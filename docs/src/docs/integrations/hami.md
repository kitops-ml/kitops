# Sharing GPUs for ModelKit Workloads with HAMi

[HAMi](https://project-hami.io/) allocates GPU resources to Kubernetes workloads. KitOps delivers the model files. No changes to the ModelKit are needed: a KitOps init container unpacks the model into a shared volume, and only the application container requests a GPU.

## Prerequisites

- A Kubernetes cluster with HAMi configured for NVIDIA GPU sharing. See the [HAMi prerequisites](https://project-hami.io/docs/installation/prerequisites).
- `kubectl` configured for the cluster.
- Cluster access to the container registries and the public sample ModelKit registry used below.

## Run the example Pod

Save this manifest as `hami-modelkit.yaml`. The sample ModelKit contains BERT model weights. The CUDA container is a smoke test, not an inference server.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: hami-modelkit
spec:
  securityContext:
    fsGroup: 1001
  volumes:
    - name: modelkit
      emptyDir: {}
  initContainers:
    - name: kitops-init
      image: ghcr.io/kitops-ml/kitops-init:v1.17.0
      env:
        - name: MODELKIT_REF
          value: jozu.ml/jozu/bert-base-uncased:safetensors
        - name: UNPACK_PATH
          value: /modelkit
      volumeMounts:
        - name: modelkit
          mountPath: /modelkit
  containers:
    - name: app
      image: nvcr.io/nvidia/cuda:12.4.1-base-ubuntu22.04
      command: ["sleep", "infinity"]
      resources:
        limits:
          nvidia.com/gpu: 1
          nvidia.com/gpumem: 4096
      volumeMounts:
        - name: modelkit
          mountPath: /modelkit
          readOnly: true
```

The example requests one virtual GPU with 4096 MiB of GPU memory using [HAMi's default NVIDIA resource names](https://project-hami.io/docs/userguide/nvidia-device/examples/allocate-device-memory). It assumes the HAMi admission webhook configures the scheduler for GPU Pods. If your cluster uses custom resource names or requires an explicit `schedulerName`, use the values from your HAMi installation.

```sh
kubectl apply -f hami-modelkit.yaml
kubectl wait --for=condition=Ready pod/hami-modelkit --timeout=10m
kubectl exec hami-modelkit -c app -- test -s /modelkit/model.safetensors
kubectl exec hami-modelkit -c app -- nvidia-smi -L
```

The file check exits successfully if the weights are present and nonempty. The last command lists the GPU visible to the application. If the Pod is pending, check the scheduling events with `kubectl describe pod hami-modelkit`. If the init container fails, check `kubectl logs hami-modelkit -c kitops-init`.

To serve a model instead, use your inference image and point it at the unpacked files under `/modelkit`. Set the GPU memory request for that workload. For signature verification and other init container options, see [Deploying a ModelKit to Kubernetes](../k8s-init-container.md).

When finished, delete the Pod with `kubectl delete -f hami-modelkit.yaml`.

<!-- AGENT_MODIFIED: Human review required before merge -->
