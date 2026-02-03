# SRE Track Runbook

This runbook provides step-by-step instructions for presenters running the SRE track demonstration.

## Prerequisites

Ensure the following tools are installed before beginning:

- **Docker Desktop**: For container runtime
- **kind**: Kubernetes in Docker (for local clusters)
- **kubectl**: Kubernetes command-line tool
- **make**: Build automation tool

Verify installations:
```bash
docker --version
kind --version
kubectl version --client
make --version
```

## One-Time Setup

Perform these steps once before your first demo:

```bash
# Initialize SRE environment
make sre-setup

# Create the kind cluster
make sre-cluster-up
```

This will:
- Install Envoy Gateway
- Configure Gateway API resources
- Set up namespaces and RBAC
- Deploy base infrastructure

## Quick Start

To run a complete demo scenario:

```bash
make sre-demo SCENARIO=platform/bad-rollout
```

This single command handles:
- Pre-demo verification
- Scenario deployment
- Health checks
- Demo readiness confirmation

## Pre-Demo Verification

Before starting your presentation, verify the environment is ready:

```bash
make sre-verify SCENARIO=platform/bad-rollout
```

This checks:
- Cluster connectivity
- Gateway API resources
- Scenario prerequisites
- Service health

Expected output: All checks pass with green ✓ marks.

## During the Demo

### Monitor System Health

Run health checks during the demonstration:

```bash
make sre-health SCENARIO=platform/bad-rollout
```

This shows:
- Pod status
- Service endpoints
- Gateway routes
- Current traffic distribution

### Common Scenarios

#### Bad Rollout Scenario
Demonstrates a broken deployment and rollback:

```bash
# Deploy the scenario
make sre-demo SCENARIO=platform/bad-rollout

# Show the problem (during presentation)
make sre-health SCENARIO=platform/bad-rollout

# Demonstrate rollback (during presentation)
kubectl rollout undo deployment/fider -n fider
```

#### Traffic Splitting Scenario
Demonstrates canary deployments:

```bash
# Deploy the scenario
make sre-demo SCENARIO=platform/traffic-split

# Show traffic distribution
make sre-health SCENARIO=platform/traffic-split

# Adjust weights (example)
kubectl patch httproute fider-route -n fider --type merge -p '{"spec":{"rules":[{"backendRefs":[{"name":"fider-v1","weight":90},{"name":"fider-v2","weight":10}]}]}}'
```

## Reset Workflow

To reset a scenario between presentations:

```bash
make sre-reset SCENARIO=platform/bad-rollout
```

This:
- Removes scenario-specific resources
- Resets to baseline state
- Preserves cluster and core infrastructure

## Cleanup

### Clean Up a Specific Scenario

After the demo, clean up resources:

```bash
make sre-teardown SCENARIO=platform/bad-rollout
```

### Full Environment Teardown

To completely remove the kind cluster:

```bash
make sre-teardown-all
```

This deletes:
- The entire kind cluster
- All deployed resources
- Local state files

## Troubleshooting

### Cluster Not Responding

```bash
# Check cluster status
kubectl cluster-info

# If cluster is down, recreate it
make sre-cluster-down
make sre-cluster-up
```

### Gateway Not Ready

```bash
# Check Gateway status
kubectl get gateway -n envoy-gateway-system

# Restart Gateway controller if needed
kubectl rollout restart deployment envoy-gateway -n envoy-gateway-system
```

### Pods Not Starting

```bash
# Check pod events
kubectl get events -n fider --sort-by='.lastTimestamp'

# Check pod logs
kubectl logs -n fider deployment/fider --tail=50
```

### Image Pull Failures

Ensure images are loaded into kind:

```bash
# Load image into kind cluster
kind load docker-image <image-name>:latest --name northstar-demo
```

## Demo Flow Template

Recommended flow for presentations:

1. **Introduction** (2 min)
   - Explain the scenario objective
   - Show initial healthy state

2. **Problem Introduction** (3 min)
   - Deploy the broken change
   - Observe symptoms
   - Show monitoring/health checks

3. **Investigation** (5 min)
   - Examine logs: `kubectl logs`
   - Check events: `kubectl get events`
   - Inspect resources: `kubectl describe`

4. **Resolution** (3 min)
   - Demonstrate fix (rollback or patch)
   - Verify recovery
   - Show healthy state

5. **Lessons Learned** (2 min)
   - Summarize key takeaways
   - Discuss prevention strategies

## Available Scenarios

| Scenario | Path | Focus Area |
|----------|------|------------|
| Bad Rollout | `platform/bad-rollout` | Deployment failures, rollback |
| Traffic Split | `platform/traffic-split` | Canary deployments, progressive delivery |
| Config Change | `platform/config-change` | ConfigMap updates, pod restarts |
| Scale Event | `platform/scale-event` | HPA, resource limits |

## Tips for Presenters

- **Practice** the demo flow at least once before presenting
- **Keep a backup** terminal window with health checks running
- **Prepare for questions** about Kubernetes concepts
- **Have the reset command** ready in case something goes wrong
- **Time yourself** - most scenarios should complete in 15 minutes
- **Engage the audience** - ask what they think might be wrong before revealing

## Additional Resources

- **Architecture Overview**: See `docs/ARCHITECTURE.md`
- **Demo Guide**: See `docs/DEMO_GUIDE.md`
- **Makefile Reference**: Run `make help` for all available commands
