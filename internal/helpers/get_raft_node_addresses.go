package helpers

import (
	"net"

	"github.com/hadlow/genomdb/types"
	"github.com/hashicorp/raft"
)

func GetRaftNodeAddresses(r *raft.Raft) ([]types.RaftNodeAddress, error) {
	leader := string(r.Leader())

	future := r.GetConfiguration()
	if err := future.Error(); err != nil {
		return nil, err
	}

	nodes := make([]types.RaftNodeAddress, 0, len(future.Configuration().Servers))

	for _, server := range future.Configuration().Servers {
		address := string(server.Address)
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			host = address
			port = ""
		}

		nodes = append(nodes, types.RaftNodeAddress{
			ID:          string(server.ID),
			RaftAddress: address,
			IP:          host,
			Port:        port,
			Suffrage:    SuffrageString(server.Suffrage),
			IsLeader:    address == leader,
		})
	}

	return nodes, nil
}
