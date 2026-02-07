package sre

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/northstar-group-demo/democtl/internal/execx"
	"github.com/northstar-group-demo/democtl/internal/runtime"
)

const (
	DefaultHTTPPort   = 8080
	DefaultKubeContext = "kind-fider-demo"
)

// RuntimeOpts contains options for SRE runtime operations
type RuntimeOpts struct {
	RepoRoot    string
	KubeContext string
	HTTPPort    int
}

// ClusterExists checks if the Kind cluster exists and is accessible
func ClusterExists(kubeContext string) bool {
	// Check if cluster exists in kind
	cmd := exec.Command("kind", "get", "clusters")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	
	// Look for "fider-demo" in the output
	if !strings.Contains(string(output), "fider-demo") {
		return false
	}
	
	// Verify we can connect to it
	cmd = exec.Command("kubectl", "--context="+kubeContext, "cluster-info")
	err = cmd.Run()
	return err == nil
}

// GatewayReady checks if the Gateway API and Envoy Gateway are ready
func GatewayReady(kubeContext string) bool {
	// Check Gateway API CRDs
	cmd := exec.Command("kubectl", "--context="+kubeContext, "api-resources")
	output, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "gatewayclasses") {
		return false
	}
	
	// Check Envoy Gateway deployment
	cmd = exec.Command("kubectl", "--context="+kubeContext, "-n", "envoy-gateway-system", "get", "deployment", "envoy-gateway")
	if err := cmd.Run(); err != nil {
		return false
	}
	
	// Check rollout status
	cmd = exec.Command("kubectl", "--context="+kubeContext, "-n", "envoy-gateway-system", "rollout", "status", "deployment/envoy-gateway", "--timeout=10s")
	if err := cmd.Run(); err != nil {
		return false
	}
	
	// Check shared Gateway resource
	cmd = exec.Command("kubectl", "--context="+kubeContext, "-n", "envoy-gateway-system", "get", "gateway", "shared-gateway")
	return cmd.Run() == nil
}

// EnsureCluster creates the Kind cluster if it doesn't exist
func EnsureCluster(opts RuntimeOpts) error {
	if ClusterExists(opts.KubeContext) {
		fmt.Println("\033[0;34m==>\033[0m Kind cluster already exists (skipping creation)")
		return nil
	}
	
	fmt.Println("\033[0;34m==>\033[0m Creating Kind cluster...")
	
	// Check port availability
	if err := runtime.CheckPortAvailable(opts.HTTPPort, "SRE HTTP"); err != nil {
		return err
	}
	
	// Create cluster
	sreDir := filepath.Join(opts.RepoRoot, "demo", "sre")
	clusterConfig := filepath.Join(sreDir, "kind", "cluster.yaml")
	
	return execx.Run("kind", []string{"create", "cluster", "--config", clusterConfig}, execx.RunOpts{
		Dir: opts.RepoRoot,
	})
}

// EnsureGateway installs Gateway API and Envoy Gateway if not already installed
func EnsureGateway(opts RuntimeOpts) error {
	if GatewayReady(opts.KubeContext) {
		fmt.Println("\033[0;34m==>\033[0m Envoy Gateway already installed (skipping)")
		return nil
	}
	
	// Install Gateway API CRDs
	fmt.Println("\033[0;34m==>\033[0m Installing Gateway API CRDs...")
	gatewayAPIURL := "https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.0/standard-install.yaml"
	if err := execx.Run("kubectl", []string{"--context=" + opts.KubeContext, "apply", "-f", gatewayAPIURL}, execx.RunOpts{
		Dir: opts.RepoRoot,
	}); err != nil {
		return fmt.Errorf("failed to install Gateway API CRDs: %w", err)
	}
	
	// Install Envoy Gateway controller
	fmt.Println("\033[0;34m==>\033[0m Installing Envoy Gateway controller...")
	helmArgs := []string{
		"upgrade", "--install", "envoy-gateway",
		"oci://docker.io/envoyproxy/gateway-helm",
		"--kube-context=" + opts.KubeContext,
		"--version", "v1.2.4",
		"--namespace", "envoy-gateway-system",
		"--create-namespace",
		"--skip-crds",
		"--wait",
		"--timeout", "5m",
	}
	
	if err := execx.Run("helm", helmArgs, execx.RunOpts{
		Dir: opts.RepoRoot,
	}); err != nil {
		return fmt.Errorf("failed to install Envoy Gateway: %w", err)
	}
	
	// Apply shared Gateway resources
	fmt.Println("\033[0;34m==>\033[0m Applying shared Gateway resources...")
	sreDir := filepath.Join(opts.RepoRoot, "demo", "sre")
	gatewayYAML := filepath.Join(sreDir, "base", "gateway.yaml")
	
	if err := execx.Run("kubectl", []string{"--context=" + opts.KubeContext, "apply", "-f", gatewayYAML}, execx.RunOpts{
		Dir: opts.RepoRoot,
	}); err != nil {
		return fmt.Errorf("failed to apply Gateway resources: %w", err)
	}
	
	return nil
}

