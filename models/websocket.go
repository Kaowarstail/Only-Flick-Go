package models

import (
	"time"
)

// EventType définit les types d'événements WebSocket
type EventType string

const (
	// Messages
	EventMessageSent         EventType = "message_sent"
	EventMessageDelivered    EventType = "message_delivered"
	EventMessageRead         EventType = "message_read"
	EventPaidMessageSent     EventType = "paid_message_sent"
	EventPaidMessageUnlocked EventType = "paid_message_unlocked"

	// Typing indicators
	EventUserTyping        EventType = "user_typing"
	EventUserStoppedTyping EventType = "user_stopped_typing"

	// User status
	EventUserOnline     EventType = "user_online"
	EventUserOffline    EventType = "user_offline"
	EventUserActiveIn   EventType = "user_active_in_conversation"

	// Conversations
	EventConversationUpdated EventType = "conversation_updated"
	EventNewConversation     EventType = "new_conversation"

	// System
	EventError                 EventType = "error"
	EventConnectionEstablished EventType = "connection_established"
	EventHeartbeat             EventType = "heartbeat"
)

// WebSocketMessage représente un message WebSocket
type WebSocketMessage struct {
	Type           EventType   `json:"type"`
	Data           interface{} `json:"data"`
	Timestamp      time.Time   `json:"timestamp"`
	UserID         string      `json:"user_id,omitempty"`
	ConversationID *string     `json:"conversation_id,omitempty"`
}

// MessageSentEvent événement pour message envoyé
type MessageSentEvent struct {
	Message      Message      `json:"message"`
	Conversation Conversation `json:"conversation"`
	Sender       User         `json:"sender"`
}

// TypingEvent événement pour indicateur de frappe
type TypingEvent struct {
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	ConversationID string `json:"conversation_id"`
	IsTyping       bool   `json:"is_typing"`
}

// UserStatusEvent événement pour statut utilisateur
type UserStatusEvent struct {
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	IsOnline     bool      `json:"is_online"`
	LastActiveAt time.Time `json:"last_active_at"`
}

// ConversationUpdatedEvent événement pour mise à jour conversation
type ConversationUpdatedEvent struct {
	Conversation Conversation `json:"conversation"`
	LastMessage  *Message     `json:"last_message"`
	UnreadCount  int          `json:"unread_count"`
}

// PaidMessageUnlockedEvent événement pour message payant débloqué
type PaidMessageUnlockedEvent struct {
	MessageID    string  `json:"message_id"`
	UnlockedBy   string  `json:"unlocked_by"`
	Message      Message `json:"message"`
	Transaction  *PaidMessageTransaction `json:"transaction"`
}

// ErrorEvent événement d'erreur
type ErrorEvent struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ConnectionEstablishedEvent événement de connexion établie
type ConnectionEstablishedEvent struct {
	UserID        string    `json:"user_id"`
	ServerTime    time.Time `json:"server_time"`
	ConnectionID  string    `json:"connection_id"`
	Capabilities  []string  `json:"capabilities"`
}

// HeartbeatEvent événement de battement de coeur
type HeartbeatEvent struct {
	ServerTime time.Time `json:"server_time"`
	ClientTime time.Time `json:"client_time,omitempty"`
}

// ActiveInConversationEvent événement pour utilisateur actif dans conversation
type ActiveInConversationEvent struct {
	UserID         string `json:"user_id"`
	ConversationID string `json:"conversation_id"`
	IsActive       bool   `json:"is_active"`
}
