#!/bin/bash

set -e

# Kit configuration via env vars:
#   - UNPACK_PATH   : Where to unpack the modelkit (default: /home/user/modelkit)
#   - UNPACK_FILTER : What to unpack from the modelkit (default: model)
#   - EXTRA_FLAGS   : Additional flags for the unpack operation (default: --ignore-existing)
#   - MODELKIT_REF  : The modelkit to unpack -- required!
UNPACK_PATH=${UNPACK_PATH:-/home/user/modelkit/}
UNPACK_FILTER=${UNPACK_FILTER:-model}
if [ -z "$MODELKIT_REF" ]; then
  echo "Environment variable \$MODELKIT_REF is required"
  exit 1
fi

FLAGS_HAS_DIR="false"
read -r -a ADDITIONAL_UNPACK_FLAGS <<< "$EXTRA_FLAGS"
for flag in "${ADDITIONAL_UNPACK_FLAGS[@]}"; do
  if [[ "$flag" == "-d"* ]] || [[ "$flag" == "--dir"* ]]; then
    FLAGS_HAS_DIR="true"
  fi
done
if [ ${#ADDITIONAL_UNPACK_FLAGS[@]} == "0" ]; then
  ADDITIONAL_UNPACK_FLAGS[0]="--ignore-existing"
fi

# Variables for verifying signature via cosign. Can specify either a key to use for
# verifying or an identity and oidc issuer for keyless verification
if [ -n "$COSIGN_KEY" ]; then
  echo "Verifying signature for modelkit $MODELKIT_REF via key"
  cosign verify --key "$COSIGN_KEY" "$MODELKIT_REF"
elif [ -n "$COSIGN_CERT_IDENTITY" ] && [ -n "$COSIGN_CERT_OIDC_ISSUER" ]; then
  echo "Verifying signature for modelkit $MODELKIT_REF"
  cosign verify \
    --certificate-identity=${COSIGN_CERT_IDENTITY} \
    --certificate-oidc-issuer=${COSIGN_CERT_OIDC_ISSUER} \
    "$MODELKIT_REF"
else
  echo "Signature verification is disabled"
fi

echo "Binary version info:"
kit version

if [ "$FLAGS_HAS_DIR" == "false" ]; then
  echo "Unpacking modelkit $MODELKIT_REF to $UNPACK_PATH with filter '$UNPACK_FILTER'. Additional flags: ${ADDITIONAL_UNPACK_FLAGS[@]}"
  kit unpack "$MODELKIT_REF" --dir "$UNPACK_PATH" --filter="$UNPACK_FILTER" "${ADDITIONAL_UNPACK_FLAGS[@]}"
else
  echo "Unpacking modelkit $MODELKIT_REF with flags '--filter=$UNPACK_FILTER ${ADDITIONAL_UNPACK_FLAGS[@]}'"
  kit unpack "$MODELKIT_REF" --filter="$UNPACK_FILTER" "${ADDITIONAL_UNPACK_FLAGS[@]}"
fi
