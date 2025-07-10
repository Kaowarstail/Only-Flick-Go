package dto

import (
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/models"
)

type CreateConversationClassicRequest struct {
	OtherUserID string `json:"other_user_id" binding:"required"`
}

type ConversationClassicResponseFull struct {
	ID           string                   `json:"id"`
	Type         string                   `json:"type"`
	Participants []models.User            `json:"participants"`
	LastMessage  *MessageClassicResponse  `json:"last_message"`
	UnreadCount  int                      `json:"unread_count"`
	UpdatedAt    time.Time                `json:"updated_at"`
	IsActive     bool                     `json:"is_active"`
}

type ConversationsClassicResponse struct {
	Conversations []ConversationClassicResponseFull `json:"conversations"`
	Total         int64                             `json:"total"`
	UnreadTotal   int                               `json:"unread_total"`
}
