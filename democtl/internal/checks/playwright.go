package checks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/northstar-group-demo/democtl/internal/scenario"
)

// runPlaywright executes a Playwright test suite
func (r *Runner) runPlaywright(check scenario.Check) CheckResult {
	result := CheckResult{
		Type:        check.Type,
		Description: check.Description,
	}

	suite := check.Suite
	headed := check.Headed

	uiDir := filepath.Join(r.opts.Scenario.RepoRoot, "demo", "ui")
	
	// Check if UI directory exists
	if info, err := os.Stat(uiDir); err != nil || !info.IsDir() {
		result.Status = "skip"
		result.Message = fmt.Sprintf("UI test suite not found at %s", uiDir)
		return result
	}

	r.logVerbose(fmt.Sprintf("Running Playwright suite: %s", suite))

	// Get base URL
	baseURL := r.getBaseURL()
	
	// Get DEMO_LOGIN_KEY
	demoLoginKey := r.getDemoLoginKey()
	if demoLoginKey == "" {
		demoLoginKey = "northstar-demo-key"
	}

	// Determine stage (needed for Playwright env)
	stage := r.opts.Stage
	if stage == "" {
		var err error
		stage, err = r.getDefaultStage()
		if err != nil {
			stage = "broken" // fallback
		}
	}

	// Build headed flag
	headedFlag := ""
	if headed || os.Getenv("PLAYWRIGHT_HEADED") == "true" {
		headedFlag = "--headed"
	}

	// Build command args
	args := []string{"playwright", "test", "--grep", suite}
	if headedFlag != "" {
		args = append(args, headedFlag)
	}

	// Create command
	cmd := exec.Command("npx", args...)
	cmd.Dir = uiDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("BASE_URL=%s", baseURL),
		fmt.Sprintf("SCENARIO=%s", r.opts.Scenario.Identifier),
		fmt.Sprintf("STAGE=%s", stage),
		fmt.Sprintf("DEMO_LOGIN_KEY=%s", demoLoginKey),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run command
	if err := cmd.Run(); err != nil {
		result.Status = "fail"
		result.Message = "Playwright test failed"
		return result
	}

	result.Status = "pass"
	return result
}

// getBaseURL returns the base URL for the scenario
func (r *Runner) getBaseURL() string {
	urlHost := r.opts.Scenario.Manifest.URLHost
	if urlHost == "" {
		return ""
	}

	port := "8080" // default
	if r.opts.Scenario.Manifest.Type == scenario.TypeEngineering {
		port = "8082"
	}

	return fmt.Sprintf("http://%s:%s", urlHost, port)
}

// getDemoLoginKey retrieves the DEMO_LOGIN_KEY based on scenario type
func (r *Runner) getDemoLoginKey() string {
	switch r.opts.Scenario.Manifest.Type {
	case scenario.TypeSRE:
		// For SRE: read from ConfigMap
		namespace := r.getNamespace()
		cmd := exec.Command("kubectl", "--context="+r.opts.KubeContext, "-n", namespace,
			"get", "configmap", "fider-env", "-o", "jsonpath={.data.DEMO_LOGIN_KEY}")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
		return ""

	case scenario.TypeEngineering:
		// For Engineering: read from .state/global/secrets.env
		secretsFile := filepath.Join(r.opts.Scenario.RepoRoot, "demo", ".state", "global", "secrets.env")
		data, err := os.ReadFile(secretsFile)
		if err != nil {
			return ""
		}
		
		// Parse the file for DEMO_LOGIN_KEY
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "DEMO_LOGIN_KEY=") {
				return strings.TrimSpace(strings.TrimPrefix(line, "DEMO_LOGIN_KEY="))
			}
		}
		return ""

	default:
		return ""
	}
}
