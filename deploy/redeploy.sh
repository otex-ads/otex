#!/bin/bash
# Fast redeploy: git pull on VPS then rebuild only changed services
# Usage (from your machine):
#   bash deploy/redeploy.sh                      # rebuild everything
#   bash deploy/redeploy.sh advertiser-portal    # rebuild one service
#   bash deploy/redeploy.sh api advertiser-portal  # rebuild multiple

set -e

VPS_IP="167.233.171.202"
VPS_USER="root"
APP_DIR="/opt/adnet"
SERVICES="${@}"

echo "==> Pulling latest code on VPS..."
ssh -o StrictHostKeyChecking=no $VPS_USER@$VPS_IP "cd $APP_DIR && git pull origin main"

if [ -z "$SERVICES" ]; then
  echo "==> Rebuilding ALL services..."
  ssh -o StrictHostKeyChecking=no $VPS_USER@$VPS_IP "cd $APP_DIR && docker compose -f deploy/docker-compose.yml up --build -d"
else
  echo "==> Rebuilding: $SERVICES"
  ssh -o StrictHostKeyChecking=no $VPS_USER@$VPS_IP "cd $APP_DIR && docker compose -f deploy/docker-compose.yml up --build -d $SERVICES"
fi

echo "==> Done. Running containers:"
ssh -o StrictHostKeyChecking=no $VPS_USER@$VPS_IP "cd $APP_DIR && docker compose -f deploy/docker-compose.yml ps"
