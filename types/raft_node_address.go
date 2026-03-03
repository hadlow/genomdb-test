package types

type RaftNodeAddress struct {
	ID          string `json:"id"`
	RaftAddress string `json:"raft_address"`
	IP          string `json:"ip"`
	Port        string `json:"port"`
	Suffrage    string `json:"suffrage"`
	IsLeader    bool   `json:"is_leader"`
}
