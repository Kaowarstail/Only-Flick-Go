package models

import (
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/models"
	"gorm.io/gorm"
)

type Conversation struct {
	ID             uint             `gorm:"primaryKey" json:"id"`
	Participant1ID string           `gorm:"not null;index" json:"participant_1_id"`
	Participant2ID string           `gorm:"not null;index" json:"participant_2_id"`
	Participant1   models.User      `gorm:"foreignKey:Participant1ID" json:"participant_1"`
	Participant2   models.User      `gorm:"foreignKey:Participant2ID" json:"participant_2"`
	LastMessageID  *uint            `json:"last_message_id"`
	LastMessage    *EnhancedMessage `gorm:"foreignKey:LastMessageID" json:"last_message"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	IsActive       bool             `gorm:"default:true" json:"is_active"`
}

func (Conversation) TableName() string {
	return "conversations"
}

// Hook pour éviter doublons (participant_1_id < participant_2_id)
func (c *Conversation) BeforeCreate(tx *gorm.DB) error {
	if c.Participant1ID > c.Participant2ID {
		c.Participant1ID, c.Participant2ID = c.Participant2ID, c.Participant1ID
	}
	return nil
}

// Index composite pour optimiser les requêtes
func (Conversation) CreateIndexes(db *gorm.DB) error {
	return db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_conversation_participants ON conversations(participant_1_id, participant_2_id)").Error
}
