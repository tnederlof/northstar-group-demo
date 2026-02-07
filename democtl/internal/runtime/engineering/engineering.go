package engineering

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/northstar-group-demo/democtl/internal/env"
	"github.com/northstar-group-demo/democtl/internal/execx"
	"github.com/northstar-group-demo/democtl/internal/runtime"
	"github.com/northstar-group-demo/democtl/internal/scenario"
)

const (
	DefaultHTTPPort      = 8082
	DefaultDashboardPort = 8081
	EdgeNetworkName      = "northstar-demo"
)

// RuntimeOpts contains options for Engineering runtime operations
type RuntimeOpts struct {
	RepoRoot      string
	HTTPPort      int
	DashboardPort int
}

// EdgeReady checks if the edge proxy is running
func EdgeReady() bool {
	cmd := exec.Command("docker", "ps", "--filter", "name=northstar-edge", "--filter", "status=running", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	return strings.Contains(string(output), "northstar-edge")
}

// NetworkExists checks if the Docker network exists
func NetworkExists() bool {
	cmd := exec.Command("docker", "network", "ls", "--filter", "name="+EdgeNetworkName, "--format", "{{.Name}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	return strings.Contains(string(output), EdgeNetworkName)
}

// EnsureEdge starts the edge proxy if it's not running
func EnsureEdge(opts RuntimeOpts) error {
	if EdgeReady() {
		fmt.Println("\033[0;34m==>\033[0m Engineering edge proxy already running (skipping)")
		return nil
	}
	
	fmt.Println("\033[0;34m==>\033[0m Starting Engineering edge proxy...")
	
	// Check port availability
	if err := runtime.CheckPortAvailable(opts.HTTPPort, "Engineering HTTP"); err != nil {
		return err
	}
	
	if err := runtime.CheckPortAvailable(opts.DashboardPort, "Engineering Dashboard"); err != nil {
		return err
	}
	
	// Start edge proxy
	engDir := filepath.Join(opts.RepoRoot, "demo", "engineering")
	composeFile := filepath.Join(engDir, "compose", "edge", "docker-compose.yml")
	
	return execx.Run("docker", []string{"compose", "-f", composeFile, "up", "-d"}, execx.RunOpts{
		Dir: opts.RepoRoot,
	})
}

// EnsureNetwork ensures the Docker network exists
func EnsureNetwork() error {
	if NetworkExists() {
		return nil
	}
	
	// Network is created by edge proxy compose file
	return nil
}

// EnsureRuntime ensures the Engineering runtime (edge proxy and network) is ready
func EnsureRuntime(opts RuntimeOpts) error {
	fmt.Println()
	fmt.Println("\033[0;36mEnsuring Engineering Runtime\033[0m")
	fmt.Println("\033[0;36m" + strings.Repeat("=", 80) + "\033[0m")
	
	if err := EnsureEdge(opts); err != nil {
		return err
	}
	
	if err := EnsureNetwork(); err != nil {
		return err
	}
	
	fmt.Println("\033[0;32m==>\033[0m Engineering runtime ready")
	return nil
}

// EnsureWorktree ensures the git worktree exists for a scenario
func EnsureWorktree(opts RuntimeOpts, s *scenario.Scenario) error {
	worktreeDir := filepath.Join(s.Dir, "worktree")
	
	// Check if worktree already exists
	if _, err := os.Stat(worktreeDir); err == nil {
		return nil
	}
	
	fmt.Printf("\033[0;34m==>\033[0m Initializing worktree for %s...\n", s.Identifier)
	
	// Use Go implementation
	return WorktreeInit(opts.RepoRoot, s)
}

// DeployScenario starts an Engineering scenario using docker compose
func DeployScenario(opts RuntimeOpts, s *scenario.Scenario) error {
	fmt.Printf("\033[0;34m==>\033[0m Starting Engineering scenario: %s\n", s.Identifier)
	
	// Ensure network exists
	if err := EnsureNetwork(); err != nil {
		return fmt.Errorf("failed to ensure network: %w", err)
	}
	
	// Render environment
	envPath, err := env.Render(env.RenderOpts{
		Scenario: s,
		RepoRoot: opts.RepoRoot,
	})
	if err != nil {
		return fmt.Errorf("failed to render environment: %w", err)
	}
	
	fmt.Printf("  Environment: %s\n", envPath)
	
	// Ensure worktree exists
	if err := EnsureWorktree(opts, s); err != nil {
		return fmt.Errorf("failed to ensure worktree: %w", err)
	}
	
	// Force-reset worktree to broken stage for deterministic start
	fmt.Println("\033[0;34m==>\033[0m Ensuring worktree is in broken state...")
	if err := ResetWorktreeToStage(opts.RepoRoot, s, "broken"); err != nil {
		return fmt.Errorf("failed to reset worktree to broken: %w", err)
	}
	
	// Read secrets from env file for compose interpolation
	envVars, err := env.ReadEnvFile(envPath)
	if err != nil {
		return fmt.Errorf("failed to read env file: %w", err)
	}
	
	// Start compose
	fmt.Println()
	fmt.Println("Starting containers...")
	
	composeFile := filepath.Join(s.Dir, "docker-compose.yml")
	if _, err := os.Stat(composeFile); err != nil {
		return fmt.Errorf("docker-compose.yml not found at %s", composeFile)
	}
	
	// Run docker compose up with env vars
	if err := execx.Run("docker", []string{"compose", "--env-file", envPath, "up", "-d", "--build"}, execx.RunOpts{
		Dir: s.Dir,
		Env: envVars,
	}); err != nil {
		return err
	}
	
	// Apply seed data if needed
	if s.Manifest.Seed {
		if err := ApplySeed(opts, s); err != nil {
			return fmt.Errorf("failed to apply seed data: %w", err)
		}
	}
	
	return nil
}

// ApplySeed applies seed data to the postgres database in a scenario
func ApplySeed(opts RuntimeOpts, s *scenario.Scenario) error {
	fmt.Println()
	fmt.Println("\033[0;34m==>\033[0m Applying seed data...")
	
	// Get the container name (northstar-<slug>-postgres)
	_, slug, _ := scenario.ParseIdentifier(s.Identifier)
	container := fmt.Sprintf("northstar-%s-postgres", slug)
	
	// Seed file path
	seedFile := filepath.Join(opts.RepoRoot, "demo", "shared", "northstar", "seed.sql")
	if _, err := os.Stat(seedFile); err != nil {
		return fmt.Errorf("seed file not found: %s", seedFile)
	}
	
	fmt.Printf("Using container: %s\n", container)
	
	// Wait for postgres to accept connections
	fmt.Println("Waiting for postgres to accept connections...")
	for i := 0; i < 30; i++ {
		checkCmd := exec.Command("docker", "exec", container,
			"pg_isready", "-U", "fider", "-d", "fider")
		if err := checkCmd.Run(); err == nil {
			fmt.Println("Postgres is ready!")
			break
		}
		if i == 29 {
			return fmt.Errorf("postgres did not become ready in time")
		}
		time.Sleep(time.Second)
	}
	
	fmt.Println("Executing seed.sql...")
	
	// Open seed file for reading
	seedData, err := os.ReadFile(seedFile)
	if err != nil {
		return fmt.Errorf("failed to read seed file: %w", err)
	}
	
	// Execute seed data via direct docker exec (not compose exec)
	// This works even if the compose service isn't fully "ready" yet
	cmd := exec.Command("docker", "exec", "-i", container,
		"psql", "-U", "fider", "-d", "fider")
	cmd.Dir = s.Dir
	cmd.Stdin = strings.NewReader(string(seedData))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute seed.sql: %w", err)
	}
	
	fmt.Println()
	fmt.Println("\033[0;32m==>\033[0m Seed data applied successfully!")
	return nil
}

// StopScenario stops an Engineering scenario
func StopScenario(opts RuntimeOpts, s *scenario.Scenario) error {
	fmt.Printf("\033[0;34m==>\033[0m Stopping scenario: %s\n", s.Identifier)
	
	composeFile := filepath.Join(s.Dir, "docker-compose.yml")
	if _, err := os.Stat(composeFile); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: docker-compose.yml not found at %s\n", composeFile)
		return nil
	}
	
	return execx.Run("docker", []string{"compose", "down", "-v", "--remove-orphans"}, execx.RunOpts{
		Dir: s.Dir,
	})
}

