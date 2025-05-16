# Deploying a ModelKit to Kubernetes with the Kit Init Container

This tutorial shows you how to fetch, verify, and unpack a ModelKit into your Kubernetes Pod using the Kit init container image `ghcr.io/kitops-ml/kitops-init:latest`.

## Overview

The Kit init container lets you:

1. **Fetch** a ModelKit from an OCI registry.
2. **Verify** its signature via Sigstore [Cosign][cosign] (optional).
3. **Unpack** it to a shared volume.
4. **Use** it in your application container just like any local files.

## Prerequisites

- A working Kubernetes cluster (Minikube, Kind, GKE, EKS, etc.)
- `kubectl` [configured](https://kubernetes.io/docs/tasks/tools/) to talk to that cluster
- A ModelKit published at an OCI registry (e.g. `ghcr.io/kitops-ml/my-modelkit:latest`)
- *(Optional)* Cosign key or OIDC issuer for signature verification

## 1. Define a Shared Volume

Use an `emptyDir` volume so both containers share the unpacked files:

```yaml
volumes:
  - name: modelkit-storage
    emptyDir: {}
```

`emptyDir` have some limitations especially with large models. See the Kubernetes [docs](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for limitations and configuration of `emptyDir`.

## 2. Configure the Init Container

Add an `initContainer` that:

- Uses the Kit init image
- Sets the **required** env var `MODELKIT_REF`
- Sets **optional** verification and unpacking env vars

```yaml
initContainers:
  - name: kitops-init
    image: ghcr.io/kitops-ml/kitops-init:latest
    env:
      # 1) Which ModelKit to fetch
      - name: MODELKIT_REF
        value: "jozu.ml/jozu/bert-base-uncased:safetensors"
      # 2) Optional: where to unpack inside the volume
      - name: UNPACK_PATH
        value: "/modelkit"
      # 3) Optional: only unpack specific files (Kit CLI --filter syntax)
      - name: UNPACK_FILTER
        value: "model"
      # 4a) Optional: verify via local Cosign key
      - name: COSIGN_KEY
        value: "/keys/cosign.pub"
      # 4b) Alternately Optional: keyless OIDC verification
      - name: COSIGN_CERT_IDENTITY
        value: "kitops@jozu.com"
      - name: COSIGN_CERT_OIDC_ISSUER
        value: "https://github.com/login/oauth"
    volumeMounts:
      - name: modelkit-storage
        mountPath: "/modelkit"
    # If using a mounted key secret:
    # volumeMounts:
    #   - name: cosign-key
    #     mountPath: /keys
```

## 3. Add Your Main Container

Mount the same volume so your application sees the unpacked ModelKit:

```yaml
containers:
  - name: app
    image: alpine:latest
    command: ["sh", "-c", "echo 'Contents of /modelkit:' && ls /modelkit && sleep 3600"]
    volumeMounts:
      - name: modelkit-storage
        mountPath: /modelkit
```

> In a real app, replace the command above with your model server startup, pointing it at `/modelkit`.

## 4. Full Pod Example

Save this as `modelkit-demo.yaml`:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: modelkit-demo
spec:
  volumes:
    - name: modelkit-storage
      emptyDir: {}
  initContainers:
    - name: kitops-init
      image: ghcr.io/kitops-ml/kitops-init:latest
      env:
        - name: MODELKIT_REF
          value: "jozu.ml/jozu/bert-base-uncased:safetensors"
        - name: UNPACK_PATH
          value: "/modelkit"
      volumeMounts:
        - name: modelkit-storage
          mountPath: "/modelkit"
  containers:
    - name: app
      image: alpine:latest
      command: ["sh", "-c", "echo 'Contents of /modelkit:' && ls /modelkit && sleep 3600"]
      volumeMounts:
        - name: modelkit-storage
          mountPath: /modelkit
```

Apply it:

```bash
kubectl apply -f modelkit-demo.yaml
```

Check the init container logs and the unpacked files:

```bash
kubectl logs modelkit-demo -c kitops-init
kubectl exec modelkit-demo -- ls /modelkit
```

## 5. Scaling with a Deployment

For production-style workloads, use a Deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: modelkit-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: modelkit-app
  template:
    metadata:
      labels:
        app: modelkit-app
    spec:
      volumes:
        - name: modelkit-storage
          emptyDir: {}
      initContainers:
        - name: kitops-init
          image: ghcr.io/kitops-ml/kitops-init:latest
          env:
            - name: MODELKIT_REF
              value: "your-jozu-hub/your-model:latest
            - name: UNPACK_PATH
              value: "/modelkit"
          volumeMounts:
            - name: modelkit-storage
              mountPath: "/modelkit"
      containers:
        - name: app
          image: your-company/your-container:latest
          args: ["--model-dir=/modelkit"]
          volumeMounts:
            - name: modelkit-storage
              mountPath: /modelkit
```

Apply with:

```bash
kubectl apply -f modelkit-deployment.yaml
```

## Tips & Next Steps

- **Custom init container version**:  
  Use a specific version tag instead of `latest`. For example:

  ```yaml
  image: ghcr.io/kitops-ml/kitops-init:v1.5.1
  ```

- **Signature checks**: Store Cosign keys in Kubernetes Secrets and mount them.
- **Selective unpacking**: Use `UNPACK_FILTER` to extract only weights, configs, etc.
- **Persistent volumes**: Replace `emptyDir` with a PVC for data persistence beyond Pod restarts.

[cosign]: https://github.com/sigstore/cosign
