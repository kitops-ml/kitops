<img width="1270" alt="KitOps" src="https://github.com/kitops-ml/kitops/assets/10517533/41295471-fe49-4011-adf6-a215f29890c2" id="top">

[English](./README.md) | 简体中文

## KitOps：基于标准的 AI/ML 项目打包与版本管理工具

[![LICENSE](https://img.shields.io/badge/License-Apache%202.0-yellow.svg)](./LICENSE)
[![Discord](https://img.shields.io/discord/1098133460310294528?logo=Discord)](https://discord.gg/Tapeh8agYy)
[![Twitter](https://img.shields.io/twitter/url/http/shields.io.svg?style=social&label=Twitter)](https://twitter.com/kit_ops)

## 目录

- [什么是 KitOps？](#什么是-kitops)
- [15 分钟试用 KitOps](#15-分钟试用-kitops)
- [团队如何使用 KitOps](#团队如何使用-kitops)
- [KitOps 架构](#kitops-架构)
- [安全与合规](#安全与合规)
- [按角色了解 KitOps](#按角色了解-kitops)
- [集成](#集成)
- [加入 KitOps 社区](#加入-kitops-社区)

## 什么是 KitOps？

KitOps 是一个 CNCF（云原生计算基金会）开源工具，用于对 AI/ML 项目进行打包、版本管理和安全共享。

KitOps 基于与容器技术同源的[OCI（Open Container Initiative，开放容器倡议）](https://opencontainers.org/)标准构建，将模型开发或生产所需的一切打包成一个带版本标识、分层存储的制品（artifact），保存在你现有的容器镜像仓库中。它可以与你所有的 AI/ML、CI/CD 和 DevOps 工具集成。

作为 Kubernetes AI/ML 技术栈的一部分，对于需要私有化部署 AI 模型和智能体（agent）、注重安全的企业、政府和云运营商而言，KitOps 是进行 AI 资产打包、版本管理及治理的首选方案。

### KitOps 与 CNCF

KitOps 由[CNCF](https://www.cncf.io/)（即管理 Kubernetes、OpenTelemetry 和 Prometheus 的同一组织）托管。[本视频概要介绍了 KitOps 在 CNCF 中的情况](https://youtu.be/iK9mnU0prRU?feature=shared)。

KitOps 同时也是厂商中立的 AI/ML 交换格式——[CNCF ModelPack 规范](https://github.com/modelpack/model-spec)——的企业级实现。Kit CLI 同时透明支持 ModelKit 和 ModelPack 两种格式。ModelPack 的贡献公司包括 Red Hat、PayPal、蚂蚁集团和字节跳动。

[![访问官方网站](<https://img.shields.io/badge/-%E8%AE%BF%E9%97%AE%E5%AE%98%E6%96%B9%E7%BD%91%E7%AB%99%20%E2%86%92-rgb(255,175,82)?style=for-the-badge>)](https://kitops.org?utm_source=github&utm_medium=kitops-readme)

[![快速上手](<https://img.shields.io/badge/-KitOps%20%E5%BF%AB%E9%80%9F%E4%B8%8A%E6%89%8B%20%E2%86%92-rgb(122,140,225)?style=for-the-badge>)](https://kitops.org/docs/get-started/?utm_source=github&utm_medium=kitops-readme)

## 15 分钟试用 KitOps

1. **安装 CLI**：[适用于 MacOS、Windows 和 Linux](https://kitops.org/docs/cli/installation/)。
2. **打包你的第一个 ModelKit**，任选一种方式：
   * **从 HuggingFace 导入**：通过[HuggingFace 导入功能](https://kitops.org/docs/hf-import/)将模型从 HuggingFace 直接拉取到 ModelKit 中。
   * 进入你的项目目录，运行 `kit init .` 自动生成 Kitfile，然后按照[入门指南](https://kitops.org/docs/get-started/)进行打包、推送和拉取。
3. **推送到你的镜像仓库**：使用 `kit push`，即可开始将你现有的企业镜像仓库用作 AI 智能体、模型和 MCP 服务器的安全精选仓库。
4. **探索预构建的 ModelKit**：通过[快速入门](https://jozu.ml/organization/jozu-quickstarts)试用 LLM、计算机视觉模型等。

如果你更喜欢从源码构建，请按照[这些步骤](https://kitops.org/docs/cli/installation/#build-sources)从我们的仓库获取最新版本。

## 团队如何使用 KitOps

### 第 1 级：生产交付

大多数团队最初是在模型或智能体准备进入预发布/生产环境时，使用 KitOps 对其进行版本管理。ModelKit 作为不可变、自包含的包，简化了 CI/CD 部署、制品签名、AI SBOM 创建以及部署/回滚。这可以防止未知的 AI 工作负载进入生产环境，并保持数据集、模型权重和配置的同步且可跟踪。

了解更多：[CI/CD 集成](https://kitops.org/docs/integrations/cicd/)

### 第 2 级：模型安全

受监管行业的团队使用 KitOps 在模型进入生产环境之前进行扫描和把关。构建 ModelKit，用 Cosign 为其签名，运行安全扫描，将报告作为签名证明（attestation）附加，并只允许带有证明的 ModelKit 继续流转。KitOps 在你已使用的任何工具之上，额外提供了一层安全与审计能力。

了解更多：[保护 ModelKit 的安全](https://kitops.org/docs/security/)

### 第 3 级：全生命周期版本管理

成熟的团队将 KitOps 的使用扩展到开发阶段。每个里程碑（新数据集、调优检查点、重训练事件）都保存为一个带版本标识的 ModelKit。用一套标准体系（OCI）管理每个模型版本，实现防篡改、内容寻址的存储。

了解更多：[KitOps 的应用场景](https://kitops.org/docs/use-cases/)

## KitOps 架构

### ModelKit

KitOps 将你的项目打包成[ModelKit](https://kitops.org/docs/modelkit/intro/)——一个自包含、不可变的捆绑包，其中包含复现、测试或部署你的 AI/ML 模型所需的一切。

ModelKit 可以包含智能体、模型权重、MCP 服务器、数据集、提示词、实验运行结果与超参数、元数据、环境配置、代码等。

ModelKit 具有以下特性：
* **防篡改**——每个组件都受 SHA-256 摘要保护，确保一致性与可追溯性
* **可签名**——完整兼容[Cosign](https://github.com/sigstore/cosign)，支持加密验证
* **兼容性强**——在所有主流 OCI 容器镜像仓库中原生存储与检索
* **可按需解包**——只拉取你需要的层（只要模型、只要数据集等）

KitOps 还可以使用 CNCF model-spec 格式创建符合[ModelPack](https://github.com/modelpack/model-spec)规范的包。这两种格式都是厂商中立的标准，Kit 命令（`pull`、`push`、`unpack`、`inspect`、`list`）对两者都透明兼容。

> ModelKit 将 AI 制品提升为一等公民级别的受治理资产，就像应用代码一样。

### Kitfile

[Kitfile](https://kitops.org/docs/kitfile/kf-overview/)定义 ModelKit 中每个制品的存放位置。你可以通过 `kit init` 自动生成它。

### Kit CLI

[Kit CLI](https://kitops.org/docs/cli/cli-reference/)让你可以创建、管理、运行和部署 ModelKit。主要命令包括：

* `kit pack`——将你的项目打包成 ModelKit（添加 `--use-model-pack` 可使用 ModelPack 格式）
* `kit unpack`——从 ModelKit 中提取全部或指定层
* `kit push` / `kit pull`——通过任意 OCI 镜像仓库共享 ModelKit
* `kit init`——从现有项目目录自动生成 Kitfile
* `kit diff`——比较两个 ModelKit 之间的差异
* `kit list`——列出可用的 ModelKit 和 ModelPack
* `kit inspect`——无需解包即可查看 ModelKit 的内容

### PyKitOps Python SDK

[PyKitOps 库](https://kitops.org/docs/pykitops/)让数据科学家可以直接在 Python 中操作 ModelKit。使用它进行打包、推送、拉取和查看 ModelKit，无需离开你最常用工具的工作流。

### 观看 KitOps 实际操作

[![KitOps 视频](https://img.youtube.com/vi/j2qjHf2HzSQ/hqdefault.jpg)](https://www.youtube.com/watch?v=j2qjHf2HzSQ)

这段视频展示了 KitOps 如何通过 ModelKit 简化数据科学家、开发者和 SRE 之间的协作。

## 安全与合规

KitOps 为需要建立和维护 AI/ML 资产保管链与来源追溯（provenance）的组织提供制品和项目元数据：

* **不可变摘要**——每个 ModelKit 组件都经过 SHA-256 哈希。在拉取或获取制品时，任何文件改动都会通过 OCI 摘要校验被发现，被篡改的制品将被拒绝。
* **加密签名**——使用[Cosign](https://github.com/sigstore/cosign)为 ModelKit 签名（基于密钥，或通过 OIDC 无密钥签名）。未签名或被篡改的 ModelKit 可以在 CI/CD 中被拦截。
* **AI 物料清单（AI Bill of Materials）**——ModelKit 提供所有组件（模型权重、数据集、代码、配置）的结构化清单，并进行版本追踪，是 AI SBOM 的基础。
* **透明日志**——与[Rekor](https://github.com/sigstore/rekor)结合，获得只追加（append-only）的签名记录。
* **可审计的溯源记录**——从实验到预发布再到生产的完整版本历史，保存在你的 OCI 镜像仓库中。

这些特性使 ModelKit 适用于对制品完整性、来源验证和审计追踪有要求的合规框架，包括欧盟《AI 法案》（EU AI Act）、NIST AI RMF、ISO 42001 等类似的监管要求。

了解更多：[保护你的模型供应链](https://kitops.org/docs/security/)

KitOps 还被用于**[Jozu Hub](https://jozu.com/)**，后者增加了集中式策略管理、五层安全扫描、签名证明和防篡改审计日志。Jozu Hub 部署在你的防火墙之后，与你现有的 OCI 镜像仓库协同工作，支持私有云、数据中心或物理隔离（air-gapped）环境。

## 按角色了解 KitOps

### DevOps 与平台工程师

* 在现有 CI/CD 流水线中使用 ModelKit，支持[GitHub Actions、Dagger 等系统](https://kitops.org/docs/integrations/cicd/)
* 在你当前的容器镜像仓库中存储和管理模型
* 使用[init container](https://kitops.org/docs/integrations/k8s-init-container/)或[KServe](https://kitops.org/docs/integrations/kserve/)部署到 Kubernetes
* 为安全的 AI/ML 部署构建黄金路径（golden path）

### 数据科学家

* 使用 `kit pack` 或[PyKitOps SDK](https://kitops.org/docs/pykitops/)打包数据集和模型，无需操心基础设施
* 将模型[从 HuggingFace 导入](https://kitops.org/docs/hf-import/)受治理的 ModelKit
* 通过[MLFlow 集成](https://kitops.org/docs/integrations/mlflow/)追踪实验

### 开发者

* 像使用普通依赖一样，通过标准工具和 API 使用 AI/ML 模型
* 只拉取你需要的层（模型、数据集、代码），无需下载整个包
* 与[Kubeflow Pipelines](https://kitops.org/docs/integrations/kubeflow/)及其他 ML 工具集成

## 集成

KitOps 可以与你已经在使用的工具协同工作：

* [MLFlow](https://kitops.org/docs/integrations/mlflow/)——将实验运行打包为带版本标识的 ModelKit
* [CI/CD](https://kitops.org/docs/integrations/cicd/)——GitHub Actions、Dagger 及其他流水线工具
* [Kubernetes initContainer](https://kitops.org/docs/integrations/k8s-init-container/)——将 ModelKit 作为 init container 解包
* [KServe](https://kitops.org/docs/integrations/kserve/)——直接从 ModelKit 提供模型服务
* [Kubeflow Pipelines](https://kitops.org/docs/integrations/kubeflow/)——在 ML 流水线步骤中使用 ModelKit
* [ModelPack](https://kitops.org/docs/integrations/modelpack/)——CNCF 厂商中立打包格式
* [KubeStellar Console](https://console.kubestellar.io/missions/install-kitops)——使用 KubeStellar missions 跨集群部署和管理 KitOps（包括引导式安装、预检查、验证和故障排查）

查看完整的[集成列表](https://kitops.org/docs/integrations/integrations/)。

## 加入 KitOps 社区

如需支持、发布更新和一般性讨论，请加入[KitOps Discord](https://discord.gg/Tapeh8agYy)。关注[KitOps 的 X 账号](https://twitter.com/Kit_Ops)获取每日更新。

如需帮助，可以通过多种方式联系我们的社区和[维护者](./MAINTAINERS.md)，详见[支持文档](./SUPPORT.md)。

### 加入 KitOps 贡献者行列

我们热爱 KitOps 社区和贡献者。想了解更多贡献方式（你不需要会写代码）以及如何开始，请参阅[贡献者指南](./CONTRIBUTING.md)。贡献前请阅读我们的[治理文档](./GOVERNANCE.md)和[行为准则](./CODE-OF-CONDUCT.md)。

### 报告问题与提出功能建议

你的见解有助于 KitOps 作为 AI/ML 开放标准不断演进。我们非常重视来自社区用户的问题和功能请求。要分享你的想法，请前往 **Issues** 标签页并点击 **New Issue** 按钮。

### KitOps 社区电话会议（双周一次）

**每周三 13:30 - 14:00（America/Toronto 时区）**
- [Google Meet](https://meet.google.com/zfq-uprp-csd)
- +1 647-736-3184（PIN：144 931 404#）
- [更多号码](https://tel.meet/zfq-uprp-csd?pin=1283456375953)

### 以尊重为基石的社区

在 KitOps，包容、共情和责任感是我们的核心。请阅读我们的[行为准则](./CODE-OF-CONDUCT.md)，了解指导我们社区的价值观。
---

<div align="center" style="align-items: center;">
        <a href="#top">
            <img src="https://img.shields.io/badge/%E8%BF%94%E5%9B%9E%E9%A1%B6%E9%83%A8-black?style=for-the-badge&logo=github&logoColor=white" alt="返回顶部">
        </a>
</div>
<!-- AGENT_MODIFIED: Human review required before merge -->
