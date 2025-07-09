package routes

import (
	"github.com/Kaowarstail/Only-Flick-Go/internal/handlers"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes enregistre toutes les routes d'administration
func RegisterAdminRoutes(router *mux.Router) {
	// Groupe de routes d'administration - toutes protégées mais accessibles à tous les utilisateurs
	admin := router.PathPrefix("/admin").Subrouter()

	// Appliquer uniquement le middleware d'authentification (sans restriction admin)
	admin.Use(middleware.JWTAuth)
	// admin.Use(middleware.AdminRequired) // Commenté pour permettre l'accès à tous

	// Dashboard - statistiques générales
	admin.HandleFunc("/dashboard/stats", handlers.GetDashboardStats).Methods("GET", "OPTIONS")

	// Modération - signalements
	admin.HandleFunc("/reports", handlers.GetRecentReports).Methods("GET", "OPTIONS")
	admin.HandleFunc("/reports/update", handlers.UpdateReportStatus).Methods("PUT", "OPTIONS")

	// Gestion des utilisateurs
	admin.HandleFunc("/users", handlers.GetAdminUsers).Methods("GET", "OPTIONS")
	admin.HandleFunc("/users/{id}/role", handlers.UpdateUserRole).Methods("PUT", "OPTIONS")
	admin.HandleFunc("/users/{id}/ban", handlers.BanUser).Methods("PUT", "OPTIONS")
	admin.HandleFunc("/users/{id}/unban", handlers.UnbanUser).Methods("PUT", "OPTIONS")

	// Gestion des contenus
	admin.HandleFunc("/contents", handlers.GetAdminContents).Methods("GET", "OPTIONS")
}
