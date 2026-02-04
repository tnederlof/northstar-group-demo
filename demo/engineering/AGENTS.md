# Engineering Track

## Overview

Code-first demos. Git worktrees + CI + Compose.

## Runtime

- **Traefik edge** proxy on port 8082 (not 8080)
- **Per-scenario compose** with dedicated worktree
- **Dual networks**: `northstar-demo` (shared) + scenario-specific (private)
- **URL pattern**: `http://<slug>.localhost:8082`

## Worktree Lifecycle (Git Ref Model)

- `make run SCENARIO=<path>` - create worktree from broken tag on `ws/<path>` branch
- `make reset SCENARIO=<path>` - reset worktree to broken baseline tag
- `make fix-it SCENARIO=<path>` - reset worktree to solved baseline tag (escape hatch)
- `make eng-reset-broken SCENARIO=<path>` - reset worktree only (no restart)
- Worktrees created at: `demo/engineering/scenarios/<track>/<slug>/worktree/`
- Workshop commits: stay on local `ws/<track>/<slug>` branch
- Tags are immutable: automation uses `scenario/<track>/<slug>/broken` and `/solved`
- Use `FORCE=true` to override dirty worktree warnings

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
