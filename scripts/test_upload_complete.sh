#!/bin/bash

# Test script pour vérifier l'upload d'image avec l'API OnlyFlick

echo "=== Test d'upload d'image pour OnlyFlick ==="
echo ""

# Configuration
API_URL="http://localhost:8080/api/v1"
TEST_IMAGE="/Users/ilan/Documents/1_EEMI/Projet-RNCP/onlyflick/Only-Flick-Go/test_image.jpg"
EMAIL="test@example.com"
PASSWORD="testpassword"

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

# Fonction pour extraire l'ID du contenu
extract_content_id() {
    echo "$1" | grep -o '"id":[0-9]*' | cut -d':' -f2
}

echo "🔍 Vérification de l'API..."

# Test de santé
health_response=$(curl -s -X GET "$API_URL/health" -H "Content-Type: application/json")
if [ $? -eq 0 ]; then
    echo "✅ API accessible"
else
    show_error "API non accessible"
    exit 1
fi

echo ""
echo "🔑 Authentification..."

# Connexion pour récupérer le token
login_response=$(curl -s -X POST "$API_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

if echo "$login_response" | grep -q '"token"'; then
    token=$(extract_token "$login_response")
    echo "✅ Connexion réussie"
    echo "   Token: ${token:0:20}..."
else
    show_error "Échec de la connexion" "$login_response"
    exit 1
fi

echo ""
echo "📝 Création du contenu..."

# Créer un contenu
create_response=$(curl -s -X POST "$API_URL/contents" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $token" \
    -d '{
        "title": "Test Upload Image",
        "description": "Test d'\''upload d'\''image depuis le script",
        "type": "image",
        "is_premium": false,
        "is_published": true
    }')

if echo "$create_response" | grep -q '"id"'; then
    content_id=$(extract_content_id "$create_response")
    echo "✅ Contenu créé avec succès"
    echo "   ID: $content_id"
else
    show_error "Échec de la création du contenu" "$create_response"
    exit 1
fi

echo ""
echo "📤 Upload de l'image..."

# Upload du média
upload_response=$(curl -s -X POST "$API_URL/contents/$content_id/media" \
    -H "Authorization: Bearer $token" \
    -F "media=@$TEST_IMAGE")

if echo "$upload_response" | grep -q '"media_url"'; then
    echo "✅ Upload réussi"
    echo "   Réponse: $upload_response"
else
    show_error "Échec de l'upload" "$upload_response"
    
    # Afficher les détails de l'erreur
    echo ""
    echo "📋 Détails de l'erreur:"
    echo "$upload_response" | python3 -m json.tool 2>/dev/null || echo "$upload_response"
    exit 1
fi

echo ""
echo "🎉 Test terminé avec succès !"
echo "   - API accessible"
echo "   - Authentification OK"
echo "   - Création de contenu OK"
echo "   - Upload d'image OK"
