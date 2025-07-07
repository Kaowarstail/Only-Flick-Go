package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"

	"github.com/Kaowarstail/Only-Flick-Go/models"
	"github.com/Kaowarstail/Only-Flick-Go/internal/config"
)

// Client représente une connexion WebSocket client
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan models.WebSocketMessage
	userID string
	username string
	
	// État de la conversation
	conversationID *string
	
	// État du typing
	isTyping    bool
	typingTimer *time.Timer
	
	// Activité
	lastActivity time.Time
	
	// Configuration
	config config.WebSocketConfig
}

// NewClient crée un nouveau client WebSocket
func NewClient(hub *Hub, conn *websocket.Conn, userID string, username string) *Client {
	return &Client{
		hub:          hub,
		conn:         conn,
		send:         make(chan models.WebSocketMessage, 256),
		userID:       userID,
		username:     username,
		lastActivity: time.Now(),
		config:       config.GetWebSocketConfig(),
	}
}

// readPump gère la lecture des messages du client
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// Configuration des timeouts
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(c.config.PongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.config.PongWait))
		c.lastActivity = time.Now()
		return nil
	})

	for {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
				c.hub.metrics.IncrementErrors()
			}
			break
		}

		c.lastActivity = time.Now()
		c.hub.metrics.IncrementMessagesReceived()
		c.hub.metrics.UpdateMessageSize(len(message))
		
		c.handleMessage(messageType, message)
	}
}

