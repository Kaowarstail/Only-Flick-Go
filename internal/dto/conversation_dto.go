package dto

import (
	"errors"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/internal/models"
	userModels "github.com/Kaowarstail/Only-Flick-Go/models"
)

// CreateConversationRequest représente une demande pour créer une conversation
type CreateConversationRequest struct {
	OtherUserID string `json:"other_user_id" binding:"required"`
}

// Validate effectue une validation custom de la requête
func (req *CreateConversationRequest) Validate(currentUserID string) error {
	if req.OtherUserID == currentUserID {
		return errors.New("cannot create conversation with yourself")
	}

	if req.OtherUserID == "" {
		return errors.New("other_user_id is required")
	}

	return nil
}

// ConversationResponse représente une conversation enrichie pour la réponse
type ConversationResponse struct {
	ID           uint                  `json:"id"`
	Participants []userModels.User     `json:"participants"`
	LastMessage  *models.Message       `json:"last_message"`
	UnreadCount  int                   `json:"unread_count"`
	UpdatedAt    time.Time             `json:"updated_at"`
	IsActive     bool                  `json:"is_active"`
	OtherUser    *userModels.User      `json:"other_user"` // L'autre participant (pas moi)
}

// ConversationsResponse représente la réponse pour une liste de conversations
type ConversationsResponse struct {
	Conversations []ConversationResponse `json:"conversations"`
	Total         int64                  `json:"total"`
	UnreadTotal   int                    `json:"unread_total"`
	Page          int                    `json:"page"`
	Limit         int                    `json:"limit"`
	HasMore       bool                   `json:"has_more"`
}

// MarkConversationReadRequest représente une demande pour marquer conversation comme lue
type MarkConversationReadRequest struct {
	ConversationID uint `json:"conversation_id" binding:"required"`
}

// ConversationStatsResponse représente les statistiques de conversations
type ConversationStatsResponse struct {
	TotalConversations  int `json:"total_conversations"`
	ActiveConversations int `json:"active_conversations"`
	UnreadConversations int `json:"unread_conversations"`
	TotalUnreadMessages int `json:"total_unread_messages"`
}

// SearchConversationsRequest représente une demande de recherche dans les conversations
type SearchConversationsRequest struct {
	Query string `json:"query" binding:"required,min=1"`
	Page  int    `json:"page" binding:"omitempty,min=1"`
	Limit int    `json:"limit" binding:"omitempty,min=1,max=50"`
}

// GetConversationsRequest représente une demande pour récupérer les conversations
type GetConversationsRequest struct {
	Page   int  `json:"page" binding:"omitempty,min=1"`
	Limit  int  `json:"limit" binding:"omitempty,min=1,max=50"`
	Unread bool `json:"unread"`
}

// ConversationParticipantResponse représente un participant dans une conversation
type ConversationParticipantResponse struct {
	UserID         string    `json:"user_id"`
	Username       string    `json:"username"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	ProfilePicture string    `json:"profile_picture"`
	IsOnline       bool      `json:"is_online"`
	LastSeen       time.Time `json:"last_seen"`
}

// NewConversationResponse représente la réponse lors de la création d'une conversation
type NewConversationResponse struct {
	Conversation ConversationResponse `json:"conversation"`
	IsNew        bool                 `json:"is_new"` // false si conversation existait déjà
}
