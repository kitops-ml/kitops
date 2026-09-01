# KitOps (`kit`) — AI Agent Skill

## Overview

`kit` is the CLI for KitOps — a CNCF tool for packaging, versioning, and sharing AI/ML models
and agents. It bundles model weights, datasets, code, and configs into a versioned OCI artifact
called a **ModelKit**, stored in any standard container registry.

## Core Concepts

- **ModelKit**: An immutable, OCI-compatible bundle. Stored in any container registry (Docker Hub,
  AWS ECR, GitHub Packages, etc.). Every component is protected by SHA-256 digests.
- **Kitfile**: A YAML manifest (like a Dockerfile) describing what goes into a ModelKit. Lives in
  your project directory.

## Kitfile Format

```yaml
manifestVersion: 1.0.0

package:
  name: my-model
  version: 1.0.0
  description: A short description of the model
  authors:
    - My Name

model:
  name: My Model
  path: ./weights/model.bin   # relative to project root
  framework: pytorch
  license: apache-2.0

datasets:
  - name: training-data
    path: ./data/train.csv
    description: Training dataset

code:
  - path: ./src
    description: Training and inference code

docs:
  - path: ./README.md
```

All `path:` values are relative to the directory containing the Kitfile.

## Core Workflow

### 1. Initialize — generate a Kitfile

```bash
kit init .                                              # auto-generate from directory contents
kit init . --name mymodel --desc "My model"            # with name and description
kit init . --force                                      # overwrite existing Kitfile
kit init https://huggingface.co/org/model --remote     # generate from HuggingFace repo
```

### 2. Pack — create a ModelKit locally

```bash
kit pack .                                              # pack using Kitfile in current dir
kit pack . -t registry.io/org/mymodel:v1.0             # pack and tag in one step
kit pack . -f ./path/to/Kitfile -t registry.io/org/mymodel:v1.0  # custom Kitfile location
```

### 3. Push — upload to a registry

```bash
kit push registry.io/org/mymodel:v1.0
```

### 4. Pull — download from a registry

```bash
kit pull registry.io/org/mymodel:v1.0
```

### 5. Unpack — extract contents locally

```bash
kit unpack registry.io/org/mymodel:v1.0                          # unpack everything to current dir
kit unpack registry.io/org/mymodel:v1.0 -d ./output              # unpack to specific directory
kit unpack registry.io/org/mymodel:v1.0 --filter=model           # model weights only
kit unpack registry.io/org/mymodel:v1.0 --filter=datasets        # datasets only
kit unpack registry.io/org/mymodel:v1.0 --filter=code            # code only
kit unpack registry.io/org/mymodel:v1.0 --filter=model,datasets  # model and datasets
kit unpack registry.io/org/mymodel:v1.0 --filter=datasets:my-dataset  # specific dataset by name
```

## Other Commands

```bash
# Tag a ModelKit (like docker tag)
kit tag registry.io/org/mymodel:v1.0 registry.io/org/mymodel:latest

# List locally stored ModelKits
kit list

# Inspect a ModelKit's Kitfile without downloading all layers
kit inspect registry.io/org/mymodel:v1.0

# Import directly from HuggingFace into a ModelKit
kit import huggingface.co/org/model

# Login to a container registry
kit login registry.io

# Remove a local ModelKit
kit remove registry.io/org/mymodel:v1.0

# Show this skill file
kit skill
```

## Common Errors and Fixes

| Error | Cause | Fix |
|-------|-------|-----|
| `Kitfile already exists` | `kit init` when Kitfile is present | Add `--force` to overwrite |
| `authentication required` | Not logged in to registry | Run `kit login <registry>` first |
| `failed to resolve reference` | Tag not found locally or remotely | Verify with `kit list` or `kit inspect` |
| `no such file or directory` during pack | `path:` in Kitfile does not exist | Check all `path:` values are relative to the project root |
| `manifest unknown` | Tag doesn't exist in remote registry | Verify the tag exists in the registry |

## Reference Patterns

```bash
# Full pack-and-push pipeline
kit pack . -t registry.io/org/mymodel:v1.0 && kit push registry.io/org/mymodel:v1.0

# Pull and unpack only the model weights
kit pull registry.io/org/mymodel:v1.0
kit unpack registry.io/org/mymodel:v1.0 --filter=model -d ./weights

# Retag and push to a second registry
kit tag registry.io/org/mymodel:v1.0 other-registry.io/org/mymodel:v1.0
kit push other-registry.io/org/mymodel:v1.0
```

