package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/northstar-group-demo/democtl/internal/checks"
	"github.com/northstar-group-demo/democtl/internal/env"
	"github.com/northstar-group-demo/democtl/internal/migrate"
	"github.com/northstar-group-demo/democtl/internal/patchesvalidate"
	"github.com/northstar-group-demo/democtl/internal/prereq"
	"github.com/northstar-group-demo/democtl/internal/runtime"
	"github.com/northstar-group-demo/democtl/internal/runtime/engineering"
	"github.com/northstar-group-demo/democtl/internal/runtime/sre"
	"github.com/northstar-group-demo/democtl/internal/scenario"
	"github.com/northstar-group-demo/democtl/internal/validate"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "democtl",
		Short: "Northstar Group Demo Controller",
		Long: `democtl is a CLI tool for managing Northstar Group demo scenarios.
It provides commands for setup, verification, scenario management, and checks execution.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		newSetupCmd(),
		newVerifyCmd(),
		newDoctorCmd(),
		newRunCmd(),
		newResetCmd(),
		newSolveCmd(),
		newFixItCmd(), // Hidden alias for solve
		newResetAllCmd(),
		newListScenariosCmd(),
		newDescribeScenarioCmd(),
		newValidateScenariosCmd(),
		newValidatePatchesCmd(),
		newMigrateScenarioToPatchesCmd(),
		newEnvCmd(),
		newChecksCmd(),
		newCompletionCmd(),
	)

	return cmd
}

func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Set up the demo environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := scenario.GetRepoRoot()
			if err != nil {
				return err
			}

			fmt.Println("\033[0;36mDemo Setup\033[0m")
			fmt.Println("\033[0;36m" + strings.Repeat("=", 80) + "\033[0m")
			fmt.Println()
			fmt.Println("\033[0;34m==>\033[0m This command sets up UI testing dependencies.")
			fmt.Println("\033[0;34m==>\033[0m It does NOT start runtimes - use 'democtl run <scenario>' for that.")
			fmt.Println()

			// Ensure UI dependencies
			if err := runtime.UIEnsure(runtime.UIEnsureOpts{
				RepoRoot: repoRoot,
			}); err != nil {
				return err
			}

			fmt.Println()
			fmt.Println("\033[0;32m==>\033[0m Setup complete!")
			fmt.Println()
			fmt.Println("\033[0;34m==>\033[0m Next steps:")
			fmt.Println("\033[0;34m==>\033[0m   1. Run 'democtl verify' to check prerequisites")
			fmt.Println("\033[0;34m==>\033[0m   2. Run 'democtl run <track>/<slug>' to start a scenario")
			fmt.Println("\033[0;34m==>\033[0m   3. See 'docs/GETTING_STARTED.md' for examples")

			return nil
		},
	}
}

func newVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Verify prerequisites and validate scenarios",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := scenario.GetRepoRoot()
			if err != nil {
				return err
			}

			fmt.Println("\033[0;36mVerifying Prerequisites\033[0m")
			fmt.Println("\033[0;36m" + strings.Repeat("=", 80) + "\033[0m")
			fmt.Println()

			errors := 0

			// Validate scenarios using Go validator with strict mode
			fmt.Println("\033[0;34m==>\033[0m Validating scenario manifests...")
			result, err := validate.ValidateAll(repoRoot, true)
			if err != nil {
				return err
			}

			if result.HasErrors() {
				fmt.Println("\033[0;31m✗\033[0m Scenario manifest validation failed")
				for _, verr := range result.Errors {
					fmt.Fprintf(os.Stderr, "  %s\n", verr.Error())
				}
				errors++
			} else {
				fmt.Println("\033[0;32m✓\033[0m Scenario manifests are valid")
			}

			// Check SRE prerequisites using Go implementation
			fmt.Println()
			fmt.Println("\033[0;34m==>\033[0m Checking SRE prerequisites...")
			fmt.Println()
			fmt.Println("Required commands:")
			sreResults := prereq.CheckAllSRE()
			// Print only command and docker status checks first
			for _, check := range sreResults {
				if check.Name != "Port 8080 (SRE HTTP)" {
					var symbol string
					if check.Success {
						symbol = "\033[0;32m✓\033[0m"
					} else if check.Required {
						symbol = "\033[0;31m✗\033[0m"
					} else {
						symbol = "\033[0;33mℹ\033[0m"
					}
					fmt.Printf("%s %s\n", symbol, check.Message)
				}
			}
			// Print port availability separately
			fmt.Println()
			fmt.Println("Port availability:")
			for _, check := range sreResults {
				if check.Name == "Port 8080 (SRE HTTP)" {
					fmt.Printf("\033[0;32m✓\033[0m %s\n", check.Message)
				}
			}
			sreErrors := prereq.CountErrors(sreResults)
			if sreErrors > 0 {
				errors += sreErrors
			}

			// Check Engineering prerequisites using Go implementation
			fmt.Println()
			fmt.Println("\033[0;34m==>\033[0m Checking Engineering prerequisites...")
			fmt.Println()
			fmt.Println("Required commands:")
			engResults := prereq.CheckAllEngineering()
			// Print command and status checks, excluding ports and versions
			for _, check := range engResults {
				if !strings.HasPrefix(check.Name, "Port ") && !strings.HasSuffix(check.Name, " version") {
					var symbol string
					if check.Success {
						symbol = "\033[0;32m✓\033[0m"
					} else if check.Required {
						symbol = "\033[0;31m✗\033[0m"
					} else {
						symbol = "\033[0;33mℹ\033[0m"
					}
					fmt.Printf("%s %s\n", symbol, check.Message)
				}
			}
			// Print port availability
			fmt.Println()
			fmt.Println("Port availability:")
			for _, check := range engResults {
				if strings.HasPrefix(check.Name, "Port ") {
					fmt.Printf("\033[0;32m✓\033[0m %s\n", check.Message)
				}
			}
			// Print versions
			fmt.Println()
			fmt.Println("Tool versions:")
			for _, check := range engResults {
				if strings.HasSuffix(check.Name, " version") {
					if check.Success {
						fmt.Printf("%s: %s\n", check.Name, check.Message)
					}
				}
			}
			engErrors := prereq.CountErrors(engResults)
			if engErrors > 0 {
				errors += engErrors
			}

			// Check UI prerequisites
			fmt.Println()
			fmt.Println("\033[0;34m==>\033[0m Checking UI prerequisites...")
			uiDir := filepath.Join(repoRoot, "demo", "ui")
			if info, err := os.Stat(uiDir); err != nil || !info.IsDir() {
				fmt.Println("\033[0;31m✗\033[0m UI directory not found")
				errors++
			} else {
				nodeModules := filepath.Join(uiDir, "node_modules")
				if _, err := os.Stat(nodeModules); err == nil {
					fmt.Println("\033[0;32m✓\033[0m UI dependencies installed")
				} else {
					fmt.Println("\033[0;33mℹ\033[0m UI dependencies not installed (run 'democtl setup')")
				}
			}

			fmt.Println()
			if errors == 0 {
				fmt.Println("\033[0;32m==>\033[0m All required prerequisites met!")
				fmt.Println()
				fmt.Println("\033[0;34m==>\033[0m Ready to run scenarios with 'democtl run <track>/<slug>'")
				return nil
			}

			fmt.Fprintf(os.Stderr, "\033[0;31m==>\033[0m %d prerequisite check(s) failed\n", errors)
			fmt.Println()
			fmt.Fprintln(os.Stderr, "\033[0;31m==>\033[0m Please install missing tools and try again")
			fmt.Fprintln(os.Stderr, "\033[0;31m==>\033[0m See docs/GETTING_STARTED.md for installation instructions")
			return fmt.Errorf("%d prerequisite check(s) failed", errors)
		},
	}
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run diagnostics (never fails, safe for interactive use)",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := scenario.GetRepoRoot()
			if err != nil {
				return err
			}

			fmt.Println("\033[0;36mDemo Status\033[0m")
			fmt.Println("\033[0;36m" + strings.Repeat("=", 80) + "\033[0m")
			fmt.Println()

			// SRE Runtime
			fmt.Println("SRE Runtime:")
			sreStatus := sre.Status(sre.DefaultKubeContext, sre.DefaultHTTPPort)
			fmt.Printf("  Cluster: %s\n", sreStatus["cluster"])
			if sre.ClusterExists(sre.DefaultKubeContext) {
				fmt.Printf("  Gateway: %s\n", sreStatus["gateway"])
			}

			fmt.Println()
			fmt.Println("Engineering Runtime:")
			engStatus := engineering.Status(engineering.DefaultHTTPPort, engineering.DefaultDashboardPort)
			fmt.Printf("  Edge Proxy: %s\n", engStatus["edge_proxy"])
			if engineering.EdgeReady() {
				fmt.Printf("  HTTP Port: %s\n", engStatus["http_port"])
				fmt.Printf("  Dashboard: %s\n", engStatus["dashboard"])
			}
			fmt.Printf("  Network: %s\n", engStatus["network"])

			fmt.Println()
			fmt.Println("UI Testing:")
			uiStatus := runtime.UIStatus(repoRoot)
			fmt.Printf("  Dependencies: %s\n", uiStatus)

			fmt.Println()
			fmt.Println("Port Configuration:")
			fmt.Printf("  SRE HTTP: %d\n", sre.DefaultHTTPPort)
			fmt.Printf("  Engineering HTTP: %d\n", engineering.DefaultHTTPPort)
			fmt.Printf("  Engineering Dashboard: %d\n", engineering.DefaultDashboardPort)

			fmt.Println()
			fmt.Println("\033[0;34m==>\033[0m Run 'democtl verify' to check all prerequisites")

			// Doctor never fails
			return nil
		},
	}
}

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <scenario>",
		Short: "Run a demo scenario",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := scenario.GetRepoRoot()
			if err != nil {
				return err
			}
			scenarioID := args[0]
			runVerify, _ := cmd.Flags().GetBool("verify")
			typeStr, _ := cmd.Flags().GetString("type")

			fmt.Printf("\033[0;36mRunning Scenario: %s\033[0m\n", scenarioID)
			fmt.Println("\033[0;36m" + strings.Repeat("=", 80) + "\033[0m")
			fmt.Println()

			// Resolve scenario
			var typeOverride scenario.ScenarioType
			if typeStr != "" {
				typeOverride = scenario.ScenarioType(typeStr)
			}

			s, err := scenario.Resolve(repoRoot, scenarioID, typeOverride)
			if err != nil {
				return fmt.Errorf("failed to load scenario: %w", err)
			}

			var scenarioURL string

			if s.Manifest.Type == scenario.TypeSRE {
				// Ensure SRE runtime
				if err := sre.EnsureRuntime(sre.RuntimeOpts{
					RepoRoot:    repoRoot,
					KubeContext: sre.DefaultKubeContext,
					HTTPPort:    sre.DefaultHTTPPort,
				}); err != nil {
					return err
				}

				// Deploy scenario
				// Extract namespace from scenario (demo-<slug>)
				_, slug, _ := scenario.ParseIdentifier(s.Identifier)
				namespace := fmt.Sprintf("demo-%s", slug)
				
				if err := sre.DeployScenario(sre.RuntimeOpts{
					RepoRoot:    repoRoot,
					KubeContext: sre.DefaultKubeContext,
					HTTPPort:    sre.DefaultHTTPPort,
				}, s.Dir, namespace, s.Manifest.Seed); err != nil {
					return err
				}

				scenarioURL = fmt.Sprintf("http://%s:%d", s.Manifest.URLHost, sre.DefaultHTTPPort)

			} else if s.Manifest.Type == scenario.TypeEngineering {
				// Ensure Engineering runtime
				if err := engineering.EnsureRuntime(engineering.RuntimeOpts{
					RepoRoot:      repoRoot,
					HTTPPort:      engineering.DefaultHTTPPort,
					DashboardPort: engineering.DefaultDashboardPort,
				}); err != nil {
					return err
				}

				// Deploy scenario
				if err := engineering.DeployScenario(engineering.RuntimeOpts{
					RepoRoot:      repoRoot,
					HTTPPort:      engineering.DefaultHTTPPort,
					DashboardPort: engineering.DefaultDashboardPort,
				}, s); err != nil {
					return err
				}

				scenarioURL = fmt.Sprintf("http://%s:%d", s.Manifest.URLHost, engineering.DefaultHTTPPort)

			} else {
				return fmt.Errorf("unknown scenario type: %s", s.Manifest.Type)
			}

			// Run verification checks if requested
			if runVerify {
				fmt.Println()
				fmt.Println("\033[0;34m==>\033[0m Running verification checks...")

				runner := checks.NewRunner(checks.RunOpts{
					Scenario:    s,
					CheckType:   checks.CheckTypeVerify,
					KubeContext: sre.DefaultKubeContext,
				})

				_, err := runner.Run()
				if err != nil {
					fmt.Println("\033[0;33m==>\033[0m Verification failed (scenario may still be starting)")
				} else {
					fmt.Println("\033[0;32m==>\033[0m Verification passed")
				}
			}

			// Print summary
			fmt.Println()
			fmt.Println("\033[0;32m==>\033[0m Scenario is running!")
			fmt.Println()
			fmt.Printf("  URL: %s\n", scenarioURL)
			fmt.Println()
			fmt.Println("\033[0;34m==>\033[0m Next steps:")
			fmt.Printf("\033[0;34m==>\033[0m   • Visit the URL above to interact with the scenario\n")
			fmt.Printf("\033[0;34m==>\033[0m   • Run 'democtl checks health %s' to check health\n", scenarioID)
			fmt.Printf("\033[0;34m==>\033[0m   • Run 'democtl reset %s' to reset\n", scenarioID)
			fmt.Printf("\033[0;34m==>\033[0m   • Run 'democtl reset-all' to clean up everything\n")

			return nil
		},
	}
	cmd.Flags().Bool("verify", true, "Run verification checks after starting scenario")
	cmd.Flags().String("type", "", "Override scenario type (sre|engineering) for disambiguation")
	return cmd
}

func newResetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset <scenario>",
		Short: "Reset a demo scenario",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := scenario.GetRepoRoot()
			if err != nil {
				return err
			}
			scenarioID := args[0]
			typeStr, _ := cmd.Flags().GetString("type")

			fmt.Printf("\033[0;36mResetting Scenario: %s\033[0m\n", scenarioID)
			fmt.Println("\033[0;36m" + strings.Repeat("=", 80) + "\033[0m")
			fmt.Println()

			// Resolve scenario
			var typeOverride scenario.ScenarioType
			if typeStr != "" {
				typeOverride = scenario.ScenarioType(typeStr)
			}

			s, err := scenario.Resolve(repoRoot, scenarioID, typeOverride)
			if err != nil {
				return fmt.Errorf("failed to load scenario: %w", err)
			}

		if s.Manifest.Type == scenario.TypeSRE {
			// Reset SRE scenario (delete namespace)
			// Extract namespace from scenario (demo-<slug>)
			_, slug, _ := scenario.ParseIdentifier(s.Identifier)
			namespace := fmt.Sprintf("demo-%s", slug)
			if err := sre.ResetScenario(sre.RuntimeOpts{
				RepoRoot:    repoRoot,
				KubeContext: sre.DefaultKubeContext,
			}, namespace); err != nil {
				return err
			}
			} else if s.Manifest.Type == scenario.TypeEngineering {
				// Reset Engineering scenario
				if err := engineering.ResetScenario(engineering.RuntimeOpts{
					RepoRoot:      repoRoot,
					HTTPPort:      engineering.DefaultHTTPPort,
					DashboardPort: engineering.DefaultDashboardPort,
				}, s); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("unknown scenario type: %s", s.Manifest.Type)
			}

			fmt.Println()
			fmt.Println("\033[0;32m==>\033[0m Scenario reset complete")
			return nil
		},
	}
	cmd.Flags().String("type", "", "Override scenario type (sre|engineering) for disambiguation")
	return cmd
}

func newSolveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "solve <scenario>",
		Short: "Apply the solved state to a scenario",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := scenario.GetRepoRoot()
			if err != nil {
				return err
			}
			scenarioID := args[0]
			typeStr, _ := cmd.Flags().GetString("type")

			fmt.Printf("\033[0;36mApplying Solved State for Scenario: %s\033[0m\n", scenarioID)
			fmt.Println("\033[0;36m" + strings.Repeat("=", 80) + "\033[0m")
			fmt.Println()

			// Resolve scenario
			var typeOverride scenario.ScenarioType
			if typeStr != "" {
				typeOverride = scenario.ScenarioType(typeStr)
			}

			s, err := scenario.Resolve(repoRoot, scenarioID, typeOverride)
			if err != nil {
				return fmt.Errorf("failed to load scenario: %w", err)
			}

			if s.Manifest.Type == scenario.TypeSRE {
				fmt.Println("\033[0;33m==>\033[0m Solve is not implemented for SRE scenarios")
				fmt.Println("\033[0;34m==>\033[0m SRE scenarios use kubectl apply to fix issues")
				return fmt.Errorf("solve not supported for SRE scenarios")
			} else if s.Manifest.Type == scenario.TypeEngineering {
				// Solve Engineering scenario
				if err := engineering.SolveScenario(engineering.RuntimeOpts{
					RepoRoot:      repoRoot,
					HTTPPort:      engineering.DefaultHTTPPort,
					DashboardPort: engineering.DefaultDashboardPort,
				}, s); err != nil {
					return err
				}

				// Run verification in solved stage
				fmt.Println()
				fmt.Println("\033[0;34m==>\033[0m Running verification checks for solved state...")

				runner := checks.NewRunner(checks.RunOpts{
					Scenario:    s,
					CheckType:   checks.CheckTypeVerify,
					Stage:       "solved",
					KubeContext: sre.DefaultKubeContext,
				})

				_, err := runner.Run()
				if err != nil {
					fmt.Println("\033[0;33m==>\033[0m Solved verification failed (scenario may still be starting)")
				} else {
					fmt.Println("\033[0;32m==>\033[0m Solved verification passed")
				}
			} else {
				return fmt.Errorf("unknown scenario type: %s", s.Manifest.Type)
			}

			fmt.Println()
			fmt.Println("\033[0;32m==>\033[0m Solve complete!")
			fmt.Println()
			fmt.Println("\033[0;34m==>\033[0m The scenario is now in the solved state")
			return nil
		},
	}
	cmd.Flags().String("type", "", "Override scenario type (sre|engineering) for disambiguation")
	return cmd
}

func newFixItCmd() *cobra.Command {
	cmd := newSolveCmd()
	cmd.Use = "fix-it <scenario>"
	cmd.Short = "(deprecated) Alias for 'solve'"
	cmd.Hidden = true
	return cmd
}

func newResetAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset-all",
		Short: "Reset all demo scenarios",
		Long:  "Without --force, prints what would happen and exits 0. With --force, performs destructive cleanup.",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := scenario.GetRepoRoot()
			if err != nil {
				return err
			}
			force, _ := cmd.Flags().GetBool("force")
			nukeUI, _ := cmd.Flags().GetBool("nuke-ui")

			fmt.Println("\033[0;36mMaster Reset\033[0m")
			fmt.Println("\033[0;36m" + strings.Repeat("=", 80) + "\033[0m")
			fmt.Println()

			if !force {
				fmt.Println("This will:")
				fmt.Println("  • Stop all Engineering containers")
				fmt.Println("  • Delete the Kind cluster (SRE)")
				fmt.Println()
				fmt.Println("With --force, it will also:")
				fmt.Println("  • Remove all worktrees and local state")
				fmt.Println()
				fmt.Println("With --force and --nuke-ui, it will also:")
				fmt.Println("  • Remove UI node_modules and Playwright browsers")
				fmt.Println()
				fmt.Println("\033[0;33m==>\033[0m This is a destructive operation")
				fmt.Println()
				fmt.Println("To proceed, run: democtl reset-all --force")
				return nil
			}

			fmt.Println("\033[0;33m==>\033[0m Proceeding with master reset (--force enabled)")
			fmt.Println()

			// Engineering cleanup
			if err := engineering.StopAllContainers(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to stop containers: %v\n", err)
			}

			if err := engineering.StopEdge(repoRoot); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to stop edge proxy: %v\n", err)
			}

			if err := engineering.RemoveNetwork(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to remove network: %v\n", err)
			}

			// Remove worktrees
			if err := engineering.RemoveWorktrees(repoRoot); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to remove worktrees: %v\n", err)
			}

			// SRE cleanup
			if err := sre.DeleteAllNamespaces(sre.RuntimeOpts{
				RepoRoot:    repoRoot,
				KubeContext: sre.DefaultKubeContext,
			}); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to delete namespaces: %v\n", err)
			}

			if err := sre.DeleteCluster(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to delete cluster: %v\n", err)
			}

			// State cleanup
			stateDir := filepath.Join(repoRoot, "demo", ".state")
			if _, err := os.Stat(stateDir); err == nil {
				fmt.Println("\033[0;34m==>\033[0m Removing local state...")
				if err := os.RemoveAll(stateDir); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to remove state directory: %v\n", err)
				}
			}

			// UI cleanup (only if --nuke-ui)
			if nukeUI {
				if err := runtime.UICleanup(repoRoot); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to cleanup UI: %v\n", err)
				}
			}

			fmt.Println()
			fmt.Println("\033[0;32m==>\033[0m Master reset complete!")
			fmt.Println()
			fmt.Println("\033[0;34m==>\033[0m To start fresh, run:")
			fmt.Println("\033[0;34m==>\033[0m   democtl setup")
			fmt.Println("\033[0;34m==>\033[0m   democtl run <track>/<slug>")
			return nil
		},
	}
	cmd.Flags().Bool("force", false, "Actually perform the reset (destructive)")
	cmd.Flags().Bool("nuke-ui", false, "Additionally delete UI dependencies and playwright caches")
	return cmd
}

func newListScenariosCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-scenarios",
		Short: "List available demo scenarios",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := scenario.GetRepoRoot()
			if err != nil {
				return err
			}

			scenarios, err := scenario.Discover(repoRoot)
			if err != nil {
				return fmt.Errorf("failed to discover scenarios: %w", err)
			}

			typeFilter, _ := cmd.Flags().GetString("type")

			fmt.Println("Available scenarios:")
			fmt.Println()

			count := 0
			for _, s := range scenarios {
				// Apply type filter
				if typeFilter != "all" && string(s.Manifest.Type) != typeFilter {
					continue
				}

				fmt.Printf("  [%s] %s: %s\n", s.Manifest.Type, s.Identifier, s.Manifest.Title)
				count++
			}

			if count == 0 {
				fmt.Println("  No scenarios found.")
			}

			fmt.Println()
			fmt.Printf("Total: %d scenario(s)\n", count)

			return nil
		},
	}
	cmd.Flags().String("type", "all", "Filter by type (sre|engineering|all)")
	return cmd
}

func newDescribeScenarioCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe-scenario <scenario>",
		Short: "Describe a demo scenario",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := scenario.GetRepoRoot()
			if err != nil {
				return err
			}

			scenarioID := args[0]
			typeStr, _ := cmd.Flags().GetString("type")
			var typeOverride scenario.ScenarioType
			if typeStr != "" {
				typeOverride = scenario.ScenarioType(typeStr)
			}

			s, err := scenario.Resolve(repoRoot, scenarioID, typeOverride)
			if err != nil {
				return err
			}

			fmt.Printf("Scenario: %s\n", s.Identifier)
			fmt.Printf("Location: %s\n", s.Dir)
			fmt.Println()
			fmt.Println("Manifest:")

			// Pretty-print the manifest
			manifestJSON, err := json.MarshalIndent(s.Manifest, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal manifest: %w", err)
			}
			fmt.Println(string(manifestJSON))

			fmt.Println()
			fmt.Println("Files:")

			// List files in scenario directory
			entries, err := os.ReadDir(s.Dir)
			if err != nil {
				return fmt.Errorf("failed to read scenario directory: %w", err)
			}

			for _, entry := range entries {
				info, err := entry.Info()
				if err != nil {
					continue
				}

				mode := info.Mode().String()
				size := info.Size()
				modTime := info.ModTime().Format("Jan 02 15:04")
				name := entry.Name()

				fmt.Printf("%s %8d %s %s\n", mode, size, modTime, name)
			}

			return nil
		},
	}
	cmd.Flags().String("type", "", "Override scenario type (sre|engineering) for disambiguation")
	return cmd
}

func newValidateScenariosCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-scenarios",
		Short: "Validate scenario manifests",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := scenario.GetRepoRoot()
			if err != nil {
				return err
			}

			strict, _ := cmd.Flags().GetBool("strict")

			fmt.Println("Validating scenarios...")
			fmt.Println()

			result, err := validate.ValidateAll(repoRoot, strict)
			if err != nil {
				return err
			}

			// Print validation errors
			for _, verr := range result.Errors {
				fmt.Fprintf(os.Stderr, "  ERROR: %s\n", verr.Error())
			}

			fmt.Println()
			fmt.Printf("Validated: %d scenario(s)\n", result.Total)

			if result.HasErrors() {
				fmt.Printf("Errors: %d\n", len(result.Errors))
				return fmt.Errorf("validation failed with %d error(s)", len(result.Errors))
			}

			fmt.Println("All scenarios valid!")
			return nil
		},
	}
	cmd.Flags().Bool("strict", false, "Enable strict validation (fails on collisions)")
	return cmd
}

func newValidatePatchesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-patches",
		Short: "Validate engineering scenario patches",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := scenario.GetRepoRoot()
			if err != nil {
				return err
			}

			scenarioID, _ := cmd.Flags().GetString("scenario")
			strict, _ := cmd.Flags().GetBool("strict")

			fmt.Println("Validating engineering scenario patches...")
			fmt.Println()

			result, err := patchesvalidate.ValidateAll(repoRoot, scenarioID, strict)
			if err != nil {
				return err
			}

			// Print validation errors
			for _, verr := range result.Errors {
				fmt.Fprintf(os.Stderr, "  ERROR: %s\n", verr.Error())
			}

			fmt.Println()
			if scenarioID != "" {
				fmt.Printf("Validated patches for: %s\n", scenarioID)
			} else {
				fmt.Printf("Validated patches for: %d engineering scenario(s)\n", result.Total)
			}

			if result.HasErrors() {
				fmt.Printf("Errors: %d\n", len(result.Errors))
				return fmt.Errorf("patch validation failed with %d error(s)", len(result.Errors))
			}

			fmt.Println("All patches valid!")
			return nil
		},
	}
	cmd.Flags().String("scenario", "", "Validate patches for a specific scenario (track/slug)")
	cmd.Flags().Bool("strict", false, "Enable strict validation (check base_ref is ancestor of HEAD)")
	return cmd
}

func newMigrateScenarioToPatchesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate-scenario-to-patches <scenario>",
		Short: "Migrate a tag-based scenario to patch-based workflow",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := scenario.GetRepoRoot()
			if err != nil {
				return err
			}

			scenarioID := args[0]
			baseRef, _ := cmd.Flags().GetString("base")

			fmt.Printf("\033[0;36mMigrating Scenario to Patches: %s\033[0m\n", scenarioID)
			fmt.Println("\033[0;36m" + strings.Repeat("=", 80) + "\033[0m")

			// Resolve scenario
			s, err := scenario.Resolve(repoRoot, scenarioID, scenario.TypeEngineering)
			if err != nil {
				return fmt.Errorf("failed to load scenario: %w", err)
			}

			if s.Manifest.Type != scenario.TypeEngineering {
				return fmt.Errorf("migration only applies to engineering scenarios")
			}

			// Perform migration
			if err := migrate.MigrateScenarioToPatches(repoRoot, s, baseRef); err != nil {
				return err
			}

			fmt.Println()
			fmt.Println("\033[0;32m==>\033[0m Migration complete!")
			fmt.Println()
			fmt.Println("\033[0;34m==>\033[0m Next steps:")
			fmt.Println("\033[0;34m==>\033[0m   1. Review the generated patches in the scenario directory")
			fmt.Println("\033[0;34m==>\033[0m   2. Run 'democtl validate-scenarios --strict' to validate manifest")
			fmt.Println("\033[0;34m==>\033[0m   3. Run 'democtl validate-patches --strict' to validate patches")
			fmt.Println("\033[0;34m==>\033[0m   4. Test the scenario with 'democtl run %s'", scenarioID)

			return nil
		},
	}
	cmd.Flags().String("base", "", "Base commit SHA (defaults to parent of broken ref)")
	return cmd
}

func newEnvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Environment variable management",
	}
	cmd.AddCommand(newEnvRenderCmd())
	return cmd
}

func newEnvRenderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "render <scenario>",
		Short: "Render environment file for a scenario",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := scenario.GetRepoRoot()
			if err != nil {
				return err
			}
			scenarioID := args[0]

			// Resolve scenario to determine type
			typeStr, _ := cmd.Flags().GetString("type")
			var typeOverride scenario.ScenarioType
			if typeStr != "" {
				typeOverride = scenario.ScenarioType(typeStr)
			}

			s, err := scenario.Resolve(repoRoot, scenarioID, typeOverride)
			if err != nil {
				return err
			}

			// Render env file using Go implementation
			outputPath, err := env.Render(env.RenderOpts{
				Scenario: s,
				RepoRoot: repoRoot,
			})
			if err != nil {
				return fmt.Errorf("failed to render environment: %w", err)
			}

			// Print the output path (same as shell script)
			fmt.Println(outputPath)
			return nil
		},
	}
	cmd.Flags().String("type", "", "Override scenario type (sre|engineering) for disambiguation")
	return cmd
}

func newChecksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checks",
		Short: "Run scenario checks",
	}
	cmd.AddCommand(
		newChecksVerifyCmd(),
		newChecksHealthCmd(),
	)
	return cmd
}

func newChecksVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify <scenario>",
		Short: "Run verification checks for a scenario",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChecks(cmd, args, "verify")
		},
	}
	cmd.Flags().String("stage", "", "Specific stage to check")
	cmd.Flags().String("only", "", "Filter checks by name")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("verbose", false, "Enable verbose output")
	cmd.Flags().String("type", "", "Override scenario type (sre|engineering) for disambiguation")
	return cmd
}

func newChecksHealthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health <scenario>",
		Short: "Run health checks for a scenario",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChecks(cmd, args, "health")
		},
	}
	cmd.Flags().String("stage", "", "Specific stage to check")
	cmd.Flags().String("only", "", "Filter checks by name")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("verbose", false, "Enable verbose output")
	cmd.Flags().String("type", "", "Override scenario type (sre|engineering) for disambiguation")
	return cmd
}

// runChecks is a helper function for verify and health commands
func runChecks(cmd *cobra.Command, args []string, checkType string) error {
	repoRoot, err := scenario.GetRepoRoot()
	if err != nil {
		return err
	}
	scenarioID := args[0]

	// Resolve scenario to determine type
	typeStr, _ := cmd.Flags().GetString("type")
	var typeOverride scenario.ScenarioType
	if typeStr != "" {
		typeOverride = scenario.ScenarioType(typeStr)
	}

	s, err := scenario.Resolve(repoRoot, scenarioID, typeOverride)
	if err != nil {
		return err
	}

	// Get flags
	stage, _ := cmd.Flags().GetString("stage")
	only, _ := cmd.Flags().GetString("only")
	jsonOut, _ := cmd.Flags().GetBool("json")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Use Go checks runner
	var checksType checks.CheckType
	if checkType == "verify" {
		checksType = checks.CheckTypeVerify
	} else {
		checksType = checks.CheckTypeHealth
	}

	runner := checks.NewRunner(checks.RunOpts{
		Scenario:    s,
		CheckType:   checksType,
		Stage:       stage,
		OnlyFilter:  only,
		JSONOutput:  jsonOut,
		Verbose:     verbose,
		KubeContext: "kind-fider-demo",
	})

	result, err := runner.Run()
	if err != nil {
		return err
	}

	// Output JSON if requested
	if jsonOut {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	}

	return nil
}

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "Generate shell completion script",
		Long: `Generate shell completion script for democtl.

To load completions:

Bash:
  $ source <(democtl completion bash)

Zsh:
  $ source <(democtl completion zsh)

Fish:
  $ democtl completion fish | source
`,
		ValidArgs: []string{"bash", "zsh", "fish"},
		Args:      cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
	return cmd
}
