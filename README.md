<img width="1270" alt="KitOps" src="https://github.com/kitops-ml/kitops/assets/10517533/41295471-fe49-4011-adf6-a215f29890c2" id="top">

## KitOps: Standards-based packaging & versioning for AI/ML projects

[![LICENSE](https://img.shields.io/badge/License-Apache%202.0-yellow.svg)](./LICENSE)
[![Discord](https://img.shields.io/discord/1098133460310294528?logo=Discord)](https://discord.gg/Tapeh8agYy)
[![Twitter](https://img.shields.io/twitter/url/http/shields.io.svg?style=social&label=Twitter)](https://twitter.com/kit_ops)

## 📚 Table of Contents

- [What is KitOps?](#what-is-kitops)
- [KitOps Architecture](#kitops-architecture)
- [Try KitOps](#try-kitops-in-under-15-minutes)
- [Benefits](#key-benefits)
- [Community & Support](#join-kitops-community)

## What is KitOps?

KitOps is a CNCF open standards project for packaging, versioning, and securely sharing AI/ML projects. Built on the OCI ([Open Container Initiative](https://opencontainers.org/)) standard, it integrates seamlessly with your existing AI/ML, software development, and DevOps tools.

It’s the preferred solution for packaging, versioning, and managing assets in security-conscious enterprises, governments, and cloud operators who need to self-host AI models and agents.

### KitOps and the CNCF

KitOps is a CNCF project, and is governed by the same organization and policies that manage Kubernetes, OpenTelemetry, and Prometheus. [This video provides an outline of KitOps in the CNCF](https://youtu.be/iK9mnU0prRU?feature=shared).

KitOps is also the reference implementation of the [CNCF's ModelPack specification](https://github.com/modelpack/model-spec) for a vendor-neutral AI/ML interchange format.

[![Official Website](<https://img.shields.io/badge/-Visit%20the%20Official%20Website%20%E2%86%92-rgb(255,175,82)?style=for-the-badge>)](https://kitops.org?utm_source=github&utm_medium=kitops-readme)

[![Use Cases](<https://img.shields.io/badge/-KitOps%20Quick%20Start%20%E2%86%92-rgb(122,140,225)?style=for-the-badge>)](https://kitops.org/docs/get-started/?utm_source=github&utm_medium=kitops-readme)

## KitOps Architecture

### ModelKit

KitOps packages your project into a [ModelKit](https://kitops.org/docs/modelkit/intro/) — a self-contained, immutable bundle that includes everything required to reproduce, test, or deploy your AI/ML model.

ModelKits can include code, model weights, datasets, prompts, experiment run results and hyperparameters, metadata, environment configurations, and more.

ModelKits are:
* Tamper-proof – Ensuring consistency and traceability
* Signable – Enabling trust and verification
* Compatible – Natively stored and retrieved in all major container registries

> *ModelKits elevate AI artifacts to first-class, governed assets — just like application code.*

### Kitfile

A [Kitfile](https://kitops.org/docs/kitfile/kf-overview/) defines your ModelKit. Written in YAML, it maps where each artifact lives and how it fits into the project.

### Kit CLI

The [Kit CLI](https://kitops.org/docs/cli/cli-reference/) not only enables you to create, manage, run, and deploy ModelKits -- it lets you pull only the pieces you need.

### 🎥 Watch KitOps in Action

[![KitOps Video](https://img.youtube.com/vi/j2qjHf2HzSQ/hqdefault.jpg)](https://www.youtube.com/watch?v=j2qjHf2HzSQ)

This video shows how KitOps streamlines collaboration between data scientists, developers, and SREs using ModelKits.

## 🚀 Try KitOps in under 15 Minutes

1. **Install the CLI**: [for MacOS, Windows, and Linux](https://kitops.org/docs/cli/installation/).
2. **Pack your first ModelKit**: Learn how to pack, push, and pull using our [Getting Started](https://kitops.org/docs/get-started/) guide.
3. **Explore a Quick Start**: [Try pre-built ModelKits](https://jozu.ml/organization/jozu-quickstarts) for LLMs, CVs, and more.

For those who prefer to build from the source, follow [these steps](https://kitops.org/docs/cli/installation/#🛠️-install-from-source) to get the latest version from our repository.

## Key Benefits

KitOps was built to bring discipline to productizing AI/ML projects, with:
* 📦 Unified packaging and versioning of AI/ML assets
* 🔐 Secure, signed distribution
* 🛠️ Toolchain compatibility via OCI
* ⚙️ Production-ready for enterprise ML workflows
* 🚢 Create runnable containers for Kubernetes or docker
* 📈 Audit-ready lineage tracking

To get the most out of KitOps' ModelKits, use them with the **[Jozu Hub](https://jozu.com/)**. Jozu Hub can be installed behind your firewall and use your existing OCI registry in a private cloud, datacenter, or even in an air-gapped environment.

### Simplify Team Collaboration

ModelKits streamline handoffs between:
* Data scientists preparing and training models
* Application developers integrating models into services
* SREs deploying and maintaining models in production

This ensures reliable, repeatable workflows for both development and operations.

### Use KitOps to Speed Up and De-risk AI/ML Projects

KitOps supports packaging for a wide variety of models:
* Large language models
* Computer vision models
* Multi-modal models
* Predictive models
* Audio models
* etc...

> 🇪🇺 EU AI Act Compliance 🔒
>
> For our friends in the EU - ModelKits are the perfect way to create a library of model versions for EU AI Act compliance because they're tamper-proof, signable, and auditable.

## Join KitOps Community

For support, release updates, and general KitOps discussion, please join the [KitOps Discord](https://discord.gg/Tapeh8agYy). Follow [KitOps on X](https://twitter.com/Kit_Ops) for daily updates.

If you need help there are several ways to reach our community and [Maintainers](./MAINTAINERS.md) outlined in our [support doc](./SUPPORT.md)

### Joining the KitOps Contributors

We ❤️ our KitOps community and contributors. To learn more about the many ways you can contribute (you don't need to be a coder) and how to get started see our [Contributor's Guide](./CONTRIBUTING.md). Please read our [Governance](./GOVERNANCE.md) and our [Code of Conduct](./CODE-OF-CONDUCT.md) before contributing.


### Reporting Issues and Suggesting Features

Your insights help KitOps evolve as an open standard for AI/ML. We *deeply value* the issues and feature requests we get from users in our community :sparkling_heart:. To contribute your thoughts, navigate to the **Issues** tab and click the **New Issue** button.

### KitOps Community Calls (bi-weekly)

**📅 Wednesdays @ 13:30 – 14:00 (America/Toronto)**  
- 🔗 [Google Meet](https://meet.google.com/zfq-uprp-csd)  
- ☎️ +1 647-736-3184 (PIN: 144 931 404#)  
- 🌐 [More numbers](https://tel.meet/zfq-uprp-csd?pin=1283456375953)

### A Community Built on Respect

At KitOps, inclusivity, empathy, and responsibility are at our core. Please read our [Code of Conduct](./CODE-OF-CONDUCT.md) to understand the values guiding our community.

---

<div align="center" style="align-items: center;">
        <a href="#top">
            <img src="https://img.shields.io/badge/Back_to_Top-black?style=for-the-badge&logo=github&logoColor=white" alt="Back to Top">
        </a>
</div>
