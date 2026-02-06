# Demo Guide

This guide helps presenters choose and run demo scenarios effectively based on their audience and objectives.

## Overview

The Northstar Group Demo provides two distinct tracks for demonstrating different aspects of software development and operations:

### SRE Track
**Audience**: Infrastructure engineers, platform teams, DevOps practitioners
**Focus**: Cloud-native infrastructure, Kubernetes, Gateway API, GitOps workflows
**Runtime**: kind cluster with Envoy Gateway
**Complexity**: Medium to High

### Engineering Track
**Audience**: Application developers, QA engineers, engineering managers
**Focus**: Code quality, debugging, testing, CI/CD workflows
**Runtime**: Docker Compose with Traefik
**Complexity**: Low to Medium

## Choosing a Track

Use this decision tree to select the appropriate track:

```
Is your audience primarily focused on:

├─ Infrastructure/Platform? → SRE Track
│  ├─ Deployment strategies
│  ├─ Traffic management
│  ├─ Kubernetes operations
│  └─ Incident response
│
└─ Application Development? → Engineering Track
   ├─ Bug fixing and debugging
   ├─ Code review workflows
   ├─ Testing practices
   └─ Local development
```

## SRE Track Scenarios

### Platform/Bad Rollout
**Duration**: 15 minutes
**Difficulty**: Medium
**Persona**: Platform Engineer (Alex)

**Scenario**: A bad deployment causes service degradation. Demonstrates rollback procedures and incident response.

**Key Learning Points**:
- Kubernetes deployment strategies
- Rollback procedures
- Pod health checks
- Monitoring and observability

**Run**:
```bash
democtl run platform/bad-rollout
```

### Platform/Resource Exhaustion
**Duration**: 15 minutes
**Difficulty**: Hard
**Persona**: Platform Engineer (Alex)

**Scenario**: Memory limits too low causing OOMKilled pods and service degradation.

**Key Learning Points**:
- Resource requests and limits
- Memory profiling and debugging
- Pod restart patterns
- Capacity planning

**Run**:
```bash
democtl run platform/resource-exhaustion
```

### Platform/Network Isolation
**Duration**: 12 minutes
**Difficulty**: Medium
**Persona**: Platform Engineer (Alex)

**Scenario**: NetworkPolicy blocking egress to database causing connection failures.

**Key Learning Points**:
- NetworkPolicy configuration
- Egress and ingress rules
- Debugging network connectivity
- Security vs accessibility tradeoffs

**Run**:
```bash
democtl run platform/network-isolation
```

### Platform/Missing Metrics
**Duration**: 10 minutes
**Difficulty**: Medium
**Persona**: Platform Engineer (Alex)

**Scenario**: Metrics port not exposed and ServiceMonitor selector mismatch causing observability gaps.

**Key Learning Points**:
- Prometheus ServiceMonitor configuration
- Metrics endpoint exposure
- Label selector troubleshooting
- Observability best practices

**Run**:
```bash
democtl run platform/missing-metrics
```

## Engineering Track Scenarios

### Backend/UI Regression
**Duration**: 12 minutes
**Difficulty**: Easy
**Persona**: Backend Engineer (Sarah or Marcus)

**Scenario**: Missing null check in API handler causes 500 errors in the UI.

**Key Learning Points**:
- Defensive programming
- Error handling patterns
- Unit testing best practices
- Code review importance

**Run**:
```bash
democtl run backend/ui-regression
```

### Frontend/Missing Fallback
**Duration**: 12 minutes
**Difficulty**: Medium
**Persona**: Frontend Engineer (Jennifer)

**Scenario**: Removed React error boundary causes UI white-screens on component errors.

**Key Learning Points**:
- React error boundaries
- Graceful error handling
- User experience during failures
- Frontend testing with Playwright

**Run**:
```bash
democtl run frontend/missing-fallback
```

### Backend/Feature Flag Rollout
**Duration**: 10 minutes
**Difficulty**: Easy
**Persona**: Backend Engineer (Sarah)

**Scenario**: New feature behind feature flag demonstrates progressive delivery.

**Key Learning Points**:
- Feature flag patterns
- Environment-based configuration
- Progressive feature rollout
- A/B testing foundations

**Run**:
```bash
democtl run backend/feature-flag-rollout
```

## Demo Flow Per Persona

### Platform Engineer (Alex) - SRE Track

