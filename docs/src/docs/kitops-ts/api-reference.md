# API Reference

All functions are exported from `@kitops/kitops-ts` and return a [`CancellablePromise`](#cancellablepromise). They reject with a string error message containing the exit code and stderr output on failure, or with a `DOMException` named `'AbortError'` if `.cancel()` is called.

```typescript
import { pack, push, pull, /* ... */ } from '@kitops/kitops-ts';
```

## Cancellation

Every function returns a `CancellablePromise<T>`. Call `.cancel()` on the returned value at any time to kill the underlying `kit` process. The promise will reject with a `DOMException` named `'AbortError'`.

```typescript
const op = push('registry.example.com/org/my-model:v1.0.0');

setTimeout(() => op.cancel(), 30_000); // cancel after 30 s

try {
  await op;
} catch (err) {
  if (err instanceof DOMException && err.name === 'AbortError') {
    console.log('Cancelled');
  }
}
```

Since `CancellablePromise` extends `Promise`, existing `await` usage requires no changes — `.cancel()` is opt-in.

---

## Commands

### `init(path?, flags?)`

Scans a directory for ML artifacts and generates a `Kitfile`. Automatically detects models, datasets, code, and docs by file extension.

**Parameters**

| Name | Type | Description |
|---|---|---|
| `path` | `string` | Directory to scan. Defaults to `'.'`. |
| `flags.name` | `string` | Package name to use in the generated Kitfile. |
| `flags.desc` | `string` | Package description. |
| `flags.author` | `string` | Package author. |
| `flags.force` | `boolean` | Overwrite an existing Kitfile. |

**Returns** `CancellablePromise<InitResult>`

```typescript
type InitResult = {
  path: string;        // absolute path to the scanned directory
  kitfilePath: string; // absolute path to the generated Kitfile
}
```

**Example**

```typescript
const result = await init('./my-model', {
  name: 'my-model',
  desc: 'Sentiment analysis model',
  author: 'ML Team',
  force: true,
});
console.log(result.kitfilePath);
```

---

### `info(path, flags?)`

Returns the parsed `Kitfile` for a ModelKit. The result also carries a non-enumerable `_raw` property with the original YAML string.

**Parameters**

| Name | Type | Description |
|---|---|---|
| `path` | `string` | Full ModelKit reference (`registry/repository[:tag\|@digest]`). |
| `flags.remote` | `boolean` | Inspect from the registry without pulling first. |
| `flags.filter` | `FilterFlag` | Limit output to specific layers or paths. |

**Returns** `CancellablePromise<Kitfile>`

**Example**

```typescript
const kitfile = await info('registry.example.com/org/my-model:v1.0.0');
console.log(kitfile.package?.name);
console.log(kitfile.model?.path);
```

---

### `inspect(path, flags?)`

Returns the full OCI manifest and parsed Kitfile for a ModelKit.

**Parameters**

| Name | Type | Description |
|---|---|---|
| `path` | `string` | Full ModelKit reference (`registry/repository[:tag\|@digest]`). |
| `flags.remote` | `boolean` | Inspect directly from the registry without pulling first. |

**Returns** `CancellablePromise<InspectResult>`

```typescript
type InspectResult = {
  digest: string;
  cliVersion: string;
  kitfile: Kitfile;
  manifest: Manifest;
}
```

**Example**

```typescript
const result = await inspect('registry.example.com/org/my-model:v1.0.0', { remote: true });
console.log(result.digest);
console.log(result.manifest.layers);
```

---

### `pack(directory?, flags?)`

Packages a directory containing a `Kitfile` into a local ModelKit.

**Parameters**

| Name | Type | Description |
|---|---|---|
| `directory` | `string` | Directory to pack. Defaults to `'.'`. |
| `flags.tag` | `string` | Tag to assign to the packed ModelKit. |
| `flags.file` | `string` | Path to a Kitfile other than `./Kitfile`. |
| `flags.compression` | `string` | Compression algorithm (e.g. `'zstd'`, `'gzip'`). |
| `flags.useModelPack` | `boolean` | Use model-pack format. |

**Returns** `CancellablePromise<void>`

**Example**

```typescript
await pack('.', {
  tag: 'registry.example.com/org/my-model:v1.0.0',
  compression: 'zstd',
});
```

---

### `unpack(destination, flags?)`

Extracts a ModelKit into a directory. Use `flags.filter` to pull out specific layers or paths.

**Parameters**

| Name | Type | Description |
|---|---|---|
| `destination` | `string` | Directory to unpack into. |
| `flags.dir` | `string` | Override the working directory. |
| `flags.overwrite` | `boolean` | Overwrite existing files. |
| `flags.ignoreExisting` | `boolean` | Skip files that already exist instead of failing. |
| `flags.filter` | `FilterFlag` | Limit extraction to specific layers or paths. |

**Returns** `CancellablePromise<void>`

**Example**

```typescript
// Extract everything
await unpack('./output');

// Model layer only
await unpack('./output', { filter: 'model' });

// A specific dataset and the full docs layer
await unpack('./output', { filter: 'datasets:validation,docs' });
```

---

### `list(repository?, flags?)`

Lists ModelKits. Omit the argument to list local storage; pass a registry/repository to list remote tags.

**Parameters**

| Name | Type | Description |
|---|---|---|
| `repository` | `string` | Optional registry/repository to list remotely. |
| `flags.format` | `'table' \| 'json' \| 'template'` | Output format. Defaults to `'table'`. |

**Returns** `CancellablePromise<ModelKit[] | string>`

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

**Example**

```typescript
const local = await list();
const remote = await list('registry.example.com/org/my-model');

for (const kit of local) {
  console.log(kit.repository, kit.tag, kit.size);
}
```

---

### `push(source, destination?, flags?)`

Pushes a ModelKit to a registry. Optionally provide `destination` to push under a different reference (e.g. promote from staging to production).

**Parameters**

| Name | Type | Description |
|---|---|---|
| `source` | `string` | Local ModelKit reference to push. |
| `destination` | `string` | Optional target reference (different registry or tag). |
| `flags.tlsVerify` | `boolean` | Verify the server's TLS certificate. |
| `flags.tlsCert` | `string` | Path to a client TLS certificate file. |

**Returns** `CancellablePromise<void>`

**Example**

```typescript
await push('registry.example.com/org/my-model:v1.0.0');

// Push to a different registry under a different tag
await push('staging.example.com/my-model:rc1', 'registry.example.com/org/my-model:v1.0.0');
```

---

### `pull(reference, flags?)`

Pulls a ModelKit from a registry into local storage.

**Parameters**

| Name | Type | Description |
|---|---|---|
| `reference` | `string` | Full ModelKit reference including registry, repository, and tag/digest. |
| `flags.tlsVerify` | `boolean` | Verify the server's TLS certificate. |
| `flags.tlsCert` | `string` | Path to a client TLS certificate file. |

**Returns** `CancellablePromise<void>`

**Example**

```typescript
await pull('registry.example.com/org/my-model:v1.0.0');
```

---

### `login(registry, username, password, flags?)`

Authenticates with a registry. The password is passed via **stdin**, keeping it out of the process list. Preferred for CI and automated workflows.

**Parameters**

| Name | Type | Description |
|---|---|---|
| `registry` | `string` | Registry hostname (e.g. `'ghcr.io'`). |
| `username` | `string` | Registry username. |
| `password` | `string` | Registry password or token. |
| `flags.tlsVerify` | `boolean` | Verify the server's TLS certificate. |

**Returns** `CancellablePromise<void>`

**Example**

```typescript
await login(
  'registry.example.com',
  process.env.REGISTRY_USER!,
  process.env.REGISTRY_PASS!,
);
```

---

### `loginUnsafe(registry, username, password, flags?)`

Same as `login` but passes the password as a CLI argument, making it visible in the process list. Use only in environments where the process list is not observable (e.g. isolated containers).

**Parameters**

| Name | Type | Description |
|---|---|---|
| `registry` | `string` | Registry hostname. |
| `username` | `string` | Registry username. |
| `password` | `string` | Registry password or token (passed as a CLI flag — visible in `ps`). |
| `flags.tlsVerify` | `boolean` | Verify the server's TLS certificate. |

**Returns** `CancellablePromise<void>`

**Example**

```typescript
await loginUnsafe('registry.example.com', 'user', 'pass');
```

---

### `logout(registry)`

Removes stored credentials for a registry.

**Parameters**

| Name | Type | Description |
|---|---|---|
| `registry` | `string` | Registry hostname to log out from. |

**Returns** `CancellablePromise<void>`

**Example**

```typescript
await logout('registry.example.com');
```

---

### `tag(source, destination)`

Assigns a new tag to an existing local ModelKit without re-packing. Use `push` afterward to publish it.

**Parameters**

| Name | Type | Description |
|---|---|---|
| `source` | `string` | Existing ModelKit reference. |
| `destination` | `string` | New reference to assign. |

**Returns** `CancellablePromise<void>`

**Example**

```typescript
await tag('registry.example.com/org/my-model:rc1', 'registry.example.com/org/my-model:v1.0.0');
```

---

### `remove(path, flags?)`

Removes a ModelKit from local storage or a remote registry.

**Parameters**

| Name | Type | Description |
|---|---|---|
| `path` | `string` | Full ModelKit reference to remove. |
| `flags.force` | `boolean` | Skip confirmation prompt. |
| `flags.all` | `boolean` | Remove all locally cached ModelKits (ignores `path`). |
| `flags.remote` | `boolean` | Remove from the registry instead of local storage. |

**Returns** `CancellablePromise<void>`

**Example**

```typescript
// Remove a specific tag from local storage
await remove('registry.example.com/org/my-model:v0.9.0');

// Remove all locally cached ModelKits
await removeAll();

// Force-remove from the remote registry
await remove('registry.example.com/org/my-model:v0.9.0', { force: true, remote: true });
```

---

### `removeAll(flags?)`

Removes all locally cached ModelKits. Equivalent to `remove('', { all: true })` but without requiring a dummy path argument.

To remove all kits from a remote registry, use `remove('', { all: true, remote: true })` instead.

**Parameters**

| Name | Type | Description |
|---|---|---|
| `flags.force` | `boolean` | Skip confirmation prompt. |

**Returns** `CancellablePromise<void>`

**Example**

```typescript
await removeAll();
```

---

### `diff(reference1, reference2, flags?)`

Compares two ModelKits and returns a structured diff of their layers.

**Parameters**

| Name | Type | Description |
|---|---|---|
| `reference1` | `string` | First ModelKit reference. Prefix with `local://` or `remote://` to be explicit. |
| `reference2` | `string` | Second ModelKit reference. |
| `flags.plainHttp` | `boolean` | Allow plain HTTP connections to the registry. |

**Returns** `CancellablePromise<DiffResult>`

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

type DiffLayerEntry = {
  type: string;
  digest: string;
  size: string;
}
```

**Example**

```typescript
const result = await diff(
  'registry.example.com/org/my-model:v1.0.0',
  'registry.example.com/org/my-model:v1.1.0',
);
console.log('Shared layers:', result.sharedLayers.length);
console.log('New layers:', result.uniqueToKit2);
```

---

### `version()`

Returns version information for the installed `kit` binary.

**Returns** `CancellablePromise<VersionResult>`

```typescript
type VersionResult = {
  version: string;
  commit: string;
  built: string;
  goVersion: string;
}
```

**Example**

```typescript
const { version, commit, built, goVersion } = await version();
console.log(`kit ${version} (${commit})`);
```

---

### `kit(command, args, stdin?, options?)`

Low-level escape hatch for running any `kit` subcommand directly. Useful when the higher-level wrappers don't expose a flag you need.

**Parameters**

| Name | Type | Description |
|---|---|---|
| `command` | `KitCommand` | The `kit` subcommand to run (e.g. `'pack'`, `'push'`). |
| `args` | `string[]` | Arguments to pass to the subcommand. |
| `stdin` | `string` | Optional data to write to stdin before closing it (e.g. a password). |
| `options.cwd` | `string` | Working directory for the spawned process. |
| `options.env` | `Record<string, string>` | Extra environment variables merged on top of `process.env`. |

**Returns** `CancellablePromise<ExecResult>`

```typescript
type ExecResult = {
  stdout: string;
  stderr: string;
  exitCode: number;
}
```

**Example**

```typescript
// Run a subcommand in a specific directory
const result = await kit('pack', ['.', '--tag', 'my-model:latest'], undefined, { cwd: '/path/to/project' });
console.log(result.stdout);

// Pass credentials via stdin
const result = await kit('login', ['registry.example.com', '--username', 'user'], process.env.REGISTRY_PASS!);
```
