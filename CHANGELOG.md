# Changelog - ACSGO (Go-first ACSLite)

## [Unreleased]
### 🔒 Security
- Fix XSS vulnerability in dashboard: device fields now rendered with `textContent` instead of raw `innerHTML`
- Add request body size limit (1 MiB) on `/api/devices/import` to prevent DoS via large payloads
- Default `ADMIN_PASSWORD` changed from `admin123` to `change-me` to match installer/`.env.example`
- Warn at startup when `JWT_SECRET` or `ADMIN_PASSWORD` are still set to insecure default values

### 🚀 Improvements
- Add graceful shutdown: server handles `SIGTERM`/`SIGINT` and drains in-flight requests before exiting
- Add HTTP server timeouts (`ReadTimeout: 15s`, `WriteTimeout: 30s`, `IdleTimeout: 60s`)
- Verify database connectivity with `PingContext` immediately after `sql.Open`
- Restrict `/api/stats` and `/api/devices` to `GET` only; return `405` for other methods
- Go binary (`/acsgo`) added to `.gitignore`; build artifact removed from repository tracking
- Fix PostgreSQL DSN example in `README.md` and `install.sh` (was masked with `******`)

## [1.0.0] - 2026-01-08
### 🎉 Go-first Rewrite
- Migrated from TypeScript/Node.js to Go
- SQLite default database; PostgreSQL optional via env vars
- `systemd` service managed by `install.sh`
- Go 1.22+ enforced; installed from official go.dev binaries
- REST API: `/health`, `/dashboard`, `/api/stats`, `/api/devices`, `/api/devices/import`
- Bulk device provisioning via JSON import API
- Live dashboard with 2-second polling
- Graceful installer with conflict port detection and post-install health check
