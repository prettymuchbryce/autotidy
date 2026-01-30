#!/bin/bash
# Test: deb package installation
# Platform: Linux (Debian/Ubuntu)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

source "$SCRIPT_DIR/common.sh"

info "Testing: deb package installation"
info "Repository: $REPO_DIR"

# Check we're on Linux
if [[ "$(uname -s)" != "Linux" ]]; then
    error "This test is for Linux only"
fi

# Check for dpkg
if ! command -v dpkg &> /dev/null; then
    error "dpkg not found - this test requires Debian/Ubuntu"
fi

ensure_xdg_runtime_dir

# Build
info "Building binary..."
cd "$REPO_DIR"
make build

# Check for nfpm
if ! command -v nfpm &> /dev/null; then
    info "Installing nfpm..."
    go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
    export PATH="$HOME/go/bin:$PATH"
fi

# Build deb package
info "Building deb package..."
make package-deb

DEB_FILE="$REPO_DIR/autotidy.deb"
if [[ ! -f "$DEB_FILE" ]]; then
    error "deb package not created"
fi

info "Package created: $DEB_FILE"

# Install package (using sudo if available, otherwise try without)
info "Installing deb package..."
if command -v sudo &> /dev/null; then
    sudo dpkg -i "$DEB_FILE"
else
    dpkg -i "$DEB_FILE"
fi

# Verify binary is installed
verify_binary "/usr/local/bin/autotidy"

# Setup test environment
setup_test_environment

# Create and start systemd test service
create_systemd_test_service "/usr/local/bin/autotidy" "$CONFIG_DIR/autotidy.yaml"
start_systemd_test_service

# Run functional test
run_functional_test

# Cleanup
stop_systemd_test_service

# Uninstall package
if command -v sudo &> /dev/null; then
    sudo dpkg -r autotidy
else
    dpkg -r autotidy
fi

report_success "deb package installation"
