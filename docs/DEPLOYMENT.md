# Guide de déploiement OnlyFlick - Messagerie et Profils

## Prérequis

### Serveur
- **OS** : Ubuntu 20.04+ ou CentOS 8+
- **RAM** : 4GB minimum, 8GB recommandé
- **CPU** : 2 cores minimum, 4 cores recommandé
- **Stockage** : 50GB minimum (pour les uploads)

### Logiciels
- **Go** : Version 1.19+ 
- **PostgreSQL** : Version 13+
- **Nginx** : Pour le reverse proxy
- **Git** : Pour le déploiement

## Installation des dépendances

### 1. Installation de Go
```bash
# Télécharger et installer Go
cd /tmp
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# Ajouter Go au PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 2. Installation de PostgreSQL
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install postgresql postgresql-contrib

# CentOS/RHEL
sudo dnf install postgresql-server postgresql-contrib
sudo postgresql-setup --initdb
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

### 3. Configuration de la base de données
```bash
# Se connecter à PostgreSQL
sudo -u postgres psql

# Créer la base de données et l'utilisateur
CREATE DATABASE only_flick_db;
CREATE USER only_flick_user WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE only_flick_db TO only_flick_user;

# Activer les extensions nécessaires
\c only_flick_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
\q
```

## Déploiement de l'application

### 1. Cloner le repository
```bash
cd /opt
sudo git clone https://github.com/your-org/Only-Flick-Go.git
sudo chown -R $USER:$USER Only-Flick-Go
cd Only-Flick-Go
```

### 2. Configuration de l'environnement
```bash
# Créer le fichier .env
cat > .env << EOF
# Base de données
DB_HOST=localhost
DB_PORT=5432
DB_NAME=only_flick_db
DB_USER=only_flick_user
DB_PASSWORD=secure_password
DB_SSLMODE=require

# Serveur
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_EXPIRATION_HOURS=24

# Upload
UPLOAD_PATH=/opt/Only-Flick-Go/uploads
MAX_FILE_SIZE=52428800
MIN_PAID_MESSAGE_PRICE=0.99
MAX_PAID_MESSAGE_PRICE=500.00

# Messaging
MAX_MESSAGE_LENGTH=5000
MAX_MESSAGES_PER_MINUTE=30

