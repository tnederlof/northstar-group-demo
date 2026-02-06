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

## The Golden Path (Demo Users)

### Step 0: Build democtl

Build the `democtl` binary once (and after each `git pull`):

```bash
make build-democtl
```

**Optional: Add to PATH** for convenience:

```bash
# Fish
fish_add_path $PWD/bin

# Bash/Zsh
export PATH="$PWD/bin:$PATH"
```

**Optional: Enable shell completion:**

```bash
# Bash
echo 'source <(democtl completion bash)' >> ~/.bashrc && source ~/.bashrc

# Zsh
echo 'source <(democtl completion zsh)' >> ~/.zshrc && source ~/.zshrc

# Fish
echo 'democtl completion fish | source' >> ~/.config/fish/config.fish
```

### Step 1: One-Time Setup

Install UI testing dependencies (Playwright):

```bash
democtl setup
```

**Note**: This is only needed if you want to run UI/Playwright checks. Scenarios will run without this.

### Step 2: Verify Prerequisites

Check that you have all required tools:

```bash
democtl verify
```

If verification fails, install the missing tools and re-run.

**For demo users**: You only need Docker, kind, kubectl for SRE or Docker for Engineering. Go/Node are optional unless you're developing.

### Step 3: Run a Scenario

`democtl run` automatically:
- Detects scenario type (SRE or Engineering)
- Ensures runtime is up
- Deploys the scenario
- Runs verification checks

**SRE Example** - Kubernetes:

```bash
democtl run platform/bad-rollout
```

Visit: **http://bad-rollout.localhost:8080**

**Engineering Example** - Docker Compose:

```bash
democtl run backend/ui-regression
```

Visit: **http://ui-regression.localhost:8082**

**Engineering Workshop Flow:**

Engineering scenarios create a git worktree you can edit:

```bash
# 1. Run creates worktree at:
#    demo/engineering/scenarios/backend/ui-regression/worktree/
democtl run backend/ui-regression

# 2. Edit code in the worktree
cd demo/engineering/scenarios/backend/ui-regression/worktree/fider/
# Make your fixes, commit changes
git add . && git commit -m "fix: add null check"

# 3. Reset to broken baseline
democtl reset backend/ui-regression

# 4. Or jump to solved state
democtl fix-it backend/ui-regression
```

### Step 4: Clean Up

When done:

```bash
# Reset a specific scenario
democtl reset platform/bad-rollout

# Clean up everything
democtl reset-all --force
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
  - `platform/resource-exhaustion` - Memory limits causing OOMKilled pods
  - `platform/network-isolation` - NetworkPolicy blocking database access
  - `platform/missing-metrics` - ServiceMonitor configuration issues

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
democtl run platform/healthy

# In another terminal, start an Engineering scenario
democtl run backend/ui-regression

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
democtl list-scenarios
```

### Check Scenario Health

```bash
democtl checks health platform/healthy
```

### View Runtime Status

```bash
democtl doctor
```

### Run Without Verification

```bash
democtl run platform/healthy --verify=false
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
democtl run platform/healthy
```

### Clean Slate

To completely reset everything:

```bash
democtl reset-all --force NUKE_UI=true
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
democtl setup
democtl verify

# 2. Run an SRE scenario
democtl run platform/bad-rollout

# 3. Visit the URL
open http://bad-rollout.localhost:8080

# 4. Check health
democtl checks health platform/bad-rollout

# 5. Try an Engineering scenario (runs simultaneously!)
democtl run backend/ui-regression
open http://ui-regression.localhost:8082

# 6. Clean up
democtl reset-all --force
```

## Quick Reference

| Command | Purpose |
|---------|---------|
| `democtl setup` | Install UI testing dependencies |
| `democtl verify` | Check all prerequisites |
| `democtl run <track>/<slug>` | Run a scenario (auto-detects track) |
| `democtl reset <track>/<slug>` | Reset a scenario |
| `democtl reset-all --force` | Clean up everything |
| `democtl doctor` | Show runtime status |
| `democtl list-scenarios` | List all scenarios |
| `democtl checks health <track>/<slug>` | Check scenario health |
| `democtl --help` | Show all commands |

## Getting Help

- **Repository Issues**: [GitHub Issues](https://github.com/your-org/northstar-group-demo/issues)
- **Documentation**: Check the `docs/` and `demo/docs/` directories
- **Make Help**: Run `make help` to see all available commands
