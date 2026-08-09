#!/bin/bash

# AdNet VPS Deployment Script
# Usage: ./deploy.sh [staging|production]
# Run from Windows PowerShell (with WSL/Git Bash) or Linux terminal

set -e

VPS_IP="167.233.171.202"
VPS_USER="root"
APP_DIR="/opt/adnet"

# Anchor paths on THIS script's location, never on the caller's cwd.
# (When launched via a --login shell the cwd can be C:\Windows\System32,
#  which previously caused tar to archive the entire Windows directory.)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
# Use SSH key from WSL home directory with proper permissions
SSH_KEY="$HOME/.ssh/id_ed25519_linux"

ENV=${1:-staging}

echo "🚀 Deploying AdNet to VPS ($ENV environment)..."
echo "📁 Project dir: $PROJECT_DIR"

# Safety guard: make sure we're in the real project, not a system dir.
if [ ! -f "$PROJECT_DIR/go.mod" ]; then
  echo "❌ Refusing to deploy: $PROJECT_DIR does not look like the adnet project (no go.mod)."
  exit 1
fi

# Create tarball of project files
echo "📦 Creating project tarball..."
cd "$PROJECT_DIR"
tar -czf /tmp/adnet-deploy.tar.gz \
  --exclude='node_modules' \
  --exclude='.git' \
  --exclude='cmd/*/adserve' \
  --exclude='cmd/*/api' \
  --exclude='cmd/*/eventd' \
  --exclude='cmd/*/reconciler' \
  --exclude='cmd/*/billingd' \
  --exclude='cmd/*/seed-redis' \
  --exclude='*.exe' \
  --exclude='*.tar.gz' \
  --exclude='./api' \
  --exclude='./api-linux' \
  --exclude='**/dist' \
  --exclude='**/.output' \
  --exclude='**/.nitro' \
  --exclude='**/.vinxi' \
  .

echo "📦 Tarball size:"
du -h /tmp/adnet-deploy.tar.gz

# Copy tarball to VPS
echo "📤 Uploading tarball to VPS..."
scp -o StrictHostKeyChecking=no -i "$SSH_KEY" /tmp/adnet-deploy.tar.gz $VPS_USER@$VPS_IP:/tmp/

# Copy environment file into the directory Docker Compose reads it from
# (compose file lives in $APP_DIR/deploy, so .env must sit beside it).
echo "🔧 Copying environment file..."
ssh -o StrictHostKeyChecking=no -i "$SSH_KEY" $VPS_USER@$VPS_IP "mkdir -p $APP_DIR/deploy"
scp -o StrictHostKeyChecking=no -i "$SSH_KEY" deploy/.env.$ENV $VPS_USER@$VPS_IP:$APP_DIR/deploy/.env

# SSH into VPS and run setup
echo "🔨 Setting up VPS..."
ssh -o StrictHostKeyChecking=no -i "$SSH_KEY" $VPS_USER@$VPS_IP << 'ENDSSH'
# Create app directory
mkdir -p /opt/adnet

# Extract tarball
cd /opt/adnet
tar -xzf /tmp/adnet-deploy.tar.gz
rm /tmp/adnet-deploy.tar.gz

# Copy logo to web root
mkdir -p /var/www/otexads.com
cp /opt/adnet/deploy/otexlogo.png /var/www/otexads.com/otexlogo.png

# All compose operations run from the deploy dir (where docker-compose.yml lives).
cd /opt/adnet/deploy

# Use Docker Compose v2 (the v1 `docker-compose` binary is not installed).
DC="docker compose"

# Build and (re)start services
$DC build
$DC up -d

# Wait for services to be healthy
echo "⏳ Waiting for services to be healthy..."
sleep 10

# Check service status
$DC ps

# Run database migrations
echo "🗄️ Running database migrations..."
$DC exec -T postgres psql -U postgres -d adnet -f /docker-entrypoint-initdb.d/000001_init.up.sql

echo "✅ Deployment complete!"
ENDSSH

# Clean up local tarball
rm /tmp/adnet-deploy.tar.gz

echo "🎉 AdNet deployed successfully to $ENV!"
echo "🌐 API: http://167.233.171.202:8080"
echo "📊 Adserve: http://167.233.171.202:8081"
echo "💰 Billing callbacks: http://167.233.171.202:8083"
