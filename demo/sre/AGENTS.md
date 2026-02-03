# SRE Track

## Overview

Kubernetes-first demos. Configuration-driven breakage.

## Runtime

- **kind cluster** with Envoy Gateway
- **Hostname routing** via Gateway API HTTPRoutes
- **Pre-built image**: `ghcr.io/tnederlof/northstar-group-demo:base`

## Manifest Layering

- **Base**: `base/` (shared Gateway, Deployment, Service)
- **Overlays**: `scenarios/<track>/<slug>/` (scenario-specific patches)

## Common Tasks

- Setup: `make sre-setup`
- Cluster up: `make sre-cluster-up`
- Cluster down: `make sre-cluster-down`
- Deploy demo: `make sre-demo SCENARIO=<path>`
- Verify: `make sre-verify SCENARIO=<path>`
- Health check: `make sre-health SCENARIO=<path>`
- Reset scenario: `make sre-reset SCENARIO=<path>`
- Teardown all: `make sre-teardown-all`

## Key Directories

- `kind/` - kind cluster configuration
- `base/` - base Kubernetes manifests
- `scenarios/` - per-scenario overlays
- `scripts/` - automation scripts
- `images/` - Dockerfile and image build context
- `docs/` - SRE runbook
