package services

import (
	"errors"
	"fmt"

	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
	"github.com/Kaowarstail/Only-Flick-Go/models"
	"gorm.io/gorm"
)

// ConversationService gère la logique métier des conversations
type ConversationService struct {
	db *gorm.DB
}

// NewConversationService crée une nouvelle instance
func NewConversationService(db *gorm.DB) *ConversationService {
	return &ConversationService{
		db: db,
	}
}

// ConversationServiceInterface définit les méthodes du service
type ConversationServiceInterface interface {
	// CRUD Operations
	GetUserConversations(userID string, page, limit int) (*ConversationsResponse, error)
	GetConversationByID(conversationID, userID string) (*ConversationClassicDTO, error)
	CreateOrGetDirectConversation(userID, otherUserID string) (*ConversationClassicDTO, error)

	// Read Status
	MarkConversationAsRead(conversationID, userID string) error
	GetUnreadConversationsCount(userID string) (int64, error)

	// Management
	ArchiveConversation(conversationID, userID string) error
	DeleteConversation(conversationID, userID string) error

	// Search & Validation
	SearchConversations(userID, query string, page, limit int) ([]ConversationClassicDTO, error)
	CanUserAccessConversation(conversationID, userID string) (bool, error)
}

// ConversationsResponse représente une liste de conversations
type ConversationsResponse struct {
	Conversations []ConversationClassicDTO `json:"conversations"`
	Pagination    PaginationResponse       `json:"pagination"`
	UnreadTotal   int64                    `json:"unread_total"`
}

// PaginationResponse représente les informations de pagination
type PaginationResponse struct {
	Page    int   `json:"page"`
	Limit   int   `json:"limit"`
	Total   int64 `json:"total"`
	HasMore bool  `json:"has_more"`
}

// ConversationClassicDTO pour sérialisation API
type ConversationClassicDTO struct {
	ID           string             `json:"id"`
	Type         string             `json:"type"`
	IsActive     bool               `json:"is_active"`
	Participants []UserDTO          `json:"participants"`
	LastMessage  *MessageClassicDTO `json:"last_message,omitempty"`
	UnreadCount  int64              `json:"unread_count"`
	CreatedAt    string             `json:"created_at"`
	UpdatedAt    string             `json:"updated_at"`
}

// UserDTO structure basique pour éviter récursion
type UserDTO struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
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

// ========== CRUD Operations ==========

// GetUserConversations récupère les conversations d'un utilisateur avec pagination
func (s *ConversationService) GetUserConversations(userID string, page, limit int) (*ConversationsResponse, error) {
	// Validation paramètres
	if err := s.validatePaginationParams(page, limit); err != nil {
		return nil, err
	}

	// Vérifier que l'utilisateur existe
	if err := s.validateUserExists(userID); err != nil {
		return nil, err
	}

	// Récupérer conversations avec pagination
	offset := (page - 1) * limit
	var conversations []models.ConversationClassic
	var total int64

	// Sous-requête pour récupérer les conversations de l'utilisateur
	err := s.db.Model(&models.ConversationClassic{}).
		Joins("JOIN conversation_classic_participants ccp ON ccp.conversation_classic_id = conversation_classics.id").
		Where("ccp.user_id = ? AND conversation_classics.is_active = ?", userID, true).
		Count(&total).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count user conversations: %w", err)
	}

	err = s.db.
		Preload("Participants").
		Preload("LastMessage").
		Preload("LastMessage.Sender").
		Joins("JOIN conversation_classic_participants ccp ON ccp.conversation_classic_id = conversation_classics.id").
		Where("ccp.user_id = ? AND conversation_classics.is_active = ?", userID, true).
		Order("conversation_classics.updated_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&conversations).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user conversations: %w", err)
	}

	// Convertir en DTOs avec informations supplémentaires
	conversationDTOs := make([]ConversationClassicDTO, len(conversations))
	for i, conv := range conversations {
		// Calculer nombre de messages non lus pour cette conversation
		unreadCount, _ := s.getUnreadCountForConversation(conv.ID, userID)

		conversationDTOs[i] = s.toConversationDTO(&conv, userID, unreadCount)
	}

	// Calculer nombre total de messages non lus
	totalUnread, err := s.GetTotalUnreadMessagesCount(userID)
	if err != nil {
		totalUnread = 0 // Fallback sans échouer la requête
	}

	return &ConversationsResponse{
		Conversations: conversationDTOs,
		Pagination: PaginationResponse{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasMore: int64(page*limit) < total,
		},
		UnreadTotal: totalUnread,
	}, nil
}

