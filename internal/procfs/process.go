package procfs

import (
	"fmt"
	"os"
	"strings"
)

// ProcessInfo contains information about a process.
type ProcessInfo struct {
	PID  int
	Name string
}

// readProcessName reads the process name from /proc/[pid]/comm.
func readProcessName(pid int) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
