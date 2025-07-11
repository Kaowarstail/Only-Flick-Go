#!/bin/bash

# Script de démarrage du monitoring OnlyFlick
# Usage: ./start_monitoring.sh [start|stop|restart|status]

set -e

# Couleurs pour les logs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.monitoring.yml"
PROJECT_NAME="onlyflick_monitoring"

# Fonctions utilitaires
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Vérifier si Docker est installé
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker n'est pas installé. Veuillez installer Docker."
        exit 1
    fi
    
    if ! docker compose version &> /dev/null; then
        log_error "Docker Compose n'est pas disponible. Veuillez installer Docker Compose."
        exit 1
    fi
}

# Vérifier si le fichier de configuration existe
check_config() {
    if [ ! -f "$COMPOSE_FILE" ]; then
        log_error "Le fichier $COMPOSE_FILE n'existe pas."
        exit 1
    fi
}

# Démarrer les services de monitoring
start_monitoring() {
    log_info "Démarrage du monitoring OnlyFlick..."
    
    # Créer les dossiers nécessaires s'ils n'existent pas
    mkdir -p monitoring/grafana/dashboards
    mkdir -p monitoring/prometheus
    mkdir -p monitoring/alertmanager
    
    # Démarrer les services
    docker compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d
    
    # Attendre que les services soient prêts
    log_info "Attente du démarrage des services..."
    sleep 10
    
    # Vérifier le statut des services
    if docker compose -f $COMPOSE_FILE -p $PROJECT_NAME ps | grep -q "Up"; then
        log_success "Services de monitoring démarrés avec succès!"
        show_access_info
    else
        log_error "Erreur lors du démarrage des services."
        exit 1
    fi
}

# Arrêter les services de monitoring
stop_monitoring() {
    log_info "Arrêt du monitoring OnlyFlick..."
    docker compose -f $COMPOSE_FILE -p $PROJECT_NAME down
    log_success "Services de monitoring arrêtés."
}

# Redémarrer les services
restart_monitoring() {
    log_info "Redémarrage du monitoring OnlyFlick..."
    stop_monitoring
    sleep 5
    start_monitoring
}

# Afficher le statut des services
show_status() {
    log_info "Statut des services de monitoring:"
    docker compose -f $COMPOSE_FILE -p $PROJECT_NAME ps
}

# Afficher les informations d'accès
show_access_info() {
    echo ""
    log_success "=== INFORMATIONS D'ACCÈS ==="
    echo ""
    echo "📊 Grafana (Dashboards):"
    echo "   URL: http://localhost:3000"
    echo "   Login: admin"
    echo "   Mot de passe: admin123"
    echo ""
    echo "📈 Prometheus (Métriques):"
    echo "   URL: http://localhost:9090"
    echo ""
    echo "🚨 Alertmanager (Alertes):"
    echo "   URL: http://localhost:9093"
    echo ""
    echo "🖥️  Node Exporter (Métriques système):"
    echo "   URL: http://localhost:9100"
    echo ""
    echo "🔍 Dashboards disponibles dans Grafana:"
    echo "   - OnlyFlick API - Monitoring Principal"
    echo "   - OnlyFlick - Métriques Business"
    echo "   - OnlyFlick - Métriques Infrastructure"
    echo ""
    log_info "Assurez-vous que votre API OnlyFlick est démarrée sur le port 8080"
    log_info "pour que les métriques soient collectées correctement."
}

# Afficher les logs
show_logs() {
    log_info "Logs des services de monitoring:"
    docker compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f
}

# Nettoyer les volumes (attention: supprime les données)
cleanup() {
    log_warning "Cette action va supprimer toutes les données de monitoring."
    read -p "Êtes-vous sûr ? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Nettoyage en cours..."
        docker compose -f $COMPOSE_FILE -p $PROJECT_NAME down -v
        docker volume prune -f
        log_success "Nettoyage terminé."
    else
        log_info "Nettoyage annulé."
    fi
}

# Fonction principale
main() {
    case "${1:-start}" in
        start)
            check_docker
            check_config
            start_monitoring
            ;;
        stop)
            check_docker
            check_config
            stop_monitoring
            ;;
        restart)
            check_docker
            check_config
            restart_monitoring
            ;;
        status)
            check_docker
            check_config
            show_status
            ;;
        logs)
            check_docker
            check_config
            show_logs
            ;;
        cleanup)
            check_docker
            check_config
            cleanup
            ;;
        info)
            show_access_info
            ;;
        *)
            echo "Usage: $0 {start|stop|restart|status|logs|cleanup|info}"
            echo ""
            echo "Commandes:"
            echo "  start     - Démarrer le monitoring"
            echo "  stop      - Arrêter le monitoring"
            echo "  restart   - Redémarrer le monitoring"
            echo "  status    - Afficher le statut des services"
            echo "  logs      - Afficher les logs des services"
            echo "  cleanup   - Nettoyer les volumes (supprime les données)"
            echo "  info      - Afficher les informations d'accès"
            exit 1
            ;;
    esac
}

# Exécuter la fonction principale
main "$@"
