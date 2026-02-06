package migrate

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/northstar-group-demo/democtl/internal/scenario"
)

// MigrateScenarioToPatches converts a tag-based scenario to patch-based workflow
func MigrateScenarioToPatches(repoRoot string, s *scenario.Scenario, baseRef string) error {
	// Validate scenario has legacy fields
	if s.Manifest.Git == nil {
		return fmt.Errorf("scenario %s has no git config", s.Identifier)
	}

	brokenRef := s.Manifest.Git.BrokenRef
	solvedRef := s.Manifest.Git.SolvedRef

	// For migration, we expect legacy fields
	if brokenRef == "" || solvedRef == "" {
		// Check if already migrated (has base_ref instead)
		if s.Manifest.Git.BaseRef != "" {
			fmt.Printf("Scenario %s appears already migrated (has base_ref)\n", s.Identifier)
			return nil
		}
		return fmt.Errorf("scenario %s missing broken_ref or solved_ref", s.Identifier)
	}

	fmt.Printf("\nMigrating scenario: %s\n", s.Identifier)
	fmt.Printf("  Broken ref: %s\n", brokenRef)
	fmt.Printf("  Solved ref: %s\n", solvedRef)

	// Compute base SHA if not provided
	if baseRef == "" {
		// Default: base = parent of broken_ref
		cmd := exec.Command("git", "rev-parse", brokenRef+"^")
		cmd.Dir = repoRoot
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to compute base ref from %s^: %w", brokenRef, err)
		}
		baseRef = strings.TrimSpace(string(output))
		fmt.Printf("  Computed base ref: %s (parent of broken ref)\n", baseRef)
	} else {
		// Validate provided base exists
		cmd := exec.Command("git", "cat-file", "-e", baseRef+"^{commit}")
		cmd.Dir = repoRoot
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("provided base ref %s not found", baseRef)
		}
		fmt.Printf("  Using provided base ref: %s\n", baseRef)
	}

	// Export broken patches
	brokenPatchesDir := filepath.Join(s.Dir, "patches", "broken")
	if err := exportPatches(repoRoot, baseRef, brokenRef, brokenPatchesDir); err != nil {
		return fmt.Errorf("failed to export broken patches: %w", err)
	}

	// Export solved patches
	solvedPatchesDir := filepath.Join(s.Dir, "patches", "solved")
	if err := exportPatches(repoRoot, baseRef, solvedRef, solvedPatchesDir); err != nil {
		return fmt.Errorf("failed to export solved patches: %w", err)
	}

	// Validate scope of all patches
	fmt.Println("  Validating patch scope...")
	if err := validateAllPatchScope(brokenPatchesDir); err != nil {
		return fmt.Errorf("broken patches violate scope: %w", err)
	}
	if err := validateAllPatchScope(solvedPatchesDir); err != nil {
		return fmt.Errorf("solved patches violate scope: %w", err)
	}
	fmt.Println("  ✓ All patches only touch fider/")

	// Update scenario.json
	if err := updateManifest(s, baseRef); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	fmt.Printf("✓ Migration complete for %s\n", s.Identifier)
	return nil
}

// exportPatches exports patches from base to target into outputDir
func exportPatches(repoRoot, baseRef, targetRef, outputDir string) error {
	// Remove existing patches if any
	if err := os.RemoveAll(outputDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing patches directory: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create patches directory: %w", err)
	}

	// Export patches using git format-patch
	cmd := exec.Command("git", "format-patch", "--binary", "--output-directory", outputDir, baseRef+".."+targetRef)
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git format-patch failed: %w", err)
	}

	patchCount := len(strings.Split(strings.TrimSpace(string(output)), "\n"))
	if string(output) == "" {
		patchCount = 0
	}

	fmt.Printf("  Exported %d patch(es) to %s\n", patchCount, filepath.Base(outputDir))
	return nil
}

// validateAllPatchScope validates that all patches in a directory only touch fider/
func validateAllPatchScope(patchesDir string) error {
	entries, err := os.ReadDir(patchesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No patches is OK
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".patch") {
			continue
		}

		patchFile := filepath.Join(patchesDir, entry.Name())
		if err := validatePatchScope(patchFile); err != nil {
			return fmt.Errorf("patch %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// validatePatchScope validates that a patch only touches files under fider/
func validatePatchScope(patchFile string) error {
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
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		path := parts[2]

		// Handle renames
		if strings.Contains(path, " => ") {
			renameParts := strings.Split(path, " => ")
			if len(renameParts) == 2 {
				oldPath := strings.TrimSpace(renameParts[0])
				newPath := strings.TrimSpace(renameParts[1])

				if oldPath != "/dev/null" && !strings.HasPrefix(oldPath, "fider/") {
					return fmt.Errorf("patch affects path outside fider/: %s", oldPath)
				}
				if newPath != "/dev/null" && !strings.HasPrefix(newPath, "fider/") {
					return fmt.Errorf("patch affects path outside fider/: %s", newPath)
				}
				continue
			}
		}

		if path == "/dev/null" {
			continue
		}

		if !strings.HasPrefix(path, "fider/") {
			return fmt.Errorf("patch affects path outside fider/: %s", path)
		}
	}

	return nil
}

// updateManifest updates the scenario manifest to use patch-based git config
func updateManifest(s *scenario.Scenario, baseRef string) error {
	manifestPath := filepath.Join(s.Dir, "scenario.json")

	// Read current manifest
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	// Parse as generic map to preserve formatting and unknown fields
	var manifestMap map[string]interface{}
	if err := json.Unmarshal(data, &manifestMap); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Update git config
	gitConfig, ok := manifestMap["git"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("git config is not an object")
	}

	// Remove legacy fields
	delete(gitConfig, "maintenance_branch")
	delete(gitConfig, "broken_ref")
	delete(gitConfig, "solved_ref")

	// Add new fields
	gitConfig["base_ref"] = baseRef
	gitConfig["broken_patches_dir"] = "patches/broken"
	gitConfig["solved_patches_dir"] = "patches/solved"

	// Update checks stages if using "fixed" -> rename to "solved"
	if checks, ok := manifestMap["checks"].(map[string]interface{}); ok {
		if stages, ok := checks["stages"].(map[string]interface{}); ok {
			if fixed, ok := stages["fixed"]; ok {
				stages["solved"] = fixed
				delete(stages, "fixed")
				fmt.Println("  Renamed checks stage: fixed -> solved")
			}
		}
	}

	// Write updated manifest
	updatedData, err := json.MarshalIndent(manifestMap, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Add trailing newline
	updatedData = append(updatedData, '\n')

	if err := os.WriteFile(manifestPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	fmt.Println("  Updated scenario.json")
	return nil
}
