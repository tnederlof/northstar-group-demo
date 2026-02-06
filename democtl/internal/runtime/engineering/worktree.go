package engineering

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/northstar-group-demo/democtl/internal/scenario"
)

// WorktreeInit creates a git worktree from the base ref and applies broken patches
func WorktreeInit(repoRoot string, s *scenario.Scenario) error {
	worktreeDir := filepath.Join(s.Dir, "worktree")
	
	// Check if worktree already exists
	if _, err := os.Stat(worktreeDir); err == nil {
		fmt.Println("Worktree already exists")
		return nil
	}
	
	// Get git config from manifest
	baseRef := s.Manifest.Git.BaseRef
	workBranch := s.Manifest.Git.WorkBranch
	
	if baseRef == "" || workBranch == "" {
		return fmt.Errorf("scenario manifest missing git.base_ref or git.work_branch")
	}
	
	fmt.Printf("Creating worktree for scenario: %s\n", s.Identifier)
	fmt.Printf("  Base ref: %s\n", baseRef)
	fmt.Printf("  Workshop branch: %s\n", workBranch)
	fmt.Printf("  Path: %s\n", worktreeDir)
	
	// Best-effort fetch and prune
	fmt.Println("Fetching refs from origin...")
	cmd := exec.Command("git", "fetch", "origin", "--tags")
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: git fetch failed: %v\n", err)
	}
	
	cmd = exec.Command("git", "worktree", "prune")
	cmd.Dir = repoRoot
	_ = cmd.Run() // ignore errors
	
	// Validate base commit exists locally
	cmd = exec.Command("git", "cat-file", "-e", baseRef+"^{commit}")
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("base commit %s not found in local repository", baseRef)
	}
	
	// Create worktree at base commit (no branch creation yet)
	cmd = exec.Command("git", "worktree", "add", worktreeDir, baseRef)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}
	
	// In the worktree, force-create/reset the workshop branch to base
	cmd = exec.Command("git", "switch", "-C", workBranch, baseRef)
	cmd.Dir = worktreeDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create workshop branch: %w", err)
	}
	
	// Apply broken patches
	brokenPatchesDir := s.Manifest.Git.GetBrokenPatchesDir()
	brokenPatches, err := listPatchFiles(s, brokenPatchesDir)
	if err != nil {
		return fmt.Errorf("failed to list broken patches: %w", err)
	}
	
	if len(brokenPatches) > 0 {
		fmt.Printf("Applying %d broken patch(es)...\n", len(brokenPatches))
		if err := applyPatchSeries(worktreeDir, brokenPatches, false); err != nil {
			return fmt.Errorf("failed to apply broken patches for %s: %w", s.Identifier, err)
		}
	}
	
	fmt.Println()
	fmt.Println("Worktree created successfully!")
	fmt.Printf("You can commit your changes to the workshop branch: %s\n", workBranch)
	
	return nil
}

// ResetWorktreeToStage resets worktree to a specific stage (broken or solved) by applying patches
// This is intentionally destructive - it discards all uncommitted/untracked changes
func ResetWorktreeToStage(repoRoot string, s *scenario.Scenario, stage string) error {
	worktreeDir := filepath.Join(s.Dir, "worktree")
	
	if _, err := os.Stat(worktreeDir); err != nil {
		return fmt.Errorf("worktree does not exist, use init first")
	}
	
	baseRef := s.Manifest.Git.BaseRef
	workBranch := s.Manifest.Git.WorkBranch
	
	if baseRef == "" || workBranch == "" {
		return fmt.Errorf("scenario manifest missing git.base_ref or git.work_branch")
	}
	
	fmt.Printf("Resetting worktree to %s stage...\n", stage)
	
	// Best-effort cleanup of any in-progress git am
	_ = abortAmIfInProgress(worktreeDir)
	
	// Hard reset to base ref
	cmd := exec.Command("git", "reset", "--hard", baseRef)
	cmd.Dir = worktreeDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reset to base: %w", err)
	}
	
	// Clean untracked files and directories
	cmd = exec.Command("git", "clean", "-fdx")
	cmd.Dir = worktreeDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clean worktree: %w", err)
	}
	
	// Ensure branch is aligned with base
	cmd = exec.Command("git", "switch", "-C", workBranch, baseRef)
	cmd.Dir = worktreeDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reset workshop branch: %w", err)
	}
	
	// Apply stage-specific patches
	var patchesDir string
	if stage == "broken" {
		patchesDir = s.Manifest.Git.GetBrokenPatchesDir()
	} else if stage == "solved" {
		patchesDir = s.Manifest.Git.GetSolvedPatchesDir()
	} else {
		return fmt.Errorf("unknown stage: %s", stage)
	}
	
	patches, err := listPatchFiles(s, patchesDir)
	if err != nil {
		return fmt.Errorf("failed to list patches for %s stage: %w", stage, err)
	}
	
	if len(patches) > 0 {
		fmt.Printf("Applying %d patch(es) for %s stage...\n", len(patches), stage)
		if err := applyPatchSeries(worktreeDir, patches, false); err != nil {
			return fmt.Errorf("failed to apply %s patches for %s: %w", stage, s.Identifier, err)
		}
	}
	
	// Verify worktree is clean
	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = worktreeDir
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check worktree status: %w", err)
	}
	
	if len(strings.TrimSpace(string(output))) > 0 {
		return fmt.Errorf("worktree is not clean after applying %s patches", stage)
	}
	
	fmt.Println()
	fmt.Printf("Worktree reset to %s stage successfully\n", stage)
	
	return nil
}

// WorktreeResetToBroken resets worktree to the broken stage
func WorktreeResetToBroken(repoRoot string, s *scenario.Scenario, force bool) error {
	return ResetWorktreeToStage(repoRoot, s, "broken")
}

// WorktreeResetToSolved resets worktree to the solved stage
func WorktreeResetToSolved(repoRoot string, s *scenario.Scenario, force bool) error {
	return ResetWorktreeToStage(repoRoot, s, "solved")
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
