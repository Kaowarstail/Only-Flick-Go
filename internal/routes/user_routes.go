package routes

import (
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/internal/handlers"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/gorilla/mux"
)

// RegisterUserRoutes enregistre toutes les routes de gestion des utilisateurs
func RegisterUserRoutes(router *mux.Router) {
	// Routes publiques pour les utilisateurs
	router.HandleFunc("/users/{id}", handlers.GetUser).Methods("GET")
	router.HandleFunc("/users", handlers.GetUsers).Methods("GET")

	// Routes protégées pour les utilisateurs
	router.Handle("/users/{id}", middleware.JWTAuth(http.HandlerFunc(handlers.UpdateUser))).Methods("PUT")
	router.Handle("/users/{id}", middleware.JWTAuth(http.HandlerFunc(handlers.DeleteUser))).Methods("DELETE")
	router.Handle("/users/{id}/profile-pic", middleware.JWTAuth(http.HandlerFunc(handlers.UploadProfilePicture))).Methods("PUT")
	router.Handle("/users/{id}/following", middleware.JWTAuth(http.HandlerFunc(handlers.GetFollowing))).Methods("GET")
	router.Handle("/users/{id}/block/{targetId}", middleware.JWTAuth(http.HandlerFunc(handlers.BlockUser))).Methods("POST")
	router.Handle("/users/{id}/block/{targetId}", middleware.JWTAuth(http.HandlerFunc(handlers.UnblockUser))).Methods("DELETE")
	router.Handle("/users/{id}/blocked", middleware.JWTAuth(http.HandlerFunc(handlers.GetBlockedUsers))).Methods("GET")
	router.Handle("/users/{id}/notification-settings", middleware.JWTAuth(http.HandlerFunc(handlers.UpdateNotificationSettings))).Methods("PUT")
	router.Handle("/users/{id}/feed", middleware.JWTAuth(http.HandlerFunc(handlers.GetFeed))).Methods("GET")
}

// Les gestionnaires sont maintenant implémentés dans internal/handlers/profile.go
