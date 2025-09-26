# KFP ModelKit Component

## Overview
This Kubeflow Pipeline (KFP) component enables packaging a ModelKit model and pushing it to an OCI (Open Container Initiative) registry. It is designed to be used as a step in automated ML pipelines, making it easy to build, package, and distribute models using ModelKit within Kubeflow.

## Files
- `component.py`: Python script implementing the component logic.
- `component.yaml`: KFP component specification for pipeline integration.
- `Dockerfile`: Container definition for the component runtime environment.

## Usage
1. **Build the Docker image:**
   ```bash
   docker build -t <your-repo>/kfp-modelkit-component:latest .
   ```
2. **Push the image to your registry:**
   ```bash
   docker push <your-repo>/kfp-modelkit-component:latest
   ```
3. **Use in a KFP pipeline:**
   - Reference `component.yaml` in your pipeline definition.
   - Configure input parameters as needed (see below).

## Inputs & Parameters
Refer to `component.yaml` for the full list of parameters. Typical parameters include:
- Model source path
- Model name and version
- OCI registry URL and credentials
- Additional packaging options

## Example Pipeline Step
```python
import kfp
from kfp.components import load_component_from_file

modelkit_op = load_component_from_file('component.yaml')

# Example usage in a pipeline
def my_pipeline(...):
    modelkit_op(
        model_path='path/to/model',
        model_name='my-model',
        version='1.0.0',
        registry_url='oci://my-registry',
        # ...other params...
    )
```

## Notes
- Ensure all required environment variables and credentials are available at runtime.
- The component is designed for use with ModelKit and OCI-compliant registries.
- For advanced usage, customize `component.py` or the Dockerfile as needed.

## References
- [Kubeflow Pipelines Documentation](https://www.kubeflow.org/docs/components/pipelines/)
- [ModelKit Documentation](https://github.com/kitops-ml/modelkit)
- [OCI Registry Specification](https://opencontainers.org/)

---

Feel free to adapt this file for official documentation or as a companion PR note.