#!/bin/bash

# Script pour configurer Cloudinary avec des valeurs de test
# Ce script configure temporairement Cloudinary avec des valeurs de démonstration

echo "🔧 Configuration temporaire de Cloudinary pour le développement"
echo "==============================================================="
echo ""

# Pour le développement, nous allons utiliser des valeurs par défaut
# Ces valeurs ne fonctionneront pas avec le vrai Cloudinary, mais permettront de tester la structure
echo "📝 Mise à jour du config.json avec des valeurs de test..."

# Utiliser des valeurs temporaires
CLOUD_NAME="demo-cloud"
API_KEY="123456789012345"
API_SECRET="test-secret-key"

# Mettre à jour le config.json
cat > ../config.json << EOF
{
  "server": {
    "port": "8080",
    "timeout": 15
  },
  "database": {
    "driver": "sqlite",
    "path": "dev_database.db"
  },
  "jwt": {
    "secret": "your-secret-key",
    "expiration": 24
  },
  "cloudinary": {
    "cloud_name": "$CLOUD_NAME",
    "api_key": "$API_KEY",
    "api_secret": "$API_SECRET"
  }
}
EOF

echo "✅ Configuration temporaire mise à jour"
echo ""
echo "⚠️  ATTENTION: Ces valeurs sont temporaires et ne fonctionneront pas avec le vrai Cloudinary"
echo "📋 Pour utiliser Cloudinary en production, exécutez: ./setup_cloudinary.sh"
echo ""
echo "Pour tester maintenant:"
echo "1. Redémarrez votre serveur: cd .. && go run cmd/api/main.go"
echo "2. Lancez le test: ./test_image_upload.sh"
