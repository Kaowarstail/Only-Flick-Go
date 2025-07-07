#!/bin/bash

# Test de configuration Cloudinary pour OnlyFlick

echo "🧪 Test de configuration Cloudinary - OnlyFlick"
echo "================================================"

# Vérifier si les variables d'environnement sont définies
if [ -z "$CLOUDINARY_CLOUD_NAME" ]; then
    echo "❌ CLOUDINARY_CLOUD_NAME n'est pas défini"
    echo "💡 Définissez vos variables d'environnement Cloudinary:"
    echo "   export CLOUDINARY_CLOUD_NAME=your-cloud-name"
    echo "   export CLOUDINARY_API_KEY=your-api-key"
    echo "   export CLOUDINARY_API_SECRET=your-api-secret"
    exit 1
fi

if [ -z "$CLOUDINARY_API_KEY" ]; then
    echo "❌ CLOUDINARY_API_KEY n'est pas défini"
    exit 1
fi

if [ -z "$CLOUDINARY_API_SECRET" ]; then
    echo "❌ CLOUDINARY_API_SECRET n'est pas défini"
    exit 1
fi

echo "✅ Variables d'environnement Cloudinary détectées"

# Exécuter le test
echo "🚀 Exécution du test..."
cd "$(dirname "$0")/.."
go run scripts/test_cloudinary.go

echo ""
echo "📋 Pour tester l'upload réel, vous pouvez utiliser curl:"
echo ""
echo "# Upload d'une image de contenu"
echo "curl -X POST http://localhost:8080/api/contents/1/media \\"
echo "  -H 'Authorization: Bearer YOUR_JWT_TOKEN' \\"
echo "  -F 'media=@/path/to/your/image.jpg'"
echo ""
echo "# Upload d'une photo de profil"
echo "curl -X POST http://localhost:8080/api/users/USER_ID/profile-picture \\"
echo "  -H 'Authorization: Bearer YOUR_JWT_TOKEN' \\"
echo "  -F 'profile_picture=@/path/to/your/profile.jpg'"
echo ""
echo "# Upload d'une bannière de créateur"
echo "curl -X POST http://localhost:8080/api/creators/USER_ID/banner \\"
echo "  -H 'Authorization: Bearer YOUR_JWT_TOKEN' \\"
echo "  -F 'banner_image=@/path/to/your/banner.jpg'"
