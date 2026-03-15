package helpers

import (
	"fmt"
	"net"
	"strconv"
)

const RaftPortOffset = 1024

func RaftToHttpAddress(raftAddr string) (string, error) {
	host, port, err := net.SplitHostPort(raftAddr)
	if err != nil {
		return "", fmt.Errorf("invalid raft address %q", raftAddr)
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return "", fmt.Errorf("invalid raft port %q", port)
	}

	httpPort := portInt - RaftPortOffset
	if httpPort <= 0 {
		return "", fmt.Errorf("invalid http port derived from raft port %d", portInt)
	}

	return net.JoinHostPort(host, strconv.Itoa(httpPort)), nil
}
