package netlink

// NETLINK_SOCK_DIAG is the Netlink protocol for socket diagnostics.
const NETLINK_SOCK_DIAG = 4

// Socket represents a Netlink socket for querying socket information.
type Socket struct {
	fd int
}
