---
description: Learn how to integrate KitOps ModelKits with KServe using a ClusterStorageContainer.
---
# Integrating KitOps with KServe

ModelKits let you treat your model as a first‑class, OCI‑versioned artifact with rich metadata and strong supply‑chain guarantees—while [KServe](https://kserve.github.io/website/) handles the runtime plumbing for scalable, pluggable inference. The result is a secure, reproducible, and operationally streamlined path from model build to model serving.

KitOps can be used as a `ClusterStorageContainer` in [KServe](https://kserve.github.io/website/) to serve ModelKits directly in InferenceServices. This guide shows how to install, configure, and build the KitOps container for KServe.

## Prerequisites

- A Kubernetes cluster with [KServe](https://kserve.github.io/website/) installed.
- `kubectl` configured to access your cluster.
- A built KitOps KServe image (for example, `ghcr.io/kitops-ml/kitops-kserve:latest`).

## Installing the ClusterStorageContainer

Create a `ClusterStorageContainer` resource that uses the KitOps CLI to unpack ModelKits when `storageUri` begins with `kit://`:

```yaml
apiVersion: serving.kserve.io/v1alpha1
kind: ClusterStorageContainer
metadata:
  name: kitops
spec:
  container:
    name: storage-initializer
    image: ghcr.io/kitops-ml/kitops-kserve:latest
    imagePullPolicy: Always
    env:
      - name: KIT_UNPACK_FLAGS
        value: "" # Additional flags for `kit unpack`
    resources:
      requests:
        memory: 100Mi
        cpu: 100m
      limits:
        memory: 1Gi
  supportedUriFormats:
    - prefix: kit://
```

After applying this manifest, any InferenceService that references a ModelKit URI with the `kit://` prefix will use KitOps to fetch and unpack the model.

## Using ModelKits in an InferenceService

Reference your ModelKit in the `storageUri` field of an InferenceService spec:

```yaml
apiVersion: serving.kserve.io/v1beta1
kind: InferenceService
metadata:
  name: iris-model
spec:
  predictor:
    model:
      modelFormat:
        name: sklearn
      storageUri: kit://<your-modelkit-reference>
```

KServe will invoke the KitOps container to unpack the ModelKit layers and make them available to your model server.

## Configuration Options

### Unpack Flags

Pass extra flags to the `kit unpack` command via the `KIT_UNPACK_FLAGS` environment variable. For example:

```yaml
env:
  - name: KIT_UNPACK_FLAGS
    value: "-v --plain-http"
```

### Registry Credentials

By default, the KitOps container uses `KIT_USER` and `KIT_PASSWORD` to log in. You can inject these from a Kubernetes Secret:

```yaml
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
```

### AWS ECR (IRSA)

If you use AWS ECR and leverage IRSA (IAM Roles for Service Accounts), set `AWS_ROLE_ARN` on the service account. KitOps will detect it and perform ECR login automatically, unless `KIT_USER` and `KIT_PASSWORD` are provided (which take precedence).

### GKE Workload Identity Federation

The KitOps KServe container also supports GKE login via Workload Identity Federation. To enable this feature, set the `GCP_WIF` environment variable to `true` and specify the `GCP_GAR_LOCATION` to indicate the location of your Google Artifact Registry (GAR) repository (for example, `europe` or `europe-west3`).

When these variables are set, KitOps will use Workload Identity Federation to authenticate with GKE—unless you provide `KIT_USER` and `KIT_PASSWORD`, which always take precedence.
