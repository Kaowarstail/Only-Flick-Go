#!/bin/bash

echo "🔍 DIAGNOSTIC AVANCÉ - Upload d'images OnlyFlick"
echo "=================================================="

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

API_BASE_URL="http://localhost:8080/api"
TEST_IMAGE="test_image.jpg"
TOKEN=""

echo -e "${BLUE}1. Vérification de la configuration Cloudinary${NC}"
echo "---------------------------------------------------"

# Vérifier les variables d'environnement
echo "🔧 Variables d'environnement Cloudinary:"
if [ -z "$CLOUDINARY_CLOUD_NAME" ]; then
    echo -e "${RED}❌ CLOUDINARY_CLOUD_NAME non défini${NC}"
    CLOUDINARY_MISSING=true
else
    echo -e "${GREEN}✅ CLOUDINARY_CLOUD_NAME: $CLOUDINARY_CLOUD_NAME${NC}"
fi

if [ -z "$CLOUDINARY_API_KEY" ]; then
    echo -e "${RED}❌ CLOUDINARY_API_KEY non défini${NC}"
    CLOUDINARY_MISSING=true
else
    echo -e "${GREEN}✅ CLOUDINARY_API_KEY: ${CLOUDINARY_API_KEY:0:4}***${NC}"
fi

if [ -z "$CLOUDINARY_API_SECRET" ]; then
    echo -e "${RED}❌ CLOUDINARY_API_SECRET non défini${NC}"
    CLOUDINARY_MISSING=true
else
    echo -e "${GREEN}✅ CLOUDINARY_API_SECRET: ${CLOUDINARY_API_SECRET:0:4}***${NC}"
fi

if [ "$CLOUDINARY_MISSING" = true ]; then
    echo -e "${RED}❌ Configuration Cloudinary manquante !${NC}"
    echo "Exportez les variables d'environnement:"
    echo "export CLOUDINARY_CLOUD_NAME=\"your-cloud-name\""
    echo "export CLOUDINARY_API_KEY=\"your-api-key\""
    echo "export CLOUDINARY_API_SECRET=\"your-api-secret\""
    exit 1
fi

echo -e "\n${BLUE}2. Test de connectivité Cloudinary${NC}"
echo "-------------------------------------------"

# Test de connexion à Cloudinary
echo "🌐 Test de connexion à Cloudinary API..."
CLOUDINARY_TEST_URL="https://api.cloudinary.com/v1_1/${CLOUDINARY_CLOUD_NAME}/image/upload"
if curl -s --connect-timeout 5 -o /dev/null -w "%{http_code}" "$CLOUDINARY_TEST_URL" | grep -q "400\|401"; then
    echo -e "${GREEN}✅ Cloudinary API accessible${NC}"
else
    echo -e "${RED}❌ Cloudinary API inaccessible${NC}"
    exit 1
fi

echo -e "\n${BLUE}3. Vérification du serveur OnlyFlick${NC}"
echo "----------------------------------------------"

# Vérifier que le serveur tourne
echo "🏃 Vérification du serveur OnlyFlick..."
if curl -s --connect-timeout 3 "$API_BASE_URL/health" > /dev/null; then
    echo -e "${GREEN}✅ Serveur OnlyFlick en ligne${NC}"
else
    echo -e "${RED}❌ Serveur OnlyFlick hors ligne${NC}"
    echo "Démarrez le serveur avec: go run cmd/api/main.go"
    exit 1
fi

echo -e "\n${BLUE}4. Création d'une image de test${NC}"
echo "------------------------------------"

# Créer une image de test si elle n'existe pas
if [ ! -f "$TEST_IMAGE" ]; then
    echo "🎨 Création d'une image de test..."
    # Créer une image de test simple avec ImageMagick ou curl
    if command -v convert >/dev/null 2>&1; then
        convert -size 300x200 xc:skyblue -pointsize 30 -fill white -gravity center -annotate +0+0 "Test OnlyFlick" "$TEST_IMAGE"
        echo -e "${GREEN}✅ Image de test créée avec ImageMagick${NC}"
    else
        # Télécharger une image de test depuis un service public
        curl -s "https://picsum.photos/300/200" -o "$TEST_IMAGE"
        echo -e "${GREEN}✅ Image de test téléchargée${NC}"
    fi
else
    echo -e "${GREEN}✅ Image de test existante${NC}"
fi

echo -e "\n${BLUE}5. Authentification${NC}"
echo "---------------------"

# Créer un utilisateur de test et récupérer le token
echo "🔑 Création d'un utilisateur de test..."
TEST_USER_EMAIL="test-creator-$(date +%s)@example.com"
TEST_USER_PASSWORD="TestPass123!"

