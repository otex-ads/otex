# VPS Setup Guide for AdNet

## Prerequisites
- VPS with Ubuntu 22.04 LTS or similar
- SSH access with your ed25519 key
- Domain name pointed to VPS IP (167.233.171.202)

## Initial VPS Setup

### 1. Update System
```bash
ssh root@167.233.171.202
apt update && apt upgrade -y
```

### 2. Install Docker
```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
usermod -aG docker $USER
```

### 3. Install Docker Compose
```bash
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
```

### 4. Install Nginx
```bash
apt install nginx certbot python3-certbot-nginx -y
```

### 5. Create Application Directory
```bash
mkdir -p /opt/adnet
cd /opt/adnet
```

### 6. Configure Firewall
```bash
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable
```

## SSL Certificate Setup

### 1. Generate SSL Certificate
```bash
certbot --nginx -d api.yourdomain.com -d adserve.yourdomain.com
```

### 2. Auto-renewal (already configured by certbot)
```bash
certbot renew --dry-run
```

## Environment Configuration

Create `.env` file in `/opt/adnet`:
```bash
nano /opt/adnet/.env
```

Add the following:
```env
# Database
POSTGRES_URL=postgres://postgres:your_secure_password@postgres:5432/adnet?sslmode=disable

# Redis
REDIS_URL=redis://redis:6379

# JWT
JWT_SECRET=your_very_secure_random_secret_key_here

# M-Pesa (Sandbox for testing)
MPESA_CONSUMER_KEY=your_consumer_key
MPESA_CONSUMER_SECRET=your_consumer_secret
MPESA_ENVIRONMENT=sandbox
MPESA_SHORTCODE=174379
MPESA_PASSKEY=your_passkey
MPESA_CALLBACK_URL=https://api.yourdomain.com/callback

# Ports
API_PORT=8080
ADSERVE_PORT=8081
BILLINGD_PORT=8083
```

## Nginx Configuration

### API Server
```nginx
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/api.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Adserve Server
```nginx
server {
    listen 443 ssl http2;
    server_name adserve.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/adserve.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/adserve.yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Deploy

From your local machine:
```bash
cd deploy
chmod +x deploy.sh
./deploy.sh staging
```

## Monitoring

### Check Service Status
```bash
ssh root@167.233.171.202
cd /opt/adnet
docker-compose ps
docker-compose logs -f
```

### View Logs
```bash
docker-compose logs api
docker-compose logs adserve
docker-compose logs eventd
```

### Restart Services
```bash
docker-compose restart
```

## Database Backups

### Setup Automated Backups
```bash
# Create backup script
cat > /opt/adnet/backup.sh << 'EOF'
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
docker-compose exec -T postgres pg_dump -U postgres adnet > /backups/adnet_$DATE.sql
find /backups -name "adnet_*.sql" -mtime +7 -delete
EOF

chmod +x /opt/adnet/backup.sh

# Create backup directory
mkdir -p /backups

# Add to crontab (daily at 2 AM)
crontab -e
# Add: 0 2 * * * /opt/adnet/backup.sh
```

## Security Checklist

- [ ] Change default PostgreSQL password
- [ ] Use strong JWT secret
- [ ] Enable SSL only (redirect HTTP to HTTPS)
- [ ] Configure fail2ban for SSH protection
- [ ] Regular security updates
- [ ] Monitor logs for suspicious activity
- [ ] Use M-Pesa production credentials (not sandbox) for live
