package services

import (
	"errors"

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

// GetMessagesClassic récupère les messages d'une conversation classique
func (s *MessageService) GetMessagesClassic(conversationID string, userID string, page, limit int) (*dto.MessagesClassicResponse, error) {
	var messages []internalmodels.MessageClassic
	var total int64

	offset := (page - 1) * limit

	// Vérifier que l'utilisateur est participant de la conversation
	var conversation internalmodels.ConversationClassic
	err := s.db.Preload("Participants").Where("id = ?", conversationID).First(&conversation).Error
	if err != nil {
		return nil, errors.New("conversation not found")
	}

	isParticipant := false
	for _, participant := range conversation.Participants {
		if participant.ID == userID {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		return nil, errors.New("access denied")
	}

	// Compter le total de messages
	err = s.db.Model(&internalmodels.MessageClassic{}).
		Where("conversation_id = ?", conversationID).
		Count(&total).Error
	if err != nil {
		return nil, err
	}

	// Récupérer les messages avec le sender
	err = s.db.Preload("Sender").
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	hasMore := int64(offset+limit) < total

	return &dto.MessagesClassicResponse{
		Messages: messages,
		Total:    total,
		Page:     page,
		Limit:    limit,
		HasMore:  hasMore,
	}, nil
}

// SendMessageClassic envoie un message classique
func (s *MessageService) SendMessageClassic(req dto.SendMessageClassicRequest, senderID string) (*internalmodels.MessageClassic, error) {
	// Vérifier que l'utilisateur est participant de la conversation
	var conversation internalmodels.ConversationClassic
	err := s.db.Preload("Participants").Where("id = ?", req.ConversationID).First(&conversation).Error
	if err != nil {
		return nil, errors.New("conversation not found")
	}

	isParticipant := false
	for _, participant := range conversation.Participants {
		if participant.ID == senderID {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		return nil, errors.New("access denied")
	}

	// Créer le message
	message := internalmodels.MessageClassic{
		ConversationID: req.ConversationID,
		SenderID:       senderID,
		Content:        req.Content,
		MediaURL:       req.MediaURL,
		MediaType:      req.MediaType,
		MessageType:    internalmodels.MessageType(req.MessageType),
		Status:         internalmodels.MessageStatusSent,
	}

	err = s.db.Create(&message).Error
	if err != nil {
		return nil, err
	}

	// Mettre à jour le dernier message de la conversation
	conversation.LastMessageID = &message.ID
	s.db.Save(&conversation)

	// Recharger le message avec les relations
	err = s.db.Preload("Sender").First(&message, "id = ?", message.ID).Error
	if err != nil {
		return nil, err
	}

	return &message, nil
}

// MarkMessageAsRead marque un message comme lu
func (s *MessageService) MarkMessageAsRead(conversationID string, userID string) error {
	// Créer ou mettre à jour le statut de lecture
	readStatus := internalmodels.ConversationClassicReadStatus{
		ConversationID: conversationID,
		UserID:         userID,
	}

	err := s.db.Save(&readStatus).Error
	return err
}
