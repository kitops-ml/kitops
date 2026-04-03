---
title: Why Use KitOps - Standards-Based AI/ML Packaging
description: Discover why KitOps is the leading open-source solution for packaging, versioning, and deploying AI/ML models. Built on OCI standards and compatible with your existing tools.
keywords: kitops, modelkit, modelpack, ml model versioning, ai packaging, ai ml deployment, reproducible machine learning, OCI ML tools, MLOps alternative, secure AI packaging
---

# Why Use KitOps?

AI projects are complex. KitOps makes them manageable.

Every AI project has many moving parts - models, datasets, prompts, agent skill files, MCP server configurations, code, and documentation. There's no consistent way to version, package, or deploy them together. KitOps solves this with an open-source, OCI-based standard for packaging and sharing complete AI projects.

## The Problem: No Standard Way to Package AI Projects

AI project artifacts are scattered across tools:
- Code in Jupyter notebooks or Git
- Model weights in S3 buckets or model hubs
- Datasets in object storage or DVC
- Prompts hardcoded in application code or lost in Slack threads
- Agent skill files copy-pasted between repos
- MCP server configs managed ad-hoc per environment
- Pipeline configs hidden in proprietary systems

Each has its own versioning (if any) and its own storage. Nothing ties the whole project together. Getting the wrong combination means delays, risk, and failure.

This makes it hard to answer critical questions:
- What code and data was used to train this model?
- Which prompt version is running in production?
- What changed between the working agent and the broken one?
- Who built and approved it, and when?

It also slows down collaboration between teams and introduces risks for security, compliance, and reproducibility.

## The KitOps Solution

KitOps brings DevOps-style packaging to AI workflows - for both self-hosted models and agentic AI systems.

Using KitOps, you can:
- Package models, prompts, skill files, MCP configs, datasets, and code into a single, versioned ModelKit
- Store and share using any OCI-compliant registry (Docker Hub, GitLab, Harbor)
- Unpack only what you need (e.g., just the prompts, just the model)
- Reproduce any previous state across development, testing, and production
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
1. Create a central source of truth for AI project artifacts - models, prompts, skills, configs, and datasets
1. Speed up deployment with standardized, versioned artifacts
1. Track changes and approvals for compliance and audits
1. Enforce consistency across environments

➡️ Try our simple [Get Started](../get-started.md)

## KitOps Helps You Answer:

- What data trained this model?
- Which prompt version is deployed in production?
- What exact combination of model + prompts + skills was running when the agent broke?
- Who signed off on it, and when?
- What changed between version 3 and version 4?

➡️ See how [security is built-in](../security.md)

## How KitOps Compares

KitOps doesn’t replace your favorite AI tools - it complements them with the secure, standardized packaging they are missing.

### KitOps and MLOps Platforms

Tools like MLflow, Weights & Biases, and others are great for tracking experiments. But they aren’t designed to package and version full AI projects for handoff across teams or deployment pipelines.

KitOps:
- Creates secure, immutable packages
- Stores them in standard OCI registries for security
- Integrates with any DevOps or MLOps tool
- Is free, open source, and governed by the independent CNCF organization

➡️ Integrate [KitOps with experiment trackers](../integrations/mlflow.md)

### KitOps and Agentic AI Frameworks

Agent frameworks like LangChain, CrewAI, as well as models like Claude can handle agentic orchestration and execution. But they don’t solve the packaging and versioning problem: which prompts, skills, and MCP server configs were running when the agent worked correctly?

KitOps:
- Versions prompts, skill files, and MCP configs as immutable artifacts
- Ties the exact prompt version to the exact model version and skill set
- Lets you roll back to a known-good agent configuration with one command
- Makes agent state reproducible across environments

### KitOps and Jupyter Notebooks

Notebooks are great for prototyping, but poor at versioning and reproducibility.

ModelKits:
- Package serialized models, code, and data
- Preserve state outside the notebook
- Let others reuse your work without opening a notebook

Tip: Add a Kitfile to your notebook project and run `kit pack` at the end of each run.

➡️ Add [KitOps to your notebook](https://www.youtube.com/watch?v=OQPp7QEvk7Q)

### KitOps and Containers

Containers are ideal for deployment, but they’re awkward for tracking datasets, configs, or experiment metadata.

ModelKits:
- Can include Dockerfiles and container artifacts
- Are easier for data teams to create than full containers
- Provide a clean handoff point between DS and DevOps teams
- Can be turned into containers when deployment is needed

➡️ Learn about [deployments with KitOps](../deploy.md)

### KitOps and Git

Git is built for code, not large binary files like models or datasets.

ModelKits:
- Handle binaries gracefully (no LFS nightmares)
- Keep everything in sync - model weights, prompts, skills, configs, and code
- Can still include code snapshots from Git when needed

### KitOps and Ad-Hoc Storage

AI project artifacts are often scattered across S3 buckets, local drives, config management tools, and hardcoded strings. That makes it hard to answer questions like:
- Which model used this dataset?
- Which prompt version was running when the agent failed?
- Is this MCP server config safe to deploy?

ModelKits:
- Version all project artifacts together - datasets, prompts, skills, configs, and code
- Reduce duplication and risk
- Support reproducibility and audit trails

### Recap: What Makes KitOps Unique

1. Full AI project packaging - models, prompts, skills, MCP configs, datasets, and code
1. Built on OCI standards
1. Works with your existing registries and CI/CD
1. Open source, vendor-neutral
1. Designed for team collaboration and governance

## Try It Today
- [Get Started](../get-started.md)
- [Explore CLI commands](../cli/cli-reference.md)
- [Learn about ModelKits](../modelkit/intro.md)
- [Contribute to the community](../../contributing.md)
