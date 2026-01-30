#!/bin/bash
# Common functions for installation tests
# Source this file in test scripts: source "$(dirname "$0")/common.sh"

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test directories
export TEST_DIR="${TEST_DIR:-$(mktemp -d)}"
export CONFIG_DIR="$TEST_DIR/config"
export WATCH_DIR="$TEST_DIR/watch"
export DEST_DIR="$TEST_DIR/dest"

info() {
    echo -e "${GREEN}[INFO]${NC} $*"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

error() {
    echo -e "${RED}[ERROR]${NC} $*"
    exit 1
}

# Setup test directories and config
setup_test_environment() {
    info "Setting up test environment in $TEST_DIR"
    mkdir -p "$CONFIG_DIR" "$WATCH_DIR" "$DEST_DIR"

    # Create a simple test rule config
    cat > "$CONFIG_DIR/autotidy.yaml" << EOF
rules:
  - name: test-rule
    locations:
      - $WATCH_DIR
    filters:
      - extension: txt
    actions:
      - move:
          dest: $DEST_DIR
EOF

    info "Created test config at $CONFIG_DIR/autotidy.yaml"
}

# Verify autotidy binary works
verify_binary() {
    local autotidy_path="${1:-autotidy}"

    info "Verifying binary at: $autotidy_path"

    # For full paths, check file exists and is executable
    if [[ "$autotidy_path" == /* ]]; then
        if [[ ! -f "$autotidy_path" ]]; then
            ls -la "$(dirname "$autotidy_path")"
            error "Binary file not found: $autotidy_path"
        fi
        if [[ ! -x "$autotidy_path" ]]; then
            ls -la "$autotidy_path"
            error "Binary not executable: $autotidy_path"
        fi
    else
        # For commands (no path), check if in PATH
        if ! command -v "$autotidy_path" &> /dev/null; then
            error "Command not found in PATH: $autotidy_path"
        fi
    fi

    # Verify binary runs (use help since there's no --version flag)
    if ! "$autotidy_path" --help &> /dev/null; then
        error "Binary failed to run: $autotidy_path"
    fi
    info "Binary verified successfully"
}

# Start daemon and wait for it to be ready
start_daemon() {
    local autotidy_path="${1:-autotidy}"
    local config_path="${2:-$CONFIG_DIR/autotidy.yaml}"

    info "Starting daemon with config: $config_path"

    # Start daemon in background
    "$autotidy_path" daemon --config "$config_path" &
    DAEMON_PID=$!
    export DAEMON_PID

    # Wait for daemon to start
    sleep 2

    if ! kill -0 "$DAEMON_PID" 2>/dev/null; then
        error "Daemon failed to start"
    fi

    info "Daemon started with PID: $DAEMON_PID"
}

# Stop daemon
stop_daemon() {
    if [[ -n "${DAEMON_PID:-}" ]]; then
        info "Stopping daemon (PID: $DAEMON_PID)"
        kill "$DAEMON_PID" 2>/dev/null || true
        wait "$DAEMON_PID" 2>/dev/null || true
        unset DAEMON_PID
    fi
}

# Run functional test: create file and verify it moves
run_functional_test() {
    local test_file="$WATCH_DIR/test_$(date +%s).txt"
    local expected_dest="$DEST_DIR/$(basename "$test_file")"

    info "Creating test file: $test_file"
    echo "test content" > "$test_file"

    # Wait for daemon to process
    info "Waiting for daemon to process file..."
    local attempts=0
    local max_attempts=10

    while [[ $attempts -lt $max_attempts ]]; do
        if [[ -f "$expected_dest" ]]; then
            info "File successfully moved to: $expected_dest"
            return 0
        fi
        sleep 1
        attempts=$((attempts + 1))
    done

    error "File was not moved within ${max_attempts}s. Expected at: $expected_dest"
}

# Cleanup test environment
cleanup() {
    info "Cleaning up test environment"
    stop_daemon

    if [[ -d "$TEST_DIR" ]]; then
        rm -rf "$TEST_DIR"
    fi
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Source platform-specific helpers
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
case "$(uname -s)" in
    Linux)
        source "$SCRIPT_DIR/common_linux.sh"
        ;;
    Darwin)
        source "$SCRIPT_DIR/common_darwin.sh"
        ;;
esac

# Report test result
report_success() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  TEST PASSED: $1${NC}"
    echo -e "${GREEN}========================================${NC}"
}

report_failure() {
    echo ""
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}  TEST FAILED: $1${NC}"
    echo -e "${RED}========================================${NC}"
    exit 1
}
