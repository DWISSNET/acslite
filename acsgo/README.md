# ACSGO — High-Performance Distributed TR-069 ACS Server

[![Go](https://img.shields.io/badge/Go-1.22-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**ACSGO** transforms the DWISSNET/acslite TypeScript ACS server into a production-grade,
distributed TR-069/CWMP ACS server written in Go — capable of handling **millions of
concurrent devices** with enterprise reliability.

---

## 🚀 Performance Targets

| Metric             | Target       |
|--------------------|--------------|
| Concurrent devices | 10M+         |
| Informs/second     | 100K+        |
| RPC latency (p99)  | < 100ms      |
| Availability       | 99.99%       |
| Memory per device  | < 1KB        |

---

## 🏗 Architecture

```
┌───────────────────────────────────────────────────────────┐
│                        Devices (CPE)                       │
└───────────────────┬───────────────────────────────────────┘
                    │ TR-069 / CWMP (port 7547)
          ┌─────────▼──────────┐
          │    CWMP Server     │ ← stateful session tracking
          │  (services/cwmp/)  │ ← SOAP/XML marshaling
          └────────┬───────────┘
                   │
     ┌─────────────┼───────────────┐
     │             │               │
┌────▼────┐  ┌─────▼─────┐  ┌─────▼──────┐
│ Redis   │  │ PostgreSQL│  │ RabbitMQ   │
│ Cache   │  │ Database  │  │ Command Q  │
│ Sessions│  │ Devices   │  │ Dead-Letter│
└─────────┘  └─────┬─────┘  └────────────┘
                   │
          ┌────────▼──────────┐
          │    REST API        │ ← Gin framework (port 7548)
          │  + WebSocket Hub   │ ← real-time dashboard
          │  + Prometheus      │ ← /metrics endpoint
          └───────────────────┘
```

### Key Components

| Component              | Location                  | Technology        |
|------------------------|---------------------------|-------------------|
| TR-069/CWMP server     | `services/cwmp/`          | net/http           |
| Session state machine  | `services/cwmp/session/`  | Go + Redis         |
| SOAP/XML marshaling    | `services/cwmp/soap/`     | encoding/xml       |
| RPC command queue      | `services/queue/`         | RabbitMQ AMQP      |
| Distributed cache      | `pkg/cache/`              | Redis              |
| Device repository      | `pkg/repository/`         | PostgreSQL (pgx)   |
| Prometheus metrics     | `services/metrics/`       | prometheus/client  |
| Analytics engine       | `services/analytics/`     | Redis pub/sub      |
| Workflow automation    | `services/automation/`    | Event-driven       |
| Background scheduler   | `services/scheduler/`     | Distributed locks  |
| REST API               | `api/http/`               | Gin                |
| WebSocket dashboard    | `api/ws/`                 | gorilla/websocket  |
| PostgreSQL schema      | `schema/`                 | SQL migrations     |
| Kubernetes manifests   | `deploy/k8s/`             | K8s StatefulSet    |

---

## 📁 Directory Structure

```
acsgo/
├── cmd/acsgo/main.go          # Entry point
├── pkg/
│   ├── config/config.go       # Configuration from env
│   ├── models/                # Data structures (Device, Command, etc.)
│   ├── repository/            # PostgreSQL data access layer
│   └── cache/redis.go         # Redis cache + distributed locks
├── services/
│   ├── cwmp/                  # TR-069/CWMP HTTP server
│   │   ├── server.go          # Main CWMP handler
│   │   ├── soap/soap.go       # SOAP/XML builder & parser
│   │   └── session/manager.go # Session state machine
│   ├── queue/rabbitmq.go      # RabbitMQ producer/consumer
│   ├── metrics/prometheus.go  # Prometheus metrics
│   ├── scheduler/scheduler.go # Distributed background jobs
│   ├── analytics/aggregator.go # Real-time analytics engine
│   └── automation/engine.go   # Event-driven automation rules
├── api/
│   ├── http/                  # Gin REST API server
│   │   ├── server.go          # Routes & middleware
│   │   └── handlers/          # Request handlers
│   └── ws/hub.go              # WebSocket hub
├── schema/001_initial.sql     # PostgreSQL schema
├── deploy/
│   ├── k8s/deployment.yaml    # Kubernetes manifests
│   └── prometheus.yml         # Prometheus config
├── web/
│   ├── index.html             # Real-time dashboard
│   └── login.html             # Login page
├── docker-compose.yml         # Local dev stack
├── Dockerfile                 # Multi-stage build
├── Makefile                   # Dev commands
└── .env.example               # Configuration template
```

---

## ⚡ Quick Start

### Local Development (Docker Compose)

```bash
# 1. Clone & enter directory
cd acsgo

# 2. Copy and configure environment
cp .env.example .env
# Edit .env with your values

# 3. Start all services
docker compose up -d

# 4. Access the dashboard
open http://localhost:7548/login
# Default: admin / admin123
```

### Manual Build

```bash
# Prerequisites: Go 1.22+, PostgreSQL 15+, Redis 7+

# Install dependencies
go mod download

# Apply schema
psql "$DATABASE_URL" -f schema/001_initial.sql

# Run
make run
# or
go run ./cmd/acsgo/main.go
```

### Kubernetes Deployment

```bash
# Apply manifests
kubectl apply -f deploy/k8s/deployment.yaml

# Scale CWMP pods
kubectl scale statefulset acsgo-cwmp --replicas=10 -n acsgo
```

---

## 📡 TR-069 Device Configuration

Point your CPE/modem to the ACSGO server:

| Setting              | Value                              |
|----------------------|------------------------------------|
| ACS URL              | `http://YOUR_SERVER:7547/cwmp`     |
| ACS Username         | `acsadmin` (configurable)          |
| ACS Password         | See `ACS_PASSWORD` env var         |
| Connection Request   | Port 7547                          |

---

## 🔌 REST API

| Method | Path                            | Description                  |
|--------|---------------------------------|------------------------------|
| POST   | `/api/auth/login`               | Authenticate, get JWT        |
| GET    | `/api/devices`                  | List devices (paginated)     |
| GET    | `/api/devices/:id`              | Get device details           |
| PUT    | `/api/devices/:id`              | Update customer/location/tags|
| GET    | `/api/devices/:id/parameters`   | Get TR-069 parameters        |
| POST   | `/api/devices/:id/command`      | Send RPC command             |
| POST   | `/api/devices/bulk-command`     | Bulk RPC (all devices)       |
| GET    | `/metrics`                      | Prometheus metrics           |
| GET    | `/healthz`                      | Liveness probe               |
| GET    | `/readyz`                       | Readiness probe              |
| WS     | `/ws`                           | Real-time WebSocket stream   |

### Example: Reboot a device

```bash
curl -X POST http://localhost:7548/api/devices/DEVICE_UUID/command \
  -H "Authorization: ******" \
  -H "Content-Type: application/json" \
  -d '{"type": "Reboot"}'
```

### Example: Bulk firmware update

```bash
curl -X POST http://localhost:7548/api/devices/bulk-command \
  -H "Authorization: ******" \
  -H "Content-Type: application/json" \
  -d '{
    "device_ids": ["id1", "id2", "id3"],
    "type": "Download",
    "parameters": {
      "file_type": "1 Firmware Upgrade Image",
      "url": "http://firmware.example.com/v2.0.0.bin",
      "file_size": "10485760",
      "target_filename": "firmware.bin"
    }
  }'
```

---

## 📊 Supported RPC Commands

| Command               | TR-069 Method              |
|-----------------------|----------------------------|
| Reboot                | `Reboot`                   |
| Factory Reset         | `FactoryReset`             |
| Get Parameters        | `GetParameterValues`       |
| Set Parameters        | `SetParameterValues`       |
| Get Parameter Names   | `GetParameterNames`        |
| Firmware Download     | `Download`                 |
| Upload                | `Upload`                   |
| Schedule Inform       | `ScheduleInform`           |

---

## 🔧 Configuration Reference

| Variable                | Default                   | Description                  |
|-------------------------|---------------------------|------------------------------|
| `PORT`                  | `7548`                    | REST API port                |
| `CWMP_PORT`             | `7547`                    | TR-069/CWMP port             |
| `DATABASE_URL`          | (required)                | PostgreSQL connection string |
| `REDIS_URL`             | `redis://localhost:6379`  | Redis URL                    |
| `RABBITMQ_URL`          | (optional)                | RabbitMQ AMQP URL            |
| `JWT_SECRET`            | (required)                | JWT signing secret           |
| `ACS_USERNAME`          | `acsadmin`                | CWMP Basic Auth username     |
| `ACS_PASSWORD`          | (empty = disabled)        | CWMP Basic Auth password     |
| `PROMETHEUS_ENABLED`    | `true`                    | Enable /metrics endpoint     |
| `LOG_LEVEL`             | `info`                    | debug/info/warn/error        |

---

## 🏆 Why Go Over TypeScript?

| Aspect               | ACSGO (Go)                    | acslite (TypeScript)          |
|----------------------|-------------------------------|-------------------------------|
| Concurrency          | Goroutines (< 2KB each)       | Event loop (single thread)    |
| Memory per conn      | ~1-2KB                        | ~10-100KB                     |
| Startup time         | < 50ms                        | 500ms+ (V8 init)              |
| Binary               | Single static executable      | Needs Node.js runtime         |
| HTTP/2               | Native net/http               | Via library                   |
| Type safety          | Compile-time                  | Erased at runtime             |
| Clustering           | Native (Redis + gRPC)         | Requires sticky sessions      |

---

## 📈 Monitoring

- **Prometheus**: `http://localhost:9090/metrics`
- **Grafana**: `http://localhost:3000` (admin / configured password)
- **RabbitMQ Management**: `http://localhost:15672`

---

## 📜 License

MIT — see [LICENSE](../LICENSE)
