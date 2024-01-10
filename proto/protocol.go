package proto

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type Status byte

func (s Status) String() string {
	switch s {
	case StatusError:
		return "ERR"
	case StatusOK:
		return "OK"
	case StatusKeyNotFound:
		return "KEYNOTFOUND"
	default:
		return "NONE"
	}
}

const (
	StatusNone Status = iota
	StatusOK
	StatusError
	StatusKeyNotFound
)

type Command byte

const (
	CmdNonce Command = iota
	CMDSet
	CMDGet
	CMDDel
	CMDJoin
)

type SetResponse struct {
	Status Status
}

func (r SetResponse) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, r.Status)

	return buf.Bytes()
}

func ParseSetResponse(r io.Reader) (*SetResponse, error) {
	resp := &SetResponse{}

	err := binary.Read(r, binary.LittleEndian, &resp.Status)

	return resp, err
}

type GetResponse struct {
	Status Status
	Value  []byte
}

func (r *GetResponse) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, r.Status)

	valueLen := int32(len(r.Value))
	binary.Write(buf, binary.LittleEndian, valueLen)
	binary.Write(buf, binary.LittleEndian, r.Value)

	return buf.Bytes()
}

func ParseGetResponse(r io.Reader) (*GetResponse, error) {
	resp := &GetResponse{}

	binary.Read(r, binary.LittleEndian, &resp.Status)

	var valueLen int32
	binary.Read(r, binary.LittleEndian, &valueLen)
	resp.Value = make([]byte, valueLen)
	binary.Read(r, binary.LittleEndian, &resp.Value)

	return resp, nil
}

type JoinCommand struct{}

type SetCommand struct {
	Key   []byte
	Value []byte
	TTL   int
}

func (c *SetCommand) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, CMDSet)

	binary.Write(buf, binary.LittleEndian, int32(len(c.Key)))
	binary.Write(buf, binary.LittleEndian, c.Key)

	binary.Write(buf, binary.LittleEndian, int32(len(c.Value)))
	binary.Write(buf, binary.LittleEndian, c.Value)

	binary.Write(buf, binary.LittleEndian, int64(c.TTL))

	return buf.Bytes()
}

type GetCommand struct {
	Key []byte
}

func (c *GetCommand) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, CMDGet)

	binary.Write(buf, binary.LittleEndian, int32(len(c.Key)))
	binary.Write(buf, binary.LittleEndian, c.Key)

	return buf.Bytes()
}

func ParseCommand(r io.Reader) (any, error) {

	var cmd Command
	if err := binary.Read(r, binary.LittleEndian, &cmd); err != nil {
		return nil, err
	}

	switch cmd {
	case CMDSet:
		return parseSetCommand(r), nil
	case CMDGet:
		return parseGetCommand(r), nil
	case CMDJoin:
		return &JoinCommand{}, nil
	default:
		return nil, fmt.Errorf("invalid command")
	}

}

func parseSetCommand(r io.Reader) *SetCommand {
	cmd := &SetCommand{}

	var keyLen int32
	binary.Read(r, binary.LittleEndian, &keyLen)
	cmd.Key = make([]byte, keyLen)
	binary.Read(r, binary.LittleEndian, &cmd.Key)

	var valueLen int32
	binary.Read(r, binary.LittleEndian, &valueLen)
	cmd.Value = make([]byte, valueLen)
	binary.Read(r, binary.LittleEndian, &cmd.Value)

	var ttl int64
	binary.Read(r, binary.LittleEndian, &ttl)
	cmd.TTL = int(ttl)

	return cmd
}

func parseGetCommand(r io.Reader) *GetCommand {
	cmd := &GetCommand{}

	var keyLen int32
	binary.Read(r, binary.LittleEndian, &keyLen)
	cmd.Key = make([]byte, keyLen)
	binary.Read(r, binary.LittleEndian, &cmd.Key)

	return cmd
}
