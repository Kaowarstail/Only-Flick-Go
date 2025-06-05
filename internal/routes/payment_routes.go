package routes

import (
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/gorilla/mux"
)

// RegisterPaymentRoutes enregistre toutes les routes de paiement
func RegisterPaymentRoutes(router *mux.Router) {
	// Routes pour les méthodes de paiement
	router.Handle("/payments/methods", middleware.JWTAuth(handleGetPaymentMethods)).Methods("GET")
	router.Handle("/payments/methods", middleware.JWTAuth(handleAddPaymentMethod)).Methods("POST")
	router.Handle("/payments/methods/{id}", middleware.JWTAuth(handleDeletePaymentMethod)).Methods("DELETE")

	// Routes pour les transactions
	router.Handle("/transactions", middleware.JWTAuth(handleGetTransactions)).Methods("GET")
	router.Handle("/transactions/{id}", middleware.JWTAuth(handleGetTransaction)).Methods("GET")

	// Routes pour les versements
	router.Handle("/payouts/request", middleware.JWTAuth(handleRequestPayout)).Methods("POST")
	router.Handle("/payouts", middleware.JWTAuth(handleGetPayouts)).Methods("GET")
	router.Handle("/payouts/{id}", middleware.JWTAuth(handleGetPayout)).Methods("GET")
}

// Définitions temporaires des gestionnaires
var (
	// Payment method handlers
	handleGetPaymentMethods = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleAddPaymentMethod = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleDeletePaymentMethod = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})

	// Transaction handlers
	handleGetTransactions = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetTransaction = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})

	// Payout handlers
	handleRequestPayout = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetPayouts = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetPayout = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
)
