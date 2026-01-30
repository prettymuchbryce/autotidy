#!/bin/bash
# Test: Nix flake package build
# Platform: Linux/macOS (with Nix installed)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

source "$SCRIPT_DIR/common.sh"

info "Testing: Nix flake package"
info "Repository: $REPO_DIR"

# Check for Nix
if ! command -v nix &> /dev/null; then
    error "Nix not found - install via DeterminateSystems/nix-installer-action"
fi

cd "$REPO_DIR"

# Build the package
info "Building Nix package..."
nix build .#default --no-link --print-out-paths

# Get the output path
NIX_RESULT=$(nix build .#default --no-link --print-out-paths)
AUTOTIDY_BIN="$NIX_RESULT/bin/autotidy"

if [[ ! -x "$AUTOTIDY_BIN" ]]; then
    error "Binary not found at: $AUTOTIDY_BIN"
fi

# Verify binary
verify_binary "$AUTOTIDY_BIN"

# Setup test environment
setup_test_environment

# Start daemon
start_daemon "$AUTOTIDY_BIN"

# Run functional test
run_functional_test

# Stop daemon
stop_daemon

report_success "Nix flake package"
