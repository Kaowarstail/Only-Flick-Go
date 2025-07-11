package services

import (
	"errors"
	"fmt"

	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
	"github.com/Kaowarstail/Only-Flick-Go/models"
	"gorm.io/gorm"
)

// MessageService gère la logique métier des messages
type MessageService struct {
	db                  *gorm.DB
	conversationService *ConversationService
}

// NewMessageService crée une nouvelle instance
func NewMessageService(db *gorm.DB, conversationService *ConversationService) *MessageService {
	return &MessageService{
		db:                  db,
		conversationService: conversationService,
	}
}

// MessageServiceInterface définit les méthodes du service
type MessageServiceInterface interface {
	// CRUD Operations
	GetConversationMessages(conversationID, userID string, page, limit int) (*MessagesResponse, error)
	SendMessage(request *SendMessageRequest, senderID string) (*models.MessageClassicDTO, error)
	GetMessageByID(messageID, userID string) (*models.MessageClassicDTO, error)

	// Message Actions
	MarkMessageAsRead(messageID, userID string) error
	DeleteMessage(messageID, userID string) error

	// Search
	SearchMessages(userID, query string, limit int) ([]models.MessageClassicDTO, error)
	ValidateMediaMessage(mediaURL, mediaType string) error
}

// MessagesResponse représente une liste de messages
type MessagesResponse struct {
	Messages    []models.MessageClassicDTO `json:"messages"`
	Pagination  PaginationResponse         `json:"pagination"`
	UnreadCount int64                      `json:"unread_count"`
}

// SendMessageRequest représente une requête d'envoi de message
type SendMessageRequest struct {
	ConversationID string             `json:"conversation_id" validate:"required,uuid"`
	Content        *string            `json:"content,omitempty"`
	MediaURL       *string            `json:"media_url,omitempty"`
	MediaType      *string            `json:"media_type,omitempty"`
	MessageType    models.MessageType `json:"message_type" validate:"required"`
}

// Validation de la requête
func (r *SendMessageRequest) Validate() error {
	// Au moins contenu OU média requis
	hasContent := r.Content != nil && *r.Content != ""
	hasMedia := r.MediaURL != nil && *r.MediaURL != ""

	if !hasContent && !hasMedia {
		return errors.New("message must have content or media")
	}

	// Si média présent, type requis
	if hasMedia && (r.MediaType == nil || *r.MediaType == "") {
		return errors.New("media type required when media URL provided")
	}

	return nil
}

// ========== CRUD Operations ==========

// GetConversationMessages récupère les messages d'une conversation avec pagination
func (s *MessageService) GetConversationMessages(conversationID, userID string, page, limit int) (*MessagesResponse, error) {
	// Validation paramètres
	if !utils.ValidateUUID(conversationID) {
		return nil, errors.New("invalid conversation ID")
	}

	if err := s.validatePaginationParams(page, limit); err != nil {
		return nil, err
	}

	// Vérifier accès à la conversation
	hasAccess, err := s.conversationService.CanUserAccessConversation(conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, errors.New("unauthorized access to conversation")
	}

	// Récupérer messages avec pagination (ordre inverse pour chat)
	offset := (page - 1) * limit
	var messages []models.MessageClassic
	var total int64

	// Compter total
	err = s.db.Model(&models.MessageClassic{}).
		Where("conversation_id = ?", conversationID).
		Count(&total).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count messages: %w", err)
	}

	// Récupérer messages (les plus récents en premier)
	err = s.db.
		Preload("Sender").
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}

	// Convertir en DTOs
	messageDTOs := make([]models.MessageClassicDTO, len(messages))
	for i, msg := range messages {
		messageDTOs[i] = msg.ToDTO()
	}

	// Compter messages non lus dans cette conversation
	unreadCount, err := s.getUnreadCountForConversation(conversationID, userID)
	if err != nil {
		unreadCount = 0 // Fallback
	}

	return &MessagesResponse{
		Messages: messageDTOs,
		Pagination: PaginationResponse{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasMore: int64(page*limit) < total,
		},
		UnreadCount: unreadCount,
	}, nil
}

