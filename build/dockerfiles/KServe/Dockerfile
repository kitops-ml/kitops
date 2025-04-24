ARG KIT_BASE_IMAGE=ghcr.io/kitops-ml/kitops:next
FROM $KIT_BASE_IMAGE

USER 0
RUN apk --no-cache upgrade && \
    apk add --no-cache aws-cli
USER 1001

COPY entrypoint.sh /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

LABEL org.opencontainers.image.description="KitOps KServe container"
LABEL org.opencontainers.image.source="https://github.com/kitops-ml/kitops"
LABEL org.opencontainers.image.licenses="Apache-2.0"
