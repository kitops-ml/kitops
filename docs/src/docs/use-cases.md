---
title: KitOps Use Cases - Secure Model Delivery & Versioning
description: Learn how organizations use KitOps to securely package, track, and deploy AI/ML models through CI/CD, with security scanning, tagging, and rollback.
keywords: kitops use cases, modelpack, ai model deployment workflow, mlops handoff, versioning ai models, secure ai pipelines, eu ai act compliance, ai model governance, mlflow handoff, reproducible ml deployment
---

# How Teams Use KitOps 🛠️

KitOps helps organizations package, share, and deploy AI/ML models securely and reproducibly — using the same tools they already use for containers.

Teams around the world are using KitOps for:
- Reproducible handoff from development to production
- Security and compliance (including EU AI Act, NIST AI, ISO 42001)
- Organizing all model versions in one standard system

➡️ See [compatible tools](../integrations/integrations.md)

## Level 1: Production Handoff

> Use Case: Reproducible, secure model handoff across teams using CI/CD

Most teams start by using KitOps to version a model when it’s ready for staging, UAT, or production. ModelKits serve as immutable, self-contained packages that simplify:
- CI/CD deployment of AI models
- Artifact signing and traceability
- App integration testing
- Secure, consistent model handoffs across teams

Organizations that are self-hosting models ❤️ KitOps because it:
- Prevents unknown models from entering production
- Enforces licensing and provenance checks (e.g. for Hugging Face imports)
- Keeps datasets, model, and code synced and trackable

### In Practice

CI/CD pipelines using GitHub Actions, Dagger, or other systems can:
1.	Pull models or data
2.	Run compliance / security tests
3.	Package project artifacts as a signed, versioned ModelKit
4.	Push the ModelKit to a private OCI registry

➡️ See [how CI/CD with KitOps works](../integrations/cicd.md)

## Level 2: Model Security

> Use Case: Scan and gate models during development or before release

Teams working in regulated industries or secure environments use KitOps to enforce security and integrity before a model is accepted into production.

### In Practice

1. Build a ModelKit for each experiment run in [MLFlow](../integrations/mlflow.md) / Weights & Biases
1. Sign the ModelKit
1. Scan the ModelKit using your preferred security scanning tools
1. Attach the security report as a signed attestation to the ModelKit
1. Only allow signed and attested ModelKits to move into forward environments
1. Track which model passed, which failed, and prevent risky surprises

Even when using other tools (MLFlow, Hugging Face, notebooks), KitOps provides a reliable security and auditing layer that protects environments from unsecure, or mistaken deployments.

## Level 3: Versioning Everything

> Use Case: Full model, code, and dataset lifecycle tracking

Mature teams — especially those under compliance scrutiny — extend KitOps to development. Every milestone (new dataset, tuning checkpoint, retraining event) is stored as a versioned ModelKit.

Benefits:
- One standard system (OCI) for every model version
- Tamper-evident and content-addressable storage
- Eliminates confusion over which assets belong together

### In Practice

1. Build a set of approved ModelKits by [importing from Hugging Face](../hf-import.md) or adding your own internal artifacts
1. Push ModelKits to your OCI registry
1. Eliminate duplicate work by starting projects from approved ModelKits
1. Version datasets as ModelKits and link them from project ModelKits
1. Perform signing, security testing and attestations as projects progress
1. Enforce policies using [OPA](https://www.openpolicyagent.org/) or similar technologies

➡️ [Get started](../get-started.md) with KitOps in your team.

---

**Have feedback or questions?**
Open an [issue on GitHub](https://github.com/kitops-ml/kitops/issues) or [join us on Discord](https://discord.gg/Tapeh8agYy).
