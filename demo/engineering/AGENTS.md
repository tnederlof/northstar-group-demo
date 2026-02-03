# Engineering Track

## Overview

Code-first demos. Git worktrees + CI + Compose.

## Runtime

- **Traefik edge** proxy on port 8080
- **Per-scenario compose** with dedicated worktree
- **Dual networks**: `northstar-demo` (shared) + scenario-specific (private)

## Worktree Lifecycle

- `make eng-scenario-init SCENARIO=<path>` - initialize worktree
- `make eng-scenario-reset SCENARIO=<path>` - reset to initial state
- `make eng-scenario-clean SCENARIO=<path>` - remove worktree

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
