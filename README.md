<img width="1270" alt="KitOps" src="https://github.com/kitops-ml/kitops/assets/10517533/41295471-fe49-4011-adf6-a215f29890c2" id="top">

## KitOps: Standards-based packaging & versioning for AI/ML projects

[![LICENSE](https://img.shields.io/badge/License-Apache%202.0-yellow.svg)](./LICENSE)
[![Discord](https://img.shields.io/discord/1098133460310294528?logo=Discord)](https://discord.gg/Tapeh8agYy)
[![Twitter](https://img.shields.io/twitter/url/http/shields.io.svg?style=social&label=Twitter)](https://twitter.com/kit_ops)

## Table of Contents

- [What is KitOps?](#what-is-kitops)
- [Try KitOps](#try-kitops-in-under-15-minutes)
- [How Teams Use KitOps](#how-teams-use-kitops)
- [KitOps Architecture](#kitops-architecture)
- [Security and Compliance](#security-and-compliance)
- [KitOps by Role](#kitops-by-role)
- [Integrations](#integrations)
- [Community and Support](#join-kitops-community)

## What is KitOps?

KitOps is a CNCF open source tool for packaging, versioning, and securely sharing AI/ML projects.

Built on the same [OCI (Open Container Initiative)](https://opencontainers.org/) technology that underlies containers, KitOps packages everything your model needs for development or production into a versioned and layered artifact stored in your existing container registry. It integrates with all your AI/ML, CI/CD, and DevOps tools.

As part of the Kubernetes AI/ML technology stack, KitOps is the preferred solution for packaging, versioning, and managing AI assets in security-conscious enterprises, governments, and cloud operators who need to self-host AI models and agents.

### KitOps and the CNCF

KitOps is governed by the [CNCF](https://www.cncf.io/) (the same organization that manages Kubernetes, OpenTelemetry, and Prometheus). [This video provides an outline of KitOps in the CNCF](https://youtu.be/iK9mnU0prRU?feature=shared).

KitOps is also the enterprise implementation of the [CNCF ModelPack specification](https://github.com/modelpack/model-spec) for a vendor-neutral AI/ML interchange format. The Kit CLI supports both ModelKit and ModelPack formats transparently. Contributing companies to ModelPack include Red Hat, PayPal, ANT Group, and ByteDance.

[![Official Website](<https://img.shields.io/badge/-Visit%20the%20Official%20Website%20%E2%86%92-rgb(255,175,82)?style=for-the-badge>)](https://kitops.org?utm_source=github&utm_medium=kitops-readme)

[![Use Cases](<https://img.shields.io/badge/-KitOps%20Quick%20Start%20%E2%86%92-rgb(122,140,225)?style=for-the-badge>)](https://kitops.org/docs/get-started/?utm_source=github&utm_medium=kitops-readme)

## Try KitOps in Under 15 Minutes

1. **Install the CLI**: [for MacOS, Windows, and Linux](https://kitops.org/docs/cli/installation/).
2. **Pack your first ModelKit**: Either:
   * **Import from HuggingFace**: Pull models directly from HuggingFace into a ModelKit with [HuggingFace Import](https://kitops.org/docs/hf-import/).
   * Navigate to your project directory and run `kit init .` to auto-generate a Kitfile, then follow the [Getting Started](https://kitops.org/docs/get-started/) guide to pack, push, and pull.
3. **Push it to your registry**: Use `kit push` to start using your existing enterprise registry as a secure and curated registry for AI agents, models, and MCP servers.
4. **Explore pre-built ModelKits**: [Try quick starts](https://jozu.ml/organization/jozu-quickstarts) for LLMs, computer vision models, and more.

For those who prefer to build from source, follow [these steps](https://kitops.org/docs/cli/installation/#build-sources) to get the latest version from our repository.

## How Teams Use KitOps

### Level 1: Production Handoff

Most teams start by using KitOps to version a model or agent when it's ready for staging or production. ModelKits serve as immutable, self-contained packages that simplify CI/CD deployment, artifact signing, AI SBOM creation, and deployment / rollback. This prevents unknown AI workloads from entering production and keeps datasets, model weights, and config synced and trackable.

Learn more: [CI/CD integration](https://kitops.org/docs/integrations/cicd/)

### Level 2: Model Security

Teams in regulated industries use KitOps to scan and gate models before they reach production. Build a ModelKit, sign it with Cosign, run security scans, attach reports as signed attestations, and only allow attested ModelKits to move forward. KitOps provides a security and auditing layer on top of whatever tools you already use.

Learn more: [Securing ModelKits](https://kitops.org/docs/security/)

### Level 3: Full Lifecycle Versioning

Mature teams extend KitOps to development. Every milestone (new dataset, tuning checkpoint, retraining event) is stored as a versioned ModelKit. One standard system (OCI) for every model version, with tamper-evident and content-addressable storage.

Learn more: [How KitOps is Used](https://kitops.org/docs/use-cases/)

## KitOps Architecture

### ModelKit

KitOps packages your project into a [ModelKit](https://kitops.org/docs/modelkit/intro/) - a self-contained, immutable bundle that includes everything required to reproduce, test, or deploy your AI/ML model.

ModelKits can include agents, model weights, MCP servers, datasets, prompts, experiment run results and hyperparameters, metadata, environment configurations, code, and more.

ModelKits are:
* **Tamper-proof** - Every component protected by SHA-256 digests, ensuring consistency and traceability
* **Signable** - Full [Cosign](https://github.com/sigstore/cosign) compatibility for cryptographic verification
* **Compatible** - Natively stored and retrieved in all major OCI container registries
* **Selectively unpacked** - Pull only the layers you need (just the model, just the dataset, etc.)

KitOps can also create [ModelPack](https://github.com/modelpack/model-spec)-compliant packages using the CNCF model-spec format. Both formats are vendor-neutral standards, and Kit commands (`pull`, `push`, `unpack`, `inspect`, `list`) work transparently with both.

> ModelKits elevate AI artifacts to first-class, governed assets, just like application code.

### Kitfile

A [Kitfile](https://kitops.org/docs/kitfile/kf-overview/) defines where each artifact lives in your ModelKit. You can generate one automatically with `kit init`.

### Kit CLI

The [Kit CLI](https://kitops.org/docs/cli/cli-reference/) lets you create, manage, run, and deploy ModelKits. Key commands include:

* `kit pack` - Package your project into a ModelKit (add `--use-model-pack` for ModelPack format)
* `kit unpack` - Extract all or specific layers from a ModelKit
* `kit push` / `kit pull` - Share ModelKits through any OCI registry
* `kit init` - Auto-generate a Kitfile from an existing project directory
* `kit diff` - Compare differences between two ModelKits
* `kit list` - List available ModelKits and ModelPacks
* `kit inspect` - View the contents of a ModelKit without unpacking

### PyKitOps Python SDK

The [PyKitOps library](https://kitops.org/docs/pykitops/) lets data scientists work with ModelKits in Python. Use it to pack, push, pull, and inspect ModelKits without leaving your favorite tool's workflow.

### Watch KitOps in Action

[![KitOps Video](https://img.youtube.com/vi/j2qjHf2HzSQ/hqdefault.jpg)](https://www.youtube.com/watch?v=j2qjHf2HzSQ)

This video shows how KitOps streamlines collaboration between data scientists, developers, and SREs using ModelKits.

## Security and Compliance

KitOps provides artifact and project metadata for organizations that need to establish and maintain chain-of-custody and provenance for their AI/ML assets:

* **Immutable digests** - Every ModelKit component is SHA-256 hashed. Any modification to any file is detected via OCI digest verification when the artifact is pulled or fetched, and the tampered artifact is rejected.
* **Cryptographic signatures** - Sign ModelKits with [Cosign](https://github.com/sigstore/cosign) (key-based or keyless via OIDC). Unsigned or tampered ModelKits can be blocked in CI/CD.
* **AI Bill of Materials** - ModelKits provide a structured inventory of all components (model weights, datasets, code, configs) with version tracking, serving as the foundation for AI SBOMs.
* **Transparency logging** - Combine with [Rekor](https://github.com/sigstore/rekor) for append-only signature records.
* **Audit-ready lineage** - Full version history from experiment through staging to production, stored in your OCI registry.

These properties make ModelKits suitable for compliance frameworks that require artifact integrity, provenance verification, and audit trails, including the EU AI Act, NIST AI RMF, ISO 42001, and similar regulatory requirements.

Learn more: [Securing Your Model Supply Chain](https://kitops.org/docs/security/)

KitOps is also used by **[Jozu Hub](https://jozu.com/)**, that adds centralized policy administration, five-layer security scanning, signed attestations, and tamper-evident audit logs. Jozu Hub installs behind your firewall and works with your existing OCI registry in private cloud, datacenter, or air-gapped environments.

## KitOps by Role

### DevOps and Platform Engineers

* Use ModelKits in existing CI/CD pipelines with [GitHub Actions, Dagger, and other systems](https://kitops.org/docs/integrations/cicd/)
* Store and manage models in your current container registry
* Deploy to Kubernetes using the [init container](https://kitops.org/docs/integrations/k8s-init-container/) or [KServe](https://kitops.org/docs/integrations/kserve/)
* Build golden paths for secure AI/ML deployment

### Data Scientists

* Package datasets and models without infrastructure hassle using `kit pack` or the [PyKitOps SDK](https://kitops.org/docs/pykitops/)
* [Import models from HuggingFace](https://kitops.org/docs/hf-import/) into governed ModelKits
* Track experiments with [MLFlow integration](https://kitops.org/docs/integrations/mlflow/)

### Developers

* Use AI/ML models like any dependency with standard tools and APIs
* Pull only the layers you need (model, dataset, code) without downloading the full package
* Integrate with [Kubeflow Pipelines](https://kitops.org/docs/integrations/kubeflow/) and other ML tooling

## Integrations

KitOps works with the tools you already use:

* [MLFlow](https://kitops.org/docs/integrations/mlflow/) - Package experiment runs as versioned ModelKits
* [CI/CD](https://kitops.org/docs/integrations/cicd/) - GitHub Actions, Dagger, and other pipeline tools
* [Kubernetes initContainer](https://kitops.org/docs/integrations/k8s-init-container/) - Unpack ModelKits as init containers
* [KServe](https://kitops.org/docs/integrations/kserve/) - Serve models directly from ModelKits
* [Kubeflow Pipelines](https://kitops.org/docs/integrations/kubeflow/) - Use ModelKits in ML pipeline steps
* [ModelPack](https://kitops.org/docs/integrations/modelpack/) - CNCF vendor-neutral packaging format

See the full [integration list](https://kitops.org/docs/integrations/integrations/).

## Join KitOps Community

For support, release updates, and general KitOps discussion, please join the [KitOps Discord](https://discord.gg/Tapeh8agYy). Follow [KitOps on X](https://twitter.com/Kit_Ops) for daily updates.

If you need help there are several ways to reach our community and [Maintainers](./MAINTAINERS.md) outlined in our [support doc](./SUPPORT.md)

### Joining the KitOps Contributors

We love our KitOps community and contributors. To learn more about the many ways you can contribute (you don't need to be a coder) and how to get started see our [Contributor's Guide](./CONTRIBUTING.md). Please read our [Governance](./GOVERNANCE.md) and our [Code of Conduct](./CODE-OF-CONDUCT.md) before contributing.

### Reporting Issues and Suggesting Features

Your insights help KitOps evolve as an open standard for AI/ML. We deeply value the issues and feature requests we get from users in our community. To contribute your thoughts, navigate to the **Issues** tab and click the **New Issue** button.

### KitOps Community Calls (bi-weekly)

**Wednesdays @ 13:30 - 14:00 (America/Toronto)**
- [Google Meet](https://meet.google.com/zfq-uprp-csd)
- +1 647-736-3184 (PIN: 144 931 404#)
- [More numbers](https://tel.meet/zfq-uprp-csd?pin=1283456375953)

### A Community Built on Respect

At KitOps, inclusivity, empathy, and responsibility are at our core. Please read our [Code of Conduct](./CODE-OF-CONDUCT.md) to understand the values guiding our community.
---

<div align="center" style="align-items: center;">
        <a href="#top">
            <img src="https://img.shields.io/badge/Back_to_Top-black?style=for-the-badge&logo=github&logoColor=white" alt="Back to Top">
        </a>
</div>
