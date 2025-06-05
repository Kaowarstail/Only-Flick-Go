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
	router.Handle("/users/{id}/profile-pic", middleware.JWTAuth(handleUploadProfilePicture)).Methods("PUT")
	router.Handle("/users/{id}/following", middleware.JWTAuth(handleGetFollowing)).Methods("GET")
	router.Handle("/users/{id}/block/{targetId}", middleware.JWTAuth(handleBlockUser)).Methods("POST")
	router.Handle("/users/{id}/block/{targetId}", middleware.JWTAuth(handleUnblockUser)).Methods("DELETE")
	router.Handle("/users/{id}/blocked", middleware.JWTAuth(handleGetBlockedUsers)).Methods("GET")
	router.Handle("/users/{id}/notification-settings", middleware.JWTAuth(handleUpdateNotificationSettings)).Methods("PUT")
	router.Handle("/users/{id}/feed", middleware.JWTAuth(handleGetFeed)).Methods("GET")
}

// Définitions temporaires des gestionnaires
var (
	handleUploadProfilePicture = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetFollowing = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleBlockUser = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleUnblockUser = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetBlockedUsers = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleUpdateNotificationSettings = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetFeed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
)
