package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/northstar-group-demo/democtl/internal/scenario"
)

func TestGenerateSecret(t *testing.T) {
	secret := generateSecret()

	// Should be 32 characters (16 bytes in hex)
	if len(secret) != 32 {
		t.Errorf("expected secret length 32, got %d", len(secret))
	}

	// Should be different each time
	secret2 := generateSecret()
	if secret == secret2 {
		t.Error("expected different secrets, got identical values")
	}

	// Should be valid hex
	for _, c := range secret {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("secret contains invalid hex character: %c", c)
		}
	}
}

func TestGetHTTPPort(t *testing.T) {
	tests := []struct {
		name         string
		scenarioType scenario.ScenarioType
		expected     string
	}{
		{
			name:         "SRE track",
			scenarioType: scenario.TypeSRE,
			expected:     "8080",
		},
		{
			name:         "Engineering track",
			scenarioType: scenario.TypeEngineering,
			expected:     "8082",
		},
		{
			name:         "Unknown track defaults to 8080",
			scenarioType: scenario.ScenarioType("unknown"),
			expected:     "8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getHTTPPort(tt.scenarioType)
			if result != tt.expected {
				t.Errorf("expected port %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSubstituteValue(t *testing.T) {
	secrets := &GlobalSecrets{
		JWTSecret:    "test-jwt-secret",
		DemoLoginKey: "test-demo-login-key",
	}

	tests := []struct {
		name     string
		key      string
		value    string
		slug     string
		httpPort string
		expected string
	}{
		{
			name:     "JWT_SECRET substitution",
			key:      "JWT_SECRET",
			value:    "<generated>",
			slug:     "test-slug",
			httpPort: "8080",
			expected: "test-jwt-secret",
		},
		{
			name:     "DEMO_LOGIN_KEY substitution",
			key:      "DEMO_LOGIN_KEY",
			value:    "<generated>",
			slug:     "test-slug",
			httpPort: "8080",
			expected: "test-demo-login-key",
		},
		{
			name:     "slug substitution",
			key:      "BASE_URL",
			value:    "http://<slug>.localhost:<http_port>",
			slug:     "test-scenario",
			httpPort: "8080",
			expected: "http://test-scenario.localhost:8080",
		},
		{
			name:     "http_port substitution",
			key:      "PORT",
			value:    "<http_port>",
			slug:     "test-slug",
			httpPort: "8082",
			expected: "8082",
		},
		{
			name:     "no substitution needed",
			key:      "EMAIL",
			value:    "none",
			slug:     "test-slug",
			httpPort: "8080",
			expected: "none",
		},
		{
			name:     "unknown generated key creates new secret",
			key:      "UNKNOWN_SECRET",
			value:    "<generated>",
			slug:     "test-slug",
			httpPort: "8080",
			expected: "[generated]", // Will be tested for length and format
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := substituteValue(tt.key, tt.value, tt.slug, tt.httpPort, secrets)

			if tt.expected == "[generated]" {
				// Verify it's a valid secret (32 char hex string)
				if len(result) != 32 {
					t.Errorf("expected generated secret length 32, got %d", len(result))
				}
			} else {
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

func TestSaveAndLoadSecrets(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	secretsPath := filepath.Join(tmpDir, "secrets.env")

	// Original secrets
	original := &GlobalSecrets{
		JWTSecret:    "original-jwt-secret",
		DemoLoginKey: "original-demo-key",
	}

	// Save secrets
	if err := saveSecrets(secretsPath, original); err != nil {
		t.Fatalf("failed to save secrets: %v", err)
	}

	// Load secrets
	loaded, err := loadSecrets(secretsPath)
	if err != nil {
		t.Fatalf("failed to load secrets: %v", err)
	}

	// Verify they match
	if loaded.JWTSecret != original.JWTSecret {
		t.Errorf("JWT_SECRET mismatch: expected %s, got %s", original.JWTSecret, loaded.JWTSecret)
	}
	if loaded.DemoLoginKey != original.DemoLoginKey {
		t.Errorf("DEMO_LOGIN_KEY mismatch: expected %s, got %s", original.DemoLoginKey, loaded.DemoLoginKey)
	}

	// Verify file contents
	content, err := os.ReadFile(secretsPath)
	if err != nil {
		t.Fatalf("failed to read secrets file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "JWT_SECRET=original-jwt-secret") {
		t.Error("secrets file missing JWT_SECRET")
	}
	if !strings.Contains(contentStr, "DEMO_LOGIN_KEY=original-demo-key") {
		t.Error("secrets file missing DEMO_LOGIN_KEY")
	}
	if !strings.Contains(contentStr, "# Auto-generated secrets") {
		t.Error("secrets file missing header comment")
	}
}

func TestLoadOrGenerateSecrets(t *testing.T) {
	t.Run("generates new secrets when file doesn't exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		secretsPath := filepath.Join(tmpDir, "new", "secrets.env")

		secrets, err := loadOrGenerateSecrets(secretsPath)
		if err != nil {
			t.Fatalf("failed to load/generate secrets: %v", err)
		}

		// Verify secrets were generated
		if secrets.JWTSecret == "" {
			t.Error("JWT_SECRET is empty")
		}
		if secrets.DemoLoginKey == "" {
			t.Error("DEMO_LOGIN_KEY is empty")
		}
		if len(secrets.JWTSecret) != 32 {
			t.Errorf("JWT_SECRET wrong length: expected 32, got %d", len(secrets.JWTSecret))
		}

		// Verify file was created
		if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
			t.Error("secrets file was not created")
		}
	})

	t.Run("loads existing secrets when file exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		secretsPath := filepath.Join(tmpDir, "secrets.env")

		// Create initial secrets
		original := &GlobalSecrets{
			JWTSecret:    "existing-jwt-secret",
			DemoLoginKey: "existing-demo-key",
		}
		if err := os.MkdirAll(filepath.Dir(secretsPath), 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		if err := saveSecrets(secretsPath, original); err != nil {
			t.Fatalf("failed to save secrets: %v", err)
		}

		// Load them
		loaded, err := loadOrGenerateSecrets(secretsPath)
		if err != nil {
			t.Fatalf("failed to load/generate secrets: %v", err)
		}

		// Verify they match the original
		if loaded.JWTSecret != original.JWTSecret {
			t.Errorf("JWT_SECRET mismatch: expected %s, got %s", original.JWTSecret, loaded.JWTSecret)
		}
		if loaded.DemoLoginKey != original.DemoLoginKey {
			t.Errorf("DEMO_LOGIN_KEY mismatch: expected %s, got %s", original.DemoLoginKey, loaded.DemoLoginKey)
		}
	})
}

func TestGenerateRuntimeEnv(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create a minimal contract file
	contractFile := filepath.Join(tmpDir, "contract.env")
	contractContent := `# Test contract file
HOST_MODE=single
BASE_URL=http://<slug>.localhost:<http_port>
DATABASE_URL=postgres://fider:fider@postgres:5432/fider?sslmode=disable
JWT_SECRET=<generated>
DEMO_LOGIN_KEY=<generated>
EMAIL=none
`
	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		t.Fatalf("failed to write contract file: %v", err)
	}

	outputFile := filepath.Join(tmpDir, "runtime.env")
	secrets := &GlobalSecrets{
		JWTSecret:    "test-jwt-secret-12345678901234567890",
		DemoLoginKey: "test-demo-key-12345678901234567890",
	}

	// Generate runtime env
	err := generateRuntimeEnv(contractFile, outputFile, "test-scenario", "8080", secrets)
	if err != nil {
		t.Fatalf("failed to generate runtime env: %v", err)
	}

	// Read and verify output
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Check for header
	if !strings.Contains(contentStr, "# Generated runtime environment for test-scenario") {
		t.Error("missing header comment")
	}

	// Check for substituted values
	expectedLines := []string{
		"HOST_MODE=single",
		"BASE_URL=http://test-scenario.localhost:8080",
		"DATABASE_URL=postgres://fider:fider@postgres:5432/fider?sslmode=disable",
		"JWT_SECRET=test-jwt-secret-12345678901234567890",
		"DEMO_LOGIN_KEY=test-demo-key-12345678901234567890",
		"EMAIL=none",
	}

	for _, expected := range expectedLines {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("output missing expected line: %s", expected)
		}
	}

	// Verify contract comments are preserved
	if !strings.Contains(contentStr, "# Test contract file") {
		t.Error("contract comments not preserved")
	}
}

func TestRender(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create contract file
	contractDir := filepath.Join(tmpDir, "demo", "shared", "contract")
	if err := os.MkdirAll(contractDir, 0755); err != nil {
		t.Fatalf("failed to create contract directory: %v", err)
	}

	contractFile := filepath.Join(contractDir, "fider.env.example")
	contractContent := `# Fider Demo Environment Contract
HOST_MODE=single
BASE_URL=http://<slug>.localhost:<http_port>
JWT_SECRET=<generated>
DEMO_LOGIN_KEY=<generated>
`
	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		t.Fatalf("failed to write contract file: %v", err)
	}

	// Create test scenario
	s := &scenario.Scenario{
		Manifest: scenario.Manifest{
			Track: "platform",
			Slug:  "test-scenario",
			Type:  scenario.TypeSRE,
		},
		Identifier: "platform/test-scenario",
		RepoRoot:   tmpDir,
	}

	// Render env file
	outputPath, err := Render(RenderOpts{
		Scenario: s,
		RepoRoot: tmpDir,
	})
	if err != nil {
		t.Fatalf("failed to render env: %v", err)
	}

	// Verify output path
	// Note: path uses type (sre) not track (platform)
	expectedPath := filepath.Join(tmpDir, "demo", ".state", "sre", "test-scenario", "runtime.env")
	if outputPath != expectedPath {
		t.Errorf("unexpected output path: expected %s, got %s", expectedPath, outputPath)
	}

	// Verify output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("output file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Verify substitutions
	if !strings.Contains(contentStr, "BASE_URL=http://test-scenario.localhost:8080") {
		t.Error("BASE_URL not correctly substituted")
	}

	// Verify secrets were generated (should be 32 char hex strings)
	lines := strings.Split(contentStr, "\n")
	var jwtFound, demoKeyFound bool
	for _, line := range lines {
		if strings.HasPrefix(line, "JWT_SECRET=") {
			jwtFound = true
			secret := strings.TrimPrefix(line, "JWT_SECRET=")
			if len(secret) != 32 {
				t.Errorf("JWT_SECRET wrong length: expected 32, got %d", len(secret))
			}
		}
		if strings.HasPrefix(line, "DEMO_LOGIN_KEY=") {
			demoKeyFound = true
			secret := strings.TrimPrefix(line, "DEMO_LOGIN_KEY=")
			if len(secret) != 32 {
				t.Errorf("DEMO_LOGIN_KEY wrong length: expected 32, got %d", len(secret))
			}
		}
	}

	if !jwtFound {
		t.Error("JWT_SECRET not found in output")
	}
	if !demoKeyFound {
		t.Error("DEMO_LOGIN_KEY not found in output")
	}

	// Verify global secrets were persisted
	globalSecretsPath := filepath.Join(tmpDir, "demo", ".state", "global", "secrets.env")
	if _, err := os.Stat(globalSecretsPath); os.IsNotExist(err) {
		t.Error("global secrets file was not created")
	}

	// Run render again and verify secrets are reused
	outputPath2, err := Render(RenderOpts{
		Scenario: s,
		RepoRoot: tmpDir,
	})
	if err != nil {
		t.Fatalf("failed to render env second time: %v", err)
	}

	content2, err := os.ReadFile(outputPath2)
	if err != nil {
		t.Fatalf("failed to read second output file: %v", err)
	}

	// Extract secrets from both runs
	extractSecret := func(content, key string) string {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, key+"=") {
				return strings.TrimPrefix(line, key+"=")
			}
		}
		return ""
	}

	jwt1 := extractSecret(contentStr, "JWT_SECRET")
	jwt2 := extractSecret(string(content2), "JWT_SECRET")
	if jwt1 != jwt2 {
		t.Error("JWT_SECRET changed between renders (should be stable)")
	}

	key1 := extractSecret(contentStr, "DEMO_LOGIN_KEY")
	key2 := extractSecret(string(content2), "DEMO_LOGIN_KEY")
	if key1 != key2 {
		t.Error("DEMO_LOGIN_KEY changed between renders (should be stable)")
	}
}

func TestRenderEngineeringTrack(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create contract file
	contractDir := filepath.Join(tmpDir, "demo", "shared", "contract")
	if err := os.MkdirAll(contractDir, 0755); err != nil {
		t.Fatalf("failed to create contract directory: %v", err)
	}

	contractFile := filepath.Join(contractDir, "fider.env.example")
	contractContent := `BASE_URL=http://<slug>.localhost:<http_port>
`
	if err := os.WriteFile(contractFile, []byte(contractContent), 0644); err != nil {
		t.Fatalf("failed to write contract file: %v", err)
	}

	// Create engineering scenario
	s := &scenario.Scenario{
		Manifest: scenario.Manifest{
			Track: "backend",
			Slug:  "ui-regression",
			Type:  scenario.TypeEngineering,
		},
		Identifier: "backend/ui-regression",
		RepoRoot:   tmpDir,
	}

	// Render env file
	outputPath, err := Render(RenderOpts{
		Scenario: s,
		RepoRoot: tmpDir,
	})
	if err != nil {
		t.Fatalf("failed to render env: %v", err)
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Verify engineering track uses port 8082
	if !strings.Contains(contentStr, "BASE_URL=http://ui-regression.localhost:8082") {
		t.Errorf("Engineering track should use port 8082, got: %s", contentStr)
	}
}
