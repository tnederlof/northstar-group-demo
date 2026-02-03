#!/usr/bin/env bash
# demo.sh - Main SRE demo orchestration script
#
# Usage: demo.sh <command> <scenario>
# Commands:
#   up       - Deploy scenario
#   down     - Remove scenario namespace
#   reset    - down + up (full reset)
#   down-all - Teardown all demo namespaces

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DEMO_DIR="$(cd "$SRE_DIR/.." && pwd)"
SHARED_DIR="$DEMO_DIR/shared"

KUBE_CONTEXT="${KUBE_CONTEXT:-kind-fider-demo}"

usage() {
    cat <<EOF
Usage: $0 <command> [scenario]

Commands:
  up <scenario>       Deploy a scenario (e.g., platform/bad-rollout)
  down <scenario>     Remove a scenario namespace
  reset <scenario>    Full reset (down + up)
  down-all            Teardown all demo namespaces

Examples:
  $0 up platform/bad-rollout
  $0 down platform/bad-rollout
  $0 reset platform/bad-rollout
  $0 down-all
EOF
    exit 1
}

get_slug() {
    "$SHARED_DIR/scripts/scenario-slug.sh" "$1"
}

get_namespace() {
    local slug
    slug=$(get_slug "$1")
    echo "demo-$slug"
}

get_scenario_dir() {
    echo "$SRE_DIR/scenarios/$1"
}

cmd_up() {
    local scenario="$1"
    local scenario_dir
    scenario_dir=$(get_scenario_dir "$scenario")
    local namespace
    namespace=$(get_namespace "$scenario")
    local slug
    slug=$(get_slug "$scenario")
    
    if [[ ! -d "$scenario_dir" ]]; then
        echo "Error: Scenario not found: $scenario" >&2
        exit 1
    fi
    
    echo "Deploying scenario: $scenario"
    echo "  Namespace: $namespace"
    echo "  Slug: $slug"
    echo ""
    
    # Create namespace
    kubectl --context="$KUBE_CONTEXT" create namespace "$namespace" --dry-run=client -o yaml | \
        kubectl --context="$KUBE_CONTEXT" apply -f -
    
    # Apply kustomize overlay
    kubectl --context="$KUBE_CONTEXT" apply -k "$scenario_dir"
    
    # Wait for postgres to be ready before seeding
    echo "Waiting for postgres to be ready..."
    kubectl --context="$KUBE_CONTEXT" -n "$namespace" wait --for=condition=available \
        deployment/postgres --timeout=120s
    
    # Apply seed data
    echo "Applying seed data..."
    "$SHARED_DIR/scripts/apply-seed.sh" k8s "$scenario"
    
    echo ""
    echo "Scenario deployed!"
    echo "Access at: http://$slug.localhost:8080"
}

cmd_down() {
    local scenario="$1"
    local namespace
    namespace=$(get_namespace "$scenario")
    
    echo "Removing scenario: $scenario"
    echo "  Namespace: $namespace"
    
    kubectl --context="$KUBE_CONTEXT" delete namespace "$namespace" --ignore-not-found
    
    echo "Scenario removed."
}

cmd_reset() {
    local scenario="$1"
    cmd_down "$scenario"
    sleep 2
    cmd_up "$scenario"
}

cmd_down_all() {
    echo "Removing all demo namespaces..."
    
    local namespaces
    namespaces=$(kubectl --context="$KUBE_CONTEXT" get namespaces -o name | \
        grep "^namespace/demo-" | cut -d'/' -f2 || true)
    
    if [[ -z "$namespaces" ]]; then
        echo "No demo namespaces found."
        return
    fi
    
    for ns in $namespaces; do
        echo "Deleting namespace: $ns"
        kubectl --context="$KUBE_CONTEXT" delete namespace "$ns" --ignore-not-found
    done
    
    echo "All demo namespaces removed."
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
    down-all)
        cmd_down_all
        ;;
    *)
        echo "Unknown command: $1" >&2
        usage
        ;;
esac
