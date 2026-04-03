---
title: ModelKit Security - Integrity, Signing, and Verification
description: Learn how KitOps ModelKits ensure AI/ML model integrity with built-in SHA-256 verification, optional Cosign signing, and integration with transparency logs. Secure your AI supply chain from development to deployment.
keywords: modelkit security, modelpack, ai model integrity, verify model artifact, cosign signing ml, oci security mlops, secure ai packaging, reproducible ml deployment, transparency logs, digests sha256, rekor
---

# Securing Your AI Supply Chain with KitOps

AI artifact integrity matters - and it’s not just about models anymore. Prompts, agent skill files, and MCP server configurations are all attack surfaces. A tampered prompt can change agent behavior without touching a single line of code. A compromised MCP server config can route tool calls to an attacker’s endpoint. KitOps helps you protect all of these assets from corruption, tampering, or unauthorized changes.

Whether you’re deploying self-hosted models, managing agentic AI systems, or distributing MCP server configurations, KitOps provides built-in verification and open standards for secure packaging.

## Why Use OCI Artifacts for AI Security?

OCI (Open Container Initiative) artifacts are content-addressed and immutable, which means:
- Once a ModelKit is packaged and pushed, its contents can’t silently change. Every file - model weights, prompts, skill files, MCP configs - is hashed and referenced by digest.
- The ModelKit is tamper-evident: if any artifact is altered, the digest no longer matches and unpacking will fail.
- OCI registries already support access control, logging, and redundancy, giving you security without building new infrastructure.

This gives KitOps a strong security foundation out of the box, for models, agent configurations, and everything in between.

## Built-in Integrity Checks

Every time you run `kit unpack` or `kit pull` KitOps automatically:
1. Reads the OCI manifest for the ModelKit
2. Computes the SHA-256 digest of each layer (model weights, datasets, prompts, skill files, code, MCP configs, Kitfile, etc.)
3. Compares each digest to the expected value from the manifest

If any artifact has been modified - even by one byte - unpacking fails with a clear digest mismatch error. This applies equally to a model weight file and a system prompt. You get the same tamper-evident guarantee for every artifact in the ModelKit.

> You don’t need to write or manage any checksum code. This is handled automatically by the KitOps CLI and PyKitOps SDK.

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

## Why This Matters for Agentic AI

Agent security isn’t just about the model. When your agent’s behavior is defined by a combination of model + prompts + skills + MCP server configs, you need integrity guarantees across all of them. A few scenarios:

- A prompt change that removes safety guardrails should be caught before deployment, not after an incident.
- An MCP server config that redirects tool calls to an untrusted endpoint should fail verification.
- A skill file modified outside your approved workflow should be detectable.

By packaging all of these artifacts into a signed ModelKit, you get one verification step that covers the entire agent configuration. If anything changed, the signature breaks.

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
- Security scanning of all AI artifacts - models, agents, and MCP servers
- Signed attestations for security reports attached to ModelKits
- Integration with Open Policy Agent
- Auto-generated secure containers and Kubernetes deployments
- UI and API access to historical metadata
- MCP Registry API for centrally curated, security-scanned MCP servers

## Learn More or Get Help
- [KitOps CLI](../cli/cli-reference.md) Reference
- [Deploy Secure ModelKits](../deploy.md)
- Join the [KitOps Community](https://discord.gg/Tapeh8agYy)
