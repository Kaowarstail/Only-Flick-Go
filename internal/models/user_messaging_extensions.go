package models

import (
	"github.com/Kaowarstail/Only-Flick-Go/models"
	"gorm.io/gorm"
)

// UserMessagingExtensions contient les méthodes d'extension pour User liées à la messagerie
type UserMessagingExtensions struct{}

// GetConversations retourne les conversations d'un utilisateur
func GetUserConversations(db *gorm.DB, userID string, page, limit int) ([]ConversationClassic, error) {
	var conversations []ConversationClassic

	offset := (page - 1) * limit

	err := db.
		Preload("Participants").
		Preload("LastMessage").
		Preload("LastMessage.Sender").
		Joins("JOIN conversation_classic_participants ccp ON ccp.conversation_classic_id = conversation_classics.id").
		Where("ccp.user_id = ? AND conversation_classics.is_active = ?", userID, true).
		Order("conversation_classics.updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&conversations).Error

	return conversations, err
}

// GetConversationWith trouve ou crée une conversation avec un autre utilisateur
func GetUserConversationWith(db *gorm.DB, userID, otherUserID string) (*ConversationClassic, bool, error) {
	var conversation ConversationClassic

	// Chercher conversation existante
	err := db.
		Preload("Participants").
		Preload("LastMessage").
		Joins("JOIN conversation_classic_participants ccp1 ON ccp1.conversation_classic_id = conversation_classics.id").
		Joins("JOIN conversation_classic_participants ccp2 ON ccp2.conversation_classic_id = conversation_classics.id").
		Where("ccp1.user_id = ? AND ccp2.user_id = ? AND conversation_classics.type = 'direct'", userID, otherUserID).
		First(&conversation).Error

	if err == nil {
		// Conversation trouvée
		return &conversation, false, nil
	}

	if err != gorm.ErrRecordNotFound {
		// Erreur autre que "pas trouvé"
		return nil, false, err
	}

	// Créer nouvelle conversation
	conversation = ConversationClassic{
		Type:     "direct",
		IsActive: true,
	}

	// Transaction pour créer conversation + participants
	err = db.Transaction(func(tx *gorm.DB) error {
		// Créer conversation
		if err := tx.Create(&conversation).Error; err != nil {
			return err
		}

		// Ajouter participants
		var users []models.User
		if err := tx.Find(&users, "id IN ?", []string{userID, otherUserID}).Error; err != nil {
			return err
		}

		if len(users) != 2 {
			return gorm.ErrRecordNotFound
		}

		return tx.Model(&conversation).Association("Participants").Append(users)
	})

	if err != nil {
		return nil, false, err
	}

	// Recharger avec relations
	err = db.
		Preload("Participants").
		First(&conversation, "id = ?", conversation.ID).Error

	return &conversation, true, err
}

// GetUnreadMessagesCount retourne le nombre total de messages non lus
func GetUserUnreadMessagesCount(db *gorm.DB, userID string) (int64, error) {
	var count int64

	// Sous-requête pour dernière lecture par conversation
	subQuery := db.Model(&ConversationClassicReadStatus{}).
		Select("conversation_id, last_read_at").
		Where("user_id = ?", userID)

	// Compter messages non lus
	err := db.Model(&MessageClassic{}).
		Joins("JOIN conversation_classic_participants ccp ON ccp.conversation_classic_id = message_classics.conversation_id").
		Joins("LEFT JOIN (?) crs ON crs.conversation_id = message_classics.conversation_id", subQuery).
		Where("ccp.user_id = ? AND message_classics.sender_id != ?", userID, userID).
		Where("message_classics.created_at > COALESCE(crs.last_read_at, '1970-01-01')").
		Count(&count).Error

	return count, err
}

// CanMessageUser vérifie si l'utilisateur peut envoyer un message
func CanUserMessageUser(db *gorm.DB, userID, targetUserID string) (bool, error) {
	// Vérifications basiques
	if userID == targetUserID {
		return false, nil // Pas de message à soi-même
	}

	// Vérifier si l'utilisateur cible existe
	var targetUser models.User
	if err := db.First(&targetUser, "id = ?", targetUserID).Error; err != nil {
		return false, err
	}

	// TODO: Ajouter logique de blocage si nécessaire
	// - Vérifier si bloqué
	// - Vérifier paramètres de confidentialité
	// - Vérifier si creator et abonné

	return true, nil
}
