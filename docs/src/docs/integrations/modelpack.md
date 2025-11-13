---
description: Use KitOps to create CNCF ModelPack-compliant packages.
keywords: modelpack, modelpack tools, oci ai packaging, ml model integration, ai model tools, machine learning pipelines, ml model serving, docker for ai models
---
# Using KitOps to Create ModelPacks

The Cloud Native Computing Foundation (CNCF) [ModelPack specification](https://github.com/modelpack/model-spec) is a vendor-neutral standard for packaging everything needed to share, deploy, and manage AI/ML projects for Kubernetes.

KitOps supports the ModelPack standard natively and transparently. To create a ModelPack artifact simply pack it with the modelpack flag:

```sh
# Create a ModelPack OCI Artifact
kit pack . --use-model-pack
```

Everything else you do with the ModelPack-compliant Artifact will be handled transparently by KitOps - simple!

**Questions or suggestions?** Drop an [issue in our GitHub repository](https://github.com/kitops-ml/kitops/issues) or join [our Discord server](https://discord.gg/Tapeh8agYy) to get support or share your feedback.