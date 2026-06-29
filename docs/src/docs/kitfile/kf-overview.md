---
title: Kitfile Format - Define AI/ML Projects for KitOps ModelKits
description: Learn how to create and use a Kitfile to securely package AI/ML projects for enterprise use. Define models, code, and datasets using a simple YAML format for versioned sharing and deployment.
keywords: kitfile format, kitops kitfile example, modelkit yaml, ai model manifest, ml model packaging yaml, versioned ai project, oci model definition, share ml project, kitops packaging
---

# Kitfile Format for ModelKits

The **Kitfile** is the blueprint for every AI project packaged with KitOps.

It’s a simple, YAML-based manifest that defines what goes into a ModelKit: models, datasets, code, prompts, agent skill files, MCP server configurations, and documentation.

ModelKits use Kitfiles to package and version your project so it can be shared, stored, or deployed from any OCI-compatible registry.

## Why Use a Kitfile?

A Kitfile helps you:
- Package your AI project with full traceability - whether it’s a model, an agent configuration, or both
- Make your work reproducible across environments
- Collaborate with DevOps, ML engineers, AI engineers, and app developers using a shared format
- Integrate into CI/CD pipelines and registries

## Kitfile Sections

Each Kitfile includes one or more of the following sections:

| Section       | Description | Example contents |
|---------------|-------------|------------------|
| `package`     | Project metadata (name, version, license, authors) | - |
| `model`       | The serialized model and framework info | Model weights, GGUF files |
| `datasets`    | Training, validation, or other datasets | CSV files, Parquet datasets |
| `code`        | Code, scripts, and server configurations | Jupyter notebooks, MCP server code and config files |
| `prompts`     | Prompt files and agent skill files | System prompts, SKILL.md, .cursorrules |
| `docs`        | Additional documentation | README, usage guides |
| `mcpServers`  | [MCP Bundle (`.mcpb`)](https://github.com/modelcontextprotocol/mcpb) archives for local MCP servers | Packaged MCP server bundles |

The `prompts` section is the natural home for agent skill files and prompt templates. The `code` section is where MCP server code and configuration files go. Use `mcpServers` to package pre-built [MCP Bundle (`.mcpb`)](https://github.com/modelcontextprotocol/mcpb) archives — these are stored verbatim so the bundle is byte-identical when unpacked. You can organize these however fits your project - the key is that everything gets versioned together.

You can extract the Kitfile from any existing ModelKit:

```sh
kit unpack [registry/repo:tag] --config -d .
```

## Minimal Kitfile Examples

The only required fields are:
- `manifestVersion`
- At least one of `code`, `model`, `datasets`, `docs`, `prompts`, or `mcpServers` section

A ModelKit for a dataset:
```yaml
manifestVersion: v1.0.0

datasets:
  - name: training data
    path: ./data/train.csv
  - name: validation data
    path: ./data/test.csv
```

A ModelKit for agent skills and prompts (no model required):
```yaml
manifestVersion: v1.0.0

package:
  name: customer-support-agent
  description: Skill files and prompts for the customer support agent
  version: 2.1.0

prompts:
  - path: ./prompts/system.prompt.md
    description: System prompt for the support agent
  - path: ./skills/escalation.skill.md
    description: Skill for escalating tickets to human agents
  - path: ./skills/refund-processing.skill.md
    description: Skill for processing refund requests
```

### Notes on Kitfile Behavior
- Kitfiles must use relative paths (not absolute)
- ModelKits can include only one model, but multiple datasets, code entries, or prompts
- You can currently only build ModelKits from local files, but support for remote sources (e.g. DVC, S3) is coming soon

## Example: Full Kitfile with Model

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

## Example: Agentic AI Kitfile

This Kitfile packages agent skill files, prompts, MCP server source code, and pre-built [MCP Bundle (`.mcpb`)](https://github.com/modelcontextprotocol/mcpb) archives together. No model is needed when the agent uses a hosted API (e.g., Claude, GPT). Use `code` for MCP server source and configuration files; use `mcpServers` for pre-built `.mcpb` bundles that are stored and restored byte-for-byte.

```yaml
manifestVersion: v1.0.0

package:
  name: sales-research-agent
  description: Agent configuration for automated sales research pipeline
  authors:
    - Platform Team
  version: 3.0.1

prompts:
  - path: ./prompts/system-prompt.md
    description: Core system prompt defining agent persona and constraints
  - path: ./prompts/research-prompt.md
    description: Prompt template for company research tasks
  - path: ./skills/linkedin-enrichment.skill.md
    description: Skill for enriching leads with LinkedIn data
  - path: ./skills/competitive-analysis.skill.md
    description: Skill for generating competitive analysis summaries

code:
  - path: ./mcp-servers/crm-connector
    description: MCP server for CRM data access
  - path: ./mcp-servers/web-scraper
    description: MCP server for web scraping with rate limiting

mcpServers:
  - name: filesystem
    path: ./bundles/filesystem.mcpb
    description: Pre-built MCP server bundle providing local filesystem access
  - name: web-search
    path: ./bundles/web-search.mcpb
    description: Pre-built MCP server bundle for web search capabilities

docs:
  - path: ./README.md
    description: Agent setup and deployment guide
```

## Learn More

➡️ Explore the [Kitfile structure](../format.md) in detail

---

**Questions or suggestions?** Drop an [issue in our GitHub repository](https://github.com/kitops-ml/kitops/issues) or join [our Discord server](https://discord.gg/Tapeh8agYy) to get support or share your feedback.

<!-- AGENT_MODIFIED: Human review required before merge -->
