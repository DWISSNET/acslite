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
LOG_FILE="${LOG_FILE:-/tmp/acsgo-installer.log}"
REQUIRED_GO_VERSION="${REQUIRED_GO_VERSION:-1.22.0}"
GO_DOWNLOAD_VERSION="${GO_DOWNLOAD_VERSION:-1.22.5}"

mkdir -p "$(dirname "$LOG_FILE")"
touch "$LOG_FILE"
exec > >(tee -a "$LOG_FILE") 2>&1

log() { printf '[acsgo-installer] %s\n' "$*"; }
fail() { printf '[acsgo-installer] ERROR: %s\n' "$*" >&2; exit 1; }
on_error() {
  local line="$1"
  fail "Installer failed at line ${line}. See log: ${LOG_FILE}"
}
trap 'on_error $LINENO' ERR

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
    ca-certificates curl git build-essential pkg-config sqlite3 tar
}

version_gte() {
  local lhs="$1" rhs="$2"
  [[ "$(printf '%s\n%s\n' "$rhs" "$lhs" | sort -V | head -n1)" == "$rhs" ]]
}

installed_go_version() {
  if ! command -v go >/dev/null 2>&1; then
    return 1
  fi
  go version 2>/dev/null | awk '{print $3}' | sed 's/^go//'
}

cleanup_old_go() {
  log "Cleaning old Go installations..."
  rm -rf /usr/local/go

  if dpkg -s golang-go >/dev/null 2>&1; then
    apt-get remove -y golang-go || true
  fi
  apt-get autoremove -y || true

  rm -f /usr/bin/go /usr/local/bin/go
}

install_or_upgrade_go() {
  local current_go=""
  current_go="$(installed_go_version || true)"
  if [[ -n "$current_go" ]] && version_gte "$current_go" "$REQUIRED_GO_VERSION"; then
    log "Go ${current_go} already satisfies requirement >= ${REQUIRED_GO_VERSION}."
    return
  fi

  local arch go_arch tarball url
  arch="$(uname -m)"
  case "$arch" in
    x86_64) go_arch="amd64" ;;
    aarch64|arm64) go_arch="arm64" ;;
    *)
      fail "Unsupported architecture for Go official binary: ${arch}"
      ;;
  esac

  cleanup_old_go
  tarball="go${GO_DOWNLOAD_VERSION}.linux-${go_arch}.tar.gz"
  url="https://go.dev/dl/${tarball}"
  log "Installing Go ${GO_DOWNLOAD_VERSION} from official binary (${url})..."
  curl -fsSL "$url" -o "/tmp/${tarball}"
  tar -C /usr/local -xzf "/tmp/${tarball}"
  rm -f "/tmp/${tarball}"
  ln -sf /usr/local/go/bin/go /usr/local/bin/go
  ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt
}

verify_go() {
  local current_go=""
  current_go="$(installed_go_version || true)"
  if [[ -z "$current_go" ]]; then
    fail "Go binary is not available after installation."
  fi
  if ! version_gte "$current_go" "$REQUIRED_GO_VERSION"; then
    fail "Go ${current_go} is too old. Require >= ${REQUIRED_GO_VERSION}."
  fi
  log "Using Go ${current_go}."
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
  export PATH="/usr/local/go/bin:${PATH}"
  verify_go
  go mod download
  CGO_ENABLED=0 go build -o "$BINARY_PATH" ./cmd/acsgo
  [[ -x "$BINARY_PATH" ]] || fail "Build did not create executable at ${BINARY_PATH}"
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

read_env_value() {
  local key="$1"
  awk -F= -v k="$key" '$1 == k {print substr($0, index($0,$2)); exit}' "$ENV_FILE"
}

kill_conflicting_ports() {
  local acs_port cwmp_port port pids pid
  systemctl stop acsgo.service 2>/dev/null || true
  acs_port="$(read_env_value PORT)"
  cwmp_port="$(read_env_value CWMP_PORT)"

  for port in "$acs_port" "$cwmp_port"; do
    [[ -n "$port" ]] || continue
    pids="$(ss -ltnp "sport = :${port}" 2>/dev/null | awk -F'pid=' 'NR>1 && /pid=/{split($2,a,","); print a[1]}' | sort -u)"
    if [[ -n "$pids" ]]; then
      log "Found process(es) on port ${port}: ${pids}. Stopping them."
      for pid in $pids; do
        [[ -n "$pid" ]] || continue
        kill "$pid" 2>/dev/null || true
      done
    fi
  done
}

verify_service_health() {
  local port status_url attempt
  port="$(read_env_value PORT)"
  [[ -n "$port" ]] || port="7548"
  status_url="http://127.0.0.1:${port}/health"

  for attempt in $(seq 1 30); do
    if curl -fsS "$status_url" >/dev/null 2>&1; then
      log "Health check passed (${status_url}) on attempt ${attempt}."
      return
    fi
    sleep 2
  done

  journalctl -u acsgo -n 50 --no-pager || true
  fail "Health check failed after retries: ${status_url}"
}

post_install_info() {
  local port admin_email admin_secret
  port="$(read_env_value PORT)"
  [[ -n "$port" ]] || port="7548"
  admin_email="$(read_env_value ADMIN_EMAIL)"
  [[ -n "$admin_email" ]] || admin_email="admin@acsgo.local"
  admin_secret="$(read_env_value ADMIN_PASSWORD)"
  [[ -n "$admin_secret" ]] || admin_secret="change-me"

  log "Installation complete."
  log "Installer log: ${LOG_FILE}"
  log "Service status: systemctl status acsgo --no-pager"
  log "Health check: curl -fsS http://127.0.0.1:${port}/health"
  log "Dashboard: http://127.0.0.1:${port}/dashboard"
  log "Admin login: ${admin_email}"
  log "Admin password: ${admin_secret}"
}

main() {
  require_root
  install_deps
  install_or_upgrade_go
  verify_go
  sync_repo
  ensure_service_user
  write_env_file
  build_binary
  kill_conflicting_ports
  write_systemd_service
  verify_service_health
  post_install_info
}

main "$@"
