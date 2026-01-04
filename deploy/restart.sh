#!/bin/bash

echo "=== Restarting Smart Forms Service ==="

# Restart the service
echo "Restarting smart-forms service..."
sudo systemctl restart smart-forms

# Wait a moment for service to start
sleep 2

# Check status
echo ""
echo "=== Service Status ==="
sudo systemctl status smart-forms --no-pager -l | head -10

# Test if API is responding
echo ""
echo "=== Testing API ==="
curl -s http://localhost:3030/ 2>/dev/null && echo "" || echo "❌ API not responding"

# Show recent logs
echo ""
echo "=== Recent Logs (last 10 lines) ==="
sudo journalctl -u smart-forms -n 10 --no-pager

echo ""
echo "✅ Restart complete!"
echo ""
echo "View live logs: sudo journalctl -u smart-forms -f"
