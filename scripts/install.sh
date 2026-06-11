#!/usr/bin/env bash
# Downloads the latest core-utils-cli (cu) release for Linux/macOS and
# installs the `cu` binary into ~/.local/bin.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/mdwcoder/core-utils-cli/main/scripts/install.sh | bash
set -euo pipefail

REPO="mdwcoder/core-utils-cli"
INSTALL_DIR="${HOME}/.local/bin"
BIN_NAME="cu"

os="$(uname -s)"
case "${os}" in
  Linux)
    asset="cu-linux-amd64.zip"
    ;;
  Darwin)
    asset="cu-macos-universal.dmg"
    ;;
  *)
    echo "Unsupported OS: ${os}" >&2
    echo "Download a release manually from https://github.com/${REPO}/releases/latest" >&2
    exit 1
    ;;
esac

tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT

url="https://github.com/${REPO}/releases/latest/download/${asset}"
echo "Downloading ${asset}..."
curl -fsSL -o "${tmp_dir}/${asset}" "${url}"

mkdir -p "${INSTALL_DIR}"

case "${os}" in
  Linux)
    unzip -q "${tmp_dir}/${asset}" -d "${tmp_dir}/extract"
    install -Dm755 "${tmp_dir}/extract/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
    ;;
  Darwin)
    mount_point="${tmp_dir}/mnt"
    mkdir -p "${mount_point}"
    hdiutil attach "${tmp_dir}/${asset}" -mountpoint "${mount_point}" -nobrowse -quiet
    install -m755 "${mount_point}/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
    hdiutil detach "${mount_point}" -quiet
    ;;
esac

echo "Installed ${BIN_NAME} to ${INSTALL_DIR}/${BIN_NAME}"
if ! command -v "${BIN_NAME}" >/dev/null 2>&1; then
  echo "NOTE: ${INSTALL_DIR} is not in your PATH. Add it to your shell profile."
fi

"${INSTALL_DIR}/${BIN_NAME}" version || true