// GetConversationByID récupère une conversation spécifique
func (s *ConversationService) GetConversationByID(conversationID, userID string) (*ConversationClassicDTO, error) {
	// Validation paramètres
	if !utils.ValidateUUID(conversationID) {
		return nil, errors.New("invalid conversation ID")
	}

	// Vérifier accès utilisateur
	hasAccess, err := s.CanUserAccessConversation(conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, errors.New("unauthorized access to conversation")
	}

	// Récupérer conversation
	var conversation models.ConversationClassic
	err = s.db.
		Preload("Participants").
		Preload("LastMessage").
		Preload("LastMessage.Sender").
		Where("id = ?", conversationID).
		First(&conversation).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("conversation not found")
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Calculer messages non lus
	unreadCount, _ := s.getUnreadCountForConversation(conversation.ID, userID)

	conversationDTO := s.toConversationDTO(&conversation, userID, unreadCount)
	return &conversationDTO, nil
}

// CreateOrGetDirectConversation crée ou récupère une conversation directe
func (s *ConversationService) CreateOrGetDirectConversation(userID, otherUserID string) (*ConversationClassicDTO, error) {
	// Validations
	if userID == otherUserID {
		return nil, errors.New("cannot create conversation with yourself")
	}

	if !utils.ValidateUUID(userID) || !utils.ValidateUUID(otherUserID) {
		return nil, errors.New("invalid user IDs")
	}

	// Vérifier que les utilisateurs existent
	if err := s.validateUserExists(userID); err != nil {
		return nil, err
	}
	if err := s.validateUserExists(otherUserID); err != nil {
		return nil, fmt.Errorf("target user not found")
	}

	// Vérifier permissions de messagerie
	canMessage, err := s.canUserMessageOther(userID, otherUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check messaging permissions: %w", err)
	}
	if !canMessage {
		return nil, errors.New("cannot send message to this user")
	}

	// Chercher conversation existante
	var existingConv models.ConversationClassic
	err = s.db.
		Preload("Participants").
		Joins(`JOIN conversation_classic_participants ccp1 ON ccp1.conversation_classic_id = conversation_classics.id AND ccp1.user_id = ?`, userID).
		Joins(`JOIN conversation_classic_participants ccp2 ON ccp2.conversation_classic_id = conversation_classics.id AND ccp2.user_id = ?`, otherUserID).
		Where("conversation_classics.type = ?", models.ConversationTypeDirect).
		First(&existingConv).Error

	if err == nil {
		// Conversation trouvée
		unreadCount, _ := s.getUnreadCountForConversation(existingConv.ID, userID)
		conversationDTO := s.toConversationDTO(&existingConv, userID, unreadCount)
		return &conversationDTO, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to search existing conversation: %w", err)
	}

	// Créer nouvelle conversation en transaction
	var newConv models.ConversationClassic
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Créer la conversation
		newConv = models.ConversationClassic{
			ID:       utils.GenerateUUID(),
			Type:     models.ConversationTypeDirect,
			IsActive: true,
		}

		if err := tx.Create(&newConv).Error; err != nil {
			return fmt.Errorf("failed to create conversation: %w", err)
		}

		// Ajouter les participants
		participants := []models.ConversationClassicParticipant{
			{
				ConversationClassicID: newConv.ID,
				UserID:                userID,
			},
			{
				ConversationClassicID: newConv.ID,
				UserID:                otherUserID,
			},
		}

		for _, participant := range participants {
			if err := tx.Create(&participant).Error; err != nil {
				return fmt.Errorf("failed to add participant: %w", err)
			}
		}

		// Preload les relations pour le retour
		return tx.Preload("Participants").First(&newConv, "id = ?", newConv.ID).Error
	})

	if err != nil {
		return nil, err
	}

	conversationDTO := s.toConversationDTO(&newConv, userID, 0)
	return &conversationDTO, nil
}

