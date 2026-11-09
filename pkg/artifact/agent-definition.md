# Agent Definition Specification

## 1. Overview

An Agent Definition is a YAML document that declares an AI agent as a composable, governed unit: its framework, the runtime values it consumes, and the modules (LLM providers, MCP servers, policies, skills) the resolver must configure to run it.

An Agent Definition travels in a ModelKit in one of two forms (section 4): as an `agent:` block in the Kitfile, or as a standalone file in the `code:` layer. Either way it is packaged alongside the prompts, skills, model weights, and datasets the agent uses. The resolver pulls the ModelKit, locates the definition, and acts on its `spec:` block.

Agents compose through `spec.base` references, enabling restricted variants, capability extensions, and environment-specific configurations without duplicating definitions.

An **agent resolver** is any runtime that can read an Agent Definition, resolve its references, and produce a running agent. Resolvers differ in what they guarantee, not in how they read the definition: a developer tool that runs the agent locally and a governed runtime that adds a gateway, policy enforcement and audit read the same document. Section 12 lists what a conforming resolver must do.

---

## 2. Conventions

The key words "MUST", "MUST NOT", "SHOULD", "SHOULD NOT", and "MAY" in this document are to be interpreted as described in [RFC 2119](https://www.rfc-editor.org/info/rfc2119).

---

## 3. Design Principles

This section is informative. It records the decisions the normative sections follow, so that a case they do not anticipate is settled the same way.

1. **Modules are instructions to the resolver.** Every module requires the resolver to act: configure an LLM endpoint or serve a model, launch or connect an MCP server, enforce a policy, or mount a skill where the framework discovers it. A field the resolver would not act on does not belong in a module.

2. **Needs deliver values to the runtime, not to module fields.** A need produces an environment variable or a file, and the agent and its local MCP servers read it from there. The only fields that reference a need are the ones the gateway applies on the agent's behalf (section 6.5), and the agent never receives those values.

3. **Content layers are packaging concerns; modules say what the resolver acts on.** Content is packaged in ModelKit layers using standard Kitfile semantics, and the definition does not restate how. A module names the entry the resolver acts on — a bundle it launches, a skill it places, a model it serves — and content no module names is mounted alongside the agent for the framework to discover by convention.

4. **Single authority per decision.** Each concern has exactly one place where it is defined. Tool authorization is a policy decision and inference parameters are a framework decision; the definition duplicates or overrides neither.

---

## 4. Packaging

An Agent Definition is carried in a ModelKit in one of two forms. Both are valid, carry the same content, and resolve identically.

| | `agent:` block | Definition file |
|---|----------------|-----------------|
| Where | the Kitfile's `agent` key | `agent-definition.yaml` or `.agent-definition.yaml`, packed as a file in the `code:` layer |
| Carries | the contents of `spec` | the full document (section 5) |
| `apiVersion`, `kind` | omitted | required |
| `metadata.name` | omitted; the Kitfile's `package.name` supplies it | required |

A resolver MUST accept both forms and MUST normalize them into one internal shape before any consumer observes the resolved manifest. Nothing downstream of loading may depend on which form was used.

### 4.1 The `agent:` block

```yaml
manifestVersion: 1.0.0
package:
  name: code-review-agent
  version: 1.0.0
prompts:
  - path: CLAUDE.md
agent:
  framework: claude-code
  needs:
    - name: GITHUB_TOKEN
  includes:
    - name: github
      type: mcp
      source:
        endpoint: https://api.githubcopilot.com/mcp/
      transport: http
      headers:
        Authorization: "Bearer ${needs.GITHUB_TOKEN}"
```

The block is the `spec` of a definition document (section 5.1) and nothing else. `apiVersion` and `kind` are implied by the Kitfile, and `metadata.name` is the Kitfile's `package.name`.

The block is the form a tool writes when it derives a definition from a project: it adds no file to the project, and a Kitfile is the one manifest every ModelKit has.

### 4.2 The definition file

A definition file is a self-contained YAML document:

```yaml
apiVersion: kitops.org/v1
kind: AgentDefinition
metadata:
  name: my-agent
spec:
  framework: claude-code
  needs:
    - ...
  includes:
    - ...
```

It is packaged inside a ModelKit as a file in the Kitfile's `code:` layer. Representative directory layout:

```
my-agent/
├── Kitfile
├── agent-definition.yaml
├── prompts/
│   └── system.md
├── skills/
│   └── my-skill/
```

Representative Kitfile:

```yaml
manifestVersion: 1.0.0
package:
  name: my-agent
  version: 1.0.0
code:
  - path: ./agent-definition.yaml
prompts:
  - path: ./prompts/system.md
  - name: my-skill
    path: ./skills/my-skill/
model:
  name: weights
  path: ./models/weights
```

The file has two accepted names. `agent-definition.yaml` is the conventional one; `.agent-definition.yaml` serves projects that keep tool configuration out of the directory listing. They are two spellings of one file, not two files: a ModelKit carries at most one of them.

A ModelKit MAY contain at most one document with `kind: AgentDefinition` across all files in the code layer, and that includes a ModelKit carrying both spellings of the file. A YAML file MAY contain multiple documents separated by `---`; the resolver MUST locate the first document with `kind: AgentDefinition` and consume it. The match is on the document's `kind`, not on the filename: a file by either name holding something else is not a definition.

### 4.3 Precedence

When a Kitfile has an `agent:` block, that block is the Agent Definition, and a resolver MUST NOT consult any definition file in the code layer. Without a block, the definition is the single `kind: AgentDefinition` document in the code layer. A resolver that finds a block and a definition file together SHOULD report that the file was shadowed, naming it, so that a developer who edits the file and sees no effect can tell why.

The precedence is deliberate. The Kitfile is the one manifest every ModelKit has and every consumer reads first, so a block there is always found without fetching a layer; ranking it first keeps resolution decidable from the config alone in the common case. It also means a definition file can be overridden without being edited: a team can repackage a published agent with a Kitfile block in front of its definition, leaving the original file intact in the layer for reference.

### 4.4 In the packed ModelKit

Packing turns each part of an agent's Kitfile into a part of the ModelKit:

| In the Kitfile | In the packed ModelKit |
|----------------|------------------------|
| `agent:` block, KitOps format | inside the config blob, which is the Kitfile itself |
| `agent:` block, ModelPack format | inside the `ml.kitops.modelkit.kitfile` manifest annotation, where KitOps keeps the whole Kitfile when the config is a ModelPack config |
| Definition file | its own code layer (`application/vnd.kitops.modelkit.code.v1.tar`), with `org.cncf.model.filepath` naming the file and `ml.kitops.modelkit.layer-subtype: agent` marking the layer |
| Prompts, rules, local skills | prompt layers: the code media type with `ml.kitops.modelkit.layer-subtype: prompt` |
| `.mcpb` bundles | MCPB layers (`application/vnd.kitops.modelkit.mcpb.v1.raw`), one per `mcpServers` entry, per the Kitfile reference |

An **agent ModelKit** carries the definition and the agent's content — prompts, rules, local skills, bundles — and not the repository the agent works on. The repository is the agent's subject: a resolver runs the agent in a working directory of the operator's choosing, and an artifact that also carried the code it was authored in could not be run anywhere else.

A packer MUST set the annotation `ml.kitops.modelkit.layer-subtype: agent` on the code layer that holds a definition file, and MUST NOT set it on any other layer. The annotation marks the layer as holding the definition and nothing more: it carries no field of the definition, since the effective framework and modules are known only after the `base` chain resolves. A layer so marked holds exactly one file, the definition, so that it can be fetched on its own and a consumer can tell from the manifest alone that a ModelKit carries a definition file. Locating the definition follows section 4.3: the resolver reads the config first, and only when it holds no `agent:` block does it look for the marked layer. A ModelKit packed before this annotation existed carries its definition file in an unmarked code layer, so a resolver that finds no marked layer MUST fall back to scanning the code layers for the document; the annotation's absence never means "no definition".

---

## 5. Document Structure

An Agent Definition is a single YAML document with the following root fields:

| Field | Required | Description |
|-------|----------|-------------|
| `apiVersion` | Yes | `kitops.org/v1`. A resolver MAY also accept `agentguard.jozu.dev/v1`, the value this specification's predecessor used, as an alias with the same meaning, so that definitions written under it keep resolving. |
| `kind` | Yes | `AgentDefinition` |
| `metadata.name` | Yes | Identifier for the agent within its publishing namespace |
| `spec` | Yes | The agent's configuration |

In the `agent:` block form (section 4.1) only the contents of `spec` appear: `apiVersion` and `kind` are implied by the Kitfile, and `metadata.name` is the Kitfile's `package.name`.

### 5.1 spec fields

| Field | Required | Description |
|-------|----------|-------------|
| `framework` | Yes (root agents) | Identifier for the agent framework (e.g., `claude-code`, `pydantic-ai`). Inherited from base in composed agents. |
| `base` | No | Reference to a base agent ModelKit for composition (section 10). |
| `needs` | No | List of runtime value declarations (see section 6). |
| `includes` | No | List of module declarations (see sections 7 and 8). |
| `egress` | No | List of hosts the agent runtime may reach, some with headers the gateway sets on their requests (see section 9). |

A root agent (no `spec.base`) MUST declare `spec.framework`. A composed agent (with `spec.base`) inherits `framework` from its base and MUST NOT override it. A composed agent MAY declare `spec.framework` as a validation assertion — the resolver MUST reject the definition if it does not match the base.

Every field of a definition is one this specification defines. A resolver MUST reject a definition carrying a field this specification does not define, at any level other than an opaque `config` map (section 8). A field a resolver needs that the definition lacks belongs in a revision of this specification, not in an extension one resolver understands and another silently ignores.

---

## 6. Needs

`spec.needs` is the declaration of values the agent consumes at runtime — environment variables, credentials, file-backed secrets, and inline configuration. Each entry binds a logical name to a resolver. The resolver produces a value; the value is delivered as an environment variable or as a file in the agent runtime, depending on the entry's `target`.

### 6.1 Schema

```yaml
spec:
  needs:
    - name: <string>            # required, matches [A-Za-z_][A-Za-z0-9_]*
      target: env | file        # optional, default "env"
      mountPath: <path>         # required when target=file, forbidden when target=env
      optional: <bool>          # optional, default false
      from: <Resolver>          # optional when target=env (see 6.3)
```

Names MUST be unique within `spec.needs` and MUST match `[A-Za-z_][A-Za-z0-9_]*` so they are safe to use as environment variable names.

When `target` is `env`, the resolved value is delivered to the agent runtime as an environment variable with the same name as `name`. When `target` is `file`, the resolved value is written as a file at `mountPath` inside the agent runtime, on a tmpfs that does not persist across runs. The resolver's output for file targets MUST be raw bytes, not a host path.

### 6.2 Resolvers

A resolver is specified via the `from:` field with exactly one resolver kind. Six kinds are defined in this specification:

```yaml
# Host environment variable
from:
  envFrom: <var-name>

# Host file contents (raw bytes)
from:
  file: <host-path>

# Inline literal (non-secret config)
from:
  text: <string>

# macOS Keychain entry
from:
  keychain:
    service: <string>
    account: <string>

# HashiCorp Vault
from:
  vault:
    path: <vault-path>
    field: <field-name>

# Short-lived OIDC workload identity token
from:
  oidc:
    audience: <audience>
```

The `keychain` resolver reads an existing item from the host keychain; both `service` and `account` are required. A published definition SHOULD declare a bare need (section 6.3), which falls back to the keychain under the resolver's own convention, or leave the source to an operator-side override (section 6.4), rather than name a host's keychain layout.

The `oidc` resolver MUST be used with `target: file`. The resolver writes a signed JWT (the agent's workload identity token) to `mountPath`. Downstream tools, such as the AWS SDK reading `AWS_WEB_IDENTITY_TOKEN_FILE`, exchange it for short-lived credentials at the cloud provider.

This specification enumerates the resolver kinds defined for this revision. A conforming resolver MUST implement every kind defined here. New resolver kinds (e.g., `kubernetesSecret`, `azureKeyVault`, `awsSecretsManager`) are added through normal revision of this specification, not through operator-level extension. Unknown resolver kinds MUST cause a validation error rather than silent fallthrough.

A gateway consumer (section 6.5) applies a need's value as a static secret, so it MUST reference a need with `target: env`; a need the `oidc` resolver produces is a file and cannot be referenced by one.

### 6.3 Defaults

When all of the following are true:

- `target` is `env` (explicit or by default)
- `from` is omitted

then the resolver defaults to `envFrom: <name>`. That is, `- name: ANTHROPIC_API_KEY` is shorthand for:

```yaml
- name: ANTHROPIC_API_KEY
  target: env
  optional: false
  from:
    envFrom: ANTHROPIC_API_KEY
```

The resolver MUST normalize the implicit default into the explicit form during validation so that audit logs, policy evaluation, and downstream code observe one canonical shape.

A need whose resolver is `envFrom` naming the need's own name — the canonical form of a bare need — resolves in this order:

1. The host environment variable of that name. When it is set, its value is used.
2. When the variable is unset, the host keychain item under the resolver's keychain service name, with the need's name as the account. When that item exists, its value is used.
3. When neither exists, the need is unresolved: a required need causes deploy-time rejection and an optional need is skipped with a deploy-time warning (section 10.6).

The environment variable takes precedence when both exist. The resolver MUST record in the need's audit entry which source answered: `envFrom` when the variable supplied the value, `keychain` when the keychain item did. An `envFrom` resolver naming any other variable, and every other resolver kind, reads only the source it names.

The keychain service name is the resolver's choice, not this specification's. A resolver MUST document it, MUST name it wherever it reports an unresolved need, and MUST store nothing else under it. That convention is the whole mechanism: an operator stores, inspects and removes those items with the host's own keychain tools.

The default exists to keep common cases free of boilerplate. A definition declares a need by name; whether the value comes from the shell or the keychain is a per-host choice that the definition does not encode. Policies MAY still forbid implicit-env resolution where it would be inappropriate (section 12).

### 6.4 Operator-side resolver configuration

The agent definition declares *what* values the agent needs; how those values are produced on a given host is an operator concern. Resolver backends (Vault address, OIDC issuer, per-need overrides) live in operator configuration outside the agent definition. A conforming resolver MUST provide a documented mechanism for operator-side configuration of:

- Backend addresses and auth methods for backed resolvers (vault, oidc, etc.)
- Per-need overrides that replace the definition's `from:` with a different resolver kind on a given host
- Precedence rules between CLI flag overrides, host configuration, and the definition's declared `from:`

An operator override MUST NOT change `target` or `mountPath` — those are part of the agent definition's contract.

### 6.5 Needs withheld from the agent

A **gateway consumer** is a field whose value the gateway applies on the agent's behalf. There are three:

- an `llm` module's `credential` (section 8.1);
- a remote MCP module's `headers` (section 8.2);
- the `headers` of an object entry in `spec.egress` (section 9.3).

A need that any gateway consumer references is **withheld**. The gateway obtains its value from the resolver and applies it; the agent runtime and local MCP servers never receive it. The login a registry entry in `spec.egress` applies is not a need, and section 9.7 withholds it the same way.

The resolver MUST NOT deliver a withheld need's value to the agent runtime's environment or to a local MCP server's environment, and MUST NOT place any other value from the host under the need's name there, except the placeholder. This takes precedence over the delivery requirements of sections 6.1 and 8.2.

In place of the value, the agent runtime and every local MCP server receive:

| Need referenced by | Delivered |
|--------------------|-----------|
| The `headers` of an egress object entry, whether or not another consumer also references it | A placeholder, as an environment variable under the need's name |
| Only an `llm` module's `credential`, a remote MCP module's `headers`, or both | Nothing |

The placeholder's form, and what becomes of the provider's conventional variable names when an `llm` module declares `credential`, are set out in Appendix A.1.

Withholding is decided by the definition's declarations alone. A need is withheld because a gateway consumer references it, never because of its name or its value. Withholding, including the provider variable names of Appendix A.1, is computed before any reference to an optional need that failed to resolve is dropped (sections 8.1, 8.2 and 9.6), so it does not depend on what resolved on the host. It is computed on the merged chain (sections 10.1 and 10.7), so a derived definition that adds a gateway consumer referencing an inherited need withholds that need from every consumer, including a local consumer that relied on the real value under the base. The resolver MUST report, at agent start, each need it withheld.

---

## 7. Module Types

Modules are instructions to the resolver. They are declared under `spec.includes`. Four types are defined:

| Type | What the Resolver Does | Source Options |
|------|----------------------|----------------|
| `llm` | Configures an inference endpoint (remote API or local serving) | `source.model` (remote, with optional `source.endpoint` override) or `source.modelkit` (local: a ModelKit in the registry, or `./name` for this ModelKit's model) |
| `mcp` | Launches a local MCP server process or connects to a remote one | `source.modelkit` (local: a ModelKit in the registry, or `./name` for a bundle in this ModelKit) or `source.endpoint` (remote) |
| `policy` | Loads enforcement rules into the policy engine | `source.modelkit` |
| `skill` | Mounts a skill where the framework discovers it | `source.modelkit` (a ModelKit in the registry, or `./name` for a skill in this ModelKit) |

Local skills are packed into the ModelKit's `prompts:` layers and declared with a `skill` module whose `source.modelkit` names the entry relatively. Local prompts and rules are packed into `prompts:` layers and need no module. External skills from the registry use a plain `source.modelkit`.

Every module MUST have a `name` (unique within `spec.includes`) and a `type`.

**Relative references.** A `source.modelkit` that begins with `./` is a relative reference: it names an entry of the ModelKit that carries the declaring document, by that entry's `name`, in the section the module's type reads from — `mcpServers` for `mcp`, `prompts` for `skill`, `model` for `llm`. The form exists for content packaged beside the agent that uses it. A plain `source.modelkit` names a different ModelKit in a registry, and a bare name would be one too, so the prefix is what tells them apart. A relative reference MUST resolve to exactly one entry: a reference to no entry, or to more than one, MUST be rejected. In a composed definition the ModelKit is the one the module was declared in, so a derived definition that inherits such a module uses the base's entry, not one of its own, and replaces it by declaring a module of the same name (section 10.1). A `policy` module has no relative form, because a Kitfile has no policy section.

---

## 8. Module Specifications

### 8.1 `type: llm`

Declares an LLM provider the agent requires. The resolver registers the provider so that the framework can route inference requests through it. Where the gateway reaches the provider on the agent's behalf, the module names the need it authenticates with in `credential`. Where the framework reaches the provider itself, the framework reads its credential from the runtime environment produced by `spec.needs`.

**Remote API — well-known provider:**

```yaml
- name: primary-llm
  type: llm
  source:
    model: anthropic/claude-opus-4-5
```

**Remote API — custom endpoint (Azure, self-hosted, etc.):**

```yaml
- name: azure-llm
  type: llm
  source:
    model: azure/gpt-4o
    endpoint: https://my-deployment.openai.azure.com
```

When `source.endpoint` is omitted, the resolver uses the well-known endpoint for the provider indicated by the `source.model` prefix (e.g., `anthropic/` → `https://api.anthropic.com`). When `source.endpoint` is specified, the resolver routes requests to that URL instead.

**Local serving:**

```yaml
- name: local-llm
  type: llm
  source:
    modelkit: jozu.com/models/llama3:8b
```

When `source.modelkit` is specified, the resolver MUST start a local inference server from the referenced artifact and register it as an available provider.

**Local from this ModelKit:**

```yaml
- name: local-llm
  type: llm
  source:
    modelkit: ./llama3-8b
```

A relative reference (section 7) names this ModelKit's `model` by its `name`. The resolver MUST serve it as it serves a model pulled from the registry. A reference that is not the name of that ModelKit's `model`, including one to a `model` entry with no `name`, MUST be rejected.

#### Credential

```yaml
spec:
  needs:
    - name: ANTHROPIC_KEY
      from:
        envFrom: ANTHROPIC_API_KEY
  includes:
    - name: primary-llm
      type: llm
      source:
        model: anthropic/claude-sonnet-4-5
      credential: ANTHROPIC_KEY
```

`credential` is the name of one need. The gateway applies the need's value to every request it sends to the provider, in the provider's API-key scheme. The need is withheld from the agent (section 6.5).

Constraints:

- `credential` MUST name a need declared or inherited in `spec.needs`.
- A module with `credential` MUST name, in `source.model`, a provider the provider table has an entry for, written `<provider>/<model>`. A `source.model` with no provider prefix, a provider the table has no entry for, and a module without `source.model` MUST be rejected.
- The need MUST have `target: env` (section 6.2). A `credential` naming a file-target need MUST be rejected.
- `credential` is allowed only where the gateway holds the module's provider entry, which is where the gateway reaches the provider on the agent's behalf. Which of the two applies is decided by the framework the resolver runs. Where the framework reaches its provider itself, with a credential from the runtime environment, a module that declares `credential` MUST be rejected: withholding the provider's variable names (section 6.5) would take the credential from the framework and give it to nothing.
- `credential` is permitted on `llm` modules only. It MUST be rejected on any other module type.

A module whose `credential` names an optional need that failed to resolve MUST be skipped: the resolver does not register its provider, and MUST report the module it skipped. A required need that fails to resolve aborts resolution (section 10.6).

The **provider table** is the resolver's: the LLM providers whose credential its gateway can apply, keyed by the provider prefix of `source.model`, each with the variable names that conventionally carry that provider's credential (Appendix A.2).

#### Structured fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique identifier within `spec.includes` |
| `type` | Yes | `llm` |
| `source.model` | If remote | Provider and model identifier (e.g., `anthropic/claude-opus-4-5`). The provider prefix indicates the API format. |
| `source.endpoint` | No | URL override for the LLM API. Required for custom deployments (Azure, self-hosted), optional for well-known providers. |
| `source.modelkit` | If local | ModelKit reference to a servable model artifact, or `./name` for this ModelKit's `model` (section 7). |
| `credential` | No | Name of the need the gateway authenticates to the provider with. Requires `source.model` naming a provider in the provider table. |
| `config` | No | Map of name → literal string: non-secret settings for the provider, such as a region or an API version. Needs are not interpolated into it; a secret belongs in the need `credential` names. |

A module MUST specify exactly one of `source.model` or `source.modelkit`.

---

### 8.2 `type: mcp`

Declares an MCP server the agent requires. The resolver either launches a local process or connects to a remote endpoint, and makes the server's tools available to the framework.

A module MAY declare a `tools` allow-list naming the tools the agent is permitted to reach on that server, defined below. Where a module declares none, the resolver MUST make all of that server's tools available to the framework.

The allow-list is not a substitute for policy, and the two are evaluated independently: a tool invocation MUST satisfy both. The allow-list is the static shape of the module, declared by whoever packages the agent, and cannot inspect an invocation's arguments. Policy is the operator's argument-aware judgement at invocation time, and cannot remove a tool from discovery.

Local MCP servers receive the values produced by `spec.needs`: every successfully resolved need with `target: env`, other than one withheld under section 6.5, and every file-target need at its `mountPath`, whether the server shares the agent runtime's process boundary or the resolver delivers them to it explicitly (Appendix A.3). Local MCP servers read their own configuration from environment variables or files, as documented by the MCP server itself. The agent definition does not template values into MCP module fields for local MCPs.

Remote MCP servers cannot read the agent runtime's environment. Their module MAY declare a `headers:` map whose values are sent on every request to the endpoint. Header values MAY reference needs via `${needs.NAME}`. The gateway substitutes them when it connects to the server or sends it a request, and the agent runtime never receives them: each need the headers reference is withheld (section 6.5). `headers:` is forbidden on local MCP modules.

**Local from ModelKit:**

```yaml
- name: filesystem
  type: mcp
  source:
    modelkit: jozu.com/mcps/filesystem:v1
  config:
    allowedPaths: /workspace
```

When `source.modelkit` is specified, the resolver MUST pull the referenced ModelKit, read its entrypoint metadata, and launch the server. Transport is always stdio for local MCP servers.

**Local from this ModelKit:**

```yaml
- name: filesystem
  type: mcp
  source:
    modelkit: ./filesystem
```

A relative reference (section 7) names this ModelKit's `mcpServers` entry by its `name` (see `mcpServers` in the Kitfile reference). The resolver MUST extract that entry's bundle and launch it as it launches a bundle pulled from the registry.

**Remote HTTP, with authentication:**

```yaml
- name: github
  type: mcp
  source:
    endpoint: https://mcp.github.com/mcp
  transport: http
  headers:
    Authorization: "Bearer ${needs.GITHUB_TOKEN}"
```

**Remote SSE:**

```yaml
- name: example-remote
  type: mcp
  source:
    endpoint: https://mcp.example.com/sse
  transport: sse
```

#### Header interpolation

A header value MAY contain one or more `${needs.NAME}` references. `NAME` MUST match a need declared (or inherited) in `spec.needs`. The gateway MUST substitute each occurrence with the need's resolved value before the header is sent, and MUST NOT expose the substituted value to the agent runtime. Constraints:

- A header value that references a need with `target: file` MUST be rejected — file-target needs deliver bytes to a path, not strings.
- A header that references an optional need which failed to resolve MUST be dropped: the gateway does not send it, and does not substitute the empty string. The module's other headers are still sent.
- Any `${…}` token that is not of the form `${needs.NAME}` for a known need MUST be rejected at validation.
- A header value MUST be a valid HTTP field value. A value containing a control character, such as CR or LF, MUST be rejected at validation, so that a definition cannot split one header into two or carry a value no request can send.

A need a remote MCP module's headers reference is withheld, and the agent runtime and local MCP servers receive nothing in its place (section 6.5). An agent that also uses the same value itself, against a host it calls directly, declares an object entry in `spec.egress` for that host (section 9.3). The resolver reports each such need at agent start, as section 6.5 requires.

An egress entry's `headers` follow the same rules, and also accept the `basic` form (section 9.3). A remote MCP module's header values are strings.

#### Tool allow-list

A module MAY declare `tools`, a list of tool names the agent is permitted to reach on that server:

```yaml
- name: github
  type: mcp
  source:
    endpoint: https://mcp.github.com/mcp
  transport: http
  tools:
    - create_issue
    - list_issues
```

Constraints:

- Each entry MUST be a tool name as the server reports it in its own `tools/list` response. The resolver MUST NOT interpret globs, patterns, or any composed form such as a framework-namespaced or gateway-prefixed name. Matching is case-insensitive, so a server exposing two tools whose names differ only in case cannot be narrowed between them; authors SHOULD write each entry in the case the server reports.
- An absent `tools` field means every tool the server exposes. This is the default, so a definition written before this field behaves as it did.
- An empty list MUST be rejected at validation. Omitting the field is how a module declares no restriction, so an empty list can only be an authoring mistake, and reading it as "no tools" would strand the module silently.
- `tools` is permitted on MCP modules only, local and remote alike. It MUST be rejected on any other module type.
- Entries are not substitutable: a `${…}` token in a tool name MUST be rejected at validation.
- The resolver SHOULD reject a duplicate name rather than deduplicating it silently.

Where a module declares `tools`, the resolver MUST NOT make any other tool of that server reachable (Appendix A.3).

#### Structured fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique identifier within `spec.includes` |
| `type` | Yes | `mcp` |
| `source.modelkit` | If local | ModelKit reference containing the MCP server, or `./name` for a bundle in this ModelKit (section 7) |
| `source.endpoint` | If remote | URL of the remote MCP server |
| `transport` | If remote | `http` or `sse` |
| `headers` | No (remote only) | Map of header name → string value the gateway sends on every request. Values MAY contain `${needs.NAME}` references. Forbidden for local MCPs. |
| `tools` | No | List of tool names the agent may reach, spelled as the server reports them. Absent means all. Empty is invalid. |
| `config` | No | Opaque. Passed through to the MCP server. |

A module MUST specify exactly one of `source.modelkit` or `source.endpoint`. When `source.endpoint` is specified, `transport` MUST also be specified.

---

### 8.3 `type: policy`

Declares a policy the resolver MUST enforce. The policy ModelKit itself declares its kind, evaluation point, and rules. The agent definition does not re-describe any of that — it only references the artifact.

```yaml
- name: tool-restrictions
  type: policy
  source:
    modelkit: jozu.com/policies/strict-tools:v1
```

#### Structured fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique identifier within `spec.includes` |
| `type` | Yes | `policy` |
| `source.modelkit` | Yes | ModelKit reference to the policy artifact |

---

### 8.4 `type: skill`

Declares a skill the agent requires. The resolver mounts the skill's contents where the framework discovers skills at runtime, as `<name>/SKILL.md` with the skill's other files beside it. The skill comes from one of two places: a separate ModelKit in the registry, or a `prompts` entry of the ModelKit that carries the declaring document.

**From the registry:**

```yaml
- name: code-review-skill
  type: skill
  source:
    modelkit: jozu.com/skills/code-review:v1
```

```yaml
- name: security-analysis
  type: skill
  source:
    modelkit: registry.jozu.com/my-org/security-analysis:v2
```

**From this ModelKit:**

```yaml
- name: code-review
  type: skill
  source:
    modelkit: ./code-review
```

A relative reference (section 7) names this ModelKit's `prompts` entry by its `name` (see `prompts` in the Kitfile reference). The resolver MUST take that entry's contents and mount them as it mounts a skill pulled from the registry.

A skill packed in `prompts:` that no module names is still mounted with the ModelKit's other content (section 12), at its own path. The module is what places it where the framework looks.

#### Structured fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique identifier within `spec.includes` |
| `type` | Yes | `skill` |
| `source.modelkit` | Yes | ModelKit reference to the skill artifact, or `./name` for a skill in this ModelKit's `prompts` (section 7) |

---

## 9. Egress

`spec.egress` declares the hosts the agent runtime may reach beyond what its modules already provide: source hosting, package registries, internal services. Each entry is a string, which grants reachability, an object, which also names headers the gateway sets on every request to the host, or a registry entry, which also has the gateway sign in to a container registry with the login the host holds (section 9.7).

The **gateway** is the part of the resolver through which the agent runtime's HTTPS traffic leaves and which reaches LLM providers and remote MCP servers on the agent's behalf. A **provider host** is a host whose connections the gateway terminates to apply policy to LLM traffic. The set of provider hosts belongs to the resolver, not to the definition.

Entries apply to the agent runtime, to the processes it spawns, and to local MCP servers alike.

### 9.1 Schema

```yaml
spec:
  needs:
    - name: GITHUB_TOKEN
    - name: ARTIFACTS_TOKEN
  egress:
    - pypi.org
    - "*.pythonhosted.org"
    - "tcp://db.internal:5432"
    - oci://ghcr.io
    - host: api.github.com
      headers:
        Authorization: "Bearer ${needs.GITHUB_TOKEN}"
    - host: artifacts.example.com
      headers:
        Authorization:
          basic: "ci:${needs.ARTIFACTS_TOKEN}"
```

### 9.2 String entries

A string entry takes one of three forms:

| Entry | Grants |
|-------|--------|
| `example.com`, `*.example.com` | HTTPS on port 443 to that name, or to any name under the suffix |
| `tcp://host:port` | TCP to that name on that port |
| `oci://registry` | HTTPS on port 443 to the registry's hosts, with the host's login for the registry applied by the gateway (section 9.7) |

Constraints:

- Every host is a DNS name. An IP address MUST be rejected, as MUST a name whose last label is all digits or begins `0x`, which some resolvers read as an address. A definition is published and runs on many networks, and an address names a different machine on each; addresses are a run-time concern of the host that runs the agent, outside the definition.
- Names are compared without regard to case. `*.example.com` covers every name ending in `.example.com` and does not cover `example.com` itself.
- A `tcp://` entry is a URL with a host and a port and nothing else. A trailing slash after the port MAY be accepted; a path, query, fragment or user MUST be rejected. `tcp` and `oci` are the only schemes an entry takes, and an entry with any other scheme MUST be rejected: an HTTPS host is written as a bare name.
- A `tcp://` entry MUST name a port between 1 and 65535. It MUST NOT name port 443, which the plain name form covers.
- A `tcp://` entry MUST name an exact host. A wildcard MUST be rejected: only the gateway checks a name, and only on port 443, so a wildcard on another port would open the address of any name under the suffix that the agent cared to resolve.
- A host MAY appear in several entries and in both forms.

An entry is a grant and carries no value. It is never a condition for the agent to start: a host that does not resolve or cannot be reached on the network the agent runs on MUST NOT cause deploy-time rejection.

A provider host stays on the gateway's terminating path whatever `spec.egress` says. Listing a provider host, or a wildcard that covers one, neither removes the host from that path nor adds a second route to it.

### 9.3 Object entries

An object entry names one host and the headers the gateway sets on its requests:

| Field | Required | Description |
|-------|----------|-------------|
| `host` | Yes | An exact DNS name. Port 443 is implied. |
| `headers` | Yes | Map of header name → value. A value is a string, a literal or containing `${needs.NAME}` references, or a mapping with the single key `basic`. |

An object entry grants what the string entry for its host would grant, and also requires the gateway to apply its headers (section 9.4).

Constraints:

- `host` MUST be an exact DNS name under the rules of section 9.2. A wildcard MUST be rejected, because headers applied to every name under a suffix go to whoever registers one. The `tcp://` form MUST be rejected, because the gateway terminates only on port 443. An address MUST be rejected.
- An object entry without `headers` MUST be treated exactly as the string entry for its `host`, including the wildcard and `tcp://` forms of section 9.2; the constraints below apply only to an entry that carries `headers`.
- Each header name MUST be a valid HTTP field name. It MUST NOT be `Host`, `Content-Length`, `Transfer-Encoding`, `Connection`, `Keep-Alive`, `TE`, `Trailer` or `Upgrade`, and MUST NOT begin with `Proxy-`, compared without regard to case. The gateway owns those.
- A header value follows the rules for header interpolation in section 8.2: `NAME` MUST match a need declared or inherited in `spec.needs`, a reference to a need with `target: file` MUST be rejected, and any other `${…}` token MUST be rejected. The value is substituted in the gateway at request time, and the agent never receives it (section 6.5).
- A header value MAY instead be a mapping with the single key `basic`, whose value is a string template under the same rules. The gateway renders the template, encodes the result in base64 and sends `Basic`, a space, and the encoding. A credential a service expects as a Basic user-password pair (RFC 7617) is written `basic: "user:${needs.NAME}"`, and the need holds the raw secret. A mapping with any other key, or with more than one key, MUST be rejected, as MUST a `basic` template that is empty.
- An empty or null value MUST be rejected.
- A host MUST appear in at most one object entry per document.

### 9.4 Gateway behaviour

For each object entry in the resolved set, the gateway terminates the host's connections and applies the entry's headers as Appendix A.4 requires. A resolver that cannot apply an object entry that way MUST reject the definition rather than run it without the entry.

### 9.5 Validation

Validation runs on the merged chain, after the merge of section 10.7, because a derived definition can add or remove the modules and needs the checks depend on. The resolver MUST reject the definition if:

- an object entry's header references a need that is not declared anywhere in the chain;
- an object entry's header references a need with `target: file`;
- an object entry's host is a provider host;
- any entry, string or object, names the endpoint host of a remote MCP module, or is a wildcard that covers one. The gateway reaches a remote MCP server holding the module's headers, under policy, and enforcing the module's `tools` allow-list; a direct route would bypass all three;
- any entry, string or object, names the `source.endpoint` host of an `llm` module whose endpoint the gateway reaches on the agent's behalf, or is a wildcard that covers one. Where the framework instead reaches the endpoint itself, the endpoint host is an ordinary host and this rule does not apply.

### 9.6 Unresolved needs

A header whose value references an optional need that failed to resolve is dropped from its object entry; the gateway does not apply it. An object entry left with no headers remains in the resolved set as the string entry for its host. A required need that fails to resolve aborts resolution (section 10.6).

### 9.7 Registry entries

A registry entry names a container registry the agent pulls from or pushes to, and has the gateway sign in to it with the login the host already holds:

```yaml
spec:
  egress:
    - oci://ghcr.io
    - oci://123456789012.dkr.ecr.us-east-2.amazonaws.com
```

The entry carries no reference to a login. A registry login belongs to whoever runs the agent, not to the definition, and the container tools on the host already keep one per registry. The resolver reads the host's **registry login store**, the store those tools read, through any credential helper it names, under the name those tools use for the registry. Which store a resolver reads on a given host is operator configuration (section 6.4).

Constraints:

- The entry is a URL with the scheme `oci` and a host, and nothing else. A trailing slash after the host MAY be accepted; a port, path, query, fragment or user MUST be rejected. The gateway terminates on port 443 only, so a registry on another port cannot take a login this way. A path would name a repository, and the login applies to the registry, not to a repository.
- The host MUST be an exact DNS name under the rules of section 9.2. A wildcard MUST be rejected, as MUST an address.
- A registry MAY also appear in a string entry. A document MUST NOT list a host of a registry entry in an object entry: the entry's login and the object entry's headers would both claim the request's `Authorization`. Section 10.7 applies the same rule across the chain.
- The hosts of a registry entry are subject to the checks of section 9.5 as the host of an object entry is.

The gateway signs in to the registry with the store's login, and the agent never receives it (Appendix A.5). Where the store holds no login for the registry, the agent reaches it as an anonymous client and the resolver reports so at agent start. The hosts a registry's responses name, its token service and content hosts, are admitted under the rules of Appendix A.5. An entry is still a grant and not a condition for the agent to start, as section 9.2 says of every entry.

---

## 10. Composition

An agent definition MAY reference a base agent via `spec.base`:

```yaml
spec:
  base:
    modelkit: registry.example.com/org/base-agent:v1
```

`spec.base.modelkit` references another agent ModelKit, one that carries an Agent Definition of its own. The derived agent inherits the base's `needs`, `includes`, `egress`, and content layers, and can add, override, or restrict. The resolver MUST resolve the full base chain before merging.

### 10.1 Module merge rules

The resolver MUST apply these rules when merging a derived agent's `spec.includes` with its base:

| Scenario | Behavior |
|----------|----------|
| New name with type/source | Added to the module set |
| Same name with type/source | Overrides the base module entirely, unless the base module is a policy (section 10.2) |
| Same name with nothing else | Removes that module from the resolved set, unless the base module is a policy (section 10.2) |

Every module other than a policy is the derived agent's to override or remove: its LLM, its MCP servers, its skills. Policies are the exception, and the only one.

**Override example:**

```yaml
- name: primary-llm
  type: llm
  source:
    model: anthropic/claude-sonnet-4-20250514
```

**Remove example:**

```yaml
- name: github
```

### 10.2 Policy accumulation

Policies are additive and protected. A derived agent MUST NOT remove or replace a policy from its base: the name of a policy module anywhere in the chain is reserved, and a derived definition that declares a module of that name — a name-only entry, a policy with a different source, or a module of any other type — is a validation error the resolver MUST reject. A derived agent adds policies; it never touches the ones it inherits.

This is a deliberate security property: restrictions only accumulate up the composition chain. To create a less restricted variant, compose from a less restricted base.

### 10.3 Needs merge rules

The resolver MUST apply these rules when merging a derived agent's `spec.needs` with its base:

| Scenario | Behavior |
|----------|----------|
| New name | Added to the resolved `needs` set |
| Existing name, same `target` and `mountPath`, new `from` | `from` is replaced (definition-level override) |
| Existing name, `optional: true` in base, `optional: false` in derived | Tightening: allowed |
| Existing name, `optional: false` in base, `optional: true` in derived | Loosening: allowed when no gateway consumer references the need in the merged chain (below); otherwise MUST be rejected |
| Existing name, different `target` or `mountPath` | MUST be rejected (changes the consumption contract) |
| Name-only removal entry | MUST be rejected (needs are additive only) |

Needs are never removed: a need the base declares binds every descendant. A need the derived definition leaves with nothing to consume it, because it removed the module that read it, is **dangling**, and what the resolver can do about it depends on how the need was consumed.

A need a gateway consumer references (section 6.5) is consumed only through that reference, so the resolver can see the reference go. A need that some definition in the chain references from a gateway consumer and that none references in the merged chain is dangling: the resolver MUST NOT deliver it to the agent runtime or to local MCP servers, MUST NOT require or resolve it, and MUST report it at agent start, naming the definition that referenced it and the module whose removal left it unreferenced. A derived definition that references it again, from a module of its own, makes it an ordinary withheld need.

A need consumed through the environment has no reference the resolver can watch, so only the derived definition can say it is no longer read, and it says so by loosening the need to `optional: true`: the need stays in the chain, resolves when the host supplies it, and no longer blocks the agent when it does not. Loosening is permitted only while no gateway consumer in the merged chain references the need. A need a kept module's `credential` or `headers`, or an inherited egress entry, still references MUST stay as the base declared it, since an unset optional need would drop the header or skip the module and change what a vetted base sends (section 10.7). The resolver MUST report each loosened need at agent start.

### 10.4 Content layers

Content layers (`code:`, `prompts:`, `model:`, `datasets:`) belong to the ModelKit, not the Agent Definition. Layer composition follows the underlying ModelKit composition semantics: layers from a base ModelKit are inherited and mounted alongside the derived ModelKit's content. This specification does not redefine layer merge rules. A module with a relative `source.modelkit` (section 7) keeps naming the ModelKit it was declared in when it is inherited; a derived definition replaces what the module refers to by declaring a module of the same name (section 10.1).

### 10.5 Composition depth

`spec.base` references MAY chain (A extends B extends C). The resolver MUST flatten the chain at deploy time, applying merge rules at each level bottom-up. The resolver MUST reject circular references.

### 10.6 Resolution order

The resolver MUST flatten in this order:

1. Walk the base chain bottom-up, merging `needs`, `includes`, `egress`, and content layers per the rules above.
2. Normalize default `target`, `optional`, and `from` values on every `needs` entry.
3. Resolve every `needs` entry using its `from:` (or operator override). Required entries that fail to resolve MUST cause deploy-time rejection. Optional entries that fail to resolve are skipped and produce a deploy-time warning.
4. Configure the resolved modules with the runtime environment that `needs` produced.

### 10.7 Egress merge rules

The resolved `spec.egress` is the union of the entries of every definition in the chain. An entry contributed by more than one definition appears once. The resolver MUST apply these rules when merging a derived agent's `spec.egress` with its base:

| Scenario | Behavior |
|----------|----------|
| New string entry | Added to the resolved `egress` set |
| New object entry, for a host no ancestor lists | Added to the resolved `egress` set |
| Object entry for a host an ancestor lists only in string form | Added; the gateway applies its headers to that host |
| String entry for a host an ancestor lists as an object entry | Allowed; the inherited object entry is unchanged |
| Object entry for a host an ancestor lists as an object entry | MUST be rejected |
| Registry entry for a registry an ancestor lists as a registry entry | Appears once |
| Registry entry for a registry whose hosts an ancestor lists in an object entry, or object entry for a host of a registry an ancestor lists as a registry entry | MUST be rejected |

A derived definition MAY add entries and MUST NOT remove an inherited one. An inherited object entry MUST NOT be changed or redeclared, whether its headers would differ or not.

Changing an inherited object entry would change which credential a vetted base sends to a host. Object entries therefore accumulate up the chain as policies do (section 10.2).

Validation of the merged set (section 9.5) and withholding (section 6.5) are computed after this merge.

---

## 11. Examples

Each example shows a definition in file form; the contents of its `spec` serve unchanged as an `agent:` block (section 4.1). The surrounding ModelKit (Kitfile, content layers) is omitted; it is packaged with standard ModelKit semantics and the resolver mounts content layers as described in section 4.

### 11.1 Root Agent — Code Review

```yaml
apiVersion: kitops.org/v1
kind: AgentDefinition
metadata:
  name: code-review-agent
spec:
  framework: claude-code
  needs:
    - name: ANTHROPIC_API_KEY
  includes:
    - name: primary-llm
      type: llm
      source:
        model: anthropic/claude-opus-4-5

    - name: filesystem
      type: mcp
      source:
        modelkit: jozu.com/mcps/filesystem:v1
      config:
        allowedPaths: /workspace

    - name: tool-restrictions
      type: policy
      source:
        modelkit: jozu.com/policies/strict-tools:v1

    - name: content-guardrail
      type: policy
      source:
        modelkit: jozu.com/policies/content-safety:v1
```

The Anthropic SDK in the framework reads `ANTHROPIC_API_KEY` from the runtime environment. The agent-definition.yaml declares the need; the SDK finds it. Prompts are packed into the ModelKit's `prompts:` layers, mounted by the resolver, and discovered by the framework; a local skill would be packed there too and declared with a `skill` module that names it (section 8.4).

### 11.2 Composed Agent — Restricted Variant

```yaml
apiVersion: kitops.org/v1
kind: AgentDefinition
metadata:
  name: code-review-agent-restricted
spec:
  base:
    modelkit: registry.example.com/org/code-review-agent:v1
  includes:
    - name: primary-llm
      type: llm
      source:
        model: anthropic/claude-sonnet-4-20250514

    - name: no-database-access
      type: policy
      source:
        modelkit: jozu.com/policies/no-database-access:v1
```

**Resolved composition:**

- `framework: claude-code` — inherited
- `needs: ANTHROPIC_API_KEY` — inherited from base
- `primary-llm` — overridden to Sonnet
- `filesystem` MCP — inherited
- `tool-restrictions` policy — inherited
- `content-guardrail` policy — inherited
- `no-database-access` policy — added
- Prompts and skills — inherited from base content layers

### 11.3 Composed Agent — Air-Gapped Environment

```yaml
apiVersion: kitops.org/v1
kind: AgentDefinition
metadata:
  name: code-review-agent-onprem
spec:
  base:
    modelkit: registry.example.com/org/code-review-agent:v1
  needs:
    - name: ANTHROPIC_API_KEY
      optional: true
  includes:
    - name: primary-llm
      type: llm
      source:
        modelkit: jozu.com/models/llama3:70b

    - name: local-git
      type: mcp
      source:
        modelkit: jozu.com/mcps/git:v1
      config:
        repoPath: /repos
```

The local LLM does not require an API key, so the inherited `ANTHROPIC_API_KEY` need would be dangling. No gateway consumer references it (the base's `primary-llm` declares no `credential`), so the derived definition loosens it to `optional: true` (section 10.3): the need stays in the chain, and an air-gapped host where the variable is unset no longer blocks the agent.

### 11.4 Root Agent — Local Model with Guardrails

```yaml
apiVersion: kitops.org/v1
kind: AgentDefinition
metadata:
  name: financial-analyst-agent
spec:
  framework: pydantic-ai
  includes:
    - name: analyst-llm
      type: llm
      source:
        modelkit: ./mixtral-8x7b-finance

    - name: filesystem
      type: mcp
      source:
        modelkit: jozu.com/mcps/filesystem:v1
      config:
        allowedPaths: /data/financial-reports

    - name: spreadsheet
      type: mcp
      source:
        modelkit: jozu.com/mcps/spreadsheet:v1

    - name: ratio-analysis
      type: skill
      source:
        modelkit: ./ratio-analysis

    - name: financial-compliance
      type: policy
      source:
        modelkit: jozu.com/policies/financial-compliance:v1

    - name: pii-filter
      type: policy
      source:
        modelkit: jozu.com/policies/pii-detection:v1
```

The fine-tuned model weights, system prompts, analysis skills, and regulatory reference datasets are packaged in the ModelKit's `model:`, `prompts:` and `datasets:` layers. `analyst-llm` names the packed model, so the resolver serves it locally; `ratio-analysis` names a packed skill, so the resolver places it where the framework looks for skills. The prompts and datasets no module names are mounted with the rest and discovered by the framework.

### 11.5 Root Agent — Short-Lived AWS Credentials via OIDC

```yaml
apiVersion: kitops.org/v1
kind: AgentDefinition
metadata:
  name: incident-response-agent
spec:
  framework: claude-code
  needs:
    - name: ANTHROPIC_API_KEY

    - name: AWS_WEB_IDENTITY_TOKEN_FILE
      target: file
      mountPath: /run/creds/oidc/aws-token
      from:
        oidc:
          audience: sts.amazonaws.com

    - name: AWS_ROLE_ARN
      from:
        text: arn:aws:iam::123456789012:role/incident-response-readonly

    - name: AWS_REGION
      from:
        text: us-east-1

    - name: KUBECONFIG
      target: file
      mountPath: /run/creds/k8s/kubeconfig
      from:
        vault:
          path: secret/data/agents/incident-response
          field: kubeconfig
  includes:
    - name: primary-llm
      type: llm
      source:
        model: anthropic/claude-sonnet-4-20250514

    - name: kubectl
      type: mcp
      source:
        modelkit: jozu.com/mcps/kubectl:v3
      config:
        default_namespace: production
        read_only: "true"

    - name: aws-cloudwatch
      type: mcp
      source:
        modelkit: jozu.com/mcps/cloudwatch:v1

    - name: incident-response-policies
      type: policy
      source:
        modelkit: jozu.com/policies/incident-response:v2
```

The AWS SDK in the cloudwatch MCP respects the standard `AWS_WEB_IDENTITY_TOKEN_FILE` + `AWS_ROLE_ARN` convention: the SDK reads the JWT file written by the OIDC resolver and calls `sts:AssumeRoleWithWebIdentity` to obtain short-lived credentials. The kubectl MCP reads `KUBECONFIG` from its environment. No long-lived cloud credentials touch the agent runtime, and no value is templated from `needs:` into module fields.

---

## 12. Conformance

A resolver conforms to this specification when it meets every requirement below. Each item names the section that defines it, except the few that are stated only here. A conforming resolver MUST:

1. Pull a ModelKit and locate the Agent Definition (section 4): the Kitfile's `agent:` block when there is one, otherwise the document with `kind: AgentDefinition` in its code layer. If neither exists, the ModelKit is not a valid agent ModelKit. If more than one such document exists in the code layer, the resolver MUST reject.
2. Parse the Agent Definition and validate it per section 5: its `spec` in either form, and `apiVersion` and `kind` in file form. Reject a field this specification does not define.
3. Resolve the full `spec.base` chain if present, rejecting circular references.
4. Merge `spec.includes` according to the rules in section 10.1.
5. Reject any attempt to remove or replace a policy module inherited from the base chain (section 10.2).
6. Merge `spec.needs` according to the rules in section 10.3: reject target changes, removal entries, and the loosening of a need a gateway consumer references in the merged chain; report each loosened need; and treat as dangling each need that a gateway consumer references in some definition of the chain and none references in the merged chain, delivering it nowhere, not requiring it, and reporting it.
7. Validate that `spec.framework` is consistent across the composition chain (section 5.1).
8. Flatten the chain, normalize `needs` defaults, resolve every need and configure the modules in the order of section 10.6, rejecting when a required need fails to resolve and skipping an optional one with a warning.
9. Deliver `env`-target needs as environment variables in the agent runtime, except needs withheld under section 6.5, which receive their placeholder or nothing as that section requires. Deliver `file`-target needs as files at their declared `mountPath`, on a non-persistent filesystem.
10. Provide an operator-side configuration mechanism for resolver backends and per-need overrides (section 6.4).
11. Allow policies to inspect the resolved `needs` and `includes` shape, so admission rules can forbid (for example) implicit-env resolution for credential-shaped names.
12. For each `llm` module, register an inference provider accessible to the framework, serving the model locally where the module names one through `source.modelkit`, with the gateway applying the module's `credential` where it declares one (section 8.1).
13. For each `mcp` module, launch or connect the server and make its tools available to the framework, restricted to the module's `tools` allow-list where it declares one (section 8.2 and Appendix A.3).
14. For each `policy` module, load the policy artifact and enforce it at the evaluation points declared by the policy itself.
15. For each `skill` module, take the skill from the referenced ModelKit, or from the `prompts` entry of the ModelKit the module was declared in when the reference is relative (section 7), and mount it where the framework discovers skills (section 8.4).
16. Mount the ModelKit's content layers (code, prompts, model, datasets), whether or not a module names them, so the framework can discover them by convention.
17. Never write resolved `needs` values to disk on the host outside of the agent runtime. Audit logs MUST record the resolver kind that produced each value, but MUST NOT record the value itself.
18. Resolve a bare need from the host environment variable of its name and, when that variable is unset, from the host keychain item with the resolver's documented service name and the need's name as account, recording which of the two answered (section 6.3).
19. Validate `spec.egress` per section 9 on each document and on the merged chain, and merge it according to the rules in section 10.7.
20. Withhold every need a gateway consumer references, and the provider variable names of each `llm` module that declares `credential`, from the agent runtime and local MCP servers, delivering a placeholder where section 6.5 requires one, in the form of Appendix A.1, and report each need it withheld.
21. Apply every object egress entry at the gateway as Appendix A.4 requires, or reject the definition.
22. Skip, and report, each `llm` module whose `credential` names an optional need that failed to resolve, and drop each remote MCP or egress header that references one (sections 8.1, 8.2 and 9.6).
23. Apply every registry entry at the gateway as section 9.7 and Appendix A.5 require, with the login the host's registry login store holds for the registry, never delivering that login to the agent runtime, and report each registry entry for which the store holds no login.
24. Record each value given to the gateway for a gateway consumer, with the consumer and its source, and never the value itself, and withhold the value when that record cannot be written.

A tool that reads a definition without meeting these requirements — a developer tool that runs the agent locally with no gateway, for instance — is not a conforming resolver, however faithfully it reads the definition. What such a tool does with a definition is its own specification's concern.

---

## Appendix A. Gateway and Resolver Mechanics

This appendix is normative. It states what a resolver does, at its gateway and around the processes it launches, to honour the declarations of sections 6, 8 and 9. Section 12 requires it of a conforming resolver. The body of this specification says what a definition declares and when it is valid; this appendix says how a resolver makes those declarations hold.

### A.1 Withholding

The placeholder's form is the resolver's choice, not this specification's: the resolver MUST document it, it MUST be distinct per need, and it MUST be a value the resolver's own gateway recognizes and cannot mistake for a real credential. The placeholder is delivered for each withheld need referenced by a header the gateway applies (section 9.6). Requests to an egress host originate in the agent's own processes, and clients commonly refuse to send a request when their credential is unset; the placeholder lets them send one, and the gateway replaces it before the request leaves. Requests to an LLM provider or a remote MCP server through a gateway consumer originate in the gateway, so nothing in the agent runtime needs a value for them.

For each `llm` module that declares `credential`, every variable name the provider table lists for the module's provider (Appendix A.2) MUST be absent from the agent runtime's environment and from every local MCP server's environment, even where the host environment supplies it, except as a placeholder section 6.5 requires under that name. Otherwise a value the host happens to hold under the provider's conventional name reaches the agent beside the one the gateway applies.

### A.2 Provider table

The **provider table** belongs to the resolver, as the set of provider hosts does (section 9). It has an entry for each LLM provider whose credential the gateway can apply, identified by the provider prefix of `source.model`, and each entry lists the variable names that conventionally carry the provider's credential, which govern withholding (section 6.5). Where a resolver's table has an entry for one of the following providers, the entry MUST list these variable names:

| Provider | Variable names |
|----------|----------------|
| `anthropic` | `ANTHROPIC_API_KEY` |
| `openai` | `OPENAI_API_KEY` |
| `gemini` | `GEMINI_API_KEY`, `GOOGLE_API_KEY` |

Several names in one entry are spellings of the same credential.

### A.3 Local MCP servers

Every successfully resolved need with `target: env`, other than one withheld under section 6.5, MUST be present in a local MCP server's environment when it is launched. Every successfully resolved need with `target: file` MUST be readable by the server at its declared `mountPath`. A resolver that runs the server inside the agent runtime's process boundary satisfies both by inheritance: the server inherits the runtime environment and sees the runtime's filesystem. A resolver that isolates the server from the agent runtime MUST deliver both explicitly, writing file-target needs at the same `mountPath` inside the server's filesystem, under the same non-persistence requirement as section 6.1. Optional needs that fail to resolve are skipped (section 10.6) and impose no delivery requirement.

Where a module declares `tools`, the resolver MUST NOT make any other tool of that server reachable, and SHOULD omit them from what the framework discovers rather than merely refusing the invocation. A tool the agent cannot see is one it cannot attempt, and an agent that can enumerate a withheld tool learns what the server exposes.

Nothing the agent controls may widen the list. In particular, a resolver that mediates MCP through a gateway MUST NOT let a request-level parameter, header, or framework-side configuration file expand what a module declared.

A resolver SHOULD record, at agent start, each module whose tools were narrowed and the names it allowed. Absent that, a tool missing by declaration is indistinguishable from a tool the server has stopped serving, and the two have different remedies.

### A.4 Object egress entries

For each object entry in the resolved set, the gateway MUST:

- terminate connections to the entry's host on port 443;
- set each declared header on every request to the host, replacing any value the client sent under that name, matched without regard to case;
- render each header's template at request time from the value the resolver holds, and encode a `basic` value in the gateway, so that neither the value nor its encoding reaches the agent;
- leave headers the entry does not declare unchanged;
- refuse a request whose `Host` differs from the connection's TLS server name, and refuse a request whose target is in absolute form, so that a client cannot steer the headers to another host behind a shared address;
- refuse a `TRACE` request before applying any header, since the upstream would echo the request, headers included, back to the client;
- forward requests only to addresses resolved for the entry's host before the agent starts;
- verify the upstream server's certificate for the entry's host against the system's trust roots, and close the client connection if verification fails.

### A.5 Registry entries

A registry is more than one host. A client reaches the registry host, the token service that the registry's `WWW-Authenticate` challenge names, and the content hosts that blob downloads redirect to, and only the first is known before the agent starts. The resolver MUST admit the others as the registry's own responses name them, on port 443 only, and MUST hold each admitted host to the rules a declared entry satisfies (sections 9.2 and 9.5): it MUST refuse an address, a wildcard, a provider host, the endpoint host of a remote MCP module, and a host the gateway terminates for another purpose. A host the definition already lists as a plain entry stays plain and receives no login. The resolver MUST record each admission and each refusal in the run's egress record. How it discovers these hosts, which registry names it rewrites to the host clients actually contact, as `docker.io` is reached at `registry-1.docker.io`, and how it bounds admission are its own documentation's concern.

For each registry entry in the resolved set, the gateway MUST sign in to the registry host and its token service with the store's login, as HTTP Basic (RFC 7617), on every request that carries no `Authorization` of its own, and MUST leave a request's own `Authorization` unchanged: a client signs in once at the token service and then presents the bearer token it was issued. The login MUST NOT be sent to a content host, whose redirect carries its own authorization. The resolver MAY obtain the login from the store again during the run, since a credential helper can issue one that expires.

The login MUST NOT reach the agent runtime. The resolver MUST NOT deliver it as an environment variable, as a file or as an entry in the agent runtime's own registry login store, and that store MUST hold no login for the entry's hosts. The audit log records the registry as `oci://<registry>` and never the login (section 12).

Where the store holds no login for the registry, the resolver MUST report it at agent start and the gateway sets no `Authorization`: the agent reaches the registry as an anonymous client, and public content stays reachable. ---
