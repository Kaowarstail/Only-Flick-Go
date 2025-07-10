package models

import (
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/models"
	"gorm.io/gorm"
)

// ConversationClassicReadStatus suit le statut de lecture par utilisateur
type ConversationClassicReadStatus struct {
	ID        string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	ConversationID string              `gorm:"type:varchar(36);not null;index:idx_conversation_user" json:"conversation_id"`
	Conversation   ConversationClassic `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`

	UserID string      `gorm:"type:varchar(36);not null;index:idx_conversation_user" json:"user_id"`
	User   models.User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Dernière lecture
	LastReadAt time.Time `gorm:"not null" json:"last_read_at"`
}

// TableName spécifie le nom de table
func (ConversationClassicReadStatus) TableName() string {
	return "conversation_classic_read_statuses"
}

// BeforeCreate hook GORM
func (crs *ConversationClassicReadStatus) BeforeCreate(tx *gorm.DB) error {
	if crs.ID == "" {
		uuid, err := generateUUID()
		if err != nil {
			return err
		}
		crs.ID = uuid
	}

	if crs.LastReadAt.IsZero() {
		crs.LastReadAt = time.Now()
	}

	return nil
}

// UpdateLastRead met à jour la dernière lecture
func (crs *ConversationClassicReadStatus) UpdateLastRead(db *gorm.DB) error {
	crs.LastReadAt = time.Now()
	return db.Save(crs).Error
}
