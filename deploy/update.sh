#!/bin/bash
set -e

echo "=== Updating Smart Forms ==="

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
APP_DIR="$(dirname "$SCRIPT_DIR")"

cd "$APP_DIR"

# Pull latest code
echo "1. Pulling latest code from GitHub..."
git pull origin master

# Rebuild binary
echo "2. Rebuilding optimized binary..."
GOOS=linux \
GOARCH=amd64 \
CGO_ENABLED=0 \
go build \
  -trimpath \
  -ldflags="-s -w -extldflags '-static'" \
  -tags netgo \
  -installsuffix netgo \
  -o smart-forms-backend

chmod +x smart-forms-backend
BINARY_SIZE=$(ls -lh smart-forms-backend | awk '{print $5}')
echo "✓ Binary rebuilt: $BINARY_SIZE"

# Restart service
echo "3. Restarting service..."
sudo systemctl restart smart-forms

# Wait and verify
sleep 2
echo ""
echo "=== Verification ==="
sudo systemctl is-active smart-forms && echo "✓ Service is running" || echo "❌ Service failed to start"

curl -s http://localhost:3030/ 2>/dev/null && echo "✓ API is responding" || echo "❌ API not responding"

echo ""
echo "✅ Update complete!"
echo ""
echo "View logs: sudo journalctl -u smart-forms -f"
