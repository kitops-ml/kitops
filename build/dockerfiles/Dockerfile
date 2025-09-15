FROM docker.io/library/golang:1.25-alpine AS builder

RUN apk --no-cache upgrade && apk add --no-cache git
ARG KIT_VERSION=next

WORKDIR /build

# Cache dependencies in separate layer
COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY . .
RUN \
    CGO_ENABLED=0 go build \
    -o kit \
    -ldflags="-s -w -X github.com/kitops-ml/kitops/pkg/lib/constants.Version=${KIT_VERSION} -X github.com/kitops-ml/kitops/pkg/lib/constants.GitCommit=$(git rev-parse --short HEAD) -X github.com/kitops-ml/kitops/pkg/lib/constants.BuildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ')"

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

USER ${USER_ID}

COPY --from=builder /build/kit /usr/local/bin/kit

ENTRYPOINT ["kit"]

LABEL org.opencontainers.image.description="Kit CLI container built from source"
LABEL org.opencontainers.image.source="https://github.com/kitops-ml/kitops"
LABEL org.opencontainers.image.licenses="Apache-2.0"
