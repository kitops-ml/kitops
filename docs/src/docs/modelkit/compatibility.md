---
description: Ensure compatibility of your ModelKits with various AI frameworks. Learn about supported models and environments.
---
# Compatible Tools

ModelKit packages can be pushed to any OCI 1.1-compliant registry, whether in the cloud, on-premises, or locally. This makes ModelKits easy to find because they're in the same place as the rest of your application's containers and artifacts. It also makes them easy to control since the registry already includes authentication and authorization.

ModelKits themselves use standards like JSON, YAML, and TAR files so whatever MLOps or DevOps tools you're using...they'll work with ModelKits.

If you've tried using Kit with your favorite tool and are having trouble, please [open an issue](https://github.com/kitops-ml/kitops/issues/new/choose) in our GitHub repository.

If you've used KitOps with a product or project we've missed, please open a pull request updating this file.

## Compliant OCI Registries

The most fully-featured repository for ModelKits is the [Jozu Hub](https://jozu.ml/), however, many users find it easiest to store their ModelKits in an existing enterprise container registry:

* Amazon Elastic Container Registry (ECR)
* Azure Container Registry
* Docker Hub
* GitHub Packages Container Registry
* GitLab Container Registry
* Google Artifact Registry
* Harbor
* IBM Cloud Container Registry
* JFrog Artifactory
* Jozu Hub
* Red Hat Quay.io
* Sonatype Nexus

## CI/CD & Pipline Tools

### Pre-Built Workflows

* Dagger: see [Kit modules for Dagger](https://github.com/kitops-ml/daggerverse) in the Daggerverse
* GitHub Actions: Kit CLI for [GitHub Actions](https://github.com/marketplace/actions/setup-kit-cli)

## Other Compatible Tools

* Amazon SageMaker
* Amazon Elastic Kubernetes Service (EKS)
* Amazon Elastic Compute Cloud (EC2)
* Amazon Fargate
* Amazon Lambda
* Amazon S3
* Argo CD
* Azure ML
* Azure Kubernetes Service (AKS)
* Azure Cloud
* Circle CI
* Clear ML
* Comet ML
* Databricks
* DataRobot
* Domino
* DvC
* Git
* Git LFS
* GitHub
* GitLab
* Google Vertex
* Google Kubernetes Service (GKS)
* Google Container Platform (GCP)
* Hugging Face
* IBM Cloud
* Jenkins CI/CD
* Jupyter notebooks
* Kubernetes
* Kserve
* Marimo
* MLFlow
* ModelScan
* Neptune.ai
* NVIDIA Triton and Run.ai
* OctoML
* Prefect
* Ray
* Red Hat InstructLab
* Red Hat OpenShift
* Red Hat OpenShift AI
* Seldon
* Tensorflow Hub
* VMware
* Weights & Biases
* ZenML
