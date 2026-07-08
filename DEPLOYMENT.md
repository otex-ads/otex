# AdNet Deployment Guide

## Quick Start (Linux Deployment)

### 1. VPS Initial Setup
Follow `deploy/VPS_SETUP.md` to set up your VPS (167.233.171.202) with Docker, Docker Compose, and Nginx.

### 2. Configure Environment
Copy the environment template and fill in your values:
```bash
cp deploy/.env.example deploy/.env.staging
# Edit deploy/.env.staging with your actual values
```

### 3. Deploy from Linux Machine
The deployment script is designed to run from your Linux machine (penguin) where SSH key authentication works:
```bash
cd deploy
chmod +x deploy.sh
./deploy.sh staging
```

## Deployment Files

- `deploy/deploy.sh` - Automated deployment script
- `deploy/VPS_SETUP.md` - Complete VPS setup instructions
- `deploy/.env.example` - Environment variables template
- `deploy/nginx-api.conf` - Nginx config for API server
- `deploy/nginx-adserve.conf` - Nginx config for adserve server
- `deploy/docker-compose.yml` - Docker services configuration

## Services

After deployment, the following services will be running:

- **API**: https://api.yourdomain.com (port 8080)
- **Adserve**: https://adserve.yourdomain.com (port 8081)
- **Eventd**: Internal (event processing)
- **Reconciler**: Internal (budget sync)
- **Billingd**: Internal (M-Pesa callbacks on port 8083)
- **PostgreSQL**: Internal (port 5432)
- **Redis**: Internal (port 6379)

## Monitoring

### Check Service Status
```bash
ssh root@167.233.171.202
cd /opt/adnet
docker-compose ps
```

### View Logs
```bash
docker-compose logs -f api
docker-compose logs -f adserve
docker-compose logs -f eventd
```

### Restart Services
```bash
docker-compose restart
```

## Database Backups

Automated backups are configured via cron. Manual backup:
```bash
docker-compose exec postgres pg_dump -U postgres adnet > backup.sql
```

## SSL Certificates

SSL certificates are managed by Let's Encrypt via certbot. Auto-renewal is configured.

Renew manually:
```bash
certbot renew
```

## Troubleshooting

### Services won't start
```bash
docker-compose logs
docker-compose down
docker-compose up -d
```

### Database connection issues
```bash
docker-compose exec postgres psql -U postgres -d adnet
```

### Redis connection issues
```bash
docker-compose exec redis redis-cli ping
```

## Security

- Change default passwords in `.env`
- Use strong JWT secret
- Enable firewall (ufw)
- Regular system updates
- Monitor logs for suspicious activity

## Scaling

For high traffic:
- Deploy adserve on separate instances
- Use load balancer (Nginx)
- Add Redis cluster
- Add PostgreSQL read replica
