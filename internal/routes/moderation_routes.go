package routes

import (
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/gorilla/mux"
)

// RegisterModerationRoutes enregistre toutes les routes de modération et d'administration
func RegisterModerationRoutes(router *mux.Router) {
	// Routes pour les signalements
	router.Handle("/reports", middleware.JWTAuth(handleCreateReport)).Methods("POST")
	router.Handle("/reports", middleware.JWTAuth(handleGetReports)).Methods("GET")
	router.Handle("/reports/{id}", middleware.JWTAuth(handleGetReport)).Methods("GET")
	router.Handle("/reports/{id}", middleware.JWTAuth(handleProcessReport)).Methods("PUT")

	// Routes pour l'administration (toutes protégées)
	admin := router.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.JWTAuth) // Middleware global pour toutes les routes admin

	admin.HandleFunc("/audit-logs", handleGetAuditLogs).Methods("GET")
	admin.HandleFunc("/users/{id}/ban", handleBanUser).Methods("PUT")
	admin.HandleFunc("/users/{id}/unban", handleUnbanUser).Methods("PUT")
}

// Définitions temporaires des gestionnaires
var (
	// Report handlers
	handleCreateReport = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetReports = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetReport = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleProcessReport = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})

	// Admin handlers
	handleGetAuditLogs = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleBanUser = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleUnbanUser = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
)