// EnsureRuntime ensures the SRE runtime (cluster and gateway) is ready
func EnsureRuntime(opts RuntimeOpts) error {
	fmt.Println()
	fmt.Println("\033[0;36mEnsuring SRE Runtime\033[0m")
	fmt.Println("\033[0;36m" + strings.Repeat("=", 80) + "\033[0m")
	
	if err := EnsureCluster(opts); err != nil {
		return err
	}
	
	if err := EnsureGateway(opts); err != nil {
		return err
	}
	
	fmt.Println("\033[0;32m==>\033[0m SRE runtime ready")
	return nil
}

// DeployScenario deploys an SRE scenario using kubectl apply -k
func DeployScenario(opts RuntimeOpts, scenarioDir string, namespace string, needsSeed bool) error {
	fmt.Println("\033[0;34m==>\033[0m Deploying SRE scenario...")
	
	// Apply kustomization
	if err := execx.Run("kubectl", []string{
		"--context=" + opts.KubeContext,
		"apply",
		"-k",
		scenarioDir,
	}, execx.RunOpts{
		Dir: opts.RepoRoot,
	}); err != nil {
		return err
	}
	
	// Run migrations
	if err := RunMigrations(opts, namespace); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	
	// Apply seed data if needed
	if needsSeed {
		if err := ApplySeed(opts, namespace); err != nil {
			return fmt.Errorf("failed to apply seed data: %w", err)
		}
	}
	
	return nil
}

// RunMigrations runs database migrations using a one-off fider pod
func RunMigrations(opts RuntimeOpts, namespace string) error {
	fmt.Println()
	fmt.Println("\033[0;34m==>\033[0m Running database migrations...")
	
	// Wait for postgres deployment to be available
	fmt.Println("Waiting for postgres to be ready...")
	if err := execx.Run("kubectl", []string{
		"--context=" + opts.KubeContext,
		"-n", namespace,
		"wait",
		"--for=condition=available",
		"deployment/postgres",
		"--timeout=120s",
	}, execx.RunOpts{
		Dir: opts.RepoRoot,
	}); err != nil {
		return fmt.Errorf("postgres deployment not ready: %w", err)
	}
	
	// Find the postgres pod
	fmt.Println("Finding postgres pod...")
	cmd := exec.Command("kubectl",
		"--context="+opts.KubeContext,
		"-n", namespace,
		"get", "pods",
		"-l", "app=postgres",
		"-o", "jsonpath={.items[0].metadata.name}",
	)
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		return fmt.Errorf("could not find postgres pod in namespace %s", namespace)
	}
	postgresPod := strings.TrimSpace(string(output))
	fmt.Printf("Found postgres pod: %s\n", postgresPod)
	
	// Wait for postgres to accept connections
	fmt.Println("Waiting for postgres to accept connections...")
	for i := 0; i < 30; i++ {
		checkCmd := exec.Command("kubectl",
			"--context="+opts.KubeContext,
			"-n", namespace,
			"exec", postgresPod,
			"--",
			"pg_isready", "-U", "fider", "-d", "fider",
		)
		if err := checkCmd.Run(); err == nil {
			fmt.Println("Postgres is ready!")
			break
		}
		if i == 29 {
			return fmt.Errorf("postgres did not become ready in time")
		}
		time.Sleep(time.Second)
	}
	
	// Run migrations in a one-off pod
	fmt.Println("Running fider migrate...")
	if err := execx.Run("kubectl", []string{
		"--context=" + opts.KubeContext,
		"-n", namespace,
		"run", "fider-migrate",
		"--image=ghcr.io/tnederlof/northstar-group-demo:base",
		"--restart=Never",
		"--rm",
		"--attach",
		"--env=DATABASE_URL=postgres://fider:fider@postgres:5432/fider?sslmode=disable",
		"--env=JWT_SECRET=northstar-demo-jwt-secret-not-for-production",
		"--env=EMAIL=none",
		"--env=EMAIL_NOREPLY=noreply@northstar.io",
		"--env=HOST_MODE=single",
		"--env=BASE_URL=http://localhost:8080",
		"--command",
		"--",
		"./fider", "migrate",
	}, execx.RunOpts{
		Dir: opts.RepoRoot,
	}); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	
	fmt.Println("\033[0;32m==>\033[0m Migrations completed successfully!")
	return nil
}

