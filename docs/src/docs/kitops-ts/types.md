# Type Definitions

All types are exported from `@kitops/kitops-ts`.

```typescript
import type { Kitfile, ModelKit, FilterFlag, /* ... */ } from '@kitops/kitops-ts';
```

---

## Core Types

### `Layer`

The content layer kinds a ModelKit can contain.

```typescript
type Layer =
  | 'model'
  | 'datasets'
  | 'code'
  | 'docs'
  | 'prompts'
```

---

### `FilterFlag`

A filter expression used by [`unpack`](../api-reference.md#unpackdestination-flags) and [`info`](../api-reference.md#inforepository-tag-flags) to select specific layers or paths within a layer.

```typescript
type FilterFlag = string  // see accepted formats below
```

Accepted formats:

| Example | Meaning |
|---|---|
| `'model'` | Entire model layer |
| `'datasets:training'` | Named dataset |
| `'docs:./README.md'` | Single file within the docs layer |
| `'model,datasets:validation'` | Comma-separated — multiple layers |

---

### `KitCommand`

Union of all subcommands accepted by the `kit` binary.

```typescript
type KitCommand =
  | 'diff'
  | 'info'
  | 'inspect'
  | 'init'
  | 'list'
  | 'login'
  | 'logout'
  | 'pack'
  | 'pull'
  | 'push'
  | 'remove'
  | 'tag'
  | 'unpack'
  | 'version'
```

---

### `ExecResult`

Raw output from a spawned `kit` process. Returned by the low-level `runCommand` helper.

```typescript
type ExecResult = {
  stdout: string;
  stderr: string;
  exitCode: number;
}
```

---

### `ModelKit`

A single entry returned by [`list`](../api-reference.md#listrepository).

```typescript
type ModelKit = {
  repository: string;
  tag: string;
  maintainer: string;
  name: string;
  size: string;
  digest: string;
}
```

| Property | Type | Description |
|---|---|---|
| `repository` | `string` | Registry/repository path. |
| `tag` | `string` | Image tag. |
| `maintainer` | `string` | Maintainer field from the Kitfile package. |
| `name` | `string` | Name field from the Kitfile package. |
| `size` | `string` | Human-readable size of the ModelKit. |
| `digest` | `string` | OCI digest of the manifest (`sha256:…`). |

---

## Flag Types

Flags objects are passed as the last (optional) argument to each command function.

### `InitFlags`

```typescript
type InitFlags = {
  name?: string;
  desc?: string;
  author?: string;
  force?: boolean;
}
```

| Property | Type | Description |
|---|---|---|
| `name` | `string` | Package name to use in the generated Kitfile. |
| `desc` | `string` | Package description. |
| `author` | `string` | Package author. |
| `force` | `boolean` | Overwrite an existing Kitfile. |

---

### `InfoFlags`

```typescript
type InfoFlags = {
  remote?: boolean;
  filter?: FilterFlag;
}
```

| Property | Type | Description |
|---|---|---|
| `remote` | `boolean` | Inspect from the registry without pulling first. |
| `filter` | `FilterFlag` | Limit output to specific layers or paths. |

---

### `InspectFlags`

```typescript
type InspectFlags = {
  remote?: boolean;
}
```

| Property | Type | Description |
|---|---|---|
| `remote` | `boolean` | Inspect from the registry without pulling first. |

---

### `UnpackFlags`

```typescript
type UnpackFlags = {
  dir?: string;
  overwrite?: boolean;
  ignoreExisting?: boolean;
  filter?: FilterFlag;
}
```

| Property | Type | Description |
|---|---|---|
| `dir` | `string` | Destination directory. Defaults to the current directory. |
| `overwrite` | `boolean` | Overwrite existing files. |
| `ignoreExisting` | `boolean` | Skip files that already exist. |
| `filter` | `FilterFlag` | Unpack only specific layers or paths. |

---

### `PackFlags`

```typescript
type PackFlags = {
  file?: string;
  tag?: string;
  compression?: string;
  useModelPack?: boolean;
}
```

| Property | Type | Description |
|---|---|---|
| `file` | `string` | Path to the Kitfile. Defaults to `./Kitfile`. |
| `tag` | `string` | Tag to assign to the packed ModelKit (e.g. `registry.example.com/org/model:v1`). |
| `compression` | `string` | Compression algorithm to use (e.g. `'gzip'`, `'zstd'`). |
| `useModelPack` | `boolean` | Use model-pack format for the model layer. |

---

### `RemoveFlags`

```typescript
type RemoveFlags = {
  force?: boolean;
  all?: boolean;
  remote?: boolean;
}
```

| Property | Type | Description |
|---|---|---|
| `force` | `boolean` | Skip confirmation prompt. |
| `all` | `boolean` | Remove all tags for the repository. |
| `remote` | `boolean` | Remove from the registry instead of local storage. |

---

### `ListFlags`

```typescript
type ListFlags = {
  format: 'table' | 'json' | 'template';
}
```

| Property | Type | Description |
|---|---|---|
| `format` | `'table' \| 'json' \| 'template'` | Output format for the underlying `kit list` call. |

---

### `DiffFlags`

```typescript
type DiffFlags = {
  plainHttp?: boolean;
}
```

| Property | Type | Description |
|---|---|---|
| `plainHttp` | `boolean` | Use plain HTTP instead of HTTPS when communicating with the registry. |

---

### `TLSFlags`

Shared TLS options extended by commands that interact with a registry.

```typescript
type TLSFlags = {
  tlsVerify?: boolean;
  tlsCert?: string;
}
```

| Property | Type | Description |
|---|---|---|
| `tlsVerify` | `boolean` | Verify the server's TLS certificate. |
| `tlsCert` | `string` | Path to a client TLS certificate file. |

---

## Result Types

### `InitResult`

Returned by [`init`](../api-reference.md#initpath-flags).

```typescript
type InitResult = {
  path: string;        // absolute path to the scanned directory
  kitfilePath: string; // absolute path to the generated Kitfile
}
```

---

### `InspectResult`

Returned by [`inspect`](../api-reference.md#inspectrepository-tag-flags).

```typescript
type InspectResult = {
  digest: string;
  cliVersion: string;
  kitfile: Kitfile;
  manifest: Manifest;
}
```

| Property | Type | Description |
|---|---|---|
| `digest` | `string` | OCI digest of the manifest. |
| `cliVersion` | `string` | Version of the `kit` CLI that created the ModelKit. |
| `kitfile` | [`Kitfile`](#kitfile) | Parsed Kitfile contents. |
| `manifest` | [`Manifest`](#manifest) | Full OCI image manifest. |

---

### `VersionResult`

Returned by [`version`](../api-reference.md#version).

```typescript
type VersionResult = {
  version: string;
  commit: string;
  built: string;
  goVersion: string;
}
```

| Property | Type | Description |
|---|---|---|
| `version` | `string` | Semantic version of the `kit` binary. |
| `commit` | `string` | Git commit hash the binary was built from. |
| `built` | `string` | Build timestamp. |
| `goVersion` | `string` | Go runtime version used to compile the binary. |

---

### `DiffResult`

Returned by [`diff`](../api-reference.md#diffmodelkit1-modelkit2-flags).

```typescript
type DiffResult = {
  modelKit1: string;
  modelKit2: string;
  configsDiffer: boolean;
  config1Digest?: string;
  config2Digest?: string;
  annotationsIdentical: boolean;
  sharedLayers: DiffLayerEntry[];
  uniqueToKit1: DiffLayerEntry[];
  uniqueToKit2: DiffLayerEntry[];
}
```

| Property | Type | Description |
|---|---|---|
| `modelKit1` | `string` | Reference to the first ModelKit. |
| `modelKit2` | `string` | Reference to the second ModelKit. |
| `configsDiffer` | `boolean` | Whether the Kitfile configs are different. |
| `config1Digest` | `string` | Digest of the first ModelKit's config, if configs differ. |
| `config2Digest` | `string` | Digest of the second ModelKit's config, if configs differ. |
| `annotationsIdentical` | `boolean` | Whether the OCI annotations are identical. |
| `sharedLayers` | [`DiffLayerEntry[]`](#difflayerentry) | Layers present in both ModelKits. |
| `uniqueToKit1` | [`DiffLayerEntry[]`](#difflayerentry) | Layers only in the first ModelKit. |
| `uniqueToKit2` | [`DiffLayerEntry[]`](#difflayerentry) | Layers only in the second ModelKit. |

---

### `DiffLayerEntry`

A single layer entry within a [`DiffResult`](#diffresult).

```typescript
type DiffLayerEntry = {
  type: string;
  digest: string;
  size: string;
}
```

| Property | Type | Description |
|---|---|---|
| `type` | `string` | Layer media type. |
| `digest` | `string` | OCI digest of the layer blob. |
| `size` | `string` | Human-readable size of the layer. |

---

## Kitfile Types

These types mirror the Kitfile schema. They are returned by [`info`](../api-reference.md#inforepository-tag-flags) and [`inspect`](../api-reference.md#inspectrepository-tag-flags), and can also be used when constructing Kitfiles programmatically.

### `Kitfile`

Typed representation of a Kitfile.

```typescript
type Kitfile = {
  manifestVersion: string;
  package?: Package;
  model?: Model;
  code?: Code[];
  datasets?: Dataset[];
  docs?: Doc[];
  prompts?: Prompt[];
}
```

| Property | Type | Description |
|---|---|---|
| `manifestVersion` | `string` | Kitfile schema version. See the [KitOps Kitfile spec](https://kitops.org/docs/kitfile/kf-spec.html) for the authoritative value. |
| `package` | [`Package`](#package) | Package-level metadata. |
| `model` | [`Model`](#model) | Model artifact definition. |
| `code` | [`Code[]`](#code) | Source code artifact definitions. |
| `datasets` | [`Dataset[]`](#dataset) | Dataset artifact definitions. |
| `docs` | [`Doc[]`](#doc) | Documentation artifact definitions. |
| `prompts` | [`Prompt[]`](#prompt) | Prompt artifact definitions. |

> When returned by `info()`, the object also carries a non-enumerable `_raw` string property containing the original YAML source.

---

### `Package`

Package-level metadata in a Kitfile.

```typescript
type Package = {
  name?: string;
  authors?: string[];
  version?: string;
  license?: string;
  description?: string;
}
```

---

### `Model`

Model artifact definition in a Kitfile.

```typescript
interface Model {
  name?: string;
  path?: string;
  parts?: ModelPart[];
  license?: string;
  description?: string;
  format?: string;
  parameters?: Record<string, string | number>;
}
```

| Property | Type | Description |
|---|---|---|
| `name` | `string` | Display name for the model. |
| `path` | `string` | Path to the model file or directory. |
| `parts` | [`ModelPart[]`](#modelpart) | Sub-parts of a composite model (e.g. split weight files). |
| `license` | `string` | SPDX license identifier. |
| `description` | `string` | Human-readable description. |
| `format` | `string` | Model format (e.g. `'gguf'`, `'safetensors'`). |
| `parameters` | `Record<string, string \| number>` | Arbitrary model parameters (e.g. context length, quantization). |

---

### `ModelPart`

A sub-part of a composite model.

```typescript
interface ModelPart {
  path?: string;
  type?: string;
  license?: string;
}
```

---

### `Dataset`

Dataset artifact definition in a Kitfile.

```typescript
interface Dataset {
  name?: string;
  path?: string;
  description?: string;
  license?: string;
  parameters?: unknown;
}
```

---

### `Code`

Source code artifact definition in a Kitfile.

```typescript
interface Code {
  path: string;
  description: string;
}
```

---

### `Doc`

Documentation artifact definition in a Kitfile.

```typescript
interface Doc {
  path: string;
  description: string;
}
```

---

### `Prompt`

Prompt artifact definition in a Kitfile.

```typescript
interface Prompt {
  path?: string;
  description?: string;
}
```

---

## Manifest Types

These types represent the OCI image manifest returned by [`inspect`](../api-reference.md#inspectrepository-tag-flags).

### `Manifest`

The OCI image manifest for a ModelKit. `config` holds the Kitfile descriptor; `layers` holds one descriptor per content layer.

```typescript
interface Manifest {
  schemaVersion: number;
  mediaType: string;
  artifactType?: string;
  config: ManifestDescriptor;
  layers: ManifestDescriptor[];
  annotations: ManifestAnnotations;
}
```

---

### `ManifestDescriptor`

OCI content descriptor — identifies a blob by media type, digest, and size.

```typescript
type ManifestDescriptor = {
  mediaType: string;
  digest: string;
  size: number;
  urls?: string[];
  annotations?: ManifestAnnotations;
  data?: string;
  artifactType?: string;
}
```

| Property | Type | Description |
|---|---|---|
| `mediaType` | `string` | IANA media type of the referenced content. |
| `digest` | `string` | Content-addressed digest (`sha256:…`). |
| `size` | `number` | Size of the blob in bytes. |
| `urls` | `string[]` | Optional list of locations the content can be downloaded from. |
| `annotations` | [`ManifestAnnotations`](#manifestannotations) | Arbitrary key/value metadata. |
| `data` | `string` | Optional inline base64-encoded content (for small blobs). |
| `artifactType` | `string` | Media type of the artifact this descriptor refers to. |

---

### `ManifestAnnotations`

Arbitrary string key/value pairs attached to an OCI descriptor or manifest.

```typescript
type ManifestAnnotations = {
  [key: string]: string;
}
```
