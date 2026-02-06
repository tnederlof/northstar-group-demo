# Northstar Group Demo

A dual-track demo repository showcasing debugging and development workflows with Warp. Features a fictional enterprise application (Fider) deployed in two ways:

- **SRE Track**: Kubernetes-first, configuration-driven debugging
- **Engineering Track**: Code-first, CI/CD-integrated development

## Quick Start

### For Demo Users (Presenters)

Just want to run scenarios? Follow this path:

```bash
# 1. Build democtl (one-time, or after git pull)
make build-democtl

# 2. Add to PATH for convenience (optional)
fish_add_path $PWD/bin  # fish
# OR for bash/zsh:
export PATH="$PWD/bin:$PATH"

# 3. Setup (installs UI testing dependencies)
democtl setup

# 4. Verify prerequisites
democtl verify

# 5. Run a scenario - it auto-detects track and starts runtime
democtl run platform/bad-rollout     # SRE (Kubernetes)
democtl run backend/ui-regression    # Engineering (Docker Compose)

# 6. Check status
democtl doctor

# 7. Clean up when done
democtl reset-all --force
```

**URLs:**
- SRE scenarios: `http://<slug>.localhost:8080`
- Engineering scenarios: `http://<slug>.localhost:8082`

**Both tracks can run simultaneously!**

### For Maintainers (Scenario Developers)

See [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) for:
- Creating new scenarios
- Maintaining git refs (Engineering track)
- Writing checks and tests
- Rebasing scenario branches

### Prerequisites

| Track | Required | Optional |
|-------|----------|----------|
| **SRE** | Docker, kind, kubectl, jq, curl, helm | - |
| **Engineering** | Docker, git | go (1.21+) for maintainers |
| **UI Testing** | Node.js (18+), npm | - |

**Note for demo users**: Go/Node are only needed for building `democtl` and UI tests. The scenarios themselves run in Docker.

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
| `backend/feature-flag-rollout` | Feature flag implementation | Feature development |

## Key Invariants

- **SRE URL Pattern**: `http://<slug>.localhost:8080`
- **Engineering URL Pattern**: `http://<slug>.localhost:8082`
- **Demo Login**: `/__demo/login/<persona>?key=<DEMO_LOGIN_KEY>`
- **Both tracks can run simultaneously** (different ports)

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
- [Contributing Guide](docs/CONTRIBUTING.md) - How to add/maintain scenarios

## Maintaining Scenarios

**For maintainers only.** See [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) for details on:
- Rebasing Engineering scenario branches when `main` changes
- Managing git tags and refs
- Scenario manifest format
- The git ref contract

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
