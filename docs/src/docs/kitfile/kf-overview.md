---
title: Kitfile Format - Define AI/ML Projects for KitOps ModelKits
description: Learn how to create and use a Kitfile to securely package AI/ML projects for enterprise use. Define models, code, and datasets using a simple YAML format for versioned sharing and deployment.
keywords: kitfile format, kitops kitfile example, modelkit yaml, ai model manifest, ml model packaging yaml, versioned ai project, oci model definition, share ml project, kitops packaging
---

# Kitfile Format for ModelKits

The **Kitfile** is the blueprint for every AI/ML project packaged with KitOps.

It’s a simple, YAML-based manifest that defines what goes into a ModelKit: models, datasets, notebooks, documentation, and metadata.

ModelKits use Kitfiles to package and version your project so it can be shared, stored, or deployed from any OCI-compatible registry.

## ✨ Why Use a Kitfile?

A Kitfile helps you:
- Package your AI/ML project with full traceability
- Make your work reproducible across environments
- Collaborate with DevOps, ML engineers, and app developers using a shared format
- Integrate seamlessly into CI/CD pipelines and registries

## 📚 Kitfile Sections

Each Kitfile includes one or more of the following sections:

| Section       | Description |
|---------------|-------------|
| `package`     | Project metadata (name, version, license, authors) |
| `code`        | Paths to notebooks or scripts |
| `model`       | The serialized model and framework info |
| `datasets`    | Training, validation, or other datasets |
| `docs`        | Additional documentation to include |
| `prompts`     | Prompt files for use with LLMs |

You can extract the Kitfile from any existing ModelKit:

```sh
kit unpack [registry/repo:tag] --config -d .
```

## ✅ Minimal Kitfile Example

The only required fields are:
- `manifestVersion`
- At least one of `code`, `model`, `datasets`, `docs`, or `prompts` section

Example:
```yaml
manifestVersion: v1.0.0

datasets:
  - name: training data
    path: ./data/train.csv
  - name: validation data
    path: ./data/test.csv
```

### Notes on Kitfile Behavior
- Kitfiles must use relative paths (not absolute)
- ModelKits can include only one model, but multiple datasets or code bases
- You can currently only build ModelKits from local files — but support for remote sources (e.g. DVC, S3) is coming soon

## 🧱 Example: Full Kitfile

```yaml
manifestVersion: v1.0.0

package:
  authors:
    - Jozu
  description: Updated model to analyze flight trait and passenger satisfaction data
  license: Apache-2.0
  name: FlightSatML

code:
  - description: Jupyter notebook with model training code in Python
    path: ./notebooks

model:
  description: Flight satisfaction and trait analysis model using Scikit-learn
  framework: Scikit-learn
  license: Apache-2.0
  name: joblib Model
  path: ./models/scikit_class_model_v2.joblib
  version: 1.0.0

datasets:
  - name: training data
    description: Flight traits and traveller satisfaction training data (tabular)
    path: ./data/train.csv
  - name: validation data
    description: validation data (tabular)
    path: ./data/test.csv

prompts:
  - path: system.prompt.md
    description: System prompt for model inference
```

## Learn More

➡️ Explore the [Kitfile structure](../format.md) in detail

---

**Questions or suggestions?** Drop an [issue in our GitHub repository](https://github.com/kitops-ml/kitops/issues) or join [our Discord server](https://discord.gg/Tapeh8agYy) to get support or share your feedback.
