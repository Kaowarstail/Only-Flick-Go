package routes

import (
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/internal/handlers"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/gorilla/mux"
)

// RegisterAuthRoutes enregistre toutes les routes d'authentification
func RegisterAuthRoutes(router *mux.Router) {
	// Groupe de routes d'authentification
	auth := router.PathPrefix("/auth").Subrouter()

	// Routes publiques d'authentification
	auth.HandleFunc("/register", handlers.Register).Methods("POST")
	auth.HandleFunc("/login", handlers.Login).Methods("POST")
	auth.HandleFunc("/logout", handlers.Logout).Methods("POST")
	auth.HandleFunc("/refresh-token", handlers.RefreshToken).Methods("POST")
	auth.HandleFunc("/reset-password", handleRequestResetPassword).Methods("POST")
	auth.HandleFunc("/reset-password/{token}", handleResetPassword).Methods("PUT")

	// Routes protégées d'authentification
	auth.Handle("/me", middleware.JWTAuth(http.HandlerFunc(handlers.GetCurrentUser))).Methods("GET")
}

// Certains gestionnaires sont encore à implémenter
var (
	handleRequestResetPassword = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Réinitialisation de mot de passe non encore implémentée"))
	})
	handleResetPassword = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Réinitialisation de mot de passe non encore implémentée"))
	})
)
