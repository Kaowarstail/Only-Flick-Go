#!/bin/bash

echo "🔧 Test de l'upload d'image avec Cloudinary"
echo "=============================================="

# Naviguer vers le répertoire de l'API
cd /Users/ilan/Documents/1_EEMI/Projet-RNCP/onlyflick/Only-Flick-Go

# Vérifier si l'API est en cours d'exécution
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "⚠️  L'API n'est pas en cours d'exécution. Démarrage..."
    # Démarrer l'API en arrière-plan
    go run cmd/api/main.go &
    API_PID=$!
    
    # Attendre que l'API démarre
    echo "⏳ Attente du démarrage de l'API..."
    sleep 10
    
    # Vérifier si l'API est maintenant disponible
    if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "❌ L'API n'a pas pu démarrer"
        kill $API_PID 2>/dev/null
        exit 1
    fi
    
    STARTED_API=true
    echo "✅ API démarrée avec succès"
else
    echo "✅ API déjà en cours d'exécution"
    STARTED_API=false
fi

# Préparer un fichier d'image de test
if [ ! -f "test_image.jpg" ]; then
    echo "⚠️  Fichier test_image.jpg non trouvé, création d'un fichier fictif..."
    # Créer un fichier minimal pour le test
    echo "Test image content" > test_image.jpg
fi

# Créer un JWT de test pour l'authentification
echo "🔐 Génération d'un token JWT de test..."
JWT_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username": "testuser", "password": "testpass"}' | \
    jq -r '.token' 2>/dev/null || echo "")

if [ -z "$JWT_TOKEN" ] || [ "$JWT_TOKEN" = "null" ]; then
    echo "⚠️  Impossible d'obtenir un token JWT valide, test sans authentification"
    JWT_TOKEN=""
fi

# Test de l'upload d'image
echo "📤 Test de l'upload d'image..."
echo "==============================================="

# Créer un Content et uploader une image
RESPONSE=$(curl -s -X POST http://localhost:8080/api/content/upload \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -F "title=Test Image Upload" \
    -F "description=Test d'upload d'image via script" \
    -F "type=image" \
    -F "image=@test_image.jpg")

echo "📋 Réponse du serveur :"
echo "$RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"

# Analyser la réponse pour extraire les informations importantes
if echo "$RESPONSE" | jq . > /dev/null 2>&1; then
    CONTENT_ID=$(echo "$RESPONSE" | jq -r '.content.id // empty')
    MEDIA_URL=$(echo "$RESPONSE" | jq -r '.content.media_url // empty')
    
    echo ""
    echo "📊 Résultats du test :"
    echo "======================"
    
    if [ -n "$CONTENT_ID" ]; then
        echo "✅ Content créé avec ID: $CONTENT_ID"
    else
        echo "❌ Échec de la création du contenu"
    fi
    
    if [ -n "$MEDIA_URL" ] && [ "$MEDIA_URL" != "null" ] && [ "$MEDIA_URL" != "" ]; then
        echo "✅ URL de l'image: $MEDIA_URL"
        echo "✅ Upload Cloudinary réussi !"
    else
        echo "❌ URL de l'image manquante - problème avec Cloudinary"
    fi
else
    echo "❌ Réponse invalide du serveur"
fi

# Vérification dans la base de données
echo ""
echo "🔍 Vérification en base de données..."
echo "===================================="

# Connexion à la base pour vérifier les données
if [ -n "$CONTENT_ID" ]; then
    echo "SELECT id, title, media_url FROM contents WHERE id = '$CONTENT_ID';" | \
    psql -h db.ckzjonximwvlbkqhdkoo.supabase.co -p 5432 -U postgres -d postgres 2>/dev/null | \
    head -10
else
    echo "SELECT id, title, media_url FROM contents ORDER BY created_at DESC LIMIT 5;" | \
    psql -h db.ckzjonximwvlbkqhdkoo.supabase.co -p 5432 -U postgres -d postgres 2>/dev/null | \
    head -10
fi

# Nettoyer si on a démarré l'API
if [ "$STARTED_API" = true ] && [ -n "$API_PID" ]; then
    echo ""
    echo "🧹 Nettoyage - Arrêt de l'API..."
    kill $API_PID 2>/dev/null
    wait $API_PID 2>/dev/null
fi

echo ""
echo "🔍 Test terminé"
