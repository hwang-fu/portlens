package capture

import "syscall"

// Socket represents a raw packet capture socket.
type Socket struct {
	fd int
}

// NewSocket creates a new AF_PACKET socket for capturing raw packets.
func NewSocket() (*Socket, error) {
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
}

// htons converts a short (uint16) from host to network byte order.
func htons(i uint16) uint16 {
	return (i<<8)&0xff00 | i>>8
}
