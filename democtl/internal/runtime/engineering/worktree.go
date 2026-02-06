package engineering

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/northstar-group-demo/democtl/internal/scenario"
)

// WorktreeInit creates a git worktree from the broken baseline tag
func WorktreeInit(repoRoot string, s *scenario.Scenario) error {
	worktreeDir := filepath.Join(s.Dir, "worktree")
	
	// Check if worktree already exists
	if _, err := os.Stat(worktreeDir); err == nil {
		fmt.Println("Worktree already exists")
		return nil
	}
	
	// Get git refs from manifest
	brokenRef := s.Manifest.Git.BrokenRef
	workBranch := s.Manifest.Git.WorkBranch
	
	if brokenRef == "" || workBranch == "" {
		return fmt.Errorf("scenario manifest missing git.broken_ref or git.work_branch")
	}
	
	fmt.Printf("Creating worktree for scenario: %s\n", s.Identifier)
	fmt.Printf("  Broken ref: %s\n", brokenRef)
	fmt.Printf("  Workshop branch: %s\n", workBranch)
	fmt.Printf("  Path: %s\n", worktreeDir)
	
	// Fetch refs from origin
	if err := fetchGitRefs(repoRoot); err != nil {
		return fmt.Errorf("failed to fetch refs: %w", err)
	}
	
	// Create worktree on local workshop branch starting from broken tag
	cmd := exec.Command("git", "worktree", "add", "-b", workBranch, worktreeDir, brokenRef)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}
	
	fmt.Println()
	fmt.Println("Worktree created successfully!")
	fmt.Printf("You can commit your changes to the workshop branch: %s\n", workBranch)
	
	return nil
}

// WorktreeResetToBroken resets worktree to the broken baseline
func WorktreeResetToBroken(repoRoot string, s *scenario.Scenario, force bool) error {
	worktreeDir := filepath.Join(s.Dir, "worktree")
	
	if _, err := os.Stat(worktreeDir); err != nil {
		return fmt.Errorf("worktree does not exist, use init first")
	}
	
	brokenRef := s.Manifest.Git.BrokenRef
	if brokenRef == "" {
		return fmt.Errorf("scenario manifest missing git.broken_ref")
	}
	
	return resetWorktreeToRef(repoRoot, worktreeDir, brokenRef, "broken baseline", force)
}

// WorktreeResetToSolved resets worktree to the solved baseline (fix-it)
func WorktreeResetToSolved(repoRoot string, s *scenario.Scenario, force bool) error {
	worktreeDir := filepath.Join(s.Dir, "worktree")
	
	if _, err := os.Stat(worktreeDir); err != nil {
		return fmt.Errorf("worktree does not exist, use init first")
	}
	
	solvedRef := s.Manifest.Git.SolvedRef
	if solvedRef == "" {
		return fmt.Errorf("scenario manifest missing git.solved_ref")
	}
	
	// Create backup branch if there are uncommitted changes
	if !force {
		if hasUncommittedChanges(worktreeDir) {
			backupBranch := fmt.Sprintf("ws/backup-%d", os.Getpid())
			fmt.Printf("Creating backup branch: %s\n", backupBranch)
			
			cmd := exec.Command("git", "branch", backupBranch, "HEAD")
			cmd.Dir = worktreeDir
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create backup branch: %v\n", err)
			}
		}
	}
	
	return resetWorktreeToRef(repoRoot, worktreeDir, solvedRef, "solved baseline", force)
}

// WorktreeRemove removes and prunes a worktree
func WorktreeRemove(repoRoot string, s *scenario.Scenario) error {
	worktreeDir := filepath.Join(s.Dir, "worktree")
	
	if _, err := os.Stat(worktreeDir); err != nil {
		fmt.Println("Worktree does not exist")
		return nil
	}
	
	fmt.Printf("Removing worktree for scenario: %s\n", s.Identifier)
	
	// Try git worktree remove first
	cmd := exec.Command("git", "worktree", "remove", worktreeDir, "--force")
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		// Fall back to rm -rf
		if err := os.RemoveAll(worktreeDir); err != nil {
			return fmt.Errorf("failed to remove worktree: %w", err)
		}
	}
	
	// Prune worktree metadata
	cmd = exec.Command("git", "worktree", "prune")
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to prune worktrees: %v\n", err)
	}
	
	fmt.Println("Worktree removed.")
	return nil
}

// RemoveWorktrees removes all engineering worktrees (for reset-all)
func RemoveWorktrees(repoRoot string) error {
	fmt.Println("\033[0;34m==>\033[0m Removing engineering worktrees...")
	
	// List all worktrees
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}
	
	// Parse worktree list to find engineering scenario worktrees
	lines := strings.Split(string(output), "\n")
	var worktreePaths []string
	
	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			// Check if it's an engineering scenario worktree
			if strings.Contains(path, "demo/engineering/scenarios") && strings.HasSuffix(path, "/worktree") {
				worktreePaths = append(worktreePaths, path)
			}
		}
	}
	
	// Remove each worktree
	for _, path := range worktreePaths {
		fmt.Printf("  Removing worktree: %s\n", path)
		
		cmd := exec.Command("git", "worktree", "remove", path, "--force")
		cmd.Dir = repoRoot
		if err := cmd.Run(); err != nil {
			// Fall back to rm -rf
			if err := os.RemoveAll(path); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to remove %s: %v\n", path, err)
			}
		}
	}
	
	// Prune worktree metadata
	cmd = exec.Command("git", "worktree", "prune")
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to prune worktrees: %v\n", err)
	}
	
	return nil
}

// Helper functions

func fetchGitRefs(repoRoot string) error {
	fmt.Println("Fetching tags and refs from origin...")
	
	cmd := exec.Command("git", "fetch", "--tags", "origin")
	cmd.Dir = repoRoot
	// Suppress output but allow errors
	if err := cmd.Run(); err != nil {
		// Don't fail on fetch errors, just warn
		fmt.Fprintf(os.Stderr, "Warning: git fetch failed: %v\n", err)
	}
	
	return nil
}

func resetWorktreeToRef(repoRoot, worktreeDir, targetRef, refName string, force bool) error {
	// Fetch refs
	if err := fetchGitRefs(repoRoot); err != nil {
		return err
	}
	
	fmt.Printf("Resetting worktree to %s (%s)\n", refName, targetRef)
	
	// Check for uncommitted changes
	if !force && hasUncommittedChanges(worktreeDir) {
		return fmt.Errorf(`
Error: Uncommitted changes present in worktree
Commit your work or use --force to discard changes

Tip: Create a backup branch before forcing reset:
  cd %s
  git branch ws-backup-$(date +%%s)
`, worktreeDir)
	}
	
	// Hard reset to target ref
	cmd := exec.Command("git", "reset", "--hard", targetRef)
	cmd.Dir = worktreeDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reset worktree: %w", err)
	}
	
	// Clean untracked files
	cmd = exec.Command("git", "clean", "-fd")
	cmd.Dir = worktreeDir
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to clean worktree: %v\n", err)
	}
	
	fmt.Println()
	fmt.Printf("Worktree reset to %s successfully\n", refName)
	
	return nil
}

func hasUncommittedChanges(worktreeDir string) bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = worktreeDir
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	return len(strings.TrimSpace(string(output))) > 0
}
