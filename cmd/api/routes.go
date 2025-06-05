package main

import (
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/internal/handlers"
	"github.com/gorilla/mux"
)

func registerRoutes(router *mux.Router) {
	// Route pour la racine "/"
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Bonjour ! j'ai changé la ci/cd"))
	}).Methods("GET")

	// API versioning
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	// Health check
	apiV1.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Exemple de routes - à adapter selon les besoins
	// Users routes
	apiV1.HandleFunc("/users", handleGetUsers).Methods("GET")
	apiV1.HandleFunc("/users", handleCreateUser).Methods("POST")
	apiV1.HandleFunc("/users/{id}", handleGetUser).Methods("GET")
	apiV1.HandleFunc("/users/{id}", handleUpdateUser).Methods("PUT")
	apiV1.HandleFunc("/users/{id}", handleDeleteUser).Methods("DELETE")

	// Auth routes
	auth := apiV1.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/login", handleLogin).Methods("POST")
	auth.HandleFunc("/logout", handleLogout).Methods("POST")

	// Garder les anciennes routes pour compatibilité
	apiV1.HandleFunc("/login", handleLogin).Methods("POST")
	apiV1.HandleFunc("/logout", handleLogout).Methods("POST")

	// Autres routes...
}

// Pas besoin d'importation supplémentaire, déjà importé en haut du fichier

// Handlers pour les routes
func handleGetUsers(w http.ResponseWriter, r *http.Request) {
	handlers.GetUsers(w, r)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	handlers.CreateUser(w, r)
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	handlers.GetUser(w, r)
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	handlers.UpdateUser(w, r)
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	handlers.DeleteUser(w, r)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	handlers.Login(w, r)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	handlers.Logout(w, r)
}
