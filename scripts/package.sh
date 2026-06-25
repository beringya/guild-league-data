#!/usr/bin/env bash
set -Eeuo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
if [ -z "${VERSION:-}" ] && [ -f "${ROOT}/.env.example" ]; then
  VERSION="$(grep '^APP_VERSION=' "${ROOT}/.env.example" | cut -d= -f2- || true)"
fi
VERSION="${VERSION:-$(date +%Y%m%d%H%M%S)}"
NAME="nsh-guild-analytics-${VERSION}"
OUT_DIR="${ROOT}/release"
PKG_DIR="${OUT_DIR}/${NAME}"

rm -rf "$PKG_DIR"
mkdir -p "$PKG_DIR"

copy_item() {
  cp -a "$ROOT/$1" "$PKG_DIR/$1"
}

for item in deployment scripts docker-compose.yml .env.example README.md LICENSE; do
  [ -e "$ROOT/$item" ] && copy_item "$item"
done

mkdir -p "$PKG_DIR/data/uploads" "$PKG_DIR/backups"
touch "$PKG_DIR/data/uploads/.gitkeep" "$PKG_DIR/backups/.gitkeep"

if grep -q '^APP_VERSION=' "$PKG_DIR/.env.example"; then
  sed -i "s|^APP_VERSION=.*|APP_VERSION=${VERSION}|" "$PKG_DIR/.env.example"
else
  printf 'APP_VERSION=%s\n' "$VERSION" >> "$PKG_DIR/.env.example"
fi
if grep -q '^APP_IMAGE=' "$PKG_DIR/.env.example"; then
  sed -i "s|^APP_IMAGE=.*|APP_IMAGE=ghcr.io/beringya/guild-league-data:${VERSION}|" "$PKG_DIR/.env.example"
else
  printf 'APP_IMAGE=ghcr.io/beringya/guild-league-data:%s\n' "$VERSION" >> "$PKG_DIR/.env.example"
fi
if grep -q '^APP_IMAGE_REPOSITORY=' "$PKG_DIR/.env.example"; then
  sed -i "s|^APP_IMAGE_REPOSITORY=.*|APP_IMAGE_REPOSITORY=ghcr.io/beringya/guild-league-data|" "$PKG_DIR/.env.example"
else
  printf 'APP_IMAGE_REPOSITORY=ghcr.io/beringya/guild-league-data\n' >> "$PKG_DIR/.env.example"
fi

cd "$OUT_DIR"
tar -czf "${NAME}.tar.gz" "$NAME"
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum "${NAME}.tar.gz" > "${NAME}.tar.gz.sha256"
fi

echo "Packaged ${OUT_DIR}/${NAME}.tar.gz"
