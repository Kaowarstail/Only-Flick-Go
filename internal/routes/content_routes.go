package routes

import (
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/internal/handlers"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/gorilla/mux"
)

// RegisterContentRoutes enregistre toutes les routes de gestion du contenu
func RegisterContentRoutes(router *mux.Router) {
	// Routes publiques pour le contenu
	router.HandleFunc("/contents", handlers.GetContents).Methods("GET")
	router.HandleFunc("/contents/{id}", handlers.GetContent).Methods("GET")
	router.HandleFunc("/contents/search", handlers.SearchContents).Methods("GET")
	router.HandleFunc("/contents/trending", handlers.GetTrendingContents).Methods("GET")
	router.HandleFunc("/contents/{id}/comments", handlers.GetContentComments).Methods("GET")

	// Routes protégées pour le contenu
	router.Handle("/contents", middleware.JWTAuth(http.HandlerFunc(handlers.CreateContent))).Methods("POST", "OPTIONS")
	router.Handle("/contents/{id}", middleware.JWTAuth(http.HandlerFunc(handlers.UpdateContent))).Methods("PUT", "OPTIONS")
	router.Handle("/contents/{id}", middleware.JWTAuth(http.HandlerFunc(handlers.DeleteContent))).Methods("DELETE", "OPTIONS")
	router.Handle("/contents/{id}/media", middleware.JWTAuth(http.HandlerFunc(handlers.UploadContentMedia))).Methods("POST", "OPTIONS")
	router.Handle("/contents/{id}/thumbnail", middleware.JWTAuth(http.HandlerFunc(handlers.UploadContentThumbnail))).Methods("POST", "OPTIONS")
	router.Handle("/contents/{id}/comments", middleware.JWTAuth(http.HandlerFunc(handlers.AddComment))).Methods("POST", "OPTIONS")
	router.Handle("/contents/{id}/likes", middleware.JWTAuth(http.HandlerFunc(handlers.LikeContent))).Methods("POST", "OPTIONS")
	router.Handle("/contents/{id}/likes", middleware.JWTAuth(http.HandlerFunc(handlers.UnlikeContent))).Methods("DELETE", "OPTIONS")
	router.Handle("/contents/{id}/toggle-like", middleware.JWTAuth(http.HandlerFunc(handlers.ToggleLike))).Methods("POST", "OPTIONS")

	// Routes pour les commentaires
	router.Handle("/comments/{id}", middleware.JWTAuth(http.HandlerFunc(handlers.UpdateComment))).Methods("PUT", "OPTIONS")
	router.Handle("/comments/{id}", middleware.JWTAuth(http.HandlerFunc(handlers.DeleteComment))).Methods("DELETE", "OPTIONS")
}
