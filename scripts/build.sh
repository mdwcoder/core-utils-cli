#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BUILD_DIR="${PROJECT_ROOT}/bin"

mkdir -p "${BUILD_DIR}"

cd "${PROJECT_ROOT}"

echo "Building cu..."
GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"

go build -o "${BUILD_DIR}/cu" ./cmd/cu

echo "Built: ${BUILD_DIR}/cu (${GOOS}/${GOARCH})"
