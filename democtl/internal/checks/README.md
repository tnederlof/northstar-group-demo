# Checks Package

The `checks` package implements scenario verification and health checking for the Northstar Group demo harness.

## Overview

This package replaces the shell-based checks runner (`demo/shared/scripts/run-checks.sh`) with a typed Go implementation that provides:

- **Stage Selection**: Automatically determines which stage to run checks for based on scenario configuration
- **Check Types**: Supports HTTP, Kubernetes, and Playwright check types
- **Output Modes**: Both human-readable and JSON output formats
- **Filtering**: Run specific subsets of checks using `--only` flag
- **Fail-Fast**: Verification checks stop on first failure, health checks continue

## Architecture

### Core Components

- `checks.go`: Main runner logic, stage selection, and output handling
- `http.go`: HTTP GET checks with retry logic and status code assertions
- `k8s.go`: Kubernetes resource checks using kubectl
- `playwright.go`: UI test execution via Playwright
- `checks_test.go`: Unit tests for stage selection, filtering, and output

### Check Types

#### HTTP Checks

**Type**: `http.get`

Makes HTTP requests and validates status codes with retry logic.

**Fields**:
- `url`: Target URL
- `expect.status`: Expected status codes (positive assertion)
- `expect.status_not`: Disallowed status codes (negative assertion)
- `timeout_seconds`: Total timeout (default: 30)
- `retry_interval`: Retry interval in seconds (default: 2)

**Example**:
```json
{
  "type": "http.get",
  "description": "Health endpoint is responding",
  "url": "http://example.localhost:8080/_health",
  "expect": { "status": [200] },
  "timeout_seconds": 60,
  "retry_interval": 5
}
```

#### Kubernetes Checks

All K8s checks use the configured `KUBE_CONTEXT` (defaults to `kind-fider-demo`) and target the scenario's namespace (`demo-<slug>`).

##### k8s.jqEquals

Validates a K8s resource field using jq.

**Fields**:
- `resource.kind`: Resource kind (e.g., "ConfigMap", "Deployment")
- `resource.name`: Resource name
- `jq`: jq expression to extract field
- `equals`: Expected value

**Example**:
```json
{
  "type": "k8s.jqEquals",
  "description": "ConfigMap has correct value",
  "resource": { "kind": "ConfigMap", "name": "fider-env" },
  "jq": ".data.DATABASE_URL",
  "equals": "postgres://fider:fider@postgres:5432/fider"
}
```

##### k8s.podsContainLog

Checks pod logs for a substring.

**Fields**:
- `selector`: Pod label selector
- `contains`: Substring to search for
- `since_seconds`: Log window in seconds (default: 300)

**Example**:
```json
{
  "type": "k8s.podsContainLog",
  "description": "App logs show successful startup",
  "selector": "app=fider",
  "contains": "Server started on port 3000",
  "since_seconds": 60
}
```

##### k8s.podTerminationReason

Validates pod termination reason.

**Fields**:
- `selector`: Pod label selector
- `reason`: Expected termination reason (e.g., "OOMKilled", "Error")

##### k8s.podRestartCount

Checks that pods have restarted at least a minimum number of times.

**Fields**:
- `selector`: Pod label selector
- `min_restarts`: Minimum restart count (default: 1)

##### k8s.deploymentAvailable

Waits for a deployment to become available.

**Fields**:
- `name`: Deployment name
- `timeout_seconds`: Total timeout (default: 60)
- `wait_seconds`: Initial wait before checking (default: 0)

##### k8s.resourceExists

Checks if a K8s resource exists.

**Fields**:
- `resource.kind`: Resource kind
- `resource.name`: Resource name

##### k8s.serviceMissingPort

Checks that a service is missing a specific port (useful for verifying broken configurations).

**Fields**:
- `name`: Service name
- `port_name`: Port name that should be missing

#### Playwright Checks

**Type**: `playwright.run`

Runs Playwright test suites against the scenario's UI.

**Fields**:
- `suite`: Playwright test suite name (used with `--grep`)
- `headed`: Run in headed mode (default: false, can be overridden with `PLAYWRIGHT_HEADED=true`)

