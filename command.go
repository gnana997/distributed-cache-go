package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"errors"
)

type Command string

const (
	CMDSet    Command = "SET"
	CMDGet    Command = "GET"
	CMDHas    Command = "HAS"
	CMDDelete Command = "DELETE"
)

type Message struct {
	Cmd   Command
	Key   []byte
	Value []byte
	TTL   time.Duration
}

func (m *Message) ToBytes() []byte {
	switch m.Cmd {
	case CMDSet:
		cmd := fmt.Sprintf("%s %s %s %d", m.Cmd, m.Key, m.Value, m.TTL)
		return []byte(cmd)
	case CMDGet:
		cmd := fmt.Sprintf("%s %s", m.Cmd, m.Key)
		return []byte(cmd)
	default:
		panic("unknown command")
	}
}

func parseCommand(raw []byte) (*Message, error) {
	var (
		rawStr = string(raw)
		parts  = strings.Split(rawStr, " ")
	)

	if len(parts) < 2 {
		return nil, errors.New("invalid command")
	}

	msg := &Message{
		Cmd: Command(parts[0]),
		Key: []byte(parts[1]),
	}

	switch msg.Cmd {
	case CMDSet:
		if len(parts) < 4 {
			return nil, errors.New("invalid SET command format")
		}
		msg.Value = []byte(parts[2])
		ttl, err := strconv.Atoi(parts[3])
		if err != nil {
			return nil, errors.New("invalid SET command ttl format")
		}
		msg.TTL = time.Duration(ttl)

		return msg, nil

	case CMDGet:
		if len(parts) < 2 {
			return nil, errors.New("invalid GET command format")
		}

		return msg, nil

	case CMDHas:
		if len(parts) < 2 {
			return nil, errors.New("invalid HAS command format")
		}

		return msg, nil

	case CMDDelete:
		if len(parts) < 2 {
			return nil, errors.New("invalid DELETE command format")
		}

		return msg, nil
	}

	return nil, errors.New("invalid command")
}
