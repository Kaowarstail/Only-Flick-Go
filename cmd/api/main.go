package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/config"
	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/Kaowarstail/Only-Flick-Go/internal/routes"
	"github.com/Kaowarstail/Only-Flick-Go/internal/services"
	"github.com/Kaowarstail/Only-Flick-Go/seed"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "github.com/joho/godotenv/autoload"
)

// healthHandler gère l'endpoint de santé
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "OnlyFlick API"}`))
}

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

	// 📊 Initialiser la mise à jour des métriques
	services.InitMetricsUpdater()
	log.Println("📊 Mise à jour des métriques démarrée ✅")

	// Setup du routeur
	router := mux.NewRouter()

	// Middleware de métriques (doit être ajouté avant les routes)
	router.Use(middleware.MetricsMiddleware)

	// Endpoint de santé
	router.HandleFunc("/health", healthHandler).Methods("GET")

	// Endpoint de métriques Prometheus
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Enregistrer les autres routes
	routes.RegisterRoutes(router)

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
