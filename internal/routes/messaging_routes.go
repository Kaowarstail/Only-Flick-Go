package routes

import (
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/internal/handlers"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// SetupMessagingRoutes configure les routes de messagerie
func SetupMessagingRoutes(router *mux.Router, db *gorm.DB) {
	// Créer le handler de messagerie
	messagingHandler := handlers.NewMessagingHandler(db)

	// Sous-routeur pour les routes de messagerie avec authentification JWT
	messagingRouter := router.PathPrefix("/api").Subrouter()
	messagingRouter.Use(middleware.JWTAuth) // Toutes les routes nécessitent une authentification

	// ========== Conversation Routes ==========

	// GET /api/conversations - Récupérer les conversations de l'utilisateur
	messagingRouter.HandleFunc("/conversations", messagingHandler.GetUserConversations).Methods("GET")

	// POST /api/conversations - Créer ou récupérer une conversation directe
	messagingRouter.HandleFunc("/conversations", messagingHandler.CreateOrGetConversation).Methods("POST")

	// GET /api/conversations/{conversationId}/messages - Récupérer les messages d'une conversation
	messagingRouter.HandleFunc("/conversations/{conversationId}/messages", messagingHandler.GetConversationMessages).Methods("GET")

	// PUT /api/conversations/{conversationId}/read - Marquer une conversation comme lue
	messagingRouter.HandleFunc("/conversations/{conversationId}/read", messagingHandler.MarkConversationAsRead).Methods("PUT")

	// PUT /api/conversations/read-all - Marquer toutes les conversations comme lues
	messagingRouter.HandleFunc("/conversations/read-all", messagingHandler.MarkAllConversationsAsRead).Methods("PUT")

	// ========== Message Routes ==========

	// POST /api/messages - Envoyer un message
	messagingRouter.HandleFunc("/messages", messagingHandler.SendMessage).Methods("POST")

	// ========== Dashboard & Stats Routes ==========

	// GET /api/messaging/dashboard - Récupérer le tableau de bord de messagerie
	messagingRouter.HandleFunc("/messaging/dashboard", messagingHandler.GetMessagingDashboard).Methods("GET")

	// GET /api/messaging/stats - Récupérer les statistiques de messagerie
	messagingRouter.HandleFunc("/messaging/stats", messagingHandler.GetMessagingStats).Methods("GET")

	// ========== Search Routes ==========

	// GET /api/messaging/search - Rechercher dans conversations et messages
	messagingRouter.HandleFunc("/messaging/search", messagingHandler.SearchMessaging).Methods("GET")

	// ========== Advanced Operations Routes ==========

	// POST /api/messaging/start - Créer une conversation et envoyer le premier message
	messagingRouter.HandleFunc("/messaging/start", messagingHandler.StartConversationAndSendMessage).Methods("POST")
}

// SetupMessagingRoutesWithCORS configure les routes de messagerie avec CORS
func SetupMessagingRoutesWithCORS(router *mux.Router, db *gorm.DB) {
	// Configuration des routes principales
	SetupMessagingRoutes(router, db)

	// Sous-routeur pour les routes CORS (OPTIONS)
	corsRouter := router.PathPrefix("/api").Subrouter()

	// Ajouter les routes OPTIONS pour CORS
	corsRoutes := []string{
		"/conversations",
		"/conversations/{conversationId}/messages",
		"/conversations/{conversationId}/read",
		"/conversations/read-all",
		"/messages",
		"/messaging/dashboard",
		"/messaging/stats",
		"/messaging/search",
		"/messaging/start",
	}

	for _, route := range corsRoutes {
		corsRouter.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
		}).Methods("OPTIONS")
	}
}