# Environment
ENVIRONMENT=production
LOG_LEVEL=info
EOF
```

### 3. Créer les dossiers nécessaires
```bash
mkdir -p uploads/{messages,avatars,banners,thumbnails}
chmod 755 uploads
chmod 755 uploads/*
```

### 4. Installation des dépendances Go
```bash
go mod download
go mod verify
```

### 5. Exécuter les migrations
```bash
# Appliquer les migrations
psql -h localhost -U only_flick_user -d only_flick_db -f internal/database/migrations/001_initial_schema.sql
psql -h localhost -U only_flick_user -d only_flick_db -f internal/database/migrations/002_messaging_profile_tables.sql
```

### 6. Compiler l'application
```bash
cd cmd/api
go build -o onlyflick-api .
```

### 7. Créer le service systemd
```bash
sudo cat > /etc/systemd/system/onlyflick-api.service << EOF
[Unit]
Description=OnlyFlick API Service
After=network.target postgresql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/Only-Flick-Go
Environment=PATH=/usr/local/go/bin:/usr/bin:/bin
EnvironmentFile=/opt/Only-Flick-Go/.env
ExecStart=/opt/Only-Flick-Go/cmd/api/onlyflick-api
Restart=always
RestartSec=10

# Sécurité
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/opt/Only-Flick-Go/uploads
ProtectHome=true

[Install]
WantedBy=multi-user.target
EOF

# Recharger systemd et démarrer le service
sudo systemctl daemon-reload
sudo systemctl enable onlyflick-api
sudo systemctl start onlyflick-api
```

## Configuration Nginx

### 1. Installation de Nginx
```bash
# Ubuntu/Debian
sudo apt install nginx

# CentOS/RHEL
sudo dnf install nginx
```

### 2. Configuration du reverse proxy
```bash
sudo cat > /etc/nginx/sites-available/onlyflick << EOF
server {
    listen 80;
    server_name api.onlyflick.com;

    # Redirections HTTPS
    return 301 https://\$server_name\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.onlyflick.com;

    # Certificats SSL (Let's Encrypt)
    ssl_certificate /etc/letsencrypt/live/api.onlyflick.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.onlyflick.com/privkey.pem;
    
    # Configuration SSL moderne
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Headers de sécurité
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload";

    # Limites des uploads
    client_max_body_size 100M;
    client_body_timeout 120s;
    client_header_timeout 120s;

    # API endpoints
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Servir les fichiers uploadés
    location /uploads/ {
        alias /opt/Only-Flick-Go/uploads/;
        expires 1y;
        add_header Cache-Control "public, immutable";
        
        # Sécurité pour les fichiers
        location ~* \.(php|jsp|asp|sh|bash)$ {
            deny all;
        }
    }

    # Health check
    location /health {
        proxy_pass http://127.0.0.1:8080/api/v1/health;
        access_log off;
    }

    # Logs
    access_log /var/log/nginx/onlyflick-access.log;
    error_log /var/log/nginx/onlyflick-error.log;
}
EOF

# Activer le site
sudo ln -s /etc/nginx/sites-available/onlyflick /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## Configuration SSL (Let's Encrypt)

### 1. Installation de Certbot
```bash
# Ubuntu/Debian
sudo apt install certbot python3-certbot-nginx

# CentOS/RHEL
sudo dnf install certbot python3-certbot-nginx
```

### 2. Obtenir le certificat
```bash
sudo certbot --nginx -d api.onlyflick.com
```

### 3. Renouvellement automatique
```bash
# Ajouter un cron job pour le renouvellement
echo "0 12 * * * /usr/bin/certbot renew --quiet" | sudo crontab -
```

## Monitoring et logs

### 1. Configuration des logs
```bash
# Créer les dossiers de logs
sudo mkdir -p /var/log/onlyflick
sudo chown www-data:www-data /var/log/onlyflick

# Rotation des logs
sudo cat > /etc/logrotate.d/onlyflick << EOF
/var/log/onlyflick/*.log {
    daily
    missingok
    rotate 52
    compress
    delaycompress
    notifempty
    create 644 www-data www-data
    postrotate
        systemctl reload onlyflick-api
    endscript
}
EOF
```

### 2. Monitoring avec systemd
```bash
# Surveiller l'état du service
sudo systemctl status onlyflick-api

# Voir les logs en temps réel
sudo journalctl -u onlyflick-api -f

# Vérifier les métriques
sudo systemctl show onlyflick-api --property=MemoryCurrent,CPUUsageNSec
```

## Sauvegarde

### 1. Script de sauvegarde de la base de données
```bash
sudo cat > /opt/scripts/backup-database.sh << EOF
#!/bin/bash

# Configuration
DB_NAME="only_flick_db"
DB_USER="only_flick_user"
BACKUP_DIR="/opt/backups/database"
DATE=\$(date +%Y%m%d_%H%M%S)

# Créer le dossier de sauvegarde
mkdir -p \$BACKUP_DIR

# Effectuer la sauvegarde
pg_dump -h localhost -U \$DB_USER \$DB_NAME | gzip > \$BACKUP_DIR/onlyflick_\$DATE.sql.gz

# Nettoyer les sauvegardes anciennes (garder 30 jours)
find \$BACKUP_DIR -name "onlyflick_*.sql.gz" -mtime +30 -delete

echo "Sauvegarde terminée: \$BACKUP_DIR/onlyflick_\$DATE.sql.gz"
EOF

chmod +x /opt/scripts/backup-database.sh

# Programmer la sauvegarde quotidienne
echo "0 2 * * * /opt/scripts/backup-database.sh" | sudo crontab -
```

### 2. Sauvegarde des fichiers uploadés
```bash
sudo cat > /opt/scripts/backup-uploads.sh << EOF
#!/bin/bash

UPLOAD_DIR="/opt/Only-Flick-Go/uploads"
BACKUP_DIR="/opt/backups/uploads"
DATE=\$(date +%Y%m%d_%H%M%S)

mkdir -p \$BACKUP_DIR
tar -czf \$BACKUP_DIR/uploads_\$DATE.tar.gz -C \$UPLOAD_DIR .

# Garder 7 jours de sauvegardes
find \$BACKUP_DIR -name "uploads_*.tar.gz" -mtime +7 -delete

echo "Sauvegarde uploads terminée: \$BACKUP_DIR/uploads_\$DATE.tar.gz"
EOF

chmod +x /opt/scripts/backup-uploads.sh

# Programmer la sauvegarde quotidienne
echo "0 3 * * * /opt/scripts/backup-uploads.sh" | sudo crontab -
```

## Sécurité

### 1. Firewall
```bash
# UFW (Ubuntu)
sudo ufw enable
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw deny 8080/tcp  # Bloquer l'accès direct à l'API

# Firewalld (CentOS)
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

### 2. Fail2ban
```bash
# Installation
sudo apt install fail2ban  # Ubuntu
sudo dnf install fail2ban  # CentOS

# Configuration
sudo cat > /etc/fail2ban/jail.local << EOF
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5

[nginx-http-auth]
enabled = true
port = http,https
logpath = /var/log/nginx/onlyflick-error.log

[nginx-noscript]
enabled = true
port = http,https
logpath = /var/log/nginx/onlyflick-access.log
maxretry = 6

[nginx-badbots]
enabled = true
port = http,https
logpath = /var/log/nginx/onlyflick-access.log
maxretry = 2
EOF

sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

## Mise à jour et maintenance

### 1. Script de déploiement
```bash
sudo cat > /opt/scripts/deploy.sh << EOF
#!/bin/bash

set -e

echo "Début du déploiement OnlyFlick..."

# Naviguer vers le dossier de l'app
cd /opt/Only-Flick-Go

# Arrêter le service
sudo systemctl stop onlyflick-api

# Sauvegarder la version actuelle
cp cmd/api/onlyflick-api cmd/api/onlyflick-api.backup

# Mettre à jour le code
git pull origin main

# Installer les dépendances
go mod download

# Compiler la nouvelle version
cd cmd/api
go build -o onlyflick-api .

# Exécuter les nouvelles migrations (si nécessaire)
# psql -h localhost -U only_flick_user -d only_flick_db -f internal/database/migrations/new_migration.sql

# Redémarrer le service
sudo systemctl start onlyflick-api

# Vérifier que le service fonctionne
sleep 5
if sudo systemctl is-active --quiet onlyflick-api; then
    echo "✅ Déploiement réussi"
else
    echo "❌ Échec du déploiement, restauration..."
    cp cmd/api/onlyflick-api.backup cmd/api/onlyflick-api
    sudo systemctl start onlyflick-api
    exit 1
fi

echo "Déploiement terminé avec succès"
EOF

chmod +x /opt/scripts/deploy.sh
```

### 2. Surveillance de la santé
```bash
sudo cat > /opt/scripts/health-check.sh << EOF
#!/bin/bash

API_URL="https://api.onlyflick.com/api/v1/health"
SLACK_WEBHOOK="your-slack-webhook-url"

# Vérifier la santé de l'API
if ! curl -f -s \$API_URL > /dev/null; then
    echo "❌ API OnlyFlick down"
    
    # Envoyer une alerte Slack (optionnel)
    if [ -n "\$SLACK_WEBHOOK" ]; then
        curl -X POST -H 'Content-type: application/json' \
            --data '{"text":"🚨 API OnlyFlick est down!"}' \
            \$SLACK_WEBHOOK
    fi
    
    # Redémarrer le service
    sudo systemctl restart onlyflick-api
    
    exit 1
else
    echo "✅ API OnlyFlick healthy"
fi
EOF

chmod +x /opt/scripts/health-check.sh

# Programmer la vérification toutes les 5 minutes
echo "*/5 * * * * /opt/scripts/health-check.sh" | sudo crontab -
```

## Optimisations de performance

### 1. Configuration PostgreSQL
```bash
# Éditer la configuration PostgreSQL
sudo nano /etc/postgresql/13/main/postgresql.conf

# Optimisations recommandées
# shared_buffers = 256MB
# effective_cache_size = 1GB
# work_mem = 4MB
# maintenance_work_mem = 64MB
# max_connections = 100
# checkpoint_completion_target = 0.9

# Redémarrer PostgreSQL
sudo systemctl restart postgresql
```

### 2. Mise en cache avec Redis (optionnel)
```bash
# Installation de Redis
sudo apt install redis-server

# Configuration de base
sudo nano /etc/redis/redis.conf
# maxmemory 256mb
# maxmemory-policy allkeys-lru

sudo systemctl enable redis-server
sudo systemctl start redis-server
```

Ce guide couvre tous les aspects essentiels du déploiement de l'API OnlyFlick en production, de l'installation à la maintenance continue.
