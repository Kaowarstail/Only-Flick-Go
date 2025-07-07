#!/bin/bash

# Script pour configurer Cloudinary avec OnlyFlick
# Ce script vous aide à configurer Cloudinary pour votre projet OnlyFlick

echo "🌟 Configuration de Cloudinary pour OnlyFlick"
echo "=============================================="
echo ""

echo "Pour configurer Cloudinary, vous devez:"
echo "1. Créer un compte gratuit sur https://cloudinary.com"
echo "2. Aller dans votre Dashboard Cloudinary"
echo "3. Copier vos informations d'identification"
echo ""

echo "Vos informations Cloudinary se trouvent dans le Dashboard:"
echo "- Cloud Name: Nom de votre cloud"
echo "- API Key: Clé API"
echo "- API Secret: Secret API"
echo ""

echo "Entrez vos informations Cloudinary:"
echo ""

read -p "Cloud Name: " cloud_name
read -p "API Key: " api_key
read -s -p "API Secret: " api_secret
echo ""

# Créer un fichier .env pour Cloudinary
cat > ../.env.cloudinary << EOF
CLOUDINARY_CLOUD_NAME=$cloud_name
CLOUDINARY_API_KEY=$api_key
CLOUDINARY_API_SECRET=$api_secret
EOF

echo ""
echo "✅ Configuration sauvegardée dans .env.cloudinary"
echo ""

# Mettre à jour le config.json
echo "📝 Mise à jour du config.json..."

# Utiliser jq pour mettre à jour le fichier JSON
jq --arg cloud_name "$cloud_name" \
   --arg api_key "$api_key" \
   --arg api_secret "$api_secret" \
   '.cloudinary.cloud_name = $cloud_name | .cloudinary.api_key = $api_key | .cloudinary.api_secret = $api_secret' \
   ../config.json > ../config.json.tmp && mv ../config.json.tmp ../config.json

echo "✅ config.json mis à jour avec succès"
echo ""

echo "🎉 Configuration terminée !"
echo "Vous pouvez maintenant utiliser l'upload d'images avec Cloudinary."
echo ""
echo "Pour tester l'upload, redémarrez votre serveur Go:"
echo "cd /Users/ilan/Documents/1_EEMI/Projet-RNCP/onlyflick/Only-Flick-Go && go run cmd/api/main.go"