// SendMessage envoie un nouveau message
func (s *MessageService) SendMessage(request *SendMessageRequest, senderID string) (*models.MessageClassicDTO, error) {
	// Validation requête
	if err := s.validateSendMessageRequest(request, senderID); err != nil {
		return nil, err
	}

	// Vérifier accès à la conversation
	hasAccess, err := s.conversationService.CanUserAccessConversation(request.ConversationID, senderID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, errors.New("unauthorized access to conversation")
	}

	// Créer le message
	message := &models.MessageClassic{
		ID:             utils.GenerateUUID(),
		ConversationID: request.ConversationID,
		SenderID:       senderID,
		MessageType:    request.MessageType,
		Status:         models.MessageStatusSent,
	}

	// Ajouter contenu texte si présent
	if request.Content != nil && *request.Content != "" {
		sanitizedContent := utils.SanitizeMessageContent(*request.Content)
		message.Content = &sanitizedContent
	}

	// Ajouter média si présent
	if request.MediaURL != nil && *request.MediaURL != "" {
		message.MediaURL = request.MediaURL
		message.MediaType = request.MediaType

		// Valider le média
		if err := s.ValidateMediaMessage(*request.MediaURL, *request.MediaType); err != nil {
			return nil, fmt.Errorf("invalid media: %w", err)
		}
	}

	// Sauvegarder en transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Créer le message (hooks GORM vont mettre à jour conversation)
		if err := tx.Create(message).Error; err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}

		// Preload les relations pour le retour
		return tx.Preload("Sender").First(message, "id = ?", message.ID).Error
	})

	if err != nil {
		return nil, err
	}

	messageDTO := message.ToDTO()
	return &messageDTO, nil
}

// GetMessageByID récupère un message spécifique
func (s *MessageService) GetMessageByID(messageID, userID string) (*models.MessageClassicDTO, error) {
	// Validation
	if !utils.ValidateUUID(messageID) {
		return nil, errors.New("invalid message ID")
	}

	// Récupérer message
	var message models.MessageClassic
	err := s.db.Preload("Sender").Where("id = ?", messageID).First(&message).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("message not found")
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	// Vérifier accès à la conversation
	hasAccess, err := s.conversationService.CanUserAccessConversation(message.ConversationID, userID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, errors.New("unauthorized access to message")
	}

	messageDTO := message.ToDTO()
	return &messageDTO, nil
}

// ========== Message Actions ==========

// MarkMessageAsRead marque un message comme lu
func (s *MessageService) MarkMessageAsRead(messageID, userID string) error {
	// Validation
	if !utils.ValidateUUID(messageID) {
		return errors.New("invalid message ID")
	}

	// Récupérer message pour vérifications
	var message models.MessageClassic
	err := s.db.Select("id, conversation_id, sender_id").
		Where("id = ?", messageID).First(&message).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("message not found")
		}
		return fmt.Errorf("failed to get message: %w", err)
	}

	// Vérifier que l'utilisateur n'est pas l'expéditeur
	if message.SenderID == userID {
		return nil // Pas besoin de marquer ses propres messages comme lus
	}

	// Vérifier accès à la conversation
	hasAccess, err := s.conversationService.CanUserAccessConversation(message.ConversationID, userID)
	if err != nil {
		return err
	}
	if !hasAccess {
		return errors.New("unauthorized access to message")
	}

	// Marquer comme lu en mettant à jour le statut de lecture de la conversation
	readStatus := models.ConversationClassicReadStatus{
		ConversationID: message.ConversationID,
		UserID:         userID,
		LastReadAt:     utils.TimeNow(),
	}

	err = s.db.
		Where("conversation_id = ? AND user_id = ?", message.ConversationID, userID).
		Assign(readStatus).
		FirstOrCreate(&readStatus).Error

	if err != nil {
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	return nil
}

