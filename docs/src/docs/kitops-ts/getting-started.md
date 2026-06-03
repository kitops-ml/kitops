# Getting Started

## Requirements

- **Node.js >= 23**
- **KitOps CLI** installed and available on `PATH`
  — [Installation instructions](/docs/cli/installation/)
  — Or set the `KITOPS_CLI_PATH` environment variable to the full path of the `kit` binary

## Installation

```bash
npm install @kitops/kitops-ts
# or
pnpm add @kitops/kitops-ts
```

## Quick start

The following example logs in to a registry, packs the current directory into a ModelKit, and pushes it.

```typescript
import { login, pack, push } from '@kitops/kitops-ts';

// Authenticate — password is passed via stdin to keep it out of the process list
await login(
  'registry.example.com',
  process.env.REGISTRY_USER!,
  process.env.REGISTRY_PASS!,
);

// Pack the current directory using the Kitfile found there
await pack('.', { tag: 'registry.example.com/org/my-model:v1.0.0' });

// Push the ModelKit to the registry
await push('registry.example.com/org/my-model:v1.0.0');
```

## Cancellation

Every function returns a `CancellablePromise<T>` — a standard `Promise` extended with a `.cancel()` method. Calling `.cancel()` kills the underlying `kit` process and rejects the promise with a `DOMException` named `'AbortError'`.

```typescript
import { pull } from '@kitops/kitops-ts';

const op = pull('registry.example.com/org/my-model:v1.0.0');

// Cancel if the pull takes more than 60 seconds
const timeout = setTimeout(() => op.cancel(), 60_000);

try {
  await op;
} catch (err) {
  if (err instanceof DOMException && err.name === 'AbortError') {
    console.log('Pull was cancelled');
  } else {
    throw err; // propagate real errors
  }
} finally {
  clearTimeout(timeout);
}
```

`CancellablePromise` extends `Promise`, so all existing `await` code works without any changes — `.cancel()` is purely opt-in.

## Error handling

All functions return `CancellablePromise`s and reject with a descriptive string on failure. Wrap calls in `try/catch`:

```typescript
try {
  await push('registry.example.com/org/my-model:v1.0.0');
} catch (err) {
  if (err instanceof DOMException && err.name === 'AbortError') {
    console.log('Operation was cancelled');
  } else {
    console.error(err); // "Kit command failed with exit code 1: ..."
    process.exit(1);
  }
}
```

## Environment variables

| Variable | Description |
|---|---|
| `KITOPS_CLI_PATH` | Full path to the `kit` binary. Defaults to `kit` (resolved via `PATH`). |
