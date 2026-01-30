#!/bin/bash
# Test: Source build with launchd user agent
# Platform: macOS

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

source "$SCRIPT_DIR/common.sh"

info "Testing: Source build with launchd user agent"
info "Repository: $REPO_DIR"

# Check we're on macOS
if [[ "$(uname -s)" != "Darwin" ]]; then
    error "This test is for macOS only"
fi

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

# Create and start launchd test agent
create_launchd_test_agent "$LOCAL_PREFIX/bin/autotidy" "$CONFIG_DIR/autotidy.yaml" "$TEST_DIR"
start_launchd_test_agent

# Run functional test
run_functional_test

# Stop and cleanup agent
stop_launchd_test_agent

report_success "Source build with launchd"
