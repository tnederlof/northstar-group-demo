package prereq

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/northstar-group-demo/democtl/internal/runtime"
)

// CheckResult represents the result of a prerequisite check
type CheckResult struct {
	Name     string
	Success  bool
	Message  string
	Required bool
}

// CheckAllSRE runs all SRE prerequisite checks
func CheckAllSRE() []CheckResult {
	var results []CheckResult

	// Required commands
	results = append(results, CheckCommand("docker", "docker", true))
	results = append(results, CheckCommand("kind", "kind", true))
	results = append(results, CheckCommand("kubectl", "kubectl", true))
	results = append(results, CheckCommand("jq", "jq", true))
	results = append(results, CheckCommand("curl", "curl", true))

	// Optional commands
	results = append(results, CheckCommand("helm", "helm (for Envoy Gateway)", false))

	// Docker status
	results = append(results, CheckDockerRunning())

	// Port availability (informational only)
	results = append(results, CheckPortInfo(8080, "SRE HTTP"))

	return results
}

// CheckAllEngineering runs all Engineering prerequisite checks
func CheckAllEngineering() []CheckResult {
	var results []CheckResult

	// Required commands
	results = append(results, CheckCommand("docker", "docker", true))
	results = append(results, CheckDockerCompose())
	results = append(results, CheckCommand("git", "git", true))
	results = append(results, CheckCommand("go", "go", true))
	results = append(results, CheckCommand("node", "node", true))
	results = append(results, CheckCommand("npm", "npm", true))
	results = append(results, CheckCommand("jq", "jq", true))
	results = append(results, CheckCommand("curl", "curl", true))

	// Optional commands
	results = append(results, CheckCommand("golangci-lint", "golangci-lint (for linting)", false))

	// Docker status
	results = append(results, CheckDockerRunning())

	// Port availability (informational only)
	results = append(results, CheckPortInfo(8082, "Engineering HTTP"))
	results = append(results, CheckPortInfo(8083, "Engineering Dashboard"))

	// Version info (informational)
	results = append(results, GetNodeVersion())
	results = append(results, GetGoVersion())

	return results
}

// CheckCommand checks if a command is available in PATH
func CheckCommand(command, displayName string, required bool) CheckResult {
	_, err := exec.LookPath(command)
	if err == nil {
		return CheckResult{
			Name:     displayName,
			Success:  true,
			Message:  fmt.Sprintf("%s found", displayName),
			Required: required,
		}
	}

	return CheckResult{
		Name:     displayName,
		Success:  false,
		Message:  fmt.Sprintf("%s not found", displayName),
		Required: required,
	}
}

// CheckDockerCompose checks if docker compose subcommand is available
func CheckDockerCompose() CheckResult {
	// First check if docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		return CheckResult{
			Name:     "docker-compose",
			Success:  false,
			Message:  "docker-compose not found (docker not in PATH)",
			Required: true,
		}
	}

	// Check if docker compose subcommand works
	cmd := exec.Command("docker", "compose", "version")
	err := cmd.Run()
	if err == nil {
		return CheckResult{
			Name:     "docker-compose",
			Success:  true,
			Message:  "docker-compose found",
			Required: true,
		}
	}

	return CheckResult{
		Name:     "docker-compose",
		Success:  false,
		Message:  "docker-compose not found",
		Required: true,
	}
}

// CheckDockerRunning checks if Docker daemon is running
func CheckDockerRunning() CheckResult {
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	if err == nil {
		return CheckResult{
			Name:     "Docker status",
			Success:  true,
			Message:  "Docker is running",
			Required: true,
		}
	}

	return CheckResult{
		Name:     "Docker status",
		Success:  false,
		Message:  "Docker is not running",
		Required: true,
	}
}

// CheckPortInfo checks port availability (informational, not required)
func CheckPortInfo(port int, description string) CheckResult {
	err := runtime.CheckPortAvailable(port, description)
	if err == nil {
		return CheckResult{
			Name:     fmt.Sprintf("Port %d (%s)", port, description),
			Success:  true,
			Message:  fmt.Sprintf("%s (port %d) is available", description, port),
			Required: false,
		}
	}

	// Port is in use - this is just informational
	// Parse the error message to get details
	errMsg := err.Error()
	lines := strings.Split(errMsg, "\n")
	if len(lines) > 0 {
		// Extract the first line which has the main info
		msg := strings.TrimSpace(lines[0])
		return CheckResult{
			Name:     fmt.Sprintf("Port %d (%s)", port, description),
			Success:  true, // Not a failure, just informational
			Message:  msg + " (only an issue if starting runtime)",
			Required: false,
		}
	}

	return CheckResult{
		Name:     fmt.Sprintf("Port %d (%s)", port, description),
		Success:  true,
		Message:  fmt.Sprintf("%s (port %d) is in use (only an issue if starting runtime)", description, port),
		Required: false,
	}
}

// GetNodeVersion gets the Node.js version (informational)
func GetNodeVersion() CheckResult {
	cmd := exec.Command("node", "--version")
	output, err := cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(output))
		return CheckResult{
			Name:     "Node version",
			Success:  true,
			Message:  version,
			Required: false,
		}
	}

	return CheckResult{
		Name:     "Node version",
		Success:  false,
		Message:  "Unable to get Node version",
		Required: false,
	}
}

// GetGoVersion gets the Go version (informational)
func GetGoVersion() CheckResult {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(output))
		return CheckResult{
			Name:     "Go version",
			Success:  true,
			Message:  version,
			Required: false,
		}
	}

	return CheckResult{
		Name:     "Go version",
		Success:  false,
		Message:  "Unable to get Go version",
		Required: false,
	}
}

// CountErrors counts the number of failed required checks
func CountErrors(results []CheckResult) int {
	errors := 0
	for _, result := range results {
		if !result.Success && result.Required {
			errors++
		}
	}
	return errors
}

// FormatResults formats check results for display
func FormatResults(results []CheckResult) string {
	var lines []string
	
	for _, result := range results {
		var symbol string
		if result.Success {
			if result.Required {
				symbol = "\033[0;32m✓\033[0m" // Green checkmark
			} else {
				symbol = "\033[0;32m✓\033[0m" // Green checkmark for optional
			}
		} else {
			if result.Required {
				symbol = "\033[0;31m✗\033[0m" // Red X
			} else {
				symbol = "\033[0;33mℹ\033[0m" // Yellow info
			}
		}
		
		lines = append(lines, fmt.Sprintf("%s %s", symbol, result.Message))
	}
	
	return strings.Join(lines, "\n")
}
