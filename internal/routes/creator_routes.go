package routes

import (
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/handlers"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/gorilla/mux"
)

// RegisterCreatorRoutes enregistre toutes les routes de gestion des créateurs
func RegisterCreatorRoutes(router *mux.Router) {
	// Créer l'instance du handler des abonnements pour la route des plans
	subscriptionHandler := handlers.NewSubscriptionHandler(database.DB)

	// Routes publiques pour les créateurs
	router.HandleFunc("/creators", handlers.GetCreators).Methods("GET", "OPTIONS")
	router.HandleFunc("/creators/{id}", handlers.GetCreator).Methods("GET", "OPTIONS")
	router.HandleFunc("/creators/{id}/contents", handlers.GetCreatorContents).Methods("GET", "OPTIONS")
	router.HandleFunc("/creators/featured", handlers.GetFeaturedCreators).Methods("GET", "OPTIONS")
	router.HandleFunc("/creators/search", handlers.SearchCreators).Methods("GET", "OPTIONS")
	router.HandleFunc("/creators/{id}/subscription-plans", subscriptionHandler.GetCreatorSubscriptionPlans).Methods("GET", "OPTIONS")

	// Routes protégées pour les créateurs
	router.Handle("/creators/become", middleware.JWTAuth(http.HandlerFunc(handlers.BecomeCreator))).Methods("POST", "OPTIONS")
	router.Handle("/creators/{id}", middleware.JWTAuth(http.HandlerFunc(handlers.UpdateCreator))).Methods("PUT", "OPTIONS")
	router.Handle("/creators/{id}/banner", middleware.JWTAuth(http.HandlerFunc(handlers.UploadBannerImage))).Methods("PUT", "OPTIONS")
	router.Handle("/creators/{id}/subscribers", middleware.JWTAuth(handleGetCreatorSubscribers)).Methods("GET", "OPTIONS")
	router.Handle("/creators/{id}/stats", middleware.JWTAuth(handleGetCreatorStats)).Methods("GET", "OPTIONS")
}

// Définitions temporaires des gestionnaires
var (
	handleGetCreatorSubscribers = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetCreatorStats = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
)
