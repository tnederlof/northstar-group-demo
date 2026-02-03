# Northstar Group Demo

## Overview

Dual-track demo repository showcasing Warp workflows. Two runtimes (SRE: Kubernetes, Engineering: CI+Compose) sharing one contract.

## Quick Navigation

- **Application code**: `fider/` (see `fider/AGENTS.md`)
- **Demo harness**: `demo/` (see `demo/AGENTS.md`)
- **Narrative docs**: `docs/`

## Key Invariants

- Single source of truth: `demo/shared/contract/`
- URL contract: `http://<slug>.localhost:8080`
- Scenario paths: always `<track>/<slug>` (2 levels)
- Scenarios are manifest-driven: `scenario.json` required

## Common Tasks

- List scenarios: `make list-scenarios`
- Run SRE demo: `make sre-demo SCENARIO=platform/bad-rollout`
- Run engineering demo: `make eng-ci SCENARIO=...` && `make eng-up SCENARIO=...`
- Validate conformance: `make validate-scenarios`

## Related Docs

- [Demo Guide](docs/DEMO_GUIDE.md)
- [SRE Runbook](demo/sre/docs/RUNBOOK.md)
- [Engineering Runbook](demo/engineering/docs/RUNBOOK.md)

---