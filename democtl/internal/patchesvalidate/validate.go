package patchesvalidate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/northstar-group-demo/democtl/internal/scenario"
)

// ValidationError represents a patch validation error
type ValidationError struct {
	ScenarioID string
	Stage      string
	PatchFile  string
	Message    string
}

func (e ValidationError) Error() string {
	if e.PatchFile != "" {
		return fmt.Sprintf("%s (%s stage, patch %s): %s", e.ScenarioID, e.Stage, e.PatchFile, e.Message)
	}
	return fmt.Sprintf("%s (%s stage): %s", e.ScenarioID, e.Stage, e.Message)
}

// ValidationResult holds the results of patch validation
type ValidationResult struct {
	Errors []ValidationError
	Total  int
}

// HasErrors returns true if there are any validation errors
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// ValidateAll validates patches for all engineering scenarios
func ValidateAll(repoRoot string, scenarioID string, strict bool) (*ValidationResult, error) {
	scenarios, err := scenario.Discover(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to discover scenarios: %w", err)
	}

	result := &ValidationResult{}

	// Filter to specific scenario if requested
	if scenarioID != "" {
		found := false
		for _, s := range scenarios {
			if s.Identifier == scenarioID && s.Manifest.Type == scenario.TypeEngineering {
				result.Total = 1
				if errs := ValidateScenarioPatches(repoRoot, s, strict); len(errs) > 0 {
					result.Errors = append(result.Errors, errs...)
				}
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("engineering scenario not found: %s", scenarioID)
		}
	} else {
		// Validate all engineering scenarios
		for _, s := range scenarios {
			if s.Manifest.Type != scenario.TypeEngineering {
				continue
			}
			result.Total++
			if errs := ValidateScenarioPatches(repoRoot, s, strict); len(errs) > 0 {
				result.Errors = append(result.Errors, errs...)
			}
		}
	}

	return result, nil
}

// ValidateScenarioPatches validates patches for a single engineering scenario
func ValidateScenarioPatches(repoRoot string, s *scenario.Scenario, strict bool) []ValidationError {
	var errors []ValidationError

	if s.Manifest.Git == nil {
		errors = append(errors, ValidationError{
			ScenarioID: s.Identifier,
			Message:    "missing git config",
		})
		return errors
	}

	baseRef := s.Manifest.Git.BaseRef
	if baseRef == "" {
		errors = append(errors, ValidationError{
			ScenarioID: s.Identifier,
			Message:    "missing base_ref",
		})
		return errors
	}

	// Validate base commit exists
	if err := validateBaseCommitExists(repoRoot, baseRef); err != nil {
		errors = append(errors, ValidationError{
			ScenarioID: s.Identifier,
			Message:    fmt.Sprintf("base commit validation failed: %v", err),
		})
		return errors
	}

	// Optional: check if base_ref is an ancestor of HEAD
	if strict {
		if err := validateBaseIsAncestor(repoRoot, baseRef); err != nil {
			errors = append(errors, ValidationError{
				ScenarioID: s.Identifier,
				Message:    fmt.Sprintf("base commit is not an ancestor of HEAD: %v", err),
			})
		}
	}

	// Validate broken stage patches
	brokenPatchesDir := s.Manifest.Git.GetBrokenPatchesDir()
	if brokenErrs := validateStagePatches(repoRoot, s, baseRef, brokenPatchesDir, "broken"); len(brokenErrs) > 0 {
		errors = append(errors, brokenErrs...)
	}

	// Validate solved stage patches
	solvedPatchesDir := s.Manifest.Git.GetSolvedPatchesDir()
	if solvedErrs := validateStagePatches(repoRoot, s, baseRef, solvedPatchesDir, "solved"); len(solvedErrs) > 0 {
		errors = append(errors, solvedErrs...)
	}

	return errors
}

// validateBaseCommitExists checks if the base commit exists in the local repo
func validateBaseCommitExists(repoRoot, baseRef string) error {
	cmd := exec.Command("git", "cat-file", "-e", baseRef+"^{commit}")
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("base commit %s not found", baseRef)
	}
	return nil
}

// validateBaseIsAncestor checks if base commit is an ancestor of HEAD
func validateBaseIsAncestor(repoRoot, baseRef string) error {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", baseRef, "HEAD")
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("base commit is not reachable from HEAD")
	}
	return nil
}

