package main

import (
	"fmt"
	"log"

	"github.com/Kaowarstail/Only-Flick-Go/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Erreur lors du chargement de la configuration:", err)
	}

	fmt.Printf("=== CONFIGURATION CLOUDINARY ===\n")
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
