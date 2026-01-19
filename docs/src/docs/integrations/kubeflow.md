---
description: Learn how to build end-to-end MLOps pipelines using KitOps, Kubeflow Pipelines, and KServe.
---

# End-to-End MLOps with Kubeflow and KServe

KitOps bridges the gap between your experimentation, training, and deployment workflows by providing a standard, versioned artifact format—the **ModelKit**. By integrating KitOps with [Kubeflow Pipelines (KFP)](https://www.kubeflow.org/docs/components/pipelines/) and [KServe](https://kserve.github.io/website/master/), you can create a robust, reproducible, and secure supply chain for your ML models.

This guide covers:
1.  Using KitOps in Kubeflow Pipelines to manage code and data.
2.  Packaging trained models as immutable ModelKits.
3.  Deploying ModelKits directly to KServe.

## Architecture Overview

1.  **Ingest**: A KFP component uses `kit unpack` to retrieve the versioned training code and dataset from a ModelKit.
2.  **Train**: The training logic executes.
3.  **Package**: A KFP component uses `kit pack` and `kit push` to save the trained model, hyperparameters, and metrics as a new ModelKit layer.
4.  **Serve**: KServe pulls the ModelKit using the [KitOps Storage Container](./kserve.md) and serves the model.

## Part 1: Kubeflow Pipelines Integration

To usage KitOps in your pipeline steps, you need a container image with the `kit` CLI installed. You can use the official image `ghcr.io/kitops-ml/kitops:latest` or build a custom image containing both your training dependencies and the `kit` tool.

### Example: Training Component

Below is an example of a generic KFP component that unpacks a ModelKit, runs a training script, and pushes the result.

```python
from kfp import dsl

def train_with_kitops_op(
    source_kit_url: str,
    output_kit_tag: str,
    registry_user: str,
    registry_pass: str
):
    """
    A unified component that unpacks code/data, trains, and repacks the model.
    """
    return dsl.ContainerOp(
        name='Train and Pack',
        image='ghcr.io/kitops-ml/kitops:latest', # Use an image with python + kit
        command=['sh', '-c'],
        arguments=[
            f"""
            set -e
            
            # 1. Login to Registry
            # Note: In production, consider using better secret management
            echo "Logging in to registry..."
            kit login -u {registry_user} -p {registry_pass} $(echo {source_kit_url} | cut -d/ -f1)

            # 2. Unpack Source (Code + Data)
            echo "Unpacking source kit..."
            mkdir -p /workspace
            kit unpack {source_kit_url} -d /workspace

            # 3. Helper: Install dependencies if needed (better to have them pre-baked in image)
            # pip install -r /workspace/requirements.txt

            # 4. Train
            echo "Starting training..."
            # Assuming train.py saves model to /workspace/model.pkl
            python3 /workspace/train.py --data /workspace/data --output /workspace/model

            # 5. Create a new Kitfile for the artifact
            echo "Creating Kitfile..."
            cat <<EOF > /workspace/Kitfile
            manifestVersion: v1.0.0
            package:
              name: my-model
              version: 0.0.1
              description: Trained model from KFP run
            layers:
              - path: ./model
                type: model
            EOF

            # 6. Pack and Push
            echo "Packing and pushing artifact..."
            kit pack /workspace -t {output_kit_tag}
            kit push {output_kit_tag}
            
            echo "Successfully pushed {output_kit_tag}"
            """
        ]
    )
```

> **Note**: In a real-world scenario, you might split the "Unpack" and "Train" steps if you want to use a pure training image (e.g., standard TensorFlow/PyTorch image) for the actual compute-heavy work. You can use a `VolumeOp` or `PVC` to share the unpacked data between the `kit` step and the training step.

## Part 2: Sending to KServe

Once your pipeline has successfully pushed the trained ModelKit (e.g., `my-registry.io/models/my-model:run-123`), you can obtain this tag and pass it to a deployment step.

Kubeflow Pipelines can trigger KServe deployments by applying an `InferenceService` manifest.

### Prerequisites
Ensure your cluster is configured with the **KitOps ClusterStorageContainer** as described in the [KServe Integration Guide](./kserve.md).

### Deployment Component
Your pipeline can execute `kubectl apply` to deploy the service:

```yaml
apiVersion: serving.kserve.io/v1beta1
kind: InferenceService
metadata:
  name: "my-model-service"
spec:
  predictor:
    model:
      modelFormat:
        name: sklearn
      storageUri: "kit://my-registry.io/models/my-model:run-123" # The output tag from Part 1
```

## Best Practices

### 1. Immutable Tags
Avoid using `:latest` in your pipelines.
*   **Source**: Pin your input code/data ModelKits to a specific version or hash.
*   **Output**: Generate unique tags for every run. A common pattern is `my-model:pipeline-run-<RUN_ID>`. This makes it easy to trace exactly which code and data produced a specific model.

### 2. Separation of Layers
*   **Base Layer**: Your "Source" ModelKit should contain your training code and potentially configuration.
*   **Data Layer**: If data is large, keep it in a separate ModelKit or layer that doesn't change as often as code.
*   **Model Layer**: The output ModelKit should primarily contain the model artifact.

### 3. Secret Management
Do not hardcode credentials in your pipeline definitions.
*   Use **Kubernetes Secrets** and mount them into your pipeline components.
*   For AWS ECR or Google Artifact Registry, use Workload Identity or IRSA so that `kit` can authenticate without explicit credentials.

### 4. Efficient Caching
KFP caches steps based on inputs. Since `kit unpack` is deterministic for a given digest, KFP can efficiently skip the download step if the `kit` URL (including digest) hasn't changed, speeding up your experiments.

## Troubleshooting

*   **Authentication Errors**: Ensure the `kit` CLI has access to the registry. Check if the Kubernetes ServiceAccount running the pipeline pod has the correct permissions or if secrets are mounted correctly.
*   **Storage URI Protocol**: In KServe, ensure your `storageUri` starts with `kit://`. If it doesn't, KServe won't invoke the KitOps container.
*   **PVC Access**: If separating "unpack" and "train" steps, ensure the Persistent Volume Claim (PVC) is correctly mounted and writable by both containers.
