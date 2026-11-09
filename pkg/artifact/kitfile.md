# Kitfile AI/ML Packaging Manifest Format Reference

The Kitfile manifest for AI/ML is a YAML file designed to encapsulate all the necessary information about the package, including code, datasets, model, and their metadata. This reference documentation outlines the structure and specifications of the manifest format.

## Overview

The manifest is structured into several key sections: `manifestVersion`, `package`, `code`, `datasets`, `docs`, `prompts`, `mcpServers`, `agent`, and `model`. Each section serves a specific purpose in describing the AI/ML package components and requirements.

### `manifestVersion`

- **Description**: Specifies the manifest format version.
- **Type**: String
- **Example**: `1.0.0`

### `package`

This section provides general information about the AI/ML project.

#### `name`

- **Description**: The name of the AI/ML project.
- **Type**: String

#### `version`

- **Description**: The current version of the project.
- **Type**: String
- **Example**: `1.2.3`

#### `description`

- **Description**: A brief overview of the project's purpose and capabilities.
- **Type**: String

#### `authors`

- **Description**: A list of individuals or entities that have contributed to the project.
- **Type**: Array of Strings


### `code`

- **Description**: Information about the source code.
- **Type**: Object Array
  - `path`: Location of the source code files or directory relative to the context
  - `description`: Description of what the code does.
  - `license`: SPDX license identifier for the code.

### `datasets`

- **Description**: Information about the datasets used.
- **Type**: Object Array
  - `name`: Name of the dataset.
  - `path`: Location of the dataset file or directory relative to the context.
  - `description`: Overview of the dataset.
  - `license`: SPDX license identifier for the dataset.

### `docs`

- **Description**: Information about included documentation for the model
- **Type**: Object Array
  - `description`: Description of the documentation
  - `path`: Location of the documentation relative to the context

### `prompts`

- **Description**: Information about prompt files intended for use with LLMs
- **Type**: Object Array
  - `path`: Location of the prompt file or directory relative to the context
  - `description`: Description of the prompt

### `mcpServers`

- **Description**: Information about MCP servers packaged as MCP Bundle (`.mcpb`) archives. Each entry's path must point to a single `.mcpb` file: a ZIP archive containing a local MCP server and a `manifest.json` at its root. Each bundle is stored verbatim as its own raw, uncompressed layer (media type `application/vnd.kitops.modelkit.mcpb.v1.raw`), so the layer digest covers the exact bundle bytes and unpacking yields a byte-identical archive. The bundle's `manifest.json` remains the source of truth for the server's configuration; none of its fields are duplicated in the Kitfile.
- **Type**: Object Array
  - `name`: Name for the MCP server. Required, and must be unique within `mcpServers`.
  - `path`: Location of the `.mcpb` file relative to the context.
  - `description`: Description of the MCP server.

### `agent`

- **Description**: The Agent Definition of an agent ModelKit, in its `agent:` block form. The block carries the `spec` of an Agent Definition as defined in the Agent Definition Specification (`pkg/artifact/agent-definition.md`); `apiVersion`, `kind` and `metadata` are omitted, and `metadata.name` is the package's `name`. The block is validated when the Kitfile is loaded and is stored in the ModelKit's config. When a ModelKit carries both an `agent` block and a definition file in its `code` layer, the block is the definition and the file is ignored (Agent Definition Specification §4.3).
- **Type**: Object
  - `framework`: The framework the agent runs under. Required unless `base` is given.
  - `base`: A base agent to compose from.
    - `modelkit`: Reference to the base agent ModelKit.
  - `needs`: List of runtime values the agent consumes (Agent Definition Specification §6).
    - `name`: Required; unique within `needs`, and a valid environment variable name.
    - `target`: `env` (default) or `file`.
    - `mountPath`: Path the value is written to; required for, and only with, `target: file`.
    - `optional`: Whether resolution may fail without aborting. Default `false`.
    - `from`: Where the value comes from; exactly one of `envFrom`, `file`, `text`, `keychain`, `vault`, `oidc`. Omitted for a bare need.
  - `includes`: List of modules the resolver provides (Agent Definition Specification §7 and §8).
    - `name`: Required; unique within `includes`.
    - `type`: `llm`, `mcp`, `policy` or `skill`.
    - `source`: Where the module comes from; exactly one of `model`, `modelkit` or `endpoint`, as the type permits. A `modelkit` beginning with `./` names an entry of this Kitfile by name: an `mcpServers` entry for an `mcp` module, a `prompts` entry for a `skill` module, the `model` for an `llm` module.
    - `transport`: `http` or `sse`; required for, and only with, a remote `mcp` module.
    - `headers`: Headers sent to a remote `mcp` module's endpoint. Values may reference needs as `${needs.NAME}`.
    - `config`: Non-secret settings passed to the module.

### `model`

- **Description**: Details of the trained models included in the package.
- **Type**: Object
  - `name`: Name of the model
  - `path`: Location of the model file or directory relative to the context
  - `framework`: AI/ML framework
  - `version`: Version of the model
  - `description`: Overview of the model
  - `license`: SPDX license identifier for the dataset.
  - `parts`: List of related files for the model (e.g. LoRA weights)
    - `name`: Identifier for the part
    - `path`: Location of the file or a directory relative to the context
    - `type`: The type of the part (e.g. LoRA weights)
  - `parameters`: An arbitrary section of yaml that can be used to store any additional data that may be relevant to the current model, with a few caveats. Only a json-compatible subset of yaml is supported. Strings will be serialized without flow parameters. Numbers will be converted to decimal representations (0xFF -> 255, 1.2e+3 -> 1200). Maps will be sorted alphabetically by key.


## Example

```yaml
manifestVersion: 1.0
package:
  name: AIProjectName
  version: 1.2.3
  description: >-
    A brief description of the AI/ML project.
  authors: [Author Name, Contributor Name]
code:
  - path: src/
    description: Source code for the AI models.
    license: Apache-2.0
datasets:
  - name: DatasetName
    path: data/dataset.csv
    description: Description of the dataset.
    license: CC-BY-4.0
prompts:
  - path: system.prompt.md
    description: System prompt for the LLM.
mcpServers:
  - name: filesystem
    path: filesystem.mcpb
    description: MCP server providing filesystem access.
agent:
  framework: claude-code
  needs:
    - name: GITHUB_TOKEN
  includes:
    - name: filesystem
      type: mcp
      source:
        modelkit: ./filesystem
    - name: github
      type: mcp
      source:
        endpoint: https://api.githubcopilot.com/mcp/
      transport: http
      headers:
        Authorization: "Bearer ${needs.GITHUB_TOKEN}"
model:
    name: ModelName
    path: models/model.h5
    framework: TensorFlow
    version: 1.0
    description: Model description.
    license: Apache-2.0
```
