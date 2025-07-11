#!/bin/bash

# Script de déploiement sécurisé pour OnlyFlick Monitoring
# Usage: ./deploy_monitoring_secure.sh [deploy|update|stop|backup|restore]

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_NAME="onlyflick_monitoring"
BACKUP_DIR="$SCRIPT_DIR/monitoring_backups"
COMPOSE_FILE="docker-compose.monitoring.prod.yml"
LOG_FILE="$SCRIPT_DIR/deploy_monitoring.log"

# Couleurs pour les logs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Mode de déploiement
DEPLOYMENT_MODE="production"
VALIDATE_SECRETS=true
BACKUP_BEFORE_DEPLOY=true
SEND_NOTIFICATIONS=true

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

# Fonction pour nettoyer les fichiers sensibles
cleanup_sensitive_files() {
    log_info "Nettoyage des fichiers sensibles..."
    
    # Supprimer les fichiers temporaires avec secrets
    rm -f .env.monitoring.temp
    rm -f .env.generated
    rm -f .env.monitoring.local
    rm -f .env.monitoring.dev
    rm -f .env.monitoring.staging
    rm -f .env.monitoring.test
    rm -f .env.monitoring.backup
    
    # Supprimer les logs sensibles
    find . -name "*.log" -path "*/logs/*" -exec rm -f {} \; 2>/dev/null || true
    
    # Nettoyer les fichiers temporaires Docker
    docker system prune -f &>/dev/null || true
    
    log_success "Nettoyage terminé"
}

# Fonction pour valider les secrets
validate_secrets() {
    log_info "Validation des secrets..."
    
    # Vérifier que le script de validation existe
    if [ ! -f "$SCRIPT_DIR/validate_secrets.sh" ]; then
        log_error "Script de validation des secrets non trouvé"
        return 1
    fi
    
    # Rendre le script exécutable
    chmod +x "$SCRIPT_DIR/validate_secrets.sh"
    
    # Valider selon le mode de déploiement
    if [ "$DEPLOYMENT_MODE" = "github-actions" ]; then
        "$SCRIPT_DIR/validate_secrets.sh" --github-actions
    elif [ "$DEPLOYMENT_MODE" = "docker-secrets" ]; then
        "$SCRIPT_DIR/validate_secrets.sh" --docker-secrets
    elif [ "$DEPLOYMENT_MODE" = "kubernetes" ]; then
        "$SCRIPT_DIR/validate_secrets.sh" --kubernetes
    else
        # Mode développement/local
        "$SCRIPT_DIR/validate_secrets.sh" --verbose
    fi
    
    local result=$?
    
    if [ $result -eq 0 ]; then
        log_success "Validation des secrets réussie"
        return 0
    else
        log_error "Validation des secrets échouée"
        return 1
    fi
}

