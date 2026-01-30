#!/bin/bash
# Pre-removal script for autotidy

# Stop the service if running
systemctl --user stop autotidy 2>/dev/null || true
systemctl --user disable autotidy 2>/dev/null || true
