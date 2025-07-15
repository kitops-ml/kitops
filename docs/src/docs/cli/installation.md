---
title: Install KitOps CLI - macOS, Windows, Linux
description: Learn how to install KitOps, the open-source CLI for packaging and managing AI/ML models with ModelKits. Supports macOS, Windows, Linux, and source builds.
keywords: install kitops, kit CLI download, kitops mac brew, kitops windows zip, linux modelkit CLI, install ai model packaging tool, mlops cli tool
---

<script setup>
import vGaTrack from '@theme/directives/ga'
</script>

# Installing KitOps

`kit` is the command-line tool for building and managing ModelKits.

Pick your platform to get started:
-	[MacOS](#-macos-install)
- [Windows](#-windows-install)
- [Linux](#-linux-install)
- [Build from source](#build-sources)

Need help? [Join the community on Discord](https://discord.gg/Tapeh8agYy).

## 🍎 MacOS KitOps Install

Install with Homebrew (recommended)
```sh
brew tap kitops-ml/kitops
brew install kitops
```

➡️ [Verify your installation](#verify-the-installation)

Or use ZIP instead...

1. Get the ZIP for your Mac chip type.
   - MacOS:<a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-darwin-arm64.zip"
  v-ga-track="{
    category: 'link',
    label: 'MacOS (Apple Silicon)',
    location: 'docs/installation'
  }">
  Apple Silicon / ARM64
</a>
   - MacOS:<a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-darwin-x86_64.zip"
  v-ga-track="{
    category: 'link',
    label: 'MacOS (Intel)',
    location: 'docs/installation'
  }">
  Intel / x86_64
</a>
2. Move the kit executable to `/usr/local/bin`
3. Test it worked by running `kit version` in your terminal
4. [Verify your installation](#verify-the-installation)

## 🪟 Windows KitOps Install

1. Get the ZIP for your hardware type:
   - Windows: <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-windows-x86_64.zip"
  v-ga-track="{
    category: 'link',
    label: 'Windows (AMD64)',
    location: 'docs/installation'
  }">
  Intel / AMD, 64-bit
</a>
   - Windows: <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-windows-arm64.zip"
  v-ga-track="{
    category: 'link',
    label: 'Windows (ARM64)',
    location: 'docs/installation'
  }">
  ARM 64-bit
</a>
   - Windows: <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-windows-i386.zip"
  v-ga-track="{
    category: 'link',
    label: 'Windows (x86_32)',
    location: 'docs/installation'
  }">
  Intel / AMD, 32-bit
</a>
2. Unzip the file using “Extract All…”
3. Move `kit.exe` to a folder in your system PATH
4. [Verify your installation](#verify-the-installation)

## 🐧 Linux KitOps Install

Install with Homebrew (recommended)
```sh
brew tap kitops-ml/kitops
brew install kitops
kit version
```

Or use TAR instead...

1. Get the ZIP for your Mac chip type.
   - Linux:<a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-linux-x86_64.tar.gz"
  v-ga-track="{
    category: 'link',
    label: 'Linux (AMD64)',
    location: 'docs/installation'
  }">
  Intel / AMD, AMD 64-bit
</a>
   - Linux:<a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-linux-arm64.tar.gz"
  v-ga-track="{
    category: 'link',
    label: 'Linux (ARM64)',
    location: 'docs/installation'
  }">
  ARM 64-bit
</a>
   - Linux:<a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-linux-i386.tar.gz"
  v-ga-track="{
    category: 'link',
    label: 'Linux (x86_32)',
    location: 'docs/installation'
  }">
  Intel / AMD, 32-bit
</a>
2. In a terminal:
   ```
   tar -xzvf kitops-linux-x86_64.tar.gz
   sudo mv kit /usr/local/bin/
   ```
3. [Verify your installation](#verify-the-installation)

## Verify the KitOps Installation

After install, open a new terminal and run:
```shell
kit version
```

If Kit is installed correctly, you'll see a kit version printed.

## Next Step: Use KitOps

You’re ready to go. Head to our Quick Start guide to:
- Create your first Kitfile
- Pack a model
- Push to a registry

➡️ [Get Started tutorial](../get-started.md)

## Build KitOps from Source

If you'd rather build from source you'll need:
- Git
- Go

You can check if you have Go installed by running go version in your terminal. If you need to install Go, visit the official Go download page for instructions.

Steps:

### 1. Clone and build the project
```sh
git clone https://github.com/kitops-ml/kitops.git
cd kitops
go build -o kit
```
This command compiles the source code into an executable named `kit`. If you are on Windows, consider renaming the executable to `kit.exe`.

### 2. Move the binary:
- macOS/Linux: `sudo mv kit /usr/local/bin/`
- Windows: move kit.exe to a folder in your PATH

### 3. Verify the installation
Open a new terminal and run: `kit version`

You should see your kit version number printed in the terminal.

### 4. Optional environment settings
You can configure which directory credentials and storage are located:

* `--config` flag for a specific kit CLI execution
* `KITOPS_HOME` environment variable for permanent configurations
If the `KITOPS_HOME` is set in various places the order of precedence is:
   1. `--config` flag, if specified
   2. `$KITOPS_HOME` environment variable, if set
   3. A default OS-dependent value:
      - Linux: `$XDG_DATA_HOME/kitops`, falling back to `~/.local/share/kitops`
      - Windows: `%LOCALAPPDATA%\kitops`
      - MacOS: `~/Library/Caches/kitops`

---

**Have feedback or questions?**
Open an [issue on GitHub](https://github.com/kitops-ml/kitops/issues) or [join us on Discord](https://discord.gg/Tapeh8agYy).