# Fonction pour créer un environnement sécurisé
create_secure_environment() {
    log_info "Création de l'environnement sécurisé..."
    
    local env_file=".env.monitoring.temp"
    
    # Vérifier si on est dans GitHub Actions
    if [ -n "$GITHUB_ACTIONS" ]; then
        log_info "Déploiement via GitHub Actions détecté"
        DEPLOYMENT_MODE="github-actions"
        
        # Les secrets sont injectés via GitHub Actions
        # Le fichier .env.monitoring.prod est déjà créé par le workflow
        if [ -f ".env.monitoring.prod" ]; then
            cp ".env.monitoring.prod" "$env_file"
        else
            log_error "Fichier d'environnement non créé par GitHub Actions"
            return 1
        fi
    else
        log_info "Déploiement local détecté"
        
        # Vérifier si des secrets sont disponibles dans l'environnement
        if [ -n "$GRAFANA_ADMIN_PASSWORD" ] && [ -n "$GRAFANA_SECRET_KEY" ]; then
            log_info "Utilisation des variables d'environnement"
            
            # Créer le fichier d'environnement depuis les variables
            cat > "$env_file" << EOF
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD}
GRAFANA_SECRET_KEY=${GRAFANA_SECRET_KEY}
GRAFANA_DB_NAME=grafana
GRAFANA_DB_USER=grafana
GRAFANA_DB_PASSWORD=${GRAFANA_DB_PASSWORD}
SMTP_HOST=${SMTP_HOST:-smtp.gmail.com}
SMTP_PORT=587
SMTP_FROM=${SMTP_FROM:-monitoring@yourdomain.com}
SMTP_USERNAME=${SMTP_USERNAME:-$SMTP_FROM}
SMTP_PASSWORD=${SMTP_PASSWORD}
ADMIN_EMAIL=${ADMIN_EMAIL:-admin@yourdomain.com}
SLACK_WEBHOOK_URL=${SLACK_WEBHOOK_URL}
WEBHOOK_TOKEN=${WEBHOOK_TOKEN}
LETSENCRYPT_EMAIL=${LETSENCRYPT_EMAIL:-$ADMIN_EMAIL}
DOMAIN=${DOMAIN:-yourdomain.com}
MONITORING_DOMAIN=${MONITORING_DOMAIN:-monitoring.yourdomain.com}
API_HOST=${API_HOST:-api.yourdomain.com}
API_PORT=8080
EOF
        elif [ -f ".env.generated" ]; then
            log_info "Utilisation des secrets générés"
            cp ".env.generated" "$env_file"
        else
            log_error "Aucun secret disponible. Exécutez ./validate_secrets.sh --generate-missing"
            return 1
        fi
    fi
    
    # Vérifier que le fichier d'environnement est valide
    if [ ! -f "$env_file" ]; then
        log_error "Impossible de créer le fichier d'environnement"
        return 1
    fi
    
    # Vérifier qu'il n'y a pas de placeholders dangereux
    if grep -q "CHANGE_ME\|{{.*}}" "$env_file"; then
        log_error "Placeholders non résolus dans le fichier d'environnement"
        cat "$env_file" | grep -E "CHANGE_ME|{{.*}}" || true
        return 1
    fi
    
    # Définir les permissions correctes
    chmod 600 "$env_file"
    
    log_success "Environnement sécurisé créé"
    return 0
}

# Fonction pour vérifier les prérequis
check_prerequisites() {
    log_info "Vérification des prérequis..."
    
    # Vérifier Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker n'est pas installé"
        return 1
    fi
    
    # Vérifier Docker Compose
    if ! docker compose version &> /dev/null; then
        log_error "Docker Compose n'est pas disponible"
        return 1
    fi
    
    # Vérifier que Docker est en cours d'exécution
    if ! docker info &> /dev/null; then
        log_error "Docker n'est pas en cours d'exécution"
        return 1
    fi
    
    # Vérifier les fichiers de configuration
    if [ ! -f "$COMPOSE_FILE" ]; then
        log_error "Fichier Docker Compose non trouvé: $COMPOSE_FILE"
        return 1
    fi
    
    # Vérifier la configuration des services
    if ! docker compose -f "$COMPOSE_FILE" config &> /dev/null; then
        log_error "Configuration Docker Compose invalide"
        return 1
    fi
    
    log_success "Prérequis vérifiés"
    return 0
}

# Fonction pour créer une sauvegarde
create_backup() {
    log_info "Création d'une sauvegarde..."
    
    # Créer le répertoire de sauvegarde
    mkdir -p "$BACKUP_DIR"
    
    local backup_name="backup_$(date +%Y%m%d_%H%M%S)"
    local backup_path="$BACKUP_DIR/$backup_name"
    
    # Créer le répertoire de sauvegarde
    mkdir -p "$backup_path"
    
    # Sauvegarder les données Grafana
    if docker volume ls | grep -q "grafana-data"; then
        log_info "Sauvegarde des données Grafana..."
        docker run --rm \
            -v grafana-data:/data \
            -v "$backup_path:/backup" \
            alpine:latest \
            tar czf /backup/grafana-data.tar.gz -C /data .
    fi
    
    # Sauvegarder les données Prometheus
    if docker volume ls | grep -q "prometheus-data"; then
        log_info "Sauvegarde des données Prometheus..."
        docker run --rm \
            -v prometheus-data:/data \
            -v "$backup_path:/backup" \
            alpine:latest \
            tar czf /backup/prometheus-data.tar.gz -C /data .
    fi
    
    # Sauvegarder les configurations
    log_info "Sauvegarde des configurations..."
    tar czf "$backup_path/configurations.tar.gz" \
        monitoring/ \
        "$COMPOSE_FILE" \
        docker-compose.monitoring.yml \
        *.md \
        *.sh \
        --exclude="*.log" \
        --exclude="*.tmp" \
        --exclude="data/" \
        --exclude="logs/" 2>/dev/null || true
    
    # Créer un fichier de métadonnées
    cat > "$backup_path/metadata.json" << EOF
{
    "backup_name": "$backup_name",
    "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
    "version": "$(git rev-parse HEAD 2>/dev/null || echo 'unknown')",
    "hostname": "$(hostname)",
    "deployment_mode": "$DEPLOYMENT_MODE"
}
EOF
    
    log_success "Sauvegarde créée: $backup_path"
    echo "$backup_path"
}

