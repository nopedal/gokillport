#!/bin/sh
set -e

# gokillport installer
# Usage: curl -sSL https://raw.githubusercontent.com/nikolasopedal/gokillport/main/install.sh | sh

REPO="nikolasopedal/gokillport"
INSTALL_DIR="/usr/local/bin"
BINARY="gokillport"
ALIAS="killport"

# Colors (only if terminal supports it)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    DIM='\033[0;90m'
    RESET='\033[0m'
else
    RED='' GREEN='' DIM='' RESET=''
fi

info()  { printf "${GREEN}▸${RESET} %s\n" "$1"; }
error() { printf "${RED}error:${RESET} %s\n" "$1" >&2; exit 1; }

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux)  OS="linux" ;;
        darwin) OS="darwin" ;;
        *)      error "unsupported OS: $OS" ;;
    esac

    case "$ARCH" in
        x86_64|amd64)  ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *)             error "unsupported architecture: $ARCH" ;;
    esac
}

# Get latest release tag from GitHub
get_latest_version() {
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        error "curl or wget is required"
    fi

    if [ -z "$VERSION" ]; then
        error "could not determine latest version. Check https://github.com/${REPO}/releases"
    fi
}

# Download and install
install() {
    detect_platform
    get_latest_version

    FILENAME="${BINARY}-${OS}-${ARCH}"
    URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

    info "downloading ${BINARY} ${VERSION} for ${OS}/${ARCH}"

    TMPDIR=$(mktemp -d)
    trap 'rm -rf "$TMPDIR"' EXIT

    if command -v curl >/dev/null 2>&1; then
        curl -sSL "$URL" -o "${TMPDIR}/${BINARY}" || error "download failed. Check ${URL}"
    else
        wget -q "$URL" -O "${TMPDIR}/${BINARY}" || error "download failed. Check ${URL}"
    fi

    chmod +x "${TMPDIR}/${BINARY}"

    # Install to /usr/local/bin (may need sudo)
    if [ -w "$INSTALL_DIR" ]; then
        mv "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
        ln -sf "${INSTALL_DIR}/${BINARY}" "${INSTALL_DIR}/${ALIAS}"
    else
        info "installing to ${INSTALL_DIR} (requires sudo)"
        sudo mv "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
        sudo ln -sf "${INSTALL_DIR}/${BINARY}" "${INSTALL_DIR}/${ALIAS}"
    fi

    info "installed ${BINARY} ${VERSION} to ${INSTALL_DIR}/${BINARY}"
    info "also available as: ${ALIAS}"
    printf "\n${DIM}  usage: killport 3000${RESET}\n\n"
}

install
