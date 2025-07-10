package dto

import (
	internalmodels "github.com/Kaowarstail/Only-Flick-Go/internal/models"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

type SendMessageRequest struct {
	ConversationID uint    `json:"conversation_id" binding:"required"`
	Content        *string `json:"content"`
	MediaURL       *string `json:"media_url"`
	MediaType      *string `json:"media_type"`
	MessageType    string  `json:"message_type" binding:"required,oneof=text image video"`
}

type SendPaidMessageRequest struct {
	ConversationID uint    `json:"conversation_id" binding:"required"`
	Content        *string `json:"content"`
	MediaURL       *string `json:"media_url"`
	MediaType      *string `json:"media_type"`
	Price          float64 `json:"price" binding:"required,min=0.99,max=500"`
	MessageType    string  `json:"message_type" binding:"required,oneof=paid_text paid_media"`
}

type MessagesResponse struct {
	Messages []internalmodels.EnhancedMessage `json:"messages"`
	Total    int64                            `json:"total"`
	Page     int                              `json:"page"`
	Limit    int                              `json:"limit"`
	HasMore  bool                             `json:"has_more"`
}

type MessageResponse struct {
	ID             uint        `json:"id"`
	ConversationID uint        `json:"conversation_id"`
	Sender         models.User `json:"sender"`
	Content        *string     `json:"content"`
	MediaURL       *string     `json:"media_url"`
	MediaType      *string     `json:"media_type"`
	ThumbnailURL   *string     `json:"thumbnail_url"`
	IsPaid         bool        `json:"is_paid"`
	Price          *float64    `json:"price"`
	IsUnlocked     bool        `json:"is_unlocked"`
	MessageType    string      `json:"message_type"`
	Status         string      `json:"status"`
	CreatedAt      string      `json:"created_at"`
	ReadAt         *string     `json:"read_at"`
	CanUnlock      bool        `json:"can_unlock"`
}