// writePump gère l'écriture des messages vers le client
func (c *Client) writePump() {
	ticker := time.NewTicker(c.config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				c.hub.metrics.IncrementErrors()
				return
			}
			
			c.hub.metrics.IncrementMessagesSent()

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage traite un message reçu du client
func (c *Client) handleMessage(messageType int, data []byte) {
	if messageType != websocket.TextMessage {
		return
	}

	// Vérifier rate limiting
	if !c.hub.rateLimiter.AllowMessage(c.userID) {
		c.hub.metrics.IncrementRateLimited()
		c.sendError("RATE_LIMIT_EXCEEDED", "Too many messages, please slow down")
		return
	}

	var wsMessage models.WebSocketMessage
	if err := json.Unmarshal(data, &wsMessage); err != nil {
		c.sendError("INVALID_MESSAGE_FORMAT", "Invalid message format")
		return
	}

	// Ajouter les métadonnées du message
	wsMessage.UserID = c.userID
	wsMessage.Timestamp = time.Now()

	switch wsMessage.Type {
	case models.EventUserTyping:
		c.handleTypingEvent(wsMessage)
	case models.EventUserStoppedTyping:
		c.handleStoppedTypingEvent(wsMessage)
	case models.EventUserActiveIn:
		c.handleActiveInConversation(wsMessage)
	case models.EventHeartbeat:
		c.handleHeartbeat(wsMessage)
	default:
		c.sendError("UNKNOWN_EVENT_TYPE", "Unknown event type: "+string(wsMessage.Type))
	}
}

// handleTypingEvent traite un événement de frappe
func (c *Client) handleTypingEvent(wsMessage models.WebSocketMessage) {
	var typingData models.TypingEvent
	if err := mapstructure.Decode(wsMessage.Data, &typingData); err != nil {
		c.sendError("INVALID_TYPING_EVENT", "Invalid typing event data")
		return
	}

	// Vérifier que l'utilisateur est participant de la conversation
	isParticipant, err := c.hub.conversationService.IsParticipant(typingData.ConversationID, c.userID)
	if err != nil || !isParticipant {
		c.sendError("ACCESS_DENIED", "Access denied to conversation")
		return
	}

	// Mettre à jour l'état de typing
	c.isTyping = true
	c.conversationID = &typingData.ConversationID

	// Ajouter le client à la conversation
	c.hub.addClientToConversation(typingData.ConversationID, c)

	// Broadcaster typing à l'autre participant
	typingEvent := models.WebSocketMessage{
		Type: models.EventUserTyping,
		Data: models.TypingEvent{
			UserID:         c.userID,
			Username:       c.username,
			ConversationID: typingData.ConversationID,
			IsTyping:       true,
		},
		Timestamp:      time.Now(),
		UserID:         c.userID,
		ConversationID: &typingData.ConversationID,
	}

	c.hub.BroadcastToConversationParticipants(typingData.ConversationID, typingEvent, c.userID)
	c.hub.metrics.IncrementTypingEvents()

	// Auto-stop typing après timeout
	c.scheduleStopTyping(typingData.ConversationID)
}

// handleStoppedTypingEvent traite l'arrêt de frappe
func (c *Client) handleStoppedTypingEvent(wsMessage models.WebSocketMessage) {
	var typingData models.TypingEvent
	if err := mapstructure.Decode(wsMessage.Data, &typingData); err != nil {
		c.sendError("INVALID_TYPING_EVENT", "Invalid typing event data")
		return
	}

	// Vérifier que l'utilisateur est participant de la conversation
	isParticipant, err := c.hub.conversationService.IsParticipant(typingData.ConversationID, c.userID)
	if err != nil || !isParticipant {
		c.sendError("ACCESS_DENIED", "Access denied to conversation")
		return
	}

	c.stopTyping(typingData.ConversationID)
}

// handleActiveInConversation traite l'activité dans une conversation
func (c *Client) handleActiveInConversation(wsMessage models.WebSocketMessage) {
	var activeData models.ActiveInConversationEvent
	if err := mapstructure.Decode(wsMessage.Data, &activeData); err != nil {
		c.sendError("INVALID_ACTIVE_EVENT", "Invalid active event data")
		return
	}

	// Vérifier que l'utilisateur est participant de la conversation
	isParticipant, err := c.hub.conversationService.IsParticipant(activeData.ConversationID, c.userID)
	if err != nil || !isParticipant {
		c.sendError("ACCESS_DENIED", "Access denied to conversation")
		return
	}

	if activeData.IsActive {
		// Ajouter à la conversation
		c.conversationID = &activeData.ConversationID
		c.hub.addClientToConversation(activeData.ConversationID, c)
		
		// Marquer les messages comme lus
		c.hub.markMessagesAsRead(activeData.ConversationID, c.userID)
	} else {
		// Retirer de la conversation
		if c.conversationID != nil {
			c.hub.removeClientFromConversation(*c.conversationID, c)
		}
		c.conversationID = nil
	}
}

// handleHeartbeat traite un heartbeat
func (c *Client) handleHeartbeat(wsMessage models.WebSocketMessage) {
	var heartbeatData models.HeartbeatEvent
	if err := mapstructure.Decode(wsMessage.Data, &heartbeatData); err != nil {
		heartbeatData = models.HeartbeatEvent{}
	}

	// Répondre avec le heartbeat
	response := models.WebSocketMessage{
		Type: models.EventHeartbeat,
		Data: models.HeartbeatEvent{
			ServerTime: time.Now(),
			ClientTime: heartbeatData.ClientTime,
		},
		Timestamp: time.Now(),
		UserID:    c.userID,
	}

	c.send <- response
}

// scheduleStopTyping programme l'arrêt automatique du typing
func (c *Client) scheduleStopTyping(conversationID string) {
	if c.typingTimer != nil {
		c.typingTimer.Stop()
	}

	c.typingTimer = time.AfterFunc(c.config.TypingTimeout, func() {
		c.stopTyping(conversationID)
	})
}

// stopTyping arrête le typing
func (c *Client) stopTyping(conversationID string) {
	if !c.isTyping {
		return
	}

	c.isTyping = false

	// Broadcaster stop typing
	stopTypingEvent := models.WebSocketMessage{
		Type: models.EventUserStoppedTyping,
		Data: models.TypingEvent{
			UserID:         c.userID,
			Username:       c.username,
			ConversationID: conversationID,
			IsTyping:       false,
		},
		Timestamp:      time.Now(),
		UserID:         c.userID,
		ConversationID: &conversationID,
	}

	c.hub.BroadcastToConversationParticipants(conversationID, stopTypingEvent, c.userID)
}

// sendError envoie un message d'erreur au client
func (c *Client) sendError(code, message string) {
	errorEvent := models.WebSocketMessage{
		Type: models.EventError,
		Data: models.ErrorEvent{
			Code:    code,
			Message: message,
		},
		Timestamp: time.Now(),
		UserID:    c.userID,
	}

	select {
	case c.send <- errorEvent:
	default:
		// Canal plein, client probablement déconnecté
	}
}

// sendConnectionEstablished envoie la confirmation de connexion
func (c *Client) sendConnectionEstablished() {
	capabilities := []string{"messaging", "typing", "presence"}
	
	if c.config.EnableTypingIndicators {
		capabilities = append(capabilities, "typing_indicators")
	}
	if c.config.EnablePresence {
		capabilities = append(capabilities, "user_presence")
	}

	connectionEvent := models.WebSocketMessage{
		Type: models.EventConnectionEstablished,
		Data: models.ConnectionEstablishedEvent{
			UserID:       c.userID,
			ServerTime:   time.Now(),
			ConnectionID: c.generateConnectionID(),
			Capabilities: capabilities,
		},
		Timestamp: time.Now(),
		UserID:    c.userID,
	}

	c.send <- connectionEvent
}

// generateConnectionID génère un ID de connexion unique
func (c *Client) generateConnectionID() string {
	return c.userID + "_" + time.Now().Format("20060102150405")
}

// IsActive retourne si le client est actif
func (c *Client) IsActive() bool {
	return time.Since(c.lastActivity) < c.config.InactivityTimeout
}

// GetUserID retourne l'ID utilisateur
func (c *Client) GetUserID() string {
	return c.userID
}

// GetConversationID retourne l'ID de conversation active
func (c *Client) GetConversationID() *string {
	return c.conversationID
}

// Close ferme la connexion client
func (c *Client) Close() {
	if c.typingTimer != nil {
		c.typingTimer.Stop()
	}
	c.conn.Close()
}
