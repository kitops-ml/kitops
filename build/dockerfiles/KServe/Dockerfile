ARG KIT_BASE_IMAGE=ghcr.io/kitops-ml/kitops:next
FROM $KIT_BASE_IMAGE

USER 0

RUN apk --no-cache upgrade && \
    apk add --no-cache aws-cli curl

RUN mkdir -p /usr/local/gcloud && \
    curl -s https://dl.google.com/dl/cloudsdk/release/google-cloud-sdk.tar.gz | tar -xz -C /usr/local/gcloud && \
    /usr/local/gcloud/google-cloud-sdk/install.sh --quiet --usage-reporting false

ENV PATH="$PATH://usr/local/gcloud/google-cloud-sdk/bin"
ENV CLOUDSDK_CONFIG=/tmp/gcloud

USER 1001

COPY entrypoint.sh /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

LABEL org.opencontainers.image.description="KitOps KServe container"
LABEL org.opencontainers.image.source="https://github.com/kitops-ml/kitops"
LABEL org.opencontainers.image.licenses="Apache-2.0"
