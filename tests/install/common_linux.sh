#!/bin/bash
# Linux-specific test helpers (systemd)
# Sourced by common.sh on Linux systems

# Ensure XDG_RUNTIME_DIR is set (required for systemd --user)
ensure_xdg_runtime_dir() {
    if [[ -z "${XDG_RUNTIME_DIR:-}" ]]; then
        export XDG_RUNTIME_DIR="/run/user/$(id -u)"
        info "Set XDG_RUNTIME_DIR=$XDG_RUNTIME_DIR"
    fi
}

# Create systemd test service file
# Usage: create_systemd_test_service <binary_path> <config_path> [service_name]
create_systemd_test_service() {
    local binary_path="$1"
    local config_path="$2"
    local service_name="${3:-autotidy-test}"

    SYSTEMD_USER_DIR="$HOME/.config/systemd/user"
    mkdir -p "$SYSTEMD_USER_DIR"

    SYSTEMD_SERVICE_FILE="$SYSTEMD_USER_DIR/${service_name}.service"
    export SYSTEMD_SERVICE_FILE
    export SYSTEMD_SERVICE_NAME="$service_name"

    cat > "$SYSTEMD_SERVICE_FILE" << EOF
[Unit]
Description=autotidy daemon (test)

[Service]
Type=simple
ExecStart=$binary_path daemon --config $config_path
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
EOF

    info "Created systemd service at $SYSTEMD_SERVICE_FILE"
}

# Start systemd test service
start_systemd_test_service() {
    local service_name="${SYSTEMD_SERVICE_NAME:-autotidy-test}"

    systemctl --user daemon-reload
    systemctl --user start "$service_name"

    sleep 2

    if ! systemctl --user is-active "$service_name" > /dev/null 2>&1; then
        systemctl --user status "$service_name" || true
        error "Service failed to start"
    fi

    info "Service started successfully"
}

# Stop and cleanup systemd test service
stop_systemd_test_service() {
    local service_name="${SYSTEMD_SERVICE_NAME:-autotidy-test}"
    local service_file="${SYSTEMD_SERVICE_FILE:-}"

    info "Stopping systemd service..."
    systemctl --user stop "$service_name" 2>/dev/null || true
    systemctl --user disable "$service_name" 2>/dev/null || true

    if [[ -n "$service_file" && -f "$service_file" ]]; then
        rm -f "$service_file"
    fi

    systemctl --user daemon-reload
}
