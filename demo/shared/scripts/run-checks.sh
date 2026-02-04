#!/usr/bin/env bash
# run-checks.sh - Shared scenario verification runner
#
# Usage: run-checks.sh <type> <scenario> <command> [options]
#
# Arguments:
#   type      - sre or engineering
#   scenario  - scenario path (e.g., platform/bad-rollout)
#   command   - verify or health
#
# Options:
#   --stage <stage>   - Stage to run checks for (default: from manifest)
#   --only <filter>   - Only run checks of this type (e.g., playwright, http, k8s)
#   --json            - Output results as JSON
#   --verbose         - Show detailed output
#
# Examples:
#   ./run-checks.sh sre platform/bad-rollout verify
#   ./run-checks.sh sre platform/healthy verify --stage healthy
#   ./run-checks.sh sre platform/healthy verify --only playwright

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEMO_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Source the scenario library
# shellcheck source=lib/scenario.sh
source "$SCRIPT_DIR/lib/scenario.sh"

# Configuration
KUBE_CONTEXT="${KUBE_CONTEXT:-kind-fider-demo}"
VERBOSE="${VERBOSE:-false}"
JSON_OUTPUT="${JSON_OUTPUT:-false}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
CHECKS_PASSED=0
CHECKS_FAILED=0
CHECKS_SKIPPED=0

usage() {
    cat <<EOF
Usage: $0 <type> <scenario> <command> [options]

Arguments:
  type      - sre or engineering
  scenario  - scenario path (e.g., platform/bad-rollout)
  command   - verify or health

Options:
  --stage <stage>   - Stage to run checks for (default: from manifest)
  --only <filter>   - Only run checks of this type (e.g., playwright, http, k8s)
  --json            - Output results as JSON
  --verbose         - Show detailed output

Examples:
  $0 sre platform/bad-rollout verify
  $0 sre platform/healthy verify --stage healthy
  $0 sre platform/healthy verify --only playwright
EOF
    exit 1
}

log_info() {
    if [[ "$JSON_OUTPUT" != "true" ]]; then
        echo -e "${BLUE}[INFO]${NC} $*"
    fi
}

log_pass() {
    if [[ "$JSON_OUTPUT" != "true" ]]; then
        echo -e "${GREEN}[PASS]${NC} $*"
    fi
}

log_fail() {
    if [[ "$JSON_OUTPUT" != "true" ]]; then
        echo -e "${RED}[FAIL]${NC} $*" >&2
    fi
}

log_skip() {
    if [[ "$JSON_OUTPUT" != "true" ]]; then
        echo -e "${YELLOW}[SKIP]${NC} $*"
    fi
}

log_verbose() {
    if [[ "$VERBOSE" == "true" && "$JSON_OUTPUT" != "true" ]]; then
        echo -e "       $*"
    fi
}

# ============================================================================
# Check Primitives
# ============================================================================

# http.get - Make HTTP GET request and check status
# Args: check_json
run_http_get() {
    local check_json="$1"
    local url description
    url=$(echo "$check_json" | jq -r '.url')
    description=$(echo "$check_json" | jq -r '.description // "HTTP GET check"')
    
    local expect_status expect_status_not
    expect_status=$(echo "$check_json" | jq -c '.expect.status // []')
    expect_status_not=$(echo "$check_json" | jq -c '.expect.status_not // []')
    
    log_verbose "GET $url"
    
    local status_code
    status_code=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 5 --max-time 10 "$url" 2>/dev/null || echo "000")
    
    log_verbose "Response status: $status_code"
    
    # Check status_not first (negative assertion)
    if [[ "$expect_status_not" != "[]" ]]; then
        if echo "$expect_status_not" | jq -e "contains([$status_code])" >/dev/null 2>&1; then
            log_fail "$description (got $status_code, expected not in $expect_status_not)"
            return 1
        fi
    fi
    
    # Check status (positive assertion)
    if [[ "$expect_status" != "[]" ]]; then
        if ! echo "$expect_status" | jq -e "contains([$status_code])" >/dev/null 2>&1; then
            log_fail "$description (got $status_code, expected one of $expect_status)"
            return 1
        fi
    fi
    
    log_pass "$description"
    return 0
}

