# Northstar Group Demo

## Overview

Dual-track demo repository showcasing Warp workflows. Two runtimes (SRE: Kubernetes, Engineering: CI+Compose) sharing one contract.

## Quick Navigation

- **Application code**: `fider/` (see `fider/AGENTS.md`)
- **Demo harness**: `demo/` (see `demo/AGENTS.md`)
- **Narrative docs**: `docs/`

## Key Invariants

- Single source of truth: `demo/shared/contract/`
- SRE URL contract: `http://<slug>.localhost:8080`
- Engineering URL contract: `http://<slug>.localhost:8082`
- Scenario paths: always `<track>/<slug>` (2 levels)
- Scenarios are manifest-driven: `scenario.json` required
- Both tracks can run simultaneously

## Common Tasks

### Golden Path (Recommended)

- Setup: `make setup`
- Verify: `make verify`
- Run any scenario: `make run SCENARIO=<track>/<slug>`
- Reset: `make reset SCENARIO=<track>/<slug>`
- Clean up: `make reset-all FORCE=true`
- Status: `make doctor`

### Advanced Commands

- List scenarios: `make list-scenarios`
- Run SRE demo: `make sre-demo SCENARIO=platform/bad-rollout`
- Run engineering demo: `make eng-up SCENARIO=...`
- Validate conformance: `make validate-scenarios`

## Related Docs

- [Demo Guide](docs/DEMO_GUIDE.md)
- [SRE Runbook](demo/sre/docs/RUNBOOK.md)
- [Engineering Runbook](demo/engineering/docs/RUNBOOK.md)

---