package capture

import (
	"fmt"
	"net"
	"syscall"
)

// Socket represents a raw packet capture socket.
type Socket struct {
	fd int
}

// NewSocket creates a new AF_PACKET socket for capturing raw packets.
func NewSocket() (*Socket, error) {
	fd, err := syscall.Socket(
		syscall.AF_PACKET,
		syscall.SOCK_RAW,
		int(htons(syscall.ETH_P_ALL)),
	)
	if err != nil {
		return nil, fmt.Errorf("create socket: %w", err)
	}

	return &Socket{fd: fd}, nil
}

// Close closes the socket.
func (s *Socket) Close() error {
	return syscall.Close(s.fd)
}

// Bind binds the socket to a specific network interface.
func (s *Socket) Bind(interfaceName string) error {
	netInterface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return fmt.Errorf("get interface %s: %w", interfaceName, err)
	}

	addr := syscall.SockaddrLinklayer{
		Protocol: htons(syscall.ETH_P_ALL),
		Ifindex:  netInterface.Index,
	}

	if err := syscall.Bind(s.fd, &addr); err != nil {
		return fmt.Errorf("bind to %s: %w", interfaceName, err)
	}

	return nil
}

// ReadPacket reads a single raw packet from the socket.
// Returns the packet data and the number of bytes read.
func (s *Socket) ReadPacket(buf []byte) (int, error) {
	n, _, err := syscall.Recvfrom(s.fd, buf, 0)
	if err != nil {
		return 0, fmt.Errorf("read packet: %w", err)
	}
	return n, nil
}

// htons converts a short (uint16) from host to network byte order.
func htons(i uint16) uint16 {
	return (i<<8)&0xff00 | i>>8
}
