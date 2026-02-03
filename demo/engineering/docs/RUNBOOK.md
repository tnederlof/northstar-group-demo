# Engineering Track Runbook

This runbook provides step-by-step instructions for presenters running the Engineering track demonstration.

## Prerequisites

Ensure the following tools are installed before beginning:

- **Docker**: For container runtime
- **docker compose**: For local orchestration
- **Go**: Go 1.25+ for backend
- **Node.js**: Node 24+ and npm for frontend
- **make**: Build automation tool
- **Playwright** (optional): For end-to-end testing

Verify installations:
```bash
docker --version
docker compose version
go version
node --version
npm --version
make --version
```

Optional for E2E testing:
```bash
npx playwright --version
```

## Worktree Setup

The Engineering track uses Git worktrees to isolate scenarios. Initialize a scenario:

```bash
make eng-scenario-init SCENARIO=backend/api-regression
```

This creates:
- A Git worktree in `.worktrees/backend/api-regression`
- Checks out the scenario branch
- Prepares the environment

### Working in Worktrees

```bash
# Navigate to the worktree
cd .worktrees/backend/api-regression

# Make changes, run tests, etc.
# Your main working directory remains unchanged
```

## CI Checks

Run the same checks that CI would run:

```bash
make eng-ci SCENARIO=backend/api-regression
```

This executes:
- Linting (Go: `golangci-lint`, JS/TS: `eslint`)
- Type checking (TypeScript: `tsc`)
- Unit tests (Go: `go test`, JS: `jest`)
- Build verification

### With End-to-End Tests (Optional)

E2E tests are opt-in due to longer runtime:

```bash
E2E=true make eng-ci SCENARIO=backend/api-regression
```

This adds:
- Playwright end-to-end tests
- Integration tests
- Full application smoke tests

## Edge Proxy

The Engineering track uses Traefik as an edge proxy for routing.

### Start Edge Proxy

```bash
make eng-edge-up
```

This starts Traefik with:
- HTTP routing on port 80
- HTTPS with self-signed certificates on port 443
- Dashboard on port 8080

### Stop Edge Proxy

```bash
make eng-edge-down
```

## Application Runtime

### Start the Application

```bash
make eng-up SCENARIO=backend/api-regression
```

This starts:
- PostgreSQL database
- Fider application (backend + frontend)
- Any scenario-specific services

The application will be available at:
- **HTTP**: http://localhost
- **HTTPS**: https://localhost (self-signed cert warning expected)

### Monitor Application Logs

Use the sniff command to tail logs:

```bash
make eng-sniff SCENARIO=backend/api-regression
```

This shows real-time logs from all services.

### Stop the Application

```bash
make eng-down SCENARIO=backend/api-regression
```

## Working with Prepared PRs

Scenarios include prepared pull requests with passing CI checks.

### Viewing PR Status

Each scenario branch has:
- **Green checks**: All CI checks passing
- **Clear description**: Problem and expected solution
- **Test coverage**: Automated tests to verify fix

### Demo Flow with PRs

1. **Show the broken state** on the scenario branch
2. **Navigate to the PR** (link in scenario README)
3. **Review the changes** in the PR diff
4. **Highlight CI status** (all green ✓)
5. **Merge the PR** or cherry-pick the fix
6. **Verify the fix** with `make eng-ci`

## Making and Committing Fixes

### Standard Git Workflow in Worktrees

```bash
# In the worktree directory
cd .worktrees/backend/api-regression

# Make your changes
vim fider/app/handlers/api.go

# Run tests
make eng-ci SCENARIO=backend/api-regression

# Commit the fix
git add .
git commit -m "Fix API null pointer dereference

Added null check before accessing user object.
Fixes #123"

# Push to remote (if applicable)
git push origin HEAD
```

### Best Practices

- **Run CI checks** before committing: `make eng-ci`
- **Write descriptive commit messages** following conventional commits
- **Include test coverage** for bug fixes
- **Verify the fix** with both unit and integration tests

## Scenario Walkthroughs

### Backend API Regression

**Problem**: Missing null check causes 500 errors

```bash
# Setup
make eng-scenario-init SCENARIO=backend/api-regression
cd .worktrees/backend/api-regression

# Start app
make eng-up SCENARIO=backend/api-regression

# Reproduce the bug
curl -X GET http://localhost/api/v1/users/999

# Run CI (shows failing test)
make eng-ci SCENARIO=backend/api-regression

# Fix the bug (edit the file)
# Re-run CI
make eng-ci SCENARIO=backend/api-regression

# Cleanup
make eng-down SCENARIO=backend/api-regression
```

### Frontend Component Crash

**Problem**: Missing error boundary causes React crash

```bash
# Setup
make eng-scenario-init SCENARIO=frontend/error-boundary
cd .worktrees/frontend/error-boundary

# Start app
make eng-up SCENARIO=frontend/error-boundary

# Reproduce the bug (navigate to problematic component in browser)
open http://localhost/admin/posts

# Run CI with E2E tests
E2E=true make eng-ci SCENARIO=frontend/error-boundary

# Fix and verify
make eng-ci SCENARIO=frontend/error-boundary
```

## Troubleshooting

### Port Already in Use

```bash
# Find and kill process using port 5432 (PostgreSQL)
lsof -ti:5432 | xargs kill -9

# Or stop all docker compose services
docker compose down -v
```

### Database Connection Errors

```bash
# Restart database
docker compose restart postgres

# Reset database
make eng-reset SCENARIO=backend/api-regression
```

### Build Failures

```bash
# Clean build cache
go clean -cache
npm clean-install

# Rebuild from scratch
make eng-rebuild SCENARIO=backend/api-regression
```

### Worktree Conflicts

```bash
# Remove corrupted worktree
git worktree remove .worktrees/backend/api-regression --force

# Recreate
make eng-scenario-init SCENARIO=backend/api-regression
```

## Demo Flow Template

Recommended flow for presentations:

1. **Introduction** (2 min)
   - Explain the scenario context
   - Show the PR with green CI checks

2. **Bug Reproduction** (3 min)
   - Start the application
   - Demonstrate the bug in action
   - Show error logs/traces

3. **Code Investigation** (5 min)
   - Navigate to the problematic code
   - Explain the root cause
   - Show the failing test

4. **Fixing the Bug** (3 min)
   - Show the fix in the PR diff
   - Explain the solution approach
   - Merge or apply the fix

5. **Verification** (2 min)
   - Run CI checks (show green ✓)
   - Re-test the application
   - Confirm fix resolves the issue

## Available Scenarios

| Scenario | Path | Focus Area |
|----------|------|------------|
| API Regression | `backend/api-regression` | Backend debugging, null checks |
| Error Boundary | `frontend/error-boundary` | React error handling |
| Database Migration | `backend/migration-conflict` | Schema changes, migration ordering |
| Feature Flag | `backend/feature-flag-rollout` | Progressive feature delivery |

## Tips for Presenters

- **Test the worktree setup** before presenting
- **Have the application running** before the demo starts (saves time)
- **Keep CI output visible** to show validation
- **Prepare browser tabs** for relevant PRs and documentation
- **Practice the fix** at least once to ensure smooth presentation
- **Emphasize CI/CD** - all checks are automated and reproducible
- **Show debugging techniques** - logs, breakpoints, test-driven debugging

## Additional Resources

- **Architecture Overview**: See `docs/ARCHITECTURE.md`
- **Demo Guide**: See `docs/DEMO_GUIDE.md`
- **Makefile Reference**: Run `make help` for all available commands
- **Fider Codebase**: Familiarize yourself with `fider/` structure
