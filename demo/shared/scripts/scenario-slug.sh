#!/usr/bin/env bash
# scenario-slug.sh - Derives slug from SCENARIO path
#
# Extracts the slug (last component) from a SCENARIO path.
# Scenario paths must be exactly 2 levels deep: <track>/<slug>
#
# Usage: scenario-slug.sh <scenario>
# Example: scenario-slug.sh platform/bad-rollout
# Output: bad-rollout
#
# Valid: platform/bad-rollout, backend/ui-regression
# Invalid: bad-rollout (missing track), platform/sub/slug (too deep)

set -euo pipefail

usage() {
    echo "Usage: $0 <scenario>"
    echo "Example: $0 platform/bad-rollout"
    echo "Output: bad-rollout"
    exit 1
}

if [[ $# -ne 1 ]]; then
    usage
fi

SCENARIO="$1"

# Remove trailing slash if present
SCENARIO="${SCENARIO%/}"

# Count path depth (number of slashes)
DEPTH=$(echo "$SCENARIO" | tr -cd '/' | wc -c)

if [[ "$DEPTH" -ne 1 ]]; then
    echo "Error: Scenario path must be exactly 2 levels deep: <track>/<slug>" >&2
    echo "Got: $SCENARIO (depth: $((DEPTH + 1)))" >&2
    exit 1
fi

# Extract the slug (last component)
SLUG="${SCENARIO##*/}"

if [[ -z "$SLUG" ]]; then
    echo "Error: Could not extract slug from scenario: $SCENARIO" >&2
    exit 1
fi

echo "$SLUG"
