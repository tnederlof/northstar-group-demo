package env

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/northstar-group-demo/democtl/internal/scenario"
)

// GlobalSecrets holds the globally-shared secrets generated once per repo
type GlobalSecrets struct {
	JWTSecret    string
	DemoLoginKey string
}

// RenderOpts contains options for rendering an env file
type RenderOpts struct {
	Scenario *scenario.Scenario
	RepoRoot string
}

// Render generates a runtime environment file for a scenario
// Returns the absolute path to the generated file
func Render(opts RenderOpts) (string, error) {
	if opts.Scenario == nil {
		return "", fmt.Errorf("scenario is required")
	}

	// Parse identifier to get track and slug
	_, slug, err := scenario.ParseIdentifier(opts.Scenario.Identifier)
	if err != nil {
		return "", fmt.Errorf("failed to parse scenario identifier: %w", err)
	}

	// Paths
	// Note: shell script uses $TRACK (the CLI arg, which is actually the type: sre/engineering)
	// and $SLUG (last component of identifier), resulting in .state/<type>/<slug>
	stateDir := filepath.Join(opts.RepoRoot, "demo", ".state")
	globalSecretsPath := filepath.Join(stateDir, "global", "secrets.env")
	contractFile := filepath.Join(opts.RepoRoot, "demo", "shared", "contract", "fider.env.example")
	outputDir := filepath.Join(stateDir, string(opts.Scenario.Manifest.Type), slug)
	outputFile := filepath.Join(outputDir, "runtime.env")

	// Ensure contract file exists
	if _, err := os.Stat(contractFile); err != nil {
		return "", fmt.Errorf("contract file not found: %s", contractFile)
	}

	// Load or generate global secrets
	secrets, err := loadOrGenerateSecrets(globalSecretsPath)
	if err != nil {
		return "", fmt.Errorf("failed to load/generate secrets: %w", err)
	}

	// Determine HTTP port based on track
	httpPort := getHTTPPort(opts.Scenario.Manifest.Type)

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate runtime env file
	if err := generateRuntimeEnv(contractFile, outputFile, slug, httpPort, secrets); err != nil {
		return "", fmt.Errorf("failed to generate runtime env: %w", err)
	}

	return outputFile, nil
}

// loadOrGenerateSecrets loads existing secrets or generates new ones
func loadOrGenerateSecrets(path string) (*GlobalSecrets, error) {
	// If file exists, load it
	if _, err := os.Stat(path); err == nil {
		return loadSecrets(path)
	}

	// Generate new secrets
	secrets := &GlobalSecrets{
		JWTSecret:    generateSecret(),
		DemoLoginKey: generateSecret(),
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create secrets directory: %w", err)
	}

	// Write secrets to file
	if err := saveSecrets(path, secrets); err != nil {
		return nil, fmt.Errorf("failed to save secrets: %w", err)
	}

	return secrets, nil
}

// loadSecrets reads secrets from a file
func loadSecrets(path string) (*GlobalSecrets, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open secrets file: %w", err)
	}
	defer file.Close()

	secrets := &GlobalSecrets{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "JWT_SECRET":
			secrets.JWTSecret = value
		case "DEMO_LOGIN_KEY":
			secrets.DemoLoginKey = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading secrets file: %w", err)
	}

	// Validate required secrets are present
	if secrets.JWTSecret == "" || secrets.DemoLoginKey == "" {
		return nil, fmt.Errorf("secrets file missing required values")
	}

	return secrets, nil
}

// saveSecrets writes secrets to a file
func saveSecrets(path string, secrets *GlobalSecrets) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create secrets file: %w", err)
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "# Auto-generated secrets (do not commit)\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(file, "JWT_SECRET=%s\n", secrets.JWTSecret)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(file, "DEMO_LOGIN_KEY=%s\n", secrets.DemoLoginKey)
	if err != nil {
		return err
	}

	return nil
}

// generateSecret generates a 32-character hex string (16 random bytes)
func generateSecret() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to less secure but still reasonable random generation
		// This should never happen in practice
		panic(fmt.Sprintf("failed to generate random bytes: %v", err))
	}
	return hex.EncodeToString(bytes)
}

// getHTTPPort returns the HTTP port for a given scenario type
func getHTTPPort(scenarioType scenario.ScenarioType) string {
	switch scenarioType {
	case scenario.TypeSRE:
		return "8080"
	case scenario.TypeEngineering:
		return "8082"
	default:
		// Default to 8080 if unknown type
		return "8080"
	}
}

// generateRuntimeEnv generates the runtime.env file by processing the contract file
func generateRuntimeEnv(contractFile, outputFile, slug, httpPort string, secrets *GlobalSecrets) error {
	// Open contract file
	contract, err := os.Open(contractFile)
	if err != nil {
		return fmt.Errorf("failed to open contract file: %w", err)
	}
	defer contract.Close()

	// Create output file
	output, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()

	// Write header
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	if _, err := fmt.Fprintf(output, "# Generated runtime environment for %s\n", slug); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(output, "# Generated at: %s\n", now); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(output, ""); err != nil {
		return err
	}

	// Process contract file line by line
	scanner := bufio.NewScanner(contract)
	for scanner.Scan() {
		line := scanner.Text()

		// Preserve comments and empty lines
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			if _, err := fmt.Fprintln(output, line); err != nil {
				return err
			}
			continue
		}

		// Parse key=value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			// Not a key=value line, skip
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := parts[1] // Don't trim value to preserve intentional spacing

		// Substitute special values
		value = substituteValue(key, value, slug, httpPort, secrets)

		// Write the processed line
		if _, err := fmt.Fprintf(output, "%s=%s\n", key, value); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading contract file: %w", err)
	}

	return nil
}

// substituteValue performs substitutions on a value
func substituteValue(key, value, slug, httpPort string, secrets *GlobalSecrets) string {
	// Handle <generated> secrets
	if strings.TrimSpace(value) == "<generated>" {
		switch key {
		case "JWT_SECRET":
			return secrets.JWTSecret
		case "DEMO_LOGIN_KEY":
			return secrets.DemoLoginKey
		default:
			// Generate a new secret for unknown keys
			return generateSecret()
		}
	}

	// Substitute <slug>
	value = strings.ReplaceAll(value, "<slug>", slug)

	// Substitute <http_port>
	value = strings.ReplaceAll(value, "<http_port>", httpPort)

	return value
}

// ReadEnvFile reads an env file and returns a map of environment variables
func ReadEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open env file: %w", err)
	}
	defer file.Close()

	envVars := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		envVars[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading env file: %w", err)
	}

	return envVars, nil
}
