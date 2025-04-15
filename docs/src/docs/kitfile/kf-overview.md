---
description: Get an overview of the Kitfile format. Understand how Kitfiles define your ModelKits and ensure consistent packaging.
---
# Kitfile: Your AI/ML Project Blueprint

## What is a Kitfile?

At the core of every AI/ML project managed by KitOps lies the Kitfile, a YAML-based manifest designed to streamline the encapsulation and sharing of project artifacts. From code and datasets to models and their metadata, the Kitfile serves as a comprehensive blueprint for your project, ensuring every component is meticulously organized and easily accessible.

## Structured for Clarity

Crafted with simplicity and efficiency in mind, the Kitfile organizes project details into distinct sections:

**Project Metadata:** Offers a snapshot of your project, including its name, version, description, and authors, laying the foundation for collaboration and recognition.

**Code:** Details about the source code powering your AI/ML models, complete with licensing information to uphold software best practices.

**Datasets:** Descriptions and paths to datasets, highlighting preprocessing steps and licenses, to ensure reproducibility and ethical use of data.

**Model Specifications:** Insights into the models themselves, including framework details, training parameters, and validation metrics, to foster understanding and further development.

**Documentation:** Conveniently separated documentation files, to make getting started faster.

## Designed for Collaboration

By encapsulating the essence of your AI/ML project into a singular, version-controlled document, the Kitfile not only simplifies the packaging process but also enhances collaborative efforts. Whether you're sharing projects within your team or with the global AI/ML community, the Kitfile ensures that every artifact, from datasets to models, is accurately represented and easily accessible.

Embrace the Kitfile in your AI/ML projects to harness the power of structured packaging, efficient collaboration, and seamless artifact management. As the backbone of the KitOps ecosystem, the Kitfile is your first step towards simplifying AI/ML project management and achieving greater innovation.

## Kitfile Structure

The Kitfile defines the contents of your ModelKit. It is written in YAML and stored with the ModelKit. You can extract the Kitfile from any ModelKit with the Kit CLI:

```sh
kit unpack [registry/repo:tag] --config -d .
```

There are four main parts to a Kitfile:
1. ModelKit metadata in the `package` section
1. Path to the Jupyter notebook folder in the `code` section
1. Path to the serialized model in the `model` section
1. Path to the datasets in the `datasets` section (you can have multiple datasets in the same page)
1. Paths to documentation in the `docs` section

Here's an example Kitfile:

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
- description: Flight traits and traveller satisfaction training data (tabular)
  name: training data
  path: ./data/train.csv
- description: validation data (tabular)
  name: validation data
  path: ./data/test.csv
```

The only mandatory parts of the Kitfile are:
* `manifestVersion`
* At least one of `code`, `model`, `docs` or `datasets` sections

A ModelKit can only contain one model, but multiple datasets or code bases are allowed. Also note that you can only use relative paths (no absolute paths) in your Kitfile. Right now you can only build ModelKits from files on your local system...but don't worry we're already working towards allowing you to reference remote files. For example, building a ModelKit from a local notebook and model, but a dataset hosted on DvC, S3, or anywhere else.

So a minimal ModelKit for distributing a pair of datasets might look like this:
```yaml
manifestVersion: v1.0.0

datasets:
- name: training data
  path: ./data/train.csv
- description: validation data (tabular)
  name: validation data
  path: ./data/test.csv
```

**Questions or suggestions?** Drop an [issue in our GitHub repository](https://github.com/kitops-ml/kitops/issues) or join [our Discord server](https://discord.gg/Tapeh8agYy) to get support or share your feedback.