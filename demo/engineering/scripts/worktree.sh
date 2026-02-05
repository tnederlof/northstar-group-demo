#!/usr/bin/env bash
# worktree.sh - Git worktree lifecycle management for engineering scenarios
#
# Usage: worktree.sh <command> <scenario>
# Commands:
#   init           - Create worktree from broken tag on workshop branch
#   reset-broken   - Reset worktree to broken baseline tag
#   fix-it         - Reset worktree to solved baseline tag
#   remove         - Remove and prune worktree

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENG_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DEMO_DIR="$(cd "$ENG_DIR/.." && pwd)"
REPO_ROOT="$(cd "$DEMO_DIR/.." && pwd)"

usage() {
    cat <<EOF
Usage: $0 <command> <scenario>

Commands:
  init <scenario>          Create worktree from broken tag on workshop branch
  reset-broken <scenario>  Reset worktree to broken baseline (use FORCE=true to force)
  fix-it <scenario>        Reset worktree to solved baseline (use FORCE=true to force)
  remove <scenario>        Remove and prune worktree

Examples:
  $0 init backend/ui-regression
  $0 reset-broken backend/ui-regression
  FORCE=true $0 fix-it backend/ui-regression
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

get_manifest_path() {
    local scenario_dir
    scenario_dir=$(get_scenario_dir "$1")
    echo "$scenario_dir/scenario.json"
}

get_git_field() {
    local scenario="$1"
    local field="$2"
    local manifest
    manifest=$(get_manifest_path "$scenario")
    
    if [[ ! -f "$manifest" ]]; then
        echo "Error: scenario.json not found at $manifest" >&2
        exit 1
    fi
    
    jq -r ".git.$field" "$manifest"
}

fetch_refs() {
    echo "Fetching tags and refs from origin..."
    cd "$REPO_ROOT"
    git fetch --tags origin 2>/dev/null || true
}

cmd_init() {
    local scenario="$1"
    local worktree_dir
    worktree_dir=$(get_worktree_dir "$scenario")
    local broken_ref work_branch
    broken_ref=$(get_git_field "$scenario" "broken_ref")
    work_branch=$(get_git_field "$scenario" "work_branch")
    
    if [[ -d "$worktree_dir" ]]; then
        echo "Worktree already exists at $worktree_dir"
        echo "Use 'reset-broken' or 'fix-it' to reset, or 'remove' to delete"
        return 0
    fi
    
    # Ensure we have the latest tags
    fetch_refs
    
    echo "Creating worktree for scenario: $scenario"
    echo "  Broken ref: $broken_ref"
    echo "  Workshop branch: $work_branch"
    echo "  Path: $worktree_dir"
    
    cd "$REPO_ROOT"
    
    # Clean up any orphaned workshop branch (worktree was removed but branch remains)
    if git show-ref --verify --quiet "refs/heads/$work_branch"; then
        echo "  Cleaning up existing workshop branch: $work_branch"
        git branch -D "$work_branch" 2>/dev/null || true
    fi
    
    # Create worktree on local workshop branch starting from broken tag
    git worktree add -b "$work_branch" "$worktree_dir" "$broken_ref"
    
    echo ""
    echo "Worktree created successfully!"
    echo "You can commit your changes to the workshop branch: $work_branch"
}

reset_worktree_to_ref() {
    local scenario="$1"
    local target_ref="$2"
    local ref_name="$3"
    local worktree_dir
    worktree_dir=$(get_worktree_dir "$scenario")
    
    if [[ ! -d "$worktree_dir" ]]; then
        echo "Worktree does not exist. Use 'init' first."
        exit 1
    fi
    
    # Ensure we have the latest tags
    fetch_refs
    
    echo "Resetting worktree for scenario: $scenario"
    echo "  Target: $ref_name ($target_ref)"
    
    cd "$worktree_dir"
    
    # Check for uncommitted changes
    if [[ -n "$(git status --porcelain)" && "${FORCE:-false}" != "true" ]]; then
        echo "" >&2
        echo "Error: Uncommitted changes present in worktree" >&2
        echo "Commit your work or use FORCE=true to discard changes" >&2
        echo "" >&2
        echo "Tip: Create a backup branch before forcing reset:" >&2
        echo "  git branch ws-backup-\$(date +%s)" >&2
        exit 1
    fi
    
    # Hard reset to target ref
    git reset --hard "$target_ref"
    git clean -fd
    
    echo ""
    echo "Worktree reset to $ref_name successfully"
}

cmd_reset_broken() {
    local scenario="$1"
    local broken_ref
    broken_ref=$(get_git_field "$scenario" "broken_ref")
    reset_worktree_to_ref "$scenario" "$broken_ref" "broken baseline"
}

cmd_fix_it() {
    local scenario="$1"
    local solved_ref
    solved_ref=$(get_git_field "$scenario" "solved_ref")
    
    # Optional: create backup branch
    local worktree_dir
    worktree_dir=$(get_worktree_dir "$scenario")
    if [[ -d "$worktree_dir" ]] && [[ -n "$(cd "$worktree_dir" && git status --porcelain)" ]]; then
        local backup_branch="ws/backup-$(date +%s)"
        echo "Creating backup branch: $backup_branch"
        (cd "$worktree_dir" && git branch "$backup_branch" HEAD)
    fi
    
    reset_worktree_to_ref "$scenario" "$solved_ref" "solved baseline"
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
    reset-broken)
        cmd_reset_broken "$2"
        ;;
    fix-it)
        cmd_fix_it "$2"
        ;;
    remove)
        cmd_remove "$2"
        ;;
    *)
        echo "Unknown command: $1" >&2
        usage
        ;;
esac
