#!/bin/bash
set -e

# Warna untuk output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

clear
echo -e "${BLUE}"
echo "            "
echo "      "
echo "                "
echo "                "
echo "                  "
echo "                        "
echo -e "${NC}"
echo -e "${GREEN}         AUTO INSTALLER ACS LITE UNTUK VPS LINUX (UBUNTU/DEBIAN)${NC}"
echo "=================================================================="
echo ""

# Cek root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED} Jalankan script ini sebagai root (sudo su)!${NC}"
    exit 1
fi

# Update system
echo -e "${YELLOW} Update system packages...${NC}"
apt update -y
apt upgrade -y

# Install dependencies
echo -e "${YELLOW} Installing dependencies...${NC}"
apt install -y git nodejs npm docker.io docker-compose fail2ban ufw openssl

# Enable & start services
systemctl enable --now docker
systemctl enable --now fail2ban

# Setup firewall
echo -e "${YELLOW} Konfigurasi UFW Firewall...${NC}"
ufw allow ssh
ufw allow 7547/tcp  # TR-069 CWMP
ufw allow 7548/tcp  # Web API/Dashboard
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# Clone repository
echo -e "${YELLOW} Clone ACS repository...${NC}"
cd /opt
git clone https://github.com/DWISSNET/acslite.git
cd acslite

# Copy .env
cp .env.example .env

# Generate random password untuk keamanan
MONGO_PASS=$(openssl rand -hex 16)
ACS_PASS=$(openssl rand -hex 8)
JWT_SECRET=$(openssl rand -hex 32)
sed -i "s/your_mongodb_password/$MONGO_PASS/g" .env
sed -i "s/acs_secure_password/$ACS_PASS/g" .env
sed -i "s/change_this_jwt_secret/$JWT_SECRET/g" .env

# Build & start dengan docker-compose
echo -e "${YELLOW} Start semua container dengan Docker...${NC}"
docker-compose up -d --build

# Tampilkan info selesai
SERVER_IP=$(curl -s ifconfig.me)
echo ""
echo -e "${GREEN} INSTALLASI BERHASIL! ACS LITE SUDAH JALAN!${NC}"
echo "=================================================================="
echo -e "${BLUE} Informasi Server:${NC}"
echo "   IP Server:         $SERVER_IP"
echo "   CWMP/TR-069:       http://$SERVER_IP:7547"
echo "   Dashboard:         http://$SERVER_IP:7548/dashboard"
echo "   API:               http://$SERVER_IP:7548/api"
echo ""
echo -e "${BLUE} Credential Default:${NC}"
echo "   MongoDB Password:  $MONGO_PASS"
echo "   ACS Admin Pass:    $ACS_PASS"
echo ""
echo -e "${YELLOW}  CATAT: Simpan credential di atas dengan aman!${NC}"
echo "=================================================================="