// ResetScenario resets an Engineering scenario to the broken state
func ResetScenario(opts RuntimeOpts, s *scenario.Scenario) error {
	// Stop the scenario
	if err := StopScenario(opts, s); err != nil {
		return err
	}
	
	// Reset worktree to broken baseline using Go implementation
	fmt.Println("\033[0;34m==>\033[0m Resetting worktree to broken baseline...")
	
	return WorktreeResetToBroken(opts.RepoRoot, s, false)
}

// SolveScenario resets an Engineering scenario to the solved state
func SolveScenario(opts RuntimeOpts, s *scenario.Scenario) error {
	// Stop the scenario
	if err := StopScenario(opts, s); err != nil {
		return err
	}
	
	// Reset worktree to solved baseline
	fmt.Println("\033[0;34m==>\033[0m Resetting worktree to solved baseline...")
	
	return WorktreeResetToSolved(opts.RepoRoot, s, true)
}

// FixItScenario is a deprecated alias for SolveScenario
func FixItScenario(opts RuntimeOpts, s *scenario.Scenario) error {
	return SolveScenario(opts, s)
}

// StopAllContainers stops all Engineering containers (for reset-all)
func StopAllContainers() error {
	fmt.Println("\033[0;34m==>\033[0m Stopping Engineering containers...")
	
	// Find all containers with the northstar label
	cmd := exec.Command("docker", "ps", "-a", "--filter", "label=com.docker.compose.project=northstar", "-q")
	output, err := cmd.Output()
	if err != nil {
		return nil // No containers, that's fine
	}
	
	containerIDs := strings.Fields(string(output))
	if len(containerIDs) == 0 {
		return nil
	}
	
	// Remove containers
	removeArgs := append([]string{"rm", "-f"}, containerIDs...)
	cmd = exec.Command("docker", removeArgs...)
	return cmd.Run()
}

