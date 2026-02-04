#!/usr/bin/env bash
# scenario.sh - Shared scenario manifest library
#
# This library provides functions for loading and parsing scenario manifests.
# It is intended to be sourced by other scripts.
#
# Usage:
#   source "$(dirname "$0")/../shared/scripts/lib/scenario.sh"
#   scenario_load sre platform/bad-rollout
#   url_host=$(scenario_get_field url_host)
#
# Environment variables set after scenario_load:
#   SCENARIO_MANIFEST_PATH - absolute path to scenario.json
#   SCENARIO_DIR - directory containing scenario.json
#   SCENARIO_JSON - cached JSON content
#   SCENARIO_TYPE - sre or engineering
#   SCENARIO_TRACK - track (e.g., platform, backend)
#   SCENARIO_SLUG - slug (e.g., bad-rollout)

# Prevent double-sourcing
[[ -n "${_SCENARIO_LIB_LOADED:-}" ]] && return 0
_SCENARIO_LIB_LOADED=1

# Resolve paths (handle both bash and zsh)
if [[ -n "${BASH_SOURCE[0]:-}" ]]; then
    _SCENARIO_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
else
    # zsh fallback
    _SCENARIO_LIB_DIR="$(cd "$(dirname "${(%):-%x}")" && pwd)"
fi
_DEMO_DIR="$(cd "$_SCENARIO_LIB_DIR/../../.." && pwd)"

# Validate scenario path format (must be <track>/<slug>)
# Args: scenario_path
# Returns: 0 if valid, 1 if invalid
scenario_validate_path() {
    local scenario_path="$1"
    scenario_path="${scenario_path%/}"  # Remove trailing slash
    
    local depth
    depth=$(echo "$scenario_path" | tr -cd '/' | wc -c)
    
    if [[ "$depth" -ne 1 ]]; then
        echo "Error: Scenario path must be exactly 2 levels deep: <track>/<slug>" >&2
        echo "Got: $scenario_path" >&2
        return 1
    fi
    return 0
}

# Extract slug from scenario path
# Args: scenario_path
# Outputs: slug to stdout
scenario_extract_slug() {
    local scenario_path="$1"
    scenario_path="${scenario_path%/}"
    echo "${scenario_path##*/}"
}

# Extract track from scenario path
# Args: scenario_path
# Outputs: track to stdout
scenario_extract_track() {
    local scenario_path="$1"
    scenario_path="${scenario_path%/}"
    echo "${scenario_path%%/*}"
}

# Find scenario manifest path
# Args: type scenario_path
# Outputs: absolute path to scenario.json, or empty if not found
scenario_find_manifest() {
    local type="$1"
    local scenario_path="$2"
    scenario_path="${scenario_path%/}"
    
    local candidate=""
    
    if [[ -n "$type" ]]; then
        # Look in specific type directory
        candidate="$_DEMO_DIR/$type/scenarios/$scenario_path/scenario.json"
        if [[ -f "$candidate" ]]; then
            echo "$candidate"
            return 0
        fi
    else
        # Search both tracks
        for track_type in sre engineering; do
            candidate="$_DEMO_DIR/$track_type/scenarios/$scenario_path/scenario.json"
            if [[ -f "$candidate" ]]; then
                echo "$candidate"
                return 0
            fi
        done
    fi
    
    return 1
}

# Load scenario manifest and set environment variables
# Args: type scenario_path
# Sets: SCENARIO_MANIFEST_PATH, SCENARIO_DIR, SCENARIO_JSON, SCENARIO_TYPE, etc.
scenario_load() {
    local type="$1"
    local scenario_path="$2"
    
    # Validate path format
    scenario_validate_path "$scenario_path" || return 1
    
    # Find manifest
    local manifest
    manifest=$(scenario_find_manifest "$type" "$scenario_path")
    if [[ -z "$manifest" || ! -f "$manifest" ]]; then
        echo "Error: Scenario not found: $scenario_path" >&2
        if [[ -n "$type" ]]; then
            echo "Searched: $_DEMO_DIR/$type/scenarios/$scenario_path/scenario.json" >&2
        else
            echo "Searched in both sre and engineering tracks" >&2
        fi
        return 1
    fi
    
    # Validate JSON
    if ! jq empty "$manifest" 2>/dev/null; then
        echo "Error: Invalid JSON in $manifest" >&2
        return 1
    fi
    
    # Set environment variables
    export SCENARIO_MANIFEST_PATH="$manifest"
    export SCENARIO_DIR="$(dirname "$manifest")"
    export SCENARIO_JSON="$(cat "$manifest")"
    export SCENARIO_TYPE="$(echo "$SCENARIO_JSON" | jq -r '.type // empty')"
    export SCENARIO_TRACK="$(echo "$SCENARIO_JSON" | jq -r '.track // empty')"
    export SCENARIO_SLUG="$(echo "$SCENARIO_JSON" | jq -r '.slug // empty')"
    
    return 0
}

