#!/bin/bash
# Test: Source build with systemd user service
# Platform: Linux

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

source "$SCRIPT_DIR/common.sh"

info "Testing: Source build with systemd user service"
info "Repository: $REPO_DIR"

# Check we're on Linux
if [[ "$(uname -s)" != "Linux" ]]; then
    error "This test is for Linux only"
fi

ensure_xdg_runtime_dir

# Build
info "Building binary..."
cd "$REPO_DIR"
make build

# Install binary to a local prefix (avoid needing sudo)
LOCAL_PREFIX="$TEST_DIR/local"
mkdir -p "$LOCAL_PREFIX/bin"
cp "$REPO_DIR/autotidy" "$LOCAL_PREFIX/bin/"
chmod +x "$LOCAL_PREFIX/bin/autotidy"
export PATH="$LOCAL_PREFIX/bin:$PATH"

# Verify binary
verify_binary "autotidy"

# Setup test environment
setup_test_environment

# Create and start systemd test service
create_systemd_test_service "$LOCAL_PREFIX/bin/autotidy" "$CONFIG_DIR/autotidy.yaml"
start_systemd_test_service

# Run functional test
run_functional_test

# Stop and cleanup service
stop_systemd_test_service

report_success "Source build with systemd"
