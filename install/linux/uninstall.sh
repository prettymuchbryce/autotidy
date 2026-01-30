#!/bin/sh
# autotidy uninstaller for Linux
# Usage: curl -fsSL https://raw.githubusercontent.com/prettymuchbryce/autotidy/master/install/linux/uninstall.sh | sh
set -e

INSTALL_DIR="$HOME/.local/bin"
BINARY_PATH="$INSTALL_DIR/autotidy"

# Colors
if [ -t 1 ]; then
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    NC='\033[0m'
else
    GREEN=''
    YELLOW=''
    NC=''
fi

info() {
    printf "${GREEN}%s${NC}\n" "$1"
}

warn() {
    printf "${YELLOW}%s${NC}\n" "$1"
}

info "Uninstalling autotidy..."

# Stop and remove systemd user service
service_file="$HOME/.config/systemd/user/autotidy.service"
if [ -f "$service_file" ]; then
    systemctl --user stop autotidy 2>/dev/null || true
    systemctl --user disable autotidy 2>/dev/null || true
    rm -f "$service_file"
    systemctl --user daemon-reload
    info "Removed systemd user service"
fi

# Remove binary
if [ -f "$BINARY_PATH" ]; then
    rm -f "$BINARY_PATH"
    info "Removed $BINARY_PATH"
fi

echo ""
info "Uninstallation complete!"
echo ""
warn "Note: Configuration at ~/.config/autotidy/ was not removed."
echo "To remove it: rm -rf ~/.config/autotidy"
