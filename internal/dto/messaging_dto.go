package dto

import (
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/internal/models"
	models2 "github.com/Kaowarstail/Only-Flick-Go/models"
	"gorm.io/gorm"
)

// ConversationClassicDTO pour sérialisation API
type ConversationClassicDTO struct {
	ID           string             `json:"id"`
	Type         string             `json:"type"`
	IsActive     bool               `json:"is_active"`
	Participants []UserDTO          `json:"participants"`
	LastMessage  *MessageClassicDTO `json:"last_message,omitempty"`
	UnreadCount  int64              `json:"unread_count,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

// MessageClassicDTO pour sérialisation API
type MessageClassicDTO struct {
	ID             string               `json:"id"`
	ConversationID string               `json:"conversation_id"`
	Sender         UserDTO              `json:"sender"`
	Content        *string              `json:"content,omitempty"`
	MediaURL       *string              `json:"media_url,omitempty"`
	MediaType      *string              `json:"media_type,omitempty"`
	ThumbnailURL   *string              `json:"thumbnail_url,omitempty"`
	MessageType    models.MessageType   `json:"message_type"`
	Status         models.MessageStatus `json:"status"`
	ReadAt         *time.Time           `json:"read_at,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
}

// UserDTO structure basique pour éviter récursion
type UserDTO struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// ToConversationDTO convertit ConversationClassic vers DTO
func ToConversationDTO(c *models.ConversationClassic, currentUserID string, db *gorm.DB) ConversationClassicDTO {
	result := ConversationClassicDTO{
		ID:           c.ID,
		Type:         c.Type,
		IsActive:     c.IsActive,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
		Participants: make([]UserDTO, len(c.Participants)),
	}

	// Convertir participants
	for i, participant := range c.Participants {
		result.Participants[i] = UserDTO{
			ID:          participant.ID,
			Username:    participant.Username,
			DisplayName: getDisplayName(&participant),
			AvatarURL:   getAvatarURL(&participant),
		}
	}

	// Convertir dernier message
	if c.LastMessage != nil {
		dto := ToMessageDTO(c.LastMessage)
		result.LastMessage = &dto
	}

	// Calculer messages non lus
	if count, err := c.GetUnreadCount(db, currentUserID); err == nil {
		result.UnreadCount = count
	}

	return result
}

// ToMessageDTO convertit MessageClassic vers DTO
func ToMessageDTO(m *models.MessageClassic) MessageClassicDTO {
	return MessageClassicDTO{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		Sender: UserDTO{
			ID:          m.Sender.ID,
			Username:    m.Sender.Username,
			DisplayName: getDisplayName(&m.Sender),
			AvatarURL:   getAvatarURL(&m.Sender),
		},
		Content:      m.Content,
		MediaURL:     m.MediaURL,
		MediaType:    m.MediaType,
		ThumbnailURL: m.ThumbnailURL,
		MessageType:  m.MessageType,
		Status:       m.Status,
		ReadAt:       m.ReadAt,
		CreatedAt:    m.CreatedAt,
	}
}

// Helpers pour récupérer nom et avatar (à adapter selon votre model User)
func getDisplayName(user *models2.User) string {
	if user.FirstName != "" && user.LastName != "" {
		return user.FirstName + " " + user.LastName
	}
	return user.Username
}

func getAvatarURL(user *models2.User) string {
	if user.ProfilePicture != "" {
		return user.ProfilePicture
	}
	return ""
}
