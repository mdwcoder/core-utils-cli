[Español](README.es.md) | [English](README.en.md)

---

# Core Utils CLI (`cu`)

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go](https://img.shields.io/badge/go-1.22+-00ADD8.svg)
![Status](https://img.shields.io/badge/status-stable-green.svg)

**Core Utils CLI (`cu`)** is the package manager / launcher for the [Core Utils](https://core-utils.dev) ecosystem. It talks to the Core Utils registry to search, install, update and remove the CLI tools, scripts and apps published under [core-utils.dev](https://core-utils.dev).

---

## ⚠️ Installation is different from the other Core Utils tools

Most Core Utils projects are installed from source (`git clone` + `pipx install` / `init.sh`). **`core-utils-cli` is the exception.**

Since `cu` is the tool *used to install everything else*, it ships as a **precompiled binary** via **GitHub Releases**. There is no `pip install`, no `pipx`, and no Go toolchain required to use it — just download the build for your operating system and run it.

### 1. Download the right asset for your system

Go to the [latest release](https://github.com/mdwcoder/core-utils-cli/releases/latest) and download the file that matches your OS:

| Operating system | Asset | Notes |
|---|---|---|
| Linux (x86_64) | `cu-linux-amd64.AppImage` | Self-contained, run directly. May need `--appimage-extract-and-run` or FUSE. |
| Linux (x86_64) | `cu-linux-amd64.zip` | Plain zip containing the `cu` binary. Recommended for servers / scripting. |
| macOS (Intel + Apple Silicon) | `cu-macos-universal.dmg` | Universal binary (amd64 + arm64). Unsigned — see notes below. |
| Windows (x86_64) | `cu-windows-amd64.exe` | Unsigned — Windows SmartScreen may warn on first run. |

A `checksums.txt` file is published alongside every release with the SHA256 of each asset.

### 2. Install it

**Linux (recommended, from the zip):**

```bash
curl -fsSL -o cu.zip https://github.com/mdwcoder/core-utils-cli/releases/latest/download/cu-linux-amd64.zip
unzip cu.zip -d /tmp/cu-extract
install -Dm755 /tmp/cu-extract/cu ~/.local/bin/cu
```

Make sure `~/.local/bin` is in your `PATH`, then verify with:

```bash
cu version
cu doctor
```

You can also use the helper script, which does the steps above automatically (Linux/macOS):

```bash
curl -fsSL https://raw.githubusercontent.com/mdwcoder/core-utils-cli/main/scripts/install.sh | bash
```

**Linux (AppImage):**

```bash
curl -fsSL -o cu.AppImage https://github.com/mdwcoder/core-utils-cli/releases/latest/download/cu-linux-amd64.AppImage
chmod +x cu.AppImage
./cu.AppImage version
```

**macOS:**

1. Download `cu-macos-universal.dmg` from the [latest release](https://github.com/mdwcoder/core-utils-cli/releases/latest).
2. Open the `.dmg` and copy `cu` to `/usr/local/bin` (or any directory in your `PATH`).
3. Since the binary is unsigned, the first time you run it macOS Gatekeeper may block it. Right-click the binary → **Open**, or run:
   ```bash
   xattr -d com.apple.quarantine /usr/local/bin/cu
   ```

**Windows:**

1. Download `cu-windows-amd64.exe` from the [latest release](https://github.com/mdwcoder/core-utils-cli/releases/latest).
2. Move it to a folder of your choice and add that folder to your `PATH`, or run it directly as `cu.exe`.
3. The executable is unsigned — Windows SmartScreen may show "Windows protected your PC". Click **More info → Run anyway**.

### 3. Verify the checksum (recommended)

```bash
curl -fsSLO https://github.com/mdwcoder/core-utils-cli/releases/latest/download/checksums.txt
sha256sum -c checksums.txt --ignore-missing
```

---

## Usage

```text
Core Utils CLI (cu) — Usage:

  cu help                  Show this help
  cu version               Show CLI version
  cu registry sync         Download and cache the registry
  cu search <query>        Search tools in the registry
  cu info <tool>           Show details about a tool
  cu install <tool>        Install a tool
  cu list                  List installed tools
  cu update [tool]         Update all or one tool
  cu remove <tool>         Remove a tool
  cu doctor                Diagnose the environment

Environment:
  CORE_UTILS_REGISTRY_URL  Override registry endpoint
  NO_COLOR                 Disable colored output
```

### Examples

```bash
# Sync the registry
cu registry sync

# Search for tools
cu search note

# Show details about a tool
cu info memory-note-cli

# Install a tool for your platform
cu install memory-note-cli

# List installed tools
cu list

# Update everything (or just one tool)
cu update
cu update memory-note-cli

# Remove a tool
cu remove memory-note-cli

# Check your environment
cu doctor
```

`cu` stores its data under `~/.core-utils/` (`bin/`, `tools/<tool>/<version>/`, `config.json`, `installed.json`, `registry-cache.json`).

---

## How `cu install` works

1. Resolves the tool by id or name (case-insensitive) from the remote registry or local cache.
2. Detects your platform (`GOOS`/`GOARCH` → `linux-amd64`, `darwin-arm64`, etc.).
3. Downloads the matching asset to a temporary file.
4. Validates its SHA256 checksum when available.
5. Extracts the archive (zip / tar.gz) with path-traversal protection.
6. Installs the binary into `~/.core-utils/bin/`.
7. Records the installation in `~/.core-utils/installed.json`.

If `~/.core-utils/bin` is not in your `PATH`, `cu doctor` and `cu install` will warn you.

---

## Releases & versioning

Releases are built and published automatically by GitHub Actions whenever a tag matching `v*.*.*` is pushed. Each release includes:

- Linux AppImage and zip (amd64)
- macOS universal DMG (amd64 + arm64)
- Windows EXE (amd64)
- `checksums.txt` with SHA256 sums for all assets

`cu version` shows the build version, commit and date that were injected at build time:

```text
Core Utils CLI
Version:   v0.2.0
Commit:    8176ac2...
Build date: 2026-06-11T04:13:51Z
Platform:  linux/amd64
```

> **Note:** macOS and Windows builds are currently **unsigned** and not notarized. This is expected for an early-stage open-source project — always verify the checksum if you're unsure.

See [`RELEASE_PLAN.md`](RELEASE_PLAN.md) for the full release process.

---

## Building from source

If you'd rather build it yourself:

```bash
git clone https://github.com/mdwcoder/core-utils-cli.git
cd core-utils-cli
./scripts/build.sh
./bin/cu version
```

Requires Go 1.22+.

---

## License

MIT
