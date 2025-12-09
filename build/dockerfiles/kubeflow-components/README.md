# Kubeflow Pipeline ModelKit Components

Kubeflow Pipeline components for packaging and deploying ML artifacts as KitOps ModelKits.

## Components

### push-modelkit

Packages ML artifacts in a directory as a ModelKit and pushes it to an OCI registry.

If a `Kitfile` exists in `modelkit_dir`, it is used as-is. Otherwise, one is auto-generated via `kit init`.

**Required inputs**

- `registry` – Container registry host (e.g., `registry.io`)
- `repository` – Repository path (e.g., `myorg/mymodel`)
- `tag` – ModelKit tag (default: `latest`)
- `modelkit_dir` – Directory with model files (with or without `Kitfile`)

**Optional metadata (for Kitfile)**

- `modelkit_name` – ModelKit package name
- `modelkit_desc` – ModelKit description
- `modelkit_author` – ModelKit author

**Optional attestation metadata**

- `dataset_uri` – Dataset URI
- `code_repo` – Code repository URL
- `code_commit` – Code commit hash

**Outputs**

- `ref` – Tagged ModelKit reference (e.g., `registry.io/myorg/mymodel:v1`)
- `digest` – Digest-based ModelKit reference (e.g., `registry.io/myorg/mymodel@sha256:abc…`)

### unpack-modelkit

Pulls a ModelKit from a registry and extracts it.

**Inputs**

- `modelkit_reference` – ModelKit reference (e.g., `registry.io/repo:tag` or `registry.io/repo@sha256:…`)
- `extract_path` – Directory to extract contents (default: `/tmp/model`)

**Outputs**

- `model_path` – Directory where contents were extracted

## Usage Examples

Complete, runnable examples (including a full house-prices pipeline) are in the [`examples/`](examples/) directory.

### Basic usage

Training component that writes ML artifacts to a directory:

```python
from kfp import dsl

@dsl.component(
    packages_to_install=['pandas', 'xgboost', 'scikit-learn'],
    base_image='python:3.11-slim',
)
def train_model(modelkit_dir: dsl.Output[dsl.Artifact]):
    """Train model and save to directory."""
    import os
    import pickle

    model = train_your_model()
    os.makedirs(modelkit_dir.path, exist_ok=True)

    with open(os.path.join(modelkit_dir.path, 'model.pkl'), 'wb') as f:
        pickle.dump(model, f)

    save_dataset(os.path.join(modelkit_dir.path, 'predictions.csv'))
    save_code(os.path.join(modelkit_dir.path, 'train.py'))
    save_docs(os.path.join(modelkit_dir.path, 'README.md'))
```

Component to push the directory as a ModelKit:

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
    dataset_uri: str = '',
    code_repo: str = '',
    code_commit: str = '',
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
            f'--dataset-uri "{dataset_uri}" '
            f'--code-repo "{code_repo}" '
            f'--code-commit "{code_commit}" '
            f'&& cp /tmp/outputs/reference "{output_ref.path}" '
            f'&& cp /tmp/outputs/digest "{output_digest.path}"'
        ],
    )
```

Simple end‑to‑end pipeline:

```python
@dsl.pipeline(
    name='simple-modelkit-pipeline',
    description='Train and package as ModelKit',
)
def simple_pipeline(
    registry: str = 'jozu.ml',
    repository: str = 'team/model',
    tag: str = 'latest',
):
    train = train_model()

    push = push_modelkit(
        registry=registry,
        repository=repository,
        tag=tag,
        input_modelkit_dir=train.outputs['modelkit_dir'],
        modelkit_name='My Model',
        modelkit_desc='Description of my model',
        modelkit_author='Data Science Team',
    )

    kubernetes.use_secret_as_volume(
        push,
        secret_name='docker-config',
        mount_path='/home/user/.docker',
    )
```

### Using a custom Kitfile

If you need full control, create a `Kitfile` alongside your artifacts:

```python
@dsl.component(base_image='python:3.11-slim')
def train_with_kitfile(modelkit_dir: dsl.Output[dsl.Artifact]):
    """Train and create custom Kitfile."""
    import os

    train_and_save_model(modelkit_dir.path)

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

When a `Kitfile` is present, the component uses it instead of generating one.

### Pipeline with attestation

```python
@dsl.pipeline(
    name='production-pipeline',
    description='Production pipeline with attestation',
)
def production_pipeline(
    registry: str = 'jozu.ml',
    repository: str = 'team/prod-model',
    tag: str = 'v1.0.0',
    dataset_uri: str = 's3://bucket/data.csv',
    code_repo: str = 'github.com/org/repo',
    code_commit: str = 'abc123',
):
    train = train_model()

    push = push_modelkit(
        registry=registry,
        repository=repository,
        tag=tag,
        input_modelkit_dir=train.outputs['modelkit_dir'],
        modelkit_name='Production Model',
        modelkit_desc='Production model v1.0.0',
        modelkit_author='ML Team',
        dataset_uri=dataset_uri,
        code_repo=code_repo,
        code_commit=code_commit,
    )

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

## Secret Requirements

### Registry credentials

Create a Kubernetes secret with Docker registry credentials:

```bash
kubectl create secret generic docker-config \
  --from-file=config.json="$HOME/.docker/config.json" \
  --namespace=kubeflow
```

Or:

```bash
kubectl create secret docker-registry docker-config \
  --docker-server=jozu.ml \
  --docker-username=myuser \
  --docker-password=mypassword \
  --docker-email=user@example.com \
  --namespace=kubeflow
```

Mount in your pipeline (as shown above) using:

```python
kubernetes.use_secret_as_volume(
    push,
    secret_name='docker-config',
    mount_path='/home/user/.docker',
)
```

### Cosign keys (optional)

For ModelKit attestation signing, create a secret with cosign keys:

```bash
cosign generate-key-pair

kubectl create secret generic cosign-keys \
  --from-file=cosign.key=cosign.key \
  --from-file=cosign.pub=cosign.pub \
  --namespace=kubeflow
```

Mount it as in the attestation pipeline example:

```python
kubernetes.use_secret_as_volume(
    push,
    secret_name='cosign-keys',
    mount_path='/etc/cosign',
)
```

If cosign keys are not available, the signing step logs a warning and continues.

## Troubleshooting

### Authentication errors

**Symptom:** `Failed to push ModelKit` or `401 Unauthorized`

**Check:**

```bash
kubectl get secret docker-config -n kubeflow
kubectl get secret docker-config -n kubeflow \
  -o jsonpath='{.data.config\.json}' | base64 -d
```

`config.json` should contain registry auth for your host:

```json
{
  "auths": {
    "jozu.ml": {
      "auth": "base64(username:password)"
    }
  }
}
```

### Directory not found

**Symptom:** `ModelKit directory does not exist`

Ensure your training component creates `modelkit_dir.path` and writes artifacts into it (see `train_model` example above).
