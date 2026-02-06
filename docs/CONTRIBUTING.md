# Contributing to Northstar Group Demo

Thank you for your interest in improving the demo! This guide will help you add new scenarios, fix bugs, and enhance the documentation.

## How to Add New Scenarios

### 1. Choose a Track

Decide whether your scenario fits the **SRE track** or **Engineering track**:

- **SRE Track**: Infrastructure, Kubernetes, deployments, scaling
- **Engineering Track**: Application code, bugs, testing, CI/CD

### 2. Create Scenario Branch and Tags

For **SRE scenarios**:
```bash
git checkout -b demo/sre/your-scenario-name main
```

For **Engineering scenarios**, use the scenario maintenance branch model:
```bash
# Create maintenance branch from main
git checkout -b scenario/<track>/<slug> main

# Make changes for broken state
# ... edit code to introduce the bug ...
git commit -m "<track>/<slug>: introduce bug"

# Tag the broken baseline
git tag -a scenario/<track>/<slug>/broken -m "<track>/<slug>: broken baseline"

# Make changes for solved state
# ... edit code to fix the bug ...
git commit -m "<track>/<slug>: fix bug"

# Tag the solved baseline
git tag -a scenario/<track>/<slug>/solved -m "<track>/<slug>: solved baseline"

# Push branch and tags
git push origin scenario/<track>/<slug>
git push origin --tags
```

### 3. Implement the Scenario

#### For SRE Scenarios

Create Kubernetes manifests in `demo/sre/scenarios/your-scenario-name/`:

```
demo/sre/scenarios/your-scenario-name/
├── README.md           # Scenario description and objectives
├── manifests/
│   ├── deployment.yaml # Your Kubernetes resources
│   ├── service.yaml
│   └── gateway.yaml
└── scripts/
    ├── setup.sh        # Pre-demo setup
    └── teardown.sh     # Cleanup
```

#### For Engineering Scenarios

Create a scenario directory and manifest:

```bash
# Create scenario directory
mkdir -p demo/engineering/scenarios/<track>/<slug>

# Create scenario.json manifest
cat > demo/engineering/scenarios/<track>/<slug>/scenario.json <<EOF
{
  "track": "<track>",
  "slug": "<slug>",
  "title": "Your Scenario Title",
  "type": "engineering",
  "url_host": "<slug>.localhost",
  "seed": true,
  "reset_strategy": "worktree-reset",
  "git": {
    "maintenance_branch": "scenario/<track>/<slug>",
    "broken_ref": "scenario/<track>/<slug>/broken",
    "solved_ref": "scenario/<track>/<slug>/solved",
    "work_branch": "ws/<track>/<slug>"
  },
  "description": "Brief description",
  "symptoms": ["Symptom 1", "Symptom 2"],
  "fix_hints": ["Hint 1", "Hint 2"],
  "checks": {
    "version": 1,
    "default_stage": "broken",
    "stages": {
      "broken": {
        "verify": [{"type": "http.get", "url": "http://<slug>.localhost:8082/_health", "expect": {"status": [500]}}]
      },
      "fixed": {
        "verify": [{"type": "http.get", "url": "http://<slug>.localhost:8082/_health", "expect": {"status": [200]}}]
      }
    }
  }
}
EOF

# Introduce code changes in fider/ on the scenario branch
vim fider/app/handlers/your_handler.go
vim fider/app/handlers/your_handler_test.go
```

**Important conformance rules**:
- Automation must rely on tags (`scenario/<track>/<slug>/broken` and `/solved`), not mutable branches
- Worktrees are created from the broken tag onto a local `ws/<track>/<slug>` branch
- The maintenance branch is where you rebase/update when `main` evolves
- Tags represent stable waypoints that `make run/reset/fix-it` depend on

### 4. Document the Scenario

Every scenario needs a README.md with:

```markdown
# Scenario: Your Scenario Name

## Persona
Who would encounter this? (e.g., Backend Engineer, Platform Engineer)

## Objective
What will the presenter demonstrate?

## Problem Description
What's broken or needs improvement?

## Key Learning Points
- Point 1
- Point 2
- Point 3

## Prerequisites
What tools or setup is needed?

## Running the Scenario
```bash
make run SCENARIO=your-scenario-name
```

## Expected Outcome
What should happen when the scenario runs successfully?

## Troubleshooting
Common issues and solutions
```

### 5. Testing Requirements

All scenarios must be tested before submission:

#### SRE Scenarios

