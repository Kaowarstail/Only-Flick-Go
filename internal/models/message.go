package models

import (
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/models"
	"gorm.io/gorm"
)

type EnhancedMessage struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	ConversationID uint         `gorm:"not null;index" json:"conversation_id"`
	Conversation   Conversation `gorm:"foreignKey:ConversationID" json:"-"`
	SenderID       string       `gorm:"not null;index" json:"sender_id"`
	Sender         models.User  `gorm:"foreignKey:SenderID" json:"sender"`
	Content        *string      `gorm:"type:text" json:"content"`
	MediaURL       *string      `json:"media_url"`
	MediaType      *string      `json:"media_type"` // 'image', 'video', 'audio'
	ThumbnailURL   *string      `json:"thumbnail_url"`

	// Messages payants
	IsPaid     bool       `gorm:"default:false;index" json:"is_paid"`
	Price      *float64   `gorm:"type:decimal(10,2)" json:"price"`
	IsUnlocked bool       `gorm:"default:false" json:"is_unlocked"`
	UnlockedAt *time.Time `json:"unlocked_at"`
	UnlockedBy *string    `json:"unlocked_by"`

	MessageType string `gorm:"default:text" json:"message_type"` // 'text', 'image', 'video', 'paid_text', 'paid_media'
	Status      string `gorm:"default:sent;index" json:"status"` // 'sending', 'sent', 'delivered', 'read', 'failed'

	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at"`

	// Relations
	UnlockedByUser *models.User `gorm:"foreignKey:UnlockedBy" json:"unlocked_by_user,omitempty"`
}

func (EnhancedMessage) TableName() string {
	return "enhanced_messages"
}

// Index pour optimiser les requêtes
func (EnhancedMessage) CreateIndexes(db *gorm.DB) error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_enhanced_message_conversation_created ON enhanced_messages(conversation_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_enhanced_message_sender_created ON enhanced_messages(sender_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_enhanced_message_unread ON enhanced_messages(conversation_id, sender_id, read_at)",
	}

	for _, index := range indexes {
		if err := db.Exec(index).Error; err != nil {
			return err
		}
	}
	return nil
}
