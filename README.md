# ACSGO (Go-first ACSLite)

This repository now ships a **Go-first ACS implementation** (`acsgo`) and removes the legacy TypeScript/Node stack from `main`.

## Quick Start

```bash
curl -fsSL https://raw.githubusercontent.com/DWISSNET/acslite/main/install.sh | bash
```

Installer behavior:
- Builds ACSGO from this repository source
- Installs/updates a `systemd` service (`acsgo.service`)
- Installs/updates **Go 1.22+** from official Go binaries
- Uses **SQLite by default** (`/var/lib/acsgo/acsgo.db`)
- Supports **PostgreSQL optionally** via env vars

## PostgreSQL (optional)

Set environment variables before running installer:

```bash
DB_TYPE=postgres \
DB_URL='postgres://user:pass@127.0.0.1:5432/acsgo?sslmode=disable' \
PORT=7548 \
CWMP_PORT=7547 \
ADMIN_EMAIL=admin@acsgo.local \
ADMIN_PASSWORD='change-me' \
JWT_SECRET='change-this-secret' \
curl -fsSL https://raw.githubusercontent.com/DWISSNET/acslite/main/install.sh | bash
```

## Local Build

```bash
go build ./cmd/acsgo
./acsgo
```

Default runtime configuration (env vars):
- `PORT` (default `7548`)
- `CWMP_PORT` (default `7547`)
- `DB_TYPE` (`sqlite` default, `postgres` optional)
- `DB_URL` (default `./data/acsgo.db`)
- `ADMIN_EMAIL`
- `ADMIN_PASSWORD`
- `JWT_SECRET`

Health endpoint:

```bash
curl http://127.0.0.1:7548/health
```

Dashboard endpoint:

```bash
open http://127.0.0.1:7548/dashboard
```

## Troubleshooting

### 404 on raw install URL
- Ensure URL is exactly:
  `https://raw.githubusercontent.com/DWISSNET/acslite/main/install.sh`
- If you are using an older branch/tag, switch to `main` or use clone fallback.

### 429 from raw.githubusercontent.com
Use clone fallback:

```bash
git clone https://github.com/DWISSNET/acslite.git
cd acslite
sudo bash install.sh
```

## Migration Note

Legacy TypeScript/Node artifacts were removed from `main` (including `src/`, `public/`, `package.json`, `package-lock.json`, `tsconfig.json`, and old VPS installers) as part of the Go-first migration.
