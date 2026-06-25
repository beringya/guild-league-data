#!/bin/sh
set -eu

APP_NAME="${APP_NAME:-nsh-guild-analytics}"
INSTALL_DIR="${DEPLOY_DIR_HOST:-/opt/${APP_NAME}}"
COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-${APP_NAME}}"
APP_IMAGE_REPOSITORY="${APP_IMAGE_REPOSITORY:-ghcr.io/beringya/guild-league-data}"
APP_IMAGE="${APP_IMAGE:-${APP_IMAGE_REPOSITORY}:latest}"
LATEST_VERSION="${UPDATE_LATEST_VERSION:-}"
UPDATE_ACTION="${UPDATE_ACTION:-restart}"

if [ -z "$LATEST_VERSION" ]; then
  echo "missing UPDATE_LATEST_VERSION" >&2
  exit 2
fi

TAG="${LATEST_VERSION#v}"
IMAGE="${APP_IMAGE_REPOSITORY}:${TAG}"
PENDING_FILE="${INSTALL_DIR}/.pending-update"

if [ "$UPDATE_ACTION" = "download" ]; then
  docker pull "$IMAGE"
  if [ -d "$INSTALL_DIR" ]; then
    printf '%s\n' "$TAG" > "$PENDING_FILE"
  fi
  echo "downloaded ${IMAGE}"
  exit 0
fi

if [ "${UPDATE_HELPER:-}" != "1" ]; then
  helper_name="${COMPOSE_PROJECT_NAME}_updater_$(date +%s)"
  docker run -d --rm \
    --name "$helper_name" \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v "${INSTALL_DIR}:${INSTALL_DIR}" \
    -e UPDATE_HELPER=1 \
    -e UPDATE_ACTION=restart \
    -e UPDATE_LATEST_VERSION="$LATEST_VERSION" \
    -e APP_NAME="$APP_NAME" \
    -e DEPLOY_DIR_HOST="$INSTALL_DIR" \
    -e COMPOSE_PROJECT_NAME="$COMPOSE_PROJECT_NAME" \
    -e APP_IMAGE_REPOSITORY="$APP_IMAGE_REPOSITORY" \
    --entrypoint /app/bin/update-image.sh \
    "$APP_IMAGE"
  echo "started update helper $helper_name"
  exit 0
fi

if [ ! -d "$INSTALL_DIR" ]; then
  echo "install directory not found: $INSTALL_DIR" >&2
  exit 3
fi

cd "$INSTALL_DIR"

if [ ! -f .env ]; then
  echo ".env not found in $INSTALL_DIR" >&2
  exit 4
fi

backup_env="$(mktemp)"
tmp_env="$(mktemp)"
cleanup() {
  rm -f "$backup_env" "$tmp_env" "${tmp_env}.next"
}
trap cleanup EXIT

docker pull "$IMAGE"

cp .env "$backup_env"
if grep -q '^APP_VERSION=' .env; then
  sed "s|^APP_VERSION=.*|APP_VERSION=${TAG}|" .env > "$tmp_env"
else
  cat .env > "$tmp_env"
  printf 'APP_VERSION=%s\n' "$TAG" >> "$tmp_env"
fi
if grep -q '^APP_IMAGE=' "$tmp_env"; then
  sed "s|^APP_IMAGE=.*|APP_IMAGE=${IMAGE}|" "$tmp_env" > "${tmp_env}.next"
  mv "${tmp_env}.next" "$tmp_env"
else
  printf 'APP_IMAGE=%s\n' "$IMAGE" >> "$tmp_env"
fi
mv "$tmp_env" .env

if ! docker compose -p "$COMPOSE_PROJECT_NAME" up -d --no-build app; then
  cp "$backup_env" .env
  docker compose -p "$COMPOSE_PROJECT_NAME" up -d --no-build app || true
  exit 5
fi
rm -f "$PENDING_FILE"
docker image prune -f >/dev/null 2>&1 || true

echo "updated ${APP_NAME} to ${IMAGE}"
