# Contributing to Northstar Group Demo

Thank you for your interest in improving the demo! This guide will help you add new scenarios, fix bugs, and enhance the documentation.

## How to Add New Scenarios

### 1. Choose a Track

Decide whether your scenario fits the **SRE track** or **Engineering track**:

- **SRE Track**: Infrastructure, Kubernetes, deployments, scaling
- **Engineering Track**: Application code, bugs, testing, CI/CD

### 2. Create Scenario Branch

For **SRE scenarios**:
```bash
git checkout -b demo/sre/your-scenario-name main
```

For **Engineering scenarios**:
```bash
git checkout -b demo/engineering/your-scenario-name main
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

Introduce code changes in the `fider/` directory:

```bash
# Make your changes to introduce a bug or demonstrate a feature
vim fider/app/handlers/your_handler.go

# Add or modify tests
vim fider/app/handlers/your_handler_test.go

# Create scenario README
vim demo/engineering/scenarios/your-scenario-name/README.md
```

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
`bash
make sre-demo SCENARIO=your-scenario-name
# or
make eng-up SCENARIO=your-scenario-name
`

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
make sre-demo SCENARIO=your-scenario-name

# Verify health checks pass
make sre-health SCENARIO=your-scenario-name

# Verify teardown works
make sre-reset SCENARIO=your-scenario-name
```

#### Engineering Scenarios

```bash
# Verify CI checks run
make eng-ci SCENARIO=your-scenario-name

# Verify the app starts
make eng-up SCENARIO=your-scenario-name

# If applicable, run E2E tests
E2E=true make eng-ci SCENARIO=your-scenario-name

# Verify teardown
make eng-down SCENARIO=your-scenario-name
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
