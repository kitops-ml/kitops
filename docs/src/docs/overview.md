---
title: What is KitOps? | Open-Source AI/ML Packaging and Deployment
description: KitOps is an open-source standard for packaging, versioning, and deploying AI/ML models. Built on OCI, it simplifies collaboration across data science, DevOps, and software teams.
keywords: kitops, what is kitops, modelkit, modelpack, ai model packaging, machine learning deployment, mlops packaging, OCI AI tools, AI reproducibility, AI CI/CD, open source ml packaging
---

# What is KitOps?

KitOps is an open-source tool that helps teams securely package, version, and deploy AI/ML models using familiar DevOps practices.

It’s built for data scientists, developers, and platform engineers working together on self-hosted AI/ML models — and it makes sharing, automating, and deploying those models as simple as managing containerized apps.

Just like PDFs standardized document sharing, KitOps standardizes how AI/ML projects are packaged, shared, and deployed.

It’s a format your tools can understand, your teams can trust, and your pipelines can automate.

## Why Use KitOps?

AI/ML models are more than code. They include model weights, datasets, configuration files, prompts, and documentation. Most teams store these assets in separate, disconnected repositories. This creates security gaps and breaks traceability between models, training data, and source notebooks.

KitOps packages everything your model needs for development or production in a single versioned OCI Artifact that you store in your container registry.

You can:
- Package models into deployable artifacts
- Share datasets and code securely between teams
- Automate packaging and deployment in CI/CD pipelines
- Run or test a model anywhere — without fragile setup steps

➡️ [Get started](../get-started.md) in minutes

## What’s Included

### 🎁 ModelKit: Standardized Model Packaging

The KitOps ModelKit is a packaging format that bundles all the artifacts of your AI/ML project — including datasets, code, configs, documentation, and the model itself — into an OCI-compliant Artifact.

This means ModelKits can be stored in your existing image registry, deployed to Kubernetes (or anywhere else containers run), and managed just like any container image.

**ModelPack Support:** KitOps can also create ModelPack-compliant packages using the [CNCF model-spec format](https://github.com/modelpack/model-spec). Both ModelPack and ModelKits are vendor-neutral standards for packaging everything needed for an AI/ML project. Kit commands (pull, push, unpack, inspect, list) work transparently with both formats.

See how to [deploy ModelKits](../deploy.md)

### 📄 Kitfile: Config Made Easy

The Kitfile is a YAML configuration that describes what goes into a KitOps ModelKit. It’s designed for clarity and security — making it easy to track what’s included, and to share AI/ML projects across environments and teams.

**Using ModelPack format:** To pack in ModelPack format instead of ModelKit format, add the `--use-model-pack` flag:

```sh
kit pack . --use-model-pack -t myregistry/mymodel:latest
```

When packing as ModelPack, KitOps stores your Kitfile as a manifest annotation so you can still retrieve it later. Commands like `kit unpack`, `kit pull`, and `kit inspect` work the same way regardless of format.

### 🖥️ Kit CLI: Create, Run, Automate

The Kit CLI is the command-line tool that ties everything together. Use it to:
- Create ModelKits from your local project
- Unpack and inspect existing kits
- Run models locally or in test environments
- Automate packaging in CI/CD pipelines

## Who Uses KitOps?

KitOps helps bridge the gap between experimentation and production for AI/ML workflows. Whether you’re running in the cloud, on-prem, or at the edge, KitOps makes it easier to collaborate across roles:

### For DevOps & Platform Engineers
- Use ModelKits in existing automation pipelines
- Store and manage models in your current container registry
- Build golden paths for secure AI/ML deployment

➡️ Integrate with [CI/CD](../integrations/cicd.md)

➡️ Add KitOps to [experiment trackers](../integrations/mlflow.md)

➡️ [Build a better golden path](../use-cases/) for AI/ML projects.

### For Data Scientists
- Package datasets and models without infrastructure hassle
- Share your work with developers without “it works on my machine” issues
- Keep code and data versions aligned

📺 [See how to use KitOps with Jupyter Notebooks](https://www.youtube.com/watch?v=OQPp7QEvk7Q).

➡️ Use the [PyKitOps Python SDK in your notebooks](../pykitops/index.md)

### For Developers
- Use AI/ML models like any dependency — no deep ML knowledge required
- Drop into apps using standard tools and APIs
- Let your team innovate without breaking your pipeline

➡️ [Get started](../get-started.md)

## A Standards-Based Approach

The goal of KitOps is to become the open, vendor-neutral standard that simplifies and secures the packaging and versioning of AI/ML projects. In the same way that PDFs have helped people share documents, images, and diagrams between tools, KitOps makes it easy for teams to use the tools they prefer, but share the results safely and securely.

KitOps is governed by the CNCF and supported by contributors from across the AI and DevOps ecosystem.

### Join the Community!
- Get help and share ideas in the [KitOps Discord](https://discord.gg/Tapeh8agYy)
- [Open an issue](https://github.com/kitops-ml/kitops/issues) on GitHub
- [Contribute](https://github.com/kitops-ml/kitops/blob/main/CONTRIBUTING.md) to the project
- Help shape the [ModelPack standards specification](https://github.com/modelpack/model-spec) for AI project packaging