# k8s.jqEquals - Check K8s resource field with jq
# Args: check_json namespace
run_k8s_jq_equals() {
    local check_json="$1"
    local namespace="$2"
    local description kind name jq_expr expected
    description=$(echo "$check_json" | jq -r '.description // "K8s jq check"')
    kind=$(echo "$check_json" | jq -r '.resource.kind')
    name=$(echo "$check_json" | jq -r '.resource.name')
    jq_expr=$(echo "$check_json" | jq -r '.jq')
    expected=$(echo "$check_json" | jq -r '.equals')
    
    log_verbose "Checking $kind/$name with jq: $jq_expr"
    
    local actual
    actual=$(kubectl --context="$KUBE_CONTEXT" -n "$namespace" get "$kind" "$name" -o json 2>/dev/null | jq -r "$jq_expr" 2>/dev/null || echo "")
    
    if [[ "$actual" != "$expected" ]]; then
        log_fail "$description (expected '$expected', got '$actual')"
        return 1
    fi
    
    log_pass "$description"
    return 0
}

# k8s.podsContainLog - Check pod logs for substring
# Args: check_json namespace
run_k8s_pods_contain_log() {
    local check_json="$1"
    local namespace="$2"
    local description selector contains since_seconds
    description=$(echo "$check_json" | jq -r '.description // "K8s log check"')
    selector=$(echo "$check_json" | jq -r '.selector')
    contains=$(echo "$check_json" | jq -r '.contains')
    since_seconds=$(echo "$check_json" | jq -r '.since_seconds // 300')
    
    log_verbose "Checking logs for pods with selector: $selector"
    
    local logs
    logs=$(kubectl --context="$KUBE_CONTEXT" -n "$namespace" logs -l "$selector" --since="${since_seconds}s" --tail=500 2>/dev/null || echo "")
    
    if [[ -z "$logs" ]]; then
        log_fail "$description (no logs found)"
        return 1
    fi
    
    if ! echo "$logs" | grep -q "$contains"; then
        log_fail "$description (substring '$contains' not found in logs)"
        return 1
    fi
    
    log_pass "$description"
    return 0
}

# k8s.podTerminationReason - Check pod termination reason
# Args: check_json namespace
run_k8s_pod_termination_reason() {
    local check_json="$1"
    local namespace="$2"
    local description selector reason
    description=$(echo "$check_json" | jq -r '.description // "K8s termination reason check"')
    selector=$(echo "$check_json" | jq -r '.selector')
    reason=$(echo "$check_json" | jq -r '.reason')
    
    log_verbose "Checking termination reason for pods with selector: $selector"
    
    local pod_json
    pod_json=$(kubectl --context="$KUBE_CONTEXT" -n "$namespace" get pods -l "$selector" -o json 2>/dev/null || echo '{"items":[]}')
    
    local found_reason
    found_reason=$(echo "$pod_json" | jq -r '.items[].status.containerStatuses[]?.lastState.terminated.reason // empty' 2>/dev/null | head -1)
    
    if [[ "$found_reason" != "$reason" ]]; then
        log_fail "$description (expected '$reason', got '${found_reason:-none}')"
        return 1
    fi
    
    log_pass "$description"
    return 0
}

# k8s.podRestartCount - Check pod restart count
# Args: check_json namespace
run_k8s_pod_restart_count() {
    local check_json="$1"
    local namespace="$2"
    local description selector min_restarts
    description=$(echo "$check_json" | jq -r '.description // "K8s restart count check"')
    selector=$(echo "$check_json" | jq -r '.selector')
    min_restarts=$(echo "$check_json" | jq -r '.min_restarts // 1')
    
    log_verbose "Checking restart count for pods with selector: $selector"
    
    local restarts
    restarts=$(kubectl --context="$KUBE_CONTEXT" -n "$namespace" get pods -l "$selector" -o jsonpath='{.items[*].status.containerStatuses[*].restartCount}' 2>/dev/null || echo "0")
    
    local max_restarts=0
    for count in $restarts; do
        if [[ "$count" -gt "$max_restarts" ]]; then
            max_restarts="$count"
        fi
    done
    
    if [[ "$max_restarts" -lt "$min_restarts" ]]; then
        log_fail "$description (restart count $max_restarts < $min_restarts)"
        return 1
    fi
    
    log_pass "$description"
    return 0
}

# k8s.deploymentAvailable - Check deployment is available
# Args: check_json namespace
run_k8s_deployment_available() {
    local check_json="$1"
    local namespace="$2"
    local description name
    description=$(echo "$check_json" | jq -r '.description // "K8s deployment check"')
    name=$(echo "$check_json" | jq -r '.name')
    
    log_verbose "Checking deployment: $name"
    
    local available
    available=$(kubectl --context="$KUBE_CONTEXT" -n "$namespace" get deployment "$name" -o jsonpath='{.status.conditions[?(@.type=="Available")].status}' 2>/dev/null || echo "")
    
    if [[ "$available" != "True" ]]; then
        log_fail "$description (deployment not available)"
        return 1
    fi
    
    log_pass "$description"
    return 0
}

