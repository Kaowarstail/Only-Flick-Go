package services

import (
	"errors"
	"fmt"

	"github.com/Kaowarstail/Only-Flick-Go/internal/constants"
	"github.com/Kaowarstail/Only-Flick-Go/internal/dto"
	"github.com/Kaowarstail/Only-Flick-Go/internal/models"
	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
	"github.com/Kaowarstail/Only-Flick-Go/internal/validators"
	userModels "github.com/Kaowarstail/Only-Flick-Go/models"
	"gorm.io/gorm"
)

// MessagingService gère la logique métier de la messagerie
type MessagingService struct {
	db        *gorm.DB
	validator *validators.MessageValidator
	helpers   *utils.MessagingHelpers
}

// NewMessagingService crée une nouvelle instance du service de messagerie
func NewMessagingService(db *gorm.DB) *MessagingService {
	return &MessagingService{
		db:        db,
		validator: validators.NewMessageValidator(),
		helpers:   utils.NewMessagingHelpers(),
	}
}

// CreateOrGetConversation crée une conversation ou récupère celle existante
func (s *MessagingService) CreateOrGetConversation(req dto.CreateConversationRequest, currentUserID string) (*dto.NewConversationResponse, error) {
	// Validation
	if err := req.Validate(currentUserID); err != nil {
		return nil, err
	}

	// Vérifier que l'autre utilisateur existe
	var otherUser userModels.User
	if err := s.db.Where("id = ?", req.OtherUserID).First(&otherUser).Error; err != nil {
		return nil, errors.New("user not found")
	}

	// Chercher une conversation existante
	var existingConv models.Conversation
	participant1 := currentUserID
	participant2 := req.OtherUserID

	// S'assurer que participant1 < participant2 pour la recherche
	if participant1 > participant2 {
		participant1, participant2 = participant2, participant1
	}

	err := s.db.Where("participant_1_id = ? AND participant_2_id = ? AND is_active = ?",
		participant1, participant2, true).
		Preload("Participant1").
		Preload("Participant2").
		Preload("LastMessage").
		First(&existingConv).Error

	if err == nil {
		// Conversation existante trouvée
		response := s.helpers.BuildConversationResponse(&existingConv, currentUserID, 0)
		return &dto.NewConversationResponse{
			Conversation: response,
			IsNew:        false,
		}, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Créer nouvelle conversation
	newConv := models.Conversation{
		Participant1ID: participant1,
		Participant2ID: participant2,
		IsActive:       true,
	}

	if err := s.db.Create(&newConv).Error; err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Recharger avec les relations
	if err := s.db.Preload("Participant1").Preload("Participant2").First(&newConv, newConv.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load conversation: %w", err)
	}

	response := s.helpers.BuildConversationResponse(&newConv, currentUserID, 0)
	return &dto.NewConversationResponse{
		Conversation: response,
		IsNew:        true,
	}, nil
}

// SendMessage envoie un nouveau message
func (s *MessagingService) SendMessage(req dto.SendMessageRequest, senderID string) (*models.Message, error) {
	// Validation
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Vérifier que la conversation existe et que l'utilisateur en fait partie
	var conversation models.Conversation
	if err := s.db.First(&conversation, req.ConversationID).Error; err != nil {
		return nil, errors.New("conversation not found")
	}

	if !conversation.IsParticipant(senderID) {
		return nil, errors.New("you are not a participant in this conversation")
	}

	// Créer le message
	message := models.Message{
		ConversationID: req.ConversationID,
		SenderID:       senderID,
		MessageType:    req.MessageType,
		Status:         constants.MessageStatusSent,
	}

	// Assigner le contenu selon le type
	if req.Content != nil {
		content := s.helpers.SanitizeMessageContent(*req.Content)
		message.Content = &content
	}
	if req.MediaURL != nil {
		message.MediaURL = req.MediaURL
	}
	if req.MediaType != nil {
		message.MediaType = req.MediaType
	}

	// Validation finale
	if err := message.Validate(); err != nil {
		return nil, err
	}

	// Sauvegarder le message
	if err := s.db.Create(&message).Error; err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Mettre à jour la conversation
	if err := s.db.Model(&conversation).Updates(map[string]interface{}{
		"last_message_id": message.ID,
		"updated_at":      message.CreatedAt,
	}).Error; err != nil {
		// Log l'erreur mais ne pas faire échouer la création du message
		fmt.Printf("Warning: Failed to update conversation: %v\n", err)
	}

	// Recharger le message avec les relations
	if err := s.db.Preload("Sender").First(&message, message.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload message: %w", err)
	}

	return &message, nil
}

// GetConversations récupère les conversations d'un utilisateur
func (s *MessagingService) GetConversations(userID string, req dto.GetConversationsRequest) (*dto.ConversationsResponse, error) {
	page, limit := s.helpers.ValidateAndNormalizePagination(req.Page, req.Limit, true)
	offset := s.helpers.CalculateOffset(page, limit)

	// Query de base
	query := s.db.Model(&models.Conversation{}).
		Where("(participant_1_id = ? OR participant_2_id = ?) AND is_active = ?", userID, userID, true)

	// Compter le total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count conversations: %w", err)
	}

	// Récupérer les conversations
	var conversations []models.Conversation
	if err := query.
		Preload("Participant1").
		Preload("Participant2").
		Preload("LastMessage").
		Order("updated_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&conversations).Error; err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	// Construire les réponses enrichies
	var conversationResponses []dto.ConversationResponse
	var totalUnread int

	for _, conv := range conversations {
		// Compter les messages non lus dans cette conversation
		var unreadCount int64
		otherParticipantID := conv.GetOtherParticipantID(userID)
		s.db.Model(&models.Message{}).
			Where("conversation_id = ? AND sender_id = ? AND read_at IS NULL", conv.ID, otherParticipantID).
			Count(&unreadCount)

		totalUnread += int(unreadCount)

		response := s.helpers.BuildConversationResponse(&conv, userID, int(unreadCount))
		conversationResponses = append(conversationResponses, response)
	}

	result := s.helpers.BuildConversationsResponse(conversationResponses, total, page, limit, totalUnread)
	return &result, nil
}

// GetMessages récupère les messages d'une conversation
func (s *MessagingService) GetMessages(conversationID uint, userID string, page, limit int) (*dto.MessagesResponse, error) {
	// Vérifier que l'utilisateur fait partie de la conversation
	var conversation models.Conversation
	if err := s.db.First(&conversation, conversationID).Error; err != nil {
		return nil, errors.New("conversation not found")
	}

	if !conversation.IsParticipant(userID) {
		return nil, errors.New("you are not a participant in this conversation")
	}

	page, limit = s.helpers.ValidateAndNormalizePagination(page, limit, false)
	offset := s.helpers.CalculateOffset(page, limit)

	// Compter les messages totaux
	var total int64
	if err := s.db.Model(&models.Message{}).Where("conversation_id = ?", conversationID).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count messages: %w", err)
	}

	// Récupérer les messages
	var messages []models.Message
	if err := s.db.Where("conversation_id = ?", conversationID).
		Preload("Sender").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	// Compter les messages non lus
	otherParticipantID := conversation.GetOtherParticipantID(userID)
	var unreadCount int64
	s.db.Model(&models.Message{}).
		Where("conversation_id = ? AND sender_id = ? AND read_at IS NULL", conversationID, otherParticipantID).
		Count(&unreadCount)

	result := s.helpers.BuildMessagesResponse(messages, total, page, limit, int(unreadCount))
	return &result, nil
}

