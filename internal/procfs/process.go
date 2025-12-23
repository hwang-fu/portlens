package procfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ProcessInfo contains information about a process.
type ProcessInfo struct {
	PID  int
	Name string
}

// FindProcessBySocket finds the process that owns a socket with the given inode.
func FindProcessBySocket(inode uint64) (*ProcessInfo, error) {
	// Scan /proc/[pid]/fd/ directories
	procs, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("read /proc: %w", err)
	}

	target := fmt.Sprintf("socket:[%d]", inode)

	for _, proc := range procs {
		// Skip non-numeric entries (not PIDs)
		pid, err := strconv.Atoi(proc.Name())
		if err != nil {
			continue
		}

		// Check each file descriptor
		fdPath := filepath.Join("/proc", proc.Name(), "fd")
		fds, err := os.ReadDir(fdPath)
		if err != nil {
			continue // Permission denied or process exited
		}

		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdPath, fd.Name()))
			if err != nil {
				continue
			}
			if link == target {
				name := readProcessName(pid)
				return &ProcessInfo{PID: pid, Name: name}, nil
			}
		}
	}

	return nil, nil // Not found (not an error)
}

// readProcessName reads the process name from /proc/[pid]/comm.
func readProcessName(pid int) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