# Fonction pour déployer les services
deploy_services() {
    log_info "Déploiement des services de monitoring..."
    
    # Utiliser le fichier d'environnement temporaire
    local env_file=".env.monitoring.temp"
    
    if [ ! -f "$env_file" ]; then
        log_error "Fichier d'environnement non trouvé"
        return 1
    fi
    
    # Déployer avec Docker Compose
    docker compose -f "$COMPOSE_FILE" --env-file "$env_file" up -d
    
    local result=$?
    
    if [ $result -eq 0 ]; then
        log_success "Services déployés avec succès"
        return 0
    else
        log_error "Échec du déploiement des services"
        return 1
    fi
}

# Fonction pour vérifier l'état des services
check_services_health() {
    log_info "Vérification de l'état des services..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        log_info "Tentative $attempt/$max_attempts"
        
        # Vérifier Prometheus
        if curl -f -s "http://localhost:9090/api/v1/query?query=up" > /dev/null 2>&1; then
            log_success "Prometheus: ✅ Opérationnel"
        else
            log_warning "Prometheus: ⚠️ Non accessible"
        fi
        
        # Vérifier Grafana
        if curl -f -s "http://localhost:3000/api/health" > /dev/null 2>&1; then
            log_success "Grafana: ✅ Opérationnel"
        else
            log_warning "Grafana: ⚠️ Non accessible"
        fi
        
        # Vérifier Alertmanager
        if curl -f -s "http://localhost:9093/api/v1/status" > /dev/null 2>&1; then
            log_success "Alertmanager: ✅ Opérationnel"
        else
            log_warning "Alertmanager: ⚠️ Non accessible"
        fi
        
        # Vérifier que tous les services sont prêts
        if docker compose -f "$COMPOSE_FILE" ps | grep -q "Up"; then
            log_success "Tous les services sont démarrés"
            sleep 10
            return 0
        fi
        
        sleep 10
        ((attempt++))
    done
    
    log_error "Timeout - Les services ne sont pas tous opérationnels"
    return 1
}

# Fonction pour envoyer des notifications
send_notification() {
    local status=$1
    local message=$2
    
    if [ "$SEND_NOTIFICATIONS" != "true" ]; then
        return 0
    fi
    
    log_info "Envoi de notification: $message"
    
    # Notification Slack
    if [ -n "$SLACK_WEBHOOK_URL" ]; then
        local color="good"
        local emoji="✅"
        
        if [ "$status" != "success" ]; then
            color="danger"
            emoji="❌"
        fi
        
        curl -X POST "$SLACK_WEBHOOK_URL" \
            -H 'Content-type: application/json' \
            --data "{
                \"text\": \"$emoji OnlyFlick Monitoring - $message\",
                \"attachments\": [{
                    \"color\": \"$color\",
                    \"fields\": [{
                        \"title\": \"Environnement\",
                        \"value\": \"$DEPLOYMENT_MODE\",
                        \"short\": true
                    }, {
                        \"title\": \"Timestamp\",
                        \"value\": \"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\",
                        \"short\": true
                    }]
                }]
            }" \
            --silent --output /dev/null || true
    fi
    
    # Notification email (si configuré)
    if [ -n "$ADMIN_EMAIL" ] && [ -n "$SMTP_HOST" ]; then
        echo "Sujet: OnlyFlick Monitoring - $message" | \
        mail -s "OnlyFlick Monitoring - $message" "$ADMIN_EMAIL" 2>/dev/null || true
    fi
}

