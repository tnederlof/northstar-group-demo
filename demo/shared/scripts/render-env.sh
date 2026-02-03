#!/usr/bin/env bash
# render-env.sh - Generates per-scenario runtime env files
#
# Generates runtime env files into demo/.state/ and prints the path
# to the generated file. Supports concurrent scenarios without manual
# env file management.
#
# Usage: render-env.sh <track> <scenario>
# Example: render-env.sh engineering backend/ui-regression
# Output: demo/.state/engineering/ui-regression/runtime.env
#
# Generated State Structure:
# - demo/.state/global/secrets.env - JWT_SECRET, DEMO_LOGIN_KEY (generated once)
# - demo/.state/<track>/<slug>/runtime.env - per-scenario env with contract keys + secrets

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEMO_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
SHARED_DIR="$DEMO_DIR/shared"
STATE_DIR="$DEMO_DIR/.state"

usage() {
    echo "Usage: $0 <track> <scenario>"
    echo "Example: $0 engineering backend/ui-regression"
    exit 1
}

if [[ $# -ne 2 ]]; then
    usage
fi

TRACK="$1"
SCENARIO="$2"

# Get the slug using scenario-slug.sh
SLUG="$("$SCRIPT_DIR/scenario-slug.sh" "$SCENARIO")"

# Paths
GLOBAL_SECRETS="$STATE_DIR/global/secrets.env"
CONTRACT_FILE="$SHARED_DIR/contract/fider.env.example"
OUTPUT_DIR="$STATE_DIR/$TRACK/$SLUG"
OUTPUT_FILE="$OUTPUT_DIR/runtime.env"

# Ensure contract file exists
if [[ ! -f "$CONTRACT_FILE" ]]; then
    echo "Error: Contract file not found: $CONTRACT_FILE" >&2
    exit 1
fi

# Generate global secrets if they don't exist
generate_secret() {
    # Generate a 32-character hex string
    if command -v openssl &> /dev/null; then
        openssl rand -hex 16
    else
        head -c 16 /dev/urandom | xxd -p
    fi
}

if [[ ! -f "$GLOBAL_SECRETS" ]]; then
    mkdir -p "$(dirname "$GLOBAL_SECRETS")"
    {
        echo "# Auto-generated secrets (do not commit)"
        echo "JWT_SECRET=$(generate_secret)"
        echo "DEMO_LOGIN_KEY=$(generate_secret)"
    } > "$GLOBAL_SECRETS"
fi

# Source the global secrets
# shellcheck source=/dev/null
source "$GLOBAL_SECRETS"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Generate runtime env file
{
    echo "# Generated runtime environment for $TRACK/$SLUG"
    echo "# Generated at: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
    echo ""
    
    # Read contract file and substitute values
    while IFS= read -r line || [[ -n "$line" ]]; do
        # Skip comments and empty lines (but preserve them)
        if [[ "$line" =~ ^[[:space:]]*# ]] || [[ -z "$line" ]]; then
            echo "$line"
            continue
        fi
        
        # Skip lines that are just comments about optional variables
        if [[ "$line" =~ ^[[:space:]]*#.* ]]; then
            continue
        fi
        
        # Extract key=value pairs (skip commented out optional vars)
        if [[ "$line" =~ ^([A-Z_]+)=(.*)$ ]]; then
            KEY="${BASH_REMATCH[1]}"
            VALUE="${BASH_REMATCH[2]}"
            
            # Substitute special values
            case "$VALUE" in
                *"<slug>"*)
                    VALUE="${VALUE//<slug>/$SLUG}"
                    ;;
                "<generated>")
                    case "$KEY" in
                        JWT_SECRET)
                            VALUE="$JWT_SECRET"
                            ;;
                        DEMO_LOGIN_KEY)
                            VALUE="$DEMO_LOGIN_KEY"
                            ;;
                        *)
                            VALUE="$(generate_secret)"
                            ;;
                    esac
                    ;;
            esac
            
            echo "$KEY=$VALUE"
        fi
    done < "$CONTRACT_FILE"
} > "$OUTPUT_FILE"

# Print the path to stdout
echo "$OUTPUT_FILE"
