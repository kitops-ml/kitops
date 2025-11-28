# Kubeflow ModelKit Component Examples

This directory contains runnable Kubeflow Pipeline examples that use the `push-modelkit` and `unpack-modelkit` components.

## House Prices Pipeline (`house-prices-pipeline.py`)

An end-to-end pipeline that:

- Trains an XGBoost regression model
- Writes model artifacts into a directory (model, code, docs)
- Packages them as a ModelKit and pushes to an OCI registry
- Optionally adds attestation metadata and cosign signing

### How to Run

```bash
pip install kfp==2.14.3 kfp-kubernetes==2.14.3
python house-prices-pipeline.py
```

Upload the generated `house-prices-with-modelkit.yaml` to the Kubeflow UI (or use the KFP SDK) to execute the pipeline.

### After It Runs

Use the `kit` CLI to pull, inspect, and unpack the resulting ModelKit.

For full component reference, integration patterns, secret requirements, and troubleshooting, see the main Kubeflow components README in this directory.
