---
outline: 2
title: "Kit CLI Reference"
description: Explore the Kit CLI command reference. Get detailed information on commands available for building, versioning, pushing, pulling, and running ModelKits within your AI/ML projects.
---
<script setup>
import VersionInfo from '@theme/components/VersionInfo.vue'
</script>

<VersionInfo />
## kit cache

Manage temporary files cached by Kit

### Synopsis

Manage files stored in the temporary KitOps cache dir ($KITOPS_HOME/cache)

Normally, this directory is empty, but may contain leftover files from resumable
downloads or files that were not cleaned up due to the command being cancelled.

The $KITOPS_HOME location is system dependent:
	- Linux: $XDG_DATA_HOME/kitops with a fall back to $HOME/.local/share/kitops
	- MacOS: ~/Library/Caches/kitops
	- Windows: %LOCALAPPDATA%\kitops


### Examples

```
# Get information about size of cached files
kit cache info

# Clear files in cache
kit cache clear
```

### Options

```
  -h, --help   help for cache
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit cache clear

Clear temporary cache storage

### Synopsis

Clear temporary files from cache storage.

```
kit cache clear [flags]
```

### Options

```
  -h, --help   help for clear
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit cache info

Get information about cache disk usage

### Synopsis

Print the total size of temporary files in the cache directory.

```
kit cache info [flags]
```

### Options

```
  -h, --help   help for info
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit dev

Run models locally (experimental)

### Synopsis

Start a local server and interact with a model in the browser

### Examples

```
kit dev start
```

### Options

```
  -h, --help   help for dev
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit dev logs

View logs for development server

### Synopsis

Print any logs output by the development server.

If the development server is currently running, the logs for this server will
be printed. If it is stopped, the logs for the previous run of the server, if
available, will be printed instead.

```
kit dev logs [flags]
```

### Options

```
  -f, --follow   Stream the log file
  -h, --help     help for logs
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit dev start

Start development server (experimental)

### Synopsis

Start development server (experimental) from a modelkit

Start a development server for a modelkit. You can provide either:
- A directory path containing an unpacked modelkit with a Kitfile
- A ModelKit reference in the format registry/repository[:tag|@digest] 
  (e.g., myrepo/my-model:latest) which will be automatically extracted 
  to a temporary directory

When using a ModelKit reference, only the model components are extracted
to optimize startup time.

```
kit dev start [directory|registry/repository[:tag|@digest]] [flags]
```

### Examples

```
# Serve the model located in the current directory
kit dev start

# Serve the modelkit in ./my-model on port 8080
kit dev start ./my-model --port 8080

# Serve a ModelKit reference from local storage or registry
kit dev start myrepo/my-model:latest

# Serve a specific model with custom host and port
kit dev start registry.example.com/models/llama2:7b --host 0.0.0.0 --port 8080
```

### Options

```
  -f, --file string       Path to the kitfile
      --host string       Host for the development server (default "127.0.0.1")
      --port int          Port for development server to listen on
      --plain-http        Use plain HTTP when connecting to remote registries
      --tls-verify        Require TLS and verify certificates when connecting to remote registries (default true)
      --cert string       Path to client certificate used for authentication (can also be set via environment variable KITOPS_CLIENT_CERT)
      --key string        Path to client certificate key used for authentication (can also be set via environment variable KITOPS_CLIENT_KEY)
      --concurrency int   Maximum number of simultaneous uploads/downloads (default 5)
      --proxy string      Proxy to use for connections (overrides proxy set by environment)
  -h, --help              help for start
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit dev stop

Stop development server

### Synopsis

Stop the development server if it is running

```
kit dev stop [flags]
```

### Options

```
  -h, --help   help for stop
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit diff

Compare two ModelKits

### Synopsis

Compare two ModelKits to see the differences in their layers.
		
ModelKits can be specified from either a local or from a remote registry.
To specify a local ModelKit, prefix the reference with 'local://', e.g. 'local://jozu.ml/foo/bar'.
To specify a remote ModelKit, prefix the reference with 'remote://', e.g. 'remote://jozu.ml/foo/bar'.
If no prefix is specified, the local registry will be checked first.


```
kit diff <ModelKit1> <ModelKit2> [flags]
```

### Examples

```
# Compare two ModelKits
kit diff jozu.ml/foo:latest jozu.ml/bar:latest

