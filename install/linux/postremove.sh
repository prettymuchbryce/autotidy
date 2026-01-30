#!/bin/bash
# Post-removal script for autotidy

# Reload systemd user daemon to clean up removed service file
systemctl --user daemon-reload 2>/dev/null || true
