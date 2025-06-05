package routes

import (
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/gorilla/mux"
)

// RegisterRoutes configure toutes les routes de l'API
func RegisterRoutes(router *mux.Router) {
	// Route pour la racine "/"
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Bienvenue sur l'API OnlyFlick"))
	}).Methods("GET")

	// API versioning
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	// Middleware global pour la journalisation
	apiV1.Use(middleware.Logger)

	// Health check
	apiV1.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

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
