package checks

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/northstar-group-demo/democtl/internal/scenario"
)

// CheckType represents the type of check command (verify or health)
type CheckType string

const (
	CheckTypeVerify CheckType = "verify"
	CheckTypeHealth CheckType = "health"
)

// CheckResult represents the outcome of a single check
type CheckResult struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Status      string `json:"status"` // "pass", "fail", "skip"
	Message     string `json:"message,omitempty"`
}

// RunResult represents the overall result of running checks
type RunResult struct {
	ScenarioID string        `json:"scenario_id"`
	Stage      string        `json:"stage"`
	CheckType  string        `json:"check_type"`
	Passed     int           `json:"passed"`
	Failed     int           `json:"failed"`
	Skipped    int           `json:"skipped"`
	Results    []CheckResult `json:"results"`
}

// RunOpts contains options for running checks
type RunOpts struct {
	Scenario      *scenario.Scenario
	CheckType     CheckType
	Stage         string
	OnlyFilter    string
	JSONOutput    bool
	Verbose       bool
	KubeContext   string
	Writer        io.Writer // for output
}

// Runner executes checks for a scenario
type Runner struct {
	opts RunOpts
}

// NewRunner creates a new checks runner
func NewRunner(opts RunOpts) *Runner {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}
	if opts.KubeContext == "" {
		opts.KubeContext = "kind-fider-demo"
	}
	return &Runner{opts: opts}
}

// Run executes the checks and returns the result
func (r *Runner) Run() (*RunResult, error) {
	// Determine stage
	stage := r.opts.Stage
	if stage == "" {
		var err error
		stage, err = r.getDefaultStage()
		if err != nil {
			return nil, err
		}
	}

	// Get checks for the stage
	checks, err := r.getChecksForStage(stage)
	if err != nil {
		return nil, err
	}

	if len(checks) == 0 {
		if !r.opts.JSONOutput {
			fmt.Fprintf(r.opts.Writer, "\033[0;34m[INFO]\033[0m No %s checks defined for stage '%s'\n", r.opts.CheckType, stage)
		}
		return &RunResult{
			ScenarioID: r.opts.Scenario.Identifier,
			Stage:      stage,
			CheckType:  string(r.opts.CheckType),
		}, nil
	}

	// Print header
	if !r.opts.JSONOutput {
		namespace := r.getNamespace()
		fmt.Fprintf(r.opts.Writer, "\033[0;34m[INFO]\033[0m Running %s checks for %s (stage: %s)\n", r.opts.CheckType, r.opts.Scenario.Identifier, stage)
		fmt.Fprintf(r.opts.Writer, "\033[0;34m[INFO]\033[0m Namespace: %s\n\n", namespace)
	}

	result := &RunResult{
		ScenarioID: r.opts.Scenario.Identifier,
		Stage:      stage,
		CheckType:  string(r.opts.CheckType),
		Results:    make([]CheckResult, 0),
	}

	// Run each check
	for _, check := range checks {
		// Apply filter if specified
		if !r.shouldRunCheck(check) {
			continue
		}

		checkResult := r.runCheck(check)
		result.Results = append(result.Results, checkResult)

		switch checkResult.Status {
		case "pass":
			result.Passed++
			if !r.opts.JSONOutput {
				fmt.Fprintf(r.opts.Writer, "\033[0;32m[PASS]\033[0m %s\n", checkResult.Description)
			}
		case "fail":
			result.Failed++
			if !r.opts.JSONOutput {
				msg := checkResult.Description
				if checkResult.Message != "" {
					msg = fmt.Sprintf("%s (%s)", msg, checkResult.Message)
				}
				fmt.Fprintf(os.Stderr, "\033[0;31m[FAIL]\033[0m %s\n", msg)
			}
			// For verify command, fail fast
			if r.opts.CheckType == CheckTypeVerify {
				if !r.opts.JSONOutput {
					fmt.Fprintln(os.Stderr, "\n\033[0;31m[FAIL]\033[0m Verification failed. Stopping.")
				}
				return result, fmt.Errorf("verification failed")
			}
		case "skip":
			result.Skipped++
			if !r.opts.JSONOutput {
				fmt.Fprintf(r.opts.Writer, "\033[0;33m[SKIP]\033[0m %s\n", checkResult.Description)
			}
		}
	}

	// Print summary
	if !r.opts.JSONOutput {
		fmt.Fprintf(r.opts.Writer, "\n\033[0;34m[INFO]\033[0m Summary: %d passed, %d failed, %d skipped\n", result.Passed, result.Failed, result.Skipped)
	}

	return result, nil
}