// MarkMessagesAsRead marque des messages comme lus
func (s *MessagingService) MarkMessagesAsRead(conversationID uint, userID string) error {
	// Vérifier que l'utilisateur fait partie de la conversation
	var conversation models.Conversation
	if err := s.db.First(&conversation, conversationID).Error; err != nil {
		return errors.New("conversation not found")
	}

	if !conversation.IsParticipant(userID) {
		return errors.New("you are not a participant in this conversation")
	}

	// Marquer comme lus tous les messages de l'autre participant
	otherParticipantID := conversation.GetOtherParticipantID(userID)

	return s.db.Model(&models.Message{}).
		Where("conversation_id = ? AND sender_id = ? AND read_at IS NULL", conversationID, otherParticipantID).
		Update("read_at", gorm.Expr("NOW()")).Error
}

// GetMessagingStats retourne les statistiques de messagerie pour un utilisateur
func (s *MessagingService) GetMessagingStats(userID string) (*dto.MessageStatsResponse, error) {
	stats := &dto.MessageStatsResponse{}

	// Total conversations
	var totalConversations int64
	s.db.Model(&models.Conversation{}).
		Where("(participant_1_id = ? OR participant_2_id = ?) AND is_active = ?", userID, userID, true).
		Count(&totalConversations)

	// Messages non lus
	var unreadMessages int64
	s.db.Table("messages").
		Joins("JOIN conversations ON messages.conversation_id = conversations.id").
		Where("(conversations.participant_1_id = ? OR conversations.participant_2_id = ?) AND messages.sender_id != ? AND messages.read_at IS NULL",
			userID, userID, userID).
		Count(&unreadMessages)

	stats.UnreadMessages = int(unreadMessages)

	return stats, nil
}