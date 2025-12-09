#!/usr/bin/env bats

# Path to the script under test
SCRIPT_PATH="${BATS_TEST_DIRNAME}/../scripts/unpack-modelkit.sh"

setup() {
    # Create temporary directory for tests
    export TEST_DIR="$(mktemp -d)"
    export EXTRACT_DIR="$TEST_DIR/extract"
    export OUTPUT_DIR="/tmp/outputs"
    export LOG_LEVEL="INFO"
    export REQUEST_ID="test-unpack-modelkit"
    export DOCKER_CONFIG="$TEST_DIR/.docker"

    # Create mock docker config
    mkdir -p "$DOCKER_CONFIG"
    cat > "$DOCKER_CONFIG/config.json" << 'DOCKEREOF'
{"auths":{"registry.io":{"auth":"TU9DS19VU0VSOk1PQ0tfUEFTU1dPUkQ="}}}  # base64("MOCK_USER:MOCK_PASSWORD")
DOCKEREOF


    # Create output directory
    mkdir -p "$OUTPUT_DIR"

    # Mock commands
    export PATH="$TEST_DIR/bin:$PATH"
    mkdir -p "$TEST_DIR/bin"

    # Create mock kit command
    cat > "$TEST_DIR/bin/kit" << 'EOF'
#!/bin/bash
# Mock kit command for testing

if [[ "$1" == "version" ]]; then
    echo "kitops version v1.0.0"
    exit 0
fi

if [[ "$1" == "unpack" ]]; then
    reference="$2"
    # Parse -d flag for directory
    shift 2
    while [[ $# -gt 0 ]]; do
        case $1 in
            -d)
                dir="$2"
                shift 2
                ;;
            *)
                shift
                ;;
        esac
    done

    # Create mock unpacked content
    mkdir -p "$dir"
    echo "mock model content" > "$dir/model.bin"
    echo "mock kitfile" > "$dir/Kitfile"
    exit 0
fi

exit 1
EOF
    chmod +x "$TEST_DIR/bin/kit"

    # Create mock jq command
    cat > "$TEST_DIR/bin/jq" << 'EOF'
#!/bin/bash
# Forward to real jq
exec /usr/bin/jq "$@"
EOF
    chmod +x "$TEST_DIR/bin/jq"

    # Create failing kit command for error tests
    cat > "$TEST_DIR/bin/kit-fail" << 'EOF'
#!/bin/bash
exit 1
EOF
    chmod +x "$TEST_DIR/bin/kit-fail"
}

teardown() {
    # Clean up temporary directory
    rm -rf "$TEST_DIR"
    rm -rf "$OUTPUT_DIR"
    unset EXTRACT_DIR
    unset OUTPUT_DIR
    unset LOG_LEVEL
    unset REQUEST_ID
    unset DOCKER_CONFIG
}

# Argument validation tests

@test "fails when no arguments provided" {
    run bash "$SCRIPT_PATH"
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Usage:" ]]
    [[ "$output" =~ "modelkit_reference" ]]
}

@test "succeeds with only modelkit_reference (uses default extract path)" {
    run bash "$SCRIPT_PATH" "registry.io/myorg/mymodel:v1"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Unpack workflow completed" ]]
}

# ModelKit unpack tests

@test "successfully unpacks modelkit to specified path" {
    run bash "$SCRIPT_PATH" "registry.io/myorg/mymodel:v1" "$EXTRACT_DIR"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Starting unpack" ]]
    [[ "$output" =~ "Unpacking" ]]
    [ -f "$EXTRACT_DIR/model.bin" ]
}

@test "creates output file in /tmp/outputs" {
    run bash "$SCRIPT_PATH" "registry.io/myorg/mymodel:v1" "$EXTRACT_DIR"
    [ "$status" -eq 0 ]
    [ -f "$OUTPUT_DIR/model_path" ]
}

@test "output file contains correct value" {
    run bash "$SCRIPT_PATH" "registry.io/myorg/mymodel:v1" "$EXTRACT_DIR"
    [ "$status" -eq 0 ]

    path_content=$(cat "$OUTPUT_DIR/model_path")
    [[ "$path_content" == "$EXTRACT_DIR" ]]
}

@test "returns valid JSON output" {
    run bash "$SCRIPT_PATH" "registry.io/myorg/mymodel:v1" "$EXTRACT_DIR"
    [ "$status" -eq 0 ]

    # Extract final JSON output (the one with "status" field)
    json_output=$(echo "$output" | awk '/^{$/,/^}$/' | jq -s '.[] | select(.status != null)')
    echo "$json_output" | jq -e '.model_path'
    echo "$json_output" | jq -e '.modelkit_reference'
    echo "$json_output" | jq -e '.status == "success"'
}

# Error handling tests

@test "fails when kit command is not found" {
    export PATH="/usr/bin:/bin"
    run bash "$SCRIPT_PATH" "registry.io/myorg/mymodel:v1" "$EXTRACT_DIR"
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Required command not found: kit" ]]
}

@test "fails when DOCKER_CONFIG not set" {
    unset DOCKER_CONFIG
    run bash "$SCRIPT_PATH" "registry.io/myorg/mymodel:v1" "$EXTRACT_DIR"
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Required environment variable not set: DOCKER_CONFIG" ]]
}

@test "retries on unpack failure and eventually fails" {
    # Replace kit with failing version
    mv "$TEST_DIR/bin/kit" "$TEST_DIR/bin/kit.bak"
    mv "$TEST_DIR/bin/kit-fail" "$TEST_DIR/bin/kit"

    run bash "$SCRIPT_PATH" "registry.io/myorg/mymodel:v1" "$EXTRACT_DIR"
    [ "$status" -eq 1 ]
}

# Edge case tests

@test "handles paths with spaces" {
    extract_with_spaces="$TEST_DIR/extract with spaces"

    run bash "$SCRIPT_PATH" "registry.io/myorg/mymodel:v1" "$extract_with_spaces"
    [ "$status" -eq 0 ]
    [ -d "$extract_with_spaces" ]
    [ -f "$extract_with_spaces/model.bin" ]
}

# Integration tests

@test "creates extract directory if it does not exist" {
    nonexistent="$TEST_DIR/nonexistent/deep/path"

    run bash "$SCRIPT_PATH" "registry.io/myorg/mymodel:v1" "$nonexistent"
    [ "$status" -eq 0 ]
    [ -d "$nonexistent" ]
    [ -f "$nonexistent/model.bin" ]
}
