package validate

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/northstar-group-demo/democtl/internal/scenario"
)

// ValidationError represents a validation error for a scenario
type ValidationError struct {
	ScenarioPath string
	Field        string
	Message      string
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s: %s", e.ScenarioPath, e.Field, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.ScenarioPath, e.Message)
}

// ValidationResult holds the results of validating scenarios
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []string
	Total    int
}

// HasErrors returns true if there are any validation errors
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// ValidateAll validates all discovered scenarios
func ValidateAll(repoRoot string, strict bool) (*ValidationResult, error) {
	scenarios, err := scenario.Discover(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to discover scenarios: %w", err)
	}

	result := &ValidationResult{
		Total: len(scenarios),
	}

	// Validate each scenario
	for _, s := range scenarios {
		if errs := ValidateScenario(s, repoRoot); len(errs) > 0 {
			result.Errors = append(result.Errors, errs...)
		}
	}

	// Check for collisions in strict mode
	if strict {
		collisions := scenario.DetectCollisions(scenarios)
		for identifier, types := range collisions {
			typeStrs := make([]string, len(types))
			for i, t := range types {
				typeStrs[i] = string(t)
			}
			result.Errors = append(result.Errors, ValidationError{
				ScenarioPath: identifier,
				Message:      fmt.Sprintf("collision detected: scenario exists in multiple types (%s)", strings.Join(typeStrs, ", ")),
			})
		}
	}

	return result, nil
}

// ValidateScenario validates a single scenario
func ValidateScenario(s *scenario.Scenario, repoRoot string) []ValidationError {
	var errors []ValidationError
	relPath := getRelativePath(s.Dir, repoRoot)

	// Validate required fields
	if s.Manifest.Track == "" {
		errors = append(errors, ValidationError{
			ScenarioPath: relPath,
			Field:        "track",
			Message:      "missing required field",
		})
	}

	if s.Manifest.Slug == "" {
		errors = append(errors, ValidationError{
			ScenarioPath: relPath,
			Field:        "slug",
			Message:      "missing required field",
		})
	}

	if s.Manifest.Title == "" {
		errors = append(errors, ValidationError{
			ScenarioPath: relPath,
			Field:        "title",
			Message:      "missing required field",
		})
	}

	if s.Manifest.Type == "" {
		errors = append(errors, ValidationError{
			ScenarioPath: relPath,
			Field:        "type",
			Message:      "missing required field",
		})
	}

	if s.Manifest.URLHost == "" {
		errors = append(errors, ValidationError{
			ScenarioPath: relPath,
			Field:        "url_host",
			Message:      "missing required field",
		})
	}

	if s.Manifest.ResetStrategy == "" {
		errors = append(errors, ValidationError{
			ScenarioPath: relPath,
			Field:        "reset_strategy",
			Message:      "missing required field",
		})
	}

	// Validate type matches directory
	if strings.HasPrefix(relPath, "sre/") && s.Manifest.Type != scenario.TypeSRE {
		errors = append(errors, ValidationError{
			ScenarioPath: relPath,
			Field:        "type",
			Message:      fmt.Sprintf("scenario in sre/ directory but type is '%s' (expected 'sre')", s.Manifest.Type),
		})
	} else if strings.HasPrefix(relPath, "engineering/") && s.Manifest.Type != scenario.TypeEngineering {
		errors = append(errors, ValidationError{
			ScenarioPath: relPath,
			Field:        "type",
			Message:      fmt.Sprintf("scenario in engineering/ directory but type is '%s' (expected 'engineering')", s.Manifest.Type),
		})
	}

	// Validate path depth (should be <type>/scenarios/<track>/<slug>)
	depth := strings.Count(relPath, "/")
	if depth != 3 {
		errors = append(errors, ValidationError{
			ScenarioPath: relPath,
			Message:      fmt.Sprintf("scenario path must be exactly 4 levels deep: <type>/scenarios/<track>/<slug>, got depth %d", depth+1),
		})
	}

	// Validate checks schema
	if checksErrs := validateChecks(s, relPath); len(checksErrs) > 0 {
		errors = append(errors, checksErrs...)
	}
	
	// Validate engineering-specific git config
	if s.Manifest.Type == scenario.TypeEngineering {
		if gitErrs := validateEngineeringGitConfig(s, relPath); len(gitErrs) > 0 {
			errors = append(errors, gitErrs...)
		}
	}

	return errors
}

