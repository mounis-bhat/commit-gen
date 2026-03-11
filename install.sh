#!/bin/sh
# commit-gen installer for Linux and macOS
# Usage: curl -sSfL https://raw.githubusercontent.com/mounis-bhat/commit-gen/trunk/install.sh | sh

set -e

REPO="mounis-bhat/commit-gen"
BINARY_NAME="commit-gen"
INSTALL_DIR="${HOME}/.local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    printf "${GREEN}[INFO]${NC} %s\n" "$1"
}

warn() {
    printf "${YELLOW}[WARN]${NC} %s\n" "$1"
}

error() {
    printf "${RED}[ERROR]${NC} %s\n" "$1"
    exit 1
}

# Detect OS
detect_os() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$OS" in
        linux*)  OS="linux" ;;
        darwin*) OS="darwin" ;;
        *)       error "Unsupported operating system: $OS" ;;
    esac
    echo "$OS"
}

# Detect architecture
detect_arch() {
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64)  ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *)             error "Unsupported architecture: $ARCH" ;;
    esac
    echo "$ARCH"
}

# Get latest release tag from GitHub
get_latest_version() {
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -sSf "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        error "Neither curl nor wget found. Please install one of them."
    fi
    
    if [ -z "$VERSION" ]; then
        error "Could not determine latest version. Please check https://github.com/${REPO}/releases"
    fi
    
    echo "$VERSION"
}

# Add INSTALL_DIR to PATH in shell profile files if not already present
setup_path() {
    case ":$PATH:" in
        *":${INSTALL_DIR}:"*) return ;;
    esac

    EXPORT_LINE="export PATH=\"\$HOME/.local/bin:\$PATH\""
    MODIFIED=""

    for profile in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile"; do
        if [ -f "$profile" ]; then
            if ! grep -qF "$EXPORT_LINE" "$profile" 2>/dev/null; then
                printf "\n# Added by commit-gen installer\n%s\n" "$EXPORT_LINE" >> "$profile"
                MODIFIED="$MODIFIED $profile"
            fi
        fi
    done

    if [ -n "$MODIFIED" ]; then
        warn "${INSTALL_DIR} was not in your PATH."
        info "Added to:$MODIFIED"
        echo ""
        echo "Restart your terminal or run:"
        for f in $MODIFIED; do echo "    source $f"; done
    fi
}

# Download and install
install() {
    OS=$(detect_os)
    ARCH=$(detect_arch)
    VERSION=$(get_latest_version)

    info "Detected OS: $OS, Architecture: $ARCH"

    if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        OLD_VERSION=$("${INSTALL_DIR}/${BINARY_NAME}" --version 2>/dev/null || echo "unknown")
        info "Upgrading existing installation (${OLD_VERSION} -> ${VERSION})..."
    else
        info "Installing ${BINARY_NAME} ${VERSION}..."
    fi

    FILENAME="${BINARY_NAME}-${OS}-${ARCH}.tar.gz"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

    info "Downloading ${FILENAME}..."

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap 'rm -rf "$TMP_DIR"' EXIT

    # Download
    if command -v curl >/dev/null 2>&1; then
        curl -sSfL "$DOWNLOAD_URL" -o "${TMP_DIR}/${FILENAME}" || error "Download failed. Please check if the release exists."
    else
        wget -q "$DOWNLOAD_URL" -O "${TMP_DIR}/${FILENAME}" || error "Download failed. Please check if the release exists."
    fi

    # Extract
    info "Extracting..."
    tar -xzf "${TMP_DIR}/${FILENAME}" -C "$TMP_DIR"

    # Install
    info "Installing to ${INSTALL_DIR}..."
    mkdir -p "$INSTALL_DIR"
    mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

    info "Successfully installed ${BINARY_NAME} to ${INSTALL_DIR}/${BINARY_NAME}"

    setup_path

    echo ""
    info "commit-gen is ready. Run 'commit-gen' in any git repo."
}

install
