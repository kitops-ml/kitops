#!/usr/bin/env bats

# Path to the script under test
SCRIPT_PATH="${BATS_TEST_DIRNAME}/../scripts/push-modelkit.sh"

setup() {
    # Create temporary directory for tests
    export TEST_DIR="$(mktemp -d)"
    export MODEL_DIR="$TEST_DIR/model"
    export OUTPUT_DIR="/tmp/outputs"
    export LOG_LEVEL="INFO"
    export REQUEST_ID="test-push-modelkit"
    export DOCKER_CONFIG="$TEST_DIR/.docker"

    # Create mock model directory with Kitfile
    mkdir -p "$MODEL_DIR"
    echo "mock model content" > "$MODEL_DIR/model.bin"
    cat > "$MODEL_DIR/Kitfile" << 'KITFILEEOF'
manifestVersion: 1.0
package:
  name: test-model
model:
  path: model.bin
KITFILEEOF

    # Create mock docker config
    mkdir -p "$DOCKER_CONFIG"
    cat > "$DOCKER_CONFIG/config.json" << 'DOCKEREOF'
{"auths":{"registry.io":{"auth":"TU9DS19VU0VSOk1PQ0tfUEFTU1dPUkQ="}}} #  Mock auth is base64("MOCK_USER:MOCK_PASSWORD")
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

if [[ "$1" == "init" ]]; then
    # Mock kit init - create a basic Kitfile
    # Handle: kit init <dir> [--name NAME] [--desc DESC] [--author AUTHOR]
    dir="$2"
    shift 2
    # Consume optional flags
    while [[ $# -gt 0 ]]; do
        case $1 in
            --name|--desc|--author)
                shift 2  # Skip flag and value
                ;;
            *)
                shift
                ;;
        esac
    done
    cat > "$dir/Kitfile" << 'INITEOF'
manifestVersion: 1.0
package:
  name: auto-generated
model:
  path: model.bin
INITEOF
    exit 0
fi

if [[ "$1" == "pack" ]]; then
    # Mock pack output
    echo "Packing model..."
    exit 0
fi

if [[ "$1" == "push" ]]; then
    # Mock push output with digest
    echo "Pushed to registry"
    echo "Digest: sha256:abc123def456789012345678901234567890123456789012345678901234"
    exit 0
fi

if [[ "$1" == "inspect" ]]; then
    # Mock inspect output
    cat << 'INSPECTEOF'
{"digest":"sha256:abc123def456789012345678901234567890123456789012345678901234"}
INSPECTEOF
    exit 0
fi

exit 1
EOF
    chmod +x "$TEST_DIR/bin/kit"

    # Create mock cosign command
    cat > "$TEST_DIR/bin/cosign" << 'EOF'
#!/bin/bash
if [[ "$1" == "attest" ]]; then
    echo "Signing attestation..."
    exit 0
fi
exit 1
EOF
    chmod +x "$TEST_DIR/bin/cosign"

    # Create mock jq command
    cat > "$TEST_DIR/bin/jq" << 'EOF'
#!/bin/bash
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
    unset MODEL_DIR
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
}

@test "fails when only registry provided" {
    run bash "$SCRIPT_PATH" "registry.io"
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Usage:" ]]
}

@test "fails when only registry and repository provided" {
    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel"
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Usage:" ]]
}

@test "fails when no modelkit-dir specified" {
    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel" "v1"
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Must specify --modelkit-dir" ]]
}

# Directory mode tests

@test "successfully packs and pushes from directory with Kitfile" {
    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel" "v1" --modelkit-dir "$MODEL_DIR"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Using ModelKit directory" ]]
    [[ "$output" =~ "Packing ModelKit" ]]
    [[ "$output" =~ "Pushing to registry" ]]
}

@test "runs kit init when no Kitfile present" {
    # Remove Kitfile
    rm "$MODEL_DIR/Kitfile"

    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel" "v1" --modelkit-dir "$MODEL_DIR"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "No Kitfile found" ]]
    [[ "$output" =~ "running kit init" ]]
}

@test "recognizes lowercase kitfile" {
    # Replace Kitfile with lowercase kitfile
    mv "$MODEL_DIR/Kitfile" "$MODEL_DIR/kitfile"

    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel" "v1" --modelkit-dir "$MODEL_DIR"
    [ "$status" -eq 0 ]
    [[ ! "$output" =~ "No Kitfile found" ]]
    [[ "$output" =~ "Packing ModelKit" ]]
}

@test "recognizes dotfile .kitfile" {
    # Replace Kitfile with .kitfile
    mv "$MODEL_DIR/Kitfile" "$MODEL_DIR/.kitfile"

    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel" "v1" --modelkit-dir "$MODEL_DIR"
    [ "$status" -eq 0 ]
    [[ ! "$output" =~ "No Kitfile found" ]]
    [[ "$output" =~ "Packing ModelKit" ]]
}

