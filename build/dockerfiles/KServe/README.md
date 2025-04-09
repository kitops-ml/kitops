# KitOps ClusterStorageContainer image for KServe

The Dockerfile in this directory is used to build an image that can run as a ClusterStorageContainer for [KServe](https://kserve.github.io/website/master/).

## Installing

The following process creates a new ClusterStorageContainer that uses `kit` to support KServe InferenceServices with storage URIs that have the `kit://` prefix.

Create the following `ClusterStorageContainer` custom resource in a Kubernetes cluster with KServe installed. Note that the sample below uses the `ghcr.io/kitops-ml/kitops-kserve:next` image, although the repository includes other tags you can use.

```yaml
apiVersion: "serving.kserve.io/v1alpha1"
kind: ClusterStorageContainer
metadata:
  name: kitops
spec:
  container:
    name: storage-initializer
    image: ghcr.io/kitops-ml/kitops-kserve:next
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

Once this custom resource is installed, ModelKits can be used in InferenceServices by specifying the ModelKit URI with the `kit://` prefix in the `storageUri` field:

```yaml
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: "iris-model"
spec:
  predictor:
    model:
      modelFormat:
        name: sklearn
      storageUri: kit://<modelkit-reference>
```

## Building

To build the image, `docker` or `podman` is required. From the root of this repository, set the `$KIT_KSERVE_IMAGE`  environment variable to the image tag you want to build and run

```bash
docker build -t $KIT_KSERVE_IMAGE .
```

By default, the image will be built using `ghcr.io/kitops-ml/kitops:next` as a base. This can be overridden by specifying the build argument `KIT_BASE_IMAGE` to use a specific version of Kit. For example:

```shell
# Build the image based on Kit v1.3.0 instead of 'next'
docker build -t kitops-kserve:latest --build-arg KIT_BASE_IMAGE=ghcr.io/kitops-ml/kitops:v1.3.0 .
```

## Configuration

The Kit KServe container supports specifying additional flags for the `kit unpack` command. These flags are read from the KIT_UNPACK_FLAGS environment variable in the ClusterStorageContainer. For example, the following configuration adds `-v` and `--plain-http` for all unpack commands:

```yaml
    env:
      - name: KIT_UNPACK_FLAGS
        value: "-v --plain-http"
```

## Additional links

* [KServe ClusterStorageContainer documentation](https://kserve.github.io/website/master/modelserving/storage/storagecontainers/)
