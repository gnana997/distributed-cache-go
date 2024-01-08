package main

import (
	"context"
	"fmt"
	"gnana997/distributed-cache/cache"
	"io"
	"log/slog"
	"net"
	"strconv"
)

type ServerOpts struct {
	ListenAddr string
	IsLeader   bool
	LeaderAddr string
}

type Server struct {
	ServerOpts
	cache cache.Cacher
}

func NewServer(opts ServerOpts, c cache.Cacher) *Server {
	return &Server{
		ServerOpts: opts,
		cache:      c,
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen error: %v", err)
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

func (s *Server) handleConn(conn net.Conn) {
	defer func() {
		conn.Close()
	}()

	buf := make([]byte, 2048)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			slog.Error("read error: %v", err)
			break
		}

		slog.Info("received message: ", "msg", string(buf[:n]))

		val, err := s.handleMessage(conn, buf[:n])

		if err != nil {
			slog.Error("handle message error: %v", err)
			_, err = conn.Write([]byte(err.Error()))
			if err != nil {
				if err == io.EOF {
					slog.Info("client disconnected")
					break
				}
				slog.Error("write error: %v", err)
				conn.Close()
			}
		}

		_, err = conn.Write(val)
		if err != nil {
			slog.Error("write error: %v", err)
			if err == io.EOF {
				slog.Info("client disconnected")
				break
			}
			slog.Error("write error: %v", err)
			conn.Close()
		}
	}

}

func (s *Server) handleMessage(conn net.Conn, msg []byte) ([]byte, error) {
	var err error
	payload, err := parseCommand(msg)

	if err != nil {
		slog.Error("parse command error: %v", err)
		return nil, err
	}

	var val []byte
	switch payload.Cmd {
	case CMDSet:
		err = s.handleSetCmd(conn, payload)
	case CMDGet:
		val, err = s.handleGetCmd(conn, payload)
	case CMDHas:
		val, err = s.handleHasCmd(conn, payload)
	case CMDDelete:
		val, err = s.handleDeleteCmd(conn, payload)
	}

	if err != nil {
		slog.Error("handle command error: %v", err)
		conn.Write([]byte(err.Error()))
		return nil, err
	}

	return val, nil
}

func (s *Server) handleSetCmd(conn net.Conn, msg *Message) error {
	if err := s.cache.Set(msg.Key, msg.Value, msg.TTL); err != nil {
		return err
	}

	go s.sendToFollowers(context.TODO(), msg)

	return nil
}

func (s *Server) handleGetCmd(conn net.Conn, msg *Message) ([]byte, error) {
	val, err := s.cache.Get(msg.Key)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (s *Server) handleHasCmd(conn net.Conn, msg *Message) ([]byte, error) {
	val := []byte(strconv.FormatBool(s.cache.Has(msg.Key)))
	return val, nil
}

func (s *Server) handleDeleteCmd(conn net.Conn, msg *Message) ([]byte, error) {
	if err := s.cache.Delete(msg.Key); err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("Deleted Key %s", msg.Key)), nil
}

func (s *Server) sendToFollowers(ctx context.Context, msg *Message) error {
	return nil
}
