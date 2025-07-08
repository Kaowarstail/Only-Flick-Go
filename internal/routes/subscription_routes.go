package routes

import (
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/handlers"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/gorilla/mux"
)

// RegisterSubscriptionRoutes enregistre toutes les routes des abonnements
func RegisterSubscriptionRoutes(router *mux.Router) {
	// Créer l'instance du handler
	subscriptionHandler := handlers.NewSubscriptionHandler(database.DB)

	// Routes publiques pour les plans d'abonnement
	router.HandleFunc("/subscription-plans", subscriptionHandler.GetSubscriptionPlans).Methods("GET", "OPTIONS")
	router.HandleFunc("/subscription-plans/{id}", subscriptionHandler.GetSubscriptionPlan).Methods("GET", "OPTIONS")

	// Routes protégées pour les plans d'abonnement (Créateurs)
	router.Handle("/subscription-plans", middleware.JWTAuth(http.HandlerFunc(subscriptionHandler.CreateSubscriptionPlan))).Methods("POST", "OPTIONS")
	router.Handle("/subscription-plans/{id}", middleware.JWTAuth(http.HandlerFunc(subscriptionHandler.UpdateSubscriptionPlan))).Methods("PUT", "OPTIONS")
	router.Handle("/subscription-plans/{id}", middleware.JWTAuth(http.HandlerFunc(subscriptionHandler.DeleteSubscriptionPlan))).Methods("DELETE", "OPTIONS")

	// Routes protégées pour les abonnements (Utilisateurs authentifiés)
	router.Handle("/subscriptions", middleware.JWTAuth(http.HandlerFunc(subscriptionHandler.GetUserSubscriptions))).Methods("GET", "OPTIONS")
	router.Handle("/subscriptions", middleware.JWTAuth(http.HandlerFunc(subscriptionHandler.CreateSubscription))).Methods("POST", "OPTIONS")
	router.Handle("/subscriptions/{id}", middleware.JWTAuth(http.HandlerFunc(subscriptionHandler.GetSubscription))).Methods("GET", "OPTIONS")
	router.Handle("/subscriptions/{id}", middleware.JWTAuth(http.HandlerFunc(subscriptionHandler.CancelSubscription))).Methods("DELETE", "OPTIONS")
	router.Handle("/subscriptions/{id}/renew", middleware.JWTAuth(http.HandlerFunc(subscriptionHandler.RenewSubscription))).Methods("PUT", "OPTIONS")
}
