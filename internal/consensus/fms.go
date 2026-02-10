package consensus

import (
	"bytes"
	"encoding/gob"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

type Command struct {
	Op    string
	Key   string
	Value string
}

type FSM struct {
	mu    sync.Mutex
	store map[string]string
}

func NewFSM() *FSM {
	return &FSM{
		store: make(map[string]string),
	}
}

func (k *FSM) Get(key string) (string, bool) {
	k.mu.Lock()
	defer k.mu.Unlock()
	v, ok := k.store[key]
	return v, ok
}

func (k *FSM) Apply(log *raft.Log) interface{} {
	var cmd Command
	dec := gob.NewDecoder(bytes.NewBuffer(log.Data))
	if err := dec.Decode(&cmd); err != nil {
		return err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	switch cmd.Op {
	case "set":
		k.store[cmd.Key] = cmd.Value
	case "delete":
		delete(k.store, cmd.Key)
	}

	return nil
}

func (k *FSM) Snapshot() (raft.FSMSnapshot, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	copy := make(map[string]string)
	for k2, v := range k.store {
		copy[k2] = v
	}
	return &KVSnap{store: copy}, nil
}

func (k *FSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	var store map[string]string
	if err := gob.NewDecoder(rc).Decode(&store); err != nil {
		return err
	}

	k.mu.Lock()
	k.store = store
	k.mu.Unlock()
	return nil
}

type KVSnap struct {
	store map[string]string
}

func (s *KVSnap) Persist(sink raft.SnapshotSink) error {
	defer sink.Close()
	return gob.NewEncoder(sink).Encode(s.store)
}

func (s *KVSnap) Release() {}
