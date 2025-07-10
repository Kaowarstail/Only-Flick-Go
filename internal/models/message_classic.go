package models

import (
	"strings"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/models"
	"gorm.io/gorm"
)

// MessageType énumération des types de messages
type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeVideo MessageType = "video"
	MessageTypeAudio MessageType = "audio"
)

// MessageStatus énumération des statuts
type MessageStatus string

const (
	MessageStatusSending   MessageStatus = "sending"
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
)

// MessageClassic représente un message dans une conversation
type MessageClassic struct {
	ID        string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	ConversationID string              `gorm:"type:varchar(36);not null;index" json:"conversation_id"`
	Conversation   ConversationClassic `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`

	SenderID string      `gorm:"type:varchar(36);not null;index" json:"sender_id"`
	Sender   models.User `gorm:"foreignKey:SenderID" json:"sender"`

	// Contenu du message
	Content *string `gorm:"type:text" json:"content,omitempty"`

	// Médias (optionnel)
	MediaURL     *string `gorm:"type:varchar(500)" json:"media_url,omitempty"`
	MediaType    *string `gorm:"type:varchar(100)" json:"media_type,omitempty"`
	ThumbnailURL *string `gorm:"type:varchar(500)" json:"thumbnail_url,omitempty"`

	// Type et statut
	MessageType MessageType   `gorm:"type:varchar(20);not null;default:'text'" json:"message_type"`
	Status      MessageStatus `gorm:"type:varchar(20);not null;default:'sent'" json:"status"`

	// Métadonnées de lecture
	ReadAt *time.Time `json:"read_at,omitempty"`
}

// TableName spécifie le nom de table
func (MessageClassic) TableName() string {
	return "message_classics"
}

// BeforeCreate hook GORM
func (m *MessageClassic) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		uuid, err := generateUUID()
		if err != nil {
			return err
		}
		m.ID = uuid
	}

	// Validation contenu
	if err := m.ValidateContent(); err != nil {
		return err
	}

	// Status par défaut
	if m.Status == "" {
		m.Status = MessageStatusSent
	}

	// Type par défaut
	if m.MessageType == "" {
		m.MessageType = MessageTypeText
	}

	return nil
}

// AfterCreate met à jour last_message_id de la conversation
func (m *MessageClassic) AfterCreate(tx *gorm.DB) error {
	return tx.Model(&ConversationClassic{}).
		Where("id = ?", m.ConversationID).
		Updates(map[string]interface{}{
			"last_message_id": m.ID,
			"updated_at":      m.CreatedAt,
		}).Error
}

// ValidateContent valide le contenu du message
func (m *MessageClassic) ValidateContent() error {
	// Au moins contenu OU média requis
	hasContent := m.Content != nil && strings.TrimSpace(*m.Content) != ""
	hasMedia := m.MediaURL != nil && strings.TrimSpace(*m.MediaURL) != ""

	if !hasContent && !hasMedia {
		return gorm.ErrInvalidValue
	}

	// Validation longueur contenu
	if hasContent && len(*m.Content) > 5000 {
		return gorm.ErrInvalidValue
	}

	// Si média présent, type requis
	if hasMedia && (m.MediaType == nil || strings.TrimSpace(*m.MediaType) == "") {
		return gorm.ErrInvalidValue
	}

	return nil
}

// IsMediaMessage vérifie si le message contient un média
func (m *MessageClassic) IsMediaMessage() bool {
	return m.MediaURL != nil && *m.MediaURL != ""
}

// GetDisplayContent retourne le contenu à afficher
func (m *MessageClassic) GetDisplayContent() string {
	if m.Content != nil && *m.Content != "" {
		return *m.Content
	}

	switch m.MessageType {
	case MessageTypeImage:
		return "📸 Image"
	case MessageTypeVideo:
		return "🎥 Vidéo"
	case MessageTypeAudio:
		return "🎵 Audio"
	default:
		return "💬 Message"
	}
}

// MarkAsRead marque le message comme lu
func (m *MessageClassic) MarkAsRead(db *gorm.DB) error {
	now := time.Now()
	m.ReadAt = &now
	m.Status = MessageStatusRead

	return db.Model(m).Updates(map[string]interface{}{
		"read_at": now,
		"status":  MessageStatusRead,
	}).Error
}
