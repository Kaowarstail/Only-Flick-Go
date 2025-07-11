#!/bin/bash

# Script de test pour le système de messagerie OnlyFlick
# Ce script teste les principaux endpoints de l'API

# Configuration
BASE_URL="http://localhost:8080"
JWT_TOKEN="YOUR_JWT_TOKEN_HERE"  # Remplacer par un vrai token JWT

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 Tests du Système de Messagerie OnlyFlick${NC}"
echo "=================================================="

# Fonction pour tester un endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "\n${YELLOW}Testing: $description${NC}"
    echo "Endpoint: $method $endpoint"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -H "Content-Type: application/json" \
            "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
            -X "$method" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$BASE_URL$endpoint")
    fi
    
    http_code=$(echo "$response" | grep "HTTP_CODE" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_CODE/d')
    
    if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 201 ]; then
        echo -e "${GREEN}✅ SUCCESS (HTTP $http_code)${NC}"
        echo "Response: $body" | jq . 2>/dev/null || echo "Response: $body"
    else
        echo -e "${RED}❌ FAILED (HTTP $http_code)${NC}"
        echo "Response: $body"
    fi
}

# Vérifier si le serveur est en cours d'exécution
echo -e "\n${BLUE}1. Vérification de santé du serveur${NC}"
test_endpoint "GET" "/health" "" "Health check"

# Test de récupération des conversations
echo -e "\n${BLUE}2. Test des endpoints de conversation${NC}"
test_endpoint "GET" "/api/conversations?page=1&limit=10" "" "Récupérer les conversations"

# Test de création de conversation
conversation_data='{"other_user_id": "test-user-2"}'
test_endpoint "POST" "/api/conversations" "$conversation_data" "Créer une conversation"

# Test d'envoi de message
message_data='{"conversation_id": "test-conv-id", "content": "Hello from test!", "message_type": "text"}'
test_endpoint "POST" "/api/messages" "$message_data" "Envoyer un message"

# Test du tableau de bord
echo -e "\n${BLUE}3. Test du dashboard${NC}"
test_endpoint "GET" "/api/messaging/dashboard?page=1&limit=5" "" "Tableau de bord de messagerie"

# Test des statistiques
test_endpoint "GET" "/api/messaging/stats" "" "Statistiques de messagerie"

# Test de recherche
echo -e "\n${BLUE}4. Test de la recherche${NC}"
test_endpoint "GET" "/api/messaging/search?q=hello&type=all&limit=10" "" "Recherche dans les messages"

# Test de démarrage de conversation avec message
echo -e "\n${BLUE}5. Test des opérations avancées${NC}"
start_conv_data='{"other_user_id": "test-user-3", "message": {"content": "Hello new conversation!", "message_type": "text"}}'
test_endpoint "POST" "/api/messaging/start" "$start_conv_data" "Démarrer conversation avec message"

echo -e "\n${BLUE}=================================================="
echo -e "🏁 Tests terminés${NC}"
echo ""
echo -e "${YELLOW}Notes importantes:${NC}"
echo "1. Remplacez YOUR_JWT_TOKEN_HERE par un vrai token JWT"
echo "2. Assurez-vous que le serveur est démarré sur localhost:8080"
echo "3. La base de données doit être configurée avec les migrations"
echo "4. Certains tests échoueront si les données n'existent pas"
echo ""
echo -e "${BLUE}Pour démarrer le serveur de test:${NC}"
echo "cd examples && go run messaging_test_server.go"
