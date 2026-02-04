#!/usr/bin/env bash
# democtl.sh - Main orchestrator for demo commands
#
# This script provides high-level user-facing commands that abstract away
# the complexity of managing SRE and Engineering runtimes.
#
# Usage: democtl.sh <command> [args]
#
# Commands:
#   setup             - One-time setup (UI deps)
#   verify            - Check all prerequisites
#   run <scenario>    - Run a scenario (auto-detect track and ensure runtime)
#   reset <scenario>  - Reset a scenario
#   reset-all         - Master reset (removes all runtimes)
#   doctor            - Show status without failing

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEMO_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
SHARED_DIR="$DEMO_DIR/shared"
SRE_DIR="$DEMO_DIR/sre"
ENG_DIR="$DEMO_DIR/engineering"
UI_DIR="$DEMO_DIR/ui"
STATE_DIR="$DEMO_DIR/.state"
WORKTREE_DIR="$DEMO_DIR/.worktrees"

# Source the scenario library
# shellcheck source=lib/scenario.sh
source "$SCRIPT_DIR/lib/scenario.sh"

# Hardcoded ports
SRE_HTTP_PORT=8080
ENG_HTTP_PORT=8082
ENG_DASHBOARD_PORT=8083

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

KUBE_CONTEXT="${KUBE_CONTEXT:-kind-fider-demo}"

# ============================================================================
# Logging
# ============================================================================

log_info() {
    echo -e "${BLUE}==>${NC} $*"
}

log_success() {
    echo -e "${GREEN}==>${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}==>${NC} $*"
}

log_error() {
    echo -e "${RED}==>${NC} $*" >&2
}

log_section() {
    echo ""
    echo -e "${CYAN}$*${NC}"
    echo -e "${CYAN}$(printf '=%.0s' {1..80})${NC}"
}

# ============================================================================
# Port Availability
# ============================================================================

check_port_available() {
    local port="$1"
    local track_name="$2"
    
    if lsof -i ":$port" &> /dev/null; then
        local process
        process=$(lsof -i ":$port" -t 2>/dev/null | head -1 || echo "unknown")
        if [[ "$process" != "unknown" ]]; then
            local cmd
            cmd=$(ps -p "$process" -o comm= 2>/dev/null || echo "unknown")
            log_error "Port $port ($track_name) is in use by PID $process ($cmd)"
        else
        log_error "Port $port ($track_name) is in use"
    fi
    log_error "Stop the process using port $port and try again"
    return 1
    fi
    return 0
}

# ============================================================================
# SRE Runtime Ensure Functions
# ============================================================================

sre_cluster_exists() {
    kind get clusters 2>/dev/null | grep -q '^fider-demo$' && \
    kubectl --context="$KUBE_CONTEXT" cluster-info &>/dev/null
}

sre_ensure_cluster() {
    if sre_cluster_exists; then
        log_info "Kind cluster already exists (skipping creation)"
        return 0
    fi
    
    log_info "Creating Kind cluster..."
    check_port_available "$SRE_HTTP_PORT" "SRE HTTP" || return 1
    
    kind create cluster --config "$SRE_DIR/kind/cluster.yaml"
}

sre_gateway_ready() {
    # Check Gateway API CRDs
    kubectl --context="$KUBE_CONTEXT" api-resources 2>/dev/null | grep -q gatewayclasses || return 1
    
    # Check Envoy Gateway deployment
    kubectl --context="$KUBE_CONTEXT" -n envoy-gateway-system get deployment envoy-gateway &>/dev/null || return 1
    kubectl --context="$KUBE_CONTEXT" -n envoy-gateway-system rollout status deployment/envoy-gateway --timeout=10s &>/dev/null || return 1
    
    # Check shared Gateway resource
    kubectl --context="$KUBE_CONTEXT" -n envoy-gateway-system get gateway shared-gateway &>/dev/null || return 1
    
    return 0
}

sre_ensure_gateway() {
    if sre_gateway_ready; then
        log_info "Envoy Gateway already installed (skipping)"
        return 0
    fi
    
    log_info "Installing Gateway API CRDs..."
    kubectl --context="$KUBE_CONTEXT" apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.0/standard-install.yaml
    
    log_info "Installing Envoy Gateway controller..."
    helm upgrade --install envoy-gateway oci://docker.io/envoyproxy/gateway-helm \
        --kube-context="$KUBE_CONTEXT" \
        --version v1.2.4 \
        --namespace envoy-gateway-system --create-namespace \
        --skip-crds \
        --wait --timeout 5m
    
    log_info "Applying shared Gateway resources..."
    kubectl --context="$KUBE_CONTEXT" apply -f "$SRE_DIR/base/gateway.yaml"
}

