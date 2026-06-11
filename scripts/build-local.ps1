# Build script for Windows (PowerShell)
$ErrorActionPreference = "Stop"

$ProjectRoot = Split-Path -Parent $PSScriptRoot
$BuildDir = Join-Path $ProjectRoot "bin"
$null = New-Item -ItemType Directory -Force -Path $BuildDir

$Version = if ($env:VERSION) { $env:VERSION } else { "dev" }
$Commit = if ($env:COMMIT) { $env:COMMIT } else { "unknown" }
$Date = if ($env:DATE) { $env:DATE } else { (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ") }

$LdFlags = "-X 'github.com/mdwcoder/core-utils-cli/internal/config.Version=$Version'"
$LdFlags = "$LdFlags -X 'github.com/mdwcoder/core-utils-cli/internal/config.Commit=$Commit'"
$LdFlags = "$LdFlags -X 'github.com/mdwcoder/core-utils-cli/internal/config.Date=$Date'"

Write-Host "Building cu ($Version)..."
& go build -ldflags "$LdFlags" -o (Join-Path $BuildDir "cu.exe") ./cmd/cu

Write-Host "Built: $(Join-Path $BuildDir 'cu.exe')"
