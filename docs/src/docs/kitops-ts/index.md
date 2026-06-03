# kitops-ts

**kitops-ts** is a TypeScript/Node.js SDK for the [KitOps](https://kitops.org) CLI. It provides a type-safe functional API for packing, pushing, pulling, and inspecting ModelKits — without having to shell out manually.

Similar to [pykitops](https://github.com/kitops-ml/pykitops) but for TypeScript/JavaScript.

## What is KitOps?

[KitOps](https://kitops.org) is an open source project that packages ML models, datasets, code, and configuration into a portable, versioned artifact called a **ModelKit**. ModelKits are stored in standard OCI registries (like GHCR, Docker Hub, or any private registry), making them easy to share, reproduce, and deploy.

## What does kitops-ts do?

kitops-ts wraps the `kit` CLI binary with a fully-typed Node.js API. Instead of writing shell scripts to call `kit pack`, `kit push`, etc., you write TypeScript:

```typescript
import { login, pack, push } from '@kitops/kitops-ts';

await login('ghcr.io', process.env.REGISTRY_USER!, process.env.REGISTRY_PASS!);
await pack('.', { tag: 'ghcr.io/my-org/my-model:v1.0.0' });
await push('ghcr.io/my-org/my-model:v1.0.0');
```

Every function returns a `CancellablePromise` — a standard `Promise` extended with a `.cancel()` method that kills the underlying `kit` process at any time. Existing `await` code works without changes; cancellation is purely opt-in.

```typescript
const op = push('ghcr.io/my-org/my-model:v1.0.0');
setTimeout(() => op.cancel(), 30_000); // cancel if it takes too long
await op;
```

## Related

- [KitOps](https://kitops.org) — The ModelKit standard and CLI
- [pykitops](https://github.com/kitops-ml/pykitops) — Python SDK for KitOps
- [kitops-ts on npm](https://www.npmjs.com/package/@kitops/kitops-ts)
- [GitHub repository](https://github.com/kitops-ml/kitops-ts)
