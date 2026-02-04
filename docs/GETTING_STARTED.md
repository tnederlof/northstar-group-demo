# Getting Started with Northstar Group Demo

Welcome! This guide will walk you through running your first demo scenario in 3-4 commands.

## Prerequisites

### Required Tools

- **Docker** - [Install Docker Desktop](https://www.docker.com/products/docker-desktop)
- **kind** - [Install kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- **kubectl** - [Install kubectl](https://kubernetes.io/docs/tasks/tools/)
- **Node.js** (v18+) - [Install Node.js](https://nodejs.org/)
- **Go** (v1.21+) - [Install Go](https://go.dev/doc/install)
- **jq** - [Install jq](https://stedolan.github.io/jq/download/)
- **curl** - Usually pre-installed on macOS/Linux

### Optional Tools

- **helm** - Required for SRE track (Gateway API)
- **golangci-lint** - For linting code changes

### Installation Tips (macOS)

```bash
# Using Homebrew
brew install docker kind kubectl node go jq curl helm

# Start Docker Desktop
open -a Docker
```

## The Golden Path (3-4 Commands)

### 1. One-Time Setup

This installs UI testing dependencies (Playwright):

```bash
make setup
```

### 2. Verify Prerequisites

Check that you have all required tools installed:

```bash
make verify
```

If verification fails, install the missing tools and run `make verify` again.

### 3. Run a Scenario

The `run` command automatically:
- Detects the scenario type (SRE or Engineering)
- Ensures the required runtime is up
- Deploys/starts the scenario
- Runs verification checks

**SRE Example** - Bad Rollout (Kubernetes):

```bash
make run SCENARIO=platform/bad-rollout
```

Visit: **http://bad-rollout.localhost:8080**

**Engineering Example** - UI Regression (Docker Compose):

```bash
make run SCENARIO=backend/ui-regression
```

Visit: **http://ui-regression.localhost:8082**

**Engineering Workshop Flow:**

For engineering scenarios, `make run` creates a git worktree you can edit:

```bash
# 1. Run creates worktree at demo/engineering/scenarios/backend/ui-regression/worktree/
make run SCENARIO=backend/ui-regression

# 2. Edit code in the worktree, commit changes
cd demo/engineering/scenarios/backend/ui-regression/worktree/fider/
# make your fixes...
git add . && git commit -m "fix: add null check"

# 3. Reset to broken baseline to start over
make reset SCENARIO=backend/ui-regression

# 4. Or jump to solved state (escape hatch)
make fix-it SCENARIO=backend/ui-regression
```

ℹ️ **Workshop commits**: Your commits stay on a local `ws/backend/ui-regression` branch. Use `FORCE=true` to override dirty worktree warnings.

### 4. Clean Up

When you're done:

```bash
# Reset a specific scenario
make reset SCENARIO=platform/bad-rollout

# Or clean up everything
make reset-all FORCE=true
```

## Understanding the Demo Tracks

This demo has **two independent tracks** that can run **simultaneously** on the same machine:

### SRE Track (Kubernetes)

- **Runtime**: kind cluster + Envoy Gateway
- **Port**: 8080 (configurable in `demo/shared/contract/ports.env`)
- **URL Pattern**: `http://<slug>.localhost:8080`
- **Scenarios**:
  - `platform/healthy` - Baseline working application
  - `platform/bad-rollout` - Failed deployment rollout
  - `platform/resource-exhaustion` - OOMKilled pods
  - `platform/network-isolation` - Network policy misconfiguration
  - `platform/missing-metrics` - Observability gaps

### Engineering Track (Docker Compose)

- **Runtime**: Traefik edge proxy + Docker Compose
- **Port**: 8082 (configurable in `demo/shared/contract/ports.env`)
- **URL Pattern**: `http://<slug>.localhost:8082`
- **Scenarios**:
  - `backend/ui-regression` - Missing null check causing 500 errors

### Running Both Tracks Simultaneously

Thanks to per-track port configuration, you can run SRE and Engineering scenarios at the same time:

```bash
# Start an SRE scenario
make run SCENARIO=platform/healthy

# In another terminal, start an Engineering scenario
make run SCENARIO=backend/ui-regression

# Both are accessible:
# - http://healthy.localhost:8080 (SRE)
# - http://ui-regression.localhost:8082 (Engineering)
```

## Port Configuration

The demo uses fixed ports for each track:

- **SRE HTTP**: 8080
- **Engineering HTTP**: 8082  
- **Engineering Dashboard**: 8083 (Traefik)

These ports are hardcoded throughout the codebase for consistency.

## Additional Commands

### List Available Scenarios

```bash
make list-scenarios
```

### Check Scenario Health

```bash
make health SCENARIO=platform/healthy
```

### View Runtime Status

```bash
make doctor
```

### Reset Without Verification

```bash
make run SCENARIO=platform/healthy VERIFY=false
```

## Troubleshooting

### Port Already in Use

If you see "port already in use" errors:

1. **Check what's using the port:**
   ```bash
   lsof -i :8080
   ```

2. **Stop the process** that's using the port

3. **Run the scenario again**

### Docker Not Running

Make sure Docker Desktop is running:

```bash
docker info
```

If it fails, start Docker Desktop and wait for it to fully start.

### kind Cluster Issues

If the kind cluster is in a bad state:

```bash
# Delete and recreate
kind delete cluster --name fider-demo
make run SCENARIO=platform/healthy
```

### Clean Slate

To completely reset everything:

```bash
make reset-all FORCE=true NUKE_UI=true
```

This removes:
- All Docker containers and networks
- kind cluster
- Git worktrees
- Local state
- UI dependencies (with `NUKE_UI=true`)

## Next Steps

- **Explore scenarios**: See [demo/docs/DEMO_GUIDE.md](../demo/docs/DEMO_GUIDE.md)
- **SRE workflows**: See [demo/sre/docs/RUNBOOK.md](../demo/sre/docs/RUNBOOK.md)
- **Engineering workflows**: See [demo/engineering/docs/RUNBOOK.md](../demo/engineering/docs/RUNBOOK.md)
- **Advanced commands**: Run `make help`

## Example Session

Here's a complete example session from start to finish:

```bash
# 1. Setup
cd northstar-group-demo
make setup
make verify

# 2. Run an SRE scenario
make run SCENARIO=platform/bad-rollout

# 3. Visit the URL
open http://bad-rollout.localhost:8080

# 4. Check health
make health SCENARIO=platform/bad-rollout

# 5. Try an Engineering scenario (runs simultaneously!)
make run SCENARIO=backend/ui-regression
open http://ui-regression.localhost:8082

# 6. Clean up
make reset-all FORCE=true
```

## Quick Reference

| Command | Purpose |
|---------|---------|
| `make setup` | Install UI testing dependencies |
| `make verify` | Check all prerequisites |
| `make run SCENARIO=<track>/<slug>` | Run a scenario (auto-detects track) |
| `make reset SCENARIO=<track>/<slug>` | Reset a scenario |
| `make reset-all FORCE=true` | Clean up everything |
| `make doctor` | Show runtime status |
| `make list-scenarios` | List all scenarios |
| `make health SCENARIO=<track>/<slug>` | Check scenario health |
| `make help` | Show all commands |

## Getting Help

- **Repository Issues**: [GitHub Issues](https://github.com/your-org/northstar-group-demo/issues)
- **Documentation**: Check the `docs/` and `demo/docs/` directories
- **Make Help**: Run `make help` to see all available commands
