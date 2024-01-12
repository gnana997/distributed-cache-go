package main

import (
	"fmt"
	"gnana997/distributed-cache/cache"
	"gnana997/distributed-cache/client"
	"gnana997/distributed-cache/fsm"
	"gnana997/distributed-cache/proto"
	"io"
	"log"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"go.uber.org/zap"
)

type ServerOpts struct {
	RaftListenAddr string
	ListenAddr     string
	IsLeader       bool
	LeaderAddr     string
	RaftDir        string
}

type Server struct {
	ServerOpts
	followers map[*client.Client]struct{}
	cache     cache.Cacher
	logger    *zap.SugaredLogger
	raft      *raft.Raft
}

func NewServer(opts ServerOpts, c cache.Cacher) *Server {
	l, _ := zap.NewProduction()

	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(opts.RaftListenAddr)

	addr, err := net.ResolveTCPAddr("tcp", opts.RaftListenAddr)
	fmt.Printf("Addr %+v \n", addr)
	if err != nil {
		log.Fatalf("error occured to resolve tcp address: %+v", err)
	}

	raftDir := filepath.Join(opts.RaftDir)
	if err := os.MkdirAll(raftDir, 0700); err != nil {
		log.Fatalf("Failed to create Raft folder: %+v", err)
	}

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(raftDir, "raft-log.bolt"))
	if err != nil {
		log.Fatalf("Failed to create new Bolt Store: %+v", err)
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(raftDir, "raft-stable.bolt"))
	if err != nil {
		log.Fatalf("Failed to create new Bolt Store: %+v", err)
	}

	snapshotStore, err := raft.NewFileSnapshotStore(raftDir, 2, os.Stderr)
	if err != nil {
		log.Fatalf("Failed to create new Snapshot Store: %+v", err)
	}

	transport, err := raft.NewTCPTransport(opts.RaftListenAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		log.Fatalf("Failed to create TCP transport: %+v", err)
	}

	fsm := &fsm.FSM{
		Cache: c,
	}

	raft, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		log.Fatalf("Failed to create Raft: %+v", err)
	}

	return &Server{
		ServerOpts: opts,
		followers:  make(map[*client.Client]struct{}),
		cache:      c,
		logger:     l.Sugar(),
		raft:       raft,
	}
}

func (s *Server) Start() error {
	if s.IsLeader {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raft.ServerID(s.RaftListenAddr),
					Address: raft.ServerAddress(s.RaftListenAddr),
				},
			},
		}

		future := s.raft.BootstrapCluster(configuration)

		if err := future.Error(); err != nil {
			log.Fatalf("Failed to Bootstrap Raft CLuster: %+v", err)
		}
	}

	if !s.IsLeader {
		go func() {
			if err := s.dialLeader(); err != nil {
				log.Println(err)
			}
		}()
	}

	s.logger.Info("server starting", "addr", s.ListenAddr, "leader", s.IsLeader)

	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen error: %v", err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Error("accept error: %v", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) dialLeader() error {
	conn, err := net.Dial("tcp", s.LeaderAddr)
	if err != nil {
		return fmt.Errorf("failed to dial leader: %s", s.LeaderAddr)
	}
	s.logger.Info("connected to leader", "leaderAddr", s.LeaderAddr)

	joinCmd := &proto.JoinCommand{
		Addr: []byte(s.RaftListenAddr),
	}

	_, err = conn.Write(joinCmd.Bytes())
	if err != nil {
		return fmt.Errorf("write error %+v", err)
	}

	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	for {
		cmd, err := proto.ParseCommand(conn)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println("parse command error")
			break
		}
		go s.handleMessage(conn, cmd)
	}
}

func (s *Server) handleMessage(conn net.Conn, msg any) {
	switch v := msg.(type) {
	case *proto.SetCommand:
		s.handleSetCommand(conn, v)
	case *proto.GetCommand:
		s.handleGetCommand(conn, v)
	case *proto.JoinCommand:
		s.handleJoinCommand(conn, v)
	}
}

func (s *Server) handleJoinCommand(conn net.Conn, cmd *proto.JoinCommand) error {
	s.logger.Info("member just joined the cluster", conn.RemoteAddr())

	if s.IsLeader {
		future := s.raft.AddVoter(raft.ServerID(conn.RemoteAddr().String()), raft.ServerAddress(cmd.Addr), 0, 100*time.Millisecond)
		if err := future.Error(); err != nil {
			s.logger.Error("failed to initiate configuration change", "error", err)
			return err
		}

		select {
		case <-s.raft.LeaderCh():
			s.logger.Info("configuration change applied")
		case <-time.After(time.Second * 2):
			s.logger.Error("timeout waiting for the configuration changes")
			return fmt.Errorf("timeout waiting for the configuration changes")
		}
	}
	return nil
}

func (s *Server) handleSetCommand(conn net.Conn, cmd *proto.SetCommand) error {

	s.logger.Info("Set command", "Key", string(cmd.Key), "value", string(cmd.Value), "ttl", cmd.TTL)

	if s.IsLeader {
		go func() {
			s.replicateSetEntry(cmd)
		}()
	}

	resp := proto.SetResponse{}
	if err := s.cache.Set(cmd.Key, cmd.Value, time.Duration(cmd.TTL)); err != nil {
		resp.Status = proto.StatusError
		_, err := conn.Write(resp.Bytes())
		return err
	}

	resp.Status = proto.StatusOK
	_, err := conn.Write(resp.Bytes())

	return err
}

func (s *Server) handleGetCommand(conn net.Conn, cmd *proto.GetCommand) error {

	s.logger.Info("Set command", "Key", string(cmd.Key))

	resp := proto.GetResponse{}
	val, err := s.cache.Get(cmd.Key)
	if err != nil {
		resp.Status = proto.StatusError
		_, err := conn.Write(resp.Bytes())
		return err
	}

	resp.Status = proto.StatusOK
	resp.Value = val
	_, err = conn.Write(resp.Bytes())

	return err
}

func (s *Server) replicateSetEntry(cmd *proto.SetCommand) {
	if !s.IsLeader {
		return
	}

	buf := cmd.Bytes()

	term, index, err := s.getLeaderInfo()
	if err != nil {
		s.logger.Errorf("Error Getting Leader Info: %+v", err)
	}

	logEntry := raft.Log{
		Type:  raft.LogCommand,
		Data:  buf,
		Index: 0,
		Term:  term,
	}

	if index != 0 {
		logEntry.Index = index + 1
	}

	future := s.raft.ApplyLog(logEntry, 10*time.Second)
	if err := future.Error(); err != nil {
		s.logger.Error("failed to replicate log entry", "err", err)
	}
}

func (s *Server) getLeaderInfo() (uint64, uint64, error) {

	stats := s.raft.Stats()

	lastIndex, err := strconv.ParseUint((stats["last_log_index"]), 10, 64)
	if err != nil {
		return 0, 0, err
	}

	term, err := strconv.ParseUint(stats["term"], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return term, lastIndex, nil
}
