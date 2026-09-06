#!/bin/sh

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" >/dev/null 2>&1 && pwd)"
cd "$SCRIPT_DIR"

UI_DIR="../../../../frontend/dev-mode"

check_tool() {
    TOOL=$1
    HINT=$2
    if ! command -v "$TOOL" >/dev/null 2>&1; then
        echo "Error: '${TOOL}' is required but was not found on PATH." >&2
        if [ -n "$HINT" ]; then
            echo "$HINT" >&2
        fi
        exit 1
    fi
}

check_tool go "Please install Go: https://golang.org/doc/install"
check_tool node "Please install Node.js: https://nodejs.org/"
check_tool pnpm "Please install pnpm: https://pnpm.io/installation"
check_tool curl "Please install curl."
check_tool tar "Please install tar."

if ! command -v sha256sum >/dev/null 2>&1 && ! command -v shasum >/dev/null 2>&1; then
    echo "Error: neither 'sha256sum' nor 'shasum' was found on PATH." >&2
    exit 1
fi

LLAMAFILE_VER=$(go run ./llamafile_ver_helper.go)

build_ui() {
    UI_HOME=$1

    echo "Building harness UI from ${UI_HOME}"
    
    (
        cd "${UI_HOME}"
        pnpm install
        pnpm run build
    )
}

compress() {
    echo "Compressing $1 to $2"
    tar -czf "$2" -C "$1" .
}

generate_sha() {
    FILE=$1
    CHECKSUM_FILE=$2

    echo "Generating SHA256 checksum for ${FILE}"
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "${FILE}" | awk '{print $1 "  " FILENAME}' FILENAME="${FILE}" >> "${CHECKSUM_FILE}"
    else
        shasum -a 256 "${FILE}" | awk '{print $1 "  " FILENAME}' FILENAME="${FILE}" >> "${CHECKSUM_FILE}"
    fi
    echo "Checksum for ${FILE} saved to ${CHECKSUM_FILE}"
}


# Function to download a binary asset from a GitHub release
download_github_release_binary() {
    OWNER=$1
    REPO=$2
    RELEASE_TAG=$3
    ASSET_NAME=$4
    OUTPUT_DIR=$5

    echo "Downloading asset '${ASSET_NAME}' from release '${RELEASE_TAG}' of repository '${OWNER}/${REPO}'"

    DOWNLOAD_URL="https://github.com/${OWNER}/${REPO}/releases/download/${RELEASE_TAG}/llamafile-${RELEASE_TAG}"
    
    mkdir -p "${OUTPUT_DIR}"

    curl -L -o "${OUTPUT_DIR}/llamafile" "${DOWNLOAD_URL}"

    echo "Asset '${ASSET_NAME}' downloaded successfully to: ${OUTPUT_DIR}/${ASSET_NAME}"

    echo "${RELEASE_TAG}" > "${OUTPUT_DIR}/llamafile.version"

    COMPRESSED_FILE="llamafile.tar.gz"
    compress "${OUTPUT_DIR}" "../${COMPRESSED_FILE}"

    echo "Compressed asset saved to: ${COMPRESSED_FILE}"
}

build_ui "${UI_DIR}"
compress "${UI_DIR}/dist" "../ui.tar.gz"
download_github_release_binary "Mozilla-Ocho" "llamafile" "${LLAMAFILE_VER}" "llamafile-${LLAMAFILE_VER}" "./downloads"

CHECKSUM_FILE="../checksums.txt"
> "${CHECKSUM_FILE}"  # Clear the checksum file if it exists
generate_sha "../ui.tar.gz" "${CHECKSUM_FILE}"
generate_sha "../llamafile.tar.gz" "${CHECKSUM_FILE}"

echo "All checksums have been saved to ${CHECKSUM_FILE}"