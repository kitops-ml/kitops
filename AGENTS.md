
# Agent Instructions

## Code Generation Instructions

1. **Minimum necessary changes only**
   - Change only what's explicitly required. No refactoring adjacent code, no adding comments/docs, no "improvements" beyond scope.
   - Read existing code before modifying it. Follow existing patterns in the file.

2. **Review marker requirement (MANDATORY)**
   After modifying ANY source file, append this marker as the LAST line:

   - Go files (.go): `// AGENT_MODIFIED: Human review required before merge`
   - Markdown (.md): `<!-- AGENT_MODIFIED: Human review required before merge -->`
   - YAML (.yaml, .yml): `# AGENT_MODIFIED: Human review required before merge`
   - Shell (.sh): `# AGENT_MODIFIED: Human review required before merge`
   - Dockerfile: `# AGENT_MODIFIED: Human review required before merge`

   DO NOT add markers to:
   - Binary files
   - Generated files (*.pb.go, go.sum, package-lock.json)
   - Vendored dependencies (/vendor/)
   - Files you only read but didn't modify

   Humans will remove these markers as they review each file. PRs with remaining markers will be rejected by CI.

3. **License header requirement (MANDATORY for Go files)**
   Every `.go` file must start with the Apache 2.0 license header. CI enforces this via `addlicense` and will block the PR if missing.

   ```go
   // Copyright 2024 The KitOps Authors.
   //
   // Licensed under the Apache License, Version 2.0 (the "License");
   // you may not use this file except in compliance with the License.
   // You may obtain a copy of the License at
   //
   //     http://www.apache.org/licenses/LICENSE-2.0
   //
   // Unless required by applicable law or agreed to in writing, software
   // distributed under the License is distributed on an "AS IS" BASIS,
   // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   // See the License for the specific language governing permissions and
   // limitations under the License.
   //
   // SPDX-License-Identifier: Apache-2.0
   ```

   When creating new `.go` files, add this header before the `package` declaration. To fix missing headers in bulk:

   ```bash
   go install github.com/google/addlicense@latest
   addlicense -s -l apache -c "The KitOps Authors." $(find . -name '*.go')
   ```

## Code Review Instructions

### Pre-Review Check (MANDATORY)

Before reviewing code, search the entire PR for `AGENT_MODIFIED`. If any markers are found, reject immediately — all modified files must be human-reviewed and markers removed before merge.

### Review Principles

- **Evidence over speculation**: Every issue must cite specific code locations and concrete failure conditions. No hypothetical "this could be a problem" without a demonstrable trigger.
- **Disprove yourself first**: Before reporting an issue, search for mitigations — check callers, defensive code in parent functions, and similar patterns elsewhere in the codebase.
- **Compare intent vs. outcome**: Read the PR description. Flag mismatches between stated goals and actual changes.
- **Severity requires confidence**: Only mark critical if you have high confidence AND it breaks core functionality or security. Uncertain issues are suggestions, not blockers.

### KitOps-Specific Review Concerns

- **Path traversal in tar layers**: Verify `filesystem.VerifySubpath()` is used. Check that archive entries can't escape the context/unpack directory via `../` or absolute paths.
- **CWD assumptions**: Code in pack/unpack/import runs after `os.Chdir` — verify paths are correct relative to the shifted working directory.
- **OCI spec compliance**: Layers, manifests, and media types must conform to the OCI and ModelKit specification.
- **Registry credential handling**: Auth tokens and credentials in `pkg/lib/network/` must not leak into logs or error messages.
- **Symlink safety**: File operations in `pkg/lib/filesystem/` must not follow symlinks outside the context directory.

### Output

Categorize findings as **Critical** (must fix — breaks functionality or security), **Important** (should fix), or **Suggestion** (author's discretion). End with one of: APPROVE, REQUEST CHANGES, or COMMENT.

## Project Overview

KitOps is a CNCF open standards project for packaging, versioning, and securely sharing AI/ML projects. Built on the OCI standard, it serves as the reference implementation of the CNCF's ModelPack specification for vendor-neutral AI/ML interchange format.

## Development Commands

### Go CLI (Main Project)

```bash

# Build the CLI
go build -o kit .

# Run tests
go test ./...

# Run specific package tests
go test ./pkg/cmd/pack -v

# Install locally  
go install .
```

### Frontend Dev Mode UI

```bash
cd frontend/dev-mode
pnpm install
pnpm dev          # Development server
pnpm build        # Production build  
pnpm type-check   # TypeScript checking
pnpm lint         # ESLint with auto-fix
```

### Documentation Site

```bash
cd docs
pnpm install
pnpm docs:dev     # Development server
pnpm docs:build   # Production build
pnpm docs:preview # Preview built docs
```

## Architecture Overview

### Core CLI Structure

- **Entry point**: `main.go` → `cmd/root.go` using Cobra framework
- **Commands**: Individual CLI commands in `pkg/cmd/{command}/` (pack, unpack, push, pull, tag, list, inspect, etc.)
- **Libraries**: Core functionality in `pkg/lib/`:
  - `kitfile/`: Kitfile manifest parsing, validation, and generation
  - `repo/local/` & `repo/remote/`: Repository management for local storage and OCI registries
  - `harness/`: LLM integration with llamafile for local inference
  - `filesystem/`: File operations, caching, and ignore patterns
  - `network/`: Authentication and network operations

### Key Concepts

- **ModelKit**: Immutable OCI-compatible bundles containing code, models, datasets, and metadata
- **Kitfile**: YAML manifest format defining ModelKit contents (spec in `pkg/artifact/kitfile.md`)
- **Repository Management**: Uses OCI layout for local storage, compatible with container registries

### Frontend Components

- **Dev Mode UI**: Vue 3 + TypeScript SPA for local LLM interaction in `frontend/dev-mode/`
- **Documentation**: VitePress-based site in `docs/` with comprehensive guides and API reference

### CLI Command Pattern

All commands follow consistent structure:

- Command definition: `pkg/cmd/{command}/cmd.go`
- Main logic: `pkg/cmd/{command}/{command}.go`
- Options/configuration in separate files when complex

### CWD Switching During Pack/Unpack (Implementation Gotcha)

Pack, unpack, and import all change the process working directory via `os.Chdir` to simulate `tar -C` semantics. This ensures relative paths inside OCI tar layers are correct (relative to the context directory, not wherever the user ran `kit` from).

**Where it happens:**

- `pkg/cmd/pack/cmd.go` — changes to `contextDir`, **not restored** (CLI exits after)
- `pkg/lib/filesystem/unpack/core.go` — changes to `UnpackDir`, **restored via defer**
- `pkg/cmd/kitimport/util.go` — changes to `contextDir`, **restored via defer** (required on Windows to allow temp dir cleanup)

**Rules when modifying pack/unpack/import code:**

1. Any code running **after** `os.Chdir` in pack operates in the context directory, not the original cwd
2. If calling pack logic from a library context (not a CLI command that exits), you **must** save and restore cwd
3. New commands that reuse packing internals must account for the cwd shift
4. `filesystem.VerifySubpath()` validates that paths stay within bounds after the cwd change — do not bypass it

### Testing Structure

- Unit tests: `*_test.go` files alongside source
- Integration tests: `testing/` directory with comprehensive test data in `testing/testdata/`
- Test categories include manifest comparison, kitfile generation, pack/unpack workflows

### Output and Configuration

- All CLI output routed through `pkg/output/` for consistent formatting and logging
- Configuration handled via `pkg/lib/constants/` with environment variable support
- Caching managed in `pkg/lib/filesystem/cache/`
