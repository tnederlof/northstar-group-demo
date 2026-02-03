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

### SRE Track Commands

```bash
# One-time setup
make sre-setup
make sre-cluster-up

# Run a scenario
make sre-demo SCENARIO=platform/bad-rollout

# Health checks
make sre-health SCENARIO=platform/bad-rollout

# Reset scenario
make sre-reset SCENARIO=platform/bad-rollout

# Teardown
make sre-teardown-all
```

### Engineering Track Commands

```bash
# Setup a scenario worktree
make eng-scenario-init SCENARIO=backend/api-regression

# Start the app
make eng-up SCENARIO=backend/api-regression

# Run CI checks
make eng-ci SCENARIO=backend/api-regression

# Monitor logs
make eng-sniff SCENARIO=backend/api-regression

# Stop the app
make eng-down SCENARIO=backend/api-regression
```

## Available Scenarios

### SRE Track

| Scenario | Duration | Difficulty | Focus |
|----------|----------|------------|-------|
| `platform/bad-rollout` | 15 min | Medium | Deployment failures and rollback |
| `platform/traffic-split` | 12 min | Medium | Canary deployments with Gateway API |
| `platform/config-change` | 10 min | Easy | ConfigMap updates and rolling restarts |
| `platform/scale-event` | 15 min | Hard | HPA and resource management |

### Engineering Track

| Scenario | Duration | Difficulty | Focus |
|----------|----------|------------|-------|
| `backend/api-regression` | 12 min | Easy | Missing null check causes 500 errors |
| `frontend/error-boundary` | 12 min | Medium | React error handling |
| `backend/migration-conflict` | 15 min | Medium | Database migration conflicts |
| `backend/feature-flag-rollout` | 10 min | Easy | Progressive feature delivery |

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
- Go 1.24+ (for backend changes)
- Node.js 20+ and npm (for frontend changes)
- `make`
- Playwright (optional, for E2E tests)

## Support

- **Issues**: Found a bug? [Open an issue](https://github.com/your-org/northstar-group-demo/issues)
- **Questions**: See the [Demo Guide](../docs/DEMO_GUIDE.md) or [Contributing Guide](../docs/CONTRIBUTING.md)
- **Improvements**: Contributions welcome! See [CONTRIBUTING.md](../docs/CONTRIBUTING.md)
