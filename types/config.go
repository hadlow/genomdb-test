package types

type Config struct {
	Database string       `yaml:"database"`
	Shards   []Shard      `yaml:"shards"`
	Raft     RaftConfig   `yaml:"raft"`
	Server   ServerConfig `yaml:"server"`
}

type RaftConfig struct {
	NodeID        string   `yaml:"node_id"`
	BindAddr      string   `yaml:"bind_addr"`
	AdvertiseAddr string   `yaml:"advertise_addr"` // Optional: if not set, uses bind_addr
	DataDir       string   `yaml:"data_dir"`
	Peers         []string `yaml:"peers"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}
