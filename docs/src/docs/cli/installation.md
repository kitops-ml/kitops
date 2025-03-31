---
description: Learn how to install the Kit CLI on MacOS, Windows, and Linux. Follow our detailed instructions to set up KitOps and start managing your AI/ML ModelKits efficiently.
---

<script setup>
import vGaTrack from '@theme/directives/ga'
</script>

# Installing Kit

Kit is a command line tool for building and managing secure and shareable ModelKits. It works on Mac, Windows, and Linux computers.

This page includes instructions for:

* Installing on [MacOS](#🍎-macos-install) with Brew or ZIP
* Installing on [Windows](#🪟-windows-install) with ZIP
* Installing on [Linux](#🐧-linux-install) with Brew or TAR
* Building from the [source code](#🛠️-build-from-source-code)

[ discord banner ]

## 🍎 MacOS Install

The simplest way to install Kit on a Mac is with [Homebrew](https://brew.sh/). You can also install from [ZIP](#mac-install-from-zip).

### Mac Brew Install

1. Open a Terminal window
1. At the prompt type: `brew tap kitops-ml/kitops` (if that doesn't work use the [ZIP instructions](#mac-install-from-zip)
1. When the previous command completes, type `brew install kitops`

You can verify that `kit` is correctly installed by opening a new terminal or command prompt and typing:

```shell
kit version
```

This command should display the version number of the Kit CLI you have installed, indicating that the installation was successful.

Now follow our [Quick Start](../../get-started/) to learn how to pack and share your first ModelKit.

**Need Help?** If something isn't working [get help on our Discord channel](https://discord.gg/Tapeh8agYy).

### Mac Install from ZIP

There are two generations of Mac hardware, if you aren't sure which you have [check here](https://www.sweetwater.com/sweetcare/articles/intel-based-mac-or-mac-with-apple-silicon/#:~:text=Choose%20About%20This%20Mac.,a%20Mac%20with%20Apple%20silicon.&text=As%20of%20this%20writing%2C%20Apple,have%20an%20Intel%2Dbased%20Mac.).

1. MacOS: <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-darwin-arm64.zip"
  v-ga-track="{
    category: 'link',
    label: 'MacOS (Apple Silicon)',
    location: 'docs/installation'
  }">
  Apple Silicon / ARM64
</a>

2. MacOS: <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-darwin-x86_64.zip"
  v-ga-track="{
    category: 'link',
    label: 'MacOS (Intel)',
    location: 'docs/installation'
  }">
  Intel / x86_64
</a>

The Kit download will happen _so quickly_ on a fast connection that you might miss it...

* Open the Mac Finder and check your downloads location for a file that starts with `kitops-darwin`
* Double-click the `.zip` file to unpack it
* Select the executable file named `kit` from unpacked folder and move it to `/usr/local/bin`

You can verify that `kit` is correctly installed by opening a new terminal or command prompt and typing:

```shell
kit version
```

This command should display the version number of the Kit CLI you have installed, indicating that the installation was successful.

Now follow our [Quick Start](../../get-started/) to learn how to pack and share your first ModelKit.

**Need Help?** If something isn't working [get help on our Discord channel](https://discord.gg/Tapeh8agYy).

## 🪟 Windows Install

Make sure you get the correct download for your hardware.

1. Windows: <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-windows-x86_64.zip"
  v-ga-track="{
    category: 'link',
    label: 'Windows (AMD64)',
    location: 'docs/installation'
  }">
  Intel / AMD, 64-bit
</a>

1. Windows: <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-windows-arm64.zip"
  v-ga-track="{
    category: 'link',
    label: 'Windows (ARM64)',
    location: 'docs/installation'
  }">
  ARM 64-bit
</a>

1. Windows: <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-windows-i386.zip"
  v-ga-track="{
    category: 'link',
    label: 'Windows (x86_32)',
    location: 'docs/installation'
  }">
  Intel / AMD, 32-bit
</a>

The Kit download will happen _so quickly_ on a fast connection that you might miss it...

* Open the File Explorer and check your downloads location for a file that starts with `kitops-windows`
* Right-click the `.zip` file and select "Extract All..." to unzip the archive
* Move the extracted `kit.exe` to a directory that is <a href="https://www.computerhope.com/issues/ch000549.htm" target="_blank">included in your system's PATH variable</a> (this will allow you to run the Kit CLI from anywhere).

You can verify that `kit` is correctly installed by opening a new terminal or command prompt and typing:

```shell
kit version
```

This command should display the version number of the Kit CLI you have installed, indicating that the installation was successful.

Now follow our [Quick Start](../../get-started/) to learn how to pack and share your first ModelKit.

**Need Help?** If something isn't working [get help on our Discord channel](https://discord.gg/Tapeh8agYy).

## 🐧 Linux Install

The simplest way to install Kit on Linux is with [Homebrew](https://brew.sh/). You can also install from [TAR](#linux-tar-install).

### Linux Brew Install

1. Open a Terminal window
1. At the prompt type: `brew tap kitops-ml/kitops` (if that doesn't work use the [TAR instructions](#linux-tar-install)
1. When the previous command completes, type `brew install kitops`

You can verify that `kit` is correctly installed by opening a new terminal or command prompt and typing:

```shell
kit version
```

This command should display the version number of the Kit CLI you have installed, indicating that the installation was successful.

Now follow our [Quick Start](../../get-started/) to learn how to pack and share your first ModelKit.

**Need Help?** If something isn't working [get help on our Discord channel](https://discord.gg/Tapeh8agYy).

### Linux TAR Install

Make sure you get the correct download for your hardware.

1. Linux: <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-linux-x86_64.tar.gz"
  v-ga-track="{
    category: 'link',
    label: 'Linux (AMD64)',
    location: 'docs/installation'
  }">
  Intel / AMD, AMD 64-bit
</a>

1. Linux: <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-linux-arm64.tar.gz"
  v-ga-track="{
    category: 'link',
    label: 'Linux (ARM64)',
    location: 'docs/installation'
  }">
  ARM 64-bit
</a>

1. Linux: <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-linux-i386.tar.gz"
  v-ga-track="{
    category: 'link',
    label: 'Linux (x86_32)',
    location: 'docs/installation'
  }">
  Intel / AMD, 32-bit
