# KFP ModelKit Component

## Overview
The KFP (Kubeflow Pipeline) ModelKit Component is a pipeline component that enables seamless integration of ModelKit operations within Kubeflow Pipelines. This component allows you to package and push ModelKit models to OCI registries as part of your ML pipelines.

## Installation

### Prerequisites
- Kubeflow Pipelines environment
- Access to an OCI-compatible registry
- ModelKit CLI installed

### Component Setup
1. Clone the kitops repository
2. Build the component image:
   ```bash
   docker build -t <your-registry>/kfp-modelkit-component:latest build/kfp-modelkit-component/
   ```
3. Push the image to your registry:
   ```bash
   docker push <your-registry>/kfp-modelkit-component:latest
   ```

## Usage

### Basic Pipeline Integration
```python
import kfp
from kfp.components import load_component_from_file

# Load the component
modelkit_op = load_component_from_file('component.yaml')

# Define your pipeline
@kfp.dsl.pipeline(
    name='ModelKit Pipeline Example',
    description='Example pipeline using ModelKit component'
)
def modelkit_pipeline():
    modelkit_task = modelkit_op(
        model_path='path/to/model',
        model_name='my-model',
        version='1.0.0',
        registry_url='oci://my-registry'
    )
```

### Component Parameters

| Parameter | Type | Description | Required |
|-----------|------|-------------|----------|
| model_path | string | Path to the model directory or file | Yes |
| model_name | string | Name of the model | Yes |
| version | string | Version tag for the model | Yes |
| registry_url | string | OCI registry URL | Yes |

### Environment Variables
The component supports the following environment variables:
- `REGISTRY_USERNAME`: Registry authentication username
- `REGISTRY_PASSWORD`: Registry authentication password
- `ADDITIONAL_BUILD_ARGS`: Extra arguments for model packaging

## Examples

### Basic Model Packaging
```python
modelkit_op(
    model_path='./models/bert',
    model_name='bert-base',
    version='1.0.0',
    registry_url='oci://registry.example.com'
)
```

### With Authentication
```python
modelkit_op(
    model_path='./models/bert',
    model_name='bert-base',
    version='1.0.0',
    registry_url='oci://registry.example.com',
    env={
        'REGISTRY_USERNAME': '{{registry_username}}',
        'REGISTRY_PASSWORD': '{{registry_password}}'
    }
)
```

## Troubleshooting

### Common Issues

1. **Registry Authentication Failed**
   - Verify credentials are correctly set
   - Ensure registry URL is correct
   - Check network connectivity

2. **Model Packaging Errors**
   - Validate model path exists
   - Verify model structure
   - Check disk space

### Debugging

Enable debug logs by setting the environment variable:
```python
modelkit_op(
    ...
    env={'MODELKIT_LOG_LEVEL': 'DEBUG'}
)
```

## Contributing
Contributions are welcome! Please see our [Contributing Guide](../contributing.md) for details.

## License
This component is part of kitops and is licensed under the same terms. See the [LICENSE](../../LICENSE) file for details.