package capture

import "net"

// LocalIPs returns a set of all IP addresses on this machine.
// Returns a map for O(1) lookup.
func LocalIPs() (map[string]bool, error) {
	ips := make(map[string]bool)

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		// addr is like "192.168.1.5/24" or "::1/128"
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ips[ipNet.IP.String()] = true
	}

	return ips, nil
}
