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

The SRE runtime (kind cluster + Envoy Gateway) is created automatically when you first run a scenario. No manual setup required.

## Quick Start

To run a demo scenario:

```bash
democtl run platform/bad-rollout
```

This command:
- Ensures kind cluster and Envoy Gateway are running
- Creates namespace
- Applies Kubernetes manifests
- Runs verification checks

## Pre-Demo Verification

Before presenting, verify the deployment:

```bash
democtl checks verify platform/bad-rollout
```

For "broken" scenarios, this verifies the breakage is present (not that the app is healthy).

## During the Demo

### Monitor System Health

Run health checks during the demo:

```bash
democtl checks health platform/bad-rollout
```

### Common Scenarios

#### Bad Rollout Scenario
Demonstrates a broken deployment and rollback:

```bash
# Deploy the scenario
democtl run platform/bad-rollout

# Show the problem (during presentation)
democtl checks health platform/bad-rollout

# Demonstrate rollback (during presentation)
kubctl rollout undo deployment/fider -n demo-bad-rollout
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
democtl reset platform/bad-rollout
```

This deletes and recreates the namespace.

## Cleanup

### Clean Up Everything

To remove all scenarios and the cluster:

```bash
democtl reset-all --force
```

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
| Healthy | `platform/healthy` | Working baseline for verification |
| Bad Rollout | `platform/bad-rollout` | Deployment failures, rollback |
| Resource Exhaustion | `platform/resource-exhaustion` | OOMKilled pods, memory limits |
| Network Isolation | `platform/network-isolation` | NetworkPolicy blocking database |
| Missing Metrics | `platform/missing-metrics` | ServiceMonitor misconfiguration |

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
