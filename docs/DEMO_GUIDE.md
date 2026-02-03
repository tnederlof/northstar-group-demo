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

**Run Command**:
```bash
make sre-demo SCENARIO=platform/bad-rollout
```

### Platform/Traffic Split
**Duration**: 12 minutes
**Difficulty**: Medium
**Persona**: Platform Engineer (Alex)

**Scenario**: Demonstrates progressive delivery with canary deployments using Gateway API.

**Key Learning Points**:
- Canary deployment patterns
- Traffic routing with Gateway API
- Gradual rollout strategies
- Monitoring traffic distribution

**Run Command**:
```bash
make sre-demo SCENARIO=platform/traffic-split
```

### Platform/Config Change
**Duration**: 10 minutes
**Difficulty**: Easy
**Persona**: Platform Engineer (Alex)

**Scenario**: ConfigMap changes causing pod restarts and potential downtime.

**Key Learning Points**:
- ConfigMap vs Secret management
- Rolling updates for config changes
- Zero-downtime deployments
- Config validation strategies

**Run Command**:
```bash
make sre-demo SCENARIO=platform/config-change
```

### Platform/Scale Event
**Duration**: 15 minutes
**Difficulty**: Hard
**Persona**: Platform Engineer (Alex)

**Scenario**: Horizontal Pod Autoscaling under load, resource limits and quotas.

**Key Learning Points**:
- HPA configuration
- Resource requests and limits
- Load testing strategies
- Capacity planning

**Run Command**:
```bash
make sre-demo SCENARIO=platform/scale-event
```

## Engineering Track Scenarios

### Backend/API Regression
**Duration**: 12 minutes
**Difficulty**: Easy
**Persona**: Backend Engineer (Sarah or Marcus)

**Scenario**: Missing null check in API handler causes 500 errors.

**Key Learning Points**:
- Defensive programming
- Error handling patterns
- Unit testing best practices
- Code review importance

**Run Command**:
```bash
make eng-up SCENARIO=backend/api-regression
```

### Frontend/Error Boundary
**Duration**: 12 minutes
**Difficulty**: Medium
**Persona**: Frontend Engineer (Jennifer)

**Scenario**: Missing React error boundary causes component crashes.

**Key Learning Points**:
- React error boundaries
- Graceful error handling
- User experience during failures
- Frontend testing with Playwright

**Run Command**:
```bash
make eng-up SCENARIO=frontend/error-boundary
```

### Backend/Migration Conflict
**Duration**: 15 minutes
**Difficulty**: Medium
**Persona**: Backend Engineer (Marcus)

**Scenario**: Duplicate migration numbers cause startup failures.

**Key Learning Points**:
- Database migration best practices
- Migration ordering and dependencies
- Handling merge conflicts
- Schema evolution strategies

**Run Command**:
```bash
make eng-up SCENARIO=backend/migration-conflict
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

**Run Command**:
```bash
make eng-up SCENARIO=backend/feature-flag-rollout
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
kubectl get pods -n fider
kubectl describe deployment fider -n fider
kubectl logs -n fider deployment/fider
kubectl rollout undo deployment/fider -n fider
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
make eng-ci SCENARIO=backend/api-regression
make eng-up SCENARIO=backend/api-regression
make eng-sniff SCENARIO=backend/api-regression
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
make eng-up SCENARIO=frontend/error-boundary
E2E=true make eng-ci SCENARIO=frontend/error-boundary
```

## Troubleshooting Common Issues

### SRE Track

#### Cluster Won't Start
```bash
# Delete and recreate
make sre-teardown-all
make sre-setup
make sre-cluster-up
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
make sre-reset SCENARIO=platform/bad-rollout
make sre-demo SCENARIO=platform/bad-rollout
```

### Engineering Track

#### App Won't Start
```bash
# Check containers
docker compose ps
docker compose logs

# Restart
make eng-down SCENARIO=backend/api-regression
make eng-up SCENARIO=backend/api-regression
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
make eng-ci SCENARIO=backend/api-regression
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
