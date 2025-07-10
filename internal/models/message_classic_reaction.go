package models

import (
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/models"
	"gorm.io/gorm"
)

// MessageClassicReaction représente une réaction à un message
type MessageClassicReaction struct {
	ID        string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	MessageID string         `gorm:"type:varchar(36);not null;index:idx_message_user_reaction" json:"message_id"`
	Message   MessageClassic `gorm:"foreignKey:MessageID" json:"message,omitempty"`

	UserID string      `gorm:"type:varchar(36);not null;index:idx_message_user_reaction" json:"user_id"`
	User   models.User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Réaction
	Emoji string `gorm:"type:varchar(10);not null" json:"emoji"` // 👍, ❤️, 😂, etc.
}

// TableName spécifie le nom de table
func (MessageClassicReaction) TableName() string {
	return "message_classic_reactions"
}

// BeforeCreate hook GORM
func (mr *MessageClassicReaction) BeforeCreate(tx *gorm.DB) error {
	if mr.ID == "" {
		uuid, err := generateUUID()
		if err != nil {
			return err
		}
		mr.ID = uuid
	}

	// Validation emoji basique
	if mr.Emoji == "" {
		return gorm.ErrInvalidValue
	}

	return nil
}
