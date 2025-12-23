package procfs

import "net"

// SocketEntry represents a socket from /proc/net/tcp or /proc/net/udp.
type SocketEntry struct {
	LocalIP    net.IP
	LocalPort  uint16
	RemoteIP   net.IP
	RemotePort uint16
	Inode      uint64
}
