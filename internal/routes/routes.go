package routes

import (
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/gorilla/mux"
)

// RegisterRoutes configure toutes les routes de l'API
func RegisterRoutes(router *mux.Router) {
	// Middleware CORS global pour toutes les routes (DOIT être en premier)
	router.Use(middleware.CORS)

	// Route pour la racine "/"
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Bienvenue sur l'API OnlyFlick"))
	}).Methods("GET", "OPTIONS")

	// API versioning
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	// Middleware CORS pour les routes API aussi (sécurité)
	apiV1.Use(middleware.CORS)
	
	// Middleware global pour la journalisation
	apiV1.Use(middleware.Logger)

	// Health check
	apiV1.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET", "OPTIONS")

	// Enregistrement des différentes catégories de routes
	RegisterAuthRoutes(apiV1)
	RegisterUserRoutes(apiV1)
	RegisterCreatorRoutes(apiV1)
	RegisterContentRoutes(apiV1)
	RegisterSubscriptionRoutes(apiV1)
	RegisterCommunicationRoutes(apiV1)
	RegisterModerationRoutes(apiV1)
	RegisterPaymentRoutes(apiV1)
}
