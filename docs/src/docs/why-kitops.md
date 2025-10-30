---
title: Why Use KitOps - Standards-Based AI/ML Packaging
description: Discover why KitOps is the leading open-source solution for packaging, versioning, and deploying AI/ML models. Built on OCI standards and compatible with your existing tools.
keywords: kitops, modelkit, modelpack, ml model versioning, ai packaging, ai ml deployment, reproducible machine learning, OCI ML tools, MLOps alternative, secure AI packaging
---

# Why Use KitOps?

AI/ML projects are complex. KitOps makes them manageable.

From datasets and models to configuration and deployment artifacts, every AI/ML project has many moving parts — and no consistent way to version, package, or deploy them together. KitOps solves this with an open-source, OCI-based standard for packaging and sharing complete AI/ML projects.

## The Problem: No Standard Way to Package AI/ML Projects

Most AI/ML projects are scattered across tools:
- Code in Jupyter notebooks or Git
- Datasets in S3 buckets or DVC
- Configs hidden in pipelines or feature stores
- Pipelines in proprietary systems

Each has its own package and its own version - but nothing ties the whole project together, and getting the wrong combination means delays, risk, and failure.

This makes it hard to answer critical questions like:
- What code and data was used to train this model?
- Who built and approved it?
- What version is running in production?
- What changed — and when?

It also slows down collaboration between teams and introduces risks for security, compliance, and reproducibility.

## The KitOps Solution

KitOps brings DevOps-style packaging to AI/ML workflows.

Using KitOps, you can:
- Package model, code, datasets, metadata — into a single, versioned ModelKit
- Store and share using any OCI-compliant registry (Docker Hub, GitLab, Harbor)
- Unpack only what you need (e.g., just code, just dataset)
- Run reproducible pipelines across development, testing, and production
- Reduce risk with immutable, traceable AI project artifacts

KitOps is designed for speed, security, and interoperability.

➡️ See [compatible tools](../integrations/integrations.md)

## How It Works
- Write a [Kitfile](../kitfile/kf-overview.md) to describe your project contents
- Run `kit pack` to create a ModelKit
- Use `kit push` to upload to any OCI registry
- Share and deploy across teams using `kit pull` or `kit unpack`

Built with standards from the Open Container Initiative (OCI), KitOps is compatible with the registries, pipelines, and serving infrastructure your organization already uses.

➡️ Install the [CLI](../cli/installation.md)

## Trusted by Teams

KitOps has been downloaded by hundreds of thousands of users, and is in production use in high security organizations, governments, enterprises, and clouds.

Teams use it to:
1. Create a central source of truth for models and datasets
1. Speed up model deployment with standardized artifacts
1. Track changes and approvals for compliance and audits
1. Enforce consistency across environments

➡️ Try our simple [Get Started](../get-started.md)

## KitOps Helps You Answer:

- What data trained this model?
- Which version is deployed where?
- Who signed off on it — and when?
- What changed between model version 3 and 4?

➡️ See how [security is built-in](../security.md)

## How KitOps Compares

KitOps doesn’t replace your favorite AI/ML tools — it complements them with the secure, standardized packaging they are missing.

### KitOps and MLOps Platforms

Tools like MLflow, Weights & Biases, and others are great for tracking experiments. But they aren’t designed to package and version full AI/ML projects for handoff across teams or deployment pipelines.

KitOps:
- Creates secure, immutable packages
- Stores them in standard OCI registries for security
- Integrates with any DevOps or MLOps tool
- Is free, open source, and governed by the independent CNCF organization

➡️ Integrate [KitOps with experiment trackers](../integrations/mlflow.md)

### KitOps and Jupyter Notebooks

Notebooks are great for prototyping, but poor at versioning and reproducibility.

ModelKits:
- Package serialized models, code, and data
- Preserve state outside the notebook
- Let others reuse your work without opening a notebook

👉 Tip: Add a Kitfile to your notebook project and run `kit pack` at the end of each run.

➡️ Add [KitOps to your notebook](https://www.youtube.com/watch?v=OQPp7QEvk7Q)

### KitOps and Containers

Containers are ideal for deployment — but they’re awkward for tracking datasets, configs, or experiment metadata.

ModelKits:
- Can include Dockerfiles and container artifacts
- Are easier for data teams to create than full containers
- Provide a clean handoff point between DS and DevOps teams
- Can be turned into containers when deployment is needed

➡️ Learn about [deployments with KitOps](../deploy.md)

### KitOps and Git

Git is built for code — not large binary files like models or datasets.

ModelKits:
- Handle binaries gracefully (no LFS nightmares)
- Keep everything in sync
- Can still include code snapshots from Git when needed

### KitOps and Ad-Hoc Dataset Storage

Datasets are often scattered across S3 buckets, local drives, databases, or BI tools. That makes it hard to answer questions like:
- Which model used this dataset?
- Did this dataset change — and when?
- Is this data safe to use?

ModelKits:
- Version datasets alongside model code and configs
- Reduce duplication and risk
- Support reproducibility and audit trails

### Recap: What Makes KitOps Unique

1. Full AI/ML project packaging
1. Built on OCI standards
1. Works with your existing registries and CI/CD
1. Open source, vendor-neutral
1. Designed for team collaboration and governance

## Try It Today
- [Get Started](../get-started.md)
- [Explore CLI commands](../cli/cli-reference.md)
- [Learn about ModelKits](../modelkit/intro.md)
- [Contribute to the community](../../contributing.md)
