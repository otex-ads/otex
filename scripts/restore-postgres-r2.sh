#!/bin/bash

# PostgreSQL Restore Script from Cloudflare R2
# Downloads and restores a backup from R2 bucket

set -e

# Configuration
BACKUP_DIR="/tmp/postgres-backups"
BUCKET_NAME="otex"
DATABASE_NAME="adnet"
POSTGRES_USER="postgres"
POSTGRES_HOST="adnet-postgres"
POSTGRES_PORT="5432"

# Check if backup filename is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <backup-filename>"
    echo "Available backups:"
    aws s3 ls "s3://${BUCKET_NAME}/" | grep "adnet-backup-"
    exit 1
fi

BACKUP_FILENAME="$1"
BACKUP_PATH="${BACKUP_DIR}/${BACKUP_FILENAME}"

# Create backup directory
mkdir -p "$BACKUP_DIR"

echo "[$(date)] Downloading backup: ${BACKUP_FILENAME}"
aws s3 cp "s3://${BUCKET_NAME}/${BACKUP_FILENAME}" "${BACKUP_PATH}"

echo "[$(date)] Restoring database: ${DATABASE_NAME}"
# Drop existing database and recreate
docker exec "${POSTGRES_HOST}" psql -U "${POSTGRES_USER}" -c "DROP DATABASE IF EXISTS ${DATABASE_NAME};"
docker exec "${POSTGRES_HOST}" psql -U "${POSTGRES_USER}" -c "CREATE DATABASE ${DATABASE_NAME};"

# Restore from backup
gunzip -c "${BACKUP_PATH}" | docker exec -i "${POSTGRES_HOST}" psql -U "${POSTGRES_USER}" -d "${DATABASE_NAME}"

echo "[$(date)] Restore completed successfully"

# Clean up local backup file
rm -f "${BACKUP_PATH}"
echo "[$(date)] Local backup file removed"
