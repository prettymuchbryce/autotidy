#!/bin/bash
# Test: Source build with make install (binary only)
# Platform: Linux/macOS

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

source "$SCRIPT_DIR/common.sh"

info "Testing: Source build (make install)"
info "Repository: $REPO_DIR"

# Build
info "Building binary..."
info "Current directory: $(pwd)"
info "REPO_DIR: $REPO_DIR"
cd "$REPO_DIR"
info "After cd, current directory: $(pwd)"

if ! make build; then
    error "make build failed"
fi

# Ensure binary is executable (go build should do this, but sometimes doesn't)
chmod +x "$REPO_DIR/autotidy"

# Verify binary was built
info "Checking for binary at $REPO_DIR/autotidy"
ls -la "$REPO_DIR/autotidy"

# Test binary directly (no install, to avoid needing sudo in CI)
verify_binary "$REPO_DIR/autotidy"

# Setup test environment
setup_test_environment

# Start daemon directly with the built binary
start_daemon "$REPO_DIR/autotidy"

# Run functional test
run_functional_test

# Stop daemon (cleanup trap will also do this)
stop_daemon

report_success "Source build (binary only)"
