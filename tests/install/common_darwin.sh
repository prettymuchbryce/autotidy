#!/bin/bash
# macOS-specific test helpers (launchd)
# Sourced by common.sh on Darwin systems

# Create launchd test agent plist
# Usage: create_launchd_test_agent <binary_path> <config_path> <log_dir> [label]
create_launchd_test_agent() {
    local binary_path="$1"
    local config_path="$2"
    local log_dir="$3"
    local label="${4:-com.autotidy.test}"

    LAUNCHD_AGENTS_DIR="$HOME/Library/LaunchAgents"
    mkdir -p "$LAUNCHD_AGENTS_DIR"

    LAUNCHD_PLIST_FILE="$LAUNCHD_AGENTS_DIR/${label}.plist"
    export LAUNCHD_PLIST_FILE
    export LAUNCHD_LABEL="$label"

    cat > "$LAUNCHD_PLIST_FILE" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>$label</string>
    <key>ProgramArguments</key>
    <array>
        <string>$binary_path</string>
        <string>daemon</string>
        <string>--config</string>
        <string>$config_path</string>
    </array>
    <key>RunAtLoad</key>
    <false/>
    <key>KeepAlive</key>
    <false/>
    <key>StandardOutPath</key>
    <string>$log_dir/autotidy.out.log</string>
    <key>StandardErrorPath</key>
    <string>$log_dir/autotidy.err.log</string>
</dict>
</plist>
EOF

    info "Created launchd plist at $LAUNCHD_PLIST_FILE"
}

# Start launchd test agent
start_launchd_test_agent() {
    local label="${LAUNCHD_LABEL:-com.autotidy.test}"
    local plist_file="${LAUNCHD_PLIST_FILE:-}"

    launchctl load "$plist_file"
    launchctl start "$label"

    sleep 2

    if ! pgrep -f "autotidy daemon" > /dev/null; then
        cat "$TEST_DIR/autotidy.err.log" 2>/dev/null || true
        error "Agent failed to start"
    fi

    info "Agent started successfully"
}

# Stop and cleanup launchd test agent
stop_launchd_test_agent() {
    local label="${LAUNCHD_LABEL:-com.autotidy.test}"
    local plist_file="${LAUNCHD_PLIST_FILE:-}"

    info "Stopping launchd agent..."
    launchctl stop "$label" 2>/dev/null || true
    launchctl unload "$plist_file" 2>/dev/null || true

    if [[ -n "$plist_file" && -f "$plist_file" ]]; then
        rm -f "$plist_file"
    fi
}
