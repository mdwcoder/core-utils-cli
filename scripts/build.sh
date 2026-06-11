#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BUILD_DIR="${PROJECT_ROOT}/bin"

mkdir -p "${BUILD_DIR}"

cd "${PROJECT_ROOT}"

VERSION="${VERSION:-dev}"
COMMIT="${COMMIT:-$(git rev-parse HEAD 2>/dev/null || echo unknown)}"
DATE="${DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"

GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"

LDFLAGS="-X 'github.com/mdwcoder/core-utils-cli/internal/config.Version=${VERSION}'"
LDFLAGS="${LDFLAGS} -X 'github.com/mdwcoder/core-utils-cli/internal/config.Commit=${COMMIT}'"
LDFLAGS="${LDFLAGS} -X 'github.com/mdwcoder/core-utils-cli/internal/config.Date=${DATE}'"

echo "Building cu (${VERSION})..."
go build -ldflags="${LDFLAGS}" -o "${BUILD_DIR}/cu" ./cmd/cu

echo "Built: ${BUILD_DIR}/cu (${GOOS}/${GOARCH})"
