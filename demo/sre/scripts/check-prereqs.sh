#!/usr/bin/env bash
# check-prereqs.sh - Check prerequisites for SRE demos

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

ERRORS=0

check_command() {
    local cmd="$1"
    local name="${2:-$cmd}"
    local required="${3:-true}"
    
    if command -v "$cmd" &> /dev/null; then
        echo -e "${GREEN}✓${NC} $name found"
    else
        echo -e "${RED}✗${NC} $name not found"
        if [[ "$required" == "true" ]]; then
            ((ERRORS++))
        fi
    fi
}

check_port() {
    local port="$1"
    local name="${2:-$port}"
    
    if ! lsof -i ":$port" &> /dev/null; then
        echo -e "${GREEN}✓${NC} $name (port $port) is available"
    else
        local process
        process=$(lsof -i ":$port" -t 2>/dev/null | head -1 || echo "unknown")
        if [[ "$process" != "unknown" ]]; then
            local cmd
            cmd=$(ps -p "$process" -o comm= 2>/dev/null || echo "unknown")
            echo -e "${GREEN}ℹ${NC} $name (port $port) is in use by PID $process ($cmd)"
        else
            echo -e "${GREEN}ℹ${NC} $name (port $port) is in use"
        fi
        echo "   This is only an issue if you're about to start the SRE runtime."
    fi
}

echo "Checking prerequisites for SRE demos..."
echo ""

echo "Required commands:"
check_command docker
check_command kind
check_command kubectl
check_command jq
check_command curl

echo ""
echo "Optional commands:"
check_command helm "helm (for Envoy Gateway)" false

echo ""
echo "Port availability:"
check_port 8080 "SRE HTTP"

echo ""
echo "Docker status:"
if docker info &> /dev/null; then
    echo -e "${GREEN}✓${NC} Docker is running"
else
    echo -e "${RED}✗${NC} Docker is not running"
    ((ERRORS++))
fi

echo ""
if [[ $ERRORS -eq 0 ]]; then
    echo -e "${GREEN}All prerequisites met!${NC}"
    exit 0
else
    echo -e "${RED}$ERRORS prerequisite(s) not met${NC}"
    exit 1
fi
