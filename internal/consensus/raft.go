package consensus

import (
	"bytes"
	"encoding/gob"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

func NewRaftNode(
	dataDir, bindAddr, advertiseAddr, nodeID string,
	fsm *FSM,
) (*raft.Raft, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeID)

	// TCP transport
	// For local development with 127.0.0.1, bind to 0.0.0.0 but advertise as 127.0.0.1
	actualBindAddr := bindAddr

	// If caller provided an advertiseAddr, use it; otherwise default to bindAddr
	if advertiseAddr == "" {
		advertiseAddr = bindAddr
	}

	// If bind address is 127.0.0.1, bind to 0.0.0.0 instead but keep advertise as 127.0.0.1
	// This allows Raft to accept the address as advertisable while binding to all interfaces
	if host, port, err := net.SplitHostPort(bindAddr); err == nil && host == "127.0.0.1" {
		actualBindAddr = "0.0.0.0:" + port
		// keep advertiseAddr as-is (likely 127.0.0.1:port)
	}

	advertiseTCPAddr, err := net.ResolveTCPAddr("tcp", advertiseAddr)
	if err != nil {
		return nil, err
	}

	transport, err := raft.NewTCPTransport(actualBindAddr, advertiseTCPAddr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}

	// Storage
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-log.db"))
	if err != nil {
		return nil, err
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-stable.db"))
	if err != nil {
		return nil, err
	}

	snapshots, err := raft.NewFileSnapshotStore(dataDir, 1, os.Stderr)
	if err != nil {
		return nil, err
	}

	r, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func ApplySet(r *raft.Raft, key, value string) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(Command{
		Op:    "set",
		Key:   key,
		Value: value,
	}); err != nil {
		return err
	}

	f := r.Apply(buf.Bytes(), 5*time.Second)
	return f.Error()
}

func ApplyDelete(r *raft.Raft, key string) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(Command{
		Op:  "delete",
		Key: key,
	}); err != nil {
		return err
	}

	f := r.Apply(buf.Bytes(), 5*time.Second)
	return f.Error()
}

func RequireLeader(r *raft.Raft, w http.ResponseWriter, req *http.Request) bool {
	if r.State() == raft.Leader {
		return true
	}

	leaderAddr := r.Leader()
	if leaderAddr == "" {
		http.Error(w, "no leader elected", http.StatusServiceUnavailable)
		return false
	}

	http.Redirect(
		w,
		req,
		"http://"+string(leaderAddr)+req.URL.Path,
		http.StatusTemporaryRedirect,
	)
	return false
}