sre_ensure_runtime() {
    log_section "Ensuring SRE Runtime"
    sre_ensure_cluster || return 1
    sre_ensure_gateway || return 1
    log_success "SRE runtime ready"
}

# ============================================================================
# Engineering Runtime Ensure Functions
# ============================================================================

eng_edge_ready() {
    docker ps --filter "name=northstar-edge" --filter "status=running" | grep -q northstar-edge
}

eng_network_exists() {
    docker network ls --filter "name=northstar-demo" | grep -q northstar-demo
}

eng_ensure_edge() {
    if eng_edge_ready; then
        log_info "Engineering edge proxy already running (skipping)"
        return 0
    fi
    
    log_info "Starting Engineering edge proxy..."
    check_port_available "$ENG_HTTP_PORT" "Engineering HTTP" || return 1
    check_port_available "$ENG_DASHBOARD_PORT" "Engineering Dashboard" || return 1
    
    docker compose -f "$ENG_DIR/compose/edge/docker-compose.yml" up -d
}

eng_ensure_network() {
    if eng_network_exists; then
        return 0
    fi
    
    # Network is created by edge proxy compose file
    return 0
}

eng_ensure_runtime() {
    log_section "Ensuring Engineering Runtime"
    eng_ensure_edge || return 1
    eng_ensure_network || return 1
    log_success "Engineering runtime ready"
}

# ============================================================================
# UI Ensure Functions
# ============================================================================

ui_ensure() {
    if [[ -d "$UI_DIR/node_modules" ]] && [[ -f "$UI_DIR/package-lock.json" ]]; then
        # Check if lockfile is newer than node_modules
        if [[ "$UI_DIR/package-lock.json" -nt "$UI_DIR/node_modules" ]]; then
            log_info "Installing UI dependencies (lockfile changed)..."
            cd "$UI_DIR"
            npm ci
        else
            log_info "UI dependencies already installed"
        fi
    else
        log_info "Installing UI dependencies..."
        cd "$UI_DIR"
        npm ci
    fi
    
    # Ensure Playwright browser is installed
    if ! npx playwright --version &>/dev/null || ! ls ~/.cache/ms-playwright/chromium-* &>/dev/null 2>&1; then
        log_info "Installing Playwright browser..."
        npx playwright install chromium
    fi
}

# ============================================================================
# Command: setup
# ============================================================================

cmd_setup() {
    log_section "Demo Setup"
    
    log_info "This command sets up UI testing dependencies."
    log_info "It does NOT start runtimes - use 'make run SCENARIO=...' for that."
    echo ""
    
    ui_ensure
    
    log_success "Setup complete!"
    echo ""
    log_info "Next steps:"
    log_info "  1. Run 'make verify' to check prerequisites"
    log_info "  2. Run 'make run SCENARIO=<track>/<slug>' to start a scenario"
    log_info "  3. See 'docs/GETTING_STARTED.md' for examples"
}

# ============================================================================
# Command: verify
# ============================================================================

cmd_verify() {
    log_section "Verifying Prerequisites"
    
    local errors=0
    
    # Validate scenarios first
    log_info "Validating scenario manifests..."
    if "$SHARED_DIR/scripts/validate-scenarios.sh" validate &>/dev/null; then
        echo -e "${GREEN}✓${NC} Scenario manifests are valid"
    else
        echo -e "${RED}✗${NC} Scenario manifest validation failed"
        ((errors++))
    fi
    
    echo ""
    log_info "Checking SRE prerequisites..."
    if "$SRE_DIR/scripts/check-prereqs.sh"; then
        :  # Success message already printed by script
    else
        ((errors++))
    fi
    
    echo ""
    log_info "Checking Engineering prerequisites..."
    if "$ENG_DIR/scripts/check-prereqs.sh"; then
        :  # Success message already printed by script
    else
        ((errors++))
    fi
    
    echo ""
    log_info "Checking UI prerequisites..."
    local ui_errors=0
    
    if [[ ! -d "$UI_DIR" ]]; then
        echo -e "${RED}✗${NC} UI directory not found"
        ((ui_errors++))
    else
        if [[ -d "$UI_DIR/node_modules" ]]; then
            echo -e "${GREEN}✓${NC} UI dependencies installed"
        else
            echo -e "${YELLOW}ℹ${NC} UI dependencies not installed (run 'make setup')"
        fi
    fi
    
    echo ""
    if [[ $errors -eq 0 ]]; then
        log_success "All required prerequisites met!"
        echo ""
        log_info "Ready to run scenarios with 'make run SCENARIO=<track>/<slug>'"
        return 0
    else
        log_error "$errors prerequisite check(s) failed"
        echo ""
        log_error "Please install missing tools and try again"
        log_error "See docs/GETTING_STARTED.md for installation instructions"
        return 1
    fi
}

