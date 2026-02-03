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
    
    if command -v "$cmd" &> /dev/null; then
        echo -e "${GREEN}✓${NC} $name found"
    else
        echo -e "${RED}✗${NC} $name not found"
        ((ERRORS++))
    fi
}

check_port() {
    local port="$1"
    
    if ! lsof -i ":$port" &> /dev/null; then
        echo -e "${GREEN}✓${NC} Port $port is available"
    else
        echo -e "${RED}✗${NC} Port $port is in use"
        ((ERRORS++))
    fi
}

echo "Checking prerequisites for Engineering track demos..."
echo ""

echo "Required commands:"
check_command docker
check_command "docker compose" "docker-compose"
check_command git
check_command go
check_command node
check_command npm
check_command jq
check_command curl

echo ""
echo "Optional commands:"
check_command golangci-lint "golangci-lint (for linting)"

echo ""
echo "Port availability:"
check_port 8080

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
