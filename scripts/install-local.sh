#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BUILD_DIR="${PROJECT_ROOT}/bin"

"${SCRIPT_DIR}/build.sh"

BINARY="${BUILD_DIR}/cu"
TARGET="${HOME}/.local/bin/cu"
mkdir -p "$(dirname "${TARGET}")"

cp "${BINARY}" "${TARGET}"
chmod +x "${TARGET}"

echo "Installed locally to: ${TARGET}"
echo "Make sure ~/.local/bin is in your PATH."
