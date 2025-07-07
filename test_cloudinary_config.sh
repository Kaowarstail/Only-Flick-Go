#!/bin/bash

# Script de test pour vérifier la configuration Cloudinary

echo "=== TEST DE CONFIGURATION CLOUDINARY ==="
echo

# Compilation du backend
echo "1. Compilation du backend..."
cd /Users/ilan/Documents/1_EEMI/Projet-RNCP/onlyflick/Only-Flick-Go
go build -o server cmd/api/main.go

if [ $? -ne 0 ]; then
    echo "❌ Erreur lors de la compilation"
    exit 1
fi

echo "✅ Compilation réussie"

# Test de la configuration
echo
echo "2. Test de la configuration Cloudinary..."

# Créer un fichier de test temporaire pour vérifier la config
cat > test_config.go << 'EOF'
package main

import (
    "fmt"
    "log"
    "./config"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Erreur lors du chargement de la configuration:", err)
    }

    fmt.Printf("Cloud Name: %s\n", cfg.Cloudinary.CloudName)
    fmt.Printf("API Key: %s\n", cfg.Cloudinary.APIKey)
    fmt.Printf("API Secret: %s\n", cfg.Cloudinary.APISecret)
    
    // Vérifier si les valeurs sont correctes
    if cfg.Cloudinary.CloudName == "your-cloud-name" || 
       cfg.Cloudinary.APIKey == "your-api-key" || 
       cfg.Cloudinary.APISecret == "your-api-secret" {
        fmt.Println("❌ Configuration Cloudinary par défaut détectée")
    } else {
        fmt.Println("✅ Configuration Cloudinary chargée depuis config.json")
    }
}
EOF

go run test_config.go

# Nettoyer
rm test_config.go

echo
echo "3. Test d'upload avec l'API..."

# Démarrer le serveur en arrière-plan
./server &
SERVER_PID=$!

# Attendre que le serveur démarre
sleep 3

# Tester l'endpoint d'upload
curl -X POST \
  -H "Content-Type: multipart/form-data" \
  -F "image=@test_image.jpg" \
  -F "title=Test Upload Config" \
  -F "description=Test de configuration Cloudinary" \
  -F "content_type=image" \
  http://localhost:8080/api/contents/upload-image

echo
echo

# Arrêter le serveur
kill $SERVER_PID

echo "=== FIN DU TEST ==="
