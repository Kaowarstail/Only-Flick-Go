#!/bin/bash

# Script de debug pour tester l'upload d'image avec l'API OnlyFlick

echo "=== Debug Upload d'image OnlyFlick ==="
echo ""

# Configuration
API_URL="http://localhost:8080/api/v1"
TEST_IMAGE="/Users/ilan/Documents/1_EEMI/Projet-RNCP/onlyflick/Only-Flick-Go/test_image.jpg"
USERNAME="testuser"
EMAIL="test@example.com"
PASSWORD="TestPassword123!"

# Vérifier que l'image de test existe
if [ ! -f "$TEST_IMAGE" ]; then
    echo "❌ Fichier de test introuvable : $TEST_IMAGE"
    echo "Créons une image de test simple..."
    
    # Créer une image de test simple si elle n'existe pas
    if command -v convert &> /dev/null; then
        convert -size 100x100 xc:blue "$TEST_IMAGE"
        echo "✅ Image de test créée avec ImageMagick"
    else
        echo "❌ ImageMagick non disponible. Veuillez créer manuellement un fichier image de test."
        exit 1
    fi
fi

# Fonction pour afficher les erreurs
show_error() {
    echo "❌ $1"
    if [ ! -z "$2" ]; then
        echo "   Réponse: $2"
    fi
}

# Fonction pour extraire le token JWT
extract_token() {
    echo "$1" | grep -o '"token":"[^"]*' | cut -d'"' -f4
}

echo "🔍 Test de l'API..."

# Test de santé
health_response=$(curl -s -X GET "$API_URL/health" -H "Content-Type: application/json")
if echo "$health_response" | grep -q 'OK' || echo "$health_response" | grep -q '"status":"ok"'; then
    echo "✅ API accessible"
else
    show_error "API non accessible" "$health_response"
    exit 1
fi

echo ""
echo "🔐 Authentification..."

# Connexion utilisateur
login_response=$(curl -s -X POST "$API_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

token=$(extract_token "$login_response")

if [ -z "$token" ] || [ "$token" = "null" ]; then
    show_error "Échec de l'authentification" "$login_response"
    
    # Essayer de créer un compte utilisateur
    echo ""
    echo "🆕 Création d'un compte utilisateur..."
    
    register_response=$(curl -s -X POST "$API_URL/auth/register" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\",\"username\":\"testuser\",\"role\":\"creator\"}")
    
    if echo "$register_response" | grep -q '"token"'; then
        echo "✅ Compte créé avec succès"
        token=$(extract_token "$register_response")
    else
        show_error "Échec de la création du compte" "$register_response"
        exit 1
    fi
fi

echo "✅ Authentification réussie"
echo "   Token: ${token:0:20}..."

echo ""
echo "📤 Test upload avec CreateContentWithMedia..."

# Test avec l'endpoint CreateContentWithMedia
upload_response=$(curl -s -X POST "$API_URL/contents/upload-image" \
    -H "Authorization: Bearer $token" \
    -F "title=Test Debug Image" \
    -F "description=Test d'upload debug" \
    -F "type=image" \
    -F "is_premium=false" \
    -F "is_published=true" \
    -F "media=@$TEST_IMAGE")

echo "🔍 Réponse complète de l'upload:"
echo "$upload_response" | python3 -m json.tool 2>/dev/null || echo "$upload_response"

echo ""
echo "🔍 Vérification du champ media_url..."

if echo "$upload_response" | grep -q '"media_url"'; then
    media_url=$(echo "$upload_response" | grep -o '"media_url":"[^"]*' | cut -d'"' -f4)
    if [ -z "$media_url" ] || [ "$media_url" = "null" ] || [ "$media_url" = "" ]; then
        echo "❌ media_url est vide ou null"
    else
        echo "✅ media_url trouvé: $media_url"
        
        # Tester l'URL
        echo ""
        echo "🔍 Test de l'URL media..."
        url_test=$(curl -s -o /dev/null -w "%{http_code}" "$media_url")
        if [ "$url_test" = "200" ]; then
            echo "✅ URL media accessible (HTTP 200)"
        else
            echo "❌ URL media non accessible (HTTP $url_test)"
        fi
    fi
else
    echo "❌ Champ media_url non trouvé dans la réponse"
fi

echo ""
echo "🎯 Test terminé"
