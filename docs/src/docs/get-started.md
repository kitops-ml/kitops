---
title: KitOps Tutorial - Open Source AI/ML Packaging CLI
description: Learn how to install and use the Kit CLI to package, version, and share AI/ML models using ModelKits. Follow our step-by-step guide for setup and deployment.
keywords: kitops, kit CLI, install kitops, package ai model, version ai model, share ml models, open source ai tools, mlops cli, modelkit example, modelpack, getting started with modelkit, ai model registry, deploy machine learning models
---

<script setup>
import vGaTrack from '@theme/directives/ga'
</script>

# KitOps Getting Started Tutorial

KitOps is the open-source CLI tool for packaging and sharing complete AI/ML projects using the [ModelKit](../modelkit/intro.md) format.

In this guide, you'll:

1. Unpack and inspect a sample ModelKit
1. Package your own model and data
1. Push it to a remote registry for collaboration or deployment

> Prefer to use KitOps with your favorite MLOps or other tool? [Check out our integrations](../integrations/integrations.md)

## Prerequisites

- [Install the Kit CLI](../cli/installation.md)
- Verify it by typing `kit version` in a new terminal
  - If you get an error check your PATH
- Create and navigate to a new folder (e.g., `KitStart`)

## Step 1: Log In to a Registry

You can use any OCI-compatible registry. We’ll use Jozu Hub for this example:

```sh
kit login jozu.ml
```

Use the email and password you signed up to your registry with.

Trouble? See the [kit login docs](../cli/cli-reference.md#kit-login).

## Step 2: Unpack a Sample ModelKit

:::tip
If you already have a model or dataset on your machine navigate to the directory where the files are and run `kit init .` in your terminal to build a Kitfile automatically.
:::

We’ll pull and unpack a fine-tuned Llama 3 model:

```sh
kit unpack jozu.ml/jozu-quickstarts/fine-tuning:latest
```

This unpacks all files to the current directory:

```sh
.
├── Kitfile
├── README.md
├── llama3-8b-8B-instruct-q4_0.gguf
├── lora-adapter.gguf
└── training-data.txt
```

## Step 3: Pack your ModelKit

Use the [kit pack command](../cli/cli-reference.md#kit-pack):

```sh
// Replace <your-name> with your Jozu.ml user
// [!code word:/your-username]
kit pack . -t jozu.ml/your-username/finetune:latest
```

This saves the ModelKit locally under the `latest` tag.

:::tip Packing as ModelPack format
To create a [ModelPack](https://github.com/modelpack/model-spec)-formatted artifact instead, add the `--use-model-pack` flag:
```sh
kit pack . --use-model-pack -t jozu.ml/your-username/finetune:latest
```
All Kit commands work with both ModelKit and ModelPack formats.
:::

Verify your packed model:

```sh
kit list
```

## Step 4: Push to a Remote Registry

Now push your ModelKit to share it:

```sh
// Replace <your-name> with your Jozu.ml user
// [!code word:/your-username]
kit push jozu.ml/your-username/finetune:latest
```

💡 If you see an error, check that your target repository exists and that you have permission to push.

## Next Steps
If you'd like to learn more about using Kit, try our [Next Steps with Kit](../next-steps/) document that covers:
* Creating a container or Kubernetes deployment from a ModelKit
* Signing your ModeKit
* Making your own Kitfile
* The power of `unpack`
* Tagging ModelKits
* Keeping your registry tidy
