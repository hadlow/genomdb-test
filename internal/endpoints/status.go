package endpoints

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	"github.com/hadlow/genomdb/internal/helpers"
)

type statusResponse struct {
	NodeID        string            `json:"node_id"`
	HTTPAddress   string            `json:"http_address"`
	RaftBindAddr  string            `json:"raft_bind_address"`
	RaftAdvertise string            `json:"raft_advertise_address"`
	RaftState     string            `json:"raft_state"`
	RaftLeader    string            `json:"raft_leader"`
	Peers         []peerInfo        `json:"peers"`
	Store         map[string]string `json:"store"`
	Keys          []string          `json:"keys"`
	KeyCount      int               `json:"key_count"`
}

type peerInfo struct {
	ID       string `json:"id"`
	Address  string `json:"address"`
	Suffrage string `json:"suffrage"`
}

func Status(s ServerInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := s.GetConfig()
		raftNode := s.GetRaft()

		store := s.GetFSM().SnapshotStore()
		keys := make([]string, 0, len(store))
		for key := range store {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		response := statusResponse{
			NodeID:        cfg.Raft.NodeID,
			HTTPAddress:   cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port),
			RaftBindAddr:  cfg.Raft.BindAddr,
			RaftAdvertise: cfg.Raft.AdvertiseAddr,
			RaftState:     raftNode.State().String(),
			RaftLeader:    string(raftNode.Leader()),
			Store:         store,
			Keys:          keys,
			KeyCount:      len(keys),
		}

		if cfg.Raft.AdvertiseAddr == "" {
			response.RaftAdvertise = cfg.Raft.BindAddr
		}

		future := raftNode.GetConfiguration()
		if err := future.Error(); err == nil {
			peers := make([]peerInfo, 0, len(future.Configuration().Servers))
			for _, server := range future.Configuration().Servers {
				peers = append(peers, peerInfo{
					ID:       string(server.ID),
					Address:  string(server.Address),
					Suffrage: helpers.SuffrageString(server.Suffrage),
				})
			}
			response.Peers = peers
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