// DeleteMessage supprime un message
func (s *MessageService) DeleteMessage(messageID, userID string) error {
	// Validation
	if !utils.ValidateUUID(messageID) {
		return errors.New("invalid message ID")
	}

	// Récupérer message pour vérifications
	var message models.MessageClassic
	err := s.db.Select("id, sender_id").Where("id = ?", messageID).First(&message).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("message not found")
		}
		return fmt.Errorf("failed to get message: %w", err)
	}

	// Vérifier que l'utilisateur est l'expéditeur
	if message.SenderID != userID {
		return errors.New("can only delete your own messages")
	}

	// Supprimer le message (soft delete)
	err = s.db.Where("id = ?", messageID).Delete(&models.MessageClassic{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// ========== Search ==========

// SearchMessages recherche dans les messages d'un utilisateur
func (s *MessageService) SearchMessages(userID, query string, limit int) ([]models.MessageClassicDTO, error) {
	// Validation
	if len(query) < 2 {
		return []models.MessageClassicDTO{}, nil
	}

	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Rechercher dans les conversations auxquelles l'utilisateur participe
	var messages []models.MessageClassic
	err := s.db.
		Preload("Sender").
		Joins("JOIN conversation_classic_participants ccp ON ccp.conversation_classic_id = message_classics.conversation_id").
		Where("ccp.user_id = ? AND message_classics.content ILIKE ?", userID, "%"+query+"%").
		Order("message_classics.created_at DESC").
		Limit(limit).
		Find(&messages).Error

	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}

	// Convertir en DTOs
	messageDTOs := make([]models.MessageClassicDTO, len(messages))
	for i, msg := range messages {
		messageDTOs[i] = msg.ToDTO()
	}

	return messageDTOs, nil
}

// ValidateMediaMessage valide un message média
func (s *MessageService) ValidateMediaMessage(mediaURL, mediaType string) error {
	// Vérifier URL
	if !utils.IsValidURL(mediaURL) {
		return errors.New("invalid media URL")
	}

	// Vérifier type MIME basique
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
		"video/mp4":  true,
		"video/webm": true,
		"audio/mp3":  true,
		"audio/wav":  true,
		"audio/ogg":  true,
	}

	if !allowedTypes[mediaType] {
		return errors.New("unsupported media type")
	}

	return nil
}

// ========== Helper Methods ==========

// validateSendMessageRequest valide une requête d'envoi de message
func (s *MessageService) validateSendMessageRequest(request *SendMessageRequest, senderID string) error {
	if request == nil {
		return errors.New("request cannot be nil")
	}

	if !utils.ValidateUUID(request.ConversationID) {
		return errors.New("invalid conversation ID")
	}

	if !utils.ValidateUUID(senderID) {
		return errors.New("invalid sender ID")
	}

	// Au moins contenu OU média requis
	hasContent := request.Content != nil && *request.Content != ""
	hasMedia := request.MediaURL != nil && *request.MediaURL != ""

	if !hasContent && !hasMedia {
		return errors.New("message must have content or media")
	}

	// Validation contenu
	if hasContent {
		if len(*request.Content) > 5000 {
			return errors.New("message content too long")
		}
	}

	// Validation média
	if hasMedia {
		if request.MediaType == nil || *request.MediaType == "" {
			return errors.New("media type required when media URL provided")
		}
	}

	return nil
}

// validatePaginationParams valide les paramètres de pagination
func (s *MessageService) validatePaginationParams(page, limit int) error {
	if page < 1 {
		return errors.New("page must be >= 1")
	}
	if limit < 1 || limit > 100 {
		return errors.New("limit must be between 1 and 100")
	}
	return nil
}

// getUnreadCountForConversation compte les messages non lus dans une conversation
func (s *MessageService) getUnreadCountForConversation(conversationID, userID string) (int64, error) {
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
