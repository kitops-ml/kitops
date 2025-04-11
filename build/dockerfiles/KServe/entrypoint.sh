#!/bin/bash

set -e

echo "Binary version info:"
kit version

read -r -a UNPACK_FLAGS <<< "$KIT_UNPACK_FLAGS"
read -r AWS_ECR_REGION <<< "$AWS_ECR_REGION"
read -r AWS_ROLE_ARN <<< "$AWS_ROLE_ARN"

if [ $# != 2 ]; then
  echo "Usage: entrypoint.sh <src-uri> <dest-path>"
  exit 1
fi

REPO_NAME="${1#kit://}"
OUTPUT_DIR="$2"

if [ -n "$AWS_ROLE_ARN" ]; then
  AWS_ACCOUNT_ID=$(echo "$AWS_ROLE_ARN" | cut -d: -f5)
  echo "Logging into AWS ECR $AWS_ACCOUNT_ID.dkr.ecr.$AWS_ECR_REGION.amazonaws.com"
  aws ecr get-login-password --region $AWS_ECR_REGION | kit login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_ECR_REGION.amazonaws.com
fi

echo "Unpacking $REPO_NAME to $OUTPUT_DIR"
echo "Unpack options: ${KIT_UNPACK_FLAGS}"
kit unpack "$REPO_NAME" -d "$OUTPUT_DIR" "${UNPACK_FLAGS[@]}"

if [ -f "$OUTPUT_DIR/Kitfile" ]; then
  echo "Unpacked modelkit:"
  cat "$OUTPUT_DIR/Kitfile"
fi

