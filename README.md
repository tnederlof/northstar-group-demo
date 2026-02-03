# Northstar Group Demo

A dual-track demo repository showcasing debugging and development workflows with Warp. Features a fictional enterprise application (Fider) deployed in two ways:

- **SRE Track**: Kubernetes-first, configuration-driven debugging
- **Engineering Track**: Code-first, CI/CD-integrated development

## Quick Start

### Prerequisites

| Track | Required | Optional |
|-------|----------|----------|
| **SRE** | Docker, kind, kubectl, jq, curl | helm |
|| **Engineering** | Docker, git, go (1.25+), node (24+), npm | golangci-lint |

### SRE Track

```bash
# Setup (one-time)
make sre-setup

# Deploy a broken scenario
make sre-demo SCENARIO=platform/bad-rollout

# Access at http://bad-rollout.localhost:8080

# Teardown
make sre-down SCENARIO=platform/bad-rollout
```

### Engineering Track

```bash
# Setup (one-time)
make eng-setup

# Initialize scenario worktree
make eng-scenario-init SCENARIO=backend/ui-regression

# Start scenario stack
make eng-up SCENARIO=backend/ui-regression

# Access at http://ui-regression.localhost:8080

# Teardown
make eng-down SCENARIO=backend/ui-regression
```

## Available Scenarios

### SRE Scenarios

| Scenario | Description | Key Skills |
|----------|-------------|------------|
| `platform/bad-rollout` | Broken DATABASE_URL after deployment | kubectl logs, env vars |
| `platform/resource-exhaustion` | Memory limits too low | kubectl describe, resource tuning |
| `platform/missing-metrics` | ServiceMonitor misconfiguration | Label matching, metrics |
| `platform/network-isolation` | NetworkPolicy blocking DB | Network debugging |

### Engineering Scenarios

| Scenario | Description | Key Skills |
|----------|-------------|------------|
| `backend/ui-regression` | Missing null check | Testing, debugging |
| `frontend/missing-fallback` | React error boundary | Frontend debugging |
| `backend/migration-conflict` | Duplicate migrations | Database migrations |
| `backend/feature-flag-rollout` | Feature flag implementation | Feature development |

## Key Invariants

- **URL Pattern**: `http://<slug>.localhost:8080`
- **Demo Login**: `/__demo/login/<persona>?key=<DEMO_LOGIN_KEY>`
- **Port 8080**: Only one track can run at a time

### Demo Personas

| Slug | Name | Role |
|------|------|------|
| `alex` | Alex Rivera | Administrator |
| `sarah` | Sarah Chen | Collaborator |
| `marcus` | Marcus Wright | Visitor |
| `jennifer` | Jennifer Patel | Visitor |

## Documentation

- [Demo Guide](docs/DEMO_GUIDE.md) - Presentation tips and script
- [Architecture](docs/ARCHITECTURE.md) - Technical overview
- [SRE Runbook](demo/sre/docs/RUNBOOK.md) - SRE track details
- [Engineering Runbook](demo/engineering/docs/RUNBOOK.md) - Engineering track details

## Repository Structure

```
northstar-group-demo/
├── fider/                    # Fider application (with demo mode)
├── demo/
│   ├── shared/               # Shared configs and scripts
│   │   ├── contract/         # Environment contract
│   │   ├── northstar/        # Seed data
│   │   └── scripts/          # Shared scripts
│   ├── sre/                  # SRE track
│   │   ├── kind/             # Kind cluster config
│   │   ├── base/             # Kustomize base
│   │   ├── scenarios/        # Scenario overlays
│   │   └── scripts/          # SRE scripts
│   └── engineering/          # Engineering track
│       ├── compose/          # Docker compose files
│       ├── scenarios/        # Scenario configs
│       └── scripts/          # Engineering scripts
├── docs/                     # Documentation
└── .github/workflows/        # CI/CD
```

## License

MIT License - see [LICENSE](LICENSE) for details.
