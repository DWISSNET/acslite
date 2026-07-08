# Deployment Guide for Advanced ACS Server

## 🚀 Quick Deploy dengan Docker (REKOMENDED)
Cara termudah untuk production:

```bash
# 1. Copy env
cp .env.example .env
# Edit .env sesuai konfigurasi server kamu

# 2. Start semua services
docker-compose up -d

# Cek status
docker-compose ps
```

## 📦 Manual Install (Tanpa Docker)
### Prasyarat
- Node.js 18+
- MongoDB 6+
- Redis 7+

```bash
# Install dependencies
npm install

# Build
npm run build

# Production start
npm start
```

## 🔌 Konfigurasi Modem untuk Connect ke ACS
Tambahkan konfigurasi TR-069 di modem kamu:
- **ACS URL**: `http://[IP-SERVER-KAMU]:7547`
- **Periodic Inform Interval**: 300 (5 menit)
- **Connection Request Username/Password**: Sesuai .env

## 📊 Akses Dashboard
Buka di browser: `http://[IP-SERVER]:7548/dashboard`

## 📡 API Endpoints
| Endpoint | Method | Deskripsi |
|----------|--------|-----------|
| `/api/devices` | GET | List semua modem |
| `/api/devices/:serial` | GET | Detail satu modem |
| `/api/devices/:serial/reboot` | POST | Reboot modem |
| `/api/parameters` | GET | List semua parameter |
| `/api/provisioning/bulk` | POST | Bulk provisioning |
| `/api/stats` | GET | Server statistics |

## 🔒 Security Best Practices
1. Ganti semua default password di `.env`
2. Gunakan HTTPS untuk akses dashboard
3. Batasi akses port 7547 hanya ke IP modem
4. Backup MongoDB setiap hari
5. Update dependency secara rutin