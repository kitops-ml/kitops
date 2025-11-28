"""
Example Kubeflow Pipeline integrating push-modelkit component with house prices training.

This example demonstrates the directory-based approach where the training component
creates a directory with well-named files (model.pkl, predictions.csv, train.py, README.md)
and the push-modelkit component packages the entire directory as a ModelKit.

Uses pure KFP v2.14.3 components without v1 compatibility.
"""

from kfp import dsl, kubernetes
from typing import NamedTuple


@dsl.component(
    packages_to_install=['pandas', 'xgboost', 'scikit-learn'],
    base_image='python:3.11-slim'
)
def train_house_prices(
    modelkit_dir: dsl.Output[dsl.Artifact]
):
    """Train house prices model with synthetic data."""
    import pandas as pd
    import xgboost as xgb
    from sklearn.model_selection import train_test_split
    from sklearn.datasets import make_regression
    import pickle
    import os

    # Generate synthetic data
    X, y = make_regression(n_samples=1000, n_features=10, noise=10, random_state=42)
    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)

    # Convert to DataFrame
    feature_names = [f'feature_{i}' for i in range(X.shape[1])]
    X_train_df = pd.DataFrame(X_train, columns=feature_names)
    X_test_df = pd.DataFrame(X_test, columns=feature_names)

    # Train model
    model = xgb.XGBRegressor(n_estimators=100, max_depth=7, learning_rate=0.1, random_state=42)
    model.fit(X_train_df, y_train)

    # Evaluate
    train_score = model.score(X_train_df, y_train)
    test_score = model.score(X_test_df, y_test)
    print(f"Training R² score: {train_score:.4f}")
    print(f"Test R² score: {test_score:.4f}")

    # Create directory for modelkit artifacts
    os.makedirs(modelkit_dir.path, exist_ok=True)

    # Write files with proper names directly to the directory
    model_file = os.path.join(modelkit_dir.path, 'model.pkl')
    predictions_file = os.path.join(modelkit_dir.path, 'predictions.csv')
    training_script_file = os.path.join(modelkit_dir.path, 'train.py')
    readme_file = os.path.join(modelkit_dir.path, 'README.md')

    # Save model using pickle (avoids KFP UTF-8 issues with binary formats)
    with open(model_file, 'wb') as f:
        pickle.dump(model, f)

    # Generate predictions
    predictions = model.predict(X_test_df)
    pred_df = pd.DataFrame({'Id': range(len(predictions)), 'Prediction': predictions})
    pred_df.to_csv(predictions_file, index=False)

    # Save training script
    with open(training_script_file, 'w') as f:
        f.write("""# Training Script

This model was trained using XGBoost on synthetic regression data.

## Training Configuration
- Algorithm: XGBoost Gradient Boosting
- n_estimators: 100
- max_depth: 7
- learning_rate: 0.1
- random_state: 42

## Data
- Training samples: 800
- Test samples: 200
- Features: 10 (synthetic)
""")

    # Generate README
    with open(readme_file, 'w') as f:
        f.write(f"""# House Prices Demo Model

## Model Details
- **Framework**: XGBoost {xgb.__version__}
- **Algorithm**: Gradient Boosted Trees
- **Training R² Score**: {train_score:.4f}
- **Test R² Score**: {test_score:.4f}

## Training Data
- Training samples: {len(X_train)}
- Test samples: {len(X_test)}
- Features: {X.shape[1]} (synthetic)

## Usage
```python
import pickle
with open('model.pkl', 'rb') as f:
    model = pickle.load(f)
predictions = model.predict(X_new)
```

---
Packaged with KitOps
""")


