package repositories

import (
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/internal/models"
	usermodels "github.com/Kaowarstail/Only-Flick-Go/models"
	"gorm.io/gorm"
)

// ConversationClassicRepository encapsule les opérations DB
type ConversationClassicRepository struct {
	db *gorm.DB
}

// NewConversationClassicRepository crée un nouveau repository
func NewConversationClassicRepository(db *gorm.DB) *ConversationClassicRepository {
	return &ConversationClassicRepository{db: db}
}

// GetUserConversations récupère les conversations d'un utilisateur avec pagination
func (r *ConversationClassicRepository) GetUserConversations(userID string, page, limit int) ([]models.ConversationClassic, int64, error) {
	var conversations []models.ConversationClassic
	var total int64

	offset := (page - 1) * limit

	// Compter total
	countQuery := r.db.Model(&models.ConversationClassic{}).
		Joins("JOIN conversation_classic_participants ccp ON ccp.conversation_classic_id = conversation_classics.id").
		Where("ccp.user_id = ? AND conversation_classics.is_active = ?", userID, true)

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Récupérer conversations
	err := r.db.Scopes(
		models.ConversationScope.Active,
		models.ConversationScope.WithParticipant(userID),
		models.ConversationScope.WithParticipants,
		models.ConversationScope.WithLastMessage,
		models.ConversationScope.OrderByLastActivity,
	).Limit(limit).Offset(offset).Find(&conversations).Error

	return conversations, total, err
}

// CreateDirectConversation crée une conversation directe
func (r *ConversationClassicRepository) CreateDirectConversation(user1ID, user2ID string) (*models.ConversationClassic, error) {
	conversation := models.ConversationClassic{
		Type:     "direct",
		IsActive: true,
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Créer conversation
		if err := tx.Create(&conversation).Error; err != nil {
			return err
		}

		// Ajouter participants
		var users []usermodels.User
		if err := tx.Find(&users, "id IN ?", []string{user1ID, user2ID}).Error; err != nil {
			return err
		}

		if len(users) != 2 {
			return gorm.ErrRecordNotFound
		}

		return tx.Model(&conversation).Association("Participants").Append(users)
	})

	if err != nil {
		return nil, err
	}

	// Recharger avec relations
	err = r.db.Scopes(
		models.ConversationScope.WithParticipants,
	).First(&conversation, "id = ?", conversation.ID).Error

	return &conversation, err
}

// FindDirectConversation trouve une conversation directe existante
func (r *ConversationClassicRepository) FindDirectConversation(user1ID, user2ID string) (*models.ConversationClassic, error) {
	var conversation models.ConversationClassic

	err := r.db.Scopes(
		models.ConversationScope.WithParticipants,
		models.ConversationScope.WithLastMessage,
	).
		Joins("JOIN conversation_classic_participants ccp1 ON ccp1.conversation_classic_id = conversation_classics.id").
		Joins("JOIN conversation_classic_participants ccp2 ON ccp2.conversation_classic_id = conversation_classics.id").
		Where("ccp1.user_id = ? AND ccp2.user_id = ? AND conversation_classics.type = 'direct'", user1ID, user2ID).
		First(&conversation).Error

	return &conversation, err
}

// GetConversationByID récupère une conversation par ID
func (r *ConversationClassicRepository) GetConversationByID(conversationID string, userID string) (*models.ConversationClassic, error) {
	var conversation models.ConversationClassic

	err := r.db.Scopes(
		models.ConversationScope.WithParticipants,
		models.ConversationScope.WithLastMessage,
		models.ConversationScope.WithParticipant(userID), // Vérifier que user est participant
	).First(&conversation, "id = ?", conversationID).Error

	return &conversation, err
}

// MarkConversationAsRead marque une conversation comme lue
func (r *ConversationClassicRepository) MarkConversationAsRead(conversationID, userID string) error {
	// Upsert du statut de lecture
	readStatus := models.ConversationClassicReadStatus{
		ConversationID: conversationID,
		UserID:         userID,
		LastReadAt:     time.Now(),
	}

	return r.db.Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Assign(models.ConversationClassicReadStatus{LastReadAt: time.Now()}).
		FirstOrCreate(&readStatus).Error
}
