#!/bin/bash
set -e

# Rize installer script
# Usage: curl -sSL https://raw.githubusercontent.com/alienxp03/rize/master/scripts/install.sh | sh

INSTALL_DIR="/usr/local/bin"
BINARY_NAME="rize"
REPO="alienxp03/rize"
GITHUB_URL="https://github.com/${REPO}"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

info() {
    echo -e "${BLUE}→${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

warning() {
    echo -e "${YELLOW}!${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
    exit 1
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            ;;
    esac

    case "$OS" in
        linux|darwin)
            ;;
        *)
            error "Unsupported OS: $OS"
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    info "Detected platform: $PLATFORM"
}

# Get latest release version
get_latest_version() {
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        error "curl or wget is required"
    fi

    if [ -z "$VERSION" ]; then
        error "Failed to get latest version"
    fi

    info "Latest version: $VERSION"
}

# Download binary
download_binary() {
    DOWNLOAD_URL="${GITHUB_URL}/releases/download/${VERSION}/${BINARY_NAME}-${PLATFORM}"
    TMP_FILE=$(mktemp)

    info "Downloading from $DOWNLOAD_URL..."

    if command -v curl >/dev/null 2>&1; then
        curl -sSL "$DOWNLOAD_URL" -o "$TMP_FILE" || error "Download failed"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$TMP_FILE" "$DOWNLOAD_URL" || error "Download failed"
    fi

    success "Downloaded successfully"
}

# Install binary
install_binary() {
    info "Installing to ${INSTALL_DIR}/${BINARY_NAME}..."

    # Make executable
    chmod +x "$TMP_FILE"

    # Install
    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
    elif command -v sudo >/dev/null 2>&1; then
        warning "Elevated permissions required"
        sudo mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        error "Cannot write to $INSTALL_DIR and sudo is not available"
    fi

    success "Installed to ${INSTALL_DIR}/${BINARY_NAME}"
}

# Check prerequisites
check_prerequisites() {
    info "Checking prerequisites..."

    # Check for Docker
    if ! command -v docker >/dev/null 2>&1; then
        warning "Docker is not installed"
        info "Please install Docker: https://docs.docker.com/get-docker/"
    else
        success "Docker found"
    fi

    # Check for docker compose
    if ! docker compose version >/dev/null 2>&1; then
        warning "Docker Compose is not available"
        info "Please install Docker Compose: https://docs.docker.com/compose/install/"
    else
        success "Docker Compose found"
    fi
}

# Main installation
main() {
    echo ""
    echo "Rize Installer"
    echo "=============="
    echo ""

    detect_platform
    get_latest_version
    download_binary
    install_binary
    check_prerequisites

    echo ""
    success "Installation complete!"
    echo ""
    info "Get started:"
    echo "  rize init         # Create config file"
    echo "  rize shell        # Start interactive shell"
    echo "  rize claude       # Run Claude Code"
    echo "  rize help         # Show all commands"
    echo ""
}

main
