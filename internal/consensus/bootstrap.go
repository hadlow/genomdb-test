package consensus

import "github.com/hashicorp/raft"

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
	// Otherwise, the node will need to be added via the /join endpoint
	addFuture := r.AddVoter(
		raft.ServerID(nodeID),
		raft.ServerAddress(addr),
		0,
		0,
	)

	if err := addFuture.Error(); err != nil {
		if err == raft.ErrNotLeader {
			// Not the leader - this is expected for new nodes
			// The node can be added manually via the /join endpoint on the leader
			// For automatic joining, you can call the join endpoint programmatically
			return nil
		}
		return err
	}

	return nil
}
