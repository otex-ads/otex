#!/bin/bash

# AdNet VPS Deployment Script
# Usage: ./deploy.sh [staging|production]
# Run from Windows PowerShell (with WSL/Git Bash) or Linux terminal

set -e

VPS_IP="167.233.171.202"
VPS_USER="root"
APP_DIR="/opt/adnet"
PROJECT_DIR="$(pwd)/.."

ENV=${1:-staging}

echo "🚀 Deploying AdNet to VPS ($ENV environment)..."

# Create tarball of project files
echo "📦 Creating project tarball..."
cd $PROJECT_DIR
tar -czf /tmp/adnet-deploy.tar.gz \
  --exclude='node_modules' \
  --exclude='.git' \
  --exclude='cmd/*/adserve' \
  --exclude='cmd/*/api' \
  --exclude='cmd/*/eventd' \
  --exclude='cmd/*/reconciler' \
  --exclude='cmd/*/billingd' \
  --exclude='cmd/*/seed-redis' \
  --exclude='deploy/*.exe' \
  .

# Copy tarball to VPS
echo "📤 Uploading tarball to VPS..."
scp -o StrictHostKeyChecking=no /tmp/adnet-deploy.tar.gz $VPS_USER@$VPS_IP:/tmp/

# Copy environment file
echo "🔧 Copying environment file..."
scp -o StrictHostKeyChecking=no deploy/.env.$ENV $VPS_USER@$VPS_IP:$APP_DIR/.env

# SSH into VPS and run setup
echo "🔨 Setting up VPS..."
ssh -o StrictHostKeyChecking=no $VPS_USER@$VPS_IP << 'ENDSSH'
# Create app directory
mkdir -p /opt/adnet

# Extract tarball
cd /opt/adnet
tar -xzf /tmp/adnet-deploy.tar.gz
rm /tmp/adnet-deploy.tar.gz

# Stop existing services
docker-compose down || true

# Build and start services
docker-compose build
docker-compose up -d

# Wait for services to be healthy
echo "⏳ Waiting for services to be healthy..."
sleep 10

# Check service status
docker-compose ps

# Run database migrations
echo "🗄️ Running database migrations..."
docker-compose exec -T postgres psql -U postgres -d adnet -f /docker-entrypoint-initdb.d/000001_init.up.sql

echo "✅ Deployment complete!"
ENDSSH

# Clean up local tarball
rm /tmp/adnet-deploy.tar.gz

echo "🎉 AdNet deployed successfully to $ENV!"
echo "🌐 API: http://167.233.171.202:8080"
echo "📊 Adserve: http://167.233.171.202:8081"
echo "💰 Billing callbacks: http://167.233.171.202:8083"
