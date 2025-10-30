---
title: ModelKit Overview - OCI Packaging for AI/ML Projects
description: Learn how ModelKit standardizes the packaging of models, datasets, and code for AI/ML workflows. OCI-compliant, versioned, and easy to use across registries and tools.
keywords: modelkit, modelkit overview, oci ai packaging, ml model registry, share ai model, package machine learning model, reproducible ml model, versioned ai artifact, docker for ai models, modelpack
---

# ModelKit Overview

![ModelKit](./ModelKit_chart.svg)

>**ModelKit is a standardized, OCI-compliant packaging format for AI/ML projects.**

It bundles everything your model needs — datasets, training code, config files, documentation, and the model itself — into a single shareable artifact.

Use ModelKits to version, share, and deploy AI models across teams and environments using familiar DevOps tools like DockerHub, GitHub Packages, or private registries.

➡️ [Get started with ModelKits](../../get-started.md) in under 15 minutes
➡️ [See how security-focused teams use ModelKits](../../use-cases.md)

## 🔑 Key Features

* **OCI-compliant and tool-friendly**
  Store, tag, and version ModelKits in any container registry — no custom infrastructure needed.

* **Selective unpacking**
  Unpack only the parts you need (e.g. just the dataset or model weights) to speed up pipelines and reduce compute overhead.

* **No duplication for shared assets**
  Reuse datasets or configs across multiple kits without bloating storage.

* **Familiar versioning and tagging**
  Use registry-native tags (e.g. `:latest`, `:prod`, `:rollback`) to track model state and history.

* **Built for ML workflows**
  Supports AI-specific needs like serialized model handling, reproducible training snapshots, and data lineage.

* **Streamlined collaboration**
  Teams can pull, inspect, and repack models just like container images — making it easier to collaborate across roles and environments.

## ⚡ Why It Matters

ModelKit simplifies the messy handoff between data scientists, engineers, and operations. It gives teams a common, versioned package that works across clouds, registries, and deployment setups — without reinventing storage or delivery.

It’s more than a format — it’s a building block for secure, reproducible AI.

---

**Have feedback or questions?**
Open an [issue on GitHub](https://github.com/kitops-ml/kitops/issues) or [join us on Discord](https://discord.gg/Tapeh8agYy).
