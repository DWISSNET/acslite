#!/usr/bin/env bash
set -euo pipefail

# Go-first ACS installer for DWISSNET/acslite
# Default database is SQLite. To use PostgreSQL, set at install time:
#   DB_TYPE=postgres
#   DB_URL='******127.0.0.1:5432/acsgo?sslmode=disable'
# Optional env overrides accepted by installer: PORT, CWMP_PORT, ADMIN_EMAIL,
# ADMIN_PASSWORD, JWT_SECRET, DB_TYPE, DB_URL.

REPO_URL="${REPO_URL:-https://github.com/DWISSNET/acslite.git}"
INSTALL_DIR="${INSTALL_DIR:-/opt/acsgo/src}"
BINARY_PATH="${BINARY_PATH:-/usr/local/bin/acsgo}"
ENV_FILE="${ENV_FILE:-/etc/acsgo/acsgo.env}"
SERVICE_FILE="/etc/systemd/system/acsgo.service"
RUN_USER="acsgo"
RUN_GROUP="acsgo"

log() { printf '[acsgo-installer] %s\n' "$*"; }
fail() { printf '[acsgo-installer] ERROR: %s\n' "$*" >&2; exit 1; }

require_root() {
  if [[ ${EUID:-$(id -u)} -ne 0 ]]; then
    fail "Run as root (or use sudo)."
  fi
}

install_deps() {
  export DEBIAN_FRONTEND=noninteractive
  log "Installing required system dependencies..."
  apt-get update -y
  apt-get install -y --no-install-recommends \
    ca-certificates curl git build-essential pkg-config sqlite3

  if ! command -v go >/dev/null 2>&1; then
    log "Go not found, installing golang-go from apt..."
    apt-get install -y --no-install-recommends golang-go
  fi
}

sync_repo() {
  log "Syncing source into ${INSTALL_DIR}..."
  mkdir -p "$(dirname "$INSTALL_DIR")"

  if [[ -d "$INSTALL_DIR/.git" ]]; then
    git -C "$INSTALL_DIR" fetch --tags origin
    git -C "$INSTALL_DIR" reset --hard origin/main
    git -C "$INSTALL_DIR" clean -fd
  else
    rm -rf "$INSTALL_DIR"
    git clone "$REPO_URL" "$INSTALL_DIR"
  fi
}

build_binary() {
  log "Building ACSGO binary..."
  cd "$INSTALL_DIR"
  go mod download
  CGO_ENABLED=0 go build -o "$BINARY_PATH" ./cmd/acsgo
  chmod 0755 "$BINARY_PATH"
}

ensure_service_user() {
  if ! id -u "$RUN_USER" >/dev/null 2>&1; then
    useradd --system --home /var/lib/acsgo --shell /usr/sbin/nologin "$RUN_USER"
  fi

  mkdir -p /etc/acsgo /var/lib/acsgo /var/log/acsgo
  chown -R "$RUN_USER":"$RUN_GROUP" /var/lib/acsgo /var/log/acsgo
}

set_env_value() {
  local key="$1" value="$2"
  if grep -qE "^${key}=" "$ENV_FILE"; then
    sed -i "s|^${key}=.*|${key}=${value}|" "$ENV_FILE"
  else
    printf '%s=%s\n' "$key" "$value" >> "$ENV_FILE"
  fi
}

write_env_file() {
  touch "$ENV_FILE"
  chmod 0640 "$ENV_FILE"
  chown root:"$RUN_GROUP" "$ENV_FILE"

  # Defaults are applied only if value is not already set in file.
  grep -q '^PORT=' "$ENV_FILE" || set_env_value PORT "${PORT:-7548}"
  grep -q '^CWMP_PORT=' "$ENV_FILE" || set_env_value CWMP_PORT "${CWMP_PORT:-7547}"
  grep -q '^DB_TYPE=' "$ENV_FILE" || set_env_value DB_TYPE "${DB_TYPE:-sqlite}"
  grep -q '^DB_URL=' "$ENV_FILE" || set_env_value DB_URL "${DB_URL:-/var/lib/acsgo/acsgo.db}"
  grep -q '^ADMIN_EMAIL=' "$ENV_FILE" || set_env_value ADMIN_EMAIL "${ADMIN_EMAIL:-admin@acsgo.local}"
  grep -q '^ADMIN_PASSWORD=' "$ENV_FILE" || set_env_value ADMIN_PASSWORD "${ADMIN_PASSWORD:-change-me}"
  grep -q '^JWT_SECRET=' "$ENV_FILE" || set_env_value JWT_SECRET "${JWT_SECRET:-change-this-secret}"

  # Explicit installer overrides always win.
  [[ -n "${PORT:-}" ]] && set_env_value PORT "$PORT"
  [[ -n "${CWMP_PORT:-}" ]] && set_env_value CWMP_PORT "$CWMP_PORT"
  [[ -n "${DB_TYPE:-}" ]] && set_env_value DB_TYPE "$DB_TYPE"
  [[ -n "${DB_URL:-}" ]] && set_env_value DB_URL "$DB_URL"
  [[ -n "${ADMIN_EMAIL:-}" ]] && set_env_value ADMIN_EMAIL "$ADMIN_EMAIL"
  [[ -n "${ADMIN_PASSWORD:-}" ]] && set_env_value ADMIN_PASSWORD "$ADMIN_PASSWORD"
  [[ -n "${JWT_SECRET:-}" ]] && set_env_value JWT_SECRET "$JWT_SECRET"
}

write_systemd_service() {
  log "Writing systemd unit ${SERVICE_FILE}..."
  cat > "$SERVICE_FILE" <<UNIT
[Unit]
Description=ACSGO Service
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=${RUN_USER}
Group=${RUN_GROUP}
WorkingDirectory=/var/lib/acsgo
EnvironmentFile=-${ENV_FILE}
ExecStart=${BINARY_PATH}
Restart=always
RestartSec=3
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
UNIT

  systemctl daemon-reload
  systemctl enable --now acsgo.service
}

post_install_info() {
  log "Installation complete."
  log "Service status: systemctl status acsgo --no-pager"
  log "Health check: curl -fsS http://127.0.0.1:${PORT:-7548}/health"
}

main() {
  require_root
  install_deps
  sync_repo
  ensure_service_user
  write_env_file
  build_binary
  write_systemd_service
  post_install_info
}

main "$@"
