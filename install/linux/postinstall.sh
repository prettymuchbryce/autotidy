#!/bin/bash
# Post-installation script for autotidy

# Reload systemd user daemon to pick up new service file
systemctl --user daemon-reload 2>/dev/null || true

echo "autotidy has been installed to /usr/local/bin/autotidy"
echo ""
echo "To start the daemon as a user service:"
echo "  systemctl --user enable --now autotidy"
echo ""
echo "To check status:"
echo "  systemctl --user status autotidy"
