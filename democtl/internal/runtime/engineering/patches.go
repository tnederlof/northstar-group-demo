package engineering

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/northstar-group-demo/democtl/internal/scenario"
)

// listPatchFiles returns a sorted list of patch files in the given directory relative to scenario dir
// Returns empty list if directory doesn't exist (not an error)
func listPatchFiles(s *scenario.Scenario, relDir string) ([]string, error) {
	patchDir := filepath.Join(s.Dir, relDir)
	
	// Check if directory exists
	info, err := os.Stat(patchDir)
	if os.IsNotExist(err) {
		// Directory doesn't exist - return empty list
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat patch directory %s: %w", patchDir, err)
	}
	
	// Verify it's a directory
	if !info.IsDir() {
		return nil, fmt.Errorf("patch path exists but is not a directory: %s", patchDir)
	}
	
	// Read directory contents
	entries, err := os.ReadDir(patchDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read patch directory %s: %w", patchDir, err)
	}
	
	// Filter for .patch files and collect
	var patchFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".patch") {
			patchFiles = append(patchFiles, filepath.Join(patchDir, entry.Name()))
		}
	}
	
	// Sort lexically (e.g., 0001-, 0002-, etc.)
	sort.Strings(patchFiles)
	
	return patchFiles, nil
}

// applyPatchSeries applies a series of patches in order
// allow3way enables git am --3way for conflict resolution
func applyPatchSeries(worktreeDir string, patchFiles []string, allow3way bool) error {
	for _, patchFile := range patchFiles {
		if err := applyPatch(worktreeDir, patchFile, allow3way); err != nil {
			// Best effort abort
			_ = abortAmIfInProgress(worktreeDir)
			return fmt.Errorf("failed to apply patch %s: %w", filepath.Base(patchFile), err)
		}
	}
	return nil
}

// applyPatch applies a single patch file
func applyPatch(worktreeDir, patchFile string, allow3way bool) error {
	args := []string{"am", "--keep-cr", "--quiet"}
	if allow3way {
		args = append(args, "--3way")
	}
	args = append(args, patchFile)
	
	cmd := exec.Command("git", args...)
	cmd.Dir = worktreeDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git am failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// abortAmIfInProgress aborts any in-progress git am session (best effort)
func abortAmIfInProgress(worktreeDir string) error {
	cmd := exec.Command("git", "am", "--abort")
	cmd.Dir = worktreeDir
	// Ignore errors - there may not be an am session in progress
	_ = cmd.Run()
	return nil
}
