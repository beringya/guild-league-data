#!/usr/bin/env bash
set -Eeuo pipefail

APP_NAME="nsh-guild-analytics"
INSTALL_DIR="${INSTALL_DIR:-/opt/${APP_NAME}}"
DEPLOY_DIR_HOST="${DEPLOY_DIR_HOST:-${INSTALL_DIR}}"
APP_PORT="${APP_PORT:-18080}"
COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-${APP_NAME}}"
APP_IMAGE_REPOSITORY="${APP_IMAGE_REPOSITORY:-ghcr.io/beringya/guild-league-data}"
APP_VERSION="${APP_VERSION:-}"
APP_IMAGE="${APP_IMAGE:-}"
SOURCE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Missing command: $1"
    exit 1
  }
}

detect_compose() {
  if docker compose version >/dev/null 2>&1; then
    echo "docker compose"
  elif command -v docker-compose >/dev/null 2>&1; then
    echo "docker-compose"
  else
    echo "Docker Compose v2 is required."
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
  echo "Please run this installer as root."
  exit 1
fi

need_cmd docker
COMPOSE_CMD="$(detect_compose)"

mkdir -p "$INSTALL_DIR"

if [ "$SOURCE_DIR" != "$INSTALL_DIR" ]; then
  need_cmd tar
  tar -C "$SOURCE_DIR" \
    --exclude='./data' \
    --exclude='./backups' \
    --exclude='./release' \
    -cf - . | tar -C "$INSTALL_DIR" -xf -
fi

cd "$INSTALL_DIR"
mkdir -p data/uploads backups
if [ ! -f ".env" ]; then
  cp .env.example .env
fi

if [ -z "$APP_VERSION" ]; then
  APP_VERSION="$(grep '^APP_VERSION=' .env | cut -d= -f2- || true)"
fi
APP_VERSION="${APP_VERSION:-latest}"
if [ -z "$APP_IMAGE" ]; then
  APP_IMAGE="${APP_IMAGE_REPOSITORY}:${APP_VERSION}"
fi

POSTGRES_USER_VALUE="$(grep '^POSTGRES_USER=' .env | cut -d= -f2- || true)"
POSTGRES_USER_VALUE="${POSTGRES_USER_VALUE:-nsh_guild}"
POSTGRES_DB_VALUE="$(grep '^POSTGRES_DB=' .env | cut -d= -f2- || true)"
POSTGRES_DB_VALUE="${POSTGRES_DB_VALUE:-nsh_guild_analytics}"
POSTGRES_PASSWORD_VALUE="$(grep '^POSTGRES_PASSWORD=' .env | cut -d= -f2- || true)"
POSTGRES_PASSWORD_VALUE="${POSTGRES_PASSWORD:-${POSTGRES_PASSWORD_VALUE:-$(random_secret)}}"
SESSION_SECRET_VALUE="$(grep '^SESSION_SECRET=' .env | cut -d= -f2- || true)"
SESSION_SECRET_VALUE="${SESSION_SECRET:-${SESSION_SECRET_VALUE:-$(random_secret)}}"

replace_env APP_VERSION "$APP_VERSION"
replace_env APP_IMAGE "$APP_IMAGE"
replace_env APP_IMAGE_REPOSITORY "$APP_IMAGE_REPOSITORY"
replace_env APP_PORT "$APP_PORT"
replace_env COMPOSE_PROJECT_NAME "$COMPOSE_PROJECT_NAME"
replace_env DEPLOY_DIR_HOST "$DEPLOY_DIR_HOST"
replace_env UPDATE_GITHUB_REPO "beringya/guild-league-data"
replace_env UPDATE_INSTALL_COMMAND "cd ${INSTALL_DIR} && docker compose pull app && docker compose up -d --no-build app"
replace_env UPDATE_APPLY_ENABLED "true"
replace_env UPDATE_APPLY_COMMAND "/app/bin/update-image.sh"
replace_env POSTGRES_USER "$POSTGRES_USER_VALUE"
replace_env POSTGRES_DB "$POSTGRES_DB_VALUE"
replace_env POSTGRES_PASSWORD "$POSTGRES_PASSWORD_VALUE"
replace_env SESSION_SECRET "$SESSION_SECRET_VALUE"
replace_env DATABASE_DSN "postgres://${POSTGRES_USER_VALUE}:${POSTGRES_PASSWORD_VALUE}@postgres:5432/${POSTGRES_DB_VALUE}?sslmode=disable"

$COMPOSE_CMD -p "$COMPOSE_PROJECT_NAME" pull app
$COMPOSE_CMD -p "$COMPOSE_PROJECT_NAME" up -d --no-build

cat >/etc/systemd/system/${APP_NAME}.service <<EOF
[Unit]
Description=NSH Guild Analytics Docker Compose Service
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=${INSTALL_DIR}
ExecStart=/usr/bin/docker compose -p ${COMPOSE_PROJECT_NAME} up -d --no-build
ExecStop=/usr/bin/docker compose -p ${COMPOSE_PROJECT_NAME} down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload || true
systemctl enable ${APP_NAME}.service || true

echo
echo "Install complete."
echo "Image: ${APP_IMAGE}"
echo "URL: http://SERVER_IP:${APP_PORT}"
echo "Default admin account: admin"
echo "Check first-start logs for the generated password:"
echo "  cd ${INSTALL_DIR} && ${COMPOSE_CMD} -p ${COMPOSE_PROJECT_NAME} logs app"
