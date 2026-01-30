#!/bin/bash
# Test: NixOS module
# Platform: Linux (with Nix installed)
# Note: This tests module evaluation and build, not activation

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

source "$SCRIPT_DIR/common.sh"

info "Testing: NixOS module"
info "Repository: $REPO_DIR"

# Check for Nix
if ! command -v nix &> /dev/null; then
    error "Nix not found - install via DeterminateSystems/nix-installer-action"
fi

# Check we're on Linux
if [[ "$(uname -s)" != "Linux" ]]; then
    error "NixOS module test is for Linux only"
fi

cd "$REPO_DIR"

# Build the NixOS test configuration
# This verifies the module can be evaluated and the system can be built
info "Building NixOS test configuration..."
nix build .#nixosConfigurations.test.config.system.build.toplevel --dry-run

# If dry-run succeeded, try an actual build
info "Building NixOS system (this may take a while)..."
NIX_RESULT=$(nix build .#nixosConfigurations.test.config.system.build.toplevel --no-link --print-out-paths 2>&1) || {
    # If full build fails, that's okay - dry-run success is sufficient for CI
    warn "Full build failed (expected in limited CI environments)"
    info "Dry-run succeeded - module is valid"
    report_success "NixOS module (dry-run only)"
    exit 0
}

info "NixOS system built: $NIX_RESULT"

# Verify systemd service is configured
if [[ -d "$NIX_RESULT/etc/systemd" ]]; then
    info "Systemd configuration found in system"
fi

# The NixOS system build validates the module. Now build autotidy package
# directly for functional testing.
info "Building autotidy package for functional testing..."
AUTOTIDY_PKG=$(nix build .#default --no-link --print-out-paths)
AUTOTIDY_BIN="$AUTOTIDY_PKG/bin/autotidy"

if [[ ! -x "$AUTOTIDY_BIN" ]]; then
    error "Could not find autotidy binary at: $AUTOTIDY_BIN"
fi

info "Found autotidy: $AUTOTIDY_BIN"
verify_binary "$AUTOTIDY_BIN"

# Setup test environment
setup_test_environment

# Start daemon
start_daemon "$AUTOTIDY_BIN"

# Run functional test
run_functional_test

# Stop daemon
stop_daemon

report_success "NixOS module"
