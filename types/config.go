package types

type Config struct {
	Database string       `yaml:"database"`
	Shards   []Shard      `yaml:"shards"`
	Raft     RaftConfig   `yaml:"raft"`
	Server   ServerConfig `yaml:"server"`
}

type RaftConfig struct {
	NodeID        string   `yaml:"node_id"`
	BindAddr      string   `yaml:"bind_addr"`      // Host or host:port; runtime port is derived from server.port + 1024
	AdvertiseAddr string   `yaml:"advertise_addr"` // Optional host or host:port; runtime port is derived from server.port + 1024
	DataDir       string   `yaml:"data_dir"`
	Peers         []string `yaml:"peers"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}
