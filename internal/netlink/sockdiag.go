package netlink

import (
	"fmt"
	"syscall"
)

// NETLINK_SOCK_DIAG is the Netlink protocol for socket diagnostics.
const NETLINK_SOCK_DIAG = 4

// Socket represents a Netlink socket for querying socket information.
type Socket struct {
	fd int
}

func NewSocket() (*Socket, error) {
	fd, err := syscall.Socket(
		syscall.AF_NETLINK,
		syscall.SOCK_DGRAM,
		NETLINK_SOCK_DIAG,
	)
	if err != nil {
		return nil, fmt.Errorf("create netlink socket: %w", err)
	}
	return &Socket{fd: fd}, nil
}

// Close closes the Netlink socket.
func (s *Socket) Close() error {
	return syscall.Close(s.fd)
}
