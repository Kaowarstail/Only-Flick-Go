#!/bin/bash

echo "🚀 OnlyFlick - PostgreSQL Local Setup"
echo "======================================"

# Vérifier si PostgreSQL est déjà installé
if command -v psql &> /dev/null; then
    echo "✅ PostgreSQL est déjà installé"
    psql --version
else
    echo "📦 Installation de PostgreSQL..."
    echo "   Veuillez télécharger et installer PostgreSQL depuis:"
    echo "   https://www.postgresql.org/download/windows/"
    echo ""
    echo "   Ou utilisez chocolatey si installé:"
    echo "   choco install postgresql"
    echo ""
    echo "   Ou utilisez winget:"
    echo "   winget install PostgreSQL.PostgreSQL"
    exit 1
fi

# Configuration de la base de données OnlyFlick
echo ""
echo "🔧 Configuration de la base de données OnlyFlick..."

# Paramètres de base de données
DB_NAME="onlyflick_db"
DB_USER="onlyflick_user"
DB_PASSWORD="onlyflick123"

echo "   Base de données: $DB_NAME"
echo "   Utilisateur: $DB_USER"
echo "   Mot de passe: $DB_PASSWORD"

# Créer la base de données et l'utilisateur
echo ""
echo "📊 Création de la base de données..."

# Se connecter en tant que postgres et créer la base
psql -U postgres -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';" 2>/dev/null || echo "   Utilisateur existe déjà"
psql -U postgres -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;" 2>/dev/null || echo "   Base de données existe déjà"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;" 2>/dev/null

# Tester la connexion
echo ""
echo "🔍 Test de connexion..."
if psql -U $DB_USER -d $DB_NAME -c "SELECT version();" &> /dev/null; then
    echo "✅ Connexion réussie!"
    
    # Mettre à jour le fichier .env
    echo ""
    echo "📝 Mise à jour du fichier .env..."
    
    cat > .env << EOF
# Configuration OnlyFlick - PostgreSQL Local
DATABASE_URL=postgresql://$DB_USER:$DB_PASSWORD@localhost:5432/$DB_NAME
PORT=8080

# Base de données PostgreSQL locale
DB_TYPE=postgresql
DB_HOST=localhost
DB_PORT=5432
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD
DB_NAME=$DB_NAME
DB_SSLMODE=disable

JWT_SECRET=onlyflick-super-secret-key-2024-messaging-system
JWT_EXPIRATION=24

# Cloudinary Configuration
CLOUDINARY_CLOUD_NAME=dafiqfwsf
CLOUDINARY_API_KEY=491423787639739
CLOUDINARY_API_SECRET=Sg2N_T7Zq63V49fMCh-oO52AefE

CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:49219,http://127.0.0.1:3000,http://127.0.0.1:49219,http://127.0.0.1:35699,http://localhost:35699
CORS_ALLOW_CREDENTIALS=true
CORS_MAX_AGE=86400
EOF

    echo "✅ Fichier .env mis à jour avec la configuration PostgreSQL locale"
    
    echo ""
    echo "🎉 Configuration terminée!"
    echo "   Vous pouvez maintenant tester votre système de messagerie avec:"
    echo "   go run cmd/test-db/main.go"
    
else
    echo "❌ Échec de la connexion. Vérifiez que PostgreSQL est démarré."
    echo "   Commandes utiles:"
    echo "   - Démarrer PostgreSQL: net start postgresql-x64-15"
    echo "   - Arrêter PostgreSQL: net stop postgresql-x64-15"
fi

echo ""
echo "📖 Informations de connexion:"
echo "   Host: localhost:5432"
echo "   Database: $DB_NAME"
echo "   Username: $DB_USER"
echo "   Password: $DB_PASSWORD"
