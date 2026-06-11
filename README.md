# Core Utils CLI (cu)

Core Utils CLI is the package manager for the Core Utils ecosystem. It allows you to discover, install, update, and remove developer tools published in the Core Utils registry.

## Features

- Search and discover tools in the Core Utils registry
- Install tools with platform-specific asset detection
- Automatic updates for installed tools
- Local environment diagnosis (`cu doctor`)
- Offline support via registry cache
- Minimal dependencies, built with Go standard library

## Installation (development)

```bash
cd core-utils-projects/core-utils-cli
./scripts/build.sh
./scripts/install-local.sh
```

This builds `cu` and copies it to `~/.local/bin/cu`.

Make sure `~/.local/bin` is in your `PATH`.

## Build

```bash
# Build the binary
go build -o bin/cu ./cmd/cu

# Run tests
go test ./...

# Format code
gofmt -w .
```

## Commands

| Command | Description |
|---------|-------------|
| `cu help` | Show help |
| `cu version` | Show CLI version |
| `cu registry sync` | Download and cache the registry |
| `cu search <query>` | Search tools by name, description, category, or language |
| `cu info <tool>` | Show details about a tool |
| `cu install <tool>` | Install a tool (downloads, validates, extracts) |
| `cu list` | List installed tools |
| `cu update [tool]` | Update all tools or a specific tool |
| `cu remove <tool>` | Remove an installed tool |
| `cu doctor` | Diagnose the local environment |

## Examples

```bash
# Sync the registry
cu registry sync

# Search for note-related tools
cu search notes

# Show info about a tool
cu info memory-note-cli

# Install a tool
cu install memory-note-cli

# List installed tools
cu list

# Update all tools
cu update

# Update a specific tool
cu update pushguard

# Remove a tool
cu remove pushguard

# Diagnose the environment
cu doctor
```

## Configuration

### Registry URL

The default registry is `https://core-utils.dev/api/registry`. You can override it:

**Via environment variable:**
```bash
CORE_UTILS_REGISTRY_URL=http://localhost:8000/api/registry cu search notes
```

**Via config file (`~/.core-utils/config.json`):**
```json
{
  "registry_url": "http://localhost:8000/api/registry"
}
```

## Local Structure

After running `cu` for the first time, the following directories and files are created under `~/.core-utils/`:

```
~/.core-utils/
├── config.json          # CLI configuration
├── registry-cache.json   # Cached registry from the backend
├── installed.json        # Metadata of installed tools
├── bin/                  # Symlinks/copies of installed binaries
└── tools/                # Extracted tool assets per version
    └── <tool-id>/
        └── <version>/
```

## Security

- Assets are downloaded only over HTTPS (HTTP allowed only for local development).
- SHA256 checksums are validated when available.
- If a checksum is `CHANGE_ME` or missing, a clear warning is shown.
- Path traversal protection during archive extraction (zip and tar.gz).
- No remote scripts are executed automatically.
- No telemetry or private data is sent to the backend.

## Testing against a local backend

```bash
# Start the backend
cd core-utils-backend
source .venv/bin/activate
uvicorn app.main:app --reload

# In another terminal, use the CLI with the local backend
CORE_UTILS_REGISTRY_URL=http://localhost:8000/api/registry ./bin/cu registry sync
CORE_UTILS_REGISTRY_URL=http://localhost:8000/api/registry ./bin/cu search notes
CORE_UTILS_REGISTRY_URL=http://localhost:8000/api/registry ./bin/cu info memory-note-cli
```

## Releases

Automated releases are published via GitHub Actions when a tag matching `v*.*.*` is pushed.

### Creating a release

```bash
git tag v1.0.0
git push origin v1.0.0
```

The workflow can also be triggered manually from the Actions tab in GitHub.

### Generated artifacts

| File | Platform | Format |
|------|----------|--------|
| `cu-linux-amd64.AppImage` | Linux (amd64) | AppImage |
| `cu-linux-amd64.zip` | Linux (amd64) | ZIP archive |
| `cu-macos-universal.dmg` | macOS (universal) | DMG |
| `cu-windows-amd64.exe` | Windows (amd64) | Portable executable |
| `checksums.txt` | All | SHA256 checksums |

### Verifying downloads

**Linux:**
```bash
sha256sum -c checksums.txt
```

**macOS:**
```bash
shasum -a 256 cu-macos-universal.dmg
```

**Windows (PowerShell):**
```powershell
Get-FileHash .\cu-windows-amd64.exe -Algorithm SHA256
```

### Known limitations

- **macOS DMG** is unsigned and not notarized. Gatekeeper will block execution unless you right-click → Open.
- **Windows EXE** is unsigned. SmartScreen may show a warning.
- **AppImage** is built for Linux amd64 without AppStream metadata. On systems without FUSE, use `--appimage-extract-and-run`.
- Future releases may include `.tar.gz` archives for direct distribution.

## Environment Variables

| Variable | Description |
|----------|-------------|
| `CORE_UTILS_REGISTRY_URL` | Override the registry endpoint URL |
| `CORE_UTILS_HOME` | Override the local data directory (default `~/.core-utils`) |
| `NO_COLOR` | Disable colored output |

## License

MIT
