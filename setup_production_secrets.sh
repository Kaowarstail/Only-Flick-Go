#!/bin/bash

# =============================================================================
# SCRIPT DE CONFIGURATION SÉCURISÉ POUR LA PRODUCTION
# =============================================================================
# Ce script utilise Docker Secrets pour gérer les secrets de manière sécurisée
# Usage: ./setup_production_secrets.sh
# =============================================================================

set -e

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DOCKER_SECRETS_DIR="/var/lib/docker/secrets"
SECRETS_DIR="./secrets"
DOCKER_COMPOSE_FILE="docker-compose.monitoring.prod.yml"

echo -e "${BLUE}===============================================${NC}"
echo -e "${BLUE}    CONFIGURATION SÉCURISÉE POUR LA PRODUCTION${NC}"
echo -e "${BLUE}===============================================${NC}"
echo

# Vérifier si Docker Swarm est initialisé
if ! docker info --format '{{.Swarm.LocalNodeState}}' | grep -q "active"; then
    echo -e "${YELLOW}Docker Swarm n'est pas initialisé. Initialisation...${NC}"
    docker swarm init
    echo -e "${GREEN}✅ Docker Swarm initialisé${NC}"
fi

# Créer le répertoire des secrets locaux
mkdir -p "$SECRETS_DIR"

# Fonction pour créer un secret Docker
create_docker_secret() {
    local secret_name="$1"
    local secret_value="$2"
    
    # Vérifier si le secret existe déjà
    if docker secret ls --format "{{.Name}}" | grep -q "^${secret_name}$"; then
        echo -e "${YELLOW}Secret ${secret_name} existe déjà, suppression...${NC}"
        docker secret rm "$secret_name" || true
    fi
    
    # Créer le secret
    echo "$secret_value" | docker secret create "$secret_name" -
    echo -e "${GREEN}✅ Secret ${secret_name} créé${NC}"
}

# Fonction pour demander un secret à l'utilisateur
prompt_secret() {
    local secret_name="$1"
    local prompt_text="$2"
    local default_value="$3"
    
    echo -e "${BLUE}${prompt_text}${NC}"
    if [ -n "$default_value" ]; then
        echo -e "${YELLOW}Appuyez sur Entrée pour utiliser la valeur par défaut: ${default_value}${NC}"
    fi
    
    read -sp "Valeur: " secret_value
    echo
    
    if [ -z "$secret_value" ] && [ -n "$default_value" ]; then
        secret_value="$default_value"
    fi
    
    if [ -z "$secret_value" ]; then
        echo -e "${RED}❌ Valeur requise pour ${secret_name}${NC}"
        exit 1
    fi
    
    create_docker_secret "$secret_name" "$secret_value"
}

# Fonction pour créer un secret depuis un fichier
create_secret_from_file() {
    local secret_name="$1"
    local file_path="$2"
    
    if [ -f "$file_path" ]; then
        echo -e "${BLUE}Création du secret ${secret_name} depuis ${file_path}${NC}"
        docker secret create "$secret_name" "$file_path"
        echo -e "${GREEN}✅ Secret ${secret_name} créé depuis le fichier${NC}"
    else
        echo -e "${RED}❌ Fichier ${file_path} introuvable${NC}"
        exit 1
    fi
}

echo -e "${YELLOW}Configuration des secrets Docker...${NC}"
echo

# Secrets de base de données
echo -e "${BLUE}=== CONFIGURATION BASE DE DONNÉES ===${NC}"
prompt_secret "db_host" "Hôte de la base de données:"
prompt_secret "db_user" "Utilisateur de la base de données:"
prompt_secret "db_password" "Mot de passe de la base de données:"
prompt_secret "db_name" "Nom de la base de données:"
prompt_secret "database_url" "URL complète de la base de données (postgresql://...):"

# Secrets JWT
echo -e "${BLUE}=== CONFIGURATION JWT ===${NC}"
prompt_secret "jwt_secret" "Clé secrète JWT (générez une clé forte):"

# Secrets Stripe
echo -e "${BLUE}=== CONFIGURATION STRIPE ===${NC}"
prompt_secret "stripe_secret_key" "Clé secrète Stripe:"
prompt_secret "stripe_publishable_key" "Clé publique Stripe:"
prompt_secret "stripe_webhook_secret" "Secret webhook Stripe:"

# Secrets Cloudinary
echo -e "${BLUE}=== CONFIGURATION CLOUDINARY ===${NC}"
prompt_secret "cloudinary_cloud_name" "Nom du cloud Cloudinary:"
prompt_secret "cloudinary_api_key" "Clé API Cloudinary:"
prompt_secret "cloudinary_api_secret" "Secret API Cloudinary:"

