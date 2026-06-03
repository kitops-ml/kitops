---
title: What is KitOps? | Open-Source AI/ML Packaging and Deployment
description: KitOps is an open-source standard for packaging, versioning, and deploying AI/ML models. Built on OCI, it simplifies collaboration across data science, DevOps, and software teams.
keywords: kitops, what is kitops, modelkit, modelpack, ai model packaging, machine learning deployment, mlops packaging, OCI AI tools, AI reproducibility, AI CI/CD, open source ml packaging
---

# What is KitOps?

KitOps is an open-source tool for packaging, versioning, and sharing AI projects. It handles models, datasets, prompts, agent skill files, MCP server configurations, and code - bundling them into a single versioned OCI artifact that lives in your existing container registry.

Whether you’re deploying self-hosted models, building agentic AI systems, or managing prompt and skill libraries across teams, KitOps gives you one standard way to package, version, and distribute every artifact your AI project depends on.

Just like PDFs standardized document sharing, KitOps standardizes how AI projects are packaged, shared, and deployed. It’s a format your tools can understand, your teams can trust, and your pipelines can automate.

## Why Use KitOps?

AI projects are more than code. A self-hosted model needs weights, datasets, configuration files, and documentation. An agentic AI system adds prompts, skill files, and MCP server configurations to that list. Most teams store these assets in separate, disconnected places - Git for code, object storage for weights, hardcoded strings for prompts, scattered config files for MCP servers. This creates security gaps, breaks traceability, and makes it impossible to reproduce a known-good state.

KitOps packages everything your AI project needs into a single versioned OCI artifact stored in your container registry. When something breaks in production, you can answer: what exact combination of model, prompts, skills, and configs was running?

You can:
- Package models, prompts, skills, MCP configs, and datasets into versioned artifacts
- Share complete AI project snapshots securely between teams
- Automate packaging and deployment in CI/CD pipelines
- Reproduce any previous state by pulling a specific version

➡️ [Get started](../get-started.md) in minutes

## What’s Included

### ModelKit: Standardized AI Project Packaging

A ModelKit is an OCI-compliant artifact that bundles everything your AI project needs: models, datasets, code, prompts, agent skill files, MCP server configurations, and documentation. ModelKits work with any container registry and can be managed just like container images.

Not every ModelKit contains a model. A ModelKit might package a set of agent skills and their associated prompts, or an MCP server with its configuration. The format is flexible enough to handle any combination of AI project artifacts.

**ModelPack Support:** KitOps can also create ModelPack-compliant packages using the [CNCF model-spec format](https://github.com/modelpack/model-spec). Both ModelPack and ModelKits are vendor-neutral standards for packaging AI projects. Kit commands (pull, push, unpack, inspect, list) work transparently with both formats.

See how to [deploy ModelKits](../deploy.md)

### Kitfile: Define What Gets Packaged

The Kitfile is a YAML manifest that describes what goes into a ModelKit. It defines the components of your AI project - models, datasets, code (including MCP server configurations), prompts (including agent skill files), and documentation. It’s designed for clarity and security, making it easy to track what’s included and share projects across environments and teams.

**Using ModelPack format:** To pack in ModelPack format instead of ModelKit format, add the `--use-model-pack` flag:

```sh
kit pack . --use-model-pack -t myregistry/mymodel:latest
```

When packing as ModelPack, KitOps stores your Kitfile as a manifest annotation so you can still retrieve it later. Commands like `kit unpack`, `kit pull`, and `kit inspect` work the same way regardless of format.

### Kit CLI: Create, Run, Automate

The Kit CLI is the command-line tool that ties everything together. Use it to:
- Create ModelKits from your local project
- Unpack and inspect existing kits
- Run models locally or in test environments
- Automate packaging in CI/CD pipelines

## Who Uses KitOps?

KitOps helps bridge the gap between experimentation and production for AI workflows. Whether you’re running in the cloud, on-prem, or at the edge, KitOps makes it easier to collaborate across roles:

### For DevOps & Platform Engineers
- Use ModelKits in existing automation pipelines
- Store and manage AI artifacts in your current container registry
- Build golden paths for secure AI deployment - models, agents, and MCP servers
- Gate promotions with signing and attestation workflows

➡️ Integrate with [CI/CD](../integrations/cicd.md)

➡️ Add KitOps to [experiment trackers](../integrations/mlflow.md)

➡️ [Build a better golden path](../use-cases/) for AI projects.

### For Data Scientists
- Package datasets and models without infrastructure hassle
- Share your work with developers without “it works on my machine” issues
- Keep code and data versions aligned

[See how to use KitOps with Jupyter Notebooks](https://www.youtube.com/watch?v=OQPp7QEvk7Q).

➡️ Use the [PyKitOps Python SDK in your notebooks](../pykitops/index.md)

### For AI Engineers & Agent Developers
- Version prompts, skill files, and MCP server configs alongside your code
- Lock down reproducible agent state: exact prompt version + model version + skill version
- Share tested agent configurations across teams without copy-paste drift
- Pull only the artifacts you need (e.g., just the prompts) without downloading the full package

➡️ [Get started](../get-started.md)

### For Developers
- Use AI models and agent artifacts like any dependency
- Drop into apps using standard tools and APIs
- Let your team innovate without breaking your pipeline

➡️ [Get started](../get-started.md)

## A Standards-Based Approach

The goal of KitOps is to become the open, vendor-neutral standard that simplifies and secures the packaging and versioning of AI projects - from self-hosted models to agentic AI systems. In the same way that PDFs helped people share documents between tools, KitOps makes it easy for teams to use the tools they prefer but share the results safely and securely.

KitOps is governed by the CNCF and supported by contributors from across the AI and DevOps ecosystem.

### Join the Community!
- Get help and share ideas in the [KitOps Discord](https://discord.gg/Tapeh8agYy)
- [Open an issue](https://github.com/kitops-ml/kitops/issues) on GitHub
- [Contribute](https://github.com/kitops-ml/kitops/blob/main/CONTRIBUTING.md) to the project
- Help shape the [ModelPack standards specification](https://github.com/modelpack/model-spec) for AI project packaging