**Typical Day**: Managing infrastructure, responding to incidents, deploying services

**Demo Flow** (15 min):
1. **Setup** (2 min): Show cluster state, explain infrastructure
2. **Problem** (3 min): Deploy bad change, observe symptoms
3. **Investigation** (5 min): Use kubectl to diagnose
4. **Resolution** (3 min): Rollback or fix
5. **Lessons** (2 min): Preventive measures, best practices

**Key Commands**:
```bash
# Use the namespace matching the scenario (demo-<slug>)
kubectl get pods -n demo-bad-rollout
kubectl describe deployment fider -n demo-bad-rollout
kubectl logs -n demo-bad-rollout deployment/fider
kubectl rollout undo deployment/fider -n demo-bad-rollout
```

### Backend Engineers (Sarah, Marcus) - Engineering Track

**Typical Day**: Writing code, fixing bugs, reviewing PRs, running tests

**Demo Flow** (12 min):
1. **Context** (2 min): Show PR, explain the bug
2. **Reproduction** (2 min): Run app, demonstrate bug
3. **Investigation** (4 min): Navigate code, show failing test
4. **Fix** (2 min): Apply fix from PR
5. **Verification** (2 min): Run CI checks, verify fix

**Key Commands**:
```bash
democtl run backend/ui-regression
democtl reset backend/ui-regression
democtl fix-it backend/ui-regression
```

### Frontend Engineer (Jennifer) - Engineering Track

**Typical Day**: Building UI components, fixing React bugs, writing E2E tests

**Demo Flow** (12 min):
1. **Context** (2 min): Show component, explain error
2. **Reproduction** (2 min): Navigate to component, trigger crash
3. **Investigation** (4 min): Show error in console, locate code
4. **Fix** (2 min): Add error boundary
5. **Verification** (2 min): Run E2E tests

**Key Commands**:
```bash
democtl run frontend/missing-fallback
democtl reset frontend/missing-fallback
democtl fix-it frontend/missing-fallback
```

## Troubleshooting Common Issues

### SRE Track

#### Cluster Won't Start
```bash
# Delete and recreate
kind delete cluster --name fider-demo
democtl run platform/healthy
```

#### Gateway Not Responding
```bash
# Check Gateway status
kubectl get gateway -A
kubectl logs -n envoy-gateway-system deployment/envoy-gateway
```

#### Scenario Won't Deploy
```bash
# Reset scenario
democtl reset platform/bad-rollout
democtl run platform/bad-rollout
```

### Engineering Track

#### App Won't Start
```bash
# Check containers
docker compose ps
docker compose logs

# Restart
democtl reset backend/ui-regression
democtl run backend/ui-regression
```

#### Port Conflicts
```bash
# Kill processes on conflicting ports
lsof -ti:5432 | xargs kill -9  # PostgreSQL
lsof -ti:3000 | xargs kill -9  # Frontend
```

#### CI Checks Fail
```bash
# Clean and retry
go clean -cache
npm clean-install
democtl run backend/ui-regression
```

## Best Practices for Presenters

### Preparation
- **Practice** the demo at least once before presenting
- **Check prerequisites** (Docker running, cluster up, etc.)
- **Prepare backup** terminal windows
- **Have documentation** open and ready
- **Test all commands** in sequence

### During Presentation
- **Explain context** before showing technical details
- **Narrate your actions** - don't just type silently
- **Engage the audience** - ask questions, pause for feedback
- **Handle errors gracefully** - use them as teaching moments
- **Watch your time** - stick to the allocated duration

### After Demo
- **Clean up resources** - use teardown commands
- **Gather feedback** - ask what worked and what didn't
- **Document issues** - note any problems for next time
- **Share resources** - provide links to runbooks and docs

## Customizing Scenarios

See `docs/CONTRIBUTING.md` for guidance on:
- Adding new scenarios
- Modifying existing scenarios
- Creating custom personas
- Testing changes

## Additional Resources

- **Architecture**: `docs/ARCHITECTURE.md`
- **Personas**: `docs/PERSONAS.md`
- **SRE Runbook**: `demo/sre/docs/RUNBOOK.md`
- **Engineering Runbook**: `demo/engineering/docs/RUNBOOK.md`
- **Contributing**: `docs/CONTRIBUTING.md`
- **Contract**: `demo/docs/CONTRACT.md`
