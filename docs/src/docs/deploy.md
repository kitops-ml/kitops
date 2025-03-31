---
description: Deploy your ModelKits with ease. Learn how to package, distribute, and run your AI/ML models in production environments.
---
# Deploying ModelKits

This page outlines how to use `init` or Kit CLI containers to deploy a ModelKit-packaged AI/ML project to Kubernetes or any other container runtime. The KitOps repo provides pre-built ModelKits that can be used for both semi-turnkey solutions, and more DIY options.

## Pre-built Containers

There are currently two pre-built containers:

1. Init container: https://github.com/kitops-ml/kitops/blob/main/build/dockerfiles/init/README.md
1. Kit CLI container: https://github.com/kitops-ml/kitops/blob/main/build/dockerfiles/README.md

## Init Container

The init container unpacks the model reference from a ModelKit to a specific path and then exits. This makes it useful as a Kubernetes `init` container. This container also supports verifying signatures for containers automatically from key-based or keyless signers.

The container is configurable via environment variables:

`$MODELKIT_REF`: The ModelKit to pull (required).
`$UNPACK_PATH`: Where to unpack the ModelKit (normally you’d want a `volumeMount` here). This is required and will default to `/home/user/modelkit`.
`$UNPACK_FILTER`: Optional filter to limit what is unpacked (e.g., just the model, or model + code). The filter format is the same as the [--filter command line argument](../cli/cli-reference/) for the Kit CLI.
`$COSIGN_KEY`: Path to the key that should be used for verification, mounted inside the init container (e.g., from a Kubernetes secret).
`$COSIGN_CERT_IDENTITY`: Signing identity for keyless signing.
`$COSIGN_CERT_OIDC_ISSUER`: OIDC issuer for keyless signer identity.

### Example Kubernetes YAML

```
 apiVersion: v1
 kind: Pod
 metadata:
   name: my-modelkit-pod
 spec:
   containers:
     - name: model-server
       image: "" # Some container that expects your modelkit
       # Share a volumeMount between the init container and this one
       volumeMounts:
         - name: modelkit-storage
           mountPath: /my-modelkit

   # Run the init container to unpack the ModelKit into the volume mount and make
   # it available to the main container
   initContainers:
     - name: kitops-init
       image: ghcr.io/kitops-ml/kitops-init:latest
       env:
         - name: MODELKIT_REF
           value: "ghcr.io/kitops-ml/my-modelkit:latest"
         - name: UNPACK_PATH
           value: /tmp/my-modelkit
       volumeMounts:
         - name: modelkit-storage
           mountPath: /tmp/my-modelkit

   # Define a volume to store the ModelKit
   volumes:
     - name: modelkit-storage
       emptyDir: {}
```

## Using the Kit CLI Container

The containerized Kit CLI can be used to tailor the running of a ModelKit because you can run any Kit CLI command. This gives you flexibility, but more manual work (the world is your oyster, but it may be hard to shuck).

This container runs `kit` as its entrypoint, accepting Kit CLI arguments. So you could run the container instead of downloading and installing the Kit CLI, although you’ll need to mount a docker volume.

Docker run example:

`docker run ghcr.io/kitops-ml/kit:latest pull jozu.ml/jozu/llama3-8b:8B-instruct-q5_0`

Kubernetes example:

```
 apiVersion: v1
 kind: Pod
 metadata:
   name: my-modelkit-pod
 spec:
   containers:
     - name: me-using-kit
       image: ghcr.io/kitops-ml/kit:latest
       args: # You can put whatever you want; args is an array
         - pull
         - jozu.ml/jozu/llama3-8b:8B-instruct-q5_0
```

## Creating a Custom Container

Going a step further you can use the Kit CLI container to create your own bespoke ModelKit containers.

Example `dockerfile` for a custom container that has `my-modelkit` built into it:

```
 # Staged build to grab the ModelKit so we can use it later
 FROM ghcr.io/kitops-ml/kit:latest AS modelkit-download

 # Download your ModelKit into the container
 RUN kit unpack my-modelkit /tmp/my-modelkit

 # Actual build stage; this just uses Alpine but you would build whatever
 # container you need here
 FROM alpine:latest

 # Normal container build + setup -- depends on your use case
 # ...
 # ...

 # Copy the downloaded ModelKit into this container
 COPY --from=modelkit-download /tmp/my-modelkit /home/user/modelkit-data
```

**Questions or suggestions?** Drop an [issue in our GitHub repository](https://github.com/kitops-ml/kitops/issues) or join [our Discord server](https://discord.gg/Tapeh8agYy) to get support or share your feedback.