// StopEdge stops the edge proxy
func StopEdge(repoRoot string) error {
	fmt.Println("\033[0;34m==>\033[0m Stopping edge proxy...")
	
	engDir := filepath.Join(repoRoot, "demo", "engineering")
	composeFile := filepath.Join(engDir, "compose", "edge", "docker-compose.yml")
	
	cmd := exec.Command("docker", "compose", "-f", composeFile, "down")
	cmd.Dir = repoRoot
	_ = cmd.Run() // Ignore errors
	
	return nil
}

// RemoveNetwork removes the Docker network
func RemoveNetwork() error {
	fmt.Println("\033[0;34m==>\033[0m Removing Docker network...")
	
	cmd := exec.Command("docker", "network", "rm", EdgeNetworkName)
	_ = cmd.Run() // Ignore errors
	
	return nil
}

// Status returns the status information for the Engineering runtime
func Status(httpPort, dashboardPort int) map[string]string {
	status := make(map[string]string)
	
	if EdgeReady() {
		status["edge_proxy"] = "\033[0;32mrunning\033[0m"
		status["http_port"] = fmt.Sprintf("%d", httpPort)
		status["dashboard"] = fmt.Sprintf("http://localhost:%d", dashboardPort)
	} else {
		status["edge_proxy"] = "\033[0;33mnot running\033[0m"
	}
	
	if NetworkExists() {
		status["network"] = "\033[0;32mexists\033[0m"
	} else {
		status["network"] = "\033[0;33mnot created\033[0m"
	}
	
	return status
}
