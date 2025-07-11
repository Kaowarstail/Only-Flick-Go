package services

import (
	"gorm.io/gorm"
)

// MessagingServiceManager gère l'injection de dépendances pour les services de messagerie
type MessagingServiceManager struct {
	db                  *gorm.DB
	messagingService    MessagingServiceInterface
	conversationService ConversationServiceInterface
	messageService      MessageServiceInterface
}

// NewMessagingServiceManager crée un gestionnaire de services
func NewMessagingServiceManager(db *gorm.DB) *MessagingServiceManager {
	// Créer les services avec injection de dépendances
	conversationService := NewConversationService(db)
	messageService := NewMessageService(db, conversationService)
	messagingService := NewMessagingService(db)

	return &MessagingServiceManager{
		db:                  db,
		messagingService:    messagingService,
		conversationService: conversationService,
		messageService:      messageService,
	}
}

// GetMessagingService retourne le service principal
func (m *MessagingServiceManager) GetMessagingService() MessagingServiceInterface {
	return m.messagingService
}

// GetConversationService retourne le service de conversations
func (m *MessagingServiceManager) GetConversationService() ConversationServiceInterface {
	return m.conversationService
}

// GetMessageService retourne le service de messages
func (m *MessagingServiceManager) GetMessageService() MessageServiceInterface {
	return m.messageService
}
