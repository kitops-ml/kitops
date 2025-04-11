#!/bin/bash

set -e

echo "Binary version info:"
kit version

read -r -a UNPACK_FLAGS <<< "$KIT_UNPACK_FLAGS"

if [ $# != 2 ]; then
  echo "Usage: entrypoint.sh <src-uri> <dest-path>"
  exit 1
fi

REPO_NAME="${1#kit://}"
OUTPUT_DIR="$2"

if [ -n "$KIT_USER" ] && [ -n "$KIT_PASSWORD" ]; then
    BASE_URL=$(echo "$REPO_NAME" | cut -d '/' -f 1)
    echo "Logging in using repo url: $BASE_URL"
    echo "$KIT_PASSWORD" | kit login "$BASE_URL" -u "$KIT_USER" --password-stdin
elif [ -z "$KIT_USER" ] && [ -z "$KIT_PASSWORD" ]; then
    ## no user and password provided, check if we are running in EKS with IRSA
    TOKEN_FILE="${AWS_WEB_IDENTITY_TOKEN_FILE:-/var/run/secrets/eks.amazonaws.com/serviceaccount/token}"
    if [ -f "$TOKEN_FILE" ]; then
        BASE_URL=$(echo "$REPO_NAME" | cut -d '/' -f 1)
        echo "Logging in using IRSA token to repo url: $BASE_URL"
        cat "$TOKEN_FILE" | kit login "$BASE_URL" -u "AWS" --password-stdin
    fi
fi

echo "Unpacking $REPO_NAME to $OUTPUT_DIR"
echo "Unpack options: ${KIT_UNPACK_FLAGS}"
kit unpack "$REPO_NAME" -d "$OUTPUT_DIR" "${UNPACK_FLAGS[@]}"

echo "Unpacked modelkit:"
cat "$OUTPUT_DIR/Kitfile"