# Get a field from the loaded scenario manifest
# Args: field_path (jq path, e.g., "url_host" or ".checks.version")
# Outputs: field value to stdout
scenario_get_field() {
    local field="$1"
    
    if [[ -z "${SCENARIO_JSON:-}" ]]; then
        echo "Error: No scenario loaded. Call scenario_load first." >&2
        return 1
    fi
    
    # Handle both ".field" and "field" formats
    if [[ "$field" != .* ]]; then
        field=".$field"
    fi
    
    echo "$SCENARIO_JSON" | jq -r "$field // empty"
}

# Get URL host from loaded scenario
scenario_get_url_host() {
    scenario_get_field "url_host"
}

# Get base URL from loaded scenario (http://<url_host>:8080)
scenario_get_base_url() {
    local url_host
    url_host=$(scenario_get_url_host)
    if [[ -n "$url_host" ]]; then
        echo "http://$url_host:8080"
    fi
}

# Get checks object for a stage
# Args: stage (e.g., "broken", "healthy", "fixed")
# Outputs: JSON array of checks to stdout
scenario_get_checks_for_stage() {
    local stage="$1"
    local command="${2:-verify}"  # "verify" or "health"
    
    if [[ -z "${SCENARIO_JSON:-}" ]]; then
        echo "Error: No scenario loaded. Call scenario_load first." >&2
        return 1
    fi
    
    echo "$SCENARIO_JSON" | jq -c ".checks.stages.\"$stage\".\"$command\" // []"
}

# Get default stage from loaded scenario
scenario_get_default_stage() {
    local default_stage
    default_stage=$(scenario_get_field ".checks.default_stage")
    
    if [[ -n "$default_stage" ]]; then
        echo "$default_stage"
        return 0
    fi
    
    # If no default, try broken first, then healthy, then first defined
    local stages
    stages=$(echo "$SCENARIO_JSON" | jq -r '.checks.stages | keys[]' 2>/dev/null | head -1)
    
    for candidate in broken healthy "$stages"; do
        if [[ -n "$candidate" ]] && echo "$SCENARIO_JSON" | jq -e ".checks.stages.\"$candidate\"" >/dev/null 2>&1; then
            echo "$candidate"
            return 0
        fi
    done
    
    return 1
}

# Get list of available stages
# Outputs: newline-separated list of stages
scenario_get_stages() {
    if [[ -z "${SCENARIO_JSON:-}" ]]; then
        echo "Error: No scenario loaded. Call scenario_load first." >&2
        return 1
    fi
    
    echo "$SCENARIO_JSON" | jq -r '.checks.stages | keys[]' 2>/dev/null
}

# Check if scenario has checks defined
# Returns: 0 if checks exist, 1 otherwise
scenario_has_checks() {
    if [[ -z "${SCENARIO_JSON:-}" ]]; then
        return 1
    fi
    
    echo "$SCENARIO_JSON" | jq -e '.checks.version' >/dev/null 2>&1
}

# Get namespace for SRE scenarios (demo-<slug>)
scenario_get_namespace() {
    local slug
    slug="${SCENARIO_SLUG:-$(scenario_get_field slug)}"
    if [[ -n "$slug" ]]; then
        echo "demo-$slug"
    fi
}

# Get DEMO_LOGIN_KEY based on scenario type
# For SRE: reads from ConfigMap (requires kubectl context)
# For Engineering: reads from .state/global/secrets.env
# Args: [kube_context] (optional, for SRE)
scenario_get_demo_login_key() {
    local kube_context="${1:-kind-fider-demo}"
    local type="${SCENARIO_TYPE:-}"
    
    if [[ "$type" == "sre" ]]; then
        local namespace
        namespace=$(scenario_get_namespace)
        if [[ -n "$namespace" ]]; then
            kubectl --context="$kube_context" -n "$namespace" get configmap fider-env -o jsonpath='{.data.DEMO_LOGIN_KEY}' 2>/dev/null || echo ""
        fi
    elif [[ "$type" == "engineering" ]]; then
        local secrets_file="$_DEMO_DIR/.state/global/secrets.env"
        if [[ -f "$secrets_file" ]]; then
            grep '^DEMO_LOGIN_KEY=' "$secrets_file" 2>/dev/null | cut -d'=' -f2 || echo ""
        fi
    fi
}
