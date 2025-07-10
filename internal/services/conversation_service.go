package services

import (
	"errors"

	"github.com/Kaowarstail/Only-Flick-Go/internal/dto"
	internalmodels "github.com/Kaowarstail/Only-Flick-Go/internal/models"
	"github.com/Kaowarstail/Only-Flick-Go/models"
	"gorm.io/gorm"
)

type ConversationService struct {
	db *gorm.DB
}

func NewConversationService(db *gorm.DB) *ConversationService {
	return &ConversationService{db: db}
}

func (s *ConversationService) GetDB() *gorm.DB {
	return s.db
}

func (s *ConversationService) GetUserConversations(userID string, page, limit int) ([]dto.ConversationResponse, int64, error) {
	var conversations []internalmodels.Conversation
	var total int64

	offset := (page - 1) * limit

	// Count total
	err := s.db.Model(&internalmodels.Conversation{}).
		Where("participant_1_id = ? OR participant_2_id = ?", userID, userID).
		Where("is_active = ?", true).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get conversations avec preload
	err = s.db.Preload("Participant1").
		Preload("Participant2").
		Preload("LastMessage").
		Preload("LastMessage.Sender").
		Where("participant_1_id = ? OR participant_2_id = ?", userID, userID).
		Where("is_active = ?", true).
		Order("updated_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&conversations).Error

	if err != nil {
		return nil, 0, err
	}

	// Convert to response avec unread count
	var responses []dto.ConversationResponse
	for _, conv := range conversations {
		unreadCount := s.getUnreadCount(conv.ID, userID)

		var lastMessage *dto.MessageResponse
		if conv.LastMessage != nil {
			lastMessage = s.convertMessageToResponse(*conv.LastMessage, userID)
		}

		responses = append(responses, dto.ConversationResponse{
			ID:           conv.ID,
			Participants: []models.User{conv.Participant1, conv.Participant2},
			LastMessage:  lastMessage,
			UnreadCount:  unreadCount,
			UpdatedAt:    conv.UpdatedAt,
			IsActive:     conv.IsActive,
		})
	}

	return responses, total, nil
}

func (s *ConversationService) CreateOrGetConversation(user1ID, user2ID string) (*internalmodels.Conversation, error) {
	// S'assurer que user1ID < user2ID
	if user1ID > user2ID {
		user1ID, user2ID = user2ID, user1ID
	}

	var conversation internalmodels.Conversation

	// Chercher conversation existante
	err := s.db.Where("participant_1_id = ? AND participant_2_id = ?", user1ID, user2ID).
		First(&conversation).Error

	if err == nil {
		// Existe déjà, preload et retourner
		s.db.Preload("Participant1").Preload("Participant2").First(&conversation, conversation.ID)
		return &conversation, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Créer nouvelle conversation
	conversation = internalmodels.Conversation{
		Participant1ID: user1ID,
		Participant2ID: user2ID,
		IsActive:       true,
	}

	err = s.db.Create(&conversation).Error
	if err != nil {
		return nil, err
	}

	// Preload relations
	s.db.Preload("Participant1").Preload("Participant2").First(&conversation, conversation.ID)

	return &conversation, nil
}

func (s *ConversationService) IsParticipant(conversationID uint, userID string) (bool, error) {
	var count int64
	err := s.db.Model(&internalmodels.Conversation{}).
		Where("id = ? AND (participant_1_id = ? OR participant_2_id = ?)", conversationID, userID, userID).
		Count(&count).Error

	return count > 0, err
}

func (s *ConversationService) getUnreadCount(conversationID uint, userID string) int {
	var count int64
	s.db.Model(&internalmodels.EnhancedMessage{}).
		Where("conversation_id = ? AND sender_id != ? AND read_at IS NULL", conversationID, userID).
		Count(&count)
	return int(count)
}

func (s *ConversationService) convertMessageToResponse(message internalmodels.EnhancedMessage, viewerID string) *dto.MessageResponse {
	response := &dto.MessageResponse{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		Sender:         message.Sender,
		MessageType:    message.MessageType,
		Status:         message.Status,
		CreatedAt:      message.CreatedAt.Format("2006-01-02T15:04:05Z"),
		IsPaid:         message.IsPaid,
		IsUnlocked:     message.IsUnlocked,
		Price:          message.Price,
	}

	if message.ReadAt != nil {
		readAtStr := message.ReadAt.Format("2006-01-02T15:04:05Z")
		response.ReadAt = &readAtStr
	}

	// Logique d'affichage du contenu
	if !message.IsPaid {
		// Message gratuit - toujours visible
		response.Content = message.Content
		response.MediaURL = message.MediaURL
		response.MediaType = message.MediaType
		response.ThumbnailURL = message.ThumbnailURL
		response.CanUnlock = false
	} else {
		// Message payant
		if message.SenderID == viewerID {
			// L'expéditeur voit toujours son contenu
			response.Content = message.Content
			response.MediaURL = message.MediaURL
			response.MediaType = message.MediaType
			response.ThumbnailURL = message.ThumbnailURL
			response.CanUnlock = false
		} else if message.IsUnlocked {
			// Message débloqué
			response.Content = message.Content
			response.MediaURL = message.MediaURL
			response.MediaType = message.MediaType
			response.ThumbnailURL = message.ThumbnailURL
			response.CanUnlock = false
		} else {
			// Message verrouillé - contenu masqué
			response.Content = nil
			response.MediaURL = nil
			response.MediaType = message.MediaType
			response.ThumbnailURL = message.ThumbnailURL // Thumbnail peut être visible
			response.CanUnlock = true
		}
	}

	return response
}
