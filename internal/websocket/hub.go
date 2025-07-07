package websocket

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/models"
	"github.com/Kaowarstail/Only-Flick-Go/internal/services"
	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
	"github.com/Kaowarstail/Only-Flick-Go/internal/config"
)

// Hub gère toutes les connexions WebSocket
type Hub struct {
	// Connexions
	clients             map[*Client]bool
	userClients         map[string][]*Client
	conversationClients map[string][]*Client
	
	// Channels
	register   chan *Client
	unregister chan *Client
	broadcast  chan models.WebSocketMessage
	
	// Services
	messageService      *services.MessageService
	conversationService *services.ConversationService
	
	// Utils
	rateLimiter *utils.RateLimiter
	metrics     *utils.WebSocketMetrics
	
	// Database
	db *sql.DB
	
	// Configuration
	config config.WebSocketConfig
	
	// Mutex pour la concurrence
	mu sync.RWMutex
}

// NewHub crée un nouveau Hub WebSocket
func NewHub(messageService *services.MessageService, conversationService *services.ConversationService, db *sql.DB) *Hub {
	return &Hub{
		clients:             make(map[*Client]bool),
		userClients:         make(map[string][]*Client),
		conversationClients: make(map[string][]*Client),
		register:            make(chan *Client),
		unregister:          make(chan *Client),
		broadcast:           make(chan models.WebSocketMessage),
		messageService:      messageService,
		conversationService: conversationService,
		rateLimiter:         utils.NewRateLimiter(),
		metrics:             utils.NewWebSocketMetrics(),
		db:                  db,
		config:              config.GetWebSocketConfig(),
	}
}

// Run démarre le hub WebSocket
func (h *Hub) Run() {
	// Démarrer les tâches de nettoyage
	go h.cleanupRoutine()
	
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
			
		case client := <-h.unregister:
			h.unregisterClient(client)
			
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient enregistre un nouveau client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// Vérifier la limite de connexions
	if len(h.clients) >= h.config.ConnectionLimit {
		client.sendError("CONNECTION_LIMIT_EXCEEDED", "Server connection limit reached")
		client.Close()
		return
	}
	
	// Ajouter aux maps
	h.clients[client] = true
	h.userClients[client.userID] = append(h.userClients[client.userID], client)
	
	// Métriques
	h.metrics.IncrementConnections()
	
	// Mettre à jour l'activité utilisateur
	h.updateUserActivity(client.userID)
	
	// Notifier que l'utilisateur est en ligne
	if h.config.EnablePresence {
		h.broadcastUserStatus(client.userID, true)
	}
	
	// Confirmer la connexion
	client.sendConnectionEstablished()
	
	log.Printf("WebSocket client connected: %s (total: %d)", client.userID, len(h.clients))
}

// unregisterClient désenregistre un client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if _, ok := h.clients[client]; ok {
		// Supprimer des maps
		delete(h.clients, client)
		h.removeUserClient(client.userID, client)
		
		// Supprimer de la conversation active
		if client.conversationID != nil {
			h.removeClientFromConversation(*client.conversationID, client)
			
			// Arrêter le typing si nécessaire
			if client.isTyping {
				client.stopTyping(*client.conversationID)
			}
		}
		
		close(client.send)
		h.metrics.DecrementConnections()
		
		// Si plus de clients pour cet utilisateur, marquer offline
		if len(h.userClients[client.userID]) == 0 {
			if h.config.EnablePresence {
				h.broadcastUserStatus(client.userID, false)
			}
		}
		
		log.Printf("WebSocket client disconnected: %s (total: %d)", client.userID, len(h.clients))
	}
}

// broadcastMessage diffuse un message
func (h *Hub) broadcastMessage(message models.WebSocketMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			// Client déconnecté, le supprimer
			h.unregister <- client
		}
	}
}

// BroadcastToConversation diffuse à tous les clients d'une conversation
func (h *Hub) BroadcastToConversation(conversationID string, message models.WebSocketMessage) {
	h.mu.RLock()
	clients := make([]*Client, len(h.conversationClients[conversationID]))
	copy(clients, h.conversationClients[conversationID])
	h.mu.RUnlock()
	
	for _, client := range clients {
		select {
		case client.send <- message:
		default:
			// Client déconnecté, le supprimer
			h.unregister <- client
		}
	}
}

// BroadcastToUser diffuse à tous les clients d'un utilisateur
func (h *Hub) BroadcastToUser(userID string, message models.WebSocketMessage) {
	h.mu.RLock()
	clients := make([]*Client, len(h.userClients[userID]))
	copy(clients, h.userClients[userID])
	h.mu.RUnlock()
	
	for _, client := range clients {
		select {
		case client.send <- message:
		default:
			h.unregister <- client
		}
	}
}

