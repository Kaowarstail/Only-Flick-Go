package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	websocketPkg "github.com/Kaowarstail/Only-Flick-Go/internal/websocket"
)

// RegisterGinRoutes configure toutes les routes de l'API avec Gin
func RegisterGinRoutes(router *gin.Engine, hub *websocketPkg.Hub) {
	// Middleware CORS global
	router.Use(middleware.CORSMiddleware())

	// Route pour la racine "/"
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Bienvenue sur l'API OnlyFlick",
			"version": "1.0.0",
			"websocket": "enabled",
		})
	})

	// API versioning
	apiV1 := router.Group("/api/v1")

	// Middleware global pour la journalisation
	apiV1.Use(middleware.LoggerMiddleware())

	// Health check
	apiV1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
			"websocket_connections": hub.GetConnectionCount(),
		})
	})

	// Enregistrement des différentes catégories de routes
	RegisterGinAuthRoutes(apiV1)
	RegisterGinUserRoutes(apiV1)
	RegisterGinCreatorRoutes(apiV1)
	RegisterGinContentRoutes(apiV1)
	RegisterGinSubscriptionRoutes(apiV1)
	RegisterGinCommunicationRoutes(apiV1)
	RegisterGinMessagingRoutes(apiV1, hub)
	RegisterGinProfileRoutes(apiV1)
	RegisterGinModerationRoutes(apiV1)
	RegisterGinPaymentRoutes(apiV1)
}

// Placeholders pour les routes Gin - à implémenter selon les besoins
func RegisterGinAuthRoutes(router *gin.RouterGroup) {
	// TODO: Implémenter les routes d'authentification
}

func RegisterGinUserRoutes(router *gin.RouterGroup) {
	// TODO: Implémenter les routes utilisateur
}

func RegisterGinCreatorRoutes(router *gin.RouterGroup) {
	// TODO: Implémenter les routes créateur
}

func RegisterGinContentRoutes(router *gin.RouterGroup) {
	// TODO: Implémenter les routes de contenu
}

func RegisterGinSubscriptionRoutes(router *gin.RouterGroup) {
	// TODO: Implémenter les routes d'abonnement
}

func RegisterGinCommunicationRoutes(router *gin.RouterGroup) {
	// TODO: Implémenter les routes de communication
}

func RegisterGinMessagingRoutes(router *gin.RouterGroup, hub *websocketPkg.Hub) {
	// TODO: Implémenter les routes de messagerie avec WebSocket
}

func RegisterGinProfileRoutes(router *gin.RouterGroup) {
	// TODO: Implémenter les routes de profil
}

func RegisterGinModerationRoutes(router *gin.RouterGroup) {
	// TODO: Implémenter les routes de modération
}

func RegisterGinPaymentRoutes(router *gin.RouterGroup) {
	// TODO: Implémenter les routes de paiement
}
