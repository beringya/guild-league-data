#!/bin/sh
set -eu

mkdir -p /app/data/uploads /app/backups
exec "$@"
