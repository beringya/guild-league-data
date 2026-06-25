#!/usr/bin/env bash
set -Eeuo pipefail

APP_NAME="nsh-guild-analytics"
INSTALL_DIR="${INSTALL_DIR:-/opt/${APP_NAME}}"
APP_PORT="${APP_PORT:-18080}"
COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-${APP_NAME}}"
USE_POSTGRES="${USE_POSTGRES:-false}"
REPO_URL="${REPO_URL:-}"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "缺少命令：$1"
    exit 1
  }
}

detect_compose() {
  if docker compose version >/dev/null 2>&1; then
    echo "docker compose"
  elif command -v docker-compose >/dev/null 2>&1; then
    echo "docker-compose"
  else
    echo "缺少 Docker Compose，请先安装 Docker Compose v2。"
    exit 1
  fi
}

random_secret() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 32
  else
    tr -dc 'A-Za-z0-9' </dev/urandom | head -c 64
  fi
}

if [ "$(id -u)" -ne 0 ]; then
  echo "请使用 root 执行安装脚本。"
  exit 1
fi

need_cmd docker
COMPOSE_CMD="$(detect_compose)"

mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

if [ -n "$REPO_URL" ]; then
  need_cmd git
  if [ ! -d ".git" ]; then
    git clone "$REPO_URL" .
  else
    git pull --ff-only
  fi
else
  echo "使用当前目录中的安装包文件。"
fi

mkdir -p data/uploads backups
if [ ! -f ".env" ]; then
  cp .env.example .env
  sed -i "s|SESSION_SECRET=.*|SESSION_SECRET=$(random_secret)|" .env
  sed -i "s|POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD=$(random_secret)|" .env
fi

sed -i "s|APP_PORT=.*|APP_PORT=${APP_PORT}|" .env
sed -i "s|COMPOSE_PROJECT_NAME=.*|COMPOSE_PROJECT_NAME=${COMPOSE_PROJECT_NAME}|" .env

if [ "$USE_POSTGRES" = "true" ]; then
  $COMPOSE_CMD -p "$COMPOSE_PROJECT_NAME" -f docker-compose.yml -f docker-compose.postgres.yml up -d --build
else
  $COMPOSE_CMD -p "$COMPOSE_PROJECT_NAME" up -d --build
fi

cat >/etc/systemd/system/${APP_NAME}.service <<EOF
[Unit]
Description=NSH Guild Analytics Docker Compose Service
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=${INSTALL_DIR}
ExecStart=/usr/bin/docker compose -p ${COMPOSE_PROJECT_NAME} up -d
ExecStop=/usr/bin/docker compose -p ${COMPOSE_PROJECT_NAME} down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload || true
systemctl enable ${APP_NAME}.service || true

echo
echo "安装完成。"
echo "访问地址：http://服务器IP:${APP_PORT}"
echo "首次管理员账号：admin"
echo "请查看首次启动日志获取随机密码："
echo "  cd ${INSTALL_DIR} && ${COMPOSE_CMD} -p ${COMPOSE_PROJECT_NAME} logs app"