// validateStagePatches validates patches for a specific stage
func validateStagePatches(repoRoot string, s *scenario.Scenario, baseRef, patchesRelDir, stage string) []ValidationError {
	var errors []ValidationError

	patchesDir := filepath.Join(s.Dir, patchesRelDir)

	// Check if patches directory exists
	info, err := os.Stat(patchesDir)
	if os.IsNotExist(err) {
		// No patches directory is OK - means empty patch series
		return nil
	}
	if err != nil {
		errors = append(errors, ValidationError{
			ScenarioID: s.Identifier,
			Stage:      stage,
			Message:    fmt.Sprintf("failed to stat patches directory: %v", err),
		})
		return errors
	}
	if !info.IsDir() {
		errors = append(errors, ValidationError{
			ScenarioID: s.Identifier,
			Stage:      stage,
			Message:    "patches path exists but is not a directory",
		})
		return errors
	}

	// List patch files
	patches, err := listPatchFiles(patchesDir)
	if err != nil {
		errors = append(errors, ValidationError{
			ScenarioID: s.Identifier,
			Stage:      stage,
			Message:    fmt.Sprintf("failed to list patches: %v", err),
		})
		return errors
	}

	// If no patches, that's OK
	if len(patches) == 0 {
		return nil
	}

	// Validate scope of each patch
	for _, patchFile := range patches {
		if err := validatePatchScope(patchFile); err != nil {
			errors = append(errors, ValidationError{
				ScenarioID: s.Identifier,
				Stage:      stage,
				PatchFile:  filepath.Base(patchFile),
				Message:    fmt.Sprintf("invalid patch scope: %v", err),
			})
		}
	}

	// Validate patches apply cleanly
	if err := validatePatchesApply(repoRoot, baseRef, patches); err != nil {
		errors = append(errors, ValidationError{
			ScenarioID: s.Identifier,
			Stage:      stage,
			Message:    fmt.Sprintf("patches do not apply cleanly: %v", err),
		})
	}

	return errors
}

// listPatchFiles returns sorted list of .patch files in directory
func listPatchFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var patches []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".patch") {
			patches = append(patches, filepath.Join(dir, entry.Name()))
		}
	}

	return patches, nil
}

// validatePatchScope validates that a patch only touches files under fider/
func validatePatchScope(patchFile string) error {
	// Use git apply --numstat to get the list of affected files
	cmd := exec.Command("git", "apply", "--numstat", patchFile)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to analyze patch: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format: <added> <deleted> <path>
		// Or for binary files: - - <path>
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		path := parts[2]

		// Handle renames: "old => new" format
		if strings.Contains(path, " => ") {
			// Extract both old and new paths
			renameParts := strings.Split(path, " => ")
			if len(renameParts) == 2 {
				oldPath := strings.TrimSpace(renameParts[0])
				newPath := strings.TrimSpace(renameParts[1])

				// Check both paths
				if oldPath != "/dev/null" && !strings.HasPrefix(oldPath, "fider/") {
					return fmt.Errorf("patch affects path outside fider/: %s", oldPath)
				}
				if newPath != "/dev/null" && !strings.HasPrefix(newPath, "fider/") {
					return fmt.Errorf("patch affects path outside fider/: %s", newPath)
				}
				continue
			}
		}

		// Handle /dev/null (deletions/additions)
		if path == "/dev/null" {
			continue
		}

		// Check if path is under fider/
		if !strings.HasPrefix(path, "fider/") {
			return fmt.Errorf("patch affects path outside fider/: %s", path)
		}
	}

	return nil
}

// validatePatchesApply validates that patches apply cleanly to base ref
func validatePatchesApply(repoRoot, baseRef string, patches []string) error {
	if len(patches) == 0 {
		return nil
	}

	// Create a temporary worktree
	tmpDir, err := os.MkdirTemp("", "democtl-patch-validate-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	worktreeDir := filepath.Join(tmpDir, "worktree")

	// Create worktree at base ref
	cmd := exec.Command("git", "worktree", "add", worktreeDir, baseRef)
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create temporary worktree: %w", err)
	}

	// Ensure cleanup
	defer func() {
		cmd := exec.Command("git", "worktree", "remove", "--force", worktreeDir)
		cmd.Dir = repoRoot
		_ = cmd.Run()

		cmd = exec.Command("git", "worktree", "prune")
		cmd.Dir = repoRoot
		_ = cmd.Run()
	}()

	// Apply patches
	for _, patchFile := range patches {
		cmd := exec.Command("git", "am", "--keep-cr", "--quiet", patchFile)
		cmd.Dir = worktreeDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Abort the am session
			abortCmd := exec.Command("git", "am", "--abort")
			abortCmd.Dir = worktreeDir
			_ = abortCmd.Run()

			return fmt.Errorf("patch %s failed to apply: %w\nOutput: %s", filepath.Base(patchFile), err, string(output))
		}
	}

	return nil
}