# Secrets pour le monitoring
echo -e "${BLUE}=== CONFIGURATION MONITORING ===${NC}"
prompt_secret "grafana_admin_password" "Mot de passe admin Grafana:"
prompt_secret "grafana_db_password" "Mot de passe DB Grafana:"
prompt_secret "smtp_password" "Mot de passe SMTP pour les alertes:"

# Créer le fichier docker-compose de production avec secrets
echo -e "${BLUE}Création du fichier docker-compose avec secrets...${NC}"

cat > docker-compose.production.yml << 'EOF'
version: '3.8'

services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - DB_HOST_FILE=/run/secrets/db_host
      - DB_USER_FILE=/run/secrets/db_user
      - DB_PASSWORD_FILE=/run/secrets/db_password
      - DB_NAME_FILE=/run/secrets/db_name
      - DATABASE_URL_FILE=/run/secrets/database_url
      - JWT_SECRET_FILE=/run/secrets/jwt_secret
      - STRIPE_SECRET_KEY_FILE=/run/secrets/stripe_secret_key
      - STRIPE_PUBLISHABLE_KEY_FILE=/run/secrets/stripe_publishable_key
      - STRIPE_WEBHOOK_SECRET_FILE=/run/secrets/stripe_webhook_secret
      - CLOUDINARY_CLOUD_NAME_FILE=/run/secrets/cloudinary_cloud_name
      - CLOUDINARY_API_KEY_FILE=/run/secrets/cloudinary_api_key
      - CLOUDINARY_API_SECRET_FILE=/run/secrets/cloudinary_api_secret
    secrets:
      - db_host
      - db_user
      - db_password
      - db_name
      - database_url
      - jwt_secret
      - stripe_secret_key
      - stripe_publishable_key
      - stripe_webhook_secret
      - cloudinary_cloud_name
      - cloudinary_api_key
      - cloudinary_api_secret
    networks:
      - onlyflick-network
    deploy:
      replicas: 2
      restart_policy:
        condition: on-failure
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 256M

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus/prometheus.prod.yml:/etc/prometheus/prometheus.yml
      - ./monitoring/prometheus/alert.rules.yml:/etc/prometheus/alert.rules.yml
      - prometheus_data:/prometheus
    networks:
      - onlyflick-network

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD__FILE=/run/secrets/grafana_admin_password
      - GF_DATABASE_TYPE=postgres
      - GF_DATABASE_HOST=grafana-db:5432
      - GF_DATABASE_NAME=grafana
      - GF_DATABASE_USER=grafana
      - GF_DATABASE_PASSWORD__FILE=/run/secrets/grafana_db_password
      - GF_SMTP_ENABLED=true
      - GF_SMTP_HOST=smtp.gmail.com:587
      - GF_SMTP_USER=your-email@gmail.com
      - GF_SMTP_PASSWORD__FILE=/run/secrets/smtp_password
    secrets:
      - grafana_admin_password
      - grafana_db_password
      - smtp_password
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/provisioning:/etc/grafana/provisioning
      - ./monitoring/grafana/dashboards:/var/lib/grafana/dashboards
    networks:
      - onlyflick-network

  grafana-db:
    image: postgres:13
    environment:
      - POSTGRES_DB=grafana
      - POSTGRES_USER=grafana
      - POSTGRES_PASSWORD_FILE=/run/secrets/grafana_db_password
    secrets:
      - grafana_db_password
    volumes:
      - grafana_db_data:/var/lib/postgresql/data
    networks:
      - onlyflick-network

  alertmanager:
    image: prom/alertmanager:latest
    ports:
      - "9093:9093"
    volumes:
      - ./monitoring/alertmanager/alertmanager.prod.yml:/etc/alertmanager/alertmanager.yml
      - alertmanager_data:/alertmanager
    networks:
      - onlyflick-network

  node-exporter:
    image: prom/node-exporter:latest
    ports:
      - "9100:9100"
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/rootfs:ro
    command:
      - '--path.procfs=/host/proc'
      - '--path.sysfs=/host/sys'
      - '--collector.filesystem.ignored-mount-points=^/(sys|proc|dev|host|etc)($$|/)'
    networks:
      - onlyflick-network

secrets:
  db_host:
    external: true
  db_user:
    external: true
  db_password:
    external: true
  db_name:
    external: true
  database_url:
    external: true
  jwt_secret:
    external: true
  stripe_secret_key:
    external: true
  stripe_publishable_key:
    external: true
  stripe_webhook_secret:
    external: true
  cloudinary_cloud_name:
    external: true
  cloudinary_api_key:
    external: true
  cloudinary_api_secret:
    external: true
  grafana_admin_password:
    external: true
  grafana_db_password:
    external: true
  smtp_password:
    external: true

