#!/usr/bin/env bash
set -euo pipefail

# OctAi Install Script
# Usage: curl -fsSL https://raw.githubusercontent.com/raynaythegreat/ai-business-hq/master/install.sh | bash

BINARY_NAME="aibhq"
INSTALL_DIR="/usr/local/bin"
REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo ""
echo "  ██████╗  ██████╗████████╗ █████╗ ██╗"
echo " ██╔═══██╗██╔════╝╚══██╔══╝██╔══██╗██║"
echo " ██║   ██║██║        ██║   ███████║██║"
echo " ╚██████╔╝╚██████╗   ██║   ██╔══██║██║"
echo "  ╚═════╝  ╚═════╝   ╚═╝   ╚═╝  ╚═╝╚═╝"
echo ""
echo "  Installing OctAi..."
echo ""

# ── Prerequisites ─────────────────────────────────────────────────────────────
command -v go >/dev/null 2>&1 || { echo "Error: Go is required. Install from https://go.dev/dl/"; exit 1; }
command -v node >/dev/null 2>&1 || { echo "Error: Node.js is required. Install from https://nodejs.org/"; exit 1; }

# ── Build frontend ────────────────────────────────────────────────────────────
echo "  [1/3] Building frontend..."
cd "$REPO_DIR/web/frontend"
if command -v pnpm >/dev/null 2>&1; then
  pnpm install --frozen-lockfile 2>/dev/null || pnpm install
  pnpm run build:backend
elif command -v npm >/dev/null 2>&1; then
  npm install
  npm run build:backend
else
  echo "Error: npm or pnpm is required."
  exit 1
fi
cd "$REPO_DIR"

# ── Build binary ──────────────────────────────────────────────────────────────
echo "  [2/3] Building binary..."
go build -mod=mod -tags "goolm,stdjson" -o "$BINARY_NAME" ./cmd/aibhq

# ── Install ───────────────────────────────────────────────────────────────────
echo "  [3/3] Installing to $INSTALL_DIR/$BINARY_NAME..."
if [ -w "$INSTALL_DIR" ]; then
  cp "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
else
  sudo cp "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
fi

echo ""
echo "  ✓ OctAi installed successfully!"
echo ""
echo "  Run 'aibhq onboard' to set up your configuration."
echo "  Run 'aibhq web' to start the web interface."
echo ""
