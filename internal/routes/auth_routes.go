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
	auth.HandleFunc("/reset-password", handlers.RequestPasswordReset).Methods("POST")
	auth.HandleFunc("/reset-password/{token}", handlers.ResetPassword).Methods("PUT")
	auth.HandleFunc("/verify-email/{token}", handlers.VerifyEmail).Methods("GET")

	// Routes protégées d'authentification
	auth.Handle("/me", middleware.JWTAuth(http.HandlerFunc(handlers.GetCurrentUser))).Methods("GET")
	auth.Handle("/resend-verification", middleware.JWTAuth(http.HandlerFunc(handlers.ResendEmailVerification))).Methods("POST")
	auth.Handle("/change-password", middleware.JWTAuth(http.HandlerFunc(handlers.ChangePassword))).Methods("PUT")
}
