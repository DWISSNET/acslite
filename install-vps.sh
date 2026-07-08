# Buat install-vps.sh ulang
@"
#!/bin/bash
set -e

# Warna
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

clear
echo -e "${BLUE}███████╗███████╗██████╗ ███╗   ███╗${NC}"
echo -e "${GREEN}ACS LITE AUTO INSTALLER VPS${NC}"
echo "====================================="

# Cek root
if [ "$EUID" -ne 0 ]; then echo -e "${RED}Jalankan sebagai root!${NC}"; exit 1; fi

# Update & install
apt update -y && apt upgrade -y
apt install -y git nodejs npm docker.io docker-compose ufw fail2ban

# Firewall
ufw allow ssh && ufw allow 7547/tcp && ufw allow 7548/tcp && ufw --force enable

# Clone repo
cd /opt
git clone https://github.com/DWISSNET/acslite.git && cd acslite
cp .env.example .env

# Generate password
MONGO_PASS=$(openssl rand -hex 16)
sed -i "s/your_mongodb_password/$MONGO_PASS/g" .env

# Start docker
docker-compose up -d --build

echo -e "${GREEN}✅ Installasi Selesai!${NC}"
echo "Server IP: $(curl -s ifconfig.me)"
echo "Dashboard: http://$(curl -s ifconfig.me):7548/dashboard"
"@ | Out-File -FilePath "install-vps.sh" -Encoding utf8