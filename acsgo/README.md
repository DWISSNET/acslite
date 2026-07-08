# ACSGO — Lightweight TR-069 ACS Server in Go

A **production-ready, minimal footprint** TR-069/CWMP Auto Configuration Server written in Go.  
Drop-in replacement for acslite (TypeScript) with superior performance and single-binary deployment.

---

## ✨ Features

| Feature | Status |
|---|---|
| Full TR-069/CWMP protocol compliance | ✅ |
| Device Inform handling & auto-detection | ✅ |
| 100+ TR-069 standard + vendor-specific parameters | ✅ |
| Huawei, ZTE, TP-Link, D-Link, MikroTik support | ✅ |
| ISP templates (Indihome, First Media, Biznet, etc.) | ✅ |
| Device management (reboot, factory-reset, params) | ✅ |
| Bulk device import/provisioning | ✅ |
| Real-time web dashboard | ✅ |
| REST API | ✅ |
| WebSocket live updates | ✅ |
| Event logging & history | ✅ |
| JWT authentication | ✅ |
| SQLite embedded (zero config) | ✅ |
| PostgreSQL (for scaling) | ✅ |
| Single binary deployment | ✅ |
| Docker Compose | ✅ |
| Systemd service | ✅ |

---

## 🚀 Quick Start

### Option A — One-command VPS install (recommended)

```bash
curl -sSL https://raw.githubusercontent.com/DWISSNET/acslite/main/acsgo/deploy/install.sh | sudo bash
```

This will:
1. Install Go (if not present)
2. Build the `acsgo` binary from source
3. Create `/opt/acsgo/.env` config file
4. Install & start the `acsgo` systemd service
5. Open firewall ports 7547 (CWMP) and 7548 (Web)

After install, access the dashboard at: **http://YOUR_VPS_IP:7548/dashboard**

### Option B — Docker Compose (dev/testing)

```bash
cd acsgo/deploy
docker-compose up -d
```

### Option C — Build from source

```bash
cd acsgo
go build -o acsgo .
./acsgo
```

---

## ⚙️ Configuration

Copy `deploy/.env.example` to `acsgo/.env` and edit:

```env
ENV=production
CWMP_PORT=7547          # Port for TR-069 devices
PORT=7548               # Port for web dashboard & API
DB_TYPE=sqlite          # sqlite or postgres
DB_URL=./data/acsgo.db  # SQLite file path
JWT_SECRET=changeme     # Random secret for JWT signing
ADMIN_EMAIL=admin@acsgo.local
ADMIN_PASSWORD=admin123
```

For **PostgreSQL**:
```env
DB_TYPE=postgres
DB_URL=******localhost:5432/acsgo?sslmode=disable
```

---

## 📡 Configure Your Modem (ACS URL)

Point your modem's TR-069 ACS URL to:
```
http://YOUR_SERVER_IP:7547/cwmp
```

| Setting | Value |
|---|---|
| ACS URL | `http://YOUR_IP:7547/cwmp` |
| ACS Username | *(leave empty or set in .env)* |
| ACS Password | *(leave empty or set in .env)* |
| Inform Interval | `300` (5 minutes) |

---

## 🔌 API Reference

All API endpoints return JSON. Protected routes require:  
`Authorization: ******

### Authentication

```
POST /api/auth/login     { "email": "...", "password": "..." }
POST /api/auth/logout
GET  /api/auth/me
```

### Device Management

```
GET    /api/devices                         # List devices (pagination, search)
GET    /api/devices/:serial                 # Device details + parameters
POST   /api/devices/:serial/reboot         # Queue reboot command
POST   /api/devices/:serial/factory-reset  # Queue factory reset
PUT    /api/devices/:serial/params         # Queue SetParameterValues
GET    /api/devices/:serial/events         # Event history
DELETE /api/devices/:serial                # Remove device
```

### Parameters

```
GET /api/parameters                    # All parameters (grouped by category)
GET /api/parameters?category=WiFi      # Filter by category
GET /api/parameters?vendor=huawei      # Filter by vendor
GET /api/parameters/category/:cat      # Parameters for category
GET /api/parameters/vendor/:vendor     # Parameters for vendor
GET /api/parameters/:path              # Single parameter details
```

### Provisioning

```
POST /api/provisioning/bulk            # Bulk device import
GET  /api/provisioning/templates       # List ISP templates
GET  /api/provisioning/template/:isp   # Get ISP template (indihome, firstmedia, etc.)
```

### System

```
GET /api/health    # Health check
GET /api/stats     # Server statistics
```

### WebSocket

Connect to `ws://YOUR_SERVER:7548/ws` for real-time events:

```json
{ "event": "device:online",  "data": { "serial": "...", "ip": "...", "model": "..." } }
{ "event": "device:offline", "data": { "count": 5 } }
```

---

## 📊 Performance

| Metric | Target |
|---|---|
| Binary size | ~50MB |
| Memory idle | ~30MB |
| Memory per device | <1KB |
| Concurrent devices | 1M+ (SQLite: 500K+) |
| Informs/second | 100K+ |
| Startup time | <1 second |

---

## 🔧 Scaling: SQLite → PostgreSQL

1. Export data from SQLite:
   ```bash
   sqlite3 data/acsgo.db .dump > backup.sql
   ```

2. Update `.env`:
   ```env
   DB_TYPE=postgres
   DB_URL=******localhost:5432/acsgo?sslmode=disable
   ```

3. Restart:
   ```bash
   systemctl restart acsgo
   ```

The schema is applied automatically on first start.

---

## 🏗️ Project Structure

```
acsgo/
├── main.go              # Entry point, server setup
├── go.mod               # Go module
├── Makefile             # Build commands
├── config/              # Configuration loading
├── models/              # Data structures
├── db/                  # Migrations & schema
├── services/            # Business logic (CWMP, devices, etc.)
├── handlers/            # HTTP request handlers
├── middleware/          # Auth, CORS, logging
├── pkg/                 # SOAP/XML, JWT, utilities
├── seeds/               # TR-069 parameter templates
├── web/                 # Single-file dashboard
├── deploy/              # Install scripts, Dockerfile
└── tests/               # Unit + integration tests
```

---

## 🧪 Running Tests

```bash
cd acsgo
go test ./tests/... -v
```

---

## 🔒 Security

- JWT tokens expire in 24 hours
- Passwords compared via bcrypt (or plain text fallback for `.env` config)
- CORS configured for all origins (suitable for LAN/VPS deployments)
- Change `ADMIN_PASSWORD` and `JWT_SECRET` before production use

---

## 📋 Supported Modems

| Vendor | Models |
|---|---|
| **Huawei** | HG8245H, HG8546M, EG8141A5, EG8145V5, HS8546V5, B618, B525 |
| **ZTE** | F609, F670L, F670Y, F680, ZXHN F601, MF286R |
| **TP-Link** | XC220-G3v, Archer AX1800 |
| **MikroTik** | RB750Gr3 (hEX), RB4011iGS+, CCR1009 |
| **D-Link** | DPR-1041, DSL-2750U |

---

## 📜 License

MIT — see [LICENSE](../LICENSE) for details.