# Fonction principale de déploiement
deploy() {
    log_info "🚀 Démarrage du déploiement sécurisé OnlyFlick Monitoring"
    
    # Nettoyer les fichiers sensibles précédents
    cleanup_sensitive_files
    
    # Vérifier les prérequis
    if ! check_prerequisites; then
        send_notification "error" "Échec des prérequis"
        return 1
    fi
    
    # Valider les secrets
    if [ "$VALIDATE_SECRETS" = "true" ]; then
        if ! validate_secrets; then
            send_notification "error" "Validation des secrets échouée"
            return 1
        fi
    fi
    
    # Créer l'environnement sécurisé
    if ! create_secure_environment; then
        send_notification "error" "Création de l'environnement échouée"
        return 1
    fi
    
    # Créer une sauvegarde
    if [ "$BACKUP_BEFORE_DEPLOY" = "true" ]; then
        backup_path=$(create_backup)
        log_info "Sauvegarde créée: $backup_path"
    fi
    
    # Déployer les services
    if ! deploy_services; then
        send_notification "error" "Déploiement des services échoué"
        return 1
    fi
    
    # Vérifier l'état des services
    if ! check_services_health; then
        send_notification "error" "Vérification de santé échouée"
        return 1
    fi
    
    # Nettoyer les fichiers sensibles
    cleanup_sensitive_files
    
    # Envoyer notification de succès
    send_notification "success" "Déploiement réussi"
    
    log_success "🎉 Déploiement terminé avec succès"
    
    # Afficher les informations d'accès
    echo ""
    echo "📊 Services de monitoring disponibles:"
    echo "- Grafana: http://localhost:3000"
    echo "- Prometheus: http://localhost:9090"
    echo "- Alertmanager: http://localhost:9093"
    echo ""
    echo "🔧 Commandes utiles:"
    echo "- Voir les logs: docker compose -f $COMPOSE_FILE logs -f"
    echo "- Arrêter les services: docker compose -f $COMPOSE_FILE down"
    echo "- Redémarrer: ./deploy_monitoring_secure.sh update"
    echo ""
    
    return 0
}

# Fonction pour mettre à jour
update() {
    log_info "🔄 Mise à jour du monitoring OnlyFlick"
    
    # Créer une sauvegarde avant la mise à jour
    if [ "$BACKUP_BEFORE_DEPLOY" = "true" ]; then
        backup_path=$(create_backup)
        log_info "Sauvegarde créée: $backup_path"
    fi
    
    # Redéployer
    deploy
}

# Fonction pour arrêter
stop() {
    log_info "🛑 Arrêt du monitoring OnlyFlick"
    
    # Arrêter les services
    docker compose -f "$COMPOSE_FILE" down
    
    # Nettoyer les fichiers sensibles
    cleanup_sensitive_files
    
    # Envoyer notification
    send_notification "info" "Services arrêtés"
    
    log_success "Services arrêtés"
}

# Fonction pour afficher l'aide
show_help() {
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  deploy              Déployer le monitoring"
    echo "  update              Mettre à jour le monitoring"
    echo "  stop                Arrêter le monitoring"
    echo "  backup              Créer une sauvegarde"
    echo "  validate-secrets    Valider les secrets"
    echo "  cleanup             Nettoyer les fichiers sensibles"
    echo "  help                Afficher cette aide"
    echo ""
    echo "Options:"
    echo "  --no-backup         Ne pas créer de sauvegarde"
    echo "  --no-validation     Ne pas valider les secrets"
    echo "  --no-notifications  Ne pas envoyer de notifications"
    echo "  --mode MODE         Mode de déploiement (production, github-actions, docker-secrets, kubernetes)"
    echo ""
    echo "Examples:"
    echo "  $0 deploy"
    echo "  $0 update --no-backup"
    echo "  $0 validate-secrets"
    echo "  $0 deploy --mode github-actions"
    echo ""
}

# Parse des arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --no-backup)
            BACKUP_BEFORE_DEPLOY=false
            shift
            ;;
        --no-validation)
            VALIDATE_SECRETS=false
            shift
            ;;
        --no-notifications)
            SEND_NOTIFICATIONS=false
            shift
            ;;
        --mode)
            DEPLOYMENT_MODE="$2"
            shift 2
            ;;
        deploy|update|stop|backup|validate-secrets|cleanup|help)
            COMMAND="$1"
            shift
            ;;
        *)
            log_error "Option inconnue: $1"
            show_help
            exit 1
            ;;
    esac
done

# Commande par défaut
COMMAND="${COMMAND:-deploy}"

# Créer le répertoire de logs
mkdir -p "$(dirname "$LOG_FILE")"

# Exécuter la commande
case $COMMAND in
    deploy)
        deploy
        ;;
    update)
        update
        ;;
    stop)
        stop
        ;;
    backup)
        create_backup
        ;;
    validate-secrets)
        validate_secrets
        ;;
    cleanup)
        cleanup_sensitive_files
        ;;
    help)
        show_help
        ;;
    *)
        log_error "Commande inconnue: $COMMAND"
        show_help
        exit 1
        ;;
esac
