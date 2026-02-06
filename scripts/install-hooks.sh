#!/bin/bash
# Install git hooks for northstar-group-demo
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOOKS_DIR="$REPO_ROOT/.git/hooks"

echo "Installing git hooks..."

# Pre-commit hook to prevent infrastructure changes in scenario branches
cat > "$HOOKS_DIR/pre-commit" <<'EOF'
#!/bin/bash
# Pre-commit hook to prevent infrastructure changes in scenario branches

CURRENT_BRANCH=$(git symbolic-ref --short HEAD 2>/dev/null)

# Only check scenario branches (scenario/* or ws/*)
if [[ ! "$CURRENT_BRANCH" =~ ^(scenario/|ws/) ]]; then
    exit 0
fi

# List of infrastructure paths that should only be modified in main
INFRASTRUCTURE_PATHS=(
    "demo/engineering/compose/"
    "demo/sre/base/"
    "demo/shared/"
    "democtl/"
    "Makefile"
    "docs/"
)

# Get staged files
STAGED_FILES=$(git diff --cached --name-only)

# Check if any staged files are in infrastructure paths
INFRA_CHANGES=""
for path in "${INFRASTRUCTURE_PATHS[@]}"; do
    while IFS= read -r file; do
        if [[ "$file" =~ ^$path ]]; then
            INFRA_CHANGES="${INFRA_CHANGES}\n  - $file"
        fi
    done <<< "$STAGED_FILES"
done

if [[ -n "$INFRA_CHANGES" ]]; then
    echo "❌ ERROR: Infrastructure changes detected in scenario branch '$CURRENT_BRANCH'"
    echo ""
    echo "The following infrastructure files should only be modified in 'main':"
    echo -e "$INFRA_CHANGES"
    echo ""
    echo "To fix this:"
    echo "  1. Stash these changes: git stash"
    echo "  2. Switch to main: git checkout main"
    echo "  3. Apply and commit: git stash pop && git commit"
    echo "  4. Rebase scenario branches: git checkout $CURRENT_BRANCH && git rebase main"
    echo ""
    echo "To bypass this check (not recommended): git commit --no-verify"
    exit 1
fi

exit 0
EOF

chmod +x "$HOOKS_DIR/pre-commit"

echo "✓ Git hooks installed successfully!"
echo ""
echo "The pre-commit hook will prevent infrastructure changes in scenario branches."
echo "Infrastructure changes should be made in 'main' and then rebased into scenario branches."
