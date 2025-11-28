# Multi-platform digest for Cosign v2.4.0
ARG COSIGN_DIGEST=sha256:9d50ceb15f023eda8f58032849eedc0216236d2e2f4cfe1cdf97c00ae7798cfe
ARG KIT_BASE_IMAGE=ghcr.io/kitops-ml/kitops:next

FROM gcr.io/projectsigstore/cosign@$COSIGN_DIGEST AS cosign-install
FROM $KIT_BASE_IMAGE

# Install additional tools needed for scripts
USER 0
RUN apk add --no-cache \
    bash \
    jq
USER 1001

# Copy cosign from multi-platform build
COPY --from=cosign-install /ko-app/cosign /usr/local/bin/cosign

# Copy scripts (needs root for chmod)
USER 0
COPY scripts/ /scripts/
RUN chmod +x /scripts/*.sh
USER 1001

# Set working directory
WORKDIR /workspace

# Default entrypoint
ENTRYPOINT ["/bin/bash"]

LABEL org.opencontainers.image.description="KitOps Kubeflow Pipeline Components"
LABEL org.opencontainers.image.source="https://github.com/kitops-ml/kitops"
LABEL org.opencontainers.image.licenses="Apache-2.0"