@dsl.container_component
def push_modelkit(
    registry: str,
    repository: str,
    tag: str,
    output_uri: dsl.Output[dsl.Artifact],
    output_digest: dsl.Output[dsl.Artifact],
    input_modelkit_dir: dsl.Input[dsl.Artifact],
    modelkit_name: str = '',
    modelkit_desc: str = '',
    modelkit_author: str = '',
    dataset_uri: str = '',
    code_repo: str = '',
    code_commit: str = ''
):
    """Package and push model as ModelKit with attestation.

    Outputs:
        output_uri: Tagged URI (e.g., jozu.ml/repo:tag)
        output_digest: Digest URI (e.g., jozu.ml/repo@sha256:...)
    """
    # Build command using safe argument passing
    return dsl.ContainerSpec(
        image='kubeflow:dev',
        command=['/bin/bash', '-c'],
        args=[
            '''
            export DOCKER_CONFIG=/home/user/.docker && \
            /scripts/push-modelkit.sh \
            "$0" "$1" "$2" \
            --modelkit-dir "$3" \
            ${4:+--name "$4"} \
            ${5:+--desc "$5"} \
            ${6:+--author "$6"} \
            ${7:+--dataset-uri "$7"} \
            ${8:+--code-repo "$8"} \
            ${9:+--code-commit "$9"} \
            && cp /tmp/outputs/uri "${10}" \
            && cp /tmp/outputs/digest "${11}"
            ''',
            registry,
            repository,
            tag,
            input_modelkit_dir.path,
            modelkit_name,
            modelkit_desc,
            modelkit_author,
            dataset_uri,
            code_repo,
            code_commit,
            output_uri.path,
            output_digest.path
        ]
    )

@dsl.pipeline(
    name='house-prices-with-modelkit',
    description='Train house prices model and package as ModelKit'
)
def house_prices_pipeline(
    registry: str = 'jozu.ml',
    repository: str = 'demo/house-prices',
    tag: str = 'latest',
    dataset_source_uri: str = 'synthetic',
    code_repo: str = 'github.com/kitops-ml/kitops',
    code_commit: str = 'main'
):
    """
    Complete pipeline that trains a house prices model and packages it as a ModelKit.

    Args:
        registry: Container registry (e.g., jozu.ml)
        repository: Repository path for ModelKit (e.g., demo/house-prices)
        tag: ModelKit tag
        dataset_source_uri: Source URI of dataset for attestation
        code_repo: Code repository for attestation
        code_commit: Git commit hash for attestation
    """

    # Train model with synthetic data
    train = train_house_prices()

    # Package as ModelKit with directory of artifacts
    push = push_modelkit(
        registry=registry,
        repository=repository,
        tag=tag,
        # Pass directory containing all artifacts
        input_modelkit_dir=train.outputs['modelkit_dir'],
        # Metadata
        modelkit_name='House Prices Demo Model',
        modelkit_desc='XGBoost model trained on synthetic data for KitOps demo',
        modelkit_author='KitOps Team',
        # Attestation metadata
        dataset_uri=dataset_source_uri,
        code_repo=code_repo,
        code_commit=code_commit
    )

    # Mount docker-config secret for registry authentication
    kubernetes.use_secret_as_volume(
        push,
        secret_name='docker-config',
        mount_path='/home/user/.docker'
    )


if __name__ == '__main__':
    import kfp

    # Check KFP version and use appropriate compiler
    kfp_version = kfp.__version__
    print(f"Using KFP version: {kfp_version}")

    if kfp_version.startswith('2.'):
        # KFP v2 - compile with v1 compatibility
        from kfp import compiler
        compiler.Compiler().compile(
            pipeline_func=house_prices_pipeline,
            package_path='house-prices-with-modelkit.yaml'
        )
    else:
        # KFP v1
        import kfp.compiler as compiler
        compiler.Compiler().compile(
            pipeline_func=house_prices_pipeline,
            pipeline_name='house-prices-with-modelkit',
            package_path='house-prices-with-modelkit.yaml'
        )

    print("Pipeline compiled successfully!")
    print("Upload house-prices-with-modelkit.yaml to Kubeflow UI")
