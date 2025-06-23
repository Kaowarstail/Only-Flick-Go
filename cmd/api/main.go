package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/config"
	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/seed"
	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// Chargement de la configuration
	_, err := config.Load()
	if err != nil {
		log.Fatal("❌ Erreur de chargement de la configuration :", err)
	}

	// Connexion à la base de données
	if err := database.Initialize(); err != nil {
		log.Fatal("❌ Erreur d'initialisation de la base de données :", err)
	}
	log.Println("📦 Connexion à la base de données réussie ✅")

	// 🌱 Optionnel : seed des données si SEED=true
	if os.Getenv("SEED") == "true" {
		seed.Run()
	}

	// Setup du routeur
	router := mux.NewRouter()
	registerRoutes(router)

	// Démarrage du serveur
	port := config.Get().Server.Port
	addr := ":" + port

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀 Serveur lancé sur http://localhost%s\n", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Impossible de démarrer le serveur : %v", err)
	}
}
