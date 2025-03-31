# Multi-platform digest for Cosign v2.4.0
ARG COSIGN_DIGEST=sha256:9d50ceb15f023eda8f58032849eedc0216236d2e2f4cfe1cdf97c00ae7798cfe
ARG KIT_BASE_IMAGE=ghcr.io/kitops-ml/kitops:next

FROM gcr.io/projectsigstore/cosign@$COSIGN_DIGEST AS cosign-install
FROM $KIT_BASE_IMAGE

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
COPY --from=cosign-install /ko-app/cosign /usr/local/bin/cosign

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

LABEL org.opencontainers.image.description="Kit CLI init container"
LABEL org.opencontainers.image.source="https://github.com/kitops-ml/kitops"
LABEL org.opencontainers.image.licenses="Apache-2.0"
