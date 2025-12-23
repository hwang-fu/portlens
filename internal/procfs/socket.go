package procfs

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"
)

// SocketEntry represents a socket from /proc/net/tcp or /proc/net/udp.
type SocketEntry struct {
	LocalIP    net.IP
	LocalPort  uint16
	RemoteIP   net.IP
	RemotePort uint16
	Inode      uint64
}

// FindSocketInode finds the inode for a socket matching the given 5-tuple.
// We need the inode to later find which process owns this socket.
func FindSocketInode(
	protocol string,
	srcIP net.IP, srcPort uint16,
	dstIP net.IP, dstPort uint16,
) (uint64, error) {
	var path string
	switch protocol {
	case "tcp", "TCP":
		path = "/proc/net/tcp"
	case "udp", "UDP":
		path = "/proc/net/udp"
	default:
		return 0, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// parseAddress parses "0100007F:1F90" into IP and port.
//
// Format quirks:
//   - IP is stored in LITTLE-ENDIAN hex (bytes reversed)
//     "0100007F" = 01.00.00.7F reversed = 127.0.0.1
//   - Port is stored in BIG-ENDIAN hex (normal)
//     "1F90" = 0x1F90 = 8080
func parseAddress(s string) (net.IP, uint16, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return nil, 0, fmt.Errorf("invalid address: %s", s)
	}

	ipHex, err := hex.DecodeString(parts[0])
	if err != nil {
		return nil, 0, err
	}

	// Reverse bytes: little-endian → normal order
	// ipHex = [01, 00, 00, 7F] for "0100007F"
	// reversed = [7F, 00, 00, 01] = 127.0.0.1
	ip := net.IP{ipHex[3], ipHex[2], ipHex[1], ipHex[0]}

	// Parse port as hex (big-endian, no reversal needed)
	var port uint16
	fmt.Sscanf(parts[1], "%X", &port)

	return ip, port, nil
}
