package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"gnana997/distributed-cache/cache"
	"gnana997/distributed-cache/client"
	"gnana997/distributed-cache/proto"
	"io"
	"log"
	"log/slog"
	"net"
	"time"
)

type ServerOpts struct {
	ListenAddr string
	IsLeader   bool
	LeaderAddr string
}

type Server struct {
	ServerOpts
	followers map[*client.Client]struct{}
	cache     cache.Cacher
}

func NewServer(opts ServerOpts, c cache.Cacher) *Server {
	return &Server{
		ServerOpts: opts,
		followers:  make(map[*client.Client]struct{}),
		cache:      c,
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen error: %v", err)
	}

	if !s.IsLeader && len(s.LeaderAddr) != 0 {
		go func() {
			if err := s.dialLeader(); err != nil {
				log.Println(err)
			}
		}()
	}

	slog.Info("server starting on ", "port", s.ListenAddr)

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
	slog.Info("connected to leader", "leaderAddr", s.LeaderAddr)

	binary.Write(conn, binary.LittleEndian, proto.CMDJoin)

	s.handleConn(conn)

	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	// buf := make([]byte, 2048)

	// slog.Info("connection made", "conn", conn.RemoteAddr())

	for {
		cmd, err := proto.ParseCommand(conn)
		if err != nil {
			if err == io.EOF {
				// slog.Info("Connection Closed", "conn", conn.RemoteAddr())
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
	fmt.Println("member just joined the cluster", conn.RemoteAddr())
	s.followers[client.NewFromConn(conn)] = struct{}{}
	return nil
}

func (s *Server) handleSetCommand(conn net.Conn, cmd *proto.SetCommand) error {

	slog.Info("Set command", "Key", cmd.Key, "value", cmd.Value, "ttl", cmd.TTL)

	if s.IsLeader {
		go func() {
			for follower := range s.followers {
				if err := follower.Set(context.TODO(), cmd.Key, cmd.Value, cmd.TTL); err != nil {
					if err != nil {
						slog.Error("forward to member error", "err", err)
					}
				}
			}
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
