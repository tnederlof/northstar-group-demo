#!/usr/bin/env bash
# apply-seed.sh - Unified seed script with k8s and compose adapters
#
# Usage: apply-seed.sh <runtime> <scenario>
#   runtime: 'k8s' or 'compose'
#   scenario: path like 'platform/bad-rollout'
#
# K8s Adapter (apply-seed.sh k8s <scenario>):
#   - Derives SLUG and NAMESPACE from scenario
#   - Uses kubectl exec to run psql with seed.sql against postgres deployment
#   - Supports KUBE_CONTEXT env var (default: kind-fider-demo)
#
# Compose Adapter (apply-seed.sh compose <scenario>):
#   - Derives SLUG from scenario
#   - Uses docker compose exec to run psql with seed.sql
#   - Uses project naming: northstar-<slug>

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEMO_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
SHARED_DIR="$DEMO_DIR/shared"
SEED_FILE="$SHARED_DIR/northstar/seed.sql"

usage() {
    cat <<EOF
Usage: $0 <runtime> <scenario>

Arguments:
  runtime   The runtime environment: 'k8s' or 'compose'
  scenario  The scenario path (e.g., 'platform/bad-rollout')

Examples:
  $0 k8s platform/bad-rollout
  $0 compose backend/ui-regression

Environment Variables:
  KUBE_CONTEXT  Kubernetes context for k8s runtime (default: kind-fider-demo)
EOF
    exit 1
}

if [[ $# -ne 2 ]]; then
    usage
fi

RUNTIME="$1"
SCENARIO="$2"

# Validate runtime
if [[ "$RUNTIME" != "k8s" && "$RUNTIME" != "compose" ]]; then
    echo "Error: Invalid runtime '$RUNTIME'. Must be 'k8s' or 'compose'." >&2
    usage
fi

# Check seed file exists
if [[ ! -f "$SEED_FILE" ]]; then
    echo "Error: Seed file not found: $SEED_FILE" >&2
    exit 1
fi

# Get the slug using scenario-slug.sh
SLUG="$("$SCRIPT_DIR/scenario-slug.sh" "$SCENARIO")"

echo "Applying seed data..."
echo "  Runtime: $RUNTIME"
echo "  Scenario: $SCENARIO"
echo "  Slug: $SLUG"
echo ""

apply_k8s() {
    local context="${KUBE_CONTEXT:-kind-fider-demo}"
    local namespace="demo-$SLUG"
    
    echo "Using Kubernetes context: $context"
    echo "Target namespace: $namespace"
    echo ""
    
    # Find the postgres pod
    local postgres_pod
    postgres_pod=$(kubectl --context="$context" -n "$namespace" get pods \
        -l app=postgres -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    
    if [[ -z "$postgres_pod" ]]; then
        echo "Error: Could not find postgres pod in namespace $namespace" >&2
        exit 1
    fi
    
    echo "Found postgres pod: $postgres_pod"
    echo "Executing seed.sql..."
    
    kubectl --context="$context" -n "$namespace" exec -i "$postgres_pod" -- \
        psql -U fider -d fider < "$SEED_FILE"
    
    echo ""
    echo "Seed data applied successfully!"
}

apply_compose() {
    local project="northstar-$SLUG"
    
    echo "Using Compose project: $project"
    echo ""
    
    echo "Executing seed.sql..."
    
    docker compose -p "$project" exec -T postgres \
        psql -U fider -d fider < "$SEED_FILE"
    
    echo ""
    echo "Seed data applied successfully!"
}

case "$RUNTIME" in
    k8s)
        apply_k8s
        ;;
    compose)
        apply_compose
        ;;
esac
