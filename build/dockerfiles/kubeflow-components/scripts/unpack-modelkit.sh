#!/bin/bash
set -euo pipefail

# Usage: /scripts/unpack-modelkit.sh <modelkit_uri> [extract_path]
# Environment variables: `DOCKER_CONFIG` (path to .docker directory containing config.json)
# Unpacks ModelKit artifacts to a directory

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

# Validate arguments
if [ $# -lt 1 ]; then
    die "Usage: $0 <modelkit_uri> [extract_path]"
fi

MODELKIT_URI="$1"
EXTRACT_PATH="${2:-/tmp/model}"

log_info "Starting unpack" "{\"modelkit_uri\":\"$MODELKIT_URI\",\"extract_path\":\"$EXTRACT_PATH\"}"

require_cmd kit jq
require_env DOCKER_CONFIG

# Disable kit update notifications to keep output clean
kit version --show-update-notifications=false >/dev/null 2>&1 || true

# Create output directory
mkdir -p /tmp/outputs
mkdir -p "$EXTRACT_PATH"

# Step 1: Unpack ModelKit with retry
log_info "Unpacking"
retry 3 2 kit unpack "$MODELKIT_URI" -d "$EXTRACT_PATH" || die "Failed to unpack ModelKit"

log_info "Unpacked successfully" "{\"path\":\"$EXTRACT_PATH\"}"

# Output results
# Write to KFP output file
echo -n "$EXTRACT_PATH" > /tmp/outputs/model_path

# Output JSON to stdout
jq -n \
    --arg path "$EXTRACT_PATH" \
    --arg uri "$MODELKIT_URI" \
    --arg timestamp "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
    '{
        "model_path": $path,
        "modelkit_uri": $uri,
        "timestamp": $timestamp,
        "status": "success"
    }'

log_info "Unpack workflow completed"
