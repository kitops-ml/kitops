---
description: Import models from Hugging Face into KitOps effortlessly. Follow our guide to integrate your existing AI models seamlessly.
---
# Importing a ModelKit from Hugging Face

You can automatically generate a ModelKit from a Hugging Face repository using the `kit import` command. This speeds and simplifies the job of getting started with ModelKits.

::: tip
To customize the editor used for editing the Kitfile during import, set the `EDITOR` environment variable
:::

You can read more about the `import` command in our [CLI reference](../cli/cli-reference/#kit-import).

## Importing a Hugging Face repository

1. **Get the HF URL**: On the Hugging Face site, copy the URL from the repository you want to create a ModelKit from (e.g., https://huggingface.co/HuggingFaceTB/SmolLM-135M-Instruct). You can also customize the name or add a tag name if desired.

2. **Kit import**: In a terminal window running Kit version 1.0.0 at least, type `kit import https://huggingface.co/HuggingFaceTB/SmolLM-135M-Instruct`. This will download and build a [Kitfile](../kitfile/kf-overview/) based on the Hugging Face model and give you an opportunity to edit it before the ModelKit is packed.

::: tip
If you have a [Huggingface Access Token](https://huggingface.co/docs/hub/security-tokens) you can specify it using the `--token` flag for `kit import`.
:::

3. **Auto-generate the ModelKit**: Once the Kitfile is accepted a ModelKit will be built and saved to your local registry using the name you selected. If you didn't specify a tag name the ModelKit will be tagged `latest`.

4. **Admire your new ModelKit**: Typing `kit list` will show you the list of ModelKits on your system and you'll see the newly imported tag (e.g., `HuggingFaceTB/SmolLM-135M-Instruct:latest`).

That's it! So easy! Now go out there and start sharing your favourite Hugging Face models through your OCI registry!
