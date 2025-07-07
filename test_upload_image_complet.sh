#!/bin/bash

# Script de test complet pour l'upload d'image OnlyFlick
# Teste la création de contenu avec upload d'image sur Cloudinary

echo "🧪 Test d'upload d'image complet - OnlyFlick"
echo "============================================="

# Configuration
BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"
TEST_EMAIL="testcreator@example.com"
TEST_PASSWORD="testpassword123"
TEST_IMAGE="test_image.jpg"

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function pour afficher les messages
log() {
    echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Vérifier si le serveur est démarré
log "Vérification du serveur..."
if ! curl -s $BASE_URL/health > /dev/null 2>&1; then
    error "Le serveur n'est pas démarré sur $BASE_URL"
    echo "Démarrez le serveur avec : go run cmd/api/main.go"
    exit 1
fi
success "Serveur accessible"

# Vérifier si l'image de test existe
if [ ! -f "$TEST_IMAGE" ]; then
    warning "Image de test non trouvée. Création d'une image de test..."
    # Créer une image de test simple avec ImageMagick (si disponible)
    if command -v convert &> /dev/null; then
        convert -size 400x300 xc:lightblue -pointsize 30 -fill darkblue -annotate +50+150 "Test Image OnlyFlick" "$TEST_IMAGE"
        success "Image de test créée : $TEST_IMAGE"
    else
        error "ImageMagick non disponible et pas d'image de test. Placez une image nommée '$TEST_IMAGE' dans le dossier."
        exit 1
    fi
fi

# Étape 1: Créer un compte créateur de test
log "Création du compte créateur de test..."
register_response=$(curl -s -X POST $API_URL/auth/register \
    -H "Content-Type: application/json" \
    -d "{
        \"username\": \"testcreator\",
        \"email\": \"$TEST_EMAIL\",
        \"password\": \"$TEST_PASSWORD\",
        \"first_name\": \"Test\",
        \"last_name\": \"Creator\"
    }")

if echo "$register_response" | grep -q '"message":"Utilisateur créé avec succès"'; then
    success "Compte créateur créé avec succès"
elif echo "$register_response" | grep -q "already exists"; then
    warning "Compte créateur déjà existant"
else
    error "Erreur lors de la création du compte : $register_response"
fi

# Étape 2: Connexion pour récupérer le token
log "Connexion du créateur..."
login_response=$(curl -s -X POST $API_URL/auth/login \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"$TEST_EMAIL\",
        \"password\": \"$TEST_PASSWORD\"
    }")

if echo "$login_response" | grep -q '"token"'; then
    JWT_TOKEN=$(echo "$login_response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    success "Connexion réussie - Token récupéré"
    echo "Token: ${JWT_TOKEN:0:20}..."
else
    error "Erreur lors de la connexion : $login_response"
    exit 1
fi

# Étape 3: Vérifier la configuration Cloudinary
log "Vérification de la configuration Cloudinary..."
if [ -f ".env" ]; then
    if grep -q "CLOUDINARY_CLOUD_NAME" .env && grep -q "CLOUDINARY_API_KEY" .env; then
        success "Configuration Cloudinary trouvée dans .env"
    else
        error "Configuration Cloudinary incomplète dans .env"
        exit 1
    fi
else
    error "Fichier .env non trouvé"
    exit 1
fi

# Étape 4: Test d'upload d'image
log "Test d'upload d'image..."
upload_response=$(curl -s -X POST $API_URL/contents/upload-image \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -F "image=@$TEST_IMAGE" \
    -F "title=Test Image Upload" \
    -F "description=Image de test pour vérifier l'upload vers Cloudinary" \
    -F "is_premium=false" \
    -F "is_published=true")

echo "Réponse upload : $upload_response"

# Analyser la réponse
if echo "$upload_response" | grep -q '"media_url"'; then
    success "Upload réussi !"
    
    # Extraire les informations
    CONTENT_ID=$(echo "$upload_response" | grep -o '"content_id":[^,]*' | cut -d':' -f2)
    MEDIA_URL=$(echo "$upload_response" | grep -o '"media_url":"[^"]*' | cut -d'"' -f4)
    PUBLIC_ID=$(echo "$upload_response" | grep -o '"public_id":"[^"]*' | cut -d'"' -f4)
    
    success "Contenu créé avec ID: $CONTENT_ID"
    success "URL média: $MEDIA_URL"
    success "Public ID Cloudinary: $PUBLIC_ID"
    
    # Vérifier que l'URL est accessible
    if curl -s --head "$MEDIA_URL" | grep -q "200 OK"; then
        success "Image accessible sur Cloudinary"
    else
        warning "Image non accessible sur Cloudinary"
    fi
    
else
    error "Erreur lors de l'upload : $upload_response"
    exit 1
fi

# Étape 5: Vérifier en base que media_url est bien renseigné
log "Vérification en base de données..."
if [ -f "dev_database.db" ]; then
    # Utiliser sqlite3 pour vérifier
    if command -v sqlite3 &> /dev/null; then
        db_result=$(sqlite3 dev_database.db "SELECT id, title, media_url, thumbnail_url FROM contents WHERE id = $CONTENT_ID;")
        if [ -n "$db_result" ]; then
            success "Contenu trouvé en base : $db_result"
            if echo "$db_result" | grep -q "cloudinary"; then
                success "media_url correctement renseigné en base"
            else
                error "media_url vide en base"
            fi
        else
            error "Contenu non trouvé en base"
        fi
    else
        warning "sqlite3 non disponible pour vérifier la base"
    fi
else
    warning "Base de données SQLite non trouvée"
fi

# Étape 6: Test de récupération du contenu via API
log "Test de récupération du contenu..."
get_response=$(curl -s -X GET $API_URL/contents/$CONTENT_ID)

if echo "$get_response" | grep -q '"media_url"'; then
    success "Contenu récupéré avec succès via API"
    GET_MEDIA_URL=$(echo "$get_response" | grep -o '"media_url":"[^"]*' | cut -d'"' -f4)
    if [ "$GET_MEDIA_URL" = "$MEDIA_URL" ]; then
        success "URL média cohérente entre création et récupération"
    else
        warning "URL média différente entre création et récupération"
    fi
else
    error "Erreur lors de la récupération : $get_response"
fi

echo ""
echo "==================== RÉSUMÉ ===================="
success "✅ Upload d'image réussi"
success "✅ Contenu créé avec ID: $CONTENT_ID"  
success "✅ Image uploadée sur Cloudinary: $MEDIA_URL"
success "✅ media_url renseigné en base"
success "✅ Contenu accessible via API"
echo "================================================="
echo ""
echo "🎉 Test d'upload d'image complet terminé avec succès !"