// ========== Read Status Management ==========

// MarkConversationAsRead marque une conversation comme lue
func (s *ConversationService) MarkConversationAsRead(conversationID, userID string) error {
	// Validation
	if !utils.ValidateUUID(conversationID) {
		return errors.New("invalid conversation ID")
	}

	// Vérifier accès
	hasAccess, err := s.CanUserAccessConversation(conversationID, userID)
	if err != nil {
		return err
	}
	if !hasAccess {
		return errors.New("unauthorized access to conversation")
	}

	// Mettre à jour ou créer le statut de lecture
	readStatus := models.ConversationClassicReadStatus{
		ConversationID: conversationID,
		UserID:         userID,
		LastReadAt:     utils.TimeNow(),
	}

	// Upsert (create or update)
	err = s.db.
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Assign(readStatus).
		FirstOrCreate(&readStatus).Error

	if err != nil {
		return fmt.Errorf("failed to mark conversation as read: %w", err)
	}

	return nil
}

// GetUnreadConversationsCount compte les conversations avec messages non lus
func (s *ConversationService) GetUnreadConversationsCount(userID string) (int64, error) {
	// Vérifier utilisateur
	if err := s.validateUserExists(userID); err != nil {
		return 0, err
	}

	// Requête pour compter les conversations avec messages non lus
	var count int64

	// Sous-requête pour dernière lecture par conversation
	subQuery := s.db.Model(&models.ConversationClassicReadStatus{}).
		Select("COALESCE(last_read_at, '1970-01-01')").
		Where("conversation_id = conversation_classics.id AND user_id = ?", userID)

	err := s.db.Model(&models.ConversationClassic{}).
		Joins("JOIN conversation_classic_participants ccp ON ccp.conversation_classic_id = conversation_classics.id").
		Joins("JOIN message_classics mc ON mc.conversation_id = conversation_classics.id").
		Where("ccp.user_id = ? AND conversation_classics.is_active = ?", userID, true).
		Where("mc.sender_id != ? AND mc.created_at > (?)", userID, subQuery).
		Distinct("conversation_classics.id").
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count unread conversations: %w", err)
	}

	return count, nil
}

// GetTotalUnreadMessagesCount compte le total des messages non lus
func (s *ConversationService) GetTotalUnreadMessagesCount(userID string) (int64, error) {
	var count int64

	// Sous-requête pour dernière lecture par conversation
	subQuery := s.db.Model(&models.ConversationClassicReadStatus{}).
		Select("conversation_id, COALESCE(last_read_at, '1970-01-01') as last_read").
		Where("user_id = ?", userID)

	err := s.db.Model(&models.MessageClassic{}).
		Joins("JOIN conversation_classic_participants ccp ON ccp.conversation_classic_id = message_classics.conversation_id").
		Joins("LEFT JOIN (?) rs ON rs.conversation_id = message_classics.conversation_id", subQuery).
		Where("ccp.user_id = ? AND message_classics.sender_id != ?", userID, userID).
		Where("message_classics.created_at > COALESCE(rs.last_read, '1970-01-01')").
		Count(&count).Error

	return count, err
}

// ========== Management Operations ==========

