#!/bin/bash

# Script de validation des secrets pour le monitoring OnlyFlick
# Usage: ./validate_secrets.sh [--github-actions|--docker-secrets|--kubernetes]

set -e

# Couleurs pour les logs
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

# Configuration
MODE="environment"
VERBOSE=false
GENERATE_MISSING=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --github-actions)
            MODE="github-actions"
            shift
            ;;
        --docker-secrets)
            MODE="docker-secrets"
            shift
            ;;
        --kubernetes)
            MODE="kubernetes"
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --generate-missing|-g)
            GENERATE_MISSING=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --github-actions    Valider les secrets GitHub Actions"
            echo "  --docker-secrets    Valider les secrets Docker"
            echo "  --kubernetes        Valider les secrets Kubernetes"
            echo "  --verbose, -v       Mode verbose"
            echo "  --generate-missing  Générer les secrets manquants"
            echo "  --help, -h          Afficher cette aide"
            exit 0
            ;;
        *)
            log_error "Option inconnue: $1"
            exit 1
            ;;
    esac
done

# Secrets obligatoires
declare -A REQUIRED_SECRETS=(
    ["GRAFANA_ADMIN_PASSWORD"]="Mot de passe admin Grafana (min 8 caractères)"
    ["GRAFANA_SECRET_KEY"]="Clé secrète Grafana (min 32 caractères)"
    ["GRAFANA_DB_PASSWORD"]="Mot de passe base de données Grafana"
    ["SMTP_PASSWORD"]="Mot de passe SMTP pour les alertes"
    ["WEBHOOK_TOKEN"]="Token pour les webhooks d'alertes"
)

# Secrets optionnels
declare -A OPTIONAL_SECRETS=(
    ["SLACK_WEBHOOK_URL"]="URL du webhook Slack"
    ["LETSENCRYPT_EMAIL"]="Email pour Let's Encrypt"
    ["DOMAIN"]="Domaine principal"
    ["MONITORING_DOMAIN"]="Domaine de monitoring"
    ["API_HOST"]="Host de l'API"
    ["ADMIN_EMAIL"]="Email de l'administrateur"
    ["SMTP_HOST"]="Serveur SMTP"
    ["SMTP_FROM"]="Adresse email d'envoi"
    ["SMTP_USERNAME"]="Nom d'utilisateur SMTP"
)

# Fonction pour générer un secret sécurisé
generate_secret() {
    local type=$1
    local length=${2:-32}
    
    case $type in
        "password")
            # Générer un mot de passe avec caractères spéciaux
            openssl rand -base64 $length | tr -d "=+/" | cut -c1-$length
            ;;
        "hex")
            # Générer une clé hexadécimale
            openssl rand -hex $((length/2))
            ;;
        "uuid")
            # Générer un UUID
            uuidgen | tr -d '-' | tr '[:upper:]' '[:lower:]'
            ;;
        *)
            # Par défaut, base64
            openssl rand -base64 $length | tr -d "=+/" | cut -c1-$length
            ;;
    esac
}

