package routes

import (
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/handlers"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/gorilla/mux"
)

// RegisterMessagingRoutes enregistre toutes les routes de messagerie classique
func RegisterMessagingRoutes(router *mux.Router) {
	// Créer les handlers
	conversationHandler := handlers.NewConversationHandler(database.GetDB())
	messageHandler := handlers.NewMessageHandler(database.GetDB())

	// Routes pour les conversations classiques (toutes protégées)
	router.Handle("/conversations", middleware.JWTAuth(http.HandlerFunc(conversationHandler.GetUserConversations))).Methods("GET", "OPTIONS")
	router.Handle("/conversations", middleware.JWTAuth(http.HandlerFunc(conversationHandler.CreateConversation))).Methods("POST", "OPTIONS")
	router.Handle("/conversations/{id}/messages", middleware.JWTAuth(http.HandlerFunc(conversationHandler.GetConversationMessages))).Methods("GET", "OPTIONS")
	router.Handle("/conversations/{id}/read", middleware.JWTAuth(http.HandlerFunc(conversationHandler.MarkConversationAsRead))).Methods("PUT", "OPTIONS")

	// Routes pour les messages classiques (avec rate limiting)
	router.Handle("/messages/{id}", middleware.JWTAuth(http.HandlerFunc(messageHandler.GetMessage))).Methods("GET", "OPTIONS")

	// Envoi de message classique avec rate limiting
	router.Handle("/messages",
		middleware.JWTAuth(
			middleware.MessageRateLimit()(
				http.HandlerFunc(messageHandler.SendMessageClassic),
			),
		),
	).Methods("POST", "OPTIONS")

	// Marquer les messages comme lus
	router.Handle("/messages/read", middleware.JWTAuth(http.HandlerFunc(messageHandler.MarkAsRead))).Methods("PUT", "OPTIONS")
}
