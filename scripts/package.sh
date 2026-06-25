#!/usr/bin/env bash
set -Eeuo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${VERSION:-$(date +%Y%m%d%H%M%S)}"
NAME="nsh-guild-analytics-${VERSION}"
OUT_DIR="${ROOT}/release"
PKG_DIR="${OUT_DIR}/${NAME}"

rm -rf "$PKG_DIR"
mkdir -p "$PKG_DIR"

copy_item() {
  cp -a "$ROOT/$1" "$PKG_DIR/$1"
}

for item in backend frontend deployment scripts Dockerfile docker-compose.yml docker-compose.postgres.yml .env.example README.md DEVELOPMENT_PLAN.md IMPLEMENTATION_CHECKLIST.md LICENSE CONTRIBUTING.md .dockerignore .gitignore .gitee; do
  [ -e "$ROOT/$item" ] && copy_item "$item"
done

mkdir -p "$PKG_DIR/data/uploads" "$PKG_DIR/backups"
touch "$PKG_DIR/data/uploads/.gitkeep" "$PKG_DIR/backups/.gitkeep"

cd "$OUT_DIR"
tar -czf "${NAME}.tar.gz" "$NAME"
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum "${NAME}.tar.gz" > "${NAME}.tar.gz.sha256"
fi

echo "打包完成：${OUT_DIR}/${NAME}.tar.gz"
