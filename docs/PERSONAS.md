# Personas

This document describes the fictional personas used throughout the demo scenarios. Each persona represents a different role and focuses on scenarios relevant to their day-to-day work.

## Platform Engineer: Alex

**Role**: Site Reliability Engineer / Platform Engineer
**Experience**: 5 years
**Track**: SRE
**Responsibilities**:
- Managing Kubernetes clusters
- Deploying and scaling applications
- Incident response and troubleshooting
- Infrastructure automation and GitOps

### Alex's Typical Day

Alex starts the morning checking cluster health and reviewing alerts. They deploy new versions of services using GitOps workflows, manage traffic routing for canary deployments, and respond to incidents when services degrade. Alex works closely with development teams to ensure applications run reliably in production.

### Scenarios for Alex

| Scenario | Path | Focus |
|----------|------|-------|
| Bad Rollout | `platform/bad-rollout` | Deployment failure and rollback |
| Resource Exhaustion | `platform/resource-exhaustion` | Memory limits and OOMKilled pods |
| Network Isolation | `platform/network-isolation` | NetworkPolicy debugging |
| Missing Metrics | `platform/missing-metrics` | ServiceMonitor and observability |

### Alex's Toolbox

- `kubectl` - Kubernetes CLI
- `kind` - Local Kubernetes clusters
- Gateway API - Traffic management
- Envoy Gateway - Ingress controller
- Grafana - Monitoring dashboards
- Prometheus - Metrics collection

### What Alex Cares About

- **Reliability**: Services must stay up
- **Observability**: Need clear metrics and logs
- **Automation**: Reduce manual operations
- **Safety**: Deployments should be reversible
- **Performance**: Efficient resource utilization

## Backend Engineer: Sarah

**Role**: Senior Backend Engineer
**Experience**: 7 years
**Track**: Engineering
**Responsibilities**:
- Designing and implementing APIs
- Writing business logic in Go
- Code reviews and mentoring
- Performance optimization

### Sarah's Typical Day

Sarah spends her mornings reviewing pull requests and pair programming with junior engineers. Afternoons are for feature development, writing tests, and investigating production issues. She's passionate about clean code and test-driven development.

### Scenarios for Sarah

| Scenario | Path | Focus |
|----------|------|-------|
| UI Regression | `backend/ui-regression` | Null pointer bugs and defensive coding |
| Feature Flag Rollout | `backend/feature-flag-rollout` | Feature flag inverted logic |

### Sarah's Toolbox

- Go 1.24+ - Primary language
- `golangci-lint` - Code linting
- `go test` - Unit and integration testing
- Docker Compose - Local development
- `curl` / Postman - API testing
- Git + GitHub - Version control and collaboration

### What Sarah Cares About

- **Code Quality**: Clean, maintainable code
- **Testing**: Comprehensive test coverage
- **Performance**: Fast, efficient APIs
- **Documentation**: Clear API contracts
- **Security**: Input validation and safety

## Backend Engineer: Marcus

**Role**: Mid-Level Backend Engineer
**Experience**: 3 years
**Track**: Engineering
**Responsibilities**:
- Implementing features
- Database schema evolution
- Debugging production issues
- Writing and running migrations

### Marcus's Typical Day

Marcus works on feature tickets, often involving database changes. He carefully tests migrations locally before deploying, writes unit tests for new code, and participates in code reviews. He's learning best practices for schema evolution and backward compatibility.

### Scenarios for Marcus

| Scenario | Path | Focus |
|----------|------|-------|
| UI Regression | `backend/ui-regression` | Bug fixing and testing |

### Marcus's Toolbox

- Go - Backend language
- PostgreSQL - Database
- SQL migrations - Schema management
- Docker Compose - Local environment
- `psql` - Database CLI
- Git worktrees - Isolated development

### What Marcus Cares About

- **Data Integrity**: Migrations must be safe
- **Compatibility**: No breaking changes
- **Learning**: Growing skills and best practices
- **Collaboration**: Working well with team
- **Clarity**: Understanding requirements

## Frontend Engineer: Jennifer

**Role**: Frontend Engineer
**Experience**: 4 years
**Track**: Engineering
**Responsibilities**:
- Building React components
- Implementing UI/UX designs
- Writing end-to-end tests
- Accessibility and performance

### Jennifer's Typical Day

Jennifer starts by reviewing design mockups and user stories. She builds React components, writes Playwright tests for critical user flows, and ensures the application works across browsers. She's focused on creating delightful user experiences even when errors occur.

### Scenarios for Jennifer

| Scenario | Path | Focus |
|----------|------|-------|
| Missing Fallback | `frontend/missing-fallback` | React error boundary and graceful degradation |

### Jennifer's Toolbox

- React 18 + TypeScript - Frontend framework
- Playwright - End-to-end testing
- `eslint` + `prettier` - Code quality
- Chrome DevTools - Debugging
- Webpack / Vite - Build tools
- npm - Package management

### What Jennifer Cares About

- **User Experience**: Graceful error handling
- **Accessibility**: Inclusive design
- **Performance**: Fast page loads
- **Testing**: Reliable E2E tests
- **Modern Practices**: Latest React patterns

## Mapping Personas to Demo Scenarios

### By Experience Level

**Junior Engineers** (1-2 years):
- Start with: UI Regression, Feature Flag Rollout
- Focus: Basic debugging, testing, code review

**Mid-Level Engineers** (3-5 years):
- Recommended: Missing Fallback, Network Isolation
- Focus: Complex debugging, system understanding

**Senior Engineers** (5+ years):
- Advanced: Bad Rollout, Resource Exhaustion, Missing Metrics
- Focus: Architecture, incident response, optimization

### By Role

**Application Developers**:
- Backend: UI Regression, Feature Flag Rollout
- Frontend: Missing Fallback

**Infrastructure/Platform**:
- All SRE Track scenarios
- Focus on Kubernetes, Gateway API, deployments

**QA/Test Engineers**:
- Engineering track with emphasis on CI checks
- E2E testing scenarios with Playwright

**Engineering Managers**:
- Any scenario, focus on process and collaboration
- PR workflows and CI/CD practices

## Using Personas in Presentations

### Storytelling Approach

When presenting, adopt the persona's perspective:

**Example for Alex**:
> "Alex gets paged at 2 AM because the Fider service is returning 500 errors. Using kubectl, Alex quickly diagnoses that the latest deployment introduced a bug..."

**Example for Sarah**:
> "Sarah is reviewing a PR from a junior engineer. She notices the handler is missing a null check. Let's see what happens when this code reaches production and causes UI errors..."

### Building Empathy

Help your audience connect with the persona:
- Explain their motivations and pressures
- Show realistic workflows and tools
- Highlight common mistakes they might make
- Demonstrate how they learn and improve

## Creating Custom Personas

Want to add your own persona? Consider:

1. **Role and Responsibilities**: What do they do day-to-day?
2. **Experience Level**: Junior, mid, senior, staff?
3. **Tech Stack**: What tools do they use?
4. **Pain Points**: What challenges do they face?
5. **Learning Goals**: What do they want to improve?

See `docs/CONTRIBUTING.md` for guidance on adding new scenarios for custom personas.
