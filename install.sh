#!/bin/bash
set -e

# Conduit Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/kraitsura/conduit/main/install.sh | bash

REPO="kraitsura/conduit"
INSTALL_DIR="/usr/local/bin"
BINARY="conduit"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

echo "Installing Conduit for $OS/$ARCH..."

# Get latest release
LATEST=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/' || echo "")

if [ -z "$LATEST" ]; then
    echo "Could not determine latest version."
    echo "Download manually from: https://github.com/$REPO/releases"
    exit 1
fi

echo "Version: $LATEST"

# Download binary
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST/${BINARY}-${OS}-${ARCH}"

echo "Downloading $BINARY..."
curl -fsSL "$DOWNLOAD_URL" -o /tmp/$BINARY
chmod +x /tmp/$BINARY

# Install (may need sudo)
if [ -w "$INSTALL_DIR" ]; then
    mv /tmp/$BINARY "$INSTALL_DIR/$BINARY"
else
    echo "Installing to $INSTALL_DIR (requires sudo)..."
    sudo mv /tmp/$BINARY "$INSTALL_DIR/$BINARY"
fi

echo ""
echo "Conduit installed successfully!"
echo ""
echo "Quick start:"
echo "  conduit init          # Initialize config"
echo "  conduit daemon start  # Start daemon"
echo "  conduit status        # Check status"
echo ""
