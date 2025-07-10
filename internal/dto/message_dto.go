package dto

import (
	"errors"
	"strings"

	"github.com/Kaowarstail/Only-Flick-Go/internal/models"
)

// SendMessageRequest représente une demande pour envoyer un message
type SendMessageRequest struct {
	ConversationID uint    `json:"conversation_id" binding:"required"`
	Content        *string `json:"content"`
	MediaURL       *string `json:"media_url"`
	MediaType      *string `json:"media_type"` // 'image', 'video', 'audio'
	MessageType    string  `json:"message_type" binding:"required,oneof=text image video audio"`
}

// Validate effectue une validation custom de la requête
func (req *SendMessageRequest) Validate() error {
	// Au moins contenu OU média requis
	hasContent := req.Content != nil && strings.TrimSpace(*req.Content) != ""
	hasMedia := req.MediaURL != nil && strings.TrimSpace(*req.MediaURL) != ""

	if !hasContent && !hasMedia {
		return errors.New("message must have content or media")
	}

	// Validation longueur contenu
	if req.Content != nil && len(strings.TrimSpace(*req.Content)) > 5000 {
		return errors.New("message content too long (max 5000 characters)")
	}

	// Media type requis si URL fournie
	if hasMedia && (req.MediaType == nil || strings.TrimSpace(*req.MediaType) == "") {
		return errors.New("media_type required when media_url provided")
	}

	return nil
}

// MessagesResponse représente la réponse pour une liste de messages avec pagination
type MessagesResponse struct {
	Messages    []models.Message `json:"messages"`
	Total       int64            `json:"total"`
	Page        int              `json:"page"`
	Limit       int              `json:"limit"`
	HasMore     bool             `json:"has_more"`
	UnreadCount int              `json:"unread_count"`
}

// MessageStatsResponse représente les statistiques de messages
type MessageStatsResponse struct {
	TotalMessages int `json:"total_messages"`
	UnreadMessages int `json:"unread_messages"`
	MediaMessages int `json:"media_messages"`
	TextMessages  int `json:"text_messages"`
}

// MarkMessageReadRequest représente une demande pour marquer un message comme lu
type MarkMessageReadRequest struct {
	MessageID uint `json:"message_id" binding:"required"`
}

// MarkMessagesReadRequest représente une demande pour marquer plusieurs messages comme lus
type MarkMessagesReadRequest struct {
	ConversationID uint   `json:"conversation_id" binding:"required"`
	MessageIDs     []uint `json:"message_ids"`
}

// SearchMessagesRequest représente une demande de recherche dans les messages
type SearchMessagesRequest struct {
	Query          string `json:"query" binding:"required,min=1"`
	ConversationID *uint  `json:"conversation_id"`
	Page           int    `json:"page" binding:"omitempty,min=1"`
	Limit          int    `json:"limit" binding:"omitempty,min=1,max=50"`
}
