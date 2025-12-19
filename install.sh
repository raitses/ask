#!/bin/bash
set -e

# ask installation script
# Usage: curl -sSL https://raw.githubusercontent.com/raitses/ask/main/install.sh | bash

REPO="raitses/ask"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect OS and architecture
detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case "$OS" in
        Linux*)
            OS="linux"
            ;;
        Darwin*)
            OS="darwin"
            ;;
        MINGW*|MSYS*|CYGWIN*)
            OS="windows"
            ;;
        *)
            echo -e "${RED}Unsupported operating system: $OS${NC}"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            echo -e "${RED}Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac

    echo -e "${GREEN}Detected platform: $OS/$ARCH${NC}"
}

# Get the latest release version
get_latest_version() {
    echo -e "${YELLOW}Fetching latest release...${NC}"
    LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$LATEST_VERSION" ]; then
        echo -e "${RED}Failed to fetch latest version${NC}"
        exit 1
    fi

    echo -e "${GREEN}Latest version: $LATEST_VERSION${NC}"
}

# Download and install
install_ask() {
    BINARY_NAME="ask"
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="ask.exe"
    fi

    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_VERSION/ask_${LATEST_VERSION#v}_${OS}_${ARCH}.tar.gz"

    if [ "$OS" = "windows" ]; then
        DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_VERSION/ask_${LATEST_VERSION#v}_${OS}_${ARCH}.zip"
    fi

    echo -e "${YELLOW}Downloading from: $DOWNLOAD_URL${NC}"

    # Create temporary directory
    TMP_DIR=$(mktemp -d)
    cd "$TMP_DIR"

    # Download
    if ! curl -sL "$DOWNLOAD_URL" -o "ask-archive"; then
        echo -e "${RED}Failed to download $DOWNLOAD_URL${NC}"
        exit 1
    fi

    # Extract
    echo -e "${YELLOW}Extracting...${NC}"
    if [ "$OS" = "windows" ]; then
        unzip -q ask-archive
    else
        tar -xzf ask-archive
    fi

    # Install
    echo -e "${YELLOW}Installing to $INSTALL_DIR...${NC}"

    if [ -w "$INSTALL_DIR" ]; then
        mv "$BINARY_NAME" "$INSTALL_DIR/ask"
        chmod +x "$INSTALL_DIR/ask"
    else
        echo -e "${YELLOW}Root permissions required to install to $INSTALL_DIR${NC}"
        sudo mv "$BINARY_NAME" "$INSTALL_DIR/ask"
        sudo chmod +x "$INSTALL_DIR/ask"
    fi

    # Cleanup
    cd -
    rm -rf "$TMP_DIR"

    echo -e "${GREEN}✓ Installation complete!${NC}"
}

# Verify installation
verify_installation() {
    if command -v ask &> /dev/null; then
        echo -e "${GREEN}✓ ask is installed successfully${NC}"
        echo ""
        ask --version
        echo ""
        echo -e "${YELLOW}Next steps:${NC}"
        echo "1. Set up your API key: export ASK_API_KEY='your-api-key'"
        echo "2. Or create a config file: ~/.config/ask/.env"
        echo "3. Run 'ask --help' for usage information"
    else
        echo -e "${RED}Installation verification failed${NC}"
        echo "Please make sure $INSTALL_DIR is in your PATH"
        exit 1
    fi
}

# Main
main() {
    echo "========================================="
    echo "       ask CLI Installer"
    echo "========================================="
    echo ""

    detect_platform
    get_latest_version
    install_ask
    verify_installation
}

main
