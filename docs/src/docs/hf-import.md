---
title: Import Hugging Face Models into ModelKits
description: Learn how to use the KitOps CLI to import Hugging Face models into ModelKits. Use KitOps to create a curated, private model registry behind your firewall.
keywords: kitops hugging face, import huggingface model, huggingface oci, huggingface kitfile, modelkit huggingface import, kit import cli, package hf model, hugging face model registry, share huggingface model
---

# Import Hugging Face Models into ModelKits

KitOps makes it easy to turn any Hugging Face repository into a versioned ModelKit.
Just use the `kit import` command to:
* Download the model and metadata
* Generate a Kitfile
* Build and store the ModelKit in your local registry

You can then push it to a private OCI registry, or use it in your development and deployment pipelines behind your firewall.

### Step 1: Get the Hugging Face Model URL

Visit [huggingface.co](https://huggingface.co) and copy the model URL you want to import. For example, `https://huggingface.co/HuggingFaceTB/SmolLM-135M-Instruct`

### Step 2: Run `kit import`

Use the CLI to import the Hugging Face repo and auto-generate a Kitfile:

```sh
kit import https://huggingface.co/HuggingFaceTB/SmolLM-135M-Instruct
```

KitOps will:
- Download the Hugging Face model and files
- Generate a Kitfile
- Launch your text editor for optional edits
- Pack and store the ModelKit locally

If the model requires authentication, use the `--token` flag to add your Hugging Face access token.

> 💡 Customize your Kitfile editor by setting the `EDITOR` environment variable.

### Step 3: Verify Your ModelKit

List your local ModelKits:

```sh
kit list
```

You should see a listing for `HuggingFaceTB/SmolLM-135M-Instruct:latest`.

## Next Steps
- [Push the ModelKit](./cli/cli-reference.md#kit-push) to a private registry
- Learn more about the [import command](./cli/cli-reference.md#kit-import)
- [Customize the Kitfile](./kitfile/format.md)
