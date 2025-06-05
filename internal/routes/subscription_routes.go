package routes

import (
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/gorilla/mux"
)

// RegisterSubscriptionRoutes enregistre toutes les routes des abonnements
func RegisterSubscriptionRoutes(router *mux.Router) {
	// Routes publiques pour les plans d'abonnement
	router.HandleFunc("/subscription-plans", handleGetSubscriptionPlans).Methods("GET")
	router.HandleFunc("/subscription-plans/{id}", handleGetSubscriptionPlan).Methods("GET")

	// Routes protégées pour les plans d'abonnement
	router.Handle("/subscription-plans", middleware.JWTAuth(handleCreateSubscriptionPlan)).Methods("POST")
	router.Handle("/subscription-plans/{id}", middleware.JWTAuth(handleUpdateSubscriptionPlan)).Methods("PUT")
	router.Handle("/subscription-plans/{id}", middleware.JWTAuth(handleDeleteSubscriptionPlan)).Methods("DELETE")

	// Routes protégées pour les abonnements
	router.Handle("/subscriptions", middleware.JWTAuth(handleGetUserSubscriptions)).Methods("GET")
	router.Handle("/subscriptions", middleware.JWTAuth(handleCreateSubscription)).Methods("POST")
	router.Handle("/subscriptions/{id}", middleware.JWTAuth(handleGetSubscription)).Methods("GET")
	router.Handle("/subscriptions/{id}", middleware.JWTAuth(handleUpdateSubscription)).Methods("PUT")
	router.Handle("/subscriptions/{id}", middleware.JWTAuth(handleDeleteSubscription)).Methods("DELETE")
	router.Handle("/subscriptions/{id}/renew", middleware.JWTAuth(handleRenewSubscription)).Methods("PUT")
}

// Définitions temporaires des gestionnaires
var (
	handleGetSubscriptionPlans = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetSubscriptionPlan = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleCreateSubscriptionPlan = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleUpdateSubscriptionPlan = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleDeleteSubscriptionPlan = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetUserSubscriptions = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleCreateSubscription = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetSubscription = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleUpdateSubscription = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleDeleteSubscription = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleRenewSubscription = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
)
