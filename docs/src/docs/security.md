---
title: ModelKit Security - Integrity, Signing, and Verification
description: Learn how KitOps ModelKits ensure AI/ML model integrity with built-in SHA-256 verification, optional Cosign signing, and integration with transparency logs. Secure your AI supply chain from development to deployment.
keywords: modelkit security, modelpack, ai model integrity, verify model artifact, cosign signing ml, oci security mlops, secure ai packaging, reproducible ml deployment, transparency logs, digests sha256, rekor
---

# Securing Your Model Supply Chain with KitOps

Model integrity matters. KitOps helps you protect your AI/ML project assets from corruption, tampering, or unauthorized changes — without requiring custom security code or complex setup.

Whether you're building models for production, working in a regulated industry, or want to enforce trust boundaries between teams, KitOps provides built-in verification and open standards for secure packaging.

## Why Use OCI Artifacts for AI Security?

OCI (Open Container Initiative) artifacts are content-addressed and immutable, which means:
- Once a ModelKit is packaged and pushed, its contents can’t silently change — every file is hashed and referenced by digest.
- The ModelKit is tamper-evident: if any model file, dataset, or config is altered, the digest no longer matches and unpacking will fail.
- OCI registries already support access control, logging, and redundancy, which gives you security without building new infrastructure.

This gives KitOps a strong security foundation out-of-the-box — just like container images, but for full AI/ML projects.

And KitOps fits seamlessly into a broader enterprise security processes and tools...

## Built-in Integrity Checks

Every time you run `kit unpack` or `kit pull` KitOps automatically:
1.	Reads the OCI manifest for the ModelKit
2.	Computes the SHA-256 digest of each layer (datasets, weights, code, Kitfile, etc.)
3.	Compares each digest to the expected value from the manifest

If any artifact has been modified — even by one byte — unpacking fails with a clear digest mismatch error.

> Zero-lift for you: You don’t need to write or manage any checksum code. This is handled automatically by the KitOps CLI and PyKitOps SDK.

## Signing with Cosign

For additional cryptographic assurance — including tamper-proof authorship and verifiable approval flows — KitOps is fully compatible with [Cosign](https://github.com/sigstore/cosign).

```sh
# Sign the ModelKit
cosign sign --key cosign.key jozu.ml/brad/signed-kit:2.0.0

# Verify signature
cosign verify --key cosign.pub jozu.ml/brad/signed-kit:2.0.0
```

Any mismatch = `kit` stops the unpack.

Any missing signature = your CI/CD pipeline can block it.

### Bonus: Go keyless

Cosign supports OIDC-based keyless signing, eliminating key file management. Signatures can be optionally recorded in a transparency log (e.g., Rekor) for audit trails and compliance.

## Build it Into Your Pipeline

Combine KitOps + Cosign in [any pipeline](../integrations/cicd.md) with KitOps - that's the easiest and safest way to keep things signed and secure.

## Audit Trails & Chain of Custody

KitOps doesn’t maintain audit logs itself because it can leverage the specialty products available in the open source ecosystem. Combine KitOps with these components for full traceability:

| Feature | Supported Tooling |
| --- | --- |
| Immutable digests | KitOps (built-in) |
| Cryptographic signatures | Cosign |
| Transparency log | Rekor |
| Push metadata | Registry (e.g. Jozu Hub) |

For organizations needing detailed, queryable audit history, [Jozu Hub](https://jozu.com/) (the commercial platform backing KitOps) provides:
- Kubernetes-hosted installation behind the firewall
- Security scanning of all models
- Signed attestations for security reports attached to ModelKits
- Integration with Open Policy Agent
- Auto-generated secure containers and Kubernetes deployments
- UI and API access to historical metadata

## Learn More or Get Help
- [KitOps CLI](../cli/cli-reference.md) Reference
- [Deploy Secure ModelKits](../deploy.md)
- Join the [KitOps Community](https://discord.gg/Tapeh8agYy)
