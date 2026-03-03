package consensus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/raft"
)

func Bootstrap(r *raft.Raft, nodeID, addr string) error {
	config := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(nodeID),
				Address: raft.ServerAddress(addr),
			},
		},
	}
	return r.BootstrapCluster(config).Error()
}

func JoinCluster(r *raft.Raft, nodeID, addr string, peers []string) error {
	// If this is the first node, bootstrap
	if len(peers) == 0 {
		return Bootstrap(r, nodeID, addr)
	}

	// Get the current configuration to check if we're already in the cluster
	future := r.GetConfiguration()
	if err := future.Error(); err != nil {
		// Configuration doesn't exist, try to bootstrap
		return Bootstrap(r, nodeID, addr)
	}

	config := future.Configuration()

	// Check if we're already in the cluster
	for _, server := range config.Servers {
		if server.ID == raft.ServerID(nodeID) {
			// Already in cluster
			return nil
		}
	}

	// Try to add this server to the cluster
	// This will only work if this node is the leader
	// Otherwise, fall back to joining via peer /join endpoints.
	addFuture := r.AddVoter(
		raft.ServerID(nodeID),
		raft.ServerAddress(addr),
		0,
		0,
	)

	if err := addFuture.Error(); err != nil {
		if err != raft.ErrNotLeader {
			return err
		}

		return joinViaPeers(nodeID, addr, peers)
	}

	return nil
}

func joinViaPeers(nodeID, addr string, peers []string) error {
	payload, err := json.Marshal(map[string]string{
		"node_id":   nodeID,
		"node_addr": addr,
	})
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 3 * time.Second}
	attemptErrors := make([]string, 0, len(peers))

	for _, peerRaftAddr := range peers {
		peerHTTPAddr, err := raftAddrToHTTPAddr(peerRaftAddr)
		if err != nil {
			attemptErrors = append(attemptErrors, fmt.Sprintf("%s: %v", peerRaftAddr, err))
			continue
		}

		joinURL := "http://" + peerHTTPAddr + "/join"
		resp, err := client.Post(joinURL, "application/json", bytes.NewReader(payload))
		if err != nil {
			attemptErrors = append(attemptErrors, fmt.Sprintf("%s (%s): %v", peerRaftAddr, joinURL, err))
			continue
		}

		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return nil
		}

		attemptErrors = append(attemptErrors, fmt.Sprintf("%s (%s): status %d", peerRaftAddr, joinURL, resp.StatusCode))
	}

	return fmt.Errorf("failed joining cluster via peers: %s", strings.Join(attemptErrors, "; "))
}

func raftAddrToHTTPAddr(raftAddr string) (string, error) {
	host, port, err := net.SplitHostPort(raftAddr)
	if err != nil {
		return "", fmt.Errorf("invalid raft address %q", raftAddr)
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return "", fmt.Errorf("invalid raft port %q", port)
	}

	httpPort := portInt - 1000
	if httpPort <= 0 {
		return "", fmt.Errorf("invalid http port derived from raft port %d", portInt)
	}

	return net.JoinHostPort(host, strconv.Itoa(httpPort)), nil
}
