#!/bin/bash

# PostgreSQL Backup Script for Cloudflare R2
# Backs up the adnet database to R2 bucket with 14-day retention

set -e

# Configuration
BACKUP_DIR="/tmp/postgres-backups"
BUCKET_NAME="otex"
DATABASE_NAME="adnet"
POSTGRES_USER="postgres"
POSTGRES_HOST="adnet-postgres"
POSTGRES_PORT="5432"
RETENTION_DAYS=14

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Generate timestamp for backup filename
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
BACKUP_FILENAME="adnet-backup-${TIMESTAMP}.sql.gz"
BACKUP_PATH="${BACKUP_DIR}/${BACKUP_FILENAME}"

echo "[$(date)] Starting PostgreSQL backup for ${DATABASE_NAME}"

# Dump database and compress
echo "[$(date)] Running pg_dump..."
docker exec "${POSTGRES_HOST}" pg_dump -U "${POSTGRES_USER}" -d "${DATABASE_NAME}" | gzip > "${BACKUP_PATH}"

# Get backup size
BACKUP_SIZE=$(du -h "${BACKUP_PATH}" | cut -f1)
echo "[$(date)] Backup created: ${BACKUP_FILENAME} (${BACKUP_SIZE})"

# Upload to R2
echo "[$(date)] Uploading to R2 bucket: ${BUCKET_NAME}"
aws s3 cp "${BACKUP_PATH}" "s3://${BUCKET_NAME}/${BACKUP_FILENAME}"

# Verify upload
if aws s3 ls "s3://${BUCKET_NAME}/${BACKUP_FILENAME}" > /dev/null 2>&1; then
    echo "[$(date)] Upload successful"
else
    echo "[$(date)] ERROR: Upload failed"
    exit 1
fi

# Clean up local backup file
rm -f "${BACKUP_PATH}"
echo "[$(date)] Local backup file removed"

# Clean up old backups from R2 (older than RETENTION_DAYS)
echo "[$(date)] Cleaning up backups older than ${RETENTION_DAYS} days"
CUTOFF_DATE=$(date -d "${RETENTION_DAYS} days ago" +%Y%m%d)

# List all backups in the bucket and delete old ones
aws s3 ls "s3://${BUCKET_NAME}/" | grep "adnet-backup-" | while read -r line; do
    FILE_DATE=$(echo "$line" | awk '{print $2}' | cut -d'-' -f3 | cut -d'.' -f1)
    FILE_NAME=$(echo "$line" | awk '{print $4}')
    
    # Extract date from filename (format: adnet-backup-YYYYMMDD-HHMMSS.sql.gz)
    FILE_TIMESTAMP=$(echo "$FILE_NAME" | sed 's/adnet-backup-\([0-9]\{8\}\)-.*/\1/')
    
    if [ "$FILE_TIMESTAMP" -lt "$CUTOFF_DATE" ]; then
        echo "[$(date)] Deleting old backup: ${FILE_NAME}"
        aws s3 rm "s3://${BUCKET_NAME}/${FILE_NAME}"
    fi
done

echo "[$(date)] Backup completed successfully"
