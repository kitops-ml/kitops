---
title: Using KitOps with Kubeflow Model Registry
description: Minimal quickstart for packaging a ModelKit with KitOps and storing it behind Kubeflow Model Registry using OCI-based storage.
keywords: kubeflow model registry, kitops, modelkit, oci model registry, kubeflow oci, modelpack kubeflow, model registry oci storage, kubeflow kitops integration
---

# Using KitOps with Kubeflow Model Registry

This guide shows a **small, end-to-end quickstart** for using KitOps with the Kubeflow Model Registry:

- Package a tiny sklearn model as a **ModelKit**.
- Push it to an **OCI registry**.
- Reference that same OCI location from **Kubeflow Model Registry**.

Once this is working, you can extend it into full pipelines, CI/CD, and richer metadata workflows.

## Architecture Overview

High-level flow:

```mermaid
flowchart LR
    A[Local Model (sklearn .pkl)] --> B[KitOps ModelKit]
    B --> C[OCI Registry (e.g., ECR / GCR / Harbor)]
    C --> D[Kubeflow Model Registry]
    D --> E[Kubeflow Pipeline / Notebook]
```

- **Local / CI**: Train and export a model (for example, a sklearn `.pkl` file).
- **KitOps**: Use `kit init` and `kit pack` to create a ModelKit.
- **OCI registry**: Store the ModelKit in any OCI-compatible registry.
- **Kubeflow Model Registry**: Register a model/version whose URI points at that OCI reference.
- **Kubeflow Pipelines / notebooks**: Resolve the version from the registry and use KitOps to unpack the ModelKit for inference or further processing.

## Prerequisites

