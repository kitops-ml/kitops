---
title: KitOps Integrations - Compatible Tools & Registries
description: Discover all the tools and platforms that work with KitOps ModelKits, including OCI registries, MLOps tools, CI/CD platforms, cloud services, and model tracking systems.
keywords: kitops integrations, modelkit compatible tools, OCI registries ML, CI/CD for AI models, mlops platforms, kubernetes model deployment, open source AI packaging, model registry compatibility
---

# KitOps Compatible Tools

KitOps ModelKits integrate seamlessly with the tools your team already uses for model development, CI/CD, registry management, and production deployment.

Built on Open Container Initiative ([OCI](https://opencontainers.org/)) standards, KitOps works anywhere containers do — across cloud, on-prem, or local environments.

**⚡ KitOps is trusted by teams using Google, Amazon, Microsoft, GitLab, MLFlow, Weights & Biases, Jupyter, Hugging Face, Kubernetes, KServe, Red Hat OpenShift, and more!**

The KitOps community has built some great integrations:
- [Import directly from Hugging Face](../../cli/cli-reference.md#kit-import)
- [Pipelines and workflows](../cicd.md):
   - [Google Vertex](https://github.com/TheCoder2010-create/Building-ML-Pipelines-with-KitOps-and-Vertex-AI)
   - [ArgoCD](https://jozu.com/blog/deploying-ml-projects-with-argo-cd/)
   - [Red Hat OpenShift Pipelines](https://jozu.com/blog/how-to-turn-your-openshift-pipelines-into-an-mlops-pipeline)
   - [GitHub Actions](https://jozu.com/blog/automating-ml-pipeline-with-modelkits-github-actions/)
   - [Dagger](https://jozu.com/blog/building-an-mlops-pipeline-with-dagger-io-and-kitops/)
- Outputting ModelKits directly from [MLFlow](../mlflow.md)
- Deploying to any [Kubernetes distribution](../k8s-init-container.md)
- Working with ML in [KServe](../kserve.md)

## 🗄️ KitOps Compliant OCI Registries (A-Z)

The most fully-featured model registry integration for ModelKits is the [Jozu Hub](https://jozu.ml/), however, any container registry will work:
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
* Zed Registry

## 🤖 KitOps Compatible MLOps Tools (A-Z)

KitOps works with nearly all ML pipeline tools, AI model deployment tools, and open source MLOps tools:
* Amazon SageMaker
* Azure ML
* Clear ML
* Comet ML
* Databricks
* DataRobot
* Domino
* DvC
* Google Vertex
* Google Kubernetes Service (GKS)
* Google Container Platform (GCP)
* Hugging Face
* Jupyter notebooks
* Kubeflow
* Marimo
* Microsoft Azure
* MLFlow
* ModelScan
* Neptune.ai
* NVIDIA Triton and Run.ai
* OctoML
* Prefect
* Ray
* Red Hat InstructLab
* Red Hat OpenShift AI
* Seldon
* Tensorflow Hub
* Weights & Biases
* ZenML

## 🏭 KitOps Compatible Serving Platforms (A-Z)

KitOps works perfectly with any serving platform that accepts containers:
* Amazon Elastic Kubernetes Service (EKS)
* Amazon Elastic Compute Cloud (EC2)
* Amazon Fargate
* Amazon Lambda
* IBM Cloud
* Kubernetes
* Kserve
* Microsoft Azure Kubernetes Service (AKS)
* Microsoft Azure Cloud
* Red Hat OpenShift
* VMware

## 🛠️ KitOps Compatible Pipeline & Storage Tools (A-Z)

Thanks to its OCI-compatibility, KitOps works nearly any tool:
* Amazon S3
* Argo CD
* BitBucket Pipelines
* Circle CI
* Dagger
* Flux CI/CD
* Git
* Git LFS
* GitHub Actions
* GitLab Pipelines
* Google Vertex AI
* Jenkins CI/CD
* Kubeflow
* Spinnaker
* Tekton
* Travis CI

## 🤩 Community Feedback

If you've used KitOps with a product or project we've missed, please [open an issue](https://github.com/jozu-ai/kitops/issues/new/choose) in our GitHub repository.

For help please join our [Discord community](https://discord.gg/Tapeh8agYy).
