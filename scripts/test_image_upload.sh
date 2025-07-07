#!/bin/bash

# Script de test pour l'upload d'images avec Cloudinary
# Ce script teste l'upload d'images via l'API OnlyFlick

echo "🧪 Test d'upload d'images OnlyFlick"
echo "=================================="
echo ""

# Configuration
API_BASE_URL="http://localhost:8080/api/v1"
TEST_IMAGE_PATH="/Users/ilan/Documents/1_EEMI/Projet-RNCP/onlyflick/Only-Flick-Go/test_image.jpg"

# Vérifier si le serveur est en cours d'exécution
echo "🔍 Vérification du serveur..."
if curl -s "$API_BASE_URL/health" > /dev/null; then
    echo "✅ Serveur en cours d'exécution"
else
    echo "❌ Serveur non accessible. Assurez-vous qu'il est démarré."
    exit 1
fi

# Vérifier si l'image de test existe
if [ ! -f "$TEST_IMAGE_PATH" ]; then
    echo "📥 Téléchargement d'une image de test..."
    curl -o "$TEST_IMAGE_PATH" "https://picsum.photos/800/600.jpg"
    echo "✅ Image de test téléchargée"
fi

echo ""
echo "🔐 Connexion en tant qu'utilisateur test..."

# Connexion pour obtenir un token
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice_photographer",
    "password": "alice123"
  }')

# Extraire le token
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ Échec de la connexion. Réponse:"
    echo $LOGIN_RESPONSE
    exit 1
fi

echo "✅ Connexion réussie"

# Obtenir l'ID utilisateur
USER_ID=$(echo $LOGIN_RESPONSE | jq -r '.user.id')
echo "👤 ID utilisateur: $USER_ID"

echo ""
echo "📝 Création d'un nouveau contenu..."

# Créer un contenu
CONTENT_RESPONSE=$(curl -s -X POST "$API_BASE_URL/content" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Test Upload Cloudinary",
    "description": "Test d'upload d'image avec Cloudinary",
    "type": "image",
    "is_premium": false,
    "is_published": true
  }')

# Extraire l'ID du contenu
CONTENT_ID=$(echo $CONTENT_RESPONSE | jq -r '.id')

if [ "$CONTENT_ID" = "null" ] || [ -z "$CONTENT_ID" ]; then
    echo "❌ Échec de la création du contenu. Réponse:"
    echo $CONTENT_RESPONSE
    exit 1
fi

echo "✅ Contenu créé avec ID: $CONTENT_ID"

echo ""
echo "🌟 Upload de l'image vers Cloudinary..."

# Upload de l'image
UPLOAD_RESPONSE=$(curl -s -X POST "$API_BASE_URL/content/$CONTENT_ID/media" \
  -H "Authorization: Bearer $TOKEN" \
  -F "media=@$TEST_IMAGE_PATH")

echo "📤 Réponse de l'upload:"
echo $UPLOAD_RESPONSE | jq '.'

# Vérifier si l'upload a réussi
if echo $UPLOAD_RESPONSE | jq -e '.media_url' > /dev/null; then
    MEDIA_URL=$(echo $UPLOAD_RESPONSE | jq -r '.media_url')
    THUMBNAIL_URL=$(echo $UPLOAD_RESPONSE | jq -r '.thumbnail_url')
    PUBLIC_ID=$(echo $UPLOAD_RESPONSE | jq -r '.public_id')
    
    echo ""
    echo "🎉 Upload réussi !"
    echo "🌐 URL de l'image: $MEDIA_URL"
    echo "🖼️  URL de la miniature: $THUMBNAIL_URL"
    echo "🔑 Public ID: $PUBLIC_ID"
    
    echo ""
    echo "🔍 Vérification du contenu mis à jour..."
    curl -s -X GET "$API_BASE_URL/content/$CONTENT_ID" \
      -H "Authorization: Bearer $TOKEN" | jq '.'
else
    echo "❌ Échec de l'upload"
    echo "Erreur: $(echo $UPLOAD_RESPONSE | jq -r '.error // "Erreur inconnue"')"
fi

echo ""
echo "🏁 Test terminé"
