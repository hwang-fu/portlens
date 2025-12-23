package procfs

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
	"os"
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

// parseNetFile parses /proc/net/tcp or /proc/net/udp.
//
// File format (columns):
//
//	sl: slot number
//	local_address: hex IP:port (e.g., "0100007F:1F90" = 127.0.0.1:8080)
//	rem_address: remote hex IP:port
//	st: socket state
//	... (more fields we don't need)
//	inode: socket inode (field index 9)
func parseNetFile(path string) ([]SocketEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []SocketEntry
	scanner := bufio.NewScanner(file)

	// Skip header line ("sl  local_address rem_address ...")
	scanner.Scan()
	for scanner.Scan() {
		entry, err := parseLine(scanner.Text())
		if err != nil {
			continue // Skip malformed lines
		}
		entries = append(entries, entry)
	}

	return entries, scanner.Err()
}

// parseLine parses a single line from /proc/net/tcp or /proc/net/udp.
//
// Example line:
//
//	"   0: 0100007F:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000  0 12345 ..."
//	    ^        ^          ^          ^                                             ^
//	    slot     local      remote     state                                         inode (field 9)
func parseLine(line string) (SocketEntry, error) {
	fields := strings.Fields(line)
	if len(fields) < 10 {
		return SocketEntry{}, fmt.Errorf("not enough fields")
	}

	// fields[1] = local address (e.g., "0100007F:1F90")
	localIP, localPort, err := parseAddress(fields[1])
	if err != nil {
		return SocketEntry{}, err
	}

	// fields[2] = remote address
	remoteIP, remotePort, err := parseAddress(fields[2])
	if err != nil {
		return SocketEntry{}, err
	}

	// fields[9] = inode
	var inode uint64
	fmt.Sscanf(fields[9], "%d", &inode)

	return SocketEntry{
		LocalIP:    localIP,
		LocalPort:  localPort,
		RemoteIP:   remoteIP,
		RemotePort: remotePort,
		Inode:      inode,
	}, nil
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
