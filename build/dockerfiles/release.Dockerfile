FROM docker.io/library/alpine AS kit-download
ARG KIT_VERSION

RUN apk --no-cache upgrade && \
    apk add --no-cache bash curl

WORKDIR /kit-cli-download

RUN <<EOF
  set -e

  [ -n "$KIT_VERSION" ] || { echo "Build arg 'KIT_VERSION' is required"; exit 1; }

  export ARCH="$(uname -m | sed 's|aarch64|arm64|g')"

  echo "Downloading CLI and checksums"
  curl -fsSL \
    -o kit.tar.gz "https://github.com/kitops-ml/kitops/releases/download/$KIT_VERSION/kitops-linux-$ARCH.tar.gz" \
    -o checksums.txt "https://github.com/kitops-ml/kitops/releases/download/$KIT_VERSION/kitops_${KIT_VERSION}_checksums.txt" \

  echo "Checking SHA256 checksum"
  CHECKSUM=$(cat checksums.txt | grep "kitops-linux-$ARCH.tar.gz" | cut -f 1 -d ' ')
  echo "$CHECKSUM kit.tar.gz" | sha256sum -c

  echo "Extracting Kit CLI to $(pwd)/extracted/"
  mkdir -p ./extracted
  tar -xzvf kit.tar.gz -C ./extracted/
EOF

FROM docker.io/library/alpine

ENV USER_ID=1001 \
    USER_NAME=kit \
    HOME=/home/user/

RUN apk --no-cache upgrade && \
    apk add --no-cache bash git git-lfs && \
    git-lfs install --system && \
    mkdir -p /home/user/ && \
    adduser -D $USER_NAME -h $HOME -u $USER_ID

# Editor env var is required for kit import
ENV EDITOR=vi

COPY --from=kit-download /kit-cli-download/extracted/kit /usr/local/bin/kit
COPY --from=kit-download /kit-cli-download/extracted/README.md /kit-cli-download/extracted/LICENSE /

USER ${USER_ID}

ENTRYPOINT ["kit"]

LABEL org.opencontainers.image.description="Official release Kit CLI container"
LABEL org.opencontainers.image.source="https://github.com/kitops-ml/kitops"
LABEL org.opencontainers.image.licenses="Apache-2.0"
