package dto

import (
	internalmodels "github.com/Kaowarstail/Only-Flick-Go/internal/models"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

// SendMessageClassicRequest pour envoyer un message classique
type SendMessageClassicRequest struct {
	ConversationID string  `json:"conversation_id" binding:"required"`
	Content        *string `json:"content"`
	MediaURL       *string `json:"media_url"`
	MediaType      *string `json:"media_type"`
	MessageType    string  `json:"message_type" binding:"required,oneof=text image video audio"`
}

// MessagesClassicResponse pour la réponse des messages classiques
type MessagesClassicResponse struct {
	Messages []internalmodels.MessageClassic `json:"messages"`
	Total    int64                           `json:"total"`
	Page     int                             `json:"page"`
	Limit    int                             `json:"limit"`
	HasMore  bool                            `json:"has_more"`
}

// MessageClassicResponse pour la réponse d'un message classique
type MessageClassicResponse struct {
	ID             string      `json:"id"`
	ConversationID string      `json:"conversation_id"`
	Sender         models.User `json:"sender"`
	Content        *string     `json:"content"`
	MediaURL       *string     `json:"media_url"`
	MediaType      *string     `json:"media_type"`
	MessageType    string      `json:"message_type"`
	Status         string      `json:"status"`
	CreatedAt      string      `json:"created_at"`
}

// ConversationClassicResponse pour la réponse d'une conversation
type ConversationClassicResponse struct {
	ID           string      `json:"id"`
	Type         string      `json:"type"`
	Participants []models.User `json:"participants"`
	LastMessage  *MessageClassicResponse `json:"last_message"`
	IsActive     bool        `json:"is_active"`
	CreatedAt    string      `json:"created_at"`
	UpdatedAt    string      `json:"updated_at"`
}
