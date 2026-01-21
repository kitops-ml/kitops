---
title: Kubeflow Pipeline Components for ModelKits
description: Learn how to integrate KitOps ModelKits with Kubeflow Pipelines using push-modelkit and unpack-modelkit components.
keywords: kubeflow modelkit, kubeflow pipelines kitops, ml pipeline packaging, kubeflow oci, kubeflow modelpack, pipeline model versioning, kubeflow integration
---

# Integrating KitOps with Kubeflow Pipelines

KitOps provides Kubeflow Pipeline components that let you package ML artifacts as ModelKits directly within your pipeline workflows. This integration brings OCI-versioned, reproducible packaging to your Kubeflow training and deployment pipelines.

## Overview

The KitOps Kubeflow components enable you to:

1. **Package** trained models, datasets, code, and documentation as ModelKits
2. **Push** ModelKits to OCI registries with attestation metadata
3. **Pull** and unpack ModelKits for downstream pipeline steps
4. **Integrate** seamlessly with existing Kubeflow Pipeline workflows

## Prerequisites

- Kubeflow Pipelines installed on your Kubernetes cluster
- Access to an OCI registry (Docker Hub, Jozu Hub, etc.)
- `kubectl` configured to access your cluster
- Kubernetes secrets configured for registry authentication

## Complete Example

