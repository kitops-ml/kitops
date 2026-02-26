---
title: KitOps Integration with Kubeflow Model Registry
description: Learn how to use KitOps ModelKits with Kubeflow Model Registry for model versioning, discovery, and deployment in Kubeflow environments.
keywords: kubeflow model registry, modelkit kubeflow, oci model registry, ml model versioning, kubeflow kitops integration, model artifact management, kubeflow oci storage, modelpack kubeflow
---

# Integrating KitOps with Kubeflow Model Registry

KitOps ModelKits and the Kubeflow Model Registry both use OCI-compliant storage to manage machine learning artifacts. This guide shows how to package models with KitOps and register them in the Kubeflow Model Registry for tracking, versioning, and downstream use within Kubeflow workflows.

## Overview

This integration workflow:

1. **Packages** models, datasets, code, and configuration as OCI-compliant ModelKits using KitOps
2. **Pushes** ModelKits to an OCI-compatible registry
3. **Registers** the ModelKit URI in Kubeflow Model Registry for centralized tracking
4. **Discovers** and retrieves models via the Model Registry API
5. **Deploys** models using KServe or other Kubeflow components

```
Local Model → kit pack → OCI Registry → Model Registry → Kubeflow Pipelines/KServe
```

## Prerequisites

- [KitOps CLI installed](../../cli/installation.md)
- Kubernetes cluster with [Kubeflow](https://www.kubeflow.org/docs/started/installing-kubeflow/) installed
- [Kubeflow Model Registry](https://www.kubeflow.org/docs/components/model-registry/installation/) deployed
- Access to an OCI-compatible registry
- `kubectl` configured to access your cluster
- Python >= 3.9 (for Model Registry client)
- Kubeflow Pipelines SDK (`kfp`) installed for pipeline examples

> Note: If your OCI registry is private, configure Kubernetes imagePullSecrets or registry credentials for Kubeflow workloads before proceeding. See the **Registry Authentication** section below.

## Environment Setup

### Local Setup

Install and verify the KitOps CLI:

```sh
# Verify installation
kit version

# Log in to your OCI registry
kit login <REGISTRY>
```

### Kubeflow Setup

Install the Model Registry Python client in your notebook or pipeline environment:

```sh
pip install "model-registry>=0.3"
```

Verify connectivity to Model Registry from within a Kubeflow Notebook:

```sh
# Replace <MODEL_REGISTRY_URL> with your Model Registry service endpoint
curl <MODEL_REGISTRY_URL>/api/model_registry/v1alpha3/registered_models
```

## Workflow

### Step 1: Create and Package a ModelKit

Prepare your model artifacts and create a Kitfile:

```yaml
# Kitfile
manifestVersion: 1.0.0

package:
  name: iris-classifier
  version: 1.0.0
  description: Iris species classification model
  authors:
    - ML Team

model:
  name: sklearn-iris
  path: ./model/iris_model.joblib
  framework: scikit-learn
  version: "1.0"
  license: Apache-2.0

datasets:
  - name: training-data
    path: ./data/iris_train.csv
    description: Training dataset for iris classifier

docs:
  - path: ./README.md
```

Pack the ModelKit:

```sh
kit pack . -t <REGISTRY>/<ORG>/iris-classifier:v1.0.0
```

Verify the ModelKit:

```sh
kit list
```

### Step 2: Push to OCI Registry

Push the ModelKit to your registry:

```sh
kit push <REGISTRY>/<ORG>/iris-classifier:v1.0.0
```

Get the digest for immutable referencing:

```sh
kit inspect <REGISTRY>/<ORG>/iris-classifier:v1.0.0
```

### Step 3: Register in Kubeflow Model Registry

Use the Model Registry Python client to register the ModelKit:

```python
from model_registry import ModelRegistry

# Connect to Model Registry
registry = ModelRegistry(
    # Replace with your Model Registry service endpoint
    server_address="<MODEL_REGISTRY_URL>",
    port=8080,
    author="ml-team",
    is_secure=False
)

# Register the ModelKit with its OCI URI
rm = registry.register_model(
    "iris-classifier",
    "oci://<REGISTRY>/<ORG>/iris-classifier:v1.0.0",  # OCI URI to ModelKit
    model_format_name="sklearn",
    model_format_version="1",
    version="v1.0.0",
    description="Iris species classification model packaged with KitOps",
    metadata={
        "framework": "scikit-learn",
        "packaged_with": "kitops",
        "license": "Apache-2.0",
    }
)

print(f"Registered model with ID: {rm.id}")
```

### Step 4: Discover and Query Models

Retrieve model metadata from the registry:

```python
# Get the registered model
model = registry.get_registered_model("iris-classifier")
print(f"Model: {model.name}, ID: {model.id}")

# Get a specific version
version = registry.get_model_version("iris-classifier", "v1.0.0")
print(f"Version: {version.name}, ID: {version.id}")

# Get the artifact details (including the OCI URI)
artifact = registry.get_model_artifact("iris-classifier", "v1.0.0")
print(f"Artifact URI: {artifact.uri}")
```

List all registered models:

```python
models = registry.get_registered_models()
for m in models:
    print(f"- {m.name}")
```

### Step 5: Pull and Use ModelKits

Once you have the OCI URI from the registry, use KitOps to retrieve the model:

```sh
# Pull the ModelKit
kit pull <REGISTRY>/<ORG>/iris-classifier:v1.0.0

# Unpack to a directory
kit unpack <REGISTRY>/<ORG>/iris-classifier:v1.0.0 -d ./model-artifacts
```

Unpack only specific components:

```sh
# Unpack only the model
kit unpack <REGISTRY>/<ORG>/iris-classifier:v1.0.0 --filter=model -d ./model-only

# Unpack only datasets
kit unpack <REGISTRY>/<ORG>/iris-classifier:v1.0.0 --filter=datasets -d ./data-only
```

## Using ModelKits in Kubeflow Pipelines

### Pipeline Component for Unpacking ModelKits

Create a reusable component to unpack ModelKits within your pipeline:

```python
from kfp import dsl, kubernetes

@dsl.container_component
def unpack_modelkit(
    modelkit_reference: str,
    extract_path: str,
) -> dsl.ContainerSpec:
    return dsl.ContainerSpec(
        image='ghcr.io/kitops-ml/kitops:latest',
        command=['kit'],
        args=['unpack', modelkit_reference, '-d', extract_path, '-o'],
    )
```

Use in a pipeline:

```python
@dsl.pipeline(name="inference-pipeline")
def inference_pipeline(model_ref: str = "<REGISTRY>/<ORG>/iris-classifier:v1.0.0"):
    # Unpack the ModelKit
    unpack_task = unpack_modelkit(
        modelkit_reference=model_ref,
        extract_path="/tmp/model",
    )
    
    # Mount registry credentials
    kubernetes.use_secret_as_volume(
        unpack_task,
        secret_name='docker-config',
        mount_path='/home/user/.docker',
    )
    
    # Continue with inference using unpacked model...
```

### Integration with KServe

After registering a ModelKit in the Model Registry, deploy it with KServe. For KitOps-based model storage, use the [KitOps KServe integration](../kserve.md):

```yaml
apiVersion: serving.kserve.io/v1beta1
kind: InferenceService
metadata:
  name: iris-classifier
  labels:
    modelregistry/registered-model-id: "1"
    modelregistry/model-version-id: "1"
spec:
  predictor:
    model:
      modelFormat:
        name: sklearn
      storageUri: kit://<REGISTRY>/<ORG>/iris-classifier:v1.0.0
```

## Version Management

### Tagging Strategy

Use semantic versioning for ModelKit tags:

```sh
# Development versions
kit pack . -t <REGISTRY>/<ORG>/iris-classifier:dev
kit pack . -t <REGISTRY>/<ORG>/iris-classifier:v1.0.0-rc1

# Release versions
kit pack . -t <REGISTRY>/<ORG>/iris-classifier:v1.0.0
kit pack . -t <REGISTRY>/<ORG>/iris-classifier:latest
```

Register each version in Model Registry:

```python
# Register v1.0.0
registry.register_model(
    "iris-classifier",
    "oci://<REGISTRY>/<ORG>/iris-classifier:v1.0.0",
    model_format_name="sklearn",
    model_format_version="1",
    version="v1.0.0",
    description="Initial release",
)

# Register v1.1.0
registry.register_model(
    "iris-classifier",
    "oci://<REGISTRY>/<ORG>/iris-classifier:v1.1.0",
    model_format_name="sklearn",
    model_format_version="1",
    version="v1.1.0",
    description="Improved accuracy with additional training data",
)
```

### Digest-Based References

For production deployments, use digest-based references for immutability:

```python
# Get the digest from kit inspect output
digest_ref = "<REGISTRY>/<ORG>/iris-classifier@sha256:abc123..."

registry.register_model(
    "iris-classifier",
    f"oci://{digest_ref}",
    model_format_name="sklearn",
    model_format_version="1",
    version="v1.0.0-stable",
    description="Production-pinned release",
)
```

## Registry Authentication

### Creating Registry Secrets

For Kubeflow components to access your OCI registry:

```sh
kubectl create secret docker-registry oci-registry-creds \
  --docker-server=<REGISTRY> \
  --docker-username=myuser \
  --docker-password=mypassword \
  --docker-email=user@example.com \
  --namespace=kubeflow
```

Or use an existing Docker config:

```sh
kubectl create secret generic docker-config \
  --from-file=config.json="$HOME/.docker/config.json" \
  --namespace=kubeflow
```

### Using Secrets in Pipelines

Mount the secret in your pipeline components:

```python
kubernetes.use_secret_as_volume(
    unpack_task,
    secret_name='docker-config',
    mount_path='/home/user/.docker',
)
```

## End-to-End Example

Complete workflow from training to deployment:

```python
from model_registry import ModelRegistry
import subprocess

# 1. Pack the trained model
subprocess.run([
    "kit", "pack", ".",
    "-t", "<REGISTRY>/<ORG>/churn-predictor:v1.0.0"
], check=True)

# 2. Push to registry
subprocess.run([
    "kit", "push", "<REGISTRY>/<ORG>/churn-predictor:v1.0.0"
], check=True)

# 3. Register in Model Registry
registry = ModelRegistry(
    # Replace with your Model Registry service endpoint
    server_address="<MODEL_REGISTRY_URL>",
    port=8080,
    author="data-science-team",
    is_secure=False
)

rm = registry.register_model(
    "churn-predictor",
    "oci://<REGISTRY>/<ORG>/churn-predictor:v1.0.0",
    model_format_name="xgboost",
    model_format_version="1",
    version="v1.0.0",
    description="Customer churn prediction model",
    metadata={
        "accuracy": "0.94",
        "auc": "0.97",
        "training_date": "2025-02-01",
    }
)

# 4. Query the model
artifact = registry.get_model_artifact("churn-predictor", "v1.0.0")
print(f"Model registered at: {artifact.uri}")

# 5. Pull for inference
subprocess.run([
    "kit", "unpack", "<REGISTRY>/<ORG>/churn-predictor:v1.0.0",
    "-d", "/tmp/model", "-o"
], check=True)
```

## Troubleshooting

### Authentication Errors

**Symptom:** `unauthorized: authentication required` when pushing or pulling.

**Solution:**
```sh
# Re-authenticate with the registry
kit login <REGISTRY>

# Verify access by inspecting a remote artifact
kit inspect --remote <REGISTRY>/<ORG>/iris-classifier:v1.0.0
```

### Model Registry Connection Failures

**Symptom:** Cannot connect to Model Registry service.

**Solution:**
```sh
# Verify the service is running
kubectl get pods -n kubeflow -l app=model-registry

# Check service endpoint
kubectl get svc model-registry-service -n kubeflow
```

### ModelKit Not Found

**Symptom:** `manifest unknown` when pulling a ModelKit.

**Solution:**
```sh
# Verify the ModelKit exists
kit list <REGISTRY>/<ORG>/model-name

# Check the exact tag
kit inspect --remote <REGISTRY>/<ORG>/model-name:tag
```

### OCI URI Format

**Symptom:** Model Registry does not recognize the artifact URI.

**Solution:** Ensure the URI uses the correct format:
- For OCI artifacts: `oci://registry/repository:tag`
- For digests: `oci://registry/repository@sha256:...`

## Best Practices

1. **Version everything**: Tag ModelKits with semantic versions and register each version in Model Registry.

2. **Use digests in production**: Pin production deployments to digest-based references for immutability.

3. **Include metadata**: Add training metrics, hyperparameters, and lineage information when registering models.

4. **Automate registration**: Integrate ModelKit packing and registration into CI/CD pipelines.

5. **Secure credentials**: Use Kubernetes secrets for registry authentication; never hardcode credentials.

6. **Selective unpacking**: Use `--filter` flags to extract only the components you need, reducing storage and transfer overhead.

## Tested With

This guide was validated conceptually against:
- KitOps CLI: latest
- Kubeflow: v1.x
- Kubeflow Model Registry: v1alpha3 API

Update with exact versions after environment validation.