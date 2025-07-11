package models

import (
	"time"

	"gorm.io/gorm"
)

// ConversationType définit le type de conversation
type ConversationType string

const (
	ConversationTypeDirect ConversationType = "direct"
	ConversationTypeGroup  ConversationType = "group"
)

// MessageType définit le type de message
type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeVideo MessageType = "video"
	MessageTypeAudio MessageType = "audio"
	MessageTypeFile  MessageType = "file"
)

// MessageStatus définit le statut d'un message
type MessageStatus string

const (
	MessageStatusSending   MessageStatus = "sending"
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
)

// ConversationClassic représente une conversation moderne
type ConversationClassic struct {
	ID        string           `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Type      ConversationType `json:"type" gorm:"type:varchar(20);not null"`
	IsActive  bool             `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt   `json:"-" gorm:"index"`

	// Relations
	Participants []ConversationClassicParticipant `json:"-" gorm:"foreignKey:ConversationClassicID"`
	Messages     []MessageClassic                 `json:"-" gorm:"foreignKey:ConversationID"`
	LastMessage  *MessageClassic                  `json:"-" gorm:"foreignKey:ConversationID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

// ConversationClassicParticipant représente un participant à une conversation
type ConversationClassicParticipant struct {
	ID                    uint      `json:"id" gorm:"primaryKey"`
	ConversationClassicID string    `json:"conversation_classic_id" gorm:"type:varchar(36);not null"`
	UserID                string    `json:"user_id" gorm:"type:varchar(36);not null"`
	JoinedAt              time.Time `json:"joined_at" gorm:"autoCreateTime"`

	// Relations
	Conversation ConversationClassic `json:"-" gorm:"foreignKey:ConversationClassicID"`
	User         User                `json:"-" gorm:"foreignKey:UserID"`
}

// MessageClassic représente un message moderne
type MessageClassic struct {
	ID             string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ConversationID string         `json:"conversation_id" gorm:"type:varchar(36);not null"`
	SenderID       string         `json:"sender_id" gorm:"type:varchar(36);not null"`
	Content        *string        `json:"content,omitempty"`
	MediaURL       *string        `json:"media_url,omitempty"`
	MediaType      *string        `json:"media_type,omitempty"`
	ThumbnailURL   *string        `json:"thumbnail_url,omitempty"`
	MessageType    MessageType    `json:"message_type" gorm:"type:varchar(20);not null"`
	Status         MessageStatus  `json:"status" gorm:"type:varchar(20);default:'sent'"`
	ReadAt         *time.Time     `json:"read_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Conversation ConversationClassic `json:"-" gorm:"foreignKey:ConversationID"`
	Sender       User                `json:"-" gorm:"foreignKey:SenderID"`
}

