#!/usr/bin/env bash
set -Eeuo pipefail

APP_NAME="nsh-guild-analytics"
INSTALL_DIR="${INSTALL_DIR:-/opt/${APP_NAME}}"
APP_PORT="${APP_PORT:-18080}"
COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-${APP_NAME}}"
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

replace_env() {
  local key="$1"
  local value="$2"
  if grep -q "^${key}=" .env; then
    sed -i "s|^${key}=.*|${key}=${value}|" .env
  else
    printf '%s=%s\n' "$key" "$value" >> .env
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
  echo "使用当前目录中的离线安装包文件。"
fi

mkdir -p data/uploads backups
if [ ! -f ".env" ]; then
  cp .env.example .env
fi

POSTGRES_USER_VALUE="$(grep '^POSTGRES_USER=' .env | cut -d= -f2- || true)"
POSTGRES_USER_VALUE="${POSTGRES_USER_VALUE:-nsh_guild}"
POSTGRES_DB_VALUE="$(grep '^POSTGRES_DB=' .env | cut -d= -f2- || true)"
POSTGRES_DB_VALUE="${POSTGRES_DB_VALUE:-nsh_guild_analytics}"
POSTGRES_PASSWORD_VALUE="${POSTGRES_PASSWORD:-$(random_secret)}"
SESSION_SECRET_VALUE="${SESSION_SECRET:-$(random_secret)}"

replace_env APP_PORT "$APP_PORT"
replace_env COMPOSE_PROJECT_NAME "$COMPOSE_PROJECT_NAME"
replace_env POSTGRES_USER "$POSTGRES_USER_VALUE"
replace_env POSTGRES_DB "$POSTGRES_DB_VALUE"
replace_env POSTGRES_PASSWORD "$POSTGRES_PASSWORD_VALUE"
replace_env SESSION_SECRET "$SESSION_SECRET_VALUE"
replace_env DATABASE_DSN "postgres://${POSTGRES_USER_VALUE}:${POSTGRES_PASSWORD_VALUE}@postgres:5432/${POSTGRES_DB_VALUE}?sslmode=disable"

$COMPOSE_CMD -p "$COMPOSE_PROJECT_NAME" up -d --build

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
