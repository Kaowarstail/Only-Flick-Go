package dto

import (
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/models"
)

type CreateConversationRequest struct {
	OtherUserID string `json:"other_user_id" binding:"required"`
}

type ConversationResponse struct {
	ID           uint             `json:"id"`
	Participants []models.User    `json:"participants"`
	LastMessage  *MessageResponse `json:"last_message"`
	UnreadCount  int              `json:"unread_count"`
	UpdatedAt    time.Time        `json:"updated_at"`
	IsActive     bool             `json:"is_active"`
}

type ConversationsResponse struct {
	Conversations []ConversationResponse `json:"conversations"`
	Total         int64                  `json:"total"`
	UnreadTotal   int                    `json:"unread_total"`
}
