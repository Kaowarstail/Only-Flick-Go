package models

import (
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/models"
	"gorm.io/gorm"
)

// ConversationClassic représente une conversation entre utilisateurs
type ConversationClassic struct {
	ID        string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations many-to-many avec User
	Participants []models.User `gorm:"many2many:conversation_classic_participants;" json:"participants"`

	// Messages de la conversation
	Messages []MessageClassic `gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE;" json:"messages,omitempty"`

	// Dernier message (pour optimisation affichage)
	LastMessageID *string         `json:"last_message_id,omitempty"`
	LastMessage   *MessageClassic `gorm:"foreignKey:LastMessageID" json:"last_message,omitempty"`

	// État de la conversation
	IsActive bool `gorm:"default:true" json:"is_active"`

	// Type de conversation
	Type string `gorm:"type:varchar(50);default:'direct'" json:"type"` // 'direct', 'group' (future)
}

// TableName spécifie le nom de table
func (ConversationClassic) TableName() string {
	return "conversation_classics"
}

// BeforeCreate génère l'ID et valide
func (c *ConversationClassic) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		uuid, err := generateUUID()
		if err != nil {
			return err
		}
		c.ID = uuid
	}

	if c.Type == "" {
		c.Type = "direct"
	}

	return nil
}

// AfterCreate hook pour post-processing
func (c *ConversationClassic) AfterCreate(tx *gorm.DB) error {
	// Log ou notifications si nécessaire
	return nil
}

// GetOtherParticipant retourne l'autre participant pour conversation directe
func (c *ConversationClassic) GetOtherParticipant(currentUserID string) *models.User {
	for _, participant := range c.Participants {
		if participant.ID != currentUserID {
			return &participant
		}
	}
	return nil
}

// IsParticipant vérifie si un utilisateur fait partie de la conversation
func (c *ConversationClassic) IsParticipant(userID string) bool {
	for _, participant := range c.Participants {
		if participant.ID == userID {
			return true
		}
	}
	return false
}

// GetUnreadCount retourne le nombre de messages non lus pour un utilisateur
func (c *ConversationClassic) GetUnreadCount(db *gorm.DB, userID string) (int64, error) {
	var count int64

	// Sous-requête pour dernière lecture
	subQuery := db.Model(&ConversationClassicReadStatus{}).
		Select("last_read_at").
		Where("conversation_id = ? AND user_id = ?", c.ID, userID)

	// Compter messages non lus
	err := db.Model(&MessageClassic{}).
		Where("conversation_id = ? AND sender_id != ?", c.ID, userID).
		Where("created_at > COALESCE((?), '1970-01-01')", subQuery).
		Count(&count).Error

	return count, err
}
