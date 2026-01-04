#!/bin/bash

echo "=== Smart Forms Logs ==="
echo "Press Ctrl+C to exit"
echo ""

# Follow logs in real-time
sudo journalctl -u smart-forms -f --no-pager
