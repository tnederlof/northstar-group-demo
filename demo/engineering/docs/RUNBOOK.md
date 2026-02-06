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

The Engineering track uses Git worktrees to isolate scenarios. `democtl run` automatically creates worktrees:

```bash
democtl run backend/ui-regression
```

This creates:
- A Git worktree in `demo/engineering/scenarios/backend/ui-regression/worktree/`
- Checks out from the `scenario/backend/ui-regression/broken` tag
- Creates a local workshop branch `ws/backend/ui-regression`
- Starts the Docker Compose runtime

### Working in Worktrees

```bash
# Navigate to the worktree
cd demo/engineering/scenarios/backend/ui-regression/worktree/fider/

# Make changes, run tests, etc.
# Your main working directory remains unchanged
# Commits stay on your local ws/backend/ui-regression branch
```

### Workshop Commands

```bash
# Reset to broken baseline (start over)
democtl reset backend/ui-regression

# Jump to solved baseline (escape hatch)
democtl fix-it backend/ui-regression
```

## CI Checks

For maintainers: Run CI checks in the worktree manually:

```bash
cd demo/engineering/scenarios/backend/ui-regression/worktree/fider/

# Run linting
golangci-lint run
npx eslint .

# Run tests
go test ./...
npm test
```

## Edge Proxy

The Traefik edge proxy is started automatically by `democtl run`. No manual start required.

## Application Runtime

### Start the Application

```bash
democtl run backend/ui-regression
```

Visit: **http://ui-regression.localhost:8082**

### Verify the Deployment

Run scenario-specific checks:

```bash
# Run all checks
democtl checks verify backend/ui-regression

# Run checks for a specific stage
democtl checks verify backend/ui-regression --stage broken
democtl checks verify backend/ui-regression --stage fixed

# Run only specific check types
democtl checks verify backend/ui-regression --only playwright
```

### Monitor Application Logs

Use docker compose:

```bash
cd demo/engineering/scenarios/backend/ui-regression
docker compose logs -f
```

### Stop the Application

```bash
democtl reset backend/ui-regression
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
| UI Regression | `backend/ui-regression` | Backend debugging, null checks |

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
