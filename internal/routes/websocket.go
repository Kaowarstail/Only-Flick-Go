package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/Kaowarstail/Only-Flick-Go/internal/handlers"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	websocketPkg "github.com/Kaowarstail/Only-Flick-Go/internal/websocket"
)

// RegisterWebSocketRoutes enregistre les routes WebSocket
func RegisterWebSocketRoutes(router *gin.Engine, hub *websocketPkg.Hub) {
	// Créer le handler WebSocket
	wsHandler := handlers.NewWebSocketHandler(hub)
	
	// Groupe API v1 pour WebSocket
	v1 := router.Group("/api/v1")
	
	// Route de connexion WebSocket (nécessite JWT)
	v1.GET("/ws", middleware.JWTMiddleware(), wsHandler.HandleWebSocket)
	
	// Routes d'information WebSocket (nécessite JWT)
	wsRoutes := v1.Group("/ws")
	wsRoutes.Use(middleware.JWTMiddleware())
	{
		// Informations de connexion
		wsRoutes.GET("/info", wsHandler.GetConnectionInfo)
		
		// Liste des utilisateurs en ligne
		wsRoutes.GET("/online", wsHandler.GetOnlineUsers)
		
		// Statut d'un utilisateur spécifique
		wsRoutes.GET("/user/:user_id/status", wsHandler.GetUserStatus)
		
		// Health check
		wsRoutes.GET("/health", wsHandler.HealthCheck)
	}
	
	// Routes admin (nécessite JWT + admin)
	adminRoutes := v1.Group("/ws/admin")
	adminRoutes.Use(middleware.JWTMiddleware())
	adminRoutes.Use(middleware.AdminOnlyMiddleware())
	{
		// Métriques WebSocket
		adminRoutes.GET("/metrics", wsHandler.GetMetrics)
	}
}
