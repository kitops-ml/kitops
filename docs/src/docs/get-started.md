---
description: Begin your journey with KitOps. Learn how to download, install, and utilize the Kit CLI to create and manage ModelKits for your AI/ML projects. Follow our step-by-step guide to streamline your development process.
---

<script setup>
import vGaTrack from '@theme/directives/ga'
</script>

# KitOps Getting Started

In this guide, we'll use ModelKits and the kit CLI to easily:
* Package up a model, notebook, and datasets into a single ModelKit you can share through your existing tools
* Push that versioned ModelKit package to a registry
* Grab only the assets you need from the ModelKit for testing, integration, local running, or deployment
* Package the ModelKit as a container or Kubernetes deployment

:::info
If you're more interested in using KitOps with [MLFlow](../integrations/mlflow.md) or [CI/CD pipelines](../integrations/cicd.md) check out our integrations section.
:::

## Before we start...

1. Make sure you've got the [Kit CLI setup](../cli/installation/).
2. Create and navigate to a new folder on your filesystem - we suggest calling it `KitStart` but any name works.

## Learning to use the CLI

### 1/ Check your CLI Version

Check that the Kit CLI is properly installed by using the [version command](../cli/cli-reference/#kit-version).

```sh
kit version
```

You'll see information about the version of Kit you're running. If you get an error check to make sure you have [Kit installed and in your path](../cli/installation/).

### 2/ Login to Your Registry

You can use the [login command](../cli/cli-reference/#kit-login) to authenticate with any OCI v1.1-compatible container registry - local or remote (you can see our [list of compliant registries](../modelkit/compatibility/)). In this guide, we'll use the [Jozu Hub](https://jozu.ml/) because it's free to sign-up and provides more detail on what's inside each ModelKit like whether it's signed or has provenance. You can substitute your own repository if preferred.

```sh
kit login jozu.ml
```

After entering your username and password, you'll see `Log in successful`. If you get an error it may be that you need an HTTP vs HTTPS (default) connection. Try the login command again but with `--plain-http`.

### 3/ Get a Sample ModelKit

You can get started using one of our sample ModelKits (covered below), or you can [import a ModelKit from Hugging Face](../hf-import/).

Let's use the [unpack command](../cli/cli-reference/#kit-unpack) to pull a [sample ModelKit from Jozu Hub](https://jozu.ml/organization/jozu-quickstarts) to our machine that we can play with. In this case, we'll unpack the whole thing, but one of the great things about Kit is that you can also selectively unpack only the artifacts you need: just the model, the model and dataset, the code, the configuration...whatever you want. Check out the `unpack` [command reference](../cli/cli-reference/#kit-unpack) for details.

If you have a model already on your machine you can use that instead.

You can grab <a href="https://jozu.ml/browse"
  v-ga-track="{
    category: 'link',
    label: 'grab any of the ModelKits',
    location: 'docs/get-started'
  }">any of the ModelKits</a> from the Jozu Hub, but we've chosen a fine-tuned model based on Llama3.

The unpack command will unpack the ModelKit contents to the current directory by default. If you want it unpacked to a specific directory use the `-d /path/to/unpacked`.

```sh
kit unpack jozu.ml/jozu-quickstarts/fine-tuning:latest
```

You'll see a set of messages as Kit unpacks the configuration, code, datasets, and serialized model. Now list the directory contents:

```sh
tree

.
├── Kitfile
├── README.md
├── llama3-8b-8B-instruct-q4_0.gguf
├── lora-adapter.gguf
└── training-data.txt* A Kitfile
```

The [Kitfile](../kitfile/kf-overview/) is the manifest for our ModelKit, the serialized model, and a set of files or directories including the adapter, dataset, and docs. Every ModelKit has a Kitfile and you can use the info and inspect commands to view them from the CLI (there's more on this in our [Next Steps](../next-steps/) doc).

### 4/ Check the Local Repository

Use the [list command](../cli/cli-reference/#kit-list) to check what's in the KitOps local registry.

```sh
kit list
```

You'll see the column headings for an empty table with things like `REPOSITORY`, `TAG`, etc...

The local registry is found at `$KITOPS_HOME/storage`. The $KITOPS_HOME location is system dependent:
	- Linux: `$XDG_DATA_HOME/kitops` with a fall back to `$HOME/.local/share/kitops`
	- MacOS: `~/Library/Caches/kitops`
	- Windows: `%LOCALAPPDATA%\kitops`

### 5/ Pack the ModelKit

Since our repository is empty we'll need to use the [pack command](../cli/cli-reference/#kit-pack) to create our ModelKit. The ModelKit in your local registry will need to be named the same as your remote registry. The command will look like: `kit pack . -t [your registry address]/[your registry user or organization name]/[your repository name]:[your tag name]`

In our case, we are pushing a ModelKit tagged `latest` to:
* The [Jozu Hub](https://jozu.ml/) registry
* The `brad` user organization
* The `quick-start` repository

As a result, the command will look like:

```sh
kit pack . -t jozu.ml/brad/quick-start:latest
```

You may need to substitute your own registry, user, repository, or tag names.

Once complete, you'll see a set of `Saved ...` messages as each piece of the ModelKit is saved to the local repository.

Check your local registry again:

```sh
kit list
```

You should see an entry named based on whatever you used in your pack command.

### 6/ (Optional) Remove a ModelKit from a Local Repository

If you have a typo when packing a ModelKit you can easily remove it from your repository and try again. The [Next Steps guide includes information on how to remove ModelKits](../next-steps/#remove-command).

Once you've removed the mistaken ModelKit from the repository, you can repeat the `kit pack` command in the previous step, being sure to provide the correct organization and repository name for your ModelKit.

### 7/ Push the ModelKit to a Remote Repository

The [push command](../cli/cli-reference/#kit-push) will copy the newly built ModelKit from your local repository to the remote repository you logged into earlier. The naming of your ModelKit will need to be the same as what you see in your `kit list` command (REPOSITORY:TAG). You can even copy and paste it. In our case it looks like:

```sh
kit push jozu.ml/brad/quick-start:latest
```

Note that some registries, like Jozu Hub, don't automatically create a repository. If you receive an error from your `push` command, make sure you have created the repository in your target registry and that you have push rights to the repository.

### Congratulations

You've learned how to unpack a ModelKit, pack one up, and push it. Anyone with access to your remote repository can now pull your new ModelKit and start playing with your model using the `kit pull` or `kit unpack` commands.

If you'd like to learn more about using Kit, try our [Next Steps with Kit](../next-steps/) document that covers:
* Creating a container or Kubernetes deployment from a ModelKit
* Signing your ModeKit
* Making your own Kitfile
* The power of `unpack`
* Tagging ModelKits
* Keeping your registry tidy

Or, if you want to run an LLM-based ModelKit locally try our [dev mode](../deploy/#running-llms-locally).

Finally, if you're building workflows using Dagger you can use KitOps through our [Daggerverse modules](https://github.com/kitops-ml/daggerverse). Or get the [GitHub Action for Kit](https://github.com/marketplace/actions/setup-kit-cli).

Thanks for taking some time to play with Kit. We'd love to hear what you think. Feel free to drop us an [issue in our GitHub repository](https://github.com/kitops-ml/kitops/issues) or join [our Discord server](https://discord.gg/Tapeh8agYy).
