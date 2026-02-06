package checks

import (
	"bytes"
	"testing"

	"github.com/northstar-group-demo/democtl/internal/scenario"
)

func TestStageSelection(t *testing.T) {
	tests := []struct {
		name          string
		manifest      scenario.Manifest
		expectedStage string
		expectError   bool
	}{
		{
			name: "Uses default_stage when set",
			manifest: scenario.Manifest{
				Checks: scenario.Checks{
					DefaultStage: "healthy",
					Stages: map[string]scenario.Stage{
						"broken":  {},
						"healthy": {},
					},
				},
			},
			expectedStage: "healthy",
		},
		{
			name: "Prefers broken over healthy",
			manifest: scenario.Manifest{
				Checks: scenario.Checks{
					Stages: map[string]scenario.Stage{
						"broken":  {},
						"healthy": {},
					},
				},
			},
			expectedStage: "broken",
		},
		{
			name: "Uses healthy when broken not present",
			manifest: scenario.Manifest{
				Checks: scenario.Checks{
					Stages: map[string]scenario.Stage{
						"healthy": {},
						"fixed":   {},
					},
				},
			},
			expectedStage: "healthy",
		},
		{
			name: "Uses lexicographically first when neither broken nor healthy",
			manifest: scenario.Manifest{
				Checks: scenario.Checks{
					Stages: map[string]scenario.Stage{
						"zzz": {},
						"aaa": {},
						"mmm": {},
					},
				},
			},
			expectedStage: "aaa",
		},
		{
			name: "Errors when no stages defined",
			manifest: scenario.Manifest{
				Checks: scenario.Checks{
					Stages: map[string]scenario.Stage{},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &scenario.Scenario{
				Manifest: tt.manifest,
			}

			runner := NewRunner(RunOpts{
				Scenario:  s,
				CheckType: CheckTypeVerify,
			})

			stage, err := runner.getDefaultStage()
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if stage != tt.expectedStage {
				t.Errorf("Expected stage %q, got %q", tt.expectedStage, stage)
			}
		})
	}
}

func TestFilterLogic(t *testing.T) {
	runner := NewRunner(RunOpts{})

	tests := []struct {
		name       string
		checkType  string
		filter     string
		shouldRun  bool
	}{
		{"No filter runs all", "http.get", "", true},
		{"Playwright filter matches playwright.run", "playwright.run", "playwright", true},
		{"Playwright filter rejects http.get", "http.get", "playwright", false},
		{"HTTP filter matches http.get", "http.get", "http", true},
		{"HTTP filter rejects k8s.jqEquals", "k8s.jqEquals", "http", false},
		{"K8s filter matches k8s.jqEquals", "k8s.jqEquals", "k8s", true},
		{"K8s filter matches k8s.podsContainLog", "k8s.podsContainLog", "k8s", true},
		{"K8s filter rejects http.get", "http.get", "k8s", false},
		{"Exact prefix match", "k8s.deploymentAvailable", "k8s.deployment", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner.opts.OnlyFilter = tt.filter
			check := scenario.Check{Type: tt.checkType}
			result := runner.shouldRunCheck(check)
			if result != tt.shouldRun {
				t.Errorf("For check %q with filter %q: expected shouldRun=%v, got %v",
					tt.checkType, tt.filter, tt.shouldRun, result)
			}
		})
	}
}

func TestCheckResultOutput(t *testing.T) {
	// Test that check results are properly formatted
	s := &scenario.Scenario{
		Identifier: "platform/test",
		Manifest: scenario.Manifest{
			Slug: "test",
			Checks: scenario.Checks{
				DefaultStage: "test-stage",
				Stages: map[string]scenario.Stage{
					"test-stage": {
						Verify: []scenario.Check{
							{
								Type:        "unknown.check",
								Description: "This check type doesn't exist",
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	runner := NewRunner(RunOpts{
		Scenario:  s,
		CheckType: CheckTypeVerify,
		Writer:    &buf,
	})

	result, _ := runner.Run()

	if result.Skipped != 1 {
		t.Errorf("Expected 1 skipped check, got %d", result.Skipped)
	}

	if len(result.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result.Results))
	}

	if result.Results[0].Status != "skip" {
		t.Errorf("Expected status 'skip', got %q", result.Results[0].Status)
	}
}

func TestNamespaceGeneration(t *testing.T) {
	s := &scenario.Scenario{
		Manifest: scenario.Manifest{
			Slug: "my-scenario",
		},
	}

	runner := NewRunner(RunOpts{Scenario: s})
	ns := runner.getNamespace()

	expected := "demo-my-scenario"
	if ns != expected {
		t.Errorf("Expected namespace %q, got %q", expected, ns)
	}
}
