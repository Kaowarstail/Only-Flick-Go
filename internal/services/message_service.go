package services

import (
	"errors"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/internal/dto"
	internalmodels "github.com/Kaowarstail/Only-Flick-Go/internal/models"
	"gorm.io/gorm"
)

type MessageService struct {
	db *gorm.DB
}

func NewMessageService(db *gorm.DB) *MessageService {
	return &MessageService{db: db}
}

func (s *MessageService) GetMessages(conversationID uint, userID string, page, limit int) (*dto.MessagesResponse, error) {
	var messages []internalmodels.EnhancedMessage
	var total int64

	offset := (page - 1) * limit

	// Vérifier participant
	conversationService := NewConversationService(s.db)
	isParticipant, err := conversationService.IsParticipant(conversationID, userID)
	if err != nil || !isParticipant {
		return nil, errors.New("access denied")
	}

	// Count total
	err = s.db.Model(&internalmodels.EnhancedMessage{}).
		Where("conversation_id = ?", conversationID).
		Count(&total).Error
	if err != nil {
		return nil, err
	}

	// Get messages
	err = s.db.Preload("Sender").
		Preload("UnlockedByUser").
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	return &dto.MessagesResponse{
		Messages: messages,
		Total:    total,
		Page:     page,
		Limit:    limit,
		HasMore:  total > int64(page*limit),
	}, nil
}

func (s *MessageService) SendMessage(req dto.SendMessageRequest, senderID string) (*internalmodels.EnhancedMessage, error) {
	// Validation
	if req.Content == nil && req.MediaURL == nil {
		return nil, errors.New("message must have content or media")
	}

	// Vérifier participant
	conversationService := NewConversationService(s.db)
	isParticipant, err := conversationService.IsParticipant(req.ConversationID, senderID)
	if err != nil || !isParticipant {
		return nil, errors.New("access denied")
	}

	message := internalmodels.EnhancedMessage{
		ConversationID: req.ConversationID,
		SenderID:       senderID,
		Content:        req.Content,
		MediaURL:       req.MediaURL,
		MediaType:      req.MediaType,
		MessageType:    req.MessageType,
		Status:         "sent",
		IsPaid:         false,
	}

	// Transaction pour créer message + update conversation
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Créer message
		if err := tx.Create(&message).Error; err != nil {
			return err
		}

		// Mettre à jour conversation
		return tx.Model(&internalmodels.Conversation{}).
			Where("id = ?", req.ConversationID).
			Updates(map[string]interface{}{
				"last_message_id": message.ID,
				"updated_at":      time.Now(),
			}).Error
	})

	if err != nil {
		return nil, err
	}

	// Preload et retourner
	s.db.Preload("Sender").First(&message, message.ID)
	return &message, nil
}

func (s *MessageService) SendPaidMessage(req dto.SendPaidMessageRequest, senderID string) (*internalmodels.EnhancedMessage, error) {
	// Validation
	if req.Content == nil && req.MediaURL == nil {
		return nil, errors.New("paid message must have content or media")
	}

	if req.Price < 0.99 || req.Price > 500.00 {
		return nil, errors.New("price must be between $0.99 and $500.00")
	}

	// Vérifier participant
	conversationService := NewConversationService(s.db)
	isParticipant, err := conversationService.IsParticipant(req.ConversationID, senderID)
	if err != nil || !isParticipant {
		return nil, errors.New("access denied")
	}

	message := internalmodels.EnhancedMessage{
		ConversationID: req.ConversationID,
		SenderID:       senderID,
		Content:        req.Content,
		MediaURL:       req.MediaURL,
		MediaType:      req.MediaType,
		MessageType:    req.MessageType,
		Status:         "sent",
		IsPaid:         true,
		Price:          &req.Price,
		IsUnlocked:     false,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Créer message
		if err := tx.Create(&message).Error; err != nil {
			return err
		}

		// Update conversation
		return tx.Model(&internalmodels.Conversation{}).
			Where("id = ?", req.ConversationID).
			Updates(map[string]interface{}{
				"last_message_id": message.ID,
				"updated_at":      time.Now(),
			}).Error
	})

	if err != nil {
		return nil, err
	}

	s.db.Preload("Sender").First(&message, message.ID)
	return &message, nil
}

func (s *MessageService) UnlockPaidMessage(messageID uint, buyerID string) error {
	var message internalmodels.EnhancedMessage

	// Récupérer message
	err := s.db.Preload("Sender").First(&message, messageID).Error
	if err != nil {
		return err
	}

	// Validations
	if !message.IsPaid {
		return errors.New("message is not paid")
	}
	if message.IsUnlocked {
		return errors.New("message already unlocked")
	}
	if message.SenderID == buyerID {
		return errors.New("cannot unlock own message")
	}

	// Vérifier participant
	conversationService := NewConversationService(s.db)
	isParticipant, err := conversationService.IsParticipant(message.ConversationID, buyerID)
	if err != nil || !isParticipant {
		return errors.New("access denied")
	}

	now := time.Now()

	// Transaction pour débloquer + créer transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Débloquer message
		if err := tx.Model(&message).Updates(map[string]interface{}{
			"is_unlocked": true,
			"unlocked_at": &now,
			"unlocked_by": buyerID,
		}).Error; err != nil {
			return err
		}

		// Créer transaction
		transaction := internalmodels.PaidMessageTransaction{
			MessageID:   messageID,
			BuyerID:     buyerID,
			SellerID:    message.SenderID,
			Amount:      *message.Price,
			Status:      "completed",
			CompletedAt: &now,
		}

		return tx.Create(&transaction).Error
	})
}

func (s *MessageService) MarkAsRead(conversationID uint, userID string) error {
	// Vérifier participant
	conversationService := NewConversationService(s.db)
	isParticipant, err := conversationService.IsParticipant(conversationID, userID)
	if err != nil || !isParticipant {
		return errors.New("access denied")
	}

	now := time.Now()
	return s.db.Model(&internalmodels.EnhancedMessage{}).
		Where("conversation_id = ? AND sender_id != ? AND read_at IS NULL", conversationID, userID).
		Update("read_at", &now).Error
}

func (s *MessageService) GetMessageWithAccess(messageID uint, userID string) (*dto.MessageResponse, error) {
	var message internalmodels.EnhancedMessage
	err := s.db.Preload("Sender").
		Preload("UnlockedByUser").
		First(&message, messageID).Error
	if err != nil {
		return nil, err
	}

	// Vérifier participant
	conversationService := NewConversationService(s.db)
	isParticipant, err := conversationService.IsParticipant(message.ConversationID, userID)
	if err != nil || !isParticipant {
		return nil, errors.New("access denied")
	}

	response := conversationService.convertMessageToResponse(message, userID)
	return response, nil
}