// getDefaultStage determines the default stage to use
// Priority: checks.default_stage > "broken" > "healthy" > lexicographically first
func (r *Runner) getDefaultStage() (string, error) {
	checks := r.opts.Scenario.Manifest.Checks
	
	// Check default_stage
	if checks.DefaultStage != "" {
		return checks.DefaultStage, nil
	}

	// Try preferred stages in order
	for _, candidate := range []string{"broken", "healthy"} {
		if _, exists := checks.Stages[candidate]; exists {
			return candidate, nil
		}
	}

	// Use lexicographically first stage
	if len(checks.Stages) > 0 {
		stages := make([]string, 0, len(checks.Stages))
		for stage := range checks.Stages {
			stages = append(stages, stage)
		}
		sort.Strings(stages)
		return stages[0], nil
	}

	return "", fmt.Errorf("no stages defined in scenario")
}

// getChecksForStage retrieves the checks for a specific stage and check type
func (r *Runner) getChecksForStage(stage string) ([]scenario.Check, error) {
	stageData, exists := r.opts.Scenario.Manifest.Checks.Stages[stage]
	if !exists {
		return nil, fmt.Errorf("stage '%s' not found in scenario", stage)
	}

	switch r.opts.CheckType {
	case CheckTypeVerify:
		return stageData.Verify, nil
	case CheckTypeHealth:
		return stageData.Health, nil
	default:
		return nil, fmt.Errorf("unknown check type: %s", r.opts.CheckType)
	}
}

// shouldRunCheck determines if a check should run based on filters
func (r *Runner) shouldRunCheck(check scenario.Check) bool {
	if r.opts.OnlyFilter == "" {
		return true
	}

	filter := r.opts.OnlyFilter
	checkType := check.Type

	switch filter {
	case "playwright":
		return checkType == "playwright.run"
	case "http":
		return len(checkType) >= 4 && checkType[:4] == "http"
	case "k8s":
		return len(checkType) >= 3 && checkType[:3] == "k8s"
	default:
		// Match prefix
		return len(checkType) >= len(filter) && checkType[:len(filter)] == filter
	}
}

// getNamespace returns the namespace for the scenario
func (r *Runner) getNamespace() string {
	return fmt.Sprintf("demo-%s", r.opts.Scenario.Manifest.Slug)
}

// runCheck executes a single check and returns its result
func (r *Runner) runCheck(check scenario.Check) CheckResult {
	r.logVerbose(fmt.Sprintf("Running check: %s (%s)", check.Description, check.Type))

	switch check.Type {
	case "http.get":
		return r.runHTTPGet(check)
	case "k8s.jqEquals":
		return r.runK8sJQEquals(check)
	case "k8s.podsContainLog":
		return r.runK8sPodsContainLog(check)
	case "k8s.podTerminationReason":
		return r.runK8sPodTerminationReason(check)
	case "k8s.podRestartCount":
		return r.runK8sPodRestartCount(check)
	case "k8s.deploymentAvailable":
		return r.runK8sDeploymentAvailable(check)
	case "k8s.resourceExists":
		return r.runK8sResourceExists(check)
	case "k8s.serviceMissingPort":
		return r.runK8sServiceMissingPort(check)
	case "playwright.run":
		return r.runPlaywright(check)
	default:
		return CheckResult{
			Type:        check.Type,
			Description: check.Description,
			Status:      "skip",
			Message:     fmt.Sprintf("Unknown check type: %s", check.Type),
		}
	}
}

// logVerbose prints a verbose log message if verbose mode is enabled
func (r *Runner) logVerbose(msg string) {
	if r.opts.Verbose && !r.opts.JSONOutput {
		fmt.Fprintf(r.opts.Writer, "       %s\n", msg)
	}
}

// MarshalJSON outputs the run result as JSON
func (r *RunResult) MarshalJSON() ([]byte, error) {
	type Alias RunResult
	return json.MarshalIndent((*Alias)(r), "", "  ")
}
