package routes

import (
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/internal/handlers"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/gorilla/mux"
)

// RegisterMessagingRoutes configure les routes pour la messagerie
func RegisterMessagingRoutes(router *mux.Router) {
	// Créer les handlers
	conversationHandler := handlers.NewConversationHandler()
	messageHandler := handlers.NewMessageHandler()
	mediaHandler := handlers.NewMediaHandler("uploads", 50*1024*1024) // 50MB max

	// Groupe de routes pour la messagerie avec authentification requise
	messaging := router.PathPrefix("/messaging").Subrouter()
	messaging.Use(middleware.JWTMiddleware)

	// Routes pour les conversations
	conversations := messaging.PathPrefix("/conversations").Subrouter()
	
	// Récupérer toutes les conversations de l'utilisateur
	conversations.HandleFunc("", conversationHandler.GetUserConversations).Methods("GET")
	
	// Créer une nouvelle conversation
	conversations.HandleFunc("", conversationHandler.CreateConversation).Methods("POST")
	
	// Récupérer une conversation spécifique
	conversations.HandleFunc("/{id}", conversationHandler.GetConversation).Methods("GET")
	
	// Récupérer les messages d'une conversation
	conversations.HandleFunc("/{id}/messages", conversationHandler.GetConversationMessages).Methods("GET")
	
	// Marquer une conversation comme lue
	conversations.HandleFunc("/{id}/read", conversationHandler.MarkConversationAsRead).Methods("POST")

	// Routes pour les messages
	messages := messaging.PathPrefix("/messages").Subrouter()
	
	// Envoyer un message
	messages.HandleFunc("", messageHandler.SendMessage).Methods("POST")
	
	// Débloqur un message payant
	messages.HandleFunc("/{id}/unlock", messageHandler.UnlockPaidMessage).Methods("POST")
	
	// Récupérer l'aperçu d'un message payant
	messages.HandleFunc("/{id}/preview", messageHandler.GetPaidMessagePreview).Methods("GET")
	
	// Marquer un message comme lu
	messages.HandleFunc("/{id}/read", messageHandler.MarkMessageAsRead).Methods("POST")
	
	// Supprimer un message
	messages.HandleFunc("/{id}", messageHandler.DeleteMessage).Methods("DELETE")
	
	// Récupérer les transactions d'un message
	messages.HandleFunc("/{id}/transactions", messageHandler.GetMessageTransactions).Methods("GET")
	
	// Récupérer les statistiques de messages de l'utilisateur
	messages.HandleFunc("/stats", messageHandler.GetUserMessagesStats).Methods("GET")

	// Routes pour les médias
	media := messaging.PathPrefix("/media").Subrouter()
	
	// Uploader un fichier média
	media.HandleFunc("/upload", mediaHandler.UploadMedia).Methods("POST")
	
	// Récupérer les métadonnées d'un fichier
	media.HandleFunc("/{id}", mediaHandler.GetMedia).Methods("GET")
	
	// Supprimer un fichier
	media.HandleFunc("/{id}", mediaHandler.DeleteMedia).Methods("DELETE")
	
	// Récupérer tous les fichiers de l'utilisateur
	media.HandleFunc("", mediaHandler.GetUserMediaFiles).Methods("GET")
	
	// Récupérer les statistiques de médias
	media.HandleFunc("/stats", mediaHandler.GetMediaStats).Methods("GET")

	// Route pour servir les fichiers (sans authentification pour permettre l'affichage)
	router.PathPrefix("/uploads/").Handler(
		middleware.MediaAccessMiddleware(
			http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads/")))
		)
	).Methods("GET")

	// Route pour le nombre de messages non lus
	messaging.HandleFunc("/unread-count", conversationHandler.GetUnreadCount).Methods("GET")
}

// RegisterProfileRoutes configure les routes pour l'édition de profil
func RegisterProfileRoutes(router *mux.Router) {
	// Groupe de routes pour les profils avec authentification requise
	profiles := router.PathPrefix("/profiles").Subrouter()
	profiles.Use(middleware.JWTMiddleware)

	// Routes pour les utilisateurs
	users := profiles.PathPrefix("/users").Subrouter()
	
	// Mettre à jour le profil de l'utilisateur connecté
	users.HandleFunc("/me", handlers.UpdateUserProfile).Methods("PUT")
	
	// Récupérer le profil d'un utilisateur
	users.HandleFunc("/{id}", handlers.GetUserProfile).Methods("GET")
	
	// Mettre à jour les liens sociaux
	users.HandleFunc("/me/social-links", handlers.UpdateUserSocialLinks).Methods("PUT")
	
	// Récupérer les statistiques d'un utilisateur
	users.HandleFunc("/{id}/stats", handlers.GetUserStats).Methods("GET")

	// Routes pour les créateurs
	creators := profiles.PathPrefix("/creators").Subrouter()
	creators.Use(middleware.CreatorOnlyMiddleware)
	
	// Mettre à jour le profil du créateur
	creators.HandleFunc("/me", handlers.UpdateCreatorProfile).Methods("PUT")
	
	// Récupérer les gains du créateur
	creators.HandleFunc("/me/earnings", handlers.GetCreatorEarnings).Methods("GET")
	
	// Récupérer les gains mensuels du créateur
	creators.HandleFunc("/me/monthly-earnings", handlers.GetCreatorMonthlyEarnings).Methods("GET")
	
	// Récupérer les statistiques détaillées du créateur
	creators.HandleFunc("/me/stats", handlers.GetCreatorStats).Methods("GET")
}

// RegisterCommunicationRoutes configure les routes de communication (à mettre à jour dans routes.go)
func RegisterCommunicationRoutes(router *mux.Router) {
	RegisterMessagingRoutes(router)
	RegisterProfileRoutes(router)
}
