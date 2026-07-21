---
name: kitops
description: >-
  Package, version, share, and unpack AI/ML models as OCI-standard ModelKits
  with the KitOps `kit` CLI. Use this skill when authoring a Kitfile or running
  kit init, pack, tag, login, push, pull, or unpack.
license: Apache-2.0
---

# Using the KitOps `kit` CLI

KitOps packages models, datasets, code, docs, prompts, and MCP servers together
as a **ModelKit** — an OCI artifact you can store in any OCI registry. A ModelKit
is defined by a **Kitfile** (a YAML manifest) and built with `kit pack`.

This skill documents the real behavior of the `kit` CLI. It was generated from
and ships with a specific `kit` version, so it always matches the installed
binary. Confirm anything not covered here with `kit <command> --help`.

## Core workflow

```
kit init <dir>     # generate a Kitfile for a directory (optional starting point)
kit pack <dir>     # build a ModelKit into local storage
kit tag SRC DST    # add another reference to a local ModelKit
kit login <reg>    # authenticate to a remote registry
kit push SRC       # upload a ModelKit to a registry
kit pull REF       # download a ModelKit into local storage
kit unpack REF     # extract a ModelKit's contents onto disk
```

## Authoring a Kitfile

The Kitfile is a YAML manifest, by default named `Kitfile` at the root of the
context directory. All `path` values are **relative to the context directory**;
absolute paths are rejected. Only local filesystem paths are allowed for `code`,
`datasets`, `docs`, `prompts`, and `mcpServers` paths; `model.path` also accepts
a ModelKit reference. No two layers may declare the same path.

Sections:

- `manifestVersion` (string): use `1.0.0`. Other values produce a warning and
  are treated as `1.0.0`.
- `package`: `name`, `version`, `description`, `authors` (list of strings).
- `model` (object): `name`, `path`, `framework`, `format`, `version`,
  `description`, `license` (SPDX id), `parts` (list of `name`/`path`/`type`,
  e.g. LoRA weights), and `parameters` (arbitrary JSON-compatible YAML).
- `code` (list): each has `path`, `description`, `license`.
- `datasets` (list): each has `name`, `path`, `description`, `license`. A dataset
  may instead reference a remote source via `remotePath` (an `s3://` URL, which
  also requires `remoteHash`, or a ModelKit reference).
- `docs` (list): each has `path`, `description`.
- `prompts` (list): each has `path`, `description`. A prompt directory that
  contains a `SKILL.md` can be installed as an agent skill (see `unpack`).
- `mcpServers` (list): each has `name` (required, unique) and `path` pointing to
  a single `.mcpb` bundle file, plus `description`.

Example:

```yaml
manifestVersion: 1.0.0
package:
  name: mymodel
  version: 1.0.0
  description: Sentiment classifier
  authors: [Jane Doe]
code:
  - path: src/
    description: Training and inference code
    license: Apache-2.0
datasets:
  - name: training-set
    path: data/train.csv
    description: Labeled training data
    license: CC-BY-4.0
model:
  name: sentiment-model
  path: models/model.safetensors
  framework: PyTorch
  version: 1.0.0
  license: Apache-2.0
```

Tip: `kit init <dir>` inspects a directory and generates a starting Kitfile.
Useful flags: `--name`, `--desc`, `--author`, `--force` (overwrite an existing
Kitfile), and `--output` (`-` writes to stdout). `kit init <repo> --remote`
generates a Kitfile from a remote Hugging Face repository (with `--ref`).

## pack — build a ModelKit

`kit pack [flags] DIRECTORY`

Builds a ModelKit from the Kitfile using `DIRECTORY` as the context, and stores
it in local storage. Relative paths in the Kitfile are resolved against
`DIRECTORY`.

- `-f, --file` — path to the Kitfile (default: `Kitfile` in the context dir; use
  `-` to read from stdin).
- `-t, --tag` — assign one or more tags, comma-separated:
  `-t registry/repository:tag1,tag2`.
- `--compression` — `none` (default), `gzip`, `gzip-fastest`, or `zstd`.
- `--layer-format` — `tar` (default) or `raw`.

```
kit pack .
kit pack . -f ./Kitfile -t myregistry.com/myorg/mymodel:1.0.0
```

## tag — add a reference

`kit tag SOURCE_MODELKIT[:TAG] TARGET_MODELKIT[:TAG]`

Adds a new reference to an existing local ModelKit. Both the source and target
must include a tag or digest.