# ============================================================================
# Command: run
# ============================================================================

cmd_run() {
    local scenario="$1"
    local verify="${VERIFY:-true}"
    
    if [[ -z "$scenario" ]]; then
        log_error "SCENARIO is required"
        echo "Usage: make run SCENARIO=<track>/<slug>"
        return 1
    fi
    
    log_section "Running Scenario: $scenario"
    
    # Load scenario manifest
    if ! scenario_load "" "$scenario"; then
        log_error "Failed to load scenario manifest"
        return 1
    fi
    
    local scenario_type="$SCENARIO_TYPE"
    local url_host="$(scenario_get_url_host)"
    
    if [[ "$scenario_type" == "sre" ]]; then
        # Ensure SRE runtime
        sre_ensure_runtime || return 1
        
        # Deploy scenario
        log_info "Deploying SRE scenario..."
        KUBE_CONTEXT="$KUBE_CONTEXT" "$SRE_DIR/scripts/demo.sh" up "$scenario" || return 1
        
        local scenario_url="http://$url_host:$SRE_HTTP_PORT"
        
    elif [[ "$scenario_type" == "engineering" ]]; then
        # Ensure Engineering runtime
        eng_ensure_runtime || return 1
        
        # Start scenario
        log_info "Starting Engineering scenario..."
        "$ENG_DIR/scripts/demo.sh" up "$scenario" || return 1
        
        local scenario_url="http://$url_host:$ENG_HTTP_PORT"
        
    else
        log_error "Unknown scenario type: $scenario_type"
        return 1
    fi
    
    # Run verification checks
    if [[ "$verify" == "true" ]]; then
        echo ""
        log_info "Running verification checks..."
        if KUBE_CONTEXT="$KUBE_CONTEXT" "$SHARED_DIR/scripts/run-checks.sh" "$scenario_type" "$scenario" verify; then
            log_success "Verification passed"
        else
            log_warn "Verification failed (scenario may still be starting)"
        fi
    fi
    
    # Print summary
    echo ""
    log_success "Scenario is running!"
    echo ""
    echo "  URL: $scenario_url"
    echo ""
    log_info "Next steps:"
    log_info "  • Visit the URL above to interact with the scenario"
    log_info "  • Run 'make health SCENARIO=$scenario' to check health"
    log_info "  • Run 'make reset SCENARIO=$scenario' to reset"
    log_info "  • Run 'make reset-all' to clean up everything"
}

# ============================================================================
# Command: reset
# ============================================================================

cmd_reset() {
    local scenario="$1"
    
    if [[ -z "$scenario" ]]; then
        log_error "SCENARIO is required"
        echo "Usage: make reset SCENARIO=<track>/<slug>"
        return 1
    fi
    
    log_section "Resetting Scenario: $scenario"
    
    # Load scenario manifest to determine type
    if ! scenario_load "" "$scenario"; then
        log_error "Failed to load scenario manifest"
        return 1
    fi
    
    local scenario_type="$SCENARIO_TYPE"
    
    if [[ "$scenario_type" == "sre" ]]; then
        KUBE_CONTEXT="$KUBE_CONTEXT" "$SRE_DIR/scripts/demo.sh" reset "$scenario"
    elif [[ "$scenario_type" == "engineering" ]]; then
        "$ENG_DIR/scripts/demo.sh" reset "$scenario"
    else
        log_error "Unknown scenario type: $scenario_type"
        return 1
    fi
    
    log_success "Scenario reset complete"
}

# ============================================================================
# Command: reset-all
# ============================================================================