// ConversationClassicReadStatus suit le statut de lecture des conversations
type ConversationClassicReadStatus struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	ConversationID string    `json:"conversation_id" gorm:"type:varchar(36);not null"`
	UserID         string    `json:"user_id" gorm:"type:varchar(36);not null"`
	LastReadAt     time.Time `json:"last_read_at" gorm:"not null"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	Conversation ConversationClassic `json:"-" gorm:"foreignKey:ConversationID"`
	User         User                `json:"-" gorm:"foreignKey:UserID"`
}

// Hook GORM pour mettre à jour LastMessage après création d'un message
func (m *MessageClassic) AfterCreate(tx *gorm.DB) error {
	// Mettre à jour le lastMessage de la conversation
	return tx.Model(&ConversationClassic{}).
		Where("id = ?", m.ConversationID).
		Updates(map[string]interface{}{
			"updated_at": time.Now(),
		}).Error
}

// Hook GORM pour mettre à jour la conversation après mise à jour d'un message
func (m *MessageClassic) AfterUpdate(tx *gorm.DB) error {
	// Mettre à jour le timestamp de la conversation
	return tx.Model(&ConversationClassic{}).
		Where("id = ?", m.ConversationID).
		Update("updated_at", time.Now()).Error
}

// GetUnreadCount retourne le nombre de messages non lus dans une conversation pour un utilisateur
func (c *ConversationClassic) GetUnreadCount(db *gorm.DB, userID string) (int64, error) {
	var count int64

	// Récupérer la dernière lecture
	var readStatus ConversationClassicReadStatus
	err := db.Where("conversation_id = ? AND user_id = ?", c.ID, userID).
		First(&readStatus).Error

	lastReadAt := time.Unix(0, 0) // Epoch par défaut
	if err == nil {
		lastReadAt = readStatus.LastReadAt
	}

	// Compter les messages non lus
	err = db.Model(&MessageClassic{}).
		Where("conversation_id = ? AND sender_id != ? AND created_at > ?",
			c.ID, userID, lastReadAt).
		Count(&count).Error

	return count, err
}

// ToDTO convertit un MessageClassic en DTO
func (m *MessageClassic) ToDTO() MessageClassicDTO {
	var readAt *string
	if m.ReadAt != nil {
		readAtStr := m.ReadAt.Format("2006-01-02T15:04:05Z07:00")
		readAt = &readAtStr
	}

	return MessageClassicDTO{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		Sender: UserDTO{
			ID:          m.SenderID,
			Username:    m.Sender.Username,
			DisplayName: m.Sender.FirstName + " " + m.Sender.LastName,
			AvatarURL:   m.Sender.ProfilePicture,
		},
		Content:      m.Content,
		MediaURL:     m.MediaURL,
		MediaType:    m.MediaType,
		ThumbnailURL: m.ThumbnailURL,
		MessageType:  string(m.MessageType),
		Status:       string(m.Status),
		ReadAt:       readAt,
		CreatedAt:    m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// MessageClassicDTO pour sérialisation API
type MessageClassicDTO struct {
	ID             string  `json:"id"`
	ConversationID string  `json:"conversation_id"`
	Sender         UserDTO `json:"sender"`
	Content        *string `json:"content,omitempty"`
	MediaURL       *string `json:"media_url,omitempty"`
	MediaType      *string `json:"media_type,omitempty"`
	ThumbnailURL   *string `json:"thumbnail_url,omitempty"`
	MessageType    string  `json:"message_type"`
	Status         string  `json:"status"`
	ReadAt         *string `json:"read_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

// UserDTO structure basique pour éviter récursion
type UserDTO struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// GetUnreadMessagesCount retourne le nombre total de messages non lus pour un utilisateur
func (u *User) GetUnreadMessagesCount(db *gorm.DB) (int64, error) {
	var count int64

	// Sous-requête pour récupérer les dernières lectures
	subQuery := db.Model(&ConversationClassicReadStatus{}).
		Select("conversation_id, COALESCE(last_read_at, '1970-01-01') as last_read").
		Where("user_id = ?", u.ID)

	// Compter tous les messages non lus dans toutes les conversations
	err := db.Model(&MessageClassic{}).
		Joins("JOIN conversation_classic_participants ccp ON ccp.conversation_classic_id = message_classics.conversation_id").
		Joins("LEFT JOIN (?) rs ON rs.conversation_id = message_classics.conversation_id", subQuery).
		Where("ccp.user_id = ? AND message_classics.sender_id != ?", u.ID, u.ID).
		Where("message_classics.created_at > COALESCE(rs.last_read, '1970-01-01')").
		Count(&count).Error

	return count, err
}

// Index pour optimiser les performances
func (ConversationClassic) TableName() string {
	return "conversation_classics"
}

func (MessageClassic) TableName() string {
	return "message_classics"
}

func (ConversationClassicParticipant) TableName() string {
	return "conversation_classic_participants"
}

func (ConversationClassicReadStatus) TableName() string {
	return "conversation_classic_read_statuses"
}