A full reference is `[HOST[:PORT]/][NAMESPACE/]REPOSITORY[:TAG]`. `HOST` defaults
to `localhost` when omitted. A `TAG` may use letters, digits, `_`, `.`, `-`, must
not start with `.` or `-`, and is at most 128 characters.

```
kit tag myregistry.com/myorg/mymodel:latest myregistry.com/myorg/mymodel:v1.0.0
```

## login — authenticate

`kit login [flags] [REGISTRY]`

Authenticate before pushing or pulling from a private registry.

- `-u, --username`, `-p, --password`, `--password-stdin`.

```
kit login myregistry.com -u myuser --password-stdin
```

## push — upload to a registry

`kit push [flags] SOURCE [DESTINATION]`

Uploads a local ModelKit to a remote registry. Without a `DESTINATION`, the
ModelKit must already be tagged with a full registry reference (a registry host
is required — you cannot push to `localhost`). With a `DESTINATION`, a locally
tagged ModelKit is pushed to that reference.

```
kit push myregistry.com/myorg/mymodel:1.0.0
kit push mymodel:1.0.0 myregistry.com/myorg/mymodel:latest
```

## pull — download from a registry

`kit pull [flags] registry/repository[:tag|@digest]`

Downloads a ModelKit into local storage. A registry host is required (you cannot
pull from `localhost`).

```
kit pull myregistry.com/myorg/mymodel:1.0.0
```

## unpack — extract contents to disk

`kit unpack [flags] [registry/]repository[:tag|@digest]`

Extracts a ModelKit's contents to the filesystem. Looks in local storage first,
then the remote registry.

- `-d, --dir` — target directory (created if missing; default: current dir).
- `-f, --filter` — limit what is extracted, format `[types]:[filters]`, where
  `types` is a comma-separated subset of `kitfile,model,datasets,code,docs,prompts,mcpservers`.
  Repeatable; a layer is unpacked if it matches any filter. Example:
  `--filter=model` or `--filter=datasets:my-dataset`.
- `-o, --overwrite` — overwrite existing files.
- `-i, --ignore-existing` — skip files that already exist.
- `--as-skill[=agent,...]` — install `SKILL.md` prompt layers as agent skills
  instead of unpacking to their original paths. With no value, kit auto-detects
  installed agents; otherwise pass a comma-separated list (e.g.
  `--as-skill=claude-code,cursor`). Installs globally unless `-d` is given.

```
kit unpack myregistry.com/myorg/mymodel:1.0.0 -d ./unpacked
kit unpack myregistry.com/myorg/mymodel:1.0.0 --filter=model -o
```

## Other useful commands

- `kit list` — list ModelKits in local storage (or a remote repository).
- `kit inspect REF` — print a ModelKit's manifest and config.
- `kit info REF` — show a ModelKit's Kitfile.
- `kit remove REF` — delete a ModelKit from local storage.
- `kit version` — print the installed kit version.

## Common errors and how to resolve them

- **"registry is required when pushing" / "...when pulling"** — the reference has
  no registry host (it resolves to `localhost`). Use a full reference like
  `myregistry.com/myorg/mymodel:tag`, and `kit tag` the local ModelKit first if
  needed.
- **"No tag specified ... Using 'latest' as default"** — informational; kit
  defaulted the reference to `:latest`. Specify a tag explicitly to avoid it.
- **"reference cannot include multiple tags"** — push/pull/unpack accept a single
  reference. Split multiple tags into separate commands (or use `kit tag`).
- **"source/target ModelKit reference requires a tag or digest"** — `kit tag`
  needs a `:tag` (or `@digest`) on both arguments.
- **Kitfile validation errors** (surface on pack/init) include:
  - "absolute paths are not supported in a Kitfile" — make every `path` relative
    to the context directory.
  - "only local paths are permitted" — `code`/`datasets`/`docs`/`prompts` paths
    must be local files or directories, not URLs.
  - "... use the same path ..." — two layers point at the same path; each layer
    needs a distinct path.
  - "invalid path for mcpServer ...: path must point to a single .mcpb file" — an
    `mcpServers` entry must reference one `.mcpb` bundle.
  - "Unrecognized manifestVersion" — set `manifestVersion: 1.0.0`.
- **"Failed to pack model kit"** — a path in the Kitfile is missing or the Kitfile
  itself was not found; verify `-f`/the context directory and that every `path`
  exists.
