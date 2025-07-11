#!/bin/bash

# Script de test du monitoring OnlyFlick
# Ce script génère du trafic artificiel pour tester les métriques

set -e

# Configuration
API_URL="http://localhost:8080"
MONITORING_URL="http://localhost:3000"
PROMETHEUS_URL="http://localhost:9090"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# Vérifier si curl est installé
check_curl() {
    if ! command -v curl &> /dev/null; then
        log_error "curl n'est pas installé"
        exit 1
    fi
}

# Vérifier si l'API est accessible
check_api() {
    log_info "Vérification de l'API OnlyFlick..."
    if curl -s "$API_URL/health" > /dev/null; then
        log_success "API OnlyFlick accessible"
    else
        log_error "API OnlyFlick non accessible sur $API_URL"
        exit 1
    fi
}

# Vérifier si Prometheus est accessible
check_prometheus() {
    log_info "Vérification de Prometheus..."
    if curl -s "$PROMETHEUS_URL/-/healthy" > /dev/null; then
        log_success "Prometheus accessible"
    else
        log_error "Prometheus non accessible sur $PROMETHEUS_URL"
        exit 1
    fi
}

# Vérifier si Grafana est accessible
check_grafana() {
    log_info "Vérification de Grafana..."
    if curl -s "$MONITORING_URL/api/health" > /dev/null; then
        log_success "Grafana accessible"
    else
        log_error "Grafana non accessible sur $MONITORING_URL"
        exit 1
    fi
}

# Générer du trafic vers l'API
generate_traffic() {
    log_info "Génération de trafic vers l'API..."
    
    # Requêtes GET normales
    for i in {1..20}; do
        curl -s "$API_URL/health" > /dev/null
        curl -s "$API_URL/metrics" > /dev/null
        sleep 0.1
    done
    
    # Requêtes vers des endpoints qui pourraient ne pas exister (pour tester les erreurs 404)
    for i in {1..5}; do
        curl -s "$API_URL/nonexistent" > /dev/null 2>&1 || true
        sleep 0.1
    done
    
    log_success "Trafic généré (25 requêtes)"
}

# Vérifier les métriques dans Prometheus
check_metrics() {
    log_info "Vérification des métriques dans Prometheus..."
    
    # Vérifier quelques métriques importantes
    metrics=("http_requests_total" "http_request_duration_seconds" "go_goroutines" "go_memstats_heap_inuse_bytes")
    
    for metric in "${metrics[@]}"; do
        if curl -s "$PROMETHEUS_URL/api/v1/query?query=$metric" | grep -q "success"; then
            log_success "Métrique $metric disponible"
        else
            log_warning "Métrique $metric non trouvée"
        fi
        sleep 0.5
    done
}

# Afficher les informations d'accès
show_access_info() {
    echo ""
    log_success "=== INFORMATIONS D'ACCÈS ==="
    echo ""
    echo "📊 Grafana (Dashboards) : $MONITORING_URL"
    echo "   Login: admin / Mot de passe: admin123"
    echo ""
    echo "📈 Prometheus (Métriques) : $PROMETHEUS_URL"
    echo ""
    echo "🎯 API OnlyFlick : $API_URL"
    echo ""
    echo "📋 Dashboards disponibles :"
    echo "   - OnlyFlick - Vue d'Ensemble"
    echo "   - OnlyFlick API - Monitoring Principal"
    echo "   - OnlyFlick - Métriques Business"
    echo "   - OnlyFlick - Métriques Infrastructure"
    echo ""
}

# Test de charge (optionnel)
load_test() {
    log_info "Test de charge (30 secondes)..."
    
    # Lancer plusieurs requêtes en parallèle
    for i in {1..5}; do
        {
            for j in {1..20}; do
                curl -s "$API_URL/health" > /dev/null
                curl -s "$API_URL/metrics" > /dev/null
                sleep 0.1
            done
        } &
    done
    
    # Attendre la fin des processus
    wait
    
    log_success "Test de charge terminé"
}

# Vérifier les alertes
check_alerts() {
    log_info "Vérification des alertes..."
    
    # Vérifier si Alertmanager est accessible
    if curl -s "http://localhost:9093/api/v1/status" > /dev/null; then
        log_success "Alertmanager accessible"
        
        # Vérifier s'il y a des alertes actives
        alerts=$(curl -s "http://localhost:9093/api/v1/alerts" | grep -c '"state":"active"' || echo "0")
        if [ "$alerts" -gt 0 ]; then
            log_warning "$alerts alertes actives"
        else
            log_success "Aucune alerte active"
        fi
    else
        log_warning "Alertmanager non accessible"
    fi
}

# Test complet
run_tests() {
    log_info "=== DÉMARRAGE DES TESTS DE MONITORING ==="
    echo ""
    
    check_curl
    check_api
    check_prometheus
    check_grafana
    
    echo ""
    log_info "=== GÉNÉRATION DE TRAFIC ==="
    generate_traffic
    
    echo ""
    log_info "=== VÉRIFICATION DES MÉTRIQUES ==="
    check_metrics
    
    echo ""
    log_info "=== VÉRIFICATION DES ALERTES ==="
    check_alerts
    
    echo ""
    log_success "=== TESTS TERMINÉS ==="
    
    show_access_info
}

# Fonction principale
main() {
    case "${1:-test}" in
        test)
            run_tests
            ;;
        traffic)
            check_api
            generate_traffic
            ;;
        load)
            check_api
            load_test
            ;;
        metrics)
            check_prometheus
            check_metrics
            ;;
        alerts)
            check_alerts
            ;;
        info)
            show_access_info
            ;;
        *)
            echo "Usage: $0 {test|traffic|load|metrics|alerts|info}"
            echo ""
            echo "Commandes:"
            echo "  test     - Exécuter tous les tests"
            echo "  traffic  - Générer du trafic vers l'API"
            echo "  load     - Exécuter un test de charge"
            echo "  metrics  - Vérifier les métriques"
            echo "  alerts   - Vérifier les alertes"
            echo "  info     - Afficher les informations d'accès"
            exit 1
            ;;
    esac
}

# Exécuter la fonction principale
main "$@"