A full working example including training, packaging, and deployment is available in the [KitOps repository](https://github.com/kitops-ml/kitops/tree/main/build/dockerfiles/kubeflow-components/examples).

## Registry Authentication

Before using the KitOps components, set up registry authentication by creating a Kubernetes secret with your Docker registry credentials:

```bash
kubectl create secret generic docker-config \
  --from-file=config.json="$HOME/.docker/config.json" \
  --namespace=kubeflow
```

Or create a Docker registry secret:

```bash
kubectl create secret docker-registry docker-config \
  --docker-server=jozu.ml \
  --docker-username=myuser \
  --docker-password=mypassword \
  --docker-email=user@example.com \
  --namespace=kubeflow
```

Mount the secret in your pipeline component using:

```python
kubernetes.use_secret_as_volume(
    component_task,
    secret_name='docker-config',
    mount_path='/home/user/.docker',
)
```

## Cosign Signing (Optional)

To sign ModelKit attestations with Cosign, first generate signing keys and create a Kubernetes secret:

```bash
cosign generate-key-pair

kubectl create secret generic cosign-keys \
  --from-file=cosign.key=cosign.key \
  --from-file=cosign.pub=cosign.pub \
  --namespace=kubeflow
```

Mount the secret in your pipeline component:

```python
kubernetes.use_secret_as_volume(
    component_task,
    secret_name='cosign-keys',
    mount_path='/etc/cosign',
)
```

If cosign keys are not available, the signing step logs a warning and continues without failing the pipeline.

## Available Components

### push-modelkit

Packages ML artifacts in a directory as a ModelKit and pushes it to an OCI registry.

If a `Kitfile` exists in the model directory, it is used as-is. Otherwise, one is auto-generated via `kit init`.

**Required Inputs:**

- `registry` – Container registry host (e.g., `jozu.ml`)
- `repository` – Repository path (e.g., `myorg/mymodel`)
- `tag` – ModelKit tag (default: `latest`)
- `modelkit_dir` – Directory containing ML artifacts (with or without Kitfile)

**Optional Metadata:**

- `modelkit_name` – ModelKit package name
- `modelkit_desc` – ModelKit description
- `modelkit_author` – ModelKit author

**Optional Attestation Metadata:**

- `dataset_uri` – Dataset source URI
- `code_repo` – Code repository URL
- `code_commit` – Git commit hash

**Outputs:**

- `ref` – Tagged ModelKit reference (e.g., `jozu.ml/myorg/mymodel:v1`)
- `digest` – Digest-based ModelKit reference (e.g., `jozu.ml/myorg/mymodel@sha256:abc…`)

### unpack-modelkit

Pulls a ModelKit from a registry and extracts its contents.

**Inputs:**

- `modelkit_reference` – ModelKit reference (e.g., `jozu.ml/repo:tag` or `jozu.ml/repo@sha256:…`)
- `extract_path` – Directory to extract contents (default: `/tmp/model`)

**Outputs:**

- `model_path` – Directory where contents were extracted

## Quick Start Example

Here's how to use the `push-modelkit` component in your pipeline. The component wraps your trained model artifacts into a ModelKit and pushes it to a registry:

```python
from kfp import dsl, kubernetes

@dsl.container_component
def push_modelkit(
    registry: str,
    repository: str,
    tag: str,
    input_modelkit_dir: dsl.Input[dsl.Artifact],
    output_ref: dsl.Output[dsl.Artifact],
    output_digest: dsl.Output[dsl.Artifact],
    modelkit_name: str = '',
    modelkit_desc: str = '',
    modelkit_author: str = '',
):
    return dsl.ContainerSpec(
        image='ghcr.io/kitops-ml/kitops-kubeflow:latest',
        command=['/bin/bash', '-c'],
        args=[
            f'/scripts/push-modelkit.sh '
            f'"{registry}" "{repository}" "{tag}" '
            f'--modelkit-dir "{input_modelkit_dir.path}" '
            f'--name "{modelkit_name}" '
            f'--desc "{modelkit_desc}" '
            f'--author "{modelkit_author}" '
            f'&& cp /tmp/outputs/reference "{output_ref.path}" '
            f'&& cp /tmp/outputs/digest "{output_digest.path}"'
        ],
    )
```

Use it in your pipeline after your training step:

```python
# After your training component produces model artifacts
push = push_modelkit(
    registry='jozu.ml',
    repository='team/model',
    tag='latest',
    input_modelkit_dir=train.outputs['modelkit_dir'],
    modelkit_name='My Model',
    modelkit_desc='Description of my model',
    modelkit_author='Data Science Team',
)

# Mount registry credentials
kubernetes.use_secret_as_volume(
    push,
    secret_name='docker-config',
    mount_path='/home/user/.docker',
)
```

## Using Custom Kitfiles

If you need full control over the Kitfile structure, create one in your training component alongside your artifacts:

```python
kitfile_content = """
manifestVersion: 1.0
package:
  name: Custom Model
  description: Model with custom configuration
  authors:
    - Data Science Team
model:
  path: model.pkl
datasets:
  - path: train.csv
  - path: test.csv
code:
  - path: train.py
docs:
  - path: README.md
"""
with open(os.path.join(modelkit_dir.path, 'Kitfile'), 'w') as f:
    f.write(kitfile_content)
```

When a `Kitfile` is present, the component uses it instead of generating one automatically.

## Adding Attestation Metadata

For production pipelines, add attestation metadata to track data sources and code versions:

```python
push = push_modelkit(
    registry='jozu.ml',
    repository='team/prod-model',
    tag='v1.0.0',
    input_modelkit_dir=train.outputs['modelkit_dir'],
    modelkit_name='Production Model',
    modelkit_desc='Production model v1.0.0',
    modelkit_author='ML Team',
    dataset_uri='s3://bucket/data.csv',
    code_repo='github.com/org/repo',
    code_commit='abc123',
)

# Mount both registry credentials and cosign keys
kubernetes.use_secret_as_volume(
    push,
    secret_name='docker-config',
    mount_path='/home/user/.docker',
)
kubernetes.use_secret_as_volume(
    push,
    secret_name='cosign-keys',
    mount_path='/etc/cosign',
)
```

## Unpacking ModelKits

Use the `unpack-modelkit` component to pull and extract ModelKits in downstream pipeline steps:

```python
@dsl.container_component
def unpack_modelkit(
    modelkit_reference: str,
    extract_path: str,
    output_model_path: dsl.Output[dsl.Artifact],
):
    return dsl.ContainerSpec(
        image='ghcr.io/kitops-ml/kitops-kubeflow:latest',
        command=['/bin/bash'],
        args=[
            '/scripts/unpack-modelkit.sh',
            modelkit_reference,
            extract_path,
            f'&& cp /tmp/outputs/model_path "{output_model_path.path}"'
        ],
    )
```

Use it in your pipeline:

```python
unpack = unpack_modelkit(
    modelkit_reference='jozu.ml/team/model:v1',
    extract_path='/tmp/model',
)

kubernetes.use_secret_as_volume(
    unpack,
    secret_name='docker-config',
    mount_path='/home/user/.docker',
)

# Use unpacked model in subsequent steps
deploy = deploy_model(model_path=unpack.outputs['output_model_path'])
```
