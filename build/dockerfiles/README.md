# Containerized Kit CLI

The Dockerfiles in this directory can be used for building a containerized version of the Kit CLI. The default entrypoint
for these image is invoking the `kit` CLI, allowing for quickly running Kit commands without needing to install the CLI:
```
docker run --rm ghcr.io/kitops-ml/kitops:latest version
```

## Release build
The dockerfile named `release.Dockerfile` will build a container that includes a specified release of the Kit CLI, downloaded
from GitHub releases. To build this container to include Kit version `vX.Y.Z`, use the following command
```bash
docker build -t kit-cli:my-tag -f build/dockerfiles/release.Dockerfile --build-arg KIT_VERSION=vX.Y.Z .
```
Note, the `KIT_VERSION` build arg is required.


## Offline build
The dockerfile named `Dockerfile` will build a container by compiling the Kit CLI from local sources. To build this image
manually, run the following command from the root of this repository:

```bash
docker build -t kit-cli:next -f build/dockerfiles/Dockerfile .
```
By default, this container will set the Kit CLI version inside the container to `next`. To override this, you can specify
the build arg `KIT_VERSION`.
