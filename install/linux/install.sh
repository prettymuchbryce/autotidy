#!/bin/sh
# autotidy installer for Linux
# Usage: curl -fsSL https://raw.githubusercontent.com/prettymuchbryce/autotidy/master/install/linux/install.sh | sh
#
# For testing with a local binary:
#   ./install.sh /path/to/autotidy
set -e

REPO="prettymuchbryce/autotidy"
INSTALL_DIR="$HOME/.local/bin"
BINARY_PATH="$INSTALL_DIR/autotidy"
LOCAL_BINARY="${1:-}"

# Colors (if terminal supports it)
if [ -t 1 ]; then
    GREEN='\033[0;32m'
    RED='\033[0;31m'
    NC='\033[0m'
else
    GREEN=''
    RED=''
    NC=''
fi

info() {
    printf "${GREEN}%s${NC}\n" "$1"
}

error() {
    printf "${RED}ERROR: %s${NC}\n" "$1" >&2
    exit 1
}

detect_arch() {
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac
}

get_latest_version() {
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        error "curl or wget is required"
    fi
}

download() {
    url="$1"
    output="$2"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$url" -o "$output"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$url" -O "$output"
    fi
}

main() {
    # Check we're on Linux
    if [ "$(uname -s)" != "Linux" ]; then
        error "This installer is for Linux only. For macOS, use: brew install prettymuchbryce/tap/autotidy"
    fi

    info "Installing autotidy..."

    mkdir -p "$INSTALL_DIR"

    if [ -n "$LOCAL_BINARY" ]; then
        # Use local binary (for testing)
        if [ ! -f "$LOCAL_BINARY" ]; then
            error "Binary not found at $LOCAL_BINARY"
        fi
        cp "$LOCAL_BINARY" "$BINARY_PATH"
        chmod +x "$BINARY_PATH"
        info "Installed autotidy from local binary"
        version="local"
    else
        # Download from GitHub releases
        arch=$(detect_arch)
        version=$(get_latest_version)

        if [ -z "$version" ]; then
            error "Could not determine latest version"
        fi

        # Version without 'v' prefix for artifact name
        version_num="${version#v}"
        platform="linux_${arch}"

        info "Downloading autotidy $version for $platform..."

        tmpdir=$(mktemp -d)
        trap 'rm -rf "$tmpdir"' EXIT

        archive_url="https://github.com/$REPO/releases/download/$version/autotidy_${version_num}_${platform}.tar.gz"
        archive_path="$tmpdir/autotidy.tar.gz"

        download "$archive_url" "$archive_path" || error "Failed to download from $archive_url"

        tar -xzf "$archive_path" -C "$tmpdir"
        mv "$tmpdir/autotidy" "$BINARY_PATH"
        chmod +x "$BINARY_PATH"

        info "Installed autotidy to $BINARY_PATH"
    fi

    # Setup systemd user service
    service_dir="$HOME/.config/systemd/user"
    service_file="$service_dir/autotidy.service"

    mkdir -p "$service_dir"

    cat > "$service_file" << EOF
[Unit]
Description=autotidy daemon
After=default.target

[Service]
Type=notify
ExecStart=$BINARY_PATH daemon
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
EOF

    info "Created systemd user service"

    # Reload and enable service
    systemctl --user daemon-reload
    systemctl --user enable autotidy
    systemctl --user start autotidy

    info "Started autotidy daemon"

    # Update PATH if needed
    case ":$PATH:" in
        *":$INSTALL_DIR:"*) ;;
        *)
            shell_name=$(basename "$SHELL")
            case "$shell_name" in
                bash)
                    if [ -f "$HOME/.bashrc" ]; then
                        rc_file="$HOME/.bashrc"
                    else
                        rc_file="$HOME/.bash_profile"
                    fi
                    ;;
                zsh) rc_file="$HOME/.zshrc" ;;
                fish) rc_file="$HOME/.config/fish/config.fish" ;;
                *) rc_file="$HOME/.profile" ;;
            esac

            if ! grep -q "$INSTALL_DIR" "$rc_file" 2>/dev/null; then
                if [ "$shell_name" = "fish" ]; then
                    mkdir -p "$(dirname "$rc_file")"
                    echo "fish_add_path $INSTALL_DIR" >> "$rc_file"
                else
                    echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$rc_file"
                fi
                info "Added $INSTALL_DIR to PATH in $rc_file"
                echo "    Restart your shell or run: export PATH=\"$INSTALL_DIR:\$PATH\""
            fi
            ;;
    esac

    echo ""
    info "Installation complete!"
    echo ""
    echo "autotidy $version is now running."
    echo ""
    echo "To check status:  autotidy status"
    echo "To view config:   \$EDITOR ~/.config/autotidy/config.yaml"
}

main "$@"
