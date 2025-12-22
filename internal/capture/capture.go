package capture

import (
	"fmt"
	"syscall"
)

// Socket represents a raw packet capture socket.
type Socket struct {
	fd int
}

// NewSocket creates a new AF_PACKET socket for capturing raw packets.
func NewSocket() (*Socket, error) {
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
	if err != nil {
		return nil, fmt.Errorf("create socket: %w", err)
	}
	return &Socket{fd: fd}, nil
}

// Close closes the socket.
func (s *Socket) Close() error {
	return syscall.Close(s.fd)
}

// htons converts a short (uint16) from host to network byte order.
func htons(i uint16) uint16 {
	return (i<<8)&0xff00 | i>>8
}
