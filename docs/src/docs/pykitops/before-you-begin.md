---
title: Before you begin with PyKitOps
---
## Before You Begin

This project was created using Python v3.12, but works with Python versions >= 3.10.

### 1/ Install the Kit CLI

The PyKitOps SDK uses the Kit CLI to manage ModelKits so you'll need an up-to-date Kit CLI version on your machine.

To determine if the Kit CLI is installed in your environment:
1. Open a Terminal window
1. Run the following command:

   ```bash
   kit version
   ```

1. You should see output similar to the following:

    ```
    Version: 0.4.0
    Commit: e2e83d953823ac35648f2f76602a0cc6e8ead819
    Built: 2024-11-05T20:29:07Z
    Go version: go1.22.6
    ```

If you don't have the Kit CLI installed, follow the [Kit Installation Instructions](https://kitops.org/docs/cli/installation/).

### 2/ Prepare Your Registry

To get the most out of ModelKits we strongly suggest you [sign up for a free account at Jozu.ml](https://jozu.ml/signup).

The [Jozu Hub](https://jozu.ml/) will:
* Automatically generates a container from a ModelKit
* Show you details about the various parts of your ModelKit at a glance
* Indicate whether ModelKits are signed

Alternatively, ModelKits can be stored in any OCI 1.1-compliant container registry, however, you'll need to set the `JOZU_REGISTRY` environment variable in addition to the username, password, and namespace (see the next section for details).

### 3/ Set Your Environment

1. In the root directory of your project (the *"Project directory"*) create a `.env` file.
2. Edit the `.env` file by adding an entry for your `JOZU_USERNAME`, your `JOZU_PASSWORD` and your `JOZU_NAMESPACE` (this should match the repository name you'll be pushing to in the regsitry). If you're *not* using the Jozu Hub you'll also need to set the `JOZU_REGISTRY` variable to point to the URL for your registry.

    An example `.env` file for Jozu Hub will look like this:

    ```bash
      JOZU_USERNAME=brett@jozu.org
      JOZU_PASSWORD=my_password
      JOZU_NAMESPACE=brett
    ```

    An example `.env` file for another registry will look like this:

    ```bash
      JOZU_REGISTRY=hub.docker.com
      JOZU_USERNAME=brett@jozu.org
      JOZU_PASSWORD=my_password
      JOZU_NAMESPACE=brett
    ```

    - The Kitops Manager uses the entries in the `.env` file to login to [Jozu.ml](https://jozu.ml).
    - As an alternative to using a `.env` file, you can create Environment Variables for each of the entries above.
3. Be sure to save the changes to your .env file before continuing.

That's it! You can check out the How To Guide to see an example of how to use the SDK.
