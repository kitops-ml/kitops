import os
import subprocess
import json
from typing import Optional
from ml_metadata import metadata_store
from ml_metadata.proto import metadata_store_pb2

# This script is intended to be used as a Kubeflow Pipeline component entrypoint.
# It expects the following environment variables or arguments:
#   MODEL_DIR: Path to model artifacts
#   REGISTRY_URI: OCI registry URI (e.g., oci://my-registry/modelkit:tag)
#   PIPELINE_RUN_ID: Kubeflow pipeline run ID
#   EXPERIMENT_NAME: Kubeflow experiment name
#   KUBEFLOW_METADATA_HOST: MLMD gRPC host (default: metadata-grpc-service.kubeflow)
#   DOCKER_CONFIG_PATH: Path to Docker config for registry auth (default: /kaniko/.docker)

def main():
    model_dir = os.environ.get('MODEL_DIR')
    registry_uri = os.environ.get('REGISTRY_URI')
    pipeline_run_id = os.environ.get('PIPELINE_RUN_ID')
    experiment_name = os.environ.get('EXPERIMENT_NAME')
    kubeflow_metadata_host = os.environ.get('KUBEFLOW_METADATA_HOST', 'metadata-grpc-service.kubeflow')
    docker_config_path = os.environ.get('DOCKER_CONFIG_PATH', '/kaniko/.docker')

    if not model_dir or not registry_uri or not pipeline_run_id or not experiment_name:
        raise ValueError("Missing required environment variables.")

    # 1. Connect to ML Metadata
    store = metadata_store.MetadataStore(
        metadata_store_pb2.ConnectionConfig(
            host=kubeflow_metadata_host,
            port=8080,
        )
    )
    # 2. Extract run/experiment metadata
    runs = store.get_executions_by_type('kfp-run')
    run = next((r for r in runs if r.custom_properties.get('run_id', None) and r.custom_properties['run_id'].string_value == pipeline_run_id), None)
    metrics = run.custom_properties.get('metrics', {}).string_value if run and 'metrics' in run.custom_properties else '{}'

    # 3. Prepare ModelKit metadata
    metadata = {
        "pipeline_run_id": pipeline_run_id,
        "experiment_name": experiment_name,
        "metrics": metrics,
    }
    with open('modelkit-metadata.json', 'w') as f:
        json.dump(metadata, f)

    # 4. Package model as ModelKit (assumes modelkit CLI is installed)
    subprocess.run([
        'modelkit', 'pack',
        '--input', model_dir,
        '--metadata', 'modelkit-metadata.json',
        '--output', 'modelkit.tar'
    ], check=True)

    # 5. Push to OCI registry
    env = os.environ.copy()
    env['DOCKER_CONFIG'] = docker_config_path
    subprocess.run([
        'modelkit', 'push',
        '--input', 'modelkit.tar',
        '--destination', registry_uri
    ], check=True, env=env)

    print(f"ModelKit pushed to {registry_uri}")

if __name__ == "__main__":
    main()