- **KitOps CLI** installed — follow the [KitOps installation guide](/docs/cli/installation/).
- Access to an **OCI-compatible registry** (for example, `ghcr.io`, ECR, GCR, Harbor, Docker Hub, or JFrog Artifactory).
- A **Kubeflow cluster** with the **Model Registry** component installed. See Kubeflow’s [Model Registry getting started guide](https://www.kubeflow.org/docs/components/model-registry/getting-started/).
- `kubectl` configured for the cluster and rights to create or use registry credentials (Kubernetes Secrets).
- Python environment (with `scikit-learn` and `pickle`) available for training a tiny sklearn model.

> In this quickstart, replace `<registry>`, `<repo-namespace>`, `<k8s-namespace>`, and `my-modelkit` with values that match your environment.

## Quickstart (local)

This section focuses on **local steps**: install KitOps, package a tiny model (for example, `model.pkl`) as a ModelKit, then push it to an OCI registry that Kubeflow Model Registry can later reference.
For local cluster testing, you can use tools like `minikube` or `kind` with registry access configured. This is useful when validating Kubernetes Secrets and Kubeflow pipeline steps.

### 1. Install and verify the Kit CLI

Follow the official [installation instructions](/docs/cli/installation/) for your OS (Homebrew, ZIP/TAR, or build-from-source).

Verify the binary is available:

```bash
kit --help
```

(Optional) You can also validate these steps by building KitOps from source (using the “Build KitOps from Source” section in the install docs) and running `kit --help`.

### 2. Train and save a tiny sklearn model

Create and run a small Python script (for example, `train_and_save.py`) to produce `model.pkl`:

```python
from sklearn.linear_model import LogisticRegression
import pickle

X = [[0], [1], [2], [3]]
y = [0, 0, 1, 1]

model = LogisticRegression().fit(X, y)

with open("model.pkl", "wb") as f:
    pickle.dump(model, f)
```

After running the script you should have:

```bash
ls
# model.pkl
# train_and_save.py
```

### 3. Create and package a ModelKit

Create a directory to hold the ModelKit contents and generate a `Kitfile`:

```bash
mkdir my-modelkit
cp model.pkl my-modelkit/
cd my-modelkit

kit init . \
  --name "my-modelkit" \
  --desc "Minimal sklearn demo model" \
  --author "ML Platform Team"
```

This creates the minimal ModelKit metadata (a manifest file such as `Kitfile` or `modelkit.yaml`, depending on your KitOps version) that tells KitOps which files belong in the ModelKit.

Example manifest (illustrative — confirm manifest field names with the KitOps manifest schema in the CLI docs). For this tiny model it might look like:

```yaml
# illustrative manifest — confirm fields with the KitOps manifest schema in the CLI docs
manifestVersion: 1.0
package:
  name: my-modelkit
  description: Minimal sklearn demo model
  version: 0.1.0
  authors:
    - ML Platform Team
model:
  path: model.pkl
```

Now pack the directory into a ModelKit and tag it with your OCI reference:

```bash
kit pack . \
  -t <registry>/<repo-namespace>/my-modelkit:0.1.0
```

This builds a local ModelKit that is ready to be pushed.

### 4. Push the ModelKit to an OCI registry

First log in to the registry (only needed once per environment). The `kit` command examples below show typical flags; if you see errors, confirm the exact subcommands/flags for your version of KitOps in the [CLI reference](/docs/cli/cli-reference/):

```bash
kit login <registry> -u "<REGISTRY_USER>" -p "<REGISTRY_PASSWORD>"
```

If your registry uses Docker-native authentication, you can also run:

```bash
docker login <registry>
```

or use any registry-specific auth flow (for example, GitHub Container Registry with a personal access token).
> Example: GitHub Container Registry (GHCR) requires a personal access token (PAT) as the password.

Then push the tagged ModelKit:

```bash
kit push <registry>/<repo-namespace>/my-modelkit:0.1.0
```

**Note:** The `kit` commands shown here are illustrative. Confirm exact subcommands/flags for your installed KitOps CLI version in the [CLI reference](/docs/cli/cli-reference/).

At this point you have:

- A tiny sklearn model packaged as a **ModelKit**.
- Stored in your **OCI registry** under the reference  
  `<registry>/<repo-namespace>/my-modelkit:0.1.0`.

If your registry is private and you run Kubeflow in a Kubernetes cluster, create a Docker registry Secret so cluster workloads (including Kubeflow) can pull the ModelKit image:

```bash
kubectl create secret docker-registry kitops-regcred \
  --docker-server=<registry> \
  --docker-username=<username> \
  --docker-password=<password> \
  --docker-email=<email> \
  -n <k8s-namespace>
```

You can then reference this Secret via `imagePullSecrets` on a `ServiceAccount`, Pod spec, or Kubeflow workload so the cluster can pull from your registry.

For example, attach it to a `ServiceAccount` (replace `default` with the ServiceAccount used by your Kubeflow components or pipeline runtime):

```bash
kubectl patch serviceaccount default \
  -p '{"imagePullSecrets": [{"name": "kitops-regcred"}]}' \
  -n <k8s-namespace>
```

Replace `<repo-namespace>` (OCI repository namespace) and `<k8s-namespace>` (Kubernetes namespace) with values that match your environment.

### 5. Verify in Kubeflow Model Registry

With the ModelKit available in your registry, you can now reference it from Kubeflow Model Registry.

Use the official Kubeflow documentation to:

1. Access the Model Registry UI or API endpoint in your cluster.  
2. Create or register a model (for example, `my-modelkit`) and use the full OCI reference  
   `<registry>/<repo-namespace>/my-modelkit:0.1.0` as the backing URI or storage location.  
3. Confirm that the model and version appear in the Model Registry UI or by calling the Python client’s `list_models()` (or equivalent) in a notebook.

For detailed steps, see the Kubeflow [Model Registry getting started guide](https://www.kubeflow.org/docs/components/model-registry/getting-started/).

Once this is working, you can:

- Use Kubeflow Model Registry to track additional versions (`0.2.0`, `1.0.0`, etc.).
- Resolve the “current” version from notebooks or pipelines.
- Call KitOps (`kit pull` / `kit unpack`) inside Kubeflow workloads using that same OCI reference.

Common patterns for **Kubeflow-side authentication** include:

- Configuring `imagePullSecrets` on the `ServiceAccount` used by Kubeflow components.
- Storing registry credentials in Secrets used by controllers or pipeline runtimes.
- Mounting workload-specific Secrets into Pods that run `kit` commands.

**Tested with:**
- KitOps CLI version: TBD (TODO)
- Kubeflow version: TBD (TODO)
- Local test cluster: TBD (TODO)

## Links and References

- **KitOps docs**
  - [Install KitOps CLI](/docs/cli/installation/)
  - [Kit CLI reference](/docs/cli/cli-reference/)
  - [Getting started with ModelKits](/docs/get-started/)
- **Kubeflow Model Registry**
  - [Getting started](https://www.kubeflow.org/docs/components/model-registry/getting-started/)
  - [Python client reference](https://www.kubeflow.org/docs/components/model-registry/reference/python-client/)

If your documentation site does not render Mermaid diagrams, you can include a static PNG (for example, `kubeflow-model-registry-architecture.png`) as a fallback representation of the architecture.


