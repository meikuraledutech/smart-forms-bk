# Smart Forms - Deployment Guide

## Quick Deploy on AWS Lightsail

### 1. Create Lightsail Instance
- Go to AWS Lightsail Console
- Create instance with **Amazon Linux 2023**
- Choose **micro_3_1** or **nano_ipv6_3_1** bundle
- Open port **80** (HTTP)

### 2. SSH into Server
```bash
# Via AWS Console
Click "Connect using SSH" in Lightsail dashboard
```

### 3. Clone Repository
```bash
cd ~
git clone https://github.com/meikuraledutech/smart-forms-bk.git app
cd app
```

### 4. Configure Environment
```bash
# Copy example env file
cp .env.example .env

# Edit with your actual credentials
nano .env
```

**Required variables:**
- `DATABASE_URL` - Your PostgreSQL connection string
- `CORS_ORIGINS` - Your frontend URL(s)
- `ACCESS_TOKEN_SECRET` - Random secure string
- `REFRESH_TOKEN_SECRET` - Different random secure string

### 5. Run Deployment Script
```bash
chmod +x deploy/setup.sh
./deploy/setup.sh
```

This will:
- ✅ Install Go and nginx
- ✅ Build the binary
- ✅ Set up systemd service (auto-restart on crash)
- ✅ Configure nginx reverse proxy
- ✅ Start all services

### 6. Verify Deployment
```bash
# Check services are running
sudo systemctl status smart-forms
sudo systemctl status nginx

# Test API
curl http://localhost:3030/
```

### 7. Configure Cloudflare DNS
- Type: `AAAA`
- Name: `api` (or your subdomain)
- Content: Your server's IPv6 address
- Proxy: **ON** (orange cloud)
- SSL/TLS: **Flexible** mode

---

## Useful Commands

### Service Management
```bash
# Check status
sudo systemctl status smart-forms

# Restart service
sudo systemctl restart smart-forms

# Stop service
sudo systemctl stop smart-forms

# View logs (real-time)
sudo journalctl -u smart-forms -f

# View logs (last 100 lines)
sudo journalctl -u smart-forms -n 100
```

### Update Code
```bash
cd ~/app
git pull
./deploy/setup.sh
```

### Manual Build
```bash
cd ~/app
CGO_ENABLED=0 go build -ldflags="-s -w" -o smart-forms-backend
sudo systemctl restart smart-forms
```

### Nginx Management
```bash
# Test nginx config
sudo nginx -t

# Restart nginx
sudo systemctl restart nginx

# View nginx logs
sudo tail -f /var/log/nginx/error.log
```

---

## Architecture

```
Internet (IPv4/IPv6)
    ↓
Cloudflare (DNS + Proxy + SSL)
    ↓ (IPv6)
Nginx :80
    ↓
Smart Forms App :3030
    ↓
PostgreSQL Database
```

## Features

- ✅ **Auto-restart on crash** - Systemd restarts app if it fails
- ✅ **Zero-downtime updates** - Systemd handles graceful restarts
- ✅ **Logging** - All logs in systemd journal
- ✅ **Nginx reverse proxy** - Handles HTTP, WebSocket support
- ✅ **Cloudflare SSL** - Free HTTPS
- ✅ **Cost-effective** - $5/month (micro) or $3.50/month (nano IPv6)

## Troubleshooting

### App not starting?
```bash
# Check logs
sudo journalctl -u smart-forms -n 50

# Common issues:
# - .env file missing or incorrect
# - Database connection failed
# - Port 3030 already in use
```

### Nginx not working?
```bash
# Test config
sudo nginx -t

# Check if port 80 is open
sudo netstat -tlnp | grep :80

# Restart nginx
sudo systemctl restart nginx
```

### Can't connect from Cloudflare?
```bash
# Verify port 80 is open in Lightsail firewall
# Check IPv6 address in Cloudflare DNS matches server
ip -6 addr show
```

---

## Security Notes

1. **Never commit .env file** - Contains sensitive credentials
2. **Use strong JWT secrets** - Random 32+ character strings
3. **Keep system updated** - `sudo dnf update -y`
4. **Use Cloudflare proxy** - Hides origin IP
5. **Review logs regularly** - `sudo journalctl -u smart-forms`
