#!/bin/bash
# Test: Linux install.sh script
# Platform: Linux
#
# Tests the one-liner install script

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

source "$SCRIPT_DIR/common.sh"

info "Testing: Linux install.sh script"
info "Repository: $REPO_DIR"

# Check we're on Linux
if [[ "$(uname -s)" != "Linux" ]]; then
    error "This test is for Linux only"
fi

ensure_xdg_runtime_dir

# Build the binary first
info "Building binary..."
cd "$REPO_DIR"
make build

BINARY_PATH="$REPO_DIR/autotidy"
INSTALL_SCRIPT="$REPO_DIR/install/linux/install.sh"
UNINSTALL_SCRIPT="$REPO_DIR/install/linux/uninstall.sh"

# Verify binary was built
if [[ ! -x "$BINARY_PATH" ]]; then
    error "Binary not found at $BINARY_PATH"
fi

# Setup test environment
setup_test_environment

INSTALLED_BINARY="$HOME/.local/bin/autotidy"
SERVICE_FILE="$HOME/.config/systemd/user/autotidy.service"

# Clean any existing installation first
"$UNINSTALL_SCRIPT" 2>/dev/null || true

# Remove any PATH entries from rc files to start clean
for rc_file in "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.zshrc" "$HOME/.profile"; do
    if [[ -f "$rc_file" ]]; then
        grep -v '\.local/bin' "$rc_file" >"${rc_file}.tmp" 2>/dev/null || true
        mv "${rc_file}.tmp" "$rc_file" 2>/dev/null || true
    fi
done

# ============================================================================
# Test 1: Basic installation
# ============================================================================
info "Test 1: Basic installation"

"$INSTALL_SCRIPT" "$BINARY_PATH"

# Verify binary was installed
if [[ ! -x "$INSTALLED_BINARY" ]]; then
    error "Binary not installed at $INSTALLED_BINARY"
fi
info "  Binary installed: $INSTALLED_BINARY"

# Verify systemd service was created
if [[ ! -f "$SERVICE_FILE" ]]; then
    error "Systemd service file not created"
fi
info "  Systemd service created"

# Verify service is running
if ! systemctl --user is-active autotidy >/dev/null 2>&1; then
    systemctl --user status autotidy || true
    error "Service is not running"
fi
info "  Service is running"

# Verify binary works
if ! "$INSTALLED_BINARY" --help >/dev/null 2>&1; then
    error "Installed binary doesn't work"
fi
info "  Binary is functional"

# Check PATH was added to shell rc file
path_added=false
for rc_file in "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.zshrc" "$HOME/.profile"; do
    if [[ -f "$rc_file" ]] && grep -q '\.local/bin' "$rc_file" 2>/dev/null; then
        info "  PATH added to $rc_file"
        path_added=true
        break
    fi
done
if [[ "$path_added" == "false" ]]; then
    info "  PATH not added (already in PATH or no rc file found)"
fi

# ============================================================================
# Test 2: Functional test
# ============================================================================
info "Test 2: Functional test"

# Stop the installed service, start with test config
systemctl --user stop autotidy

"$INSTALLED_BINARY" daemon --config "$CONFIG_DIR/autotidy.yaml" &
DAEMON_PID=$!
sleep 2

if ! kill -0 "$DAEMON_PID" 2>/dev/null; then
    error "Daemon process died"
fi

run_functional_test

kill "$DAEMON_PID" 2>/dev/null || true
wait "$DAEMON_PID" 2>/dev/null || true
info "  Functional test passed"

# ============================================================================
# Test 3: Uninstall
# ============================================================================
info "Test 3: Uninstall"

"$UNINSTALL_SCRIPT"

if [[ -f "$INSTALLED_BINARY" ]]; then
    error "Binary was not removed"
fi
info "  Binary removed"

if [[ -f "$SERVICE_FILE" ]]; then
    error "Service file was not removed"
fi
info "  Service file removed"

if systemctl --user is-active autotidy >/dev/null 2>&1; then
    error "Service is still running"
fi
info "  Service stopped"

# ============================================================================
# Test 4: Reinstall (upgrade scenario)
# ============================================================================
info "Test 4: Reinstall (upgrade scenario)"

"$INSTALL_SCRIPT" "$BINARY_PATH"

if [[ ! -x "$INSTALLED_BINARY" ]]; then
    error "Binary not installed after reinstall"
fi

if ! systemctl --user is-active autotidy >/dev/null 2>&1; then
    error "Service not running after reinstall"
fi
info "  Reinstall successful"

# ============================================================================
# Test 5: Shell-specific PATH detection
# ============================================================================
info "Test 5: Shell-specific PATH detection"

# Uninstall first
"$UNINSTALL_SCRIPT"

# The install script only adds PATH to rc files if ~/.local/bin is NOT already
# in $PATH. On many CI systems, it's already in PATH, so we need to test with
# PATH modified to not include it.

# Save original PATH
ORIGINAL_PATH="$PATH"

# Remove ~/.local/bin from PATH for this test
CLEAN_PATH=$(echo "$PATH" | tr ':' '\n' | grep -v '\.local/bin' | tr '\n' ':' | sed 's/:$//')

# Test each available shell
for shell_name in bash zsh fish; do
    shell_path=$(command -v "$shell_name" 2>/dev/null || true)
    if [[ -z "$shell_path" ]]; then
        info "  $shell_name: not installed, skipping"
        continue
    fi

    # Determine rc file for this shell
    case "$shell_name" in
    bash)
        rc_file="$HOME/.bashrc"
        ;;
    zsh)
        rc_file="$HOME/.zshrc"
        ;;
    fish)
        rc_file="$HOME/.config/fish/config.fish"
        ;;
    esac

    # Create empty rc file if it doesn't exist
    mkdir -p "$(dirname "$rc_file")"
    touch "$rc_file"

    # Remove any existing .local/bin entries
    grep -v '\.local/bin' "$rc_file" >"${rc_file}.tmp" 2>/dev/null || true
    mv "${rc_file}.tmp" "$rc_file"

    # Run install with this shell and cleaned PATH
    PATH="$CLEAN_PATH" SHELL="$shell_path" "$INSTALL_SCRIPT" "$BINARY_PATH" >/dev/null 2>&1

    # Check if PATH was added
    if grep -q '\.local/bin\|fish_add_path' "$rc_file" 2>/dev/null; then
        info "  $shell_name: PATH added to $rc_file"
    else
        error "  $shell_name: PATH not added to $rc_file"
    fi

    # Uninstall for next iteration
    "$UNINSTALL_SCRIPT" >/dev/null 2>&1
done

# Restore original PATH
export PATH="$ORIGINAL_PATH"

# ============================================================================
# Cleanup
# ============================================================================
info "Cleaning up..."
"$UNINSTALL_SCRIPT" 2>/dev/null || true

report_success "Linux install.sh script"
