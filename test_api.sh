#!/bin/bash

# Script de test pour l'API OnlyFlick
# Ce script teste les principaux endpoints de messagerie et d'édition de profil

set -e

# Configuration
API_BASE_URL="http://localhost:8080/api/v1"
JWT_TOKEN="${JWT_TOKEN:-your-jwt-token-here}"

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Fonction pour afficher les messages
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Fonction pour faire des requêtes HTTP
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local content_type=${4:-"application/json"}
    
    if [ -z "$data" ]; then
        curl -s -X "$method" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -H "Content-Type: $content_type" \
            "$API_BASE_URL$endpoint"
    else
        curl -s -X "$method" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -H "Content-Type: $content_type" \
            -d "$data" \
            "$API_BASE_URL$endpoint"
    fi
}

# Vérifier que l'API est accessible
test_health_check() {
    log_info "Test de santé de l'API..."
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "$API_BASE_URL/health")
    
    if [ "$response" -eq 200 ]; then
        log_info "✓ API accessible"
    else
        log_error "✗ API non accessible (HTTP $response)"
        exit 1
    fi
}

# Test des conversations
test_conversations() {
    log_info "Test des conversations..."
    
    # Test 1: Récupérer les conversations
    log_info "1. Récupération des conversations..."
    response=$(make_request "GET" "/messaging/conversations")
    
    if echo "$response" | grep -q '"success":true'; then
        log_info "✓ Récupération des conversations réussie"
    else
        log_error "✗ Échec de la récupération des conversations"
        echo "$response"
    fi
    
    # Test 2: Créer une conversation (nécessite un autre utilisateur)
    log_info "2. Création d'une conversation..."
    conversation_data='{"participant_id":"test-user-id"}'
    response=$(make_request "POST" "/messaging/conversations" "$conversation_data")
    
    if echo "$response" | grep -q '"success":true'; then
        log_info "✓ Création de conversation réussie"
        # Extraire l'ID de la conversation pour les tests suivants
        CONVERSATION_ID=$(echo "$response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        log_info "ID de conversation: $CONVERSATION_ID"
    else
        log_warning "⚠ Création de conversation échouée (normal si utilisateur test n'existe pas)"
    fi
    
    # Test 3: Nombre de messages non lus
    log_info "3. Récupération du nombre de messages non lus..."
    response=$(make_request "GET" "/messaging/unread-count")
    
    if echo "$response" | grep -q '"success":true'; then
        log_info "✓ Récupération du nombre de messages non lus réussie"
    else
        log_error "✗ Échec de la récupération du nombre de messages non lus"
    fi
}

# Test des messages
test_messages() {
    log_info "Test des messages..."
    
    # Test 1: Envoyer un message (nécessite une conversation)
    if [ -n "$CONVERSATION_ID" ]; then
        log_info "1. Envoi d'un message..."
        message_data="{\"conversation_id\":\"$CONVERSATION_ID\",\"content\":\"Test message\",\"message_type\":\"text\"}"
        response=$(make_request "POST" "/messaging/messages" "$message_data")
        
        if echo "$response" | grep -q '"success":true'; then
            log_info "✓ Envoi de message réussi"
            MESSAGE_ID=$(echo "$response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        else
            log_error "✗ Échec de l'envoi du message"
            echo "$response"
        fi
    else
        log_warning "⚠ Pas de conversation disponible pour tester les messages"
    fi
    
    # Test 2: Statistiques des messages
    log_info "2. Récupération des statistiques de messages..."
    response=$(make_request "GET" "/messaging/messages/stats")
    
    if echo "$response" | grep -q '"success":true'; then
        log_info "✓ Récupération des statistiques de messages réussie"
    else
        log_error "✗ Échec de la récupération des statistiques de messages"
    fi
}

# Test des médias
test_media() {
    log_info "Test des médias..."
    
    # Test 1: Récupérer les fichiers de l'utilisateur
    log_info "1. Récupération des fichiers utilisateur..."
    response=$(make_request "GET" "/messaging/media")
    
    if echo "$response" | grep -q '"success":true'; then
        log_info "✓ Récupération des fichiers utilisateur réussie"
    else
        log_error "✗ Échec de la récupération des fichiers utilisateur"
    fi
    
    # Test 2: Statistiques des médias
    log_info "2. Récupération des statistiques de médias..."
    response=$(make_request "GET" "/messaging/media/stats")
    
    if echo "$response" | grep -q '"success":true'; then
        log_info "✓ Récupération des statistiques de médias réussie"
    else
        log_error "✗ Échec de la récupération des statistiques de médias"
    fi
}

# Test des profils utilisateur
test_user_profiles() {
    log_info "Test des profils utilisateur..."
    
    # Test 1: Mise à jour du profil utilisateur
    log_info "1. Mise à jour du profil utilisateur..."
    profile_data='{"first_name":"Test","last_name":"User","biography":"Test biography","location":"Test City"}'
    response=$(make_request "PUT" "/profiles/users/me" "$profile_data")
    
    if echo "$response" | grep -q '"success":true'; then
        log_info "✓ Mise à jour du profil utilisateur réussie"
    else
        log_error "✗ Échec de la mise à jour du profil utilisateur"
        echo "$response"
    fi
    
    # Test 2: Mise à jour des liens sociaux
    log_info "2. Mise à jour des liens sociaux..."
    social_data='{"links":{"instagram":"testuser","twitter":"testuser","website":"https://testuser.com"}}'
    response=$(make_request "PUT" "/profiles/users/me/social-links" "$social_data")
    
    if echo "$response" | grep -q '"success":true'; then
        log_info "✓ Mise à jour des liens sociaux réussie"
    else
        log_error "✗ Échec de la mise à jour des liens sociaux"
        echo "$response"
    fi
}

# Test des profils créateur
test_creator_profiles() {
    log_info "Test des profils créateur..."
    
    # Test 1: Récupération des gains (nécessite d'être créateur)
    log_info "1. Récupération des gains créateur..."
    response=$(make_request "GET" "/profiles/creators/me/earnings")
    
    if echo "$response" | grep -q '"success":true'; then
        log_info "✓ Récupération des gains créateur réussie"
    elif echo "$response" | grep -q '"error"' && echo "$response" | grep -q '"Accès réservé aux créateurs"'; then
        log_warning "⚠ Accès réservé aux créateurs (normal si utilisateur n'est pas créateur)"
    else
        log_error "✗ Échec de la récupération des gains créateur"
        echo "$response"
    fi
    
    # Test 2: Statistiques créateur
    log_info "2. Récupération des statistiques créateur..."
    response=$(make_request "GET" "/profiles/creators/me/stats")
    
    if echo "$response" | grep -q '"success":true'; then
        log_info "✓ Récupération des statistiques créateur réussie"
    elif echo "$response" | grep -q '"error"' && echo "$response" | grep -q '"Accès réservé aux créateurs"'; then
        log_warning "⚠ Accès réservé aux créateurs (normal si utilisateur n'est pas créateur)"
    else
        log_error "✗ Échec de la récupération des statistiques créateur"
    fi
}

# Test des erreurs d'authentification
test_authentication() {
    log_info "Test d'authentification..."
    
    # Test sans token
    log_info "1. Test sans token d'authentification..."
    response=$(curl -s -o /dev/null -w "%{http_code}" "$API_BASE_URL/messaging/conversations")
    
    if [ "$response" -eq 401 ]; then
        log_info "✓ Authentification requise correctement vérifiée"
    else
        log_error "✗ Authentification non vérifiée (HTTP $response)"
    fi
    
    # Test avec token invalide
    log_info "2. Test avec token invalide..."
    response=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Authorization: Bearer invalid-token" \
        "$API_BASE_URL/messaging/conversations")
    
    if [ "$response" -eq 401 ]; then
        log_info "✓ Token invalide correctement rejeté"
    else
        log_error "✗ Token invalide non rejeté (HTTP $response)"
    fi
}

# Fonction principale
main() {
    log_info "=== Test de l'API OnlyFlick ==="
    log_info "URL de base: $API_BASE_URL"
    
    if [ "$JWT_TOKEN" = "your-jwt-token-here" ]; then
        log_warning "Token JWT par défaut utilisé. Définissez JWT_TOKEN pour des tests complets."
    fi
    
    # Tests de base
    test_health_check
    test_authentication
    
    # Tests fonctionnels (nécessitent un token valide)
    if [ "$JWT_TOKEN" != "your-jwt-token-here" ]; then
        test_conversations
        test_messages
        test_media
        test_user_profiles
        test_creator_profiles
    else
        log_warning "Tests fonctionnels ignorés (token JWT requis)"
    fi
    
    log_info "=== Fin des tests ==="
}

# Vérifier les dépendances
check_dependencies() {
    if ! command -v curl &> /dev/null; then
        log_error "curl n'est pas installé"
        exit 1
    fi
    
    if ! command -v grep &> /dev/null; then
        log_error "grep n'est pas installé"
        exit 1
    fi
}

# Aide
show_help() {
    cat << EOF
Usage: $0 [OPTIONS]

Options:
  -h, --help          Afficher cette aide
  -t, --token TOKEN   Token JWT pour l'authentification
  -u, --url URL       URL de base de l'API (défaut: $API_BASE_URL)

Variables d'environnement:
  JWT_TOKEN          Token JWT pour l'authentification
  API_BASE_URL       URL de base de l'API

Exemples:
  $0                                    # Tests basiques
  $0 -t "your-jwt-token"               # Tests avec authentification
  JWT_TOKEN="token" $0                 # Tests avec token depuis env
  $0 -u "http://api.example.com/v1"    # Tests avec URL custom

EOF
}

# Parsing des arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -t|--token)
            JWT_TOKEN="$2"
            shift 2
            ;;
        -u|--url)
            API_BASE_URL="$2"
            shift 2
            ;;
        *)
            log_error "Option inconnue: $1"
            show_help
            exit 1
            ;;
    esac
done

# Exécuter les tests
check_dependencies
main