// ApplySeed applies seed data to the postgres database in a scenario
// Assumes migrations have already been run by RunMigrations
func ApplySeed(opts RuntimeOpts, namespace string) error {
	fmt.Println()
	fmt.Println("\033[0;34m==>\033[0m Applying seed data...")
	
	// Find the postgres pod
	fmt.Println("Finding postgres pod...")
	cmd := exec.Command("kubectl",
		"--context="+opts.KubeContext,
		"-n", namespace,
		"get", "pods",
		"-l", "app=postgres",
		"-o", "jsonpath={.items[0].metadata.name}",
	)
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		return fmt.Errorf("could not find postgres pod in namespace %s", namespace)
	}
	postgresPod := strings.TrimSpace(string(output))
	fmt.Printf("Found postgres pod: %s\n", postgresPod)
	
	// Copy seed file to pod
	seedFile := filepath.Join(opts.RepoRoot, "demo", "shared", "northstar", "seed.sql")
	if _, err := os.Stat(seedFile); err != nil {
		return fmt.Errorf("seed file not found: %s", seedFile)
	}
	
	fmt.Println("Copying seed file to pod...")
	if err := execx.Run("kubectl", []string{
		"--context=" + opts.KubeContext,
		"-n", namespace,
		"cp",
		seedFile,
		postgresPod + ":/tmp/seed.sql",
	}, execx.RunOpts{
		Dir: opts.RepoRoot,
	}); err != nil {
		return fmt.Errorf("failed to copy seed file: %w", err)
	}
	
	// Execute seed.sql
	fmt.Println("Executing seed.sql...")
	if err := execx.Run("kubectl", []string{
		"--context=" + opts.KubeContext,
		"-n", namespace,
		"exec", postgresPod,
		"--",
		"psql", "-U", "fider", "-d", "fider", "-f", "/tmp/seed.sql",
	}, execx.RunOpts{
		Dir: opts.RepoRoot,
	}); err != nil {
		return fmt.Errorf("failed to execute seed.sql: %w", err)
	}
	
	fmt.Println()
	fmt.Println("\033[0;32m==>\033[0m Seed data applied successfully!")
	return nil
}

// ResetScenario resets an SRE scenario by deleting its namespace
func ResetScenario(opts RuntimeOpts, namespace string) error {
	fmt.Println("\033[0;34m==>\033[0m Deleting namespace:", namespace)
	
	// Delete the namespace
	return execx.Run("kubectl", []string{
		"--context=" + opts.KubeContext,
		"delete",
		"namespace",
		namespace,
		"--ignore-not-found",
	}, execx.RunOpts{
		Dir: opts.RepoRoot,
	})
}

// DeleteAllNamespaces deletes all scenario namespaces (for reset-all)
func DeleteAllNamespaces(opts RuntimeOpts) error {
	fmt.Println("\033[0;34m==>\033[0m Removing SRE namespaces...")
	
	// Get all namespaces with the scenario label
	cmd := exec.Command("kubectl", "--context="+opts.KubeContext, "get", "namespaces", "-o", "name")
	output, err := cmd.Output()
	if err != nil {
		// If cluster doesn't exist, that's fine
		return nil
	}
	
	namespaces := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, ns := range namespaces {
		nsName := strings.TrimPrefix(ns, "namespace/")
		// Skip system namespaces
		if strings.HasPrefix(nsName, "kube-") || nsName == "default" || strings.HasPrefix(nsName, "envoy-gateway") {
			continue
		}
		
		// Delete the namespace
		cmd = exec.Command("kubectl", "--context="+opts.KubeContext, "delete", "namespace", nsName, "--ignore-not-found")
		_ = cmd.Run() // Ignore errors, best effort
	}
	
	return nil
}

// DeleteCluster deletes the Kind cluster
func DeleteCluster() error {
	fmt.Println("\033[0;34m==>\033[0m Deleting Kind cluster...")
	
	cmd := exec.Command("kind", "delete", "cluster", "--name", "fider-demo")
	return cmd.Run()
}

// Status returns the status information for the SRE runtime
func Status(kubeContext string, httpPort int) map[string]string {
	status := make(map[string]string)
	
	if ClusterExists(kubeContext) {
		status["cluster"] = "\033[0;32mrunning\033[0m"
		
		if GatewayReady(kubeContext) {
			status["gateway"] = "\033[0;32mready\033[0m"
		} else {
			status["gateway"] = "\033[0;33mnot ready\033[0m"
		}
	} else {
		status["cluster"] = "\033[0;33mnot running\033[0m"
	}
	
	return status
}