# k8s.resourceExists - Check if K8s resource exists
# Args: check_json namespace
run_k8s_resource_exists() {
    local check_json="$1"
    local namespace="$2"
    local description kind name
    description=$(echo "$check_json" | jq -r '.description // "K8s resource exists check"')
    kind=$(echo "$check_json" | jq -r '.resource.kind')
    name=$(echo "$check_json" | jq -r '.resource.name')
    
    log_verbose "Checking if $kind/$name exists"
    
    if ! kubectl --context="$KUBE_CONTEXT" -n "$namespace" get "$kind" "$name" >/dev/null 2>&1; then
        log_fail "$description ($kind/$name not found)"
        return 1
    fi
    
    log_pass "$description"
    return 0
}

# k8s.serviceMissingPort - Check if service is missing a port
# Args: check_json namespace
run_k8s_service_missing_port() {
    local check_json="$1"
    local namespace="$2"
    local description name port_name
    description=$(echo "$check_json" | jq -r '.description // "K8s service port check"')
    name=$(echo "$check_json" | jq -r '.name')
    port_name=$(echo "$check_json" | jq -r '.port_name')
    
    log_verbose "Checking if service $name is missing port $port_name"
    
    local ports
    ports=$(kubectl --context="$KUBE_CONTEXT" -n "$namespace" get service "$name" -o jsonpath='{.spec.ports[*].name}' 2>/dev/null || echo "")
    
    if echo "$ports" | grep -qw "$port_name"; then
        log_fail "$description (port $port_name exists, expected missing)"
        return 1
    fi
    
    log_pass "$description"
    return 0
}

# playwright.run - Run Playwright test suite
# Args: check_json
run_playwright() {
    local check_json="$1"
    local description suite headed
    description=$(echo "$check_json" | jq -r '.description // "Playwright test"')
    suite=$(echo "$check_json" | jq -r '.suite')
    headed=$(echo "$check_json" | jq -r '.headed // false')
    
    local ui_dir="$DEMO_DIR/ui"
    
    if [[ ! -d "$ui_dir" ]]; then
        log_skip "$description (UI test suite not found at $ui_dir)"
        CHECKS_SKIPPED=$((CHECKS_SKIPPED + 1))
        return 0
    fi
    
    log_verbose "Running Playwright suite: $suite"
    
    local base_url
    base_url=$(scenario_get_base_url)
    
    local demo_login_key
    demo_login_key=$(scenario_get_demo_login_key "$KUBE_CONTEXT")
    
    local headed_flag=""
    if [[ "$headed" == "true" || "${PLAYWRIGHT_HEADED:-}" == "true" ]]; then
        headed_flag="--headed"
    fi
    
    # Run playwright from the ui directory
    if (
        cd "$ui_dir"
        BASE_URL="$base_url" \
        SCENARIO="$SCENARIO_TRACK/$SCENARIO_SLUG" \
        STAGE="$STAGE" \
        DEMO_LOGIN_KEY="${demo_login_key:-northstar-demo-key}" \
        npx playwright test --grep "$suite" $headed_flag 2>&1
    ); then
        log_pass "$description"
        return 0
    else
        log_fail "$description (Playwright test failed)"
        return 1
    fi
}

# ============================================================================
# Check Router
# ============================================================================

# Run a single check based on its type
# Args: check_json namespace
run_check() {
    local check_json="$1"
    local namespace="$2"
    
    local check_type
    check_type=$(echo "$check_json" | jq -r '.type')
    
    case "$check_type" in
        http.get)
            run_http_get "$check_json"
            ;;
        k8s.jqEquals)
            run_k8s_jq_equals "$check_json" "$namespace"
            ;;
        k8s.podsContainLog)
            run_k8s_pods_contain_log "$check_json" "$namespace"
            ;;
        k8s.podTerminationReason)
            run_k8s_pod_termination_reason "$check_json" "$namespace"
            ;;
        k8s.podRestartCount)
            run_k8s_pod_restart_count "$check_json" "$namespace"
            ;;
        k8s.deploymentAvailable)
            run_k8s_deployment_available "$check_json" "$namespace"
            ;;
        k8s.resourceExists)
            run_k8s_resource_exists "$check_json" "$namespace"
            ;;
        k8s.serviceMissingPort)
            run_k8s_service_missing_port "$check_json" "$namespace"
            ;;
        playwright.run)
            run_playwright "$check_json"
            ;;
        *)
            log_skip "Unknown check type: $check_type"
            CHECKS_SKIPPED=$((CHECKS_SKIPPED + 1))
            return 0
            ;;
    esac
}

