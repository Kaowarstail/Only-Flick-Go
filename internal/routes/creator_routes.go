package routes

import (
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/gorilla/mux"
)

// RegisterCreatorRoutes enregistre toutes les routes de gestion des créateurs
func RegisterCreatorRoutes(router *mux.Router) {
	// Routes publiques pour les créateurs
	router.HandleFunc("/creators", handleGetCreators).Methods("GET")
	router.HandleFunc("/creators/{id}", handleGetCreator).Methods("GET")
	router.HandleFunc("/creators/featured", handleGetFeaturedCreators).Methods("GET")
	router.HandleFunc("/creators/search", handleSearchCreators).Methods("GET")
	router.HandleFunc("/creators/{id}/subscription-plans", handleGetCreatorSubscriptionPlans).Methods("GET")

	// Routes protégées pour les créateurs
	router.Handle("/creators", middleware.JWTAuth(handleBecomeCreator)).Methods("POST")
	router.Handle("/creators/{id}", middleware.JWTAuth(handleUpdateCreator)).Methods("PUT")
	router.Handle("/creators/{id}/banner", middleware.JWTAuth(handleUploadCreatorBanner)).Methods("PUT")
	router.Handle("/creators/{id}/subscribers", middleware.JWTAuth(handleGetCreatorSubscribers)).Methods("GET")
	router.Handle("/creators/{id}/stats", middleware.JWTAuth(handleGetCreatorStats)).Methods("GET")
	router.Handle("/creators/{id}/earnings", middleware.JWTAuth(handleGetCreatorEarnings)).Methods("GET")
}

// Définitions temporaires des gestionnaires
var (
	handleGetCreators = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetCreator = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetFeaturedCreators = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleSearchCreators = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetCreatorSubscriptionPlans = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleBecomeCreator = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleUpdateCreator = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleUploadCreatorBanner = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetCreatorSubscribers = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetCreatorStats = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetCreatorEarnings = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
)
