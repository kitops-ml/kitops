#!/bin/bash
set -euo pipefail

# Usage: /scripts/push-modelkit.sh <registry> <repository> <tag> --modelkit-dir <path> [options]
#
# Arguments:
#   <registry>                  Container registry host (e.g., jozu.ml)
#   <repository>                Repository path (e.g., myorg/mymodel)
#   <tag>                       ModelKit tag
#   --modelkit-dir <path>       Directory with ML artifacts (with or without Kitfile)
#
# Options:
#   --name <name>               ModelKit name
#   --desc <description>        ModelKit description
#   --author <author>           ModelKit author
#   --dataset-uri <uri>         Dataset URI
#   --code-repo <repo>          Code repository
#   --code-commit <hash>        Code commit
#
# Environment variables: `DOCKER_CONFIG` (path to .docker directory containing config.json)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

# Initialize variables
REGISTRY=""
REPOSITORY=""
TAG=""
MODELKIT_DIR=""
MODELKIT_NAME=""
MODELKIT_DESC=""
MODELKIT_AUTHOR=""
DATASET_URI=""
CODE_REPO=""
CODE_COMMIT=""

# Parse arguments
if [ $# -lt 3 ]; then
    die "Usage: $0 <registry> <repository> <tag> --modelkit-dir <path> [options]"
fi

# First three args are positional
REGISTRY="$1"
REPOSITORY="$2"
TAG="$3"
shift 3

# Parse optional arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --modelkit-dir)
            MODELKIT_DIR="$2"
            shift 2
            ;;
        --name)
            MODELKIT_NAME="$2"
            shift 2
            ;;
        --desc)
            MODELKIT_DESC="$2"
            shift 2
            ;;
        --author)
            MODELKIT_AUTHOR="$2"
            shift 2
            ;;
        --dataset-uri)
            DATASET_URI="$2"
            shift 2
            ;;
        --code-repo)
            CODE_REPO="$2"
            shift 2
            ;;
        --code-commit)
            CODE_COMMIT="$2"
            shift 2
            ;;
        *)
            die "Unknown argument: $1"
            ;;
    esac
done

# Validate required arguments
if [ -z "$MODELKIT_DIR" ]; then
    die "Must specify --modelkit-dir"
fi

if [ ! -d "$MODELKIT_DIR" ]; then
    die "ModelKit directory not found: $MODELKIT_DIR"
fi

# Construct ModelKit reference
MODELKIT_REF="${REGISTRY}/${REPOSITORY}:${TAG}"

log_info "Starting ModelKit push" "{\"reference\":\"$MODELKIT_REF\"}"

require_cmd kit cosign jq

# Disable kit update notifications
kit version --show-update-notifications=false >/dev/null 2>&1 || true

# Create output directory
mkdir -p /tmp/outputs

# Use the provided directory as working directory
WORK_DIR="$MODELKIT_DIR"

log_info "Using ModelKit directory" "{\"dir\":\"$MODELKIT_DIR\"}"

# Check if Kitfile exists, if not run kit init
if [ ! -f "$WORK_DIR/Kitfile" ] && [ ! -f "$WORK_DIR/kitfile" ] && [ ! -f "$WORK_DIR/.kitfile" ]; then
    log_info "No Kitfile found, running kit init"

    INIT_ARGS=()
    [ -n "$MODELKIT_NAME" ] && INIT_ARGS+=(--name "$MODELKIT_NAME")
    [ -n "$MODELKIT_DESC" ] && INIT_ARGS+=(--desc "$MODELKIT_DESC")
    [ -n "$MODELKIT_AUTHOR" ] && INIT_ARGS+=(--author "$MODELKIT_AUTHOR")

    kit init "$WORK_DIR" ${INIT_ARGS[@]+"${INIT_ARGS[@]}"} || die "Failed to initialize Kitfile"
else
    log_info "Found existing Kitfile"
fi

# Pack the ModelKit
log_info "Packing ModelKit artifacts"
kit pack "$WORK_DIR" -t "$MODELKIT_REF" || die "Failed to pack ModelKit"

# Push to registry with retry
log_info "Pushing to registry"
retry 3 2 kit push "$MODELKIT_REF" || die "Failed to push ModelKit"

# Fetch digest from registry
log_debug "Fetching digest from registry"

set +e
INSPECT_OUTPUT=$(kit inspect "$MODELKIT_REF" --remote 2>&1)
INSPECT_EXIT_CODE=$?
set -e

log_debug "Kit inspect completed" "{\"exit_code\":$INSPECT_EXIT_CODE}"

if [ $INSPECT_EXIT_CODE -eq 0 ]; then
    MODELKIT_DIGEST=$(echo "$INSPECT_OUTPUT" | jq -r '.digest' 2>/dev/null || echo "")
fi

if [ -z "$MODELKIT_DIGEST" ]; then
    die "Could not determine ModelKit digest" "{\"reference\":\"$MODELKIT_REF\",\"exit_code\":$INSPECT_EXIT_CODE}"
fi

log_debug "ModelKit digest: $MODELKIT_DIGEST"

# Construct full reference with digest
FULL_REF="${REGISTRY}/${REPOSITORY}@${MODELKIT_DIGEST}"

log_info "Push completed" "{\"reference\":\"$FULL_REF\"}"

# Create in-toto attestation predicate
ATTESTATION_PREDICATE=$(jq -nc \
    --arg reference "$FULL_REF" \
    --arg digest "$MODELKIT_DIGEST" \
    --arg dataset_uri "$DATASET_URI" \
    --arg code_repo "$CODE_REPO" \
    --arg code_commit "$CODE_COMMIT" \
    --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    '{
        predicateType: "https://kitops.ml/attestation/v1",
        predicate: {
            modelkit: {
                reference: $reference,
                digest: $digest
            },
            metadata: {
                dataset_uri: $dataset_uri,
                code_repo: $code_repo,
                code_commit: $code_commit,
                created_at: $timestamp
            }
        }
    }')

log_debug "Created attestation predicate"

# Sign with cosign (non-fatal)
if [ -f "/etc/cosign/cosign.key" ]; then
    log_info "Signing and attaching attestation"

    PREDICATE_FILE=$(mktemp)
    echo "$ATTESTATION_PREDICATE" > "$PREDICATE_FILE"

    if retry 3 2 cosign attest \
        --key /etc/cosign/cosign.key \
        --predicate "$PREDICATE_FILE" \
        --tlog-upload=false \
        --yes \
        "$FULL_REF" 2>&1; then
        log_info "Signed with cosign"
    else
        log_warn "Failed to sign with cosign, continuing"
    fi

    rm -f "$PREDICATE_FILE"
else
    log_warn "No cosign key found at /etc/cosign/cosign.key, skipping signing"
fi

# Output results
# Write to KFP output files
echo -n "$MODELKIT_REF" > /tmp/outputs/reference         # Tagged reference (e.g., jozu.ml/repo:tag)
echo -n "$FULL_REF" > /tmp/outputs/digest          # Digest reference (e.g., jozu.ml/repo@sha256:...)

# Output JSON to stdout
jq -n \
    --arg reference "$FULL_REF" \
    --arg digest "$MODELKIT_DIGEST" \
    --arg timestamp "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
    '{
        "reference": $reference,
        "digest": $digest,
        "timestamp": $timestamp,
        "status": "success"
    }'

log_info "Push workflow completed"
