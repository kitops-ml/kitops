# Contributing

Thank you for your interest in contributing to kitops-ts!

For the full contribution guide, see [CONTRIBUTING.md](https://github.com/kitops-ml/kitops-ts/blob/main/CONTRIBUTING.md) in the repository root.

## Quick reference

### Setup

```bash
pnpm install
```

### Development

```bash
pnpm dev        # watch mode — recompiles on change
pnpm build      # compile TypeScript to dist/
pnpm typecheck  # type-check without emitting files
pnpm test       # run the test suite
```

### Before opening a PR

- Run `pnpm typecheck` and `pnpm test` — both must pass
- Keep changes focused; one concern per PR
- If you're adding a new command wrapper, add a corresponding spec file under `src/commands/__tests__/`

### Project structure

```
src/
  commands/   # one file per kit subcommand
  core/       # exec utilities (spawn, arg prep, output parsers)
  types/      # TypeScript type definitions
examples/     # runnable JS examples
docs/         # this documentation
```

### Reporting bugs

Open an issue at [https://github.com/kitops-ml/kitops-ts/issues](https://github.com/kitops-ml/kitops-ts/issues). Include:

- The `kit` binary version (`kit version`)
- Your Node.js version (`node --version`)
- A minimal reproduction case
