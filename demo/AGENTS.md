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

## Common Tasks

- List scenarios: `make list-scenarios`
- Describe scenario: `make describe-scenario SCENARIO=<path>`
- Validate manifests: `make validate-scenarios`

## Track-Specific Docs

- SRE Track: `sre/AGENTS.md`
- Engineering Track: `engineering/AGENTS.md`
