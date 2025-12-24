package main

import (
	"log"
	"net"

	"github.com/hwang-fu/portlens/internal/procfs"
)

// logDebug logs a message only if debug mode is enabled.
func logDebug(format string, args ...any) {
	if cfg.debug {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// getDirection returns "in", "out", or "unknown" based on src/dst IPs.
func getDirection(srcIP, dstIP string, localIPs map[string]bool) string {
	srcLocal := localIPs[srcIP]
	dstLocal := localIPs[dstIP]

	if srcLocal && !dstLocal {
		return "out"
	}
	if !srcLocal && dstLocal {
		return "in"
	}
	return "unknown"
}

// lookupProcess finds the process owning a socket.
func lookupProcess(protocol string, srcIP, dstIP net.IP, srcPort, dstPort uint16) *procfs.ProcessInfo {
	inode, err := procfs.FindSocketInode(protocol, srcIP, srcPort, dstIP, dstPort)
	if err != nil || inode == 0 {
		return nil
	}

	proc, err := procfs.FindProcessBySocket(inode)
	if err != nil {
		return nil
	}
	return proc
}

// matchesProcessFilter checks if proc matches the configured process filters.
// Returns true if the packet should be processed, false if it should be skipped.
func matchesProcessFilter(proc *procfs.ProcessInfo) bool {
	if cfg.process != "" && (proc == nil || proc.Name != cfg.process) {
		return false
	}
	if cfg.pid != 0 && (proc == nil || proc.PID != cfg.pid) {
		return false
	}
	return true
}
