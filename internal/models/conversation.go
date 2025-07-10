package models

import (
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/models"
	"gorm.io/gorm"
)

// Conversation représente une conversation entre deux utilisateurs
type Conversation struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	Participant1ID string `gorm:"not null;index" json:"participant_1_id"`
	Participant2ID string `gorm:"not null;index" json:"participant_2_id"`

	// Relations
	Participant1  models.User `gorm:"foreignKey:Participant1ID" json:"participant_1"`
	Participant2  models.User `gorm:"foreignKey:Participant2ID" json:"participant_2"`
	LastMessageID *uint       `json:"last_message_id"`
	LastMessage   *Message    `gorm:"foreignKey:LastMessageID" json:"last_message"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
}

func (Conversation) TableName() string {
	return "conversations"
}

// Hook GORM pour éviter doublons (participant_1_id < participant_2_id)
func (c *Conversation) BeforeCreate(tx *gorm.DB) error {
	// S'assurer que participant_1_id < participant_2_id pour unicité
	if c.Participant1ID > c.Participant2ID {
		c.Participant1ID, c.Participant2ID = c.Participant2ID, c.Participant1ID
	}
	return nil
}

// CreateIndexes crée les index uniques pour éviter conversations dupliquées
func (Conversation) CreateIndexes(db *gorm.DB) error {
	// Pour SQLite, on va créer un index simple sur les deux participants
	// La logique de tri est gérée dans BeforeCreate
	return db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_conversations_participants 
		ON conversations(participant1_id, participant2_id)
		WHERE is_active = true
	`).Error
}

// Helper methods
func (c *Conversation) IsParticipant(userID string) bool {
	return c.Participant1ID == userID || c.Participant2ID == userID
}

func (c *Conversation) GetOtherParticipant(userID string) *models.User {
	if c.Participant1ID == userID {
		return &c.Participant2
	}
	if c.Participant2ID == userID {
		return &c.Participant1
	}
	return nil
}

func (c *Conversation) GetOtherParticipantID(userID string) string {
	if c.Participant1ID == userID {
		return c.Participant2ID
	}
	if c.Participant2ID == userID {
		return c.Participant1ID
	}
	return ""
}