# Inscription
REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"$TEST_USER_EMAIL\",
        \"password\": \"$TEST_USER_PASSWORD\",
        \"username\": \"testcreator$(date +%s)\",
        \"role\": \"creator\"
    }")

echo "📝 Réponse inscription: $REGISTER_RESPONSE"

# Connexion
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"$TEST_USER_EMAIL\",
        \"password\": \"$TEST_USER_PASSWORD\"
    }")

echo "📝 Réponse connexion: $LOGIN_RESPONSE"

TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | sed 's/"token":"//')

if [ -z "$TOKEN" ]; then
    echo -e "${RED}❌ Impossible d'obtenir le token d'authentification${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Token obtenu: ${TOKEN:0:20}...${NC}"

echo -e "\n${BLUE}6. Test d'upload d'image${NC}"
echo "------------------------------"

echo "📤 Upload de l'image de test..."
UPLOAD_RESPONSE=$(curl -s -X POST "$API_BASE_URL/contents/upload-image" \
    -H "Authorization: Bearer $TOKEN" \
    -F "image=@$TEST_IMAGE" \
    -F "title=Test Upload $(date +%s)" \
    -F "description=Image de test pour diagnostic" \
    -F "is_premium=false" \
    -F "is_published=true")

echo "📝 Réponse upload complète:"
echo "$UPLOAD_RESPONSE" | jq '.' 2>/dev/null || echo "$UPLOAD_RESPONSE"

# Analyser la réponse
MEDIA_URL=$(echo "$UPLOAD_RESPONSE" | grep -o '"media_url":"[^"]*' | sed 's/"media_url":"//')
CONTENT_ID=$(echo "$UPLOAD_RESPONSE" | grep -o '"content_id":[0-9]*' | sed 's/"content_id"://')

if [ -n "$MEDIA_URL" ] && [ "$MEDIA_URL" != "null" ] && [ "$MEDIA_URL" != "" ]; then
    echo -e "${GREEN}✅ Upload réussi! URL: $MEDIA_URL${NC}"
else
    echo -e "${RED}❌ Upload échoué ou media_url vide${NC}"
    echo "Analysons la réponse..."
    
    # Vérifier s'il y a une erreur spécifique
    if echo "$UPLOAD_RESPONSE" | grep -q "error"; then
        ERROR_MSG=$(echo "$UPLOAD_RESPONSE" | grep -o '"error":"[^"]*' | sed 's/"error":"//')
        echo -e "${RED}Erreur détectée: $ERROR_MSG${NC}"
    fi
    
    # Vérifier si le content_id est présent
    if [ -n "$CONTENT_ID" ]; then
        echo -e "${YELLOW}⚠️ Content créé avec ID: $CONTENT_ID mais media_url vide${NC}"
    fi
fi

echo -e "\n${BLUE}7. Vérification en base de données${NC}"
echo "-----------------------------------"

if [ -n "$CONTENT_ID" ]; then
    echo "🔍 Vérification du contenu en base..."
    
    # Test de récupération du contenu
    CONTENT_RESPONSE=$(curl -s -X GET "$API_BASE_URL/contents/$CONTENT_ID" \
        -H "Authorization: Bearer $TOKEN")
    
    echo "📝 Contenu en base:"
    echo "$CONTENT_RESPONSE" | jq '.' 2>/dev/null || echo "$CONTENT_RESPONSE"
    
    # Vérifier la media_url en base
    DB_MEDIA_URL=$(echo "$CONTENT_RESPONSE" | grep -o '"media_url":"[^"]*' | sed 's/"media_url":"//')
    
    if [ -n "$DB_MEDIA_URL" ] && [ "$DB_MEDIA_URL" != "null" ] && [ "$DB_MEDIA_URL" != "" ]; then
        echo -e "${GREEN}✅ media_url sauvegardée en base: $DB_MEDIA_URL${NC}"
    else
        echo -e "${RED}❌ media_url vide en base de données${NC}"
    fi
fi

echo -e "\n${BLUE}8. Logs du serveur${NC}"
echo "-------------------"

echo "📋 Dernières lignes des logs du serveur:"
echo -e "${YELLOW}Note: Vérifiez les logs du serveur Go pour plus de détails${NC}"

echo -e "\n${BLUE}9. Résumé du diagnostic${NC}"
echo "------------------------"

echo "🎯 Points à vérifier:"
echo "1. Configuration Cloudinary dans les variables d'environnement"
echo "2. Connectivité à Cloudinary API"
echo "3. Logs du serveur Go pendant l'upload"
echo "4. Sauvegarde GORM en base de données"
echo "5. Réponse complète de l'API d'upload"

echo -e "\n${GREEN}Diagnostic terminé !${NC}"
echo "Consultez les logs du serveur Go pour plus de détails sur l'upload."
