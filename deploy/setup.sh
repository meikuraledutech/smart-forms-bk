#!/bin/bash
set -e

echo "=== Smart Forms Deployment Setup ==="

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
APP_DIR="$(dirname "$SCRIPT_DIR")"

echo "App directory: $APP_DIR"

# 1. Install dependencies
echo "1. Installing Go and nginx..."
sudo dnf update -y
sudo dnf install -y golang nginx git

# 2. Check if .env exists
if [ ! -f "$APP_DIR/.env" ]; then
    echo "2. Creating .env file..."
    if [ -f "$APP_DIR/.env.example" ]; then
        cp "$APP_DIR/.env.example" "$APP_DIR/.env"
        echo "⚠️  Please edit .env with your actual credentials!"
    else
        echo "❌ .env.example not found!"
        exit 1
    fi
else
    echo "2. .env file already exists ✓"
fi

# 3. Build binary (optimized for AWS Lightsail)
echo "3. Building optimized Go binary for Linux amd64..."
cd "$APP_DIR"

# Build with maximum optimization for AWS Lightsail (Amazon Linux 2023, x86_64)
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
echo "✓ Binary built: $BINARY_SIZE (optimized for AWS x86_64)"

# 4. Create systemd service
echo "4. Creating systemd service..."
sudo cp "$SCRIPT_DIR/smart-forms.service" /etc/systemd/system/smart-forms.service
sudo sed -i "s|/home/ec2-user/app|$APP_DIR|g" /etc/systemd/system/smart-forms.service

# 5. Configure nginx
echo "5. Configuring nginx..."
sudo cp "$SCRIPT_DIR/nginx.conf" /etc/nginx/conf.d/smart-forms.conf

# 6. Start services
echo "6. Starting services..."
sudo systemctl daemon-reload
sudo systemctl enable smart-forms
sudo systemctl restart smart-forms
sudo systemctl enable nginx
sudo systemctl restart nginx

# 7. Verify
echo ""
echo "=== Verification ==="
sleep 2

echo "App status:"
sudo systemctl is-active smart-forms && echo "✓ Running" || echo "❌ Not running"

echo "Nginx status:"
sudo systemctl is-active nginx && echo "✓ Running" || echo "❌ Not running"

echo ""
echo "Testing API:"
curl -s http://localhost:3030/ 2>/dev/null || echo "API not responding"

echo ""
echo "✅ Deployment complete!"
echo ""
echo "Useful commands:"
echo "  Check status:  sudo systemctl status smart-forms"
echo "  View logs:     sudo journalctl -u smart-forms -f"
echo "  Restart:       sudo systemctl restart smart-forms"
echo "  Update code:   cd $APP_DIR && git pull && ./deploy/setup.sh"