cmd_reset_all() {
    local force="${FORCE:-false}"
    local nuke_ui="${NUKE_UI:-false}"
    
    log_section "Master Reset"
    
    if [[ "$force" != "true" ]]; then
        echo "This will:"
        echo "  • Stop all Engineering containers"
        echo "  • Delete the Kind cluster (SRE)"
        echo ""
        echo "With FORCE=true, it will also:"
        echo "  • Remove all worktrees and local state"
        echo ""
        echo "With FORCE=true and NUKE_UI=true, it will also:"
        echo "  • Remove UI node_modules and Playwright browsers"
        echo ""
        log_warn "This is a destructive operation"
        echo ""
        echo "To proceed, run: make reset-all FORCE=true"
        return 0
    fi
    
    log_warn "Proceeding with master reset (FORCE=true)"
    echo ""
    
    # Engineering cleanup
    log_info "Stopping Engineering containers..."
    if docker ps -a --filter "label=com.docker.compose.project=northstar" -q | grep -q .; then
        docker ps -a --filter "label=com.docker.compose.project=northstar" -q | xargs docker rm -f
    fi
    
    log_info "Stopping edge proxy..."
    docker compose -f "$ENG_DIR/compose/edge/docker-compose.yml" down 2>/dev/null || true
    
    log_info "Removing Docker network..."
    docker network rm northstar-demo 2>/dev/null || true
    
    # Remove worktrees
    if [[ -d "$WORKTREE_DIR" ]]; then
        log_info "Removing worktrees..."
        find "$WORKTREE_DIR" -mindepth 1 -maxdepth 1 -type d | while read -r wt; do
            "$ENG_DIR/scripts/worktree.sh" remove "$(basename "$wt")" 2>/dev/null || true
        done
        rm -rf "$WORKTREE_DIR"
    fi
    
    # SRE cleanup
    log_info "Removing SRE namespaces..."
    KUBE_CONTEXT="$KUBE_CONTEXT" "$SRE_DIR/scripts/demo.sh" down-all 2>/dev/null || true
    
    log_info "Deleting Kind cluster..."
    kind delete cluster --name fider-demo 2>/dev/null || true
    
    # State cleanup
    if [[ -d "$STATE_DIR" ]]; then
        log_info "Removing local state..."
        rm -rf "$STATE_DIR"
    fi
    
    # UI cleanup (only if NUKE_UI=true)
    if [[ "$nuke_ui" == "true" ]]; then
        log_info "Removing UI dependencies and Playwright browsers..."
        rm -rf "$UI_DIR/node_modules"
        rm -rf ~/.cache/ms-playwright
    fi
    
    log_success "Master reset complete!"
    echo ""
    log_info "To start fresh, run:"
    log_info "  make setup"
    log_info "  make run SCENARIO=<track>/<slug>"
}

# ============================================================================
# Command: doctor
# ============================================================================

cmd_doctor() {
    log_section "Demo Status"
    
    echo "SRE Runtime:"
    if sre_cluster_exists; then
        echo -e "  Cluster: ${GREEN}running${NC}"
        if sre_gateway_ready; then
            echo -e "  Gateway: ${GREEN}ready${NC}"
        else
            echo -e "  Gateway: ${YELLOW}not ready${NC}"
        fi
    else
        echo -e "  Cluster: ${YELLOW}not running${NC}"
    fi
    
    echo ""
    echo "Engineering Runtime:"
    if eng_edge_ready; then
        echo -e "  Edge Proxy: ${GREEN}running${NC}"
        echo -e "  HTTP Port: $ENG_HTTP_PORT"
        echo -e "  Dashboard: http://localhost:$ENG_DASHBOARD_PORT"
    else
        echo -e "  Edge Proxy: ${YELLOW}not running${NC}"
    fi
    
    if eng_network_exists; then
        echo -e "  Network: ${GREEN}exists${NC}"
    else
        echo -e "  Network: ${YELLOW}not created${NC}"
    fi
    
    echo ""
    echo "UI Testing:"
    if [[ -d "$UI_DIR/node_modules" ]]; then
        echo -e "  Dependencies: ${GREEN}installed${NC}"
    else
        echo -e "  Dependencies: ${YELLOW}not installed${NC}"
    fi
    
    echo ""
    echo "Port Configuration:"
    echo "  SRE HTTP: $SRE_HTTP_PORT"
    echo "  Engineering HTTP: $ENG_HTTP_PORT"
    echo "  Engineering Dashboard: $ENG_DASHBOARD_PORT"
    
    echo ""
    log_info "Run 'make verify' to check all prerequisites"
}

# ============================================================================
# Main
# ============================================================================

main() {
    local command="${1:-}"
    
    if [[ -z "$command" ]]; then
        echo "Usage: $0 <command> [args]"
        echo ""
        echo "Commands:"
        echo "  setup             - One-time setup (UI deps)"
        echo "  verify            - Check all prerequisites"
        echo "  run <scenario>    - Run a scenario (auto-detect track)"
        echo "  reset <scenario>  - Reset a scenario"
        echo "  reset-all         - Master reset (FORCE=true required)"
        echo "  doctor            - Show status"
        exit 1
    fi
    
    shift
    
    case "$command" in
        setup)
            cmd_setup "$@"
            ;;
        verify)
            cmd_verify "$@"
            ;;
        run)
            cmd_run "$@"
            ;;
        reset)
            cmd_reset "$@"
            ;;
        reset-all)
            cmd_reset_all "$@"
            ;;
        doctor)
            cmd_doctor "$@"
            ;;
        *)
            log_error "Unknown command: $command"
            exit 1
            ;;
    esac
}

main "$@"