# Validation des secrets dans les variables d'environnement
validate_environment_secrets() {
    log_info "Validation des secrets dans les variables d'environnement..."
    
    local errors=0
    local warnings=0
    
    # Vérifier les secrets obligatoires
    for secret in "${!REQUIRED_SECRETS[@]}"; do
        local value="${!secret}"
        local description="${REQUIRED_SECRETS[$secret]}"
        
        if [ -z "$value" ]; then
            log_error "Secret manquant: $secret ($description)"
            
            if [ "$GENERATE_MISSING" = true ]; then
                case $secret in
                    "GRAFANA_SECRET_KEY")
                        local generated=$(generate_secret "hex" 32)
                        echo "export $secret=\"$generated\"" >> .env.generated
                        log_info "Généré: $secret"
                        ;;
                    "GRAFANA_ADMIN_PASSWORD"|"GRAFANA_DB_PASSWORD"|"SMTP_PASSWORD")
                        local generated=$(generate_secret "password" 16)
                        echo "export $secret=\"$generated\"" >> .env.generated
                        log_info "Généré: $secret"
                        ;;
                    "WEBHOOK_TOKEN")
                        local generated=$(generate_secret "uuid")
                        echo "export $secret=\"$generated\"" >> .env.generated
                        log_info "Généré: $secret"
                        ;;
                esac
            fi
            
            ((errors++))
        else
            # Vérifier la longueur et la complexité
            case $secret in
                "GRAFANA_SECRET_KEY")
                    if [ ${#value} -lt 32 ]; then
                        log_error "$secret doit faire au moins 32 caractères"
                        ((errors++))
                    else
                        log_success "$secret: ✅ Valide"
                    fi
                    ;;
                "GRAFANA_ADMIN_PASSWORD"|"GRAFANA_DB_PASSWORD"|"SMTP_PASSWORD")
                    if [ ${#value} -lt 8 ]; then
                        log_error "$secret doit faire au moins 8 caractères"
                        ((errors++))
                    elif [[ "$value" =~ ^[a-zA-Z0-9]*$ ]]; then
                        log_warning "$secret devrait contenir des caractères spéciaux"
                        ((warnings++))
                    else
                        log_success "$secret: ✅ Valide"
                    fi
                    ;;
                *)
                    if [ ${#value} -lt 4 ]; then
                        log_error "$secret est trop court"
                        ((errors++))
                    else
                        log_success "$secret: ✅ Valide"
                    fi
                    ;;
            esac
            
            # Vérifier les valeurs par défaut dangereuses
            if [[ "$value" == *"CHANGE_ME"* ]] || [[ "$value" == *"password"* ]] || [[ "$value" == *"123"* ]]; then
                log_error "$secret contient une valeur par défaut dangereuse"
                ((errors++))
            fi
        fi
    done
    
    # Vérifier les secrets optionnels
    for secret in "${!OPTIONAL_SECRETS[@]}"; do
        local value="${!secret}"
        local description="${OPTIONAL_SECRETS[$secret]}"
        
        if [ -z "$value" ]; then
            if [ "$VERBOSE" = true ]; then
                log_warning "Secret optionnel manquant: $secret ($description)"
            fi
            ((warnings++))
        else
            if [ "$VERBOSE" = true ]; then
                log_success "$secret: ✅ Défini"
            fi
        fi
    done
    
    # Résumé
    if [ $errors -eq 0 ]; then
        log_success "Tous les secrets obligatoires sont valides"
    else
        log_error "$errors erreur(s) détectée(s)"
    fi
    
    if [ $warnings -gt 0 ]; then
        log_warning "$warnings avertissement(s)"
    fi
    
    return $errors
}

# Validation des secrets GitHub Actions
validate_github_actions_secrets() {
    log_info "Validation des secrets GitHub Actions..."
    
    if [ -z "$GITHUB_TOKEN" ]; then
        log_error "GITHUB_TOKEN requis pour valider les secrets GitHub Actions"
        return 1
    fi
    
    local repo="${GITHUB_REPOSITORY:-"owner/repo"}"
    local api_url="https://api.github.com/repos/$repo/actions/secrets"
    
    log_info "Vérification des secrets pour le repository: $repo"
    
    # Récupérer la liste des secrets
    local secrets_response=$(curl -s -H "Authorization: token $GITHUB_TOKEN" "$api_url")
    
    if [ $? -ne 0 ]; then
        log_error "Impossible de récupérer les secrets GitHub Actions"
        return 1
    fi
    
    local errors=0
    
    # Vérifier chaque secret obligatoire
    for secret in "${!REQUIRED_SECRETS[@]}"; do
        local description="${REQUIRED_SECRETS[$secret]}"
        
        if echo "$secrets_response" | grep -q "\"name\":\"$secret\""; then
            log_success "$secret: ✅ Configuré"
        else
            log_error "$secret: ❌ Manquant ($description)"
            ((errors++))
        fi
    done
    
    # Vérifier les secrets optionnels
    if [ "$VERBOSE" = true ]; then
        for secret in "${!OPTIONAL_SECRETS[@]}"; do
            local description="${OPTIONAL_SECRETS[$secret]}"
            
            if echo "$secrets_response" | grep -q "\"name\":\"$secret\""; then
                log_success "$secret: ✅ Configuré (optionnel)"
            else
                log_warning "$secret: ⚠️ Manquant (optionnel) - $description"
            fi
        done
    fi
    
    return $errors
}

# Validation des secrets Docker
validate_docker_secrets() {
    log_info "Validation des secrets Docker..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker n'est pas installé"
        return 1
    fi
    
    local errors=0
    
    # Vérifier chaque secret obligatoire
    for secret in "${!REQUIRED_SECRETS[@]}"; do
        local description="${REQUIRED_SECRETS[$secret]}"
        local secret_name=$(echo "$secret" | tr '[:upper:]' '[:lower:]')
        
        if docker secret ls | grep -q "$secret_name"; then
            log_success "$secret: ✅ Configuré"
        else
            log_error "$secret: ❌ Manquant ($description)"
            ((errors++))
        fi
    done
    
    return $errors
}

# Validation des secrets Kubernetes
validate_kubernetes_secrets() {
    log_info "Validation des secrets Kubernetes..."
    
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl n'est pas installé"
        return 1
    fi
    
    local secret_name="monitoring-secrets"
    local namespace="${KUBERNETES_NAMESPACE:-default}"
    
    log_info "Vérification du secret '$secret_name' dans le namespace '$namespace'"
    
    if ! kubectl get secret "$secret_name" -n "$namespace" &> /dev/null; then
        log_error "Secret Kubernetes '$secret_name' non trouvé"
        return 1
    fi
    
    local errors=0
    
    # Vérifier chaque clé dans le secret
    for secret in "${!REQUIRED_SECRETS[@]}"; do
        local description="${REQUIRED_SECRETS[$secret]}"
        local key=$(echo "$secret" | tr '[:upper:]' '[:lower:]' | tr '_' '-')
        
        if kubectl get secret "$secret_name" -n "$namespace" -o jsonpath="{.data.$key}" &> /dev/null; then
            log_success "$secret: ✅ Configuré"
        else
            log_error "$secret: ❌ Manquant ($description)"
            ((errors++))
        fi
    done
    
    return $errors
}

# Fonction principale
main() {
    log_info "Validation des secrets OnlyFlick Monitoring"
    log_info "Mode: $MODE"
    
    local result=0
    
    case $MODE in
        "environment")
            validate_environment_secrets
            result=$?
            ;;
        "github-actions")
            validate_github_actions_secrets
            result=$?
            ;;
        "docker-secrets")
            validate_docker_secrets
            result=$?
            ;;
        "kubernetes")
            validate_kubernetes_secrets
            result=$?
            ;;
        *)
            log_error "Mode non supporté: $MODE"
            exit 1
            ;;
    esac
    
    if [ $result -eq 0 ]; then
        log_success "🎉 Validation terminée avec succès"
        
        if [ "$GENERATE_MISSING" = true ] && [ -f ".env.generated" ]; then
            log_info "Secrets générés dans .env.generated"
            log_warning "ATTENTION: Sécurisez ce fichier et ne le commitez pas!"
        fi
    else
        log_error "❌ Validation échouée ($result erreur(s))"
        
        if [ "$GENERATE_MISSING" = true ]; then
            log_info "Utilisez --generate-missing pour générer les secrets manquants"
        fi
    fi
    
    return $result
}

# Exécuter la fonction principale
main "$@"
