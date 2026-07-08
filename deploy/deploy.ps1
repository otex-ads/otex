# AdNet VPS Deployment Script (PowerShell version)
# Usage: .\deploy.ps1 [staging|production]

param(
    [string]$ENV = "staging"
)

$VPS_IP = "167.233.171.202"
$VPS_USER = "root"
$APP_DIR = "/opt/adnet"
$PROJECT_DIR = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)

Write-Host "Deploying AdNet to VPS ($ENV environment)..." -ForegroundColor Green

# Create tarball of project files
Write-Host "Creating project tarball..." -ForegroundColor Yellow
$tarballPath = "$env:TEMP\adnet-deploy.tar.gz"
Push-Location $PROJECT_DIR
tar -czf $tarballPath --exclude='node_modules' --exclude='.git' --exclude='cmd/*/adserve' --exclude='cmd/*/api' --exclude='cmd/*/eventd' --exclude='cmd/*/reconciler' --exclude='cmd/*/billingd' --exclude='cmd/*/seed-redis' --exclude='deploy/*.exe' .
Pop-Location

# Copy tarball to VPS
Write-Host "Uploading tarball to VPS..." -ForegroundColor Yellow
scp -o StrictHostKeyChecking=no $tarballPath "${VPS_USER}@${VPS_IP}:/tmp/"

# Copy environment file
Write-Host "Copying environment file..." -ForegroundColor Yellow
scp -o StrictHostKeyChecking=no "$PSScriptRoot\.env.$ENV" "${VPS_USER}@${VPS_IP}:${APP_DIR}/.env"

# SSH into VPS and run setup
Write-Host "Setting up VPS..." -ForegroundColor Yellow
ssh -o StrictHostKeyChecking=no ${VPS_USER}@${VPS_IP} @"
# Create app directory
mkdir -p $APP_DIR

# Extract tarball
cd $APP_DIR
tar -xzf /tmp/adnet-deploy.tar.gz
rm /tmp/adnet-deploy.tar.gz

# Stop existing services
docker-compose down || true

# Build and start services
docker-compose build
docker-compose up -d

# Wait for services to be healthy
echo "Waiting for services to be healthy..."
sleep 10

# Check service status
docker-compose ps

# Run database migrations
echo "Running database migrations..."
docker-compose exec -T postgres psql -U postgres -d adnet -f /docker-entrypoint-initdb.d/000001_init.up.sql

echo "Deployment complete!"
"@

# Clean up local tarball
Remove-Item $tarballPath -Force

Write-Host "AdNet deployed successfully to $ENV!" -ForegroundColor Green
Write-Host "API: http://167.233.171.202:8080" -ForegroundColor Cyan
Write-Host "Adserve: http://167.233.171.202:8081" -ForegroundColor Cyan
Write-Host "Billing callbacks: http://167.233.171.202:8083" -ForegroundColor Cyan
