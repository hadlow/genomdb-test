package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/hashicorp/raft"

	"github.com/hadlow/genomdb/internal/consensus"
)

func Join(s ServerInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only leader can add nodes
		if !consensus.RequireLeader(s.GetRaft(), w, r) {
			return
		}

		var body struct {
			NodeID   string `json:"node_id"`
			NodeAddr string `json:"node_addr"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if body.NodeID == "" || body.NodeAddr == "" {
			http.Error(w, "node_id and node_addr are required", http.StatusBadRequest)
			return
		}

		// Add the node to the cluster
		future := s.GetRaft().AddVoter(
			raft.ServerID(body.NodeID),
			raft.ServerAddress(body.NodeAddr),
			0,
			0,
		)

		if err := future.Error(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "node added to cluster",
		})
	}
}
