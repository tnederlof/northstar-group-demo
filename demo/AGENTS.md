# Demo Harness

## Overview

Two demo tracks sharing one contract.

## Shared Contract

- **Versions**: `shared/contract/versions.env`
- **Env template**: `shared/contract/fider.env.example`
- **Seed data**: `shared/northstar/seed.sql`
- **Runtime env**: `shared/scripts/render-env.sh`

## Scenario Manifest Schema

Required keys in `scenario.json`:
- `track` - sre or engineering
- `slug` - unique scenario identifier (e.g., `platform/bad-rollout`)
- `title` - human-readable name
- `type` - scenario type/category
- `url_host` - hostname for routing (e.g., `bad-rollout`)
- `seed` - seed data to use
- `reset_strategy` - how to reset the scenario

Engineering scenarios also require:
- `git.base_ref` - pinned commit SHA (40 characters)
- `git.broken_patches_dir` - directory containing broken state patches (default: `patches/broken`)
- `git.solved_patches_dir` - directory containing solved state patches (default: `patches/solved`)
- `git.work_branch` - local branch for workshop commits (e.g., `ws/backend/ui-regression`)

## Scenario Maintenance

When the `fider/` codebase changes significantly, scenarios may need rebasing:

```bash
# Use migration/rebase tools
democtl rebase-scenario-patches --to <new_base_ref> --all

# Or manually re-record patches on new base
# See AGENTS.md in repo root for detailed steps
```

After updates, validate:
```bash
democtl validate-scenarios --strict
democtl validate-patches
```

## Common Tasks

- List scenarios: `make list-scenarios`
- Describe scenario: `make describe-scenario SCENARIO=<path>`
- Validate manifests: `make validate-scenarios`

## Track-Specific Docs

- SRE Track: `sre/AGENTS.md`
- Engineering Track: `engineering/AGENTS.md`