```bash
# Verify the scenario can be deployed
make run SCENARIO=platform/your-scenario-name

# Verify health checks pass
make health SCENARIO=platform/your-scenario-name

# Verify teardown works
make reset SCENARIO=platform/your-scenario-name
```

#### Engineering Scenarios

```bash
# Verify the scenario runs and creates worktree from broken tag
make run SCENARIO=<track>/<slug>

# Verify reset to broken works
make reset SCENARIO=<track>/<slug>

# Verify fix-it jumps to solved tag
make fix-it SCENARIO=<track>/<slug>

# Verify teardown
make reset SCENARIO=<track>/<slug>
```

### 6. PR Checklist

Before submitting your pull request:

- [ ] Scenario README.md is complete and clear
- [ ] All commands in the README work as documented
- [ ] CI checks pass (for Engineering scenarios)
- [ ] Scenario deploys and tears down cleanly
- [ ] Added scenario to `docs/DEMO_GUIDE.md` scenario list
- [ ] Updated appropriate persona in `docs/PERSONAS.md`
- [ ] Tested the full demo flow at least once
- [ ] Added any new make targets to root Makefile
- [ ] Documented any new environment variables in `demo/docs/CONTRACT.md`

## Fixing Bugs

### Reporting Bugs

When reporting a bug, include:

1. **Scenario Name**: Which scenario is affected?
2. **Steps to Reproduce**: Exact commands that trigger the bug
3. **Expected Behavior**: What should happen?
4. **Actual Behavior**: What actually happens?
5. **Environment**: OS, Docker version, Kubernetes version, etc.
6. **Logs**: Relevant error messages or stack traces

### Fixing Bugs

1. Create a branch: `git checkout -b fix/your-bug-description`
2. Make your changes
3. Test thoroughly
4. Submit a PR with a clear description

## Improving Documentation

Documentation improvements are always welcome!

### Types of Documentation

- **Runbooks**: Step-by-step presenter guides
- **Architecture**: Technical system overview
- **Demo Guide**: Scenario selection and usage
- **Personas**: User profiles and use cases
- **Contract**: API and environment contracts

### Documentation Standards

- Use clear, concise language
- Include code examples where appropriate
- Add diagrams for complex concepts (ASCII art is fine)
- Test all commands before documenting them
- Keep formatting consistent with existing docs

## Code Style Guidelines

### Go Code

- Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Write table-driven tests
- Add comments for exported functions

### TypeScript/React Code

- Use TypeScript strict mode
- Follow React best practices
- Use functional components and hooks
- Run `eslint` and `prettier`
- Write meaningful component names

### Shell Scripts

- Use `#!/usr/bin/env bash` shebang
- Enable strict mode: `set -euo pipefail`
- Add comments for complex logic
- Test on both macOS and Linux if possible

### YAML/Kubernetes Manifests

- Use 2-space indentation
- Add labels for resource organization
- Include comments for non-obvious configurations
- Validate with `kubectl --dry-run=client`

## Makefile Conventions

When adding make targets:

- Use `##` comments for help text
- Group related targets
- Use `.PHONY` for non-file targets
- Provide sensible defaults
- Include error handling

Example:
```makefile
.PHONY: sre-demo
sre-demo: ## Run an SRE demo scenario
\t@if [ -z "$(SCENARIO)" ]; then \
\t\techo "Error: SCENARIO is required"; \
\t\texit 1; \
\tfi
\t@echo "Running SRE scenario: $(SCENARIO)"
\t# ... implementation
```

## Testing New Scenarios

### Manual Testing Checklist

For every new scenario:

1. **Fresh Environment**
   - Start from a clean state
   - Run setup commands
   - Verify prerequisites

2. **Happy Path**
   - Follow the README exactly
   - All commands should succeed
   - Output should match documentation

3. **Error Cases**
   - Test with missing prerequisites
   - Test with wrong parameters
   - Verify error messages are helpful

4. **Cleanup**
   - Run teardown commands
   - Verify all resources are removed
   - Check for leftover state

### Automated Testing

Engineering scenarios should include:

- **Unit Tests**: Test individual functions
- **Integration Tests**: Test component interactions
- **E2E Tests** (optional): Test user workflows

Add tests in the appropriate location:
```
fider/app/handlers/your_handler_test.go
fider/public/components/YourComponent.test.tsx
```

## Getting Help

- **Questions**: Open a GitHub issue with the "question" label
- **Bugs**: Open a GitHub issue with the "bug" label
- **Ideas**: Open a GitHub issue with the "enhancement" label
- **Chat**: Join our Discord/Slack (if available)

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (see LICENSE file).