# Compare two ModelKits from a remote registry
kit diff remote://jozu.ml/foo:champion remote://jozu.ml/bar:latest

# Compare local ModelKit with a remote ModelKit
kit diff local://jozu.ml/foo:latest remote://jozu.ml/foo:latest

```

### Options

```
      --plain-http        Use plain HTTP when connecting to remote registries
      --tls-verify        Require TLS and verify certificates when connecting to remote registries (default true)
      --cert string       Path to client certificate used for authentication (can also be set via environment variable KITOPS_CLIENT_CERT)
      --key string        Path to client certificate key used for authentication (can also be set via environment variable KITOPS_CLIENT_KEY)
      --concurrency int   Maximum number of simultaneous uploads/downloads (default 5)
      --proxy string      Proxy to use for connections (overrides proxy set by environment)
  -h, --help              help for diff
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit import

Import a model from HuggingFace

### Synopsis

Download a repository from HuggingFace and package it as a ModelKit.

The repository can be specified either via a repository (e.g. myorg/myrepo) or
with a full URL (https://huggingface.co/myorg/myrepo). The repository will be
downloaded to a temporary directory and be packaged using a generated Kitfile.

In interactive settings, this command will read the EDITOR environment variable
to determine which editor should be used for editing the Kitfile.

This command supports multiple ways of downloading files from the remote
repository. The tool used can be specified using the --tool flag with one of the
options below:

  --tool=hf  : Download files using the Huggingface API. Requires REPOSITORY to
	             be a Huggingface repository. This is the default for Huggingface
							 repositories
  --tool=git : Download files using Git and Git LFS. Works for any Git
	             repository but requires that Git and Git LFS are installed.

By default, Kit will automatically select the tool based on the provided
REPOSITORY.

```
kit import [flags] REPOSITORY
```

### Examples

```
# Download repository myorg/myrepo and package it, using the default tag (myorg/myrepo:latest)
kit import myorg/myrepo

# Download repository using a different version (tag) than the default and package it
kit import myorg/myrepo --ref v1.0.0

# Download repository and tag it 'myrepository:mytag'
kit import myorg/myrepo --tag myrepository:mytag

# Download repository and pack it using an existing Kitfile
kit import myorg/myrepo --file ./path/to/Kitfile
```

### Options

```
      --ref string        Version (tag) of repository to import (default "main")
      --token string      Token to use for authenticating with repository
  -t, --tag string        Tag for the ModelKit (default is '[repository]:latest')
  -f, --file string       Path to Kitfile to use for packing (use '-' to read from standard input)
      --tool string       Tool to use for downloading files: options are 'git' and 'hf' (default: detect based on repository)
      --concurrency int   Maximum number of simultaneous downloads (for huggingface) (default 5)
  -h, --help              help for import
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit info

Show the configuration for a modelkit

### Synopsis

Print the contents of a modelkit config to the screen.

By default, kit will check local storage for the specified modelkit. To see
the configuration for a modelkit stored on a remote registry, use the
--remote flag.

```
kit info [flags] MODELKIT
```

### Examples

```
# See configuration for a local modelkit:
kit info mymodel:mytag

# See configuration for a local modelkit by digest:
kit info mymodel@sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a

# See configuration for a remote modelkit:
kit info --remote registry.example.com/my-model:1.0.0
```

### Options

```
      --plain-http        Use plain HTTP when connecting to remote registries
      --tls-verify        Require TLS and verify certificates when connecting to remote registries (default true)
      --cert string       Path to client certificate used for authentication (can also be set via environment variable KITOPS_CLIENT_CERT)
      --key string        Path to client certificate key used for authentication (can also be set via environment variable KITOPS_CLIENT_KEY)
      --concurrency int   Maximum number of simultaneous uploads/downloads (default 5)
      --proxy string      Proxy to use for connections (overrides proxy set by environment)
  -r, --remote            Check remote registry instead of local storage
  -f, --filter string     filter with node selectors
  -h, --help              help for info
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit init

Generate a Kitfile for the contents of a directory

### Synopsis

Examine the contents of a directory and attempt to generate a basic Kitfile
based on common file formats. Any files whose type (i.e. model, dataset, etc.)
cannot be determined will be included in a code layer.

By default the command will prompt for input for a name and description for the Kitfile

```
kit init [flags] PATH
```

### Examples

```
# Generate a Kitfile for the current directory:
kit init .

# Generate a Kitfile for files in ./my-model, with name "mymodel" and a description:
kit init ./my-model --name "mymodel" --desc "This is my model's description"

# Generate a Kitfile, overwriting any existing Kitfile:
kit init ./my-model --force
```

### Options

```
      --name string     Name for the ModelKit
      --desc string     Description for the ModelKit
      --author string   Author for the ModelKit
  -f, --force           Overwrite existing Kitfile if present
  -h, --help            help for init
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit inspect

Inspect a modelkit's manifest

### Synopsis

Print the contents of a modelkit manifest to the screen.

By default, kit will check local storage for the specified modelkit. To
inspect a modelkit stored on a remote registry, use the --remote flag.

```
kit inspect [flags] MODELKIT
```

### Examples

```
# Inspect a local modelkit:
kit inspect mymodel:mytag

# Inspect a local modelkit by digest:
kit inspect mymodel@sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a

# Inspect a remote modelkit:
kit inspect --remote registry.example.com/my-model:1.0.0
```

### Options

```
      --plain-http        Use plain HTTP when connecting to remote registries
      --tls-verify        Require TLS and verify certificates when connecting to remote registries (default true)
      --cert string       Path to client certificate used for authentication (can also be set via environment variable KITOPS_CLIENT_CERT)
      --key string        Path to client certificate key used for authentication (can also be set via environment variable KITOPS_CLIENT_KEY)
      --concurrency int   Maximum number of simultaneous uploads/downloads (default 5)
      --proxy string      Proxy to use for connections (overrides proxy set by environment)
  -r, --remote            Check remote registry instead of local storage
  -h, --help              help for inspect
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit list

List modelkits in a repository

### Synopsis

Displays a list of modelkits available in a repository.

This command provides an overview of modelkits stored either in the local
repository or a specified remote repository. It displays each modelkit along
with its associated tags and the cumulative size of its contents. Modelkits
comprise multiple artifacts, including models, datasets, code, and
configuration, designed to enhance reusability and modularity. However, this
command focuses on the aggregate rather than listing individual artifacts.

Each modelkit entry includes its DIGEST, a unique identifier that ensures
distinct versions of a modelkit are easily recognizable, even if they share
the same name or tags. Modelkits with multiple tags or repository names will
appear multiple times in the list, distinguished by their DIGEST.

The SIZE displayed for each modelkit represents the total storage space
occupied by all its components.

Use the --format flag to change how results are printed. Valid values are
"table", "json", or a Go template. When a value other than "table" or "json"
is supplied, the flag contents are treated as a Go template executed once per
listed modelkit.

Template placeholders:

* <code v-pre>{{ .Repo }}</code> - repository name
* <code v-pre>{{ .Tags }}</code> - slice of tags for the modelkit
* <code v-pre>{{ .Digest }}</code> - digest of the modelkit
* <code v-pre>{{ .ModelName }}</code> - name defined in the Kitfile
* <code v-pre>{{ .Size }}</code> - total size of the modelkit
* <code v-pre>{{ .Author }}</code> - author from the Kitfile

```
kit list [flags] [REPOSITORY]
```

### Examples

```
# List local modelkits
kit list

# List modelkits from a remote repository
kit list registry.example.com/my-namespace/my-model
```

### Options

```
      --format string     Output format: table, json, or Go template string (default "table")
      --plain-http        Use plain HTTP when connecting to remote registries
      --tls-verify        Require TLS and verify certificates when connecting to remote registries (default true)
      --cert string       Path to client certificate used for authentication (can also be set via environment variable KITOPS_CLIENT_CERT)
      --key string        Path to client certificate key used for authentication (can also be set via environment variable KITOPS_CLIENT_KEY)
      --concurrency int   Maximum number of simultaneous uploads/downloads (default 5)
      --proxy string      Proxy to use for connections (overrides proxy set by environment)
  -h, --help              help for list
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit login

Log in to an OCI registry

### Synopsis

Log in to a specified OCI-compatible registry. Credentials are saved and used
automatically for future CLI operations

```
kit login [flags] [REGISTRY]
```

### Examples

```
# Login to ghcr.io
kit login ghcr.io -u github_user -p personal_token

# Login to docker.io with password from stdin
kit login docker.io --password-stdin -u docker_user
```

### Options

```
  -u, --username string   registry username
  -p, --password string   registry password or token
      --password-stdin    read password from stdin
      --plain-http        Use plain HTTP when connecting to remote registries
      --tls-verify        Require TLS and verify certificates when connecting to remote registries (default true)
      --cert string       Path to client certificate used for authentication (can also be set via environment variable KITOPS_CLIENT_CERT)
      --key string        Path to client certificate key used for authentication (can also be set via environment variable KITOPS_CLIENT_KEY)
      --concurrency int   Maximum number of simultaneous uploads/downloads (default 5)
      --proxy string      Proxy to use for connections (overrides proxy set by environment)
  -h, --help              help for login
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit logout

Log out from an OCI registry

### Synopsis

Log out from a specified OCI-compatible registry. Any saved credentials are
removed from storage.

```
kit logout [flags] REGISTRY
```

### Examples

```
# Log out from ghcr.io
kit logout ghcr.io
```

### Options

```
  -h, --help   help for logout
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit pack

Pack a modelkit

### Synopsis

Pack a modelkit from a kitfile using the given context directory.

The packing process involves taking the configuration and resources defined in
your kitfile and using them to create a modelkit. This modelkit is then stored
in your local registry, making it readily available for further actions such
as pushing to a remote registry for collaboration.

Unless a different location is specified, this command looks for the kitfile
at the root of the provided context directory. Any relative paths defined
within the kitfile are interpreted as being relative to this context
directory.

```
kit pack [flags] DIRECTORY
```

### Examples

```
# Pack a modelkit using the kitfile in the current directory
kit pack .

# Pack a modelkit with a specific kitfile and tag
kit pack . -f /path/to/your/Kitfile -t registry/repository:modelv1
```

### Options

```
  -f, --file string          Specifies the path to the Kitfile explicitly (use "-" to read from standard input)
  -t, --tag string           Assigns one or more tags to the built modelkit. Example: -t registry/repository:tag1,tag2
      --compression string   Compression format to use for layers. Valid options: 'none' (default), 'gzip', 'gzip-fastest' (default "none")
  -h, --help                 help for pack
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit pull

Retrieve modelkits from a remote registry to your local environment.

### Synopsis

Downloads modelkits from a specified registry. The downloaded modelkits
are stored in the local registry.

```
kit pull [flags] registry/repository[:tag|@digest]
```

### Examples

```
# Pull the latest version of a modelkit from a remote registry
kit pull registry.example.com/my-model:latest
```

### Options

```
      --plain-http        Use plain HTTP when connecting to remote registries
      --tls-verify        Require TLS and verify certificates when connecting to remote registries (default true)
      --cert string       Path to client certificate used for authentication (can also be set via environment variable KITOPS_CLIENT_CERT)
      --key string        Path to client certificate key used for authentication (can also be set via environment variable KITOPS_CLIENT_KEY)
      --concurrency int   Maximum number of simultaneous uploads/downloads (default 5)
      --proxy string      Proxy to use for connections (overrides proxy set by environment)
  -h, --help              help for pull
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit push

Upload a modelkit to a specified registry

### Synopsis

This command pushes modelkits from local storage to a remote registry.

If specified without a destination, the ModelKit must be tagged locally before
pushing.

```
kit push [flags] SOURCE [DESTINATION]
```

### Examples

```
# Push the ModelKit tagged 'latest' to a remote registry
kit push registry.example.com/my-org/my-model:latest

# Push a ModelKit to a remote registry by digest
kit push registry.example.com/my-org/my-model@sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a

# Push local modelkit 'mymodel:1.0.0' to a remote registry
kit push mymodel:1.0.0 registry.example.com/my-org/my-model:latest
```

### Options

```
      --plain-http        Use plain HTTP when connecting to remote registries
      --tls-verify        Require TLS and verify certificates when connecting to remote registries (default true)
      --cert string       Path to client certificate used for authentication (can also be set via environment variable KITOPS_CLIENT_CERT)
      --key string        Path to client certificate key used for authentication (can also be set via environment variable KITOPS_CLIENT_KEY)
      --concurrency int   Maximum number of simultaneous uploads/downloads (default 5)
      --proxy string      Proxy to use for connections (overrides proxy set by environment)
  -h, --help              help for push
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit remove

Remove a modelkit from local storage

### Synopsis

Removes a modelkit from storage on the local disk.

The model to be removed may be specifed either by a tag or by a digest. If
specified by digest, that modelkit will be removed along with any tags that
might refer to it. If specified by tag (and the --force flag is not used),
the modelkit will only be removed if no other tags refer to it; otherwise
it is only untagged.

```
kit remove [flags] registry/repository[:tag|@digest]
```

### Examples

```
# Remove modelkit by tag
kit remove my-registry.com/my-org/my-repo:my-tag

# Remove modelkit by digest
kit remove my-registry.com/my-org/my-repo@sha256:<digest>

# Remove multiple tags for a modelkit
kit remove my-registry.com/my-org/my-repo:tag1,tag2,tag3

# Remove all untagged modelkits
kit remove --all

# Remove all locally stored modelkits
kit remove --all --force
```

### Options

```
  -f, --force             remove modelkit and all other tags that refer to it
  -a, --all               remove all untagged modelkits
  -r, --remote            remove modelkit from remote registry
      --plain-http        Use plain HTTP when connecting to remote registries
      --tls-verify        Require TLS and verify certificates when connecting to remote registries (default true)
      --cert string       Path to client certificate used for authentication (can also be set via environment variable KITOPS_CLIENT_CERT)
      --key string        Path to client certificate key used for authentication (can also be set via environment variable KITOPS_CLIENT_KEY)
      --concurrency int   Maximum number of simultaneous uploads/downloads (default 5)
      --proxy string      Proxy to use for connections (overrides proxy set by environment)
  -h, --help              help for remove
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit tag

Create a tag that refers to a modelkit

### Synopsis

Create or update a tag {target-modelkit} that refers to {source-modelkit}

This command assigns a new tag to an existing modelkit (source-modelkit) or
updates an existing tag, effectively renaming or categorizing modelkits for
better organization and version control. Tags are identifiers linked to
specific modelkit versions within a repository.

A full modelkit reference has the following format:

[HOST[:PORT_NUMBER]/][NAMESPACE/]REPOSITORY[:TAG]

  * HOST: Optional. The registry hostname where the ModelKit is located.
    Defaults to localhost if unspecified. Must follow standard DNS rules
    (excluding underscores).

  * PORT_NUMBER: Optional. Specifies the registry's port number if a hostname
    is provided.

  * NAMESPACE: Represents a user or organization's namespace, consisting of
    slash-separated components that may include lowercase letters, digits, and
    specific separators (periods, underscores, hyphens).

  * REPOSITORY: The name of the repository, typically corresponding to the
    modelkit's name.

  * TAG: A human-readable identifier for the modelkit version or variant.
    Valid ASCII characters include lowercase and uppercase letters, digits,
    underscores, periods, and hyphens. It cannot start with a period or hyphen
    and is limited to 128 characters.

Tagging is a powerful way to manage different versions or configurations of
your modelkits, making it easier to organize, retrieve, and deploy specific
iterations. Ensure tags are meaningful and consistent across your team or
organization to maintain clarity and avoid confusion.

```
kit tag SOURCE_MODELKIT[:TAG] TARGET_MODELKIT[:TAG] [flags]
```

### Examples

```
kit tag myregistry.com/myrepo/mykit:latest myregistry.com/myrepo/mykit:v1.0.0
```

### Options

```
  -h, --help   help for tag
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit unpack

Produce the components from a modelkit on the local filesystem

### Synopsis

Produces all or selected components of a modelkit on the local filesystem.

This command unpacks a modelkit's components, including models, code,
datasets, and configuration files, to a specified directory on the local
filesystem. By default, it attempts to find the modelkit in local storage; if
not found, it searches the remote registry and retrieves it. This process
ensures that the necessary components are always available for unpacking,
optimizing for efficiency by fetching only specified components from the
remote registry when necessary.

The content that is unpacked can be limited via the --filter (-f) flag. For example,
use
    --filter=model
to unpack only the model, or
    --filter=datasets:my-dataset
to unpack only the dataset named 'my-dataset'.

Valid filters have the format
    [types]:[filters]
where [types] is a comma-separated list of Kitfile fields (kitfile, model, datasets
code, or docs) and [filters] is an optional comma-separated list of additional filters
to apply, which are matched against the Kitfile to further restrict what is extracted.
Additional filters match elements of the Kitfile on either the name (if present) or
the path used.

The filter field can be specified multiple times. A layer will be unpacked if it matches
any of the specified filters

```
kit unpack [flags] [registry/]repository[:tag|@digest]
```

### Examples

```
# Unpack all components of a modelkit to the current directory
kit unpack myrepo/my-model:latest -d /path/to/unpacked

# Unpack only the model and datasets of a modelkit to a specified directory
kit unpack myrepo/my-model:latest --filter=model,datasets -d /path/to/unpacked

# Unpack only the dataset named "my-dataset" to the current directory
kit unpack myrepo/my-model:latest --filter=datasets:my-dataset

# Unpack only the docs layer with path "./README.md" to the current directory
kit unpack myrepo/my-model:latest --filter=docs:./README.md

# Unpack the model and the dataset named "validation"
kit unpack myrepo/my-model:latest --filter=model --filter=datasets:validation

# Unpack a modelkit from a remote registry with overwrite enabled
kit unpack registry.example.com/myrepo/my-model:latest -o -d /path/to/unpacked
```

### Options

```
  -d, --dir string           The target directory to unpack components into. This directory will be created if it does not exist
  -o, --overwrite            Overwrites existing files and directories in the target unpack directory without prompting
  -i, --ignore-existing      Skip unpacking files if a file with that name already exists
  -f, --filter stringArray   Filter what is unpacked from the modelkit based on type and name. Can be specified multiple times
      --kitfile              Unpack only Kitfile (deprecated: use --filter=kitfile)
      --model                Unpack only model (deprecated: use --filter=model)
      --code                 Unpack only code (deprecated: use --filter=code)
      --datasets             Unpack only datasets (deprecated: use --filter=datasets)
      --docs                 Unpack only docs (deprecated: use --filter=docs)
      --plain-http           Use plain HTTP when connecting to remote registries
      --tls-verify           Require TLS and verify certificates when connecting to remote registries (default true)
      --cert string          Path to client certificate used for authentication (can also be set via environment variable KITOPS_CLIENT_CERT)
      --key string           Path to client certificate key used for authentication (can also be set via environment variable KITOPS_CLIENT_KEY)
      --concurrency int      Maximum number of simultaneous uploads/downloads (default 5)
      --proxy string         Proxy to use for connections (overrides proxy set by environment)
  -h, --help                 help for unpack
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

## kit version

Display the version information for the CLI

### Synopsis

The version command prints detailed version information.

This information includes the current version of the tool, the Git commit that
the version was built from, the build time, and the version of Go it was
compiled with.

```
kit version [flags]
```

### Options

```
  -h, --help                        help for version
      --show-update-notifications   Enable or disable update notifications for the Kit CLI
```

### Options inherited from parent commands

```
      --config string      Alternate path to root storage directory for CLI
      --log-level string   Log messages above specified level ('trace', 'debug', 'info', 'warn', 'error') (default 'info') (default "info")
      --progress string    Configure progress bars for longer operations (options: none, plain, fancy) (default "plain")
  -v, --verbose count      Increase verbosity of output (use -vv for more)
```