volumes:
  prometheus_data:
  grafana_data:
  grafana_db_data:
  alertmanager_data:

networks:
  onlyflick-network:
    driver: overlay
EOF

echo -e "${GREEN}✅ Fichier docker-compose.production.yml créé${NC}"

# Créer un helper pour lire les secrets
cat > internal/utils/secrets.go << 'EOF'
package utils

import (
	"io/ioutil"
	"os"
	"strings"
)

// ReadSecret lit un secret depuis un fichier ou une variable d'environnement
func ReadSecret(envVar string) (string, error) {
	// Vérifier si une variable d'environnement avec suffixe _FILE existe
	if fileVar := os.Getenv(envVar + "_FILE"); fileVar != "" {
		// Lire depuis le fichier
		content, err := ioutil.ReadFile(fileVar)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(content)), nil
	}
	
	// Sinon, lire depuis la variable d'environnement normale
	return os.Getenv(envVar), nil
}
EOF

echo -e "${GREEN}✅ Helper pour lire les secrets créé${NC}"

# Créer un script de déploiement
cat > deploy_production.sh << 'EOF'
#!/bin/bash

# Script de déploiement en production avec Docker Swarm
set -e

echo "🚀 Déploiement en production..."

# Vérifier que Docker Swarm est actif
if ! docker info --format '{{.Swarm.LocalNodeState}}' | grep -q "active"; then
    echo "❌ Docker Swarm n'est pas actif. Exécutez d'abord setup_production_secrets.sh"
    exit 1
fi

# Construire l'image
echo "🔨 Construction de l'image..."
docker build -t onlyflick-api:latest .

# Déployer la stack
echo "📦 Déploiement de la stack..."
docker stack deploy -c docker-compose.production.yml onlyflick

echo "✅ Déploiement terminé!"
echo "🌐 API accessible sur: http://localhost:8080"
echo "📊 Grafana accessible sur: http://localhost:3000"
echo "🔥 Prometheus accessible sur: http://localhost:9090"
echo "🚨 Alertmanager accessible sur: http://localhost:9093"
EOF

chmod +x deploy_production.sh
echo -e "${GREEN}✅ Script de déploiement créé${NC}"

# Créer un script pour vérifier les secrets
cat > check_secrets.sh << 'EOF'
#!/bin/bash

echo "🔍 Vérification des secrets Docker..."

secrets=(
    "db_host"
    "db_user"
    "db_password"
    "db_name"
    "database_url"
    "jwt_secret"
    "stripe_secret_key"
    "stripe_publishable_key"
    "stripe_webhook_secret"
    "cloudinary_cloud_name"
    "cloudinary_api_key"
    "cloudinary_api_secret"
    "grafana_admin_password"
    "grafana_db_password"
    "smtp_password"
)

missing_secrets=()

for secret in "${secrets[@]}"; do
    if docker secret ls --format "{{.Name}}" | grep -q "^${secret}$"; then
        echo "✅ $secret"
    else
        echo "❌ $secret"
        missing_secrets+=("$secret")
    fi
done

if [ ${#missing_secrets[@]} -eq 0 ]; then
    echo "🎉 Tous les secrets sont configurés!"
else
    echo "⚠️  Secrets manquants: ${missing_secrets[*]}"
    echo "Exécutez setup_production_secrets.sh pour les créer"
fi
EOF

chmod +x check_secrets.sh
echo -e "${GREEN}✅ Script de vérification des secrets créé${NC}"

echo
echo -e "${GREEN}===============================================${NC}"
echo -e "${GREEN}    CONFIGURATION TERMINÉE AVEC SUCCÈS!${NC}"
echo -e "${GREEN}===============================================${NC}"
echo
echo -e "${YELLOW}Prochaines étapes:${NC}"
echo "1. Vérifiez vos secrets avec: ./check_secrets.sh"
echo "2. Adaptez le code pour utiliser ReadSecret() au lieu des variables d'environnement"
echo "3. Déployez en production avec: ./deploy_production.sh"
echo
echo -e "${RED}IMPORTANT:${NC}"
echo "- Gardez le fichier .env.secure en lieu sûr (sauvegarde hors ligne)"
echo "- Ne commitez jamais les vrais secrets dans Git"
echo "- Utilisez des mots de passe forts pour la production"
echo "- Configurez un reverse proxy (nginx/traefik) pour SSL"
echo
