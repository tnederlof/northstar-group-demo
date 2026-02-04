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
- `git.maintenance_branch` - mutable branch for updates (e.g., `scenario/backend/ui-regression`)
- `git.broken_ref` - stable tag for broken baseline (e.g., `scenario/backend/ui-regression/broken`)
- `git.solved_ref` - stable tag for solved baseline (e.g., `scenario/backend/ui-regression/solved`)
- `git.work_branch` - local branch for workshop commits (e.g., `ws/backend/ui-regression`)

## Common Tasks

- List scenarios: `make list-scenarios`
- Describe scenario: `make describe-scenario SCENARIO=<path>`
- Validate manifests: `make validate-scenarios`

## Track-Specific Docs

- SRE Track: `sre/AGENTS.md`
- Engineering Track: `engineering/AGENTS.md`
