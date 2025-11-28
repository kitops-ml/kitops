#!/bin/bash

# Common library for workflow utilities

# Environment variables with defaults
export LOG_LEVEL="${LOG_LEVEL:-INFO}"
export REQUEST_ID="${REQUEST_ID:-}"

# Convert LOG_LEVEL to numeric value: DEBUG=0, INFO=1, WARN=2, ERROR=3
case "$LOG_LEVEL" in
    DEBUG) LOG_LEVEL_VALUE=0 ;;
    INFO) LOG_LEVEL_VALUE=1 ;;
    WARN) LOG_LEVEL_VALUE=2 ;;
    ERROR) LOG_LEVEL_VALUE=3 ;;
    *) LOG_LEVEL_VALUE=1 ;; # Default to INFO
esac
export LOG_LEVEL_VALUE

# Logging functions
log_json() {
    local level=$1
    local message=$2
    local extra="${3-}"
    if [ -z "$extra" ]; then extra="{}"; fi

    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    jq -nc \
        --arg timestamp "$timestamp" \
        --arg level "$level" \
        --arg request_id "${REQUEST_ID:-}" \
        --arg message "$message" \
        --argjson extra "$extra" \
        '{timestamp: $timestamp, level: $level, request_id: $request_id, message: $message, extra: $extra}'
}

log_debug() {
    local extra="${2-}"
    if [ -z "$extra" ]; then extra="{}"; fi
    [ "$LOG_LEVEL_VALUE" -le 0 ] && log_json "DEBUG" "$1" "$extra"
    return 0
}

log_info() {
    local extra="${2-}"
    if [ -z "$extra" ]; then extra="{}"; fi
    [ "$LOG_LEVEL_VALUE" -le 1 ] && log_json "INFO" "$1" "$extra"
}

log_warn() {
    local extra="${2-}"
    if [ -z "$extra" ]; then extra="{}"; fi
    [ "$LOG_LEVEL_VALUE" -le 2 ] && log_json "WARN" "$1" "$extra"
}

log_error() {
    local extra="${2-}"
    if [ -z "$extra" ]; then extra="{}"; fi
    [ "$LOG_LEVEL_VALUE" -le 3 ] && log_json "ERROR" "$1" "$extra" >&2
}

# Print error message and exit
die() {
    local extra="${2-}"
    if [ -z "$extra" ]; then extra="{}"; fi
    log_error "$1" "$extra"
    exit 1
}

# Retry logic
retry() {
    local max_attempts=${1:-3}
    local delay=${2:-2}
    shift 2
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        log_debug "Attempting command (attempt $attempt/$max_attempts)"

        if "$@"; then
            return 0
        fi

        if [ $attempt -lt $max_attempts ]; then
            log_warn "Command failed, retrying in ${delay}s" "{\"attempt\":$attempt}"
            sleep $delay
        fi

        attempt=$((attempt + 1))
    done

    log_error "Command failed after $max_attempts attempts"
    return 1
}

# Check required environment variables
require_env() {
    for var in "$@"; do
        if [ -z "${!var:-}" ]; then
            die "Required environment variable not set: $var"
        fi
    done
}

# Check required commands
require_cmd() {
    for cmd in "$@"; do
        if ! command -v "$cmd" &> /dev/null; then
            die "Required command not found: $cmd"
        fi
    done
}