// ArchiveConversation archive une conversation
func (s *ConversationService) ArchiveConversation(conversationID, userID string) error {
	// Validation
	if !utils.ValidateUUID(conversationID) {
		return errors.New("invalid conversation ID")
	}

	// Vérifier accès
	hasAccess, err := s.CanUserAccessConversation(conversationID, userID)
	if err != nil {
		return err
	}
	if !hasAccess {
		return errors.New("unauthorized access to conversation")
	}

	// Marquer comme inactive
	err = s.db.Model(&models.ConversationClassic{}).
		Where("id = ?", conversationID).
		Update("is_active", false).Error

	if err != nil {
		return fmt.Errorf("failed to archive conversation: %w", err)
	}

	return nil
}

// DeleteConversation supprime une conversation (soft delete)
func (s *ConversationService) DeleteConversation(conversationID, userID string) error {
	// Validation
	if !utils.ValidateUUID(conversationID) {
		return errors.New("invalid conversation ID")
	}

	// Vérifier accès
	hasAccess, err := s.CanUserAccessConversation(conversationID, userID)
	if err != nil {
		return err
	}
	if !hasAccess {
		return errors.New("unauthorized access to conversation")
	}

	// Soft delete de la conversation
	err = s.db.Where("id = ?", conversationID).Delete(&models.ConversationClassic{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	return nil
}

// ========== Search & Validation ==========

// SearchConversations recherche dans les conversations d'un utilisateur
func (s *ConversationService) SearchConversations(userID, query string, page, limit int) ([]ConversationClassicDTO, error) {
	// Validation
	if err := s.validateUserExists(userID); err != nil {
		return nil, err
	}

	if len(query) < 2 {
		return []ConversationClassicDTO{}, nil
	}

	if err := s.validatePaginationParams(page, limit); err != nil {
		return nil, err
	}

	// Recherche dans les participants
	offset := (page - 1) * limit
	var conversations []models.ConversationClassic

	err := s.db.
		Preload("Participants").
		Preload("LastMessage").
		Preload("LastMessage.Sender").
		Joins("JOIN conversation_classic_participants ccp ON ccp.conversation_classic_id = conversation_classics.id").
		Joins("JOIN conversation_classic_participants ccp2 ON ccp2.conversation_classic_id = conversation_classics.id").
		Joins("JOIN users u ON u.id = ccp2.user_id").
		Where("ccp.user_id = ? AND conversation_classics.is_active = ?", userID, true).
		Where("ccp2.user_id != ? AND u.username ILIKE ?", userID, "%"+query+"%").
		Distinct("conversation_classics.id").
		Order("conversation_classics.updated_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&conversations).Error

	if err != nil {
		return nil, fmt.Errorf("failed to search conversations: %w", err)
	}

	var conversationDTOs []ConversationClassicDTO
	for _, conv := range conversations {
		unreadCount, _ := s.getUnreadCountForConversation(conv.ID, userID)
		conversationDTOs = append(conversationDTOs, s.toConversationDTO(&conv, userID, unreadCount))
	}

	return conversationDTOs, nil
}

// CanUserAccessConversation vérifie si un utilisateur peut accéder à une conversation
func (s *ConversationService) CanUserAccessConversation(conversationID, userID string) (bool, error) {
	// Vérifier que l'utilisateur est participant
	var count int64
	err := s.db.Model(&models.ConversationClassicParticipant{}).
		Where("conversation_classic_id = ? AND user_id = ?", conversationID, userID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check conversation access: %w", err)
	}

	return count > 0, nil
}

// ========== Helper Methods ==========

// validateUserExists vérifie qu'un utilisateur existe
func (s *ConversationService) validateUserExists(userID string) error {
	var count int64
	err := s.db.Model(&models.User{}).Where("id = ?", userID).Count(&count).Error
	if err != nil {
		return fmt.Errorf("failed to validate user: %w", err)
	}
	if count == 0 {
		return errors.New("user not found")
	}
	return nil
}

// validatePaginationParams valide les paramètres de pagination
func (s *ConversationService) validatePaginationParams(page, limit int) error {
	if page < 1 {
		return errors.New("page must be >= 1")
	}
	if limit < 1 || limit > 100 {
		return errors.New("limit must be between 1 and 100")
	}
	return nil
}

// canUserMessageOther vérifie si un utilisateur peut envoyer un message à un autre
func (s *ConversationService) canUserMessageOther(userID, otherUserID string) (bool, error) {
	// Vérifier si l'utilisateur est bloqué
	var blockCount int64
	err := s.db.Model(&models.Block{}).
		Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)",
			userID, otherUserID, otherUserID, userID).
		Count(&blockCount).Error

	if err != nil {
		return false, fmt.Errorf("failed to check block status: %w", err)
	}

	// Si il y a un blocage, pas de messagerie possible
	if blockCount > 0 {
		return false, nil
	}

	// Vérifier si l'autre utilisateur est actif
	var user models.User
	err = s.db.Select("is_active, is_banned").Where("id = ?", otherUserID).First(&user).Error
	if err != nil {
		return false, fmt.Errorf("failed to check user status: %w", err)
	}

	return user.IsActive && !user.IsBanned, nil
}

// getUnreadCountForConversation compte les messages non lus dans une conversation
func (s *ConversationService) getUnreadCountForConversation(conversationID, userID string) (int64, error) {
	var count int64

	// Sous-requête pour dernière lecture
	subQuery := s.db.Model(&models.ConversationClassicReadStatus{}).
		Select("COALESCE(last_read_at, '1970-01-01')").
		Where("conversation_id = ? AND user_id = ?", conversationID, userID)

	// Compter messages non lus
	err := s.db.Model(&models.MessageClassic{}).
		Where("conversation_id = ? AND sender_id != ?", conversationID, userID).
		Where("created_at > (?)", subQuery).
		Count(&count).Error

	return count, err
}

// toConversationDTO convertit un model en DTO
func (s *ConversationService) toConversationDTO(conv *models.ConversationClassic, currentUserID string, unreadCount int64) ConversationClassicDTO {
	// Convertir participants
	participants := make([]UserDTO, len(conv.Participants))
	for i, participant := range conv.Participants {
		participants[i] = UserDTO{
			ID:          participant.UserID,
			Username:    participant.User.Username,
			DisplayName: participant.User.FirstName + " " + participant.User.LastName,
			AvatarURL:   participant.User.ProfilePicture,
		}
	}

	// Convertir dernier message si présent
	var lastMessage *MessageClassicDTO
	if conv.LastMessage != nil {
		sender := UserDTO{
			ID:          conv.LastMessage.SenderID,
			Username:    conv.LastMessage.Sender.Username,
			DisplayName: conv.LastMessage.Sender.FirstName + " " + conv.LastMessage.Sender.LastName,
			AvatarURL:   conv.LastMessage.Sender.ProfilePicture,
		}

		var readAt *string
		if conv.LastMessage.ReadAt != nil {
			readAtStr := conv.LastMessage.ReadAt.Format("2006-01-02T15:04:05Z07:00")
			readAt = &readAtStr
		}

		lastMessage = &MessageClassicDTO{
			ID:             conv.LastMessage.ID,
			ConversationID: conv.LastMessage.ConversationID,
			Sender:         sender,
			Content:        conv.LastMessage.Content,
			MediaURL:       conv.LastMessage.MediaURL,
			MediaType:      conv.LastMessage.MediaType,
			ThumbnailURL:   conv.LastMessage.ThumbnailURL,
			MessageType:    string(conv.LastMessage.MessageType),
			Status:         string(conv.LastMessage.Status),
			ReadAt:         readAt,
			CreatedAt:      conv.LastMessage.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return ConversationClassicDTO{
		ID:           conv.ID,
		Type:         string(conv.Type),
		IsActive:     conv.IsActive,
		Participants: participants,
		LastMessage:  lastMessage,
		UnreadCount:  unreadCount,
		CreatedAt:    conv.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    conv.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
