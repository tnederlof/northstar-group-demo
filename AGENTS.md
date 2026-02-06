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

### Engineering Scenario Patch Contract

- **Base commit**: Pinned `base_ref` (full 40-char SHA) in `scenario.json`
- **Broken patches**: `patches/broken/*.patch` (git format-patch output)
- **Solved patches**: `patches/solved/*.patch` (git format-patch output)
- **Workshop branch**: `ws/<track>/<slug>` (local only, per-participant)
- **Worktree location**: `demo/engineering/scenarios/<track>/<slug>/worktree/`
- **Patch scope**: All patches MUST only modify files under `fider/`
- Workshop commits stay on local `ws/` branches and don't affect baselines

### Updating Scenarios to New Base

When the `fider/` codebase changes significantly, scenarios may need rebasing:

```bash
# Use the rebase-scenario-patches command (if available)
democtl rebase-scenario-patches --to <new_base_ref> --all

# Or manually re-record patches:
# 1. Create throwaway branch at new base
git checkout -b temp-rebase <new_base_ref>

# 2. Recreate broken state, export patches
git format-patch --binary --output-directory patches/broken <new_base_ref>..HEAD

# 3. Repeat for solved state
# 4. Update scenario.json with new base_ref
```

After updating scenarios:
- Run `democtl validate-scenarios --strict`
- Run `democtl validate-patches --strict`
- Test each scenario with `democtl run` and `democtl solve`

## Git Workflow

### Infrastructure vs Application Changes

**IMPORTANT**: Infrastructure files (shared demo harness, scripts, compose files outside scenarios) must be modified in `main` only:

**Infrastructure paths (main only)**:
- `demo/engineering/compose/` - Edge proxy, shared services
- `demo/sre/base/` - Base Kubernetes manifests
- `demo/shared/` - Shared contracts
- `democtl/` - Demo CLI implementation
- `Makefile` - Top-level automation
- `docs/` - Documentation

**Application paths (scenario branches)**:
- `fider/` - Application code (via worktrees)
- `demo/engineering/scenarios/<track>/<slug>/docker-compose.yml` - Scenario-specific compose
- `demo/ui/tests/` - Test specifications

### Setting Up Git Hooks

Git hooks are automatically installed when you run `make setup`. To manually install:

```bash
./scripts/install-hooks.sh
```

The pre-commit hook blocks infrastructure changes in `scenario/*` and `ws/*` branches.

**Note**: Git hooks are local to each clone and not tracked in version control. Each contributor must run setup or install hooks manually.

## Common Tasks

### Golden Path (Demo Users)

- Build CLI: `make build-democtl`
- Setup: `democtl setup`
- Verify: `democtl verify`
- Run any scenario: `democtl run <track>/<slug>`
- Reset: `democtl reset <track>/<slug>` (returns to broken baseline)
- Solve: `democtl solve <track>/<slug>` (jump to solved baseline, engineering only)
- Clean up: `democtl reset-all --force`
- Status: `democtl doctor`

### Advanced Commands

- List scenarios: `democtl list-scenarios`
- Describe scenario: `democtl describe-scenario <track>/<slug>`
- Run checks: `democtl checks verify <track>/<slug>`
- Validate manifests: `democtl validate-scenarios`

## Related Docs

- [Demo Guide](docs/DEMO_GUIDE.md)
- [SRE Runbook](demo/sre/docs/RUNBOOK.md)
- [Engineering Runbook](demo/engineering/docs/RUNBOOK.md)

---