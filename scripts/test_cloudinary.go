package main

import (
	"fmt"
	"log"

	"github.com/Kaowarstail/Only-Flick-Go/config"
	"github.com/Kaowarstail/Only-Flick-Go/internal/services"
)

func main() {
	fmt.Println("🧪 Test de configuration Cloudinary")
	fmt.Println("=====================================")

	// Charger la configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Erreur lors du chargement de la configuration:", err)
	}

	// Afficher la configuration (sans les secrets)
	fmt.Printf("📋 Configuration Cloudinary:\n")
	fmt.Printf("   - Cloud Name: %s\n", cfg.Cloudinary.CloudName)
	fmt.Printf("   - API Key: %s***\n", cfg.Cloudinary.APIKey[:4])
	fmt.Printf("   - API Secret: %s*** (longueur: %d)\n", cfg.Cloudinary.APISecret[:4], len(cfg.Cloudinary.APISecret))

	// Tester la création du service Cloudinary
	fmt.Println("\n🔧 Test de création du service Cloudinary...")
	cloudinaryService, err := services.NewCloudinaryService()
	if err != nil {
		log.Fatal("❌ Erreur lors de la création du service Cloudinary:", err)
	}

	fmt.Println("✅ Service Cloudinary créé avec succès!")

	// Tester la génération d'URLs
	fmt.Println("\n🔗 Test de génération d'URLs...")

	// Test avec un PublicID fictif
	testPublicID := "onlyflick/test/sample-image"

	// Générer une URL d'image
	imageURL, err := cloudinaryService.GetImageURL(testPublicID, "w_300,h_300,c_fill,q_auto,f_auto")
	if err != nil {
		fmt.Printf("❌ Erreur lors de la génération de l'URL image: %v\n", err)
	} else {
		fmt.Printf("✅ URL image générée: %s\n", imageURL)
	}

	// Générer une URL de vidéo
	videoURL, err := cloudinaryService.GetVideoURL(testPublicID, "w_640,h_360,c_fill,q_auto")
	if err != nil {
		fmt.Printf("❌ Erreur lors de la génération de l'URL vidéo: %v\n", err)
	} else {
		fmt.Printf("✅ URL vidéo générée: %s\n", videoURL)
	}

	// Générer une URL de miniature
	thumbnailURL, err := cloudinaryService.GetThumbnailURL(testPublicID, "image", 150, 150)
	if err != nil {
		fmt.Printf("❌ Erreur lors de la génération de l'URL miniature: %v\n", err)
	} else {
		fmt.Printf("✅ URL miniature générée: %s\n", thumbnailURL)
	}

	// Tester la validation des types de fichiers
	fmt.Println("\n📝 Test de validation des types de fichiers...")

	testCases := []struct {
		contentType string
		mediaType   string
		filename    string
		expected    bool
	}{
		{"image/jpeg", "image", "test.jpg", true},
		{"image/png", "image", "test.png", true},
		{"video/mp4", "video", "test.mp4", true},
		{"text/plain", "image", "test.txt", false},
		{"application/octet-stream", "image", "test.jpg", true},
		{"application/octet-stream", "video", "test.mp4", true},
	}

	for _, tc := range testCases {
		result := cloudinaryService.ValidateFileType(tc.contentType, tc.mediaType, tc.filename)
		status := "✅"
		if result != tc.expected {
			status = "❌"
		}
		fmt.Printf("%s %s + %s + %s = %v (attendu: %v)\n",
			status, tc.contentType, tc.mediaType, tc.filename, result, tc.expected)
	}

	fmt.Println("\n🎉 Tests terminés!")
	fmt.Println("\n💡 Pour tester l'upload réel, utilisez les endpoints API:")
	fmt.Println("   - POST /api/contents/:id/media")
	fmt.Println("   - POST /api/contents/:id/thumbnail")
	fmt.Println("   - POST /api/users/:id/profile-picture")
	fmt.Println("   - POST /api/creators/:id/banner")
}
