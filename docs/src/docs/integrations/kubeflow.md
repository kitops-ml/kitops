---
description: Learn how to use KitOps ModelKits in Kubeflow Pipelines by pulling and unpacking ModelKits as pipeline artifacts.
keywords: kitops kubeflow, modelkit kubeflow, modelpack kubeflow, kubeflow pipelines modelkit, oci model artifacts, ml pipelines, reproducible ml
---
# Integrating KitOps with Kubeflow

ModelKits let you package models, code, and datasets as OCI‑versioned artifacts with rich metadata and strong supply‑chain guarantees—while [Kubeflow](https://www.kubeflow.org/) provides an end‑to‑end platform for building and running ML workflows. Together, you can build reproducible pipelines that consume the exact ModelKit version you intended.

This guide shows how to pull and unpack a ModelKit inside a **Kubeflow Pipelines** step so downstream steps can use the extracted files.

## Prerequisites

- A Kubernetes cluster with [Kubeflow Pipelines](https://www.kubeflow.org/docs/components/pipelines/) installed.
- `kubectl` configured to access your cluster.
- A ModelKit pushed to an OCI registry (for example: `ghcr.io/<org>/<modelkit>:<tag>`).

## Option A: Use `kit` inside a pipeline step

The simplest approach is to run the `kit` CLI in a container step, pull the ModelKit, and unpack it to a shared path.

### Example: pipeline step (concept)

In your pipeline definition, create a step that:
- Logs into the registry (optional if the registry is public)
- Pulls the ModelKit
- Unpacks it to a shared folder (mounted volume)

Example commands:

```sh
kit login <registry> -u "$KIT_USER" -p "$KIT_PASSWORD"
kit pull <your-modelkit-reference>
kit unpack <your-modelkit-reference> -d /shared/modelkit
```

Downstream steps can then read files from `/shared/modelkit`.

### Registry credentials

Provide credentials as environment variables in the step:
- `KIT_USER`
- `KIT_PASSWORD`

In Kubeflow Pipelines, these are typically sourced from a Kubernetes Secret (how you wire this depends on your KFP version and SDK).

## Option B: Prebuild a pipeline image with `kit`

If you want fully reproducible pipelines, build a container image that includes the `kit` binary and any runtime dependencies your workflow needs, then use that image for the “pull + unpack” step.

## Notes

- ModelKit references can be any OCI reference supported by your registry (tagged or digested).
- For air‑gapped clusters or private registries, make sure the pipeline runtime can pull the container image used by the step and authenticate to the ModelKit registry.

