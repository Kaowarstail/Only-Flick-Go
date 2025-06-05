package routes

import (
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/gorilla/mux"
)

// RegisterContentRoutes enregistre toutes les routes de gestion du contenu
func RegisterContentRoutes(router *mux.Router) {
	// Routes publiques pour le contenu
	router.HandleFunc("/contents", handleGetContents).Methods("GET")
	router.HandleFunc("/contents/{id}", handleGetContent).Methods("GET")
	router.HandleFunc("/contents/search", handleSearchContents).Methods("GET")
	router.HandleFunc("/contents/trending", handleGetTrendingContents).Methods("GET")
	router.HandleFunc("/contents/{id}/comments", handleGetContentComments).Methods("GET")

	// Routes protégées pour le contenu
	router.Handle("/contents", middleware.JWTAuth(handleCreateContent)).Methods("POST")
	router.Handle("/contents/{id}", middleware.JWTAuth(handleUpdateContent)).Methods("PUT")
	router.Handle("/contents/{id}", middleware.JWTAuth(handleDeleteContent)).Methods("DELETE")
	router.Handle("/contents/{id}/media", middleware.JWTAuth(handleUploadContentMedia)).Methods("POST")
	router.Handle("/contents/{id}/thumbnail", middleware.JWTAuth(handleUploadContentThumbnail)).Methods("POST")
	router.Handle("/contents/{id}/comments", middleware.JWTAuth(handleAddComment)).Methods("POST")
	router.Handle("/contents/{id}/likes", middleware.JWTAuth(handleLikeContent)).Methods("POST")
	router.Handle("/contents/{id}/likes", middleware.JWTAuth(handleUnlikeContent)).Methods("DELETE")

	// Routes pour les commentaires
	router.Handle("/comments/{id}", middleware.JWTAuth(handleUpdateComment)).Methods("PUT")
	router.Handle("/comments/{id}", middleware.JWTAuth(handleDeleteComment)).Methods("DELETE")
}

// Définitions temporaires des gestionnaires
var (
	handleGetContents = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetContent = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleSearchContents = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetTrendingContents = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetContentComments = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetCreatorContents = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleCreateContent = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleUpdateContent = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleDeleteContent = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleUploadContentMedia = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleUploadContentThumbnail = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleAddComment = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleLikeContent = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleUnlikeContent = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleUpdateComment = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleDeleteComment = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
)
