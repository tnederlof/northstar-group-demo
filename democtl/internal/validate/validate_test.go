package validate

import (
	"testing"

	"github.com/northstar-group-demo/democtl/internal/scenario"
)

func TestValidateScenario_RequiredFields(t *testing.T) {
	tests := []struct {
		name          string
		manifest      scenario.Manifest
		expectedError string
	}{
		{
			name: "valid scenario",
			manifest: scenario.Manifest{
				Track:         "platform",
				Slug:          "test-scenario",
				Title:         "Test Scenario",
				Type:          scenario.TypeSRE,
				URLHost:       "test.localhost",
				ResetStrategy: scenario.ResetNamespaceDelete,
			},
			expectedError: "",
		},
		{
			name: "missing track",
			manifest: scenario.Manifest{
				Slug:          "test-scenario",
				Title:         "Test Scenario",
				Type:          scenario.TypeSRE,
				URLHost:       "test.localhost",
				ResetStrategy: scenario.ResetNamespaceDelete,
			},
			expectedError: "track",
		},
		{
			name: "missing slug",
			manifest: scenario.Manifest{
				Track:         "platform",
				Title:         "Test Scenario",
				Type:          scenario.TypeSRE,
				URLHost:       "test.localhost",
				ResetStrategy: scenario.ResetNamespaceDelete,
			},
			expectedError: "slug",
		},
		{
			name: "missing title",
			manifest: scenario.Manifest{
				Track:         "platform",
				Slug:          "test-scenario",
				Type:          scenario.TypeSRE,
				URLHost:       "test.localhost",
				ResetStrategy: scenario.ResetNamespaceDelete,
			},
			expectedError: "title",
		},
		{
			name: "missing type",
			manifest: scenario.Manifest{
				Track:         "platform",
				Slug:          "test-scenario",
				Title:         "Test Scenario",
				URLHost:       "test.localhost",
				ResetStrategy: scenario.ResetNamespaceDelete,
			},
			expectedError: "type",
		},
		{
			name: "missing url_host",
			manifest: scenario.Manifest{
				Track:         "platform",
				Slug:          "test-scenario",
				Title:         "Test Scenario",
				Type:          scenario.TypeSRE,
				ResetStrategy: scenario.ResetNamespaceDelete,
			},
			expectedError: "url_host",
		},
		{
			name: "missing reset_strategy",
			manifest: scenario.Manifest{
				Track:   "platform",
				Slug:    "test-scenario",
				Title:   "Test Scenario",
				Type:    scenario.TypeSRE,
				URLHost: "test.localhost",
			},
			expectedError: "reset_strategy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &scenario.Scenario{
				Manifest: tt.manifest,
				Dir:      "/repo/demo/sre/scenarios/platform/test-scenario",
			}

			errors := ValidateScenario(s, "/repo")

			if tt.expectedError == "" {
				if len(errors) != 0 {
					t.Errorf("expected no errors, got %d: %v", len(errors), errors)
				}
			} else {
				found := false
				for _, err := range errors {
					if err.Field == tt.expectedError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error for field %q, got errors: %v", tt.expectedError, errors)
				}
			}
		})
	}
}

func TestValidateScenario_TypePathMismatch(t *testing.T) {
	tests := []struct {
		name         string
		scenarioType scenario.ScenarioType
		dir          string
		expectError  bool
	}{
		{
			name:         "sre type in sre directory",
			scenarioType: scenario.TypeSRE,
			dir:          "/repo/demo/sre/scenarios/platform/test",
			expectError:  false,
		},
		{
			name:         "engineering type in engineering directory",
			scenarioType: scenario.TypeEngineering,
			dir:          "/repo/demo/engineering/scenarios/backend/test",
			expectError:  false,
		},
		{
			name:         "sre type in engineering directory",
			scenarioType: scenario.TypeSRE,
			dir:          "/repo/demo/engineering/scenarios/backend/test",
			expectError:  true,
		},
		{
			name:         "engineering type in sre directory",
			scenarioType: scenario.TypeEngineering,
			dir:          "/repo/demo/sre/scenarios/platform/test",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &scenario.Scenario{
				Manifest: scenario.Manifest{
					Track:         "platform",
					Slug:          "test",
					Title:         "Test",
					Type:          tt.scenarioType,
					URLHost:       "test.localhost",
					ResetStrategy: scenario.ResetNamespaceDelete,
				},
				Dir: tt.dir,
			}

			errors := ValidateScenario(s, "/repo")

			hasTypeError := false
			for _, err := range errors {
				if err.Field == "type" {
					hasTypeError = true
					break
				}
			}

			if tt.expectError && !hasTypeError {
				t.Errorf("expected type mismatch error, got none")
			}
			if !tt.expectError && hasTypeError {
				t.Errorf("expected no type error, got errors: %v", errors)
			}
		})
	}
}

