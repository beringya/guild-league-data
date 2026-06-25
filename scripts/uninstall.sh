#!/usr/bin/env bash
set -Eeuo pipefail

APP_NAME="nsh-guild-analytics"
INSTALL_DIR="${INSTALL_DIR:-/opt/${APP_NAME}}"
COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-${APP_NAME}}"
REMOVE_DATA="${REMOVE_DATA:-false}"

if [ "$(id -u)" -ne 0 ]; then
  echo "请使用 root 执行卸载脚本。"
  exit 1
fi

systemctl disable --now ${APP_NAME}.service >/dev/null 2>&1 || true
rm -f /etc/systemd/system/${APP_NAME}.service
systemctl daemon-reload || true

if [ -d "$INSTALL_DIR" ]; then
  cd "$INSTALL_DIR"
  if docker compose version >/dev/null 2>&1; then
    docker compose -p "$COMPOSE_PROJECT_NAME" down
  elif command -v docker-compose >/dev/null 2>&1; then
    docker-compose -p "$COMPOSE_PROJECT_NAME" down
  fi
fi

if [ "$REMOVE_DATA" = "true" ]; then
  rm -rf "$INSTALL_DIR"
  echo "已删除安装目录和数据：$INSTALL_DIR"
else
  echo "已停止服务，保留数据目录：$INSTALL_DIR"
fi
