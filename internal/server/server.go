package server

import (
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/hashicorp/raft"

	"github.com/hadlow/genomdb/internal/consensus"
	"github.com/hadlow/genomdb/internal/database"
	"github.com/hadlow/genomdb/internal/endpoints"
	"github.com/hadlow/genomdb/types"
)

type Server struct {
	database *database.Database
	close    func() error
	config   *types.Config
	raft     *raft.Raft
	fsm      *consensus.FSM
}

func (s *Server) GetDatabase() *database.Database {
	return s.database
}

func (s *Server) GetConfig() *types.Config {
	return s.config
}

func (s *Server) GetRaft() *raft.Raft {
	return s.raft
}

func (s *Server) GetFSM() *consensus.FSM {
	return s.fsm
}

func NewServer(config *types.Config) (s *Server, err error) {
	database, close, err := database.NewDatabase(config.Database)
	if err != nil {
		return nil, err
	}
	database.SetBucket("main")

	fsm := consensus.NewFSM()

	// Determine advertise address: prefer explicit config, else derive
	advertiseAddr := config.Raft.AdvertiseAddr
	if advertiseAddr == "" {
		// default to bind addr
		advertiseAddr = config.Raft.BindAddr
		// if bind host is 0.0.0.0, try to pick a non-loopback interface IP
		if host, port, err := net.SplitHostPort(config.Raft.BindAddr); err == nil && (host == "0.0.0.0" || host == "") {
			if ip := pickOutboundIP(); ip != "" {
				advertiseAddr = ip + ":" + port
			}
		}
	}

	raftNode, err := consensus.NewRaftNode(
		config.Raft.DataDir,
		config.Raft.BindAddr,
		advertiseAddr,
		config.Raft.NodeID,
		fsm,
	)

	if err != nil {
		return nil, err
	}

	// Bootstrap or join cluster
	if len(config.Raft.Peers) == 0 {
		// First node - bootstrap the cluster
		if err := consensus.Bootstrap(raftNode, config.Raft.NodeID, config.Raft.BindAddr); err != nil {
			log.Printf("Bootstrap failed (cluster may already exist): %v", err)
		}
	} else {
		// Join existing cluster
		if err := consensus.JoinCluster(raftNode, config.Raft.NodeID, config.Raft.BindAddr, config.Raft.Peers); err != nil {
			log.Printf("Join cluster failed: %v", err)
		}
	}

	return &Server{database: database, close: close, config: config, raft: raftNode, fsm: fsm}, nil
}

func (s *Server) Serve() error {
	http.HandleFunc("/ping", endpoints.WithCORS(endpoints.Ping))
	http.HandleFunc("/get", endpoints.WithCORS(endpoints.Get(s)))
	http.HandleFunc("/put", endpoints.WithCORS(endpoints.Put(s)))
	http.HandleFunc("/join", endpoints.WithCORS(endpoints.Join(s)))
	http.HandleFunc("/read-chunk", endpoints.WithCORS(endpoints.ReadChunk(s)))
	http.HandleFunc("/write-chunk", endpoints.WithCORS(endpoints.WriteChunk(s)))
	http.HandleFunc("/status", endpoints.WithCORS(endpoints.Status(s)))

	addr := s.config.Server.Host + ":" + strconv.Itoa(s.config.Server.Port)
	log.Printf("Starting server on: " + addr)
	log.Printf("Raft node ID: %s, Bind address: %s\n", s.config.Raft.NodeID, s.config.Raft.BindAddr)

	err := http.ListenAndServe(addr, nil)

	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Close() {
	s.close()
}

// pickOutboundIP returns the first non-loopback IPv4 address found on the host, or
// empty string if none are found. This is used as a fallback advertise address
// when the bind address is 0.0.0.0 and no explicit advertise address is set.
func pickOutboundIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, i := range ifaces {
		if (i.Flags & net.FlagUp) == 0 {
			continue // interface down
		}
		if (i.Flags & net.FlagLoopback) != 0 {
			continue // loopback
		}
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not ipv4
			}
			return ip.String()
		}
	}
	return ""
}