**Environment Variables Passed to Playwright**:
- `BASE_URL`: Constructed from scenario's `url_host` and track-specific port
- `SCENARIO`: Scenario identifier (`<track>/<slug>`)
- `STAGE`: Current stage being checked
- `DEMO_LOGIN_KEY`: Retrieved from K8s ConfigMap (SRE) or `.state/global/secrets.env` (Engineering)

**Example**:
```json
{
  "type": "playwright.run",
  "description": "UI loads and displays content",
  "suite": "^Baseline - Basic Navigation$"
}
```

## Stage Selection

The runner automatically selects a stage using this priority:

1. `checks.default_stage` if set in manifest
2. `"broken"` if it exists
3. `"healthy"` if it exists
4. Lexicographically first stage name

This matches the behavior of the shell script it replaces.

## Check Filtering

The `--only` flag filters checks by type prefix:

- `--only playwright`: Only runs `playwright.run` checks
- `--only http`: Only runs checks starting with `http`
- `--only k8s`: Only runs checks starting with `k8s`
- `--only k8s.pods`: Only runs checks starting with `k8s.pods`

## Output Modes

### Human-Readable (Default)

```
[INFO] Running verify checks for platform/bad-rollout (stage: broken)
[INFO] Namespace: demo-bad-rollout

[PASS] ConfigMap contains intentionally broken DB host
[PASS] App logs mention the broken hostname
[FAIL] Health endpoint is not healthy (expected) (got 0, expected not in [200])

[FAIL] Verification failed. Stopping.
```

### JSON Output (`--json`)

```json
{
  "scenario_id": "platform/bad-rollout",
  "stage": "broken",
  "check_type": "verify",
  "passed": 2,
  "failed": 1,
  "skipped": 0,
  "results": [
    {
      "type": "k8s.jqEquals",
      "description": "ConfigMap contains intentionally broken DB host",
      "status": "pass"
    },
    ...
  ]
}
```

## Usage

### From CLI

```bash
# Run verification checks (fails fast on first failure)
democtl checks verify platform/bad-rollout

# Run health checks (observational, doesn't fail)
democtl checks health platform/bad-rollout --stage healthy

# Filter to specific check types
democtl checks verify platform/bad-rollout --only k8s

# JSON output for scripting
democtl checks verify platform/bad-rollout --json

# Verbose output for debugging
democtl checks verify platform/bad-rollout --verbose
```

### From Go Code

```go
import "github.com/northstar-group-demo/democtl/internal/checks"

runner := checks.NewRunner(checks.RunOpts{
    Scenario:    scenario,
    CheckType:   checks.CheckTypeVerify,
    Stage:       "broken",
    OnlyFilter:  "k8s",
    JSONOutput:  false,
    Verbose:     true,
    KubeContext: "kind-fider-demo",
})

result, err := runner.Run()
if err != nil {
    // Handle error
}
```

## Testing

The package includes comprehensive unit tests for:

- Stage selection logic (default_stage, broken/healthy preference, lexicographic fallback)
- Check filtering (--only flag behavior)
- Output formatting
- Namespace generation

Run tests:
```bash
go test ./internal/checks/...
```

## Migration from Shell Script

This package replaces `demo/shared/scripts/run-checks.sh`. Key differences:

### Behavioral Parity

- ✅ Stage selection logic matches exactly
- ✅ All check types implemented with same semantics
- ✅ Fail-fast behavior for verify, observational for health
- ✅ Same environment variable handling for Playwright
- ✅ Timeout and retry logic preserved

### Improvements

- **Type Safety**: Structured types instead of jq/bash string manipulation
- **Testability**: Core logic can be unit tested without external dependencies
- **Error Handling**: Better error messages with context
- **Performance**: No shell subprocess overhead for orchestration
- **Maintainability**: Clear separation of concerns, easier to extend

### Exit Criteria (Phase 4)

- [x] `democtl checks verify ...` and `democtl checks health ...` no longer call the shell checks runner
- [x] `demo/shared/scripts/run-checks.sh` is no longer called by `democtl`
- [x] All check types implemented and tested
- [x] Stage selection matches shell script behavior
- [x] Output modes (human-readable, JSON, verbose) work correctly
