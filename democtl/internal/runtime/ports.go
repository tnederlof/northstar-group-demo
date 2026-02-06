package runtime

import (
	"fmt"
	"os/exec"
	"strings"
)

// CheckPortAvailable checks if a port is available (not in use)
// Returns an error if the port is in use
func CheckPortAvailable(port int, description string) error {
	// Use lsof to check if port is in use
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port))
	output, err := cmd.CombinedOutput()
	
	// If lsof exits with non-zero and no output, port is available
	if err != nil && len(output) == 0 {
		return nil
	}
	
	// If lsof found something, parse the output to get process info
	if len(output) > 0 {
		lines := strings.Split(string(output), "\n")
		// Skip header line if present
		for i, line := range lines {
			if i == 0 && strings.HasPrefix(line, "COMMAND") {
				continue
			}
			if strings.TrimSpace(line) == "" {
				continue
			}
			// Parse the line to get PID
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				pid := fields[1]
				cmd := fields[0]
				return fmt.Errorf("port %d (%s) is in use by PID %s (%s)\nStop the process using port %d and try again", port, description, pid, cmd, port)
			}
		}
		
		// Couldn't parse, but port is in use
		return fmt.Errorf("port %d (%s) is in use\nStop the process using port %d and try again", port, description, port)
	}
	
	return nil
}
