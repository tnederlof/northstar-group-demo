package scenario

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ScenarioType represents the type of scenario (sre or engineering)
type ScenarioType string

const (
	TypeSRE         ScenarioType = "sre"
	TypeEngineering ScenarioType = "engineering"
)

// ResetStrategy defines how a scenario should be reset
type ResetStrategy string

const (
	ResetNamespaceDelete ResetStrategy = "namespace-delete"
	ResetWorktreeReset   ResetStrategy = "worktree-reset"
)

// GitConfig contains git-related configuration for engineering scenarios
type GitConfig struct {
	MaintenanceBranch string `json:"maintenance_branch"`
	BrokenRef         string `json:"broken_ref"`
	SolvedRef         string `json:"solved_ref"`
	WorkBranch        string `json:"work_branch"`
}

// HTTPExpect defines expectations for HTTP checks
type HTTPExpect struct {
	Status    []int `json:"status,omitempty"`
	StatusNot []int `json:"status_not,omitempty"`
}

// K8sResource represents a Kubernetes resource reference
type K8sResource struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

// Check represents a single check in a stage
type Check struct {
	Type            string       `json:"type"`
	Description     string       `json:"description"`
	// HTTP check fields
	URL             string       `json:"url,omitempty"`
	Expect          HTTPExpect   `json:"expect,omitempty"`
	TimeoutSeconds  int          `json:"timeout_seconds,omitempty"`
	RetryInterval   int          `json:"retry_interval,omitempty"`
	// K8s check fields
	Resource        *K8sResource `json:"resource,omitempty"`
	JQ              string       `json:"jq,omitempty"`
	Equals          string       `json:"equals,omitempty"`
	Selector        string       `json:"selector,omitempty"`
	Contains        string       `json:"contains,omitempty"`
	SinceSeconds    int          `json:"since_seconds,omitempty"`
	Reason          string       `json:"reason,omitempty"`
	MinRestarts     int          `json:"min_restarts,omitempty"`
	Name            string       `json:"name,omitempty"`
	WaitSeconds     int          `json:"wait_seconds,omitempty"`
	PortName        string       `json:"port_name,omitempty"`
	// Playwright check fields
	Suite           string       `json:"suite,omitempty"`
	Headed          bool         `json:"headed,omitempty"`
}

// Stage represents a scenario stage (e.g., broken, fixed, healthy)
type Stage struct {
	Verify []Check `json:"verify,omitempty"`
	Health []Check `json:"health,omitempty"`
}

// Checks defines the checks configuration for a scenario
type Checks struct {
	Version      int              `json:"version"`
	DefaultStage string           `json:"default_stage,omitempty"`
	Stages       map[string]Stage `json:"stages"`
}

// Manifest represents a scenario.json file
type Manifest struct {
	Track         string        `json:"track"`
	Slug          string        `json:"slug"`
	Title         string        `json:"title"`
	Type          ScenarioType  `json:"type"`
	URLHost       string        `json:"url_host"`
	Seed          bool          `json:"seed"`
	ResetStrategy ResetStrategy `json:"reset_strategy"`
	Git           *GitConfig    `json:"git,omitempty"`
	Description   string        `json:"description"`
	Symptoms      []string      `json:"symptoms"`
	FixHints      []string      `json:"fix_hints"`
	Checks        Checks        `json:"checks"`
}

// Scenario represents a discovered scenario with its manifest and paths
type Scenario struct {
	Manifest   Manifest
	Dir        string // absolute path to scenario directory
	Identifier string // "<track>/<slug>" format
	RepoRoot   string // absolute path to repository root
}

// LoadManifest loads and parses a scenario.json file
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// Discover finds all scenario.json files in the demo directory
func Discover(repoRoot string) ([]*Scenario, error) {
	demoDir := filepath.Join(repoRoot, "demo")

	var scenarios []*Scenario

	// Search in both sre and engineering directories
	for _, trackDir := range []string{"sre", "engineering"} {
		scenariosDir := filepath.Join(demoDir, trackDir, "scenarios")

		err := filepath.Walk(scenariosDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip worktree directories
			if info.IsDir() && info.Name() == "worktree" {
				return filepath.SkipDir
			}

			if info.IsDir() || info.Name() != "scenario.json" {
				return nil
			}

			manifest, err := LoadManifest(path)
			if err != nil {
				return fmt.Errorf("failed to load %s: %w", path, err)
			}

			scenarioDir := filepath.Dir(path)
			identifier := fmt.Sprintf("%s/%s", manifest.Track, manifest.Slug)

			scenarios = append(scenarios, &Scenario{
				Manifest:   *manifest,
				Dir:        scenarioDir,
				Identifier: identifier,
				RepoRoot:   repoRoot,
			})

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to discover scenarios in %s: %w", trackDir, err)
		}
	}

	return scenarios, nil
}

// Resolve finds a scenario by identifier (track/slug format)
// If typeOverride is specified, it will only return scenarios of that type
// Returns an error if multiple scenarios match and no type override is given
func Resolve(repoRoot, identifier string, typeOverride ScenarioType) (*Scenario, error) {
	scenarios, err := Discover(repoRoot)
	if err != nil {
		return nil, err
	}

	var matches []*Scenario
	for _, s := range scenarios {
		if s.Identifier == identifier {
			if typeOverride != "" && s.Manifest.Type != typeOverride {
				continue
			}
			matches = append(matches, s)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("scenario not found: %s", identifier)
	}

	if len(matches) > 1 {
		return nil, fmt.Errorf("ambiguous scenario %s (found in both tracks, use --type to disambiguate)", identifier)
	}

	return matches[0], nil
}

// DetectCollisions checks for scenarios with the same identifier across different types
func DetectCollisions(scenarios []*Scenario) map[string][]ScenarioType {
	identifierTypes := make(map[string]map[ScenarioType]bool)

	for _, s := range scenarios {
		if _, exists := identifierTypes[s.Identifier]; !exists {
			identifierTypes[s.Identifier] = make(map[ScenarioType]bool)
		}
		identifierTypes[s.Identifier][s.Manifest.Type] = true
	}

	collisions := make(map[string][]ScenarioType)
	for identifier, types := range identifierTypes {
		if len(types) > 1 {
			typeList := make([]ScenarioType, 0, len(types))
			for t := range types {
				typeList = append(typeList, t)
			}
			collisions[identifier] = typeList
		}
	}

	return collisions
}

// StateDir returns the absolute path to the scenario's state directory
func (s *Scenario) StateDir() string {
	return filepath.Join(s.RepoRoot, "demo", ".state", string(s.Manifest.Type), s.Manifest.Track, s.Manifest.Slug)
}

// WorktreeDir returns the absolute path to the scenario's worktree directory (engineering only)
func (s *Scenario) WorktreeDir() string {
	if s.Manifest.Type != TypeEngineering {
		return ""
	}
	return filepath.Join(s.Dir, "worktree")
}

// GetRepoRoot finds the repository root by looking for the demo directory
func GetRepoRoot() (string, error) {
	// Start from current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Walk up the directory tree looking for demo directory
	dir := cwd
	for {
		demoDir := filepath.Join(dir, "demo")
		if info, err := os.Stat(demoDir); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding demo directory
			return "", fmt.Errorf("could not find repository root (looking for 'demo' directory)")
		}
		dir = parent
	}
}

// ParseIdentifier splits a scenario identifier into track and slug
func ParseIdentifier(identifier string) (track, slug string, err error) {
	parts := strings.SplitN(identifier, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid scenario identifier format (expected 'track/slug'): %s", identifier)
	}
	return parts[0], parts[1], nil
}