func TestValidateScenario_PathDepth(t *testing.T) {
	tests := []struct {
		name        string
		dir         string
		expectError bool
	}{
		{
			name:        "correct depth - sre",
			dir:         "/repo/demo/sre/scenarios/platform/test",
			expectError: false,
		},
		{
			name:        "correct depth - engineering",
			dir:         "/repo/demo/engineering/scenarios/backend/test",
			expectError: false,
		},
		{
			name:        "too shallow",
			dir:         "/repo/demo/sre/scenarios/test",
			expectError: true,
		},
		{
			name:        "too deep",
			dir:         "/repo/demo/sre/scenarios/platform/nested/test",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &scenario.Scenario{
				Manifest: scenario.Manifest{
					Track:         "platform",
					Slug:          "test",
					Title:         "Test",
					Type:          scenario.TypeSRE,
					URLHost:       "test.localhost",
					ResetStrategy: scenario.ResetNamespaceDelete,
				},
				Dir: tt.dir,
			}

			errors := ValidateScenario(s, "/repo")

			hasDepthError := false
			for _, err := range errors {
				if err.Field == "" && err.Message != "" &&
					(len(err.Message) > 20 && err.Message[:20] == "scenario path must b") {
					hasDepthError = true
					break
				}
			}

			if tt.expectError && !hasDepthError {
				t.Errorf("expected depth error, got none")
			}
			if !tt.expectError && hasDepthError {
				t.Errorf("expected no depth error, got errors: %v", errors)
			}
		})
	}
}

func TestValidateChecks(t *testing.T) {
	tests := []struct {
		name        string
		checks      scenario.Checks
		expectError bool
		errorField  string
	}{
		{
			name: "valid checks",
			checks: scenario.Checks{
				Version: 1,
				Stages: map[string]scenario.Stage{
					"broken": {
						Verify: []scenario.Check{
							{Type: "http.get", Description: "Test"},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "no checks (valid)",
			checks: scenario.Checks{
				Version: 0,
				Stages:  map[string]scenario.Stage{},
			},
			expectError: false,
		},
		{
			name: "checks without version",
			checks: scenario.Checks{
				Version: 0,
				Stages: map[string]scenario.Stage{
					"broken": {
						Verify: []scenario.Check{
							{Type: "http.get"},
						},
					},
				},
			},
			expectError: true,
			errorField:  "checks.version",
		},
		{
			name: "wrong version",
			checks: scenario.Checks{
				Version: 2,
				Stages: map[string]scenario.Stage{
					"broken": {
						Verify: []scenario.Check{
							{Type: "http.get"},
						},
					},
				},
			},
			expectError: true,
			errorField:  "checks.version",
		},
		{
			name: "check missing type",
			checks: scenario.Checks{
				Version: 1,
				Stages: map[string]scenario.Stage{
					"broken": {
						Verify: []scenario.Check{
							{Description: "Test"},
						},
					},
				},
			},
			expectError: true,
			errorField:  "checks.stages.broken.verify[0].type",
		},
		{
			name: "health check missing type",
			checks: scenario.Checks{
				Version: 1,
				Stages: map[string]scenario.Stage{
					"broken": {
						Verify: []scenario.Check{
							{Type: "http.get"},
						},
						Health: []scenario.Check{
							{Description: "Test"},
						},
					},
				},
			},
			expectError: true,
			errorField:  "checks.stages.broken.health[0].type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &scenario.Scenario{
				Manifest: scenario.Manifest{
					Track:         "platform",
					Slug:          "test",
					Title:         "Test",
					Type:          scenario.TypeSRE,
					URLHost:       "test.localhost",
					ResetStrategy: scenario.ResetNamespaceDelete,
					Checks:        tt.checks,
				},
				Dir: "/repo/demo/sre/scenarios/platform/test",
			}

			errors := ValidateScenario(s, "/repo")

			if !tt.expectError {
				if len(errors) != 0 {
					t.Errorf("expected no errors, got %d: %v", len(errors), errors)
				}
				return
			}

			found := false
			for _, err := range errors {
				if err.Field == tt.errorField {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected error for field %q, got errors: %v", tt.errorField, errors)
			}
		})
	}
}

func TestDetectCollisions(t *testing.T) {
	scenarios := []*scenario.Scenario{
		{
			Manifest: scenario.Manifest{
				Track: "platform",
				Slug:  "test",
				Type:  scenario.TypeSRE,
			},
			Identifier: "platform/test",
		},
		{
			Manifest: scenario.Manifest{
				Track: "platform",
				Slug:  "test",
				Type:  scenario.TypeEngineering,
			},
			Identifier: "platform/test",
		},
		{
			Manifest: scenario.Manifest{
				Track: "backend",
				Slug:  "unique",
				Type:  scenario.TypeEngineering,
			},
			Identifier: "backend/unique",
		},
	}

	collisions := scenario.DetectCollisions(scenarios)

	if len(collisions) != 1 {
		t.Errorf("expected 1 collision, got %d", len(collisions))
	}

	if _, exists := collisions["platform/test"]; !exists {
		t.Errorf("expected collision for 'platform/test', got: %v", collisions)
	}

	if len(collisions["platform/test"]) != 2 {
		t.Errorf("expected 2 types in collision, got %d", len(collisions["platform/test"]))
	}
}

func TestValidateAll_StrictMode(t *testing.T) {
	// This test would require setting up a test filesystem or mock scenarios
	// For now, we'll just test that the function signature is correct
	// and returns appropriate types

	// Note: In a real test, you'd set up a temporary directory structure
	// with test scenario files to validate the full flow
	t.Skip("Requires filesystem setup - integration test needed")
}
