#!/usr/bin/env bash
# validate-scenarios.sh - Validates scenario manifests and powers list/describe commands
#
# Subcommands:
#   list              List all scenarios with their metadata
#   describe <path>   Show detailed info for a scenario
#   validate          Validate all scenario.json manifests and required files
#
# Scenario Manifest Schema (scenario.json):
#   Required keys: track, slug, title, type, url_host, seed, reset_strategy

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEMO_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Required fields in scenario.json
REQUIRED_FIELDS=("track" "slug" "title" "type" "url_host" "seed" "reset_strategy")

usage() {
    cat <<EOF
Usage: $0 <subcommand> [args]

Subcommands:
  list              List all scenarios with their metadata
  describe <path>   Show detailed info for a scenario (e.g., platform/bad-rollout)
  validate          Validate all scenario.json manifests and required files

Examples:
  $0 list
  $0 describe platform/bad-rollout
  $0 validate
EOF
    exit 1
}

# Find all scenario.json files
find_scenarios() {
    local track_dir="$1"
    find "$track_dir" -name "scenario.json" -type f 2>/dev/null | sort
}

# Validate a single scenario
validate_scenario() {
    local manifest="$1"
    local errors=0
    
    local scenario_dir
    scenario_dir="$(dirname "$manifest")"
    local rel_path="${scenario_dir#$DEMO_DIR/}"
    
    # Check if file is valid JSON
    if ! jq empty "$manifest" 2>/dev/null; then
        echo "  ERROR: Invalid JSON in $rel_path/scenario.json" >&2
        return 1
    fi
    
    # Check required fields
    for field in "${REQUIRED_FIELDS[@]}"; do
        if ! jq -e ".$field" "$manifest" >/dev/null 2>&1; then
            echo "  ERROR: Missing required field '$field' in $rel_path/scenario.json" >&2
            errors=$((errors + 1))
        fi
    done
    
    # Validate type matches track directory
    local type
    type=$(jq -r '.type' "$manifest")
    if [[ "$rel_path" == sre/* && "$type" != "sre" ]]; then
        echo "  ERROR: Scenario in sre/ directory but type is '$type' (expected 'sre')" >&2
        errors=$((errors + 1))
    elif [[ "$rel_path" == engineering/* && "$type" != "engineering" ]]; then
        echo "  ERROR: Scenario in engineering/ directory but type is '$type' (expected 'engineering')" >&2
        errors=$((errors + 1))
    fi
    
    # Validate path depth (should be <type>/scenarios/<track>/<slug>)
    local depth
    depth=$(echo "$rel_path" | tr -cd '/' | wc -c)
    if [[ "$depth" -ne 3 ]]; then
        echo "  ERROR: Scenario path must be exactly 4 levels deep: <type>/scenarios/<track>/<slug>" >&2
        echo "         Got: $rel_path (depth: $((depth + 1)))" >&2
        errors=$((errors + 1))
    fi
    
    return $errors
}

# List all scenarios
cmd_list() {
    local found=0
    
    echo "Available scenarios:"
    echo ""
    
    for track_dir in "$DEMO_DIR/sre" "$DEMO_DIR/engineering"; do
        if [[ ! -d "$track_dir/scenarios" ]]; then
            continue
        fi
        
        local track_name
        track_name="$(basename "$track_dir")"
        
        while IFS= read -r manifest; do
            [[ -z "$manifest" ]] && continue
            found=$((found + 1))
            
            local title track slug type
            title=$(jq -r '.title // "Untitled"' "$manifest")
            track=$(jq -r '.track // "unknown"' "$manifest")
            slug=$(jq -r '.slug // "unknown"' "$manifest")
            type=$(jq -r '.type // "unknown"' "$manifest")
            
            printf "  [%s] %s/%s: %s\n" "$type" "$track" "$slug" "$title"
        done < <(find_scenarios "$track_dir/scenarios")
    done
    
    if [[ $found -eq 0 ]]; then
        echo "  No scenarios found."
    fi
    
    echo ""
    echo "Total: $found scenario(s)"
}

# Describe a specific scenario
cmd_describe() {
    local scenario_path="$1"
    
    # Remove trailing slash
    scenario_path="${scenario_path%/}"
    
    # Look for the scenario in both tracks
    local manifest=""
    for track_dir in "$DEMO_DIR/sre" "$DEMO_DIR/engineering"; do
        local candidate="$track_dir/scenarios/$scenario_path/scenario.json"
        if [[ -f "$candidate" ]]; then
            manifest="$candidate"
            break
        fi
    done
    
    if [[ -z "$manifest" ]]; then
        echo "Error: Scenario not found: $scenario_path" >&2
        echo "Try: $0 list" >&2
        exit 1
    fi
    
    local scenario_dir
    scenario_dir="$(dirname "$manifest")"
    
    echo "Scenario: $scenario_path"
    echo "Location: $scenario_dir"
    echo ""
    echo "Manifest:"
    jq '.' "$manifest"
    
    echo ""
    echo "Files:"
    ls -la "$scenario_dir"
}

# Validate all scenarios
cmd_validate() {
    local found=0
    local errors=0
    
    echo "Validating scenarios..."
    echo ""
    
    for track_dir in "$DEMO_DIR/sre" "$DEMO_DIR/engineering"; do
        if [[ ! -d "$track_dir/scenarios" ]]; then
            continue
        fi
        
        while IFS= read -r manifest; do
            [[ -z "$manifest" ]] && continue
            found=$((found + 1))
            
            local rel_path
            rel_path="$(dirname "${manifest#$DEMO_DIR/}")"
            echo "Checking: $rel_path"
            
            if ! validate_scenario "$manifest"; then
                errors=$((errors + 1))
            fi
        done < <(find_scenarios "$track_dir/scenarios")
    done
    
    echo ""
    echo "Validated: $found scenario(s)"
    
    if [[ $errors -gt 0 ]]; then
        echo "Errors: $errors"
        exit 1
    else
        echo "All scenarios valid!"
    fi
}

# Main
if [[ $# -lt 1 ]]; then
    usage
fi

case "$1" in
    list)
        cmd_list
        ;;
    describe)
        if [[ $# -lt 2 ]]; then
            echo "Error: describe requires a scenario path" >&2
            usage
        fi
        cmd_describe "$2"
        ;;
    validate)
        cmd_validate
        ;;
    *)
        echo "Unknown subcommand: $1" >&2
        usage
        ;;
esac
