# Engineering Track

## Overview

Code-first demos. Git worktrees + CI + Compose.

## Runtime

- **Traefik edge** proxy on port 8082 (not 8080)
- **Per-scenario compose** with dedicated worktree
- **Dual networks**: `northstar-demo` (shared) + scenario-specific (private)
- **URL pattern**: `http://<slug>.localhost:8082`

## Worktree Lifecycle (Patch-Based Model)

- `democtl run <track>/<slug>` - create worktree from base_ref + apply broken patches
- `democtl reset <track>/<slug>` - reset worktree to broken baseline (base + broken patches)
- `democtl solve <track>/<slug>` - reset worktree to solved baseline (base + solved patches)
- Worktrees created at: `demo/engineering/scenarios/<track>/<slug>/worktree/`
- Workshop commits: stay on local `ws/<track>/<slug>` branch
- Patches stored in: `patches/broken/` and `patches/solved/` within each scenario
- All patches MUST only modify files under `fider/`
- Reset operations are intentionally destructive (discard uncommitted changes)

## Common Tasks

- Setup: `make eng-setup`
- Edge proxy up: `make eng-edge-up`
- Edge proxy down: `make eng-edge-down`
- Test deps up: `make eng-testdeps-up`
- Test deps down: `make eng-testdeps-down`
- Run CI: `make eng-ci SCENARIO=<path>`
- Start app: `make eng-up SCENARIO=<path>`
- Sniff logs: `make eng-sniff SCENARIO=<path>`
- Stop app: `make eng-down SCENARIO=<path>`

## Key Directories

- `compose/` - Docker Compose files (edge, testdeps)
- `scenarios/` - per-scenario configurations
- `scripts/` - automation scripts (worktree.sh, check-prereqs.sh)
- `docs/` - Engineering runbook
