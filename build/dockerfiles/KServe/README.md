# KitOps ClusterStorageContainer Image for KServe

This directory’s Dockerfile builds an image that runs as a `ClusterStorageContainer` for [KServe](https://kserve.github.io/website/).

## Installing

The following steps create a new `ClusterStorageContainer` that uses the `kit` CLI to support KServe `InferenceService` resources whose `storageUri` begins with `kit://`.

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

```bash
# Example: build against Kit v1.3.0 instead of "next"
docker build \
  -t kitops-kserve:latest \
  --build-arg KIT_BASE_IMAGE=ghcr.io/kitops-ml/kitops:v1.3.0 \
  .
```

## Configuration

### Unpack Flags

Set extra flags for `kit unpack` via the `KIT_UNPACK_FLAGS` environment variable. For example:

```yaml
    env:
      - name: KIT_UNPACK_FLAGS
        value: "-v --plain-http"
```

### Registry Credentials

By default, the container logs in to a registry using `KIT_USER` and `KIT_PASSWORD`. To pull credentials from a Kubernetes Secret:

```yaml
apiVersion: serving.kserve.io/v1alpha1
kind: ClusterStorageContainer
metadata:
  name: kitops
spec:
  container:
    name: storage-initializer
    image: ghcr.io/gorkem/kit-serve:latest
    imagePullPolicy: Always
    resources:
      requests:
        memory: 100Mi
        cpu: 100m
      limits:
        memory: 1Gi
    env:
      - name: KIT_USER
        valueFrom:
          secretKeyRef:
            name: kit-secret
            key: KIT_USER
            optional: true
      - name: KIT_PASSWORD
        valueFrom:
          secretKeyRef:
            name: kit-secret
            key: KIT_PASSWORD
            optional: true
  supportedUriFormats:
    - prefix: kit://
```

This example uses the `Secret` kit-secret but it can be modified to inject any secrets.

> **Tip:** The KitOps KServe container also supports AWS ECR login via IRSA. If `AWS_ROLE_ARN` is set on the service account, it takes effect—unless you have `KIT_USER` and `KIT_PASSWORD`, which take precedence.

> **Tip:** The KitOps KServe container also supports GKE login via Workload Identity Federation. If `GCP_WIF` is set to `true` and `GCP_GAR_LOCATION` indicates the location of the GAR repository (e.g. europe or europe-west3), it takes effect—unless you have `KIT_USER` and `KIT_PASSWORD`, which take precedence.

## Additional links

* [KServe ClusterStorageContainer Documentation](https://kserve.github.io/website/docs/model-serving/storage/storage-containers/)