@test "fails when directory does not exist" {
    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel" "v1" --modelkit-dir "/nonexistent"
    [ "$status" -eq 1 ]
    [[ "$output" =~ "ModelKit directory not found" ]]
}

@test "handles directory with spaces" {
    model_with_spaces="$TEST_DIR/model with spaces"
    mkdir -p "$model_with_spaces"
    echo "mock" > "$model_with_spaces/model.bin"
    cat > "$model_with_spaces/Kitfile" << 'EOF'
manifestVersion: 1.0
model:
  path: model.bin
EOF

    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel" "v1" --modelkit-dir "$model_with_spaces"
    [ "$status" -eq 0 ]
}

@test "passes metadata to kit init when no Kitfile exists" {
    # Remove Kitfile
    rm "$MODEL_DIR/Kitfile"

    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel" "v1" \
        --modelkit-dir "$MODEL_DIR" \
        --name "My Model" \
        --desc "Test model" \
        --author "Test Author"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "running kit init" ]]
}

# Output validation tests

@test "creates output files in /tmp/outputs" {
    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel" "v1" --modelkit-dir "$MODEL_DIR"
    [ "$status" -eq 0 ]
    [ -f "$OUTPUT_DIR/uri" ]
    [ -f "$OUTPUT_DIR/digest" ]
}

@test "output files contain correct values" {
    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel" "v1" --modelkit-dir "$MODEL_DIR"
    [ "$status" -eq 0 ]

    uri_content=$(cat "$OUTPUT_DIR/uri")
    digest_content=$(cat "$OUTPUT_DIR/digest")

    [[ "$uri_content" =~ registry.io/myorg/mymodel@sha256: ]]
    [[ "$digest_content" =~ sha256:abc123def456 ]]
}

@test "returns valid JSON output" {
    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel" "v1" --modelkit-dir "$MODEL_DIR"
    [ "$status" -eq 0 ]

    # Extract final JSON output
    json_output=$(echo "$output" | awk '/^{$/,/^}$/' | jq -s '.[] | select(.status != null)')
    echo "$json_output" | jq -e '.uri'
    echo "$json_output" | jq -e '.digest'
    echo "$json_output" | jq -e '.status == "success"'
}

# Attestation metadata tests

@test "accepts attestation metadata flags" {
    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel" "v1" \
        --modelkit-dir "$MODEL_DIR" \
        --dataset-uri "s3://bucket/data" \
        --code-repo "github.com/org/repo" \
        --code-commit "abc123"
    [ "$status" -eq 0 ]
}

# Cosign signing tests

@test "signs with cosign when key exists" {
    mkdir -p /tmp/etc/cosign
    echo "mock-key" > /tmp/etc/cosign/cosign.key

    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel" "v1" --modelkit-dir "$MODEL_DIR"
    [ "$status" -eq 0 ]

    rm -rf /tmp/etc/cosign
}

@test "warns when cosign key not found" {
    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel" "v1" --modelkit-dir "$MODEL_DIR"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "No cosign key found" ]] || [[ "$output" =~ "skipping signing" ]]
}

# Error handling tests

@test "fails when kit command is not found" {
    export PATH="/usr/bin:/bin"
    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel" "v1" --modelkit-dir "$MODEL_DIR"
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Required command not found: kit" ]]
}

@test "fails when DOCKER_CONFIG not set" {
    unset DOCKER_CONFIG
    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel" "v1" --modelkit-dir "$MODEL_DIR"
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Required environment variable not set: DOCKER_CONFIG" ]]
}

@test "retries on push failure and eventually fails" {
    # Replace kit with failing version
    mv "$TEST_DIR/bin/kit" "$TEST_DIR/bin/kit.bak"
    mv "$TEST_DIR/bin/kit-fail" "$TEST_DIR/bin/kit"

    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel" "v1" --modelkit-dir "$MODEL_DIR"
    [ "$status" -eq 1 ]
}

# Edge case tests

@test "handles registry with port number" {
    run bash "$SCRIPT_PATH" "registry.io:5000" "myorg/mymodel" "v1" --modelkit-dir "$MODEL_DIR"
    [ "$status" -eq 0 ]
}

@test "handles repository with nested path" {
    run bash "$SCRIPT_PATH" "registry.io" "myorg/team/project/mymodel" "v1" --modelkit-dir "$MODEL_DIR"
    [ "$status" -eq 0 ]
}

@test "handles tag with special characters" {
    run bash "$SCRIPT_PATH" "registry.io" "myorg/mymodel" "v1.0.0-rc1" --modelkit-dir "$MODEL_DIR"
    [ "$status" -eq 0 ]
}
