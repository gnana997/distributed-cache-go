package proto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSetCommand(t *testing.T) {
	cmd := &SetCommand{
		Key:   []byte("Foo"),
		Value: []byte("Bar"),
		TTL:   2,
	}

	r := bytes.NewReader(cmd.Bytes())

	pcmd, err := ParseCommand(r)

	assert.Nil(t, err)

	assert.Equal(t, cmd, pcmd)
}

func TestParseGetCommand(t *testing.T) {
	cmd := &GetCommand{
		Key: []byte("Foo"),
	}

	r := bytes.NewReader(cmd.Bytes())

	pcmd, err := ParseCommand(r)

	assert.Nil(t, err)

	assert.Equal(t, cmd, pcmd)
}

func BenchmarkParseCommand(b *testing.B) {
	cmd := &SetCommand{
		Key:   []byte("Foo"),
		Value: []byte("Bar"),
		TTL:   2,
	}

	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(cmd.Bytes())
		ParseCommand(r)
	}
}