New-Item -Path "h:\AUTO INSTALER GENIEACS\install-vps-nodocker.sh" -ItemType File -Value @"
#!/bin/bash
set -euo pipefail

# Warna output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

clear
echo -e "${BLUE}"
echo "███████╗███████╗██████╗ ███╗   ███╗    ████████╗███████╗██████╗ ███╗   ███╗"
echo "██╔════╝██╔════╝██╔══██╗████╗ ████║    ╚══██╔══╝██╔════╝██╔══██╗████╗ ████║"
echo "█████╗  █████╗  ██████╔╝██╔████╔██║       ██║   █████╗  ██████╔╝██╔████╔██║"
echo "██╔══╝  ██╔══╝  ██╔══██╗██║╚██╔╝██║       ██║   ██╔══╝  ██╔══██╗██║╚██╔╝██║"
echo "███████╗███████╗██║  ██║██║ ╚═╝ ██║       ██║   ███████╗██║  ██║██║ ╚═╝ ██║"
echo "╚══════╝╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝       ╚═╝   ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝"
echo -e "${NC}"
echo -e "${GREEN}    ACS LITE - INSTALL TANPA DOCKER (LANGSUNG DI HOST)${NC}"
echo "=================================================================="
echo ""

# Cek root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}❌ ERROR: Jalankan script ini sebagai root! sudo su dulu${NC}"
    exit 1
fi

# Update sistem
echo -e "${YELLOW}🔄 Update sistem...${NC}"
apt update -y && apt full-upgrade -y

# Hapus nodejs lama
echo -e "${YELLOW}🧹 Bersihkan nodejs lama...${NC}"
apt remove -y nodejs libnode-dev || true
apt autoremove -y

# Install MongoDB + Redis (langsung di host)
echo -e "${YELLOW}📦 Install MongoDB 6.0 + Redis...${NC}"
wget -qO - https://www.mongodb.org/static/pgp/server-6.0.asc | apt-key add -
echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu jammy/mongodb-org/6.0 multiverse" | tee /etc/apt/sources.list.d/mongodb-org-6.0.list
add-apt-repository -y ppa:redislabs/redis
apt update -y
apt install -y mongodb-org redis-server git curl openssl ufw fail2ban build-essential

# Install Node.js 20 LTS
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt install -y nodejs

# Aktifkan semua service
systemctl enable --now mongod
systemctl enable --now redis-server

# Setup Firewall
echo -e "${YELLOW}🔥 Setup Firewall...${NC}"
ufw allow ssh
ufw allow 7547/tcp
ufw allow 7548/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# Clone repo ACS
cd /opt
if [ -d "acslite" ]; then
    cd acslite
    git pull origin main
else
    git clone https://github.com/DWISSNET/acslite.git
    cd acslite
fi

# Setup .env
if [ ! -f .env ]; then
    cp .env.example .env
fi

# Generate password
MONGO_PASS=$(openssl rand -hex 16)
ACS_PASS=$(openssl rand -hex 8)
JWT_SECRET=$(openssl rand -hex 32)
sed -i "s/your_mongodb_password/$MONGO_PASS/g" .env
sed -i "s/acs_secure_password/$ACS_PASS/g" .env
sed -i "s/change_this_jwt_secret/$JWT_SECRET/g" .env

# Setup MongoDB user
mongosh <<EOF
use admin
db.createUser({user: "acsadmin", pwd: "$MONGO_PASS", roles: [{role: "root", db: "admin"}]})
EOF

# Install NPM & build
npm install
npm run build

# Buat systemd service
cat > /etc/systemd/system/acslite.service <<EOF
[Unit]
Description=ACS Lite TR-069 Server
After=network.target mongod.target redis-server.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/acslite
ExecStart=/usr/bin/npm start
Restart=always
RestartSec=5
Environment=NODE_ENV=production

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now acslite

# Selesai
SERVER_IP=$(curl -s ifconfig.me)
echo ""
echo -e "${GREEN}🎉🎉🎉 SEMUA BERHASIL! ACS LITE BERJALAN!${NC}"
echo "=================================================================="
echo "   IP VPS:           $SERVER_IP"
echo "   TR-069:           http://$SERVER_IP:7547"
echo "   Dashboard:        http://$SERVER_IP:7548/dashboard"
echo "   MongoDB Pass:     $MONGO_PASS"
echo "   Admin Pass:       $ACS_PASS"
echo "   Cek status: systemctl status acslite"
echo "=================================================================="
"@