# ============================================================================
# Main
# ============================================================================

# Parse arguments
TYPE=""
SCENARIO=""
COMMAND=""
STAGE=""
ONLY_FILTER=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --stage)
            STAGE="$2"
            shift 2
            ;;
        --only)
            ONLY_FILTER="$2"
            shift 2
            ;;
        --json)
            JSON_OUTPUT="true"
            shift
            ;;
        --verbose|-v)
            VERBOSE="true"
            shift
            ;;
        --help|-h)
            usage
            ;;
        -*)
            echo "Unknown option: $1" >&2
            usage
            ;;
        *)
            if [[ -z "$TYPE" ]]; then
                TYPE="$1"
            elif [[ -z "$SCENARIO" ]]; then
                SCENARIO="$1"
            elif [[ -z "$COMMAND" ]]; then
                COMMAND="$1"
            else
                echo "Unexpected argument: $1" >&2
                usage
            fi
            shift
            ;;
    esac
done

# Validate required arguments
if [[ -z "$TYPE" || -z "$SCENARIO" || -z "$COMMAND" ]]; then
    echo "Error: Missing required arguments" >&2
    usage
fi

if [[ "$TYPE" != "sre" && "$TYPE" != "engineering" ]]; then
    echo "Error: type must be 'sre' or 'engineering'" >&2
    exit 1
fi

if [[ "$COMMAND" != "verify" && "$COMMAND" != "health" ]]; then
    echo "Error: command must be 'verify' or 'health'" >&2
    exit 1
fi

# Load scenario
if ! scenario_load "$TYPE" "$SCENARIO"; then
    exit 1
fi

# Check if scenario has checks defined
if ! scenario_has_checks; then
    log_info "No checks defined for scenario $SCENARIO"
    exit 0
fi

# Determine stage
if [[ -z "$STAGE" ]]; then
    STAGE=$(scenario_get_default_stage)
fi

if [[ -z "$STAGE" ]]; then
    echo "Error: Could not determine stage. Specify with --stage" >&2
    exit 1
fi

# Get namespace for K8s checks
NAMESPACE=$(scenario_get_namespace)

# Get checks for the stage
CHECKS=$(scenario_get_checks_for_stage "$STAGE" "$COMMAND")

if [[ "$CHECKS" == "[]" || -z "$CHECKS" ]]; then
    log_info "No $COMMAND checks defined for stage '$STAGE'"
    exit 0
fi

# Print header
log_info "Running $COMMAND checks for $TYPE/$SCENARIO (stage: $STAGE)"
log_info "Namespace: $NAMESPACE"
echo ""

# Run checks
CHECK_COUNT=$(echo "$CHECKS" | jq 'length')

for i in $(seq 0 $((CHECK_COUNT - 1))); do
    check_json=$(echo "$CHECKS" | jq -c ".[$i]")
    check_type=$(echo "$check_json" | jq -r '.type')
    
    # Apply filter if specified
    if [[ -n "$ONLY_FILTER" ]]; then
        case "$ONLY_FILTER" in
            playwright)
                if [[ "$check_type" != "playwright.run" ]]; then
                    continue
                fi
                ;;
            http)
                if [[ "$check_type" != http.* ]]; then
                    continue
                fi
                ;;
            k8s)
                if [[ "$check_type" != k8s.* ]]; then
                    continue
                fi
                ;;
            *)
                if [[ "$check_type" != "$ONLY_FILTER"* ]]; then
                    continue
                fi
                ;;
        esac
    fi
    
    if run_check "$check_json" "$NAMESPACE"; then
        CHECKS_PASSED=$((CHECKS_PASSED + 1))
    else
        CHECKS_FAILED=$((CHECKS_FAILED + 1))
        
        # For verify command, fail fast
        if [[ "$COMMAND" == "verify" ]]; then
            echo ""
            log_fail "Verification failed. Stopping."
            exit 1
        fi
    fi
done

# Summary
echo ""
log_info "Summary: $CHECKS_PASSED passed, $CHECKS_FAILED failed, $CHECKS_SKIPPED skipped"

# Exit code based on command type
if [[ "$COMMAND" == "verify" && "$CHECKS_FAILED" -gt 0 ]]; then
    exit 1
fi

# Health command exits 0 even with failures (observational)
exit 0
