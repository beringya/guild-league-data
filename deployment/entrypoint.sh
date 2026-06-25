#!/bin/sh
set -eu

mkdir -p "${UPLOAD_DIR:-/app/data/uploads}" "${BACKUP_DIR:-/app/backups}" "${STATIC_DIR:-/app/public}"
exec "$@"
