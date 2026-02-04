# Demo Track Selector

Welcome to the Northstar Group Demo! This directory contains two distinct demo tracks, each tailored to different audiences and learning objectives.

## Which Track Should You Use?

### 🚀 SRE Track
**Best for**: Platform engineers, SREs, DevOps practitioners, infrastructure teams

**You'll learn about**:
- Kubernetes deployment strategies
- Gateway API and Envoy Gateway
- Service reliability and incident response
- Traffic management and canary deployments
- Horizontal Pod Autoscaling

**Runtime**: kind (Kubernetes in Docker)

**Get Started**: See [`sre/docs/RUNBOOK.md`](sre/docs/RUNBOOK.md)

---

### 💻 Engineering Track
**Best for**: Application developers, QA engineers, engineering managers

**You'll learn about**:
- Bug fixing and debugging workflows
- Code review best practices
- CI/CD automation
- Testing strategies (unit, integration, E2E)
- Git worktrees for isolation

**Runtime**: Docker Compose

**Get Started**: See [`engineering/docs/RUNBOOK.md`](engineering/docs/RUNBOOK.md)

---

## Quick Reference

### Golden Path (Recommended)

```bash
# One-time setup
make setup

# Verify prerequisites
make verify

# Run any scenario (auto-detects track)
make run SCENARIO=platform/bad-rollout    # SRE at :8080
make run SCENARIO=backend/ui-regression   # Engineering at :8082

# Check health
make health SCENARIO=<track>/<slug>

# Reset
make reset SCENARIO=<track>/<slug>
make reset-all FORCE=true  # Clean up everything

# View status
make doctor
```

### Advanced: Track-Specific Commands

**SRE Track:**
```bash
make sre-setup          # Manual setup
make sre-demo SCENARIO=platform/bad-rollout
make sre-verify SCENARIO=platform/bad-rollout
make sre-reset SCENARIO=platform/bad-rollout
make sre-down-all
```

**Engineering Track:**
```bash
make eng-setup          # Manual setup
make eng-scenario-init SCENARIO=backend/ui-regression
make eng-up SCENARIO=backend/ui-regression
make eng-verify SCENARIO=backend/ui-regression
make eng-down SCENARIO=backend/ui-regression
```

### Verification Commands

```bash
# Run only UI/Playwright checks
make ui-verify TYPE=sre SCENARIO=platform/healthy

# Verify with a specific stage
make sre-verify SCENARIO=platform/bad-rollout STAGE=broken
make eng-verify SCENARIO=backend/ui-regression STAGE=fixed
```

## Available Scenarios

### SRE Track

| Scenario | Duration | Difficulty | Focus |
|----------|----------|------------|-------|
| `platform/healthy` | 5 min | Easy | Baseline verification |
| `platform/bad-rollout` | 15 min | Medium | Deployment failures and rollback |
| `platform/resource-exhaustion` | 12 min | Medium | OOMKilled pods, memory limits |
| `platform/network-isolation` | 15 min | Medium | NetworkPolicy blocking database |
| `platform/missing-metrics` | 10 min | Easy | ServiceMonitor misconfiguration |

### Engineering Track

| Scenario | Duration | Difficulty | Focus |
|----------|----------|------------|-------|
| `backend/ui-regression` | 12 min | Easy | Missing null check causes 500 errors |

## Documentation

- **[Demo Guide](../docs/DEMO_GUIDE.md)**: Comprehensive guide to choosing and running scenarios
- **[Architecture](../docs/ARCHITECTURE.md)**: Technical overview of the system
- **[Personas](../docs/PERSONAS.md)**: User profiles and use cases
- **[Contributing](../docs/CONTRIBUTING.md)**: How to add new scenarios
- **[Contract](docs/CONTRACT.md)**: URL and environment variable contracts

## Prerequisites

### SRE Track
- Docker Desktop (with Kubernetes enabled)
- `kind` - Kubernetes in Docker
- `kubectl` - Kubernetes CLI
- `make`

### Engineering Track
- Docker and Docker Compose
- Go 1.25+ (for backend changes)
- Node.js 24+ and npm (for frontend changes)
- `make`
- Playwright (optional, for E2E tests)

## Support

- **Issues**: Found a bug? [Open an issue](https://github.com/your-org/northstar-group-demo/issues)
- **Questions**: See the [Demo Guide](../docs/DEMO_GUIDE.md) or [Contributing Guide](../docs/CONTRIBUTING.md)
- **Improvements**: Contributions welcome! See [CONTRIBUTING.md](../docs/CONTRIBUTING.md)
