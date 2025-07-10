package models

import (
	"errors"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/models"
)

// Message représente un message dans une conversation (version améliorée)
type Message struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ConversationID uint      `gorm:"not null;index" json:"conversation_id"`
	SenderID       string    `gorm:"not null;index" json:"sender_id"`

	// Relations
	Conversation Conversation `gorm:"foreignKey:ConversationID" json:"-"`
	Sender       models.User  `gorm:"foreignKey:SenderID" json:"sender"`

	// Contenu
	Content      *string `gorm:"type:text" json:"content"`
	MediaURL     *string `json:"media_url"`
	MediaType    *string `json:"media_type"` // 'image', 'video', 'audio'
	ThumbnailURL *string `json:"thumbnail_url"`

	// Metadata
	MessageType string     `gorm:"default:text" json:"message_type"` // 'text', 'image', 'video', 'audio'
	Status      string     `gorm:"default:sent;index" json:"status"` // 'sending', 'sent', 'delivered', 'read', 'failed'
	CreatedAt   time.Time  `gorm:"index" json:"created_at"`
	ReadAt      *time.Time `json:"read_at"`
}

func (Message) TableName() string {
	return "messages"
}

// Validation business logic
func (m *Message) Validate() error {
	// Un message doit avoir du contenu OU un média
	if (m.Content == nil || *m.Content == "") && (m.MediaURL == nil || *m.MediaURL == "") {
		return errors.New("message must have content or media")
	}

	// Validation longueur contenu
	if m.Content != nil && len(*m.Content) > 5000 {
		return errors.New("message content too long (max 5000 characters)")
	}

	return nil
}

// Helper methods
func (m *Message) IsReadBy(userID string) bool {
	return m.ReadAt != nil || m.SenderID == userID
}

func (m *Message) MarkAsRead() {
	if m.ReadAt == nil {
		now := time.Now()
		m.ReadAt = &now
	}
}

func (m *Message) GetDisplayContent() string {
	if m.Content != nil {
		return *m.Content
	}

	// Messages médias
	switch m.MessageType {
	case "image":
		return "📸 Image"
	case "video":
		return "🎥 Vidéo"
	case "audio":
		return "🎵 Audio"
	default:
		return "💬 Message"
	}
}

func (m *Message) IsMediaMessage() bool {
	return m.MediaURL != nil && *m.MediaURL != ""
}

func (m *Message) HasContent() bool {
	return m.Content != nil && *m.Content != ""
}
