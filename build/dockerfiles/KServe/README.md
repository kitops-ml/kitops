# Kitops ClusterStorageContainer image for KServe

The Dockerfile in this directory is used to build an image that can run as a ClusterStorageContainer for [KServe](https://kserve.github.io/website/master/).

## Building
To build the image, `docker` or `podman` is required. From the root of this repository, set the `$KIT_KSERVE_IMAGE` to the image tag you want to build and run
```bash
docker build -t $KIT_KSERVE_IMAGE .
```

By default, the image will be built using `ghcr.io/kitops-ml/kitops:next` as a base. This can be overridden (to build using a specific version of Kit, for example) by using the build arg `KIT_BASE_IMAGE`:
```shell
# Build the image based on Kit v1.3.0 instead of 'next'
docker build -t kit-init-container:latest --build-arg KIT_BASE_IMAGE=ghcr.io/kitops-ml/kitops:v1.3.0 .
```

## Installing
The following process will create a new ClusterStorageContainer that uses `kit` to support KServe InferenceServices with storage URIs that have the `kit://` prefix.

Create the following ClusterStorageContainer custom resource in a Kubernetes cluster with KServe installed
```yaml
apiVersion: "serving.kserve.io/v1alpha1"
kind: ClusterStorageContainer
metadata:
  name: kitops
spec:
  container:
    name: kit-storage-initializer
    image: $KIT_KSERVE_IMAGE
    imagePullPolicy: Always
    env:
      - name: KIT_UNPACK_FLAGS
        value: "" # Add extra flags for the `kit unpack` command
    resources:
      requests:
        memory: 100Mi
        cpu: 100m
      limits:
        memory: 1Gi
    supportedUriFormats:
    - prefix: kit://
```

Once this CR is installed, modelkits can be used in InferenceServices by specifying with the `kit://` URI:
```yaml
storageUri: kit://<modelkit-reference>
```

## Configuration
The Kit KServe container supports specifying additional flags as supported by the `kit unpack` command. Additional flags are read from the `KIT_UNPACK_FLAGS` environment variable in the ClusterStorageContainer. For example, the following adds `-v` and `--plain-http` for all unpack commands:
```yaml
    env:
      - name: KIT_UNPACK_FLAGS
        value: "-v --plain-http"
```

## Additional links
* [KServe ClusterStorageContainer documentation](https://kserve.github.io/website/master/modelserving/storage/storagecontainers/)