// validateChecks validates the checks schema in a scenario
func validateChecks(s *scenario.Scenario, relPath string) []ValidationError {
	var errors []ValidationError

	// If checks.version is 0, it means checks weren't provided or version wasn't set
	// We need to check if stages exist to determine if checks were actually provided
	if len(s.Manifest.Checks.Stages) == 0 {
		// No checks provided, which is valid
		return nil
	}

	// Checks are provided, so version must be 1
	if s.Manifest.Checks.Version != 1 {
		errors = append(errors, ValidationError{
			ScenarioPath: relPath,
			Field:        "checks.version",
			Message:      fmt.Sprintf("must be 1, got %d", s.Manifest.Checks.Version),
		})
	}

	// Validate each stage
	for stageName, stage := range s.Manifest.Checks.Stages {
		// Each stage must have a verify array (can be empty)
		if stage.Verify == nil {
			errors = append(errors, ValidationError{
				ScenarioPath: relPath,
				Field:        fmt.Sprintf("checks.stages.%s.verify", stageName),
				Message:      "must be an array",
			})
			continue
		}

		// Validate each check has required fields
		for i, check := range stage.Verify {
			if check.Type == "" {
				errors = append(errors, ValidationError{
					ScenarioPath: relPath,
					Field:        fmt.Sprintf("checks.stages.%s.verify[%d].type", stageName, i),
					Message:      "missing required field",
				})
			}
		}

		// Also validate health checks if present
		if stage.Health != nil {
			for i, check := range stage.Health {
				if check.Type == "" {
					errors = append(errors, ValidationError{
						ScenarioPath: relPath,
						Field:        fmt.Sprintf("checks.stages.%s.health[%d].type", stageName, i),
						Message:      "missing required field",
					})
				}
			}
		}
	}

	return errors
}

// validateEngineeringGitConfig validates git config for engineering scenarios
func validateEngineeringGitConfig(s *scenario.Scenario, relPath string) []ValidationError {
	var errors []ValidationError
	
	if s.Manifest.Git == nil {
		errors = append(errors, ValidationError{
			ScenarioPath: relPath,
			Field:        "git",
			Message:      "git config required for engineering scenarios",
		})
		return errors
	}
	
	// Validate base_ref is present
	if s.Manifest.Git.BaseRef == "" {
		errors = append(errors, ValidationError{
			ScenarioPath: relPath,
			Field:        "git.base_ref",
			Message:      "missing required field",
		})
	} else {
		// Validate base_ref is a full 40-character SHA
		if len(s.Manifest.Git.BaseRef) != 40 {
			errors = append(errors, ValidationError{
				ScenarioPath: relPath,
				Field:        "git.base_ref",
				Message:      fmt.Sprintf("must be a full 40-character commit SHA, got %d characters", len(s.Manifest.Git.BaseRef)),
			})
		} else {
			// Validate it's hexadecimal
			for _, c := range s.Manifest.Git.BaseRef {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
					errors = append(errors, ValidationError{
						ScenarioPath: relPath,
						Field:        "git.base_ref",
						Message:      "must be a valid hexadecimal SHA",
					})
					break
				}
			}
		}
	}
	
	// Validate work_branch is present
	if s.Manifest.Git.WorkBranch == "" {
		errors = append(errors, ValidationError{
			ScenarioPath: relPath,
			Field:        "git.work_branch",
			Message:      "missing required field",
		})
	}
	
	// Validate patch dir strings are safe relative paths (no absolute paths, no .. traversal)
	if s.Manifest.Git.BrokenPatchesDir != "" {
		if err := validatePatchDirPath(s.Manifest.Git.BrokenPatchesDir); err != nil {
			errors = append(errors, ValidationError{
				ScenarioPath: relPath,
				Field:        "git.broken_patches_dir",
				Message:      err.Error(),
			})
		}
	}
	
	if s.Manifest.Git.SolvedPatchesDir != "" {
		if err := validatePatchDirPath(s.Manifest.Git.SolvedPatchesDir); err != nil {
			errors = append(errors, ValidationError{
				ScenarioPath: relPath,
				Field:        "git.solved_patches_dir",
				Message:      err.Error(),
			})
		}
	}
	
	return errors
}

// validatePatchDirPath validates that a patch directory path is safe
func validatePatchDirPath(path string) error {
	// Must not be absolute
	if filepath.IsAbs(path) {
		return fmt.Errorf("must be a relative path, not absolute")
	}
	
	// Must not contain ..
	if strings.Contains(path, "..") {
		return fmt.Errorf("must not contain '..' (parent directory references)")
	}
	
	// Clean the path and ensure it doesn't try to escape
	cleaned := filepath.Clean(path)
	if strings.HasPrefix(cleaned, "..") {
		return fmt.Errorf("path would escape scenario directory")
	}
	
	return nil
}

// getRelativePath returns the path relative to the demo directory
func getRelativePath(scenarioDir, repoRoot string) string {
	demoDir := filepath.Join(repoRoot, "demo")
	relPath, err := filepath.Rel(demoDir, scenarioDir)
	if err != nil {
		return scenarioDir
	}
	return relPath
}
