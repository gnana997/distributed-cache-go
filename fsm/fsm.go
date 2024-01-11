package fsm

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"gnana997/distributed-cache/cache"
	"gnana997/distributed-cache/proto"
	"io"
	"sync"
	"time"

	"github.com/hashicorp/raft"
)

type FSM struct {
	mu    sync.Mutex
	Cache cache.Cacher
}

func (f *FSM) Apply(log *raft.Log) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	buf := bytes.NewReader(log.Data)

	cmd, err := proto.ParseCommand(buf)
	if err != nil {
		fmt.Printf("Error decoding command %+v", err)
	}

	switch v := cmd.(type) {
	case *proto.SetCommand:
		f.Cache.Set(v.Key, v.Value, time.Duration(v.TTL))
	}

	return nil
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return &FSMSnapshot{cacheData: f.Cache.GetState()}, nil
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	decoder := gob.NewDecoder(rc)
	if err := decoder.Decode(&f.Cache); err != nil {
		return err
	}

	return nil
}

type FSMSnapshot struct {
	cacheData map[string][]byte
}

func (fs *FSMSnapshot) Persist(sink raft.SnapshotSink) error {
	encoder := gob.NewEncoder(sink)
	if err := encoder.Encode(fs.cacheData); err != nil {
		sink.Cancel()
		return err
	}

	if err := sink.Close(); err != nil {
		return err
	}

	return nil
}

func (fs *FSMSnapshot) Release() {}

func (fs *FSMSnapshot) GetData() map[string][]byte {
	return fs.cacheData
}
