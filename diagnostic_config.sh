#!/bin/bash

echo "🔍 Diagnostic de la configuration Cloudinary"
echo "================================================"

# Vérifier le répertoire de travail
echo "📁 Répertoire de travail actuel:"
pwd

# Vérifier l'existence du fichier config.json
echo ""
echo "📄 Existence du fichier config.json:"
if [ -f "config.json" ]; then
    echo "✅ config.json existe"
    echo "📄 Contenu de la section Cloudinary:"
    cat config.json | grep -A 5 '"cloudinary"'
else
    echo "❌ config.json n'existe pas dans le répertoire courant"
fi

# Naviguer vers le répertoire de l'API
cd /Users/ilan/Documents/1_EEMI/Projet-RNCP/onlyflick/Only-Flick-Go

echo ""
echo "📁 Répertoire de l'API Go:"
pwd

# Vérifier l'existence du fichier config.json dans le répertoire API
echo ""
echo "📄 Existence du fichier config.json dans le répertoire API:"
if [ -f "config.json" ]; then
    echo "✅ config.json existe dans le répertoire API"
    echo "📄 Contenu de la section Cloudinary:"
    cat config.json | grep -A 5 '"cloudinary"'
else
    echo "❌ config.json n'existe pas dans le répertoire API"
fi

# Vérifier les variables d'environnement Cloudinary
echo ""
echo "🌍 Variables d'environnement Cloudinary:"
echo "CLOUDINARY_CLOUD_NAME: ${CLOUDINARY_CLOUD_NAME:-'non définie'}"
echo "CLOUDINARY_API_KEY: ${CLOUDINARY_API_KEY:-'non définie'}"
echo "CLOUDINARY_API_SECRET: ${CLOUDINARY_API_SECRET:+***définie***}"

# Compiler et tester une application simple pour vérifier la config
echo ""
echo "🔧 Test de chargement de la configuration:"

# Créer un fichier de test temporaire
cat > config_diagnostic.go << 'EOF'
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/Kaowarstail/Only-Flick-Go/config"
)

func main() {
	fmt.Println("🔧 Test de chargement de la configuration Cloudinary")
	fmt.Println("================================================")
	
	// Afficher le répertoire de travail
	wd, _ := os.Getwd()
	fmt.Printf("📁 Répertoire de travail: %s\n", wd)
	
	// Vérifier l'existence du fichier config.json
	configPath := filepath.Join(wd, "config.json")
	if _, err := os.Stat(configPath); err != nil {
		fmt.Printf("❌ config.json non trouvé à: %s\n", configPath)
	} else {
		fmt.Printf("✅ config.json trouvé à: %s\n", configPath)
	}
	
	// Charger la configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Erreur lors du chargement de la configuration: %v\n", err)
		return
	}
	
	// Afficher la configuration Cloudinary
	fmt.Printf("🔧 Configuration Cloudinary chargée:\n")
	fmt.Printf("  CloudName: %s\n", cfg.Cloudinary.CloudName)
	fmt.Printf("  APIKey: %s\n", cfg.Cloudinary.APIKey)
	fmt.Printf("  APISecret: %s\n", cfg.Cloudinary.APISecret)
	
	// Vérifier si les valeurs sont par défaut (fallback)
	if cfg.Cloudinary.CloudName == "your-cloud-name" {
		fmt.Printf("⚠️  CloudName est la valeur par défaut - config.json non chargé\n")
	}
	if cfg.Cloudinary.APIKey == "your-api-key" {
		fmt.Printf("⚠️  APIKey est la valeur par défaut - config.json non chargé\n")
	}
	if cfg.Cloudinary.APISecret == "your-api-secret" {
		fmt.Printf("⚠️  APISecret est la valeur par défaut - config.json non chargé\n")
	}
	
	// Vérifier si les vraies valeurs sont chargées
	if cfg.Cloudinary.CloudName == "dafiqfwsf" {
		fmt.Printf("✅ CloudName correct depuis config.json\n")
	}
	if cfg.Cloudinary.APIKey == "491423787639739" {
		fmt.Printf("✅ APIKey correct depuis config.json\n")
	}
	if cfg.Cloudinary.APISecret == "Sg2N_T7Zq63V49fMCh-oO52AefE" {
		fmt.Printf("✅ APISecret correct depuis config.json\n")
	}
}
EOF

# Compiler et exécuter
echo "🔧 Compilation du test..."
go run config_diagnostic.go

# Nettoyer
rm config_diagnostic.go

echo ""
echo "🔍 Diagnostic terminé"
