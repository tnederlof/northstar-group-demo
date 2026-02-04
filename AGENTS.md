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

### Engineering Scenario Git Ref Contract

- **Maintenance branch**: `scenario/<track>/<slug>` (mutable, for rebasing)
- **Broken baseline tag**: `scenario/<track>/<slug>/broken` (stable, immutable)
- **Solved baseline tag**: `scenario/<track>/<slug>/solved` (stable, immutable)
- **Workshop branch**: `ws/<track>/<slug>` (local only, per-participant)
- **Worktree location**: `demo/engineering/scenarios/<track>/<slug>/worktree/`
- **Automation relies on tags**, not branches - tags are the single source of truth
- Workshop commits stay on local `ws/` branches and don't affect baselines

### Keeping Scenarios Current

**IMPORTANT**: When `main` changes, scenario branches must be rebased to stay current:

```bash
# For each scenario:
git checkout scenario/<track>/<slug>
git rebase main

# Update tags to point to rebased commits (broken=HEAD~1, solved=HEAD)
git tag -f -a scenario/<track>/<slug>/broken HEAD~1 -m "<track>/<slug>: broken baseline"
git tag -f -a scenario/<track>/<slug>/solved HEAD -m "<track>/<slug>: solved baseline"

# Force push (tags are automation API, this is expected)
git push origin scenario/<track>/<slug> --force-with-lease
git push origin --tags --force
```

This is required after:
- Infrastructure changes to `main`
- Removing/adding scenarios
- Updates to shared code or configurations

## Common Tasks

### Golden Path (Recommended)

- Setup: `make setup`
- Verify: `make verify`
- Run any scenario: `make run SCENARIO=<track>/<slug>`
- Reset: `make reset SCENARIO=<track>/<slug>` (returns to broken baseline)
- Fix-it: `make fix-it SCENARIO=<track>/<slug>` (jump to solved baseline, engineering only)
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