package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/northstar-group-demo/democtl/internal/execx"
)

// UIEnsureOpts contains options for ensuring UI dependencies
type UIEnsureOpts struct {
	RepoRoot string
	Verbose  bool
}

// UIEnsure ensures UI dependencies and Playwright are installed
func UIEnsure(opts UIEnsureOpts) error {
	uiDir := filepath.Join(opts.RepoRoot, "demo", "ui")
	
	// Check if node_modules exists and is up to date
	nodeModulesPath := filepath.Join(uiDir, "node_modules")
	packageLockPath := filepath.Join(uiDir, "package-lock.json")
	
	nodeModulesInfo, nodeModulesErr := os.Stat(nodeModulesPath)
	packageLockInfo, packageLockErr := os.Stat(packageLockPath)
	
	shouldInstall := false
	reason := ""
	
	if nodeModulesErr != nil || packageLockErr != nil {
		shouldInstall = true
		reason = "initial installation"
	} else if packageLockInfo.ModTime().After(nodeModulesInfo.ModTime()) {
		shouldInstall = true
		reason = "lockfile changed"
	} else {
		fmt.Println("\033[0;34m==>\033[0m UI dependencies already installed")
	}
	
	if shouldInstall {
		if reason != "" {
			fmt.Printf("\033[0;34m==>\033[0m Installing UI dependencies (%s)...\n", reason)
		} else {
			fmt.Println("\033[0;34m==>\033[0m Installing UI dependencies...")
		}
		
		cmd := exec.Command("npm", "ci")
		cmd.Dir = uiDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install UI dependencies: %w", err)
		}
	}
	
	// Check if Playwright browser is installed
	needsPlaywright := false
	
	// Check if npx playwright works
	checkCmd := exec.Command("npx", "playwright", "--version")
	checkCmd.Dir = uiDir
	if err := checkCmd.Run(); err != nil {
		needsPlaywright = true
	}
	
	// Check if chromium browser is installed
	homeDir, err := os.UserHomeDir()
	if err == nil {
		chromiumPath := filepath.Join(homeDir, ".cache", "ms-playwright", "chromium-*")
		matches, _ := filepath.Glob(chromiumPath)
		if len(matches) == 0 {
			needsPlaywright = true
		}
	}
	
	if needsPlaywright {
		fmt.Println("\033[0;34m==>\033[0m Installing Playwright browser...")
		
		cmd := exec.Command("npx", "playwright", "install", "chromium")
		cmd.Dir = uiDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install Playwright browser: %w", err)
		}
	}
	
	return nil
}

// UICleanup removes UI dependencies and Playwright browsers
func UICleanup(repoRoot string) error {
	uiDir := filepath.Join(repoRoot, "demo", "ui")
	
	fmt.Println("\033[0;34m==>\033[0m Removing UI dependencies and Playwright browsers...")
	
	// Remove node_modules
	nodeModulesPath := filepath.Join(uiDir, "node_modules")
	if err := os.RemoveAll(nodeModulesPath); err != nil {
		return fmt.Errorf("failed to remove node_modules: %w", err)
	}
	
	// Remove Playwright cache
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	
	playwrightCache := filepath.Join(homeDir, ".cache", "ms-playwright")
	if err := os.RemoveAll(playwrightCache); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove Playwright cache: %w", err)
	}
	
	return nil
}

// UIStatus returns the installation status of UI dependencies
func UIStatus(repoRoot string) string {
	uiDir := filepath.Join(repoRoot, "demo", "ui")
	nodeModulesPath := filepath.Join(uiDir, "node_modules")
	
	if _, err := os.Stat(nodeModulesPath); err == nil {
		return "\033[0;32minstalled\033[0m"
	}
	return "\033[0;33mnot installed\033[0m"
}

// GetNpmCommand returns the npm command with proper error handling
func GetNpmCommand(args []string, dir string) *exec.Cmd {
	cmd := exec.Command("npm", args...)
	cmd.Dir = dir
	return cmd
}

// RunInUI runs a command in the UI directory with proper error handling
func RunInUI(repoRoot string, command string, args []string) error {
	uiDir := filepath.Join(repoRoot, "demo", "ui")
	
	fullArgs := append([]string{command}, args...)
	cmd := exec.Command("npm", fullArgs...)
	cmd.Dir = uiDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run npm %s: %w", strings.Join(fullArgs, " "), err)
	}
	
	return nil
}

// RunCommand runs a shell command with proper output handling
func RunCommand(name string, args []string, dir string) error {
	return execx.Run(name, args, execx.RunOpts{
		Dir: dir,
	})
}
