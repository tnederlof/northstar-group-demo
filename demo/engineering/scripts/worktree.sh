#!/usr/bin/env bash
# worktree.sh - Git worktree lifecycle management for engineering scenarios
#
# Usage: worktree.sh <command> <scenario>
# Commands:
#   init    - Create worktree from scenario branch
#   reset   - Reset worktree to branch HEAD
#   remove  - Remove and prune worktree

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENG_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DEMO_DIR="$(cd "$ENG_DIR/.." && pwd)"
REPO_ROOT="$(cd "$DEMO_DIR/.." && pwd)"

usage() {
    cat <<EOF
Usage: $0 <command> <scenario>

Commands:
  init <scenario>    Create worktree from scenario branch
  reset <scenario>   Reset worktree to branch HEAD (use FORCE=true to force)
  remove <scenario>  Remove and prune worktree

Examples:
  $0 init backend/ui-regression
  $0 reset backend/ui-regression
  FORCE=true $0 reset backend/ui-regression
  $0 remove backend/ui-regression
EOF
    exit 1
}

get_scenario_dir() {
    echo "$ENG_DIR/scenarios/$1"
}

get_worktree_dir() {
    echo "$(get_scenario_dir "$1")/worktree"
}

get_branch() {
    local scenario_dir
    scenario_dir=$(get_scenario_dir "$1")
    local manifest="$scenario_dir/scenario.json"
    
    if [[ ! -f "$manifest" ]]; then
        echo "Error: scenario.json not found at $manifest" >&2
        exit 1
    fi
    
    jq -r '.branch' "$manifest"
}

cmd_init() {
    local scenario="$1"
    local worktree_dir
    worktree_dir=$(get_worktree_dir "$scenario")
    local branch
    branch=$(get_branch "$scenario")
    
    if [[ -d "$worktree_dir" ]]; then
        echo "Worktree already exists at $worktree_dir"
        echo "Use 'reset' to reset or 'remove' to delete"
        exit 1
    fi
    
    echo "Creating worktree for scenario: $scenario"
    echo "  Branch: $branch"
    echo "  Path: $worktree_dir"
    
    cd "$REPO_ROOT"
    git worktree add "$worktree_dir" "$branch"
    
    echo ""
    echo "Worktree created successfully!"
}

cmd_reset() {
    local scenario="$1"
    local worktree_dir
    worktree_dir=$(get_worktree_dir "$scenario")
    local branch
    branch=$(get_branch "$scenario")
    
    if [[ ! -d "$worktree_dir" ]]; then
        echo "Worktree does not exist. Use 'init' first."
        exit 1
    fi
    
    echo "Resetting worktree for scenario: $scenario"
    echo "  Branch: $branch"
    
    cd "$worktree_dir"
    
    if [[ "${FORCE:-false}" == "true" ]]; then
        git reset --hard "origin/$branch"
        git clean -fd
    else
        git fetch origin "$branch"
        git reset --hard "origin/$branch"
    fi
    
    echo "Worktree reset to $branch HEAD"
}

cmd_remove() {
    local scenario="$1"
    local worktree_dir
    worktree_dir=$(get_worktree_dir "$scenario")
    
    if [[ ! -d "$worktree_dir" ]]; then
        echo "Worktree does not exist at $worktree_dir"
        exit 0
    fi
    
    echo "Removing worktree for scenario: $scenario"
    
    cd "$REPO_ROOT"
    git worktree remove "$worktree_dir" --force || rm -rf "$worktree_dir"
    git worktree prune
    
    echo "Worktree removed."
}

# Main
if [[ $# -lt 2 ]]; then
    usage
fi

case "$1" in
    init)
        cmd_init "$2"
        ;;
    reset)
        cmd_reset "$2"
        ;;
    remove)
        cmd_remove "$2"
        ;;
    *)
        echo "Unknown command: $1" >&2
        usage
        ;;
esac
