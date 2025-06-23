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
elif [ -z "$KIT_USER" ] && [ -z "$KIT_PASSWORD" ] &&  [ -n "$AWS_ROLE_ARN" ]; then
  AWS_ACCOUNT_ID=$(echo "$AWS_ROLE_ARN" | cut -d: -f5)
  echo "Logging into AWS ECR $AWS_ACCOUNT_ID.dkr.ecr.$AWS_ECR_REGION.amazonaws.com"
  aws ecr get-login-password --region $AWS_ECR_REGION | kit login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_ECR_REGION.amazonaws.com
elif [ -z "$KIT_USER" ] && [ -z "$KIT_PASSWORD" ] && [ "$GCP_WIF" == "1" ]; then
  if [ -z "$GCP_GAR_LOCATION" ]; then
    echo "GCP_GAR_LOCATION env should be set and indicate the location of the GAR repository"
    exit 1
  fi
  echo "Logging into GCP Artifact Registry in location $GCP_GAR_LOCATION"
  gcloud auth print-access-token | kit login -u oauth2accesstoken --password-stdin $GCP_GAR_LOCATION-docker.pkg.dev
fi

echo "Unpacking $REPO_NAME to $OUTPUT_DIR"
echo "Unpack options: ${KIT_UNPACK_FLAGS}"
kit unpack "$REPO_NAME" -d "$OUTPUT_DIR" "${UNPACK_FLAGS[@]}"

if [ -f "$OUTPUT_DIR/Kitfile" ]; then
  echo "Unpacked modelkit:"
  cat "$OUTPUT_DIR/Kitfile"
fi
