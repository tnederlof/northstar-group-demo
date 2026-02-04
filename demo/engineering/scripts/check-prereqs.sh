#!/usr/bin/env bash
# check-prereqs.sh - Check prerequisites for Engineering track demos

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

# Check docker compose specifically (subcommand, not a standalone command)
check_docker_compose() {
    if command -v docker &> /dev/null && docker compose version &> /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} docker-compose found"
    else
        echo -e "${RED}✗${NC} docker-compose not found"
        ((ERRORS++))
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
        echo "   This is only an issue if you're about to start the Engineering runtime."
    fi
}

echo "Checking prerequisites for Engineering track demos..."
echo ""

echo "Required commands:"
check_command docker
check_docker_compose
check_command git
check_command go
check_command node
check_command npm
check_command jq
check_command curl

echo ""
echo "Optional commands:"
check_command golangci-lint "golangci-lint (for linting)" false

echo ""
echo "Port availability:"
check_port 8082 "Engineering HTTP"
check_port 8083 "Engineering Dashboard"

echo ""
echo "Docker status:"
if docker info &> /dev/null; then
    echo -e "${GREEN}✓${NC} Docker is running"
else
    echo -e "${RED}✗${NC} Docker is not running"
    ((ERRORS++))
fi

echo ""
echo "Node version:"
if command -v node &> /dev/null; then
    node --version
fi

echo ""
echo "Go version:"
if command -v go &> /dev/null; then
    go version
fi

echo ""
if [[ $ERRORS -eq 0 ]]; then
    echo -e "${GREEN}All prerequisites met!${NC}"
    exit 0
else
    echo -e "${RED}$ERRORS prerequisite(s) not met${NC}"
    exit 1
fi
