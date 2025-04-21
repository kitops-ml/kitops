---
description: Use KitOps PyKitOps Python library to automate ModelKit creation with MLFlow.
---
# MLFlow and KitOps ModelKits

Many KitOps users automate the creation of secure and tamper-proof [ModelKits](../modelkit/intro.md) for each experiment run in MLFlow. This gives the team and organization a library of models that they can reproduce on any infrastructure.

::tip
You can follow the instructions below, or in a [community-generated tutorial](https://jozu.com/blog/how-to-use-kitops-with-mlflow/)
:::

## Install PyKitOps Python Library

PyKitOps is a community provided Python library for simplifying the creation and management of ModelKits from code. First install PyKitOps into the python interpreter used for by MLFlow for experiment runs:

```sh
pip install kitops
```

## Downloading the Run Artifacts

After the experiment run is complete, download the artifacts to a local directory - this is typically after the current `with mlflow.start_run() as cur_run` execution is complete.

```py
artifact_location = mlflow.artifacts.download_artifacts(cur_run.info.run_id)
```

This will return a local directory of the artifacts (referred to as `artifact_location` below).

## Packaging the Artifacts as a ModelKit

The `ModelKitManager` is then used to pack and upload the artifacts as a ModelKit.

In the following code:
`artifact_location` = the directory where the experiment run artifacts were saved
`name` = the name of the ModelKit as it will be displayed in the [Kitfile](../kitfile/kf-overview.md)
`modelkit_tag` = a name for the ModelKit, for example "latest"

```py
# Add the KitOps ModelKitManager
from kitops.modelkit.manager import ModelKitManager
from kitops.modelkit.user import UserCredentials
from kitops.cli import kit

# A password can be read from Environment Variables or .env files: JOZU_PASSWORD=<secret password> 
# Add in your own username from the registry that you are accessing, e.g. username=bmicklea
creds = UserCredentials(username=username, registry="jozu.ml")

# Initialize the ModelKitManager with:
# - the working directory ("artifact_location")
# - reference to the creds variable
# - a tag name e.g. modelkit_tag=latest
manager = ModelKitManager(working_directory=artifact_location, user_credentials=creds, modelkit_tag=modelkit_tag)

# Log into the registry where the ModelKit will be pushed and stored
manager.login()

# Create a new Kitfile (your ModelKit's "recipe") based on the contents of the working directory
kit.init(directory=artifact_location, name=name, description="my cool project", author=username)

# Pack the ModelKit using the Kitfile recipe, and push it to the registry
manager.pack_and_push_modelkit(with_login_and_logout=False)
```

**Questions or suggestions?** Drop an [issue in our GitHub repository](https://github.com/kitops-ml/kitops/issues) or join [our Discord server](https://discord.gg/Tapeh8agYy) to get support or share your feedback.