</a>

The Kit download will happen _so quickly_ on a fast connection that you might miss it...

Open a terminal window in your downloads location and look for a file that starts with `kitops-linux`.

Use the `tar` command to extract the downloaded file. For example, if you downloaded the `kitops-linux-x86_64.tar.gz` file, you would use the following command:

```shell
tar -xzvf kitops-linux-x86_64.tar.gz
```

Move the extracted `kit` executable to a location in your system's PATH. A common choice is `/usr/local/bin`. You can do this with the `mv` command (this may require administrator privileges):

```
sudo mv kit /usr/local/bin/
```

After installation, you can verify that `kit` is correctly installed by opening a new terminal or command prompt and typing:

```shell
kit version
```

This command should display the version number of the Kit CLI you have installed, indicating that the installation was successful.

Now follow our [Quick Start](../../get-started/) to learn how to pack and share your first ModelKit.

**Need Help?** If something isn't working [get help on our Discord channel](https://discord.gg/Tapeh8agYy).

## 🛠️ Build from Source Code

For those who prefer or require building `kit` from the source code, this section will guide you through the necessary steps.

Before you begin, make sure you have the following installed on your system:

- Git
- A recent version of Go

You can check if you have Go installed by running `go version` in your terminal. If you need to install Go, visit the [official Go download page](https://golang.org/dl/) for instructions.

### Clone the Repository

First, clone the `kitops` GitHub repository to your local machine. Open a terminal and run:


```shell
git clone https://github.com/kitops-ml/kitops.git
cd kitops
```

This command clones the repository and changes your current directory to the cloned repository's root.

### Build Sources

Once inside the `kitops` directory, you can build the `kit` tool using the Go compiler. Run:

```shell
go build -o kit
```

This command compiles the source code into an executable named `kit`. If you are on Windows, you might want to name the executable `kit.exe`.

### Install the Executable

After the build process completes, you need to move the `kit` executable to a location in your system's PATH to make it accessible from anywhere in the terminal:

#### For MacOS and Linux:

```shell
sudo mv kit /usr/local/bin/
```

#### For Windows:

Move `kit.exe` to a directory that's included in your system's PATH variable. This step may vary based on your specific Windows setup.

### Verify the Installation

To verify that `kit` was installed successfully, open a new terminal window and type:

```shell
kit version
```

## Optional: Set Your Environment

You can configure which directory credentials and storage are located:
* `--config` flag for a specific kit CLI execution
* `KITOPS_HOME` environment variable for permanent configurations

If the `KITOPS_HOME` is set in various places the order of precedence is:
1. `--config` flag, if specified
1. `$KITOPS_HOME` environment variable, if set
1. A default OS-dependent value:

    Linux: `$XDG_DATA_HOME/kitops`, falling back to `~/.local/share/kitops`

    Windows: `%LOCALAPPDATA%\kitops`

    Darwin: `~/Library/Caches/kitops`


## Follow the Quick Start

Now that everything is set up you can follow our [Quick Start](../../get-started/) to learn how to pack and share your first ModelKit.

## Become a Design Partner

Interested in helping to shape the future of our project? Email <a href="mailto:feedback@jozu.com" target="blank">feedback@jozu.com</a> to learn more about our Design Partner program.
