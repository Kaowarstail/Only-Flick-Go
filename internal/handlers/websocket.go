package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/Kaowarstail/Only-Flick-Go/internal/config"
	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
	websocketPkg "github.com/Kaowarstail/Only-Flick-Go/internal/websocket"
)

// WebSocketHandler gère les connexions WebSocket
type WebSocketHandler struct {
	hub *websocketPkg.Hub
}

// NewWebSocketHandler crée un nouveau handler WebSocket
func NewWebSocketHandler(hub *websocketPkg.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
	}
}

// upgrader configure l'upgrade WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

// checkOrigin vérifie l'origine de la connexion WebSocket
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	
	// En développement, autoriser localhost
	if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
		return true
	}
	
	// En production, vérifier les origines autorisées
	allowedOrigins := config.GetWebSocketConfig().AllowedOrigins
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	
	log.Printf("WebSocket connection refused from origin: %s", origin)
	return false
}

// HandleWebSocket gère les connexions WebSocket
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Extraire l'ID utilisateur du contexte (défini par le middleware JWT)
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized: JWT token required",
		})
		return
	}
	
	// Extraire le nom d'utilisateur (optionnel)
	username := c.GetString("username")
	if username == "" {
		username = "Unknown"
	}
	
	// Upgrade de la connexion HTTP vers WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed for user %s: %v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "WebSocket upgrade failed",
		})
		return
	}
	
	// Créer un nouveau client WebSocket
	client := websocketPkg.NewClient(h.hub, conn, userID, username)
	
	// Enregistrer le client dans le hub
	h.hub.register <- client
	
	// Démarrer les goroutines de lecture et écriture
	go client.writePump()
	go client.readPump()
	
	log.Printf("WebSocket connection established for user: %s", userID)
}

// GetMetrics retourne les métriques WebSocket
func (h *WebSocketHandler) GetMetrics(c *gin.Context) {
	// Vérifier les permissions admin
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}
	
	metrics := h.hub.GetMetrics()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// GetOnlineUsers retourne la liste des utilisateurs en ligne
func (h *WebSocketHandler) GetOnlineUsers(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}
	
	onlineUsers := h.hub.GetOnlineUsers()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"online_users": onlineUsers,
			"count":        len(onlineUsers),
		},
	})
}

// GetUserStatus retourne le statut d'un utilisateur
func (h *WebSocketHandler) GetUserStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}
	
	targetUserID := c.Param("user_id")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "User ID required",
		})
		return
	}
	
	isOnline := h.hub.IsUserOnline(targetUserID)
	activeConversation := h.hub.GetUserActiveConversation(targetUserID)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user_id":             targetUserID,
			"is_online":           isOnline,
			"active_conversation": activeConversation,
		},
	})
}

// GetConnectionInfo retourne les informations de connexion
func (h *WebSocketHandler) GetConnectionInfo(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}
	
	connectionCount := h.hub.GetConnectionCount()
	isOnline := h.hub.IsUserOnline(userID)
	activeConversation := h.hub.GetUserActiveConversation(userID)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user_id":             userID,
			"is_online":           isOnline,
			"active_conversation": activeConversation,
			"total_connections":   connectionCount,
		},
	})
}

// HealthCheck vérifie la santé du service WebSocket
func (h *WebSocketHandler) HealthCheck(c *gin.Context) {
	connectionCount := h.hub.GetConnectionCount()
	
	status := "healthy"
	if connectionCount > 900 { // 90% de la limite
		status = "warning"
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"status":  status,
		"data": gin.H{
			"connections": connectionCount,
			"max_connections": config.GetWebSocketConfig().ConnectionLimit,
			"timestamp": utils.GetCurrentTime(),
		},
	})
}
