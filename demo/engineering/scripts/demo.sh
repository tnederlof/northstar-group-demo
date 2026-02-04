#!/usr/bin/env bash
# demo.sh - Engineering demo orchestration script
#
# Usage: demo.sh <command> <scenario>
# Commands:
#   up       - Start scenario (render env, ensure worktree, start compose)
#   down     - Stop scenario (compose down)
#   reset    - down + up (full reset)
#   sniff    - Follow compose logs
#   status   - Show compose status

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENG_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DEMO_DIR="$(cd "$ENG_DIR/.." && pwd)"
SHARED_DIR="$DEMO_DIR/shared"

usage() {
    cat <<EOF
Usage: $0 <command> [scenario]

Commands:
  up <scenario>       Start a scenario (e.g., backend/ui-regression)
  down <scenario>     Stop a scenario
  reset <scenario>    Full reset (down + up)
  sniff <scenario>    Follow compose logs
  status <scenario>   Show compose status

Examples:
  $0 up backend/ui-regression
  $0 down backend/ui-regression
  $0 sniff backend/ui-regression
EOF
    exit 1
}

get_slug() {
    "$SHARED_DIR/scripts/scenario-slug.sh" "$1"
}

get_scenario_dir() {
    echo "$ENG_DIR/scenarios/$1"
}

get_compose_file() {
    echo "$(get_scenario_dir "$1")/docker-compose.yml"
}

# Ensure edge proxy network exists
ensure_network() {
    if ! docker network inspect northstar-demo >/dev/null 2>&1; then
        echo "Creating northstar-demo network..."
        docker network create northstar-demo
    fi
}

# Render environment file for scenario
render_env() {
    local scenario="$1"
    echo "Rendering environment for $scenario..."
    "$SHARED_DIR/scripts/render-env.sh" engineering "$scenario"
}

# Ensure worktree exists
ensure_worktree() {
    local scenario="$1"
    local scenario_dir
    scenario_dir=$(get_scenario_dir "$scenario")
    local worktree_dir="$scenario_dir/worktree"
    
    if [[ ! -d "$worktree_dir" ]]; then
        echo "Initializing worktree for $scenario..."
        "$SCRIPT_DIR/worktree.sh" init "$scenario"
    fi
}

cmd_up() {
    local scenario="$1"
    local scenario_dir
    scenario_dir=$(get_scenario_dir "$scenario")
    local compose_file
    compose_file=$(get_compose_file "$scenario")
    local slug
    slug=$(get_slug "$scenario")
    
    if [[ ! -f "$compose_file" ]]; then
        echo "Error: docker-compose.yml not found at $compose_file" >&2
        exit 1
    fi
    
    echo "Starting scenario: $scenario"
    echo "  Slug: $slug"
    echo ""
    
    # Ensure network exists
    ensure_network
    
    # Render environment
    local env_file
    env_file=$(render_env "$scenario")
    echo "  Environment: $env_file"
    
    # Ensure worktree exists
    ensure_worktree "$scenario"
    
    # Source the env file to get secrets for compose interpolation
    # shellcheck source=/dev/null
    source "$env_file"
    
    # Start compose
    echo ""
    echo "Starting containers..."
    (
        cd "$scenario_dir"
        export JWT_SECRET
        export DEMO_LOGIN_KEY
        docker compose --env-file "$env_file" up -d --build
    )
    
    echo ""
    echo "Scenario started!"
    echo "Access at: http://$slug.localhost:8080"
}

cmd_down() {
    local scenario="$1"
    local scenario_dir
    scenario_dir=$(get_scenario_dir "$scenario")
    local compose_file
    compose_file=$(get_compose_file "$scenario")
    
    if [[ ! -f "$compose_file" ]]; then
        echo "Warning: docker-compose.yml not found at $compose_file" >&2
        return
    fi
    
    echo "Stopping scenario: $scenario"
    
    (
        cd "$scenario_dir"
        docker compose down -v --remove-orphans
    )
    
    echo "Scenario stopped."
}

cmd_reset() {
    local scenario="$1"
    cmd_down "$scenario"
    
    # Also reset the worktree
    echo "Resetting worktree..."
    "$SCRIPT_DIR/worktree.sh" reset "$scenario"
    
    sleep 2
    cmd_up "$scenario"
}

cmd_sniff() {
    local scenario="$1"
    local scenario_dir
    scenario_dir=$(get_scenario_dir "$scenario")
    local compose_file
    compose_file=$(get_compose_file "$scenario")
    
    if [[ ! -f "$compose_file" ]]; then
        echo "Error: docker-compose.yml not found at $compose_file" >&2
        exit 1
    fi
    
    echo "Following logs for: $scenario"
    echo "(Press Ctrl+C to stop)"
    echo ""
    
    (
        cd "$scenario_dir"
        docker compose logs -f
    )
}

cmd_status() {
    local scenario="$1"
    local scenario_dir
    scenario_dir=$(get_scenario_dir "$scenario")
    local compose_file
    compose_file=$(get_compose_file "$scenario")
    
    if [[ ! -f "$compose_file" ]]; then
        echo "Error: docker-compose.yml not found at $compose_file" >&2
        exit 1
    fi
    
    echo "Status for: $scenario"
    echo ""
    
    (
        cd "$scenario_dir"
        docker compose ps
    )
}

# Main
if [[ $# -lt 1 ]]; then
    usage
fi

case "$1" in
    up)
        [[ $# -lt 2 ]] && usage
        cmd_up "$2"
        ;;
    down)
        [[ $# -lt 2 ]] && usage
        cmd_down "$2"
        ;;
    reset)
        [[ $# -lt 2 ]] && usage
        cmd_reset "$2"
        ;;
    sniff)
        [[ $# -lt 2 ]] && usage
        cmd_sniff "$2"
        ;;
    status)
        [[ $# -lt 2 ]] && usage
        cmd_status "$2"
        ;;
    *)
        echo "Unknown command: $1" >&2
        usage
        ;;
esac
