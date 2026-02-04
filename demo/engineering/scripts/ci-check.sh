#!/usr/bin/env bash
# ci-check.sh - Run CI checks (lint/test) in engineering scenario worktrees
#
# Usage: ci-check.sh <scenario>

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENG_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

usage() {
    cat <<EOF
Usage: $0 <scenario>

Run CI checks (lint and test) in a scenario's worktree.

Examples:
  $0 backend/ui-regression
  $0 frontend/missing-fallback
EOF
    exit 1
}

get_scenario_dir() {
    echo "$ENG_DIR/scenarios/$1"
}

get_worktree_dir() {
    echo "$(get_scenario_dir "$1")/worktree"
}

cmd_ci_check() {
    local scenario="$1"
    local worktree_dir
    worktree_dir=$(get_worktree_dir "$scenario")
    
    if [[ ! -d "$worktree_dir" ]]; then
        echo "Error: Worktree does not exist at $worktree_dir" >&2
        echo "Run 'make run SCENARIO=$scenario' first to create the worktree" >&2
        exit 1
    fi
    
    local fider_dir="$worktree_dir/fider"
    
    if [[ ! -d "$fider_dir" ]]; then
        echo "Error: fider directory not found at $fider_dir" >&2
        exit 1
    fi
    
    echo "Running CI checks for scenario: $scenario"
    echo "  Worktree: $worktree_dir"
    echo ""
    
    # Check if Makefile exists
    if [[ -f "$fider_dir/Makefile" ]]; then
        echo "Running tests in fider directory..."
        cd "$fider_dir"
        
        # Run tests if available
        if make -n test &>/dev/null; then
            make test || {
                echo "Tests failed" >&2
                exit 1
            }
        else
            echo "No test target found in Makefile, skipping tests"
        fi
        
        # Run lint if available
        if make -n lint &>/dev/null; then
            echo ""
            echo "Running lint..."
            make lint || {
                echo "Lint failed" >&2
                exit 1
            }
        else
            echo "No lint target found in Makefile, skipping lint"
        fi
    else
        echo "No Makefile found in fider directory, skipping CI checks"
    fi
    
    echo ""
    echo "CI checks completed successfully!"
}

# Main
if [[ $# -lt 1 ]]; then
    usage
fi

cmd_ci_check "$1"
