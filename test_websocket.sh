#!/bin/bash

# Script de test WebSocket OnlyFlick
# Usage: ./test_websocket.sh [JWT_TOKEN]

set -e

# Configuration
API_URL="localhost:8080"
JWT_TOKEN=${1:-""}

if [ -z "$JWT_TOKEN" ]; then
    echo "❌ Token JWT requis"
    echo "Usage: $0 [JWT_TOKEN]"
    exit 1
fi

echo "🚀 Test WebSocket OnlyFlick"
echo "API: $API_URL"
echo "Token: ${JWT_TOKEN:0:20}..."

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Fonction de test
test_endpoint() {
    local url=$1
    local description=$2
    
    echo -e "\n${YELLOW}🔍 Test: $description${NC}"
    echo "URL: $url"
    
    response=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $JWT_TOKEN" "$url")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" -eq 200 ]; then
        echo -e "${GREEN}✅ Succès ($http_code)${NC}"
        echo "$body" | jq . 2>/dev/null || echo "$body"
    else
        echo -e "${RED}❌ Échec ($http_code)${NC}"
        echo "$body"
    fi
}

# Test des endpoints REST WebSocket
echo -e "\n${YELLOW}📡 Tests des endpoints REST WebSocket${NC}"

test_endpoint "http://$API_URL/api/v1/ws/health" "Health Check WebSocket"
test_endpoint "http://$API_URL/api/v1/ws/info" "Informations de connexion"
test_endpoint "http://$API_URL/api/v1/ws/online" "Utilisateurs en ligne"

# Test de connexion WebSocket avec l'outil Go
echo -e "\n${YELLOW}🔌 Test de connexion WebSocket${NC}"

if command -v go &> /dev/null; then
    echo "Compilation du client de test..."
    cd cmd/websocket-test
    go build -o websocket-test main.go
    
    echo "Démarrage du test WebSocket (5 secondes)..."
    timeout 5s ./websocket-test -addr="$API_URL" -token="$JWT_TOKEN" 2>&1 | head -n 20 || true
    
    echo -e "\n${GREEN}✅ Test WebSocket terminé${NC}"
    cd ../..
else
    echo -e "${YELLOW}⚠️ Go non installé, test WebSocket ignoré${NC}"
fi

# Test de charge léger
echo -e "\n${YELLOW}📊 Test de charge léger${NC}"

if command -v wscat &> /dev/null; then
    echo "Test avec wscat..."
    for i in {1..5}; do
        echo "Connexion $i/5..."
        timeout 2s wscat -c "ws://$API_URL/api/v1/ws" -H "Authorization: Bearer $JWT_TOKEN" &
    done
    
    wait
    echo -e "${GREEN}✅ Test de charge terminé${NC}"
else
    echo -e "${YELLOW}⚠️ wscat non installé, test de charge ignoré${NC}"
    echo "Installation: npm install -g wscat"
fi

# Résumé
echo -e "\n${GREEN}🎉 Tests WebSocket terminés${NC}"
echo "Endpoints testés:"
echo "  - Health Check"
echo "  - Informations de connexion"
echo "  - Utilisateurs en ligne"
echo "  - Connexion WebSocket"
echo "  - Test de charge léger"

echo -e "\n${YELLOW}📝 Pour tester manuellement:${NC}"
echo "1. Avec le client Go:"
echo "   cd cmd/websocket-test && go run main.go -token=$JWT_TOKEN"
echo ""
echo "2. Avec wscat:"
echo "   wscat -c ws://$API_URL/api/v1/ws -H \"Authorization: Bearer $JWT_TOKEN\""
echo ""
echo "3. Avec JavaScript (navigateur):"
echo "   const ws = new WebSocket('ws://$API_URL/api/v1/ws');"
echo ""
echo "4. Endpoints REST:"
echo "   curl -H \"Authorization: Bearer $JWT_TOKEN\" http://$API_URL/api/v1/ws/health"
