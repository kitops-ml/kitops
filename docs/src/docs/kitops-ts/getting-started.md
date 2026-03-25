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

## Error handling

All functions return Promises and reject with a descriptive string on failure. Wrap calls in `try/catch`:

```typescript
try {
  await push('registry.example.com/org/my-model:v1.0.0');
} catch (err) {
  console.error(err); // "Kit command failed with exit code 1: ..."
  process.exit(1);
}
```

## Environment variables

| Variable | Description |
|---|---|
| `KITOPS_CLI_PATH` | Full path to the `kit` binary. Defaults to `kit` (resolved via `PATH`). |