// BroadcastToConversationParticipants diffuse aux participants d'une conversation
func (h *Hub) BroadcastToConversationParticipants(conversationID string, message models.WebSocketMessage, excludeUserID string) {
	// Récupérer les participants de la conversation
	conversation, err := h.conversationService.GetConversationByID(conversationID)
	if err != nil {
		log.Printf("Error getting conversation %s: %v", conversationID, err)
		return
	}
	
	// Broadcaster à chaque participant (sauf l'expéditeur)
	if conversation.Participant1ID != excludeUserID {
		h.BroadcastToUser(conversation.Participant1ID, message)
	}
	if conversation.Participant2ID != excludeUserID {
		h.BroadcastToUser(conversation.Participant2ID, message)
	}
}

// BroadcastMessageSent diffuse un nouveau message envoyé
func (h *Hub) BroadcastMessageSent(message *models.Message, conversation *models.Conversation, sender *models.User) {
	messageEvent := models.WebSocketMessage{
		Type: models.EventMessageSent,
		Data: models.MessageSentEvent{
			Message:      *message,
			Conversation: *conversation,
			Sender:       *sender,
		},
		Timestamp:      time.Now(),
		UserID:         sender.ID,
		ConversationID: &message.ConversationID,
	}
	
	h.BroadcastToConversationParticipants(message.ConversationID, messageEvent, sender.ID)
}

// BroadcastPaidMessageUnlocked diffuse un message payant débloqué
func (h *Hub) BroadcastPaidMessageUnlocked(message *models.Message, unlockedBy string, transaction *models.PaidMessageTransaction) {
	unlockEvent := models.WebSocketMessage{
		Type: models.EventPaidMessageUnlocked,
		Data: models.PaidMessageUnlockedEvent{
			MessageID:   message.ID,
			UnlockedBy:  unlockedBy,
			Message:     *message,
			Transaction: transaction,
		},
		Timestamp:      time.Now(),
		ConversationID: &message.ConversationID,
	}
	
	h.BroadcastToConversationParticipants(message.ConversationID, unlockEvent, "")
}

// BroadcastConversationUpdated diffuse une mise à jour de conversation
func (h *Hub) BroadcastConversationUpdated(conversation *models.Conversation, lastMessage *models.Message, unreadCount int) {
	updateEvent := models.WebSocketMessage{
		Type: models.EventConversationUpdated,
		Data: models.ConversationUpdatedEvent{
			Conversation: *conversation,
			LastMessage:  lastMessage,
			UnreadCount:  unreadCount,
		},
		Timestamp:      time.Now(),
		ConversationID: &conversation.ID,
	}
	
	h.BroadcastToUser(conversation.Participant1ID, updateEvent)
	h.BroadcastToUser(conversation.Participant2ID, updateEvent)
}

// addClientToConversation ajoute un client à une conversation
func (h *Hub) addClientToConversation(conversationID string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// Vérifier si le client est déjà dans la liste
	for _, existingClient := range h.conversationClients[conversationID] {
		if existingClient == client {
			return
		}
	}
	
	h.conversationClients[conversationID] = append(h.conversationClients[conversationID], client)
}

