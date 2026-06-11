#!/usr/bin/env bash
set -euo pipefail

# Deploy script for core-utils-cli
# Usage: ./scripts/deploy-remote.sh
# Run this on the server (e.g., root@api.core-utils.dev or code@api.core-utils.dev)

INSTALL_DIR="/opt/core-utils-cli"
BIN_DIR="/usr/local/bin"

echo "=== Core Utils CLI Deploy ==="

if [ ! -d "$INSTALL_DIR" ]; then
    echo "[1/4] Cloning repository..."
    git clone git@github.com:mdwcoder/core-utils-cli.git "$INSTALL_DIR"
else
    echo "[1/4] Pulling latest changes..."
    cd "$INSTALL_DIR"
    git pull origin main
fi

cd "$INSTALL_DIR"

echo "[2/4] Building cu binary..."
chmod +x scripts/build.sh
./scripts/build.sh

echo "[3/4] Installing binary to $BIN_DIR..."
cp bin/cu "$BIN_DIR/cu"
chmod +x "$BIN_DIR/cu"

echo "[4/4] Verifying installation..."
cu version

echo ""
echo "=== Deploy complete ==="
echo "cu is now available at: $BIN_DIR/cu"
echo "Test with: cu doctor"
