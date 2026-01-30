#!/bin/bash
# Test: rpm package installation
# Platform: Linux (Fedora/RHEL/CentOS)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

source "$SCRIPT_DIR/common.sh"

info "Testing: rpm package installation"
info "Repository: $REPO_DIR"

# Check we're on Linux
if [[ "$(uname -s)" != "Linux" ]]; then
    error "This test is for Linux only"
fi

# Check for rpm
if ! command -v rpm &> /dev/null; then
    error "rpm not found - this test requires Fedora/RHEL/CentOS"
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

# Build rpm package
info "Building rpm package..."
make package-rpm

RPM_FILE="$REPO_DIR/autotidy.rpm"
if [[ ! -f "$RPM_FILE" ]]; then
    error "rpm package not created"
fi

info "Package created: $RPM_FILE"

# Install package (using sudo if available)
info "Installing rpm package..."
if command -v sudo &> /dev/null; then
    sudo rpm -i "$RPM_FILE"
else
    rpm -i "$RPM_FILE"
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
    sudo rpm -e autotidy
else
    rpm -e autotidy
fi

report_success "rpm package installation"