// removeClientFromConversation retire un client d'une conversation
func (h *Hub) removeClientFromConversation(conversationID string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	clients := h.conversationClients[conversationID]
	for i, existingClient := range clients {
		if existingClient == client {
			// Supprimer le client de la liste
			h.conversationClients[conversationID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	
	// Supprimer la conversation si plus de clients
	if len(h.conversationClients[conversationID]) == 0 {
		delete(h.conversationClients, conversationID)
	}
}

// removeUserClient retire un client de la liste des clients utilisateur
func (h *Hub) removeUserClient(userID string, client *Client) {
	clients := h.userClients[userID]
	for i, existingClient := range clients {
		if existingClient == client {
			h.userClients[userID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	
	// Supprimer l'utilisateur si plus de clients
	if len(h.userClients[userID]) == 0 {
		delete(h.userClients, userID)
	}
}

// updateUserActivity met à jour l'activité utilisateur
func (h *Hub) updateUserActivity(userID string) {
	query := `
		UPDATE user_stats 
		SET last_active_at = NOW(), updated_at = NOW()
		WHERE user_id = $1
	`
	_, err := h.db.Exec(query, userID)
	if err != nil {
		log.Printf("Failed to update user activity: %v", err)
	}
}

// broadcastUserStatus diffuse le statut d'un utilisateur
func (h *Hub) broadcastUserStatus(userID string, isOnline bool) {
	// Récupérer infos utilisateur
	user, err := h.getUserInfo(userID)
	if err != nil {
		log.Printf("Error getting user info for %s: %v", userID, err)
		return
	}
	
	eventType := models.EventUserOnline
	if !isOnline {
		eventType = models.EventUserOffline
	}
	
	statusEvent := models.WebSocketMessage{
		Type: eventType,
		Data: models.UserStatusEvent{
			UserID:       userID,
			Username:     user.Username,
			IsOnline:     isOnline,
			LastActiveAt: time.Now(),
		},
		Timestamp: time.Now(),
		UserID:    userID,
	}
	
	h.metrics.IncrementUserStatusEvents()
	
	// Broadcaster à tous les utilisateurs qui ont des conversations avec cet utilisateur
	h.broadcastToUserContacts(userID, statusEvent)
}

// getUserInfo récupère les informations d'un utilisateur
func (h *Hub) getUserInfo(userID string) (*models.User, error) {
	query := `
		SELECT id, username, email, created_at, updated_at
		FROM users 
		WHERE id = $1
	`
	
	var user models.User
	err := h.db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	return &user, err
}

// broadcastToUserContacts diffuse à tous les contacts d'un utilisateur
func (h *Hub) broadcastToUserContacts(userID string, message models.WebSocketMessage) {
	// Récupérer les conversations de l'utilisateur
	conversations, err := h.conversationService.GetUserConversations(userID)
	if err != nil {
		log.Printf("Error getting user conversations: %v", err)
		return
	}
	
	// Broadcaster à chaque contact
	contactsSet := make(map[string]bool)
	for _, conversation := range conversations {
		var contactID string
		if conversation.Participant1ID == userID {
			contactID = conversation.Participant2ID
		} else {
			contactID = conversation.Participant1ID
		}
		
		if !contactsSet[contactID] {
			contactsSet[contactID] = true
			h.BroadcastToUser(contactID, message)
		}
	}
}

// markMessagesAsRead marque les messages comme lus
func (h *Hub) markMessagesAsRead(conversationID, userID string) {
	query := `
		UPDATE messages 
		SET is_read = true, read_at = NOW()
		WHERE conversation_id = $1 AND sender_id != $2 AND is_read = false
	`
	
	_, err := h.db.Exec(query, conversationID, userID)
	if err != nil {
		log.Printf("Error marking messages as read: %v", err)
	}
}

// cleanupRoutine nettoie les connexions et données obsolètes
func (h *Hub) cleanupRoutine() {
	ticker := time.NewTicker(h.config.CleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			h.cleanup()
		}
	}
}

// cleanup nettoie les données obsolètes
func (h *Hub) cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// Nettoyer les clients inactifs
	var clientsToRemove []*Client
	for client := range h.clients {
		if !client.IsActive() {
			clientsToRemove = append(clientsToRemove, client)
		}
	}
	
	// Supprimer les clients inactifs
	for _, client := range clientsToRemove {
		h.unregister <- client
	}
	
	// Nettoyer le rate limiter
	h.rateLimiter.CleanupOldClients()
	
	log.Printf("WebSocket cleanup completed: removed %d inactive clients", len(clientsToRemove))
}

// GetMetrics retourne les métriques du hub
func (h *Hub) GetMetrics() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	metrics := h.metrics.GetMetrics()
	
	// Ajouter les métriques du hub
	metrics["clients_by_user"] = len(h.userClients)
	metrics["active_conversations"] = len(h.conversationClients)
	metrics["rate_limiter_stats"] = h.rateLimiter.GetStats()
	
	return metrics
}

// GetOnlineUsers retourne la liste des utilisateurs en ligne
func (h *Hub) GetOnlineUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	var onlineUsers []string
	for userID := range h.userClients {
		if len(h.userClients[userID]) > 0 {
			onlineUsers = append(onlineUsers, userID)
		}
	}
	
	return onlineUsers
}

// IsUserOnline vérifie si un utilisateur est en ligne
func (h *Hub) IsUserOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	clients := h.userClients[userID]
	return len(clients) > 0
}

// GetUserActiveConversation retourne la conversation active d'un utilisateur
func (h *Hub) GetUserActiveConversation(userID string) *string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	clients := h.userClients[userID]
	if len(clients) > 0 && clients[0].conversationID != nil {
		return clients[0].conversationID
	}
	
	return nil
}

// GetConnectionCount retourne le nombre de connexions actives
func (h *Hub) GetConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	return len(h.clients)
}
