package models

import (
	"time"

	"gorm.io/gorm"
)

// ConversationClassicScope contient les scopes pour ConversationClassic
type ConversationClassicScope struct{}

// Active filtre les conversations actives
func (ConversationClassicScope) Active(db *gorm.DB) *gorm.DB {
	return db.Where("is_active = ?", true)
}

// WithParticipant filtre par participant
func (ConversationClassicScope) WithParticipant(userID string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins("JOIN conversation_classic_participants ccp ON ccp.conversation_classic_id = conversation_classics.id").
			Where("ccp.user_id = ?", userID)
	}
}

// OrderByLastActivity trie par activité récente
func (ConversationClassicScope) OrderByLastActivity(db *gorm.DB) *gorm.DB {
	return db.Order("conversation_classics.updated_at DESC")
}

// WithLastMessage preload le dernier message
func (ConversationClassicScope) WithLastMessage(db *gorm.DB) *gorm.DB {
	return db.Preload("LastMessage").Preload("LastMessage.Sender")
}

// WithParticipants preload les participants
func (ConversationClassicScope) WithParticipants(db *gorm.DB) *gorm.DB {
	return db.Preload("Participants")
}

// MessageClassicScope contient les scopes pour MessageClassic
type MessageClassicScope struct{}

// InConversation filtre par conversation
func (MessageClassicScope) InConversation(conversationID string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("conversation_id = ?", conversationID)
	}
}

// FromSender filtre par expéditeur
func (MessageClassicScope) FromSender(senderID string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sender_id = ?", senderID)
	}
}

// NotFromSender exclut un expéditeur
func (MessageClassicScope) NotFromSender(senderID string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sender_id != ?", senderID)
	}
}

// AfterTime filtre après une date
func (MessageClassicScope) AfterTime(t time.Time) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at > ?", t)
	}
}

// Unread filtre les messages non lus pour un utilisateur
func (MessageClassicScope) Unread(userID string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		subQuery := db.Model(&ConversationClassicReadStatus{}).
			Select("last_read_at").
			Where("user_id = ? AND conversation_id = message_classics.conversation_id", userID)

		return db.Where("sender_id != ?", userID).
			Where("created_at > COALESCE((?), '1970-01-01')", subQuery)
	}
}

// ByType filtre par type de message
func (MessageClassicScope) ByType(messageType MessageType) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("message_type = ?", messageType)
	}
}

// WithSender preload l'expéditeur
func (MessageClassicScope) WithSender(db *gorm.DB) *gorm.DB {
	return db.Preload("Sender")
}

// OrderByNewest trie par plus récent
func (MessageClassicScope) OrderByNewest(db *gorm.DB) *gorm.DB {
	return db.Order("created_at DESC")
}

// OrderByOldest trie par plus ancien
func (MessageClassicScope) OrderByOldest(db *gorm.DB) *gorm.DB {
	return db.Order("created_at ASC")
}

// Instances globales pour utilisation facile
var ConversationScope ConversationClassicScope
var MessageScope MessageClassicScope
