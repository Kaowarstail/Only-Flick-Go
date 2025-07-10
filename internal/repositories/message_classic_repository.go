package repositories

import (
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/internal/models"
	"gorm.io/gorm"
)

// MessageClassicRepository encapsule les opérations DB pour messages
type MessageClassicRepository struct {
	db *gorm.DB
}

// NewMessageClassicRepository crée un nouveau repository
func NewMessageClassicRepository(db *gorm.DB) *MessageClassicRepository {
	return &MessageClassicRepository{db: db}
}

// GetConversationMessages récupère les messages d'une conversation avec pagination
func (r *MessageClassicRepository) GetConversationMessages(conversationID string, page, limit int) ([]models.MessageClassic, int64, error) {
	var messages []models.MessageClassic
	var total int64

	offset := (page - 1) * limit

	// Compter total
	countQuery := r.db.Model(&models.MessageClassic{}).
		Where("conversation_id = ?", conversationID)

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Récupérer messages (du plus ancien au plus récent pour pagination)
	err := r.db.Scopes(
		models.MessageScope.InConversation(conversationID),
		models.MessageScope.WithSender,
		models.MessageScope.OrderByOldest,
	).Limit(limit).Offset(offset).Find(&messages).Error

	return messages, total, err
}

// CreateMessage crée un nouveau message
func (r *MessageClassicRepository) CreateMessage(message *models.MessageClassic) error {
	return r.db.Create(message).Error
}

// GetMessageByID récupère un message par ID
func (r *MessageClassicRepository) GetMessageByID(messageID string) (*models.MessageClassic, error) {
	var message models.MessageClassic

	err := r.db.Scopes(
		models.MessageScope.WithSender,
	).First(&message, "id = ?", messageID).Error

	return &message, err
}

// MarkMessageAsRead marque un message comme lu
func (r *MessageClassicRepository) MarkMessageAsRead(messageID string) error {
	now := time.Now()

	return r.db.Model(&models.MessageClassic{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"read_at": now,
			"status":  models.MessageStatusRead,
		}).Error
}

// GetUnreadMessagesCount compte les messages non lus pour un utilisateur
func (r *MessageClassicRepository) GetUnreadMessagesCount(userID string) (int64, error) {
	var count int64

	// Sous-requête pour dernière lecture par conversation
	subQuery := r.db.Model(&models.ConversationClassicReadStatus{}).
		Select("conversation_id, last_read_at").
		Where("user_id = ?", userID)

	// Compter messages non lus
	err := r.db.Model(&models.MessageClassic{}).
		Joins("JOIN conversation_classic_participants ccp ON ccp.conversation_classic_id = message_classics.conversation_id").
		Joins("LEFT JOIN (?) crs ON crs.conversation_id = message_classics.conversation_id", subQuery).
		Where("ccp.user_id = ? AND message_classics.sender_id != ?", userID, userID).
		Where("message_classics.created_at > COALESCE(crs.last_read_at, '1970-01-01')").
		Count(&count).Error

	return count, err
}

// DeleteMessage supprime un message (soft delete)
func (r *MessageClassicRepository) DeleteMessage(messageID, userID string) error {
	// Vérifier que l'utilisateur est l'expéditeur
	return r.db.Where("id = ? AND sender_id = ?", messageID, userID).
		Delete(&models.MessageClassic{}).Error
}

// SearchMessages recherche dans le contenu des messages
func (r *MessageClassicRepository) SearchMessages(userID, query string, limit int) ([]models.MessageClassic, error) {
	var messages []models.MessageClassic

	err := r.db.Scopes(
		models.MessageScope.WithSender,
	).
		Joins("JOIN conversation_classic_participants ccp ON ccp.conversation_classic_id = message_classics.conversation_id").
		Where("ccp.user_id = ?", userID).
		Where("message_classics.content ILIKE ?", "%"+query+"%").
		Order("message_classics.created_at DESC").
		Limit(limit).
		Find(&messages).Error

	return messages, err
}
