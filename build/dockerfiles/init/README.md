# Kit CLI init container

This image is designed to be used as an init container on platforms such as Kubernetes and supports

* Unpacking a ModelKit to a specific path within the container
* Verifying a ModelKit's signature before unpacking via cosign

This image is configured through setting environment variables:

* The environment variable `$MODELKIT_REF` is required and should contain the name of the modelkit to unpack
* Environment variables `$UNPACK_PATH` and `$UNPACK_FILTER` are optional configure how the ModelKit is unpacked:
  * `$UNPACK_PATH` can be used to unpack the ModelKit to a specific directory within the container (default is `/home/user/modelkit`)
  * `$UNPACK_FILTER` can be used to configure which parts of the modelkit are unpacked (the default is to unpack only the model)

  For the expected format for these environment variables matches the usage of the `--dir` and `--filter` flags on `kit unpack`, respectively.

In addition, this image can be used to verify the signature on a ModelKit in a registry:

* To verify a ModelKit via a key, set the `$COSIGN_KEY` environment variable:
  * The value of `$COSIGN_KEY` can be a path to a key mounted inside the container, a URL to the key, or a KMS URI
* To verify the ModelKit using cosign's keyless verification, set the `$COSIGN_CERT_IDENTITY` and `$COSIGN_CERT_OIDC_ISSUER` environment variables:
  * `$COSIGN_CERT_IDENTITY` should contain the identity of the expected signer within the OIDC issuer
  * `$COSIGN_CERT_OIDC_ISSUER` should contain the URL to the expected OIDC issuer used to sign the modelkit (e.g. `https://github.com/login/oauth`)

  For additional documentation how verifying ModelKit signatures works, see the [`cosign` documentation](https://docs.sigstore.dev/verifying/verify/)

## Example

Below is a minimal example of using this image to download a modelkit named `ghcr.io/kitops-ml/my-modelkit:latest` to a container

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-modelkit-pod
spec:
  initContainers:
  - name: kitops-init
    image: ghcr.io/kitops-ml/kitops-init:latest
    env:
      - name: MODELKIT_REF
        value: "ghcr.io/kitops-ml/my-modelkit:latest"
      - name: UNPACK_PATH
        value: /tmp/my-modelkit
      - name: COSIGN_CERT_IDENTITY
        value: kitops@jozu.com
      - name: COSIGN_CERT_OIDC_ISSUER
        value: https://github.com/login/oauth
    volumeMounts:
      - name: modelkit-storage
        mountPath: /tmp/my-modelkit
  containers:
  - name: main-container
    image: alpine:latest
    volumeMounts:
      - name: modelkit-storage
        mountPath: /my-modelkit
    command: ["tail"]
    args: ["-f", "/dev/null"]
  volumes:
  - name: modelkit-storage
    emptyDir: {}
```

## Building the image

Building the image requires docker or podman. To build the image, run the following command in this directory:

```shell
docker build -t kit-init-container:latest .
```

By default, the image will be built using `ghcr.io/kitops-ml/kitops:next` as a base. This can be overridden (to build using a specific version of Kit, for example) by using the build arg `KIT_BASE_IMAGE`:

```shell
# Build the image based on Kit v1.3.0 instead of 'next'
docker build -t kit-init-container:latest --build-arg KIT_BASE_IMAGE=ghcr.io/kitops-ml/kitops:v1.3.0 .
```
