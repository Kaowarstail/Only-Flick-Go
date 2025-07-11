package services

import (
	"errors"
	"fmt"

	"github.com/Kaowarstail/Only-Flick-Go/models"
	"gorm.io/gorm"
)

// MessagingService orchestrateur pour toutes les opérations de messagerie
type MessagingService struct {
	db                  *gorm.DB
	conversationService *ConversationService
	messageService      *MessageService
}

// NewMessagingService crée une nouvelle instance
func NewMessagingService(db *gorm.DB) *MessagingService {
	// Créer les services avec injection de dépendances
	conversationService := NewConversationService(db)
	messageService := NewMessageService(db, conversationService)

	return &MessagingService{
		db:                  db,
		conversationService: conversationService,
		messageService:      messageService,
	}
}

// MessagingServiceInterface définit l'interface complète
type MessagingServiceInterface interface {
	// High-level operations
	StartConversationAndSendMessage(userID, otherUserID string, messageRequest *SendMessageRequest) (*ConversationMessageResponse, error)
	GetUserMessagingDashboard(userID string, page, limit int) (*MessagingDashboardResponse, error)

	// Conversation operations
	GetUserConversations(userID string, page, limit int) (*ConversationsResponse, error)
	GetOrCreateDirectConversation(userID, otherUserID string) (*ConversationClassicDTO, error)

	// Message operations
	SendMessage(request *SendMessageRequest, senderID string) (*models.MessageClassicDTO, error)
	GetConversationMessages(conversationID, userID string, page, limit int) (*MessagesResponse, error)

	// Read status operations
	MarkConversationAsRead(conversationID, userID string) error
	MarkAllConversationsAsRead(userID string) error

	// Statistics
	GetMessagingStats(userID string) (*MessagingStatsResponse, error)

	// Search
	SearchInMessaging(userID, query string, searchType string, limit int) (*MessagingSearchResponse, error)
}

// ConversationMessageResponse représente une conversation avec son premier message
type ConversationMessageResponse struct {
	Conversation ConversationClassicDTO   `json:"conversation"`
	Message      models.MessageClassicDTO `json:"message"`
}

// MessagingDashboardResponse représente le tableau de bord complet
type MessagingDashboardResponse struct {
	Conversations       []ConversationClassicDTO `json:"conversations"`
	Pagination          PaginationResponse       `json:"pagination"`
	TotalUnreadMessages int64                    `json:"total_unread_messages"`
	Stats               MessagingStatsResponse   `json:"stats"`
}

// MessagingStatsResponse représente les statistiques de messagerie
type MessagingStatsResponse struct {
	TotalConversations  int64 `json:"total_conversations"`
	ActiveConversations int64 `json:"active_conversations"`
	UnreadConversations int64 `json:"unread_conversations"`
	TotalUnreadMessages int64 `json:"total_unread_messages"`
}

// MessagingSearchResponse représente les résultats de recherche
type MessagingSearchResponse struct {
	Query         string                     `json:"query"`
	SearchType    string                     `json:"search_type"`
	Conversations []ConversationClassicDTO   `json:"conversations"`
	Messages      []models.MessageClassicDTO `json:"messages"`
}

// ========== High-Level Operations ==========

// StartConversationAndSendMessage crée une conversation et envoie le premier message
func (s *MessagingService) StartConversationAndSendMessage(userID, otherUserID string, messageRequest *SendMessageRequest) (*ConversationMessageResponse, error) {
	// Validation
	if userID == otherUserID {
		return nil, errors.New("cannot start conversation with yourself")
	}

	// Créer ou récupérer conversation
	conversation, err := s.conversationService.CreateOrGetDirectConversation(userID, otherUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Mettre à jour la requête avec l'ID de conversation
	messageRequest.ConversationID = conversation.ID

	// Envoyer le message
	message, err := s.messageService.SendMessage(messageRequest, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return &ConversationMessageResponse{
		Conversation: *conversation,
		Message:      *message,
	}, nil
}

// GetUserMessagingDashboard récupère un tableau de bord complet
func (s *MessagingService) GetUserMessagingDashboard(userID string, page, limit int) (*MessagingDashboardResponse, error) {
	// Récupérer conversations
	conversationsResponse, err := s.conversationService.GetUserConversations(userID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	// Récupérer statistiques
	stats, err := s.GetMessagingStats(userID)
	if err != nil {
		// Log error mais ne pas échouer la requête
		stats = &MessagingStatsResponse{}
	}

	return &MessagingDashboardResponse{
		Conversations:       conversationsResponse.Conversations,
		Pagination:          conversationsResponse.Pagination,
		TotalUnreadMessages: conversationsResponse.UnreadTotal,
		Stats:               *stats,
	}, nil
}

// ========== Delegated Operations ==========

// GetUserConversations délègue au ConversationService
func (s *MessagingService) GetUserConversations(userID string, page, limit int) (*ConversationsResponse, error) {
	return s.conversationService.GetUserConversations(userID, page, limit)
}

// GetOrCreateDirectConversation délègue au ConversationService
func (s *MessagingService) GetOrCreateDirectConversation(userID, otherUserID string) (*ConversationClassicDTO, error) {
	return s.conversationService.CreateOrGetDirectConversation(userID, otherUserID)
}

// SendMessage délègue au MessageService
func (s *MessagingService) SendMessage(request *SendMessageRequest, senderID string) (*models.MessageClassicDTO, error) {
	return s.messageService.SendMessage(request, senderID)
}

// GetConversationMessages délègue au MessageService
func (s *MessagingService) GetConversationMessages(conversationID, userID string, page, limit int) (*MessagesResponse, error) {
	return s.messageService.GetConversationMessages(conversationID, userID, page, limit)
}

// ========== Read Status Operations ==========

// MarkConversationAsRead délègue au ConversationService
func (s *MessagingService) MarkConversationAsRead(conversationID, userID string) error {
	return s.conversationService.MarkConversationAsRead(conversationID, userID)
}

// MarkAllConversationsAsRead marque toutes les conversations comme lues
func (s *MessagingService) MarkAllConversationsAsRead(userID string) error {
	// Récupérer toutes les conversations de l'utilisateur
	rows, err := s.db.Model(&models.ConversationClassic{}).
		Select("conversation_classics.id").
		Joins("JOIN conversation_classic_participants ccp ON ccp.conversation_classic_id = conversation_classics.id").
		Where("ccp.user_id = ? AND conversation_classics.is_active = ?", userID, true).
		Rows()

	if err != nil {
		return fmt.Errorf("failed to get user conversations: %w", err)
	}
	defer rows.Close()

	// Marquer chaque conversation comme lue
	for rows.Next() {
		var conversationID string
		if err := rows.Scan(&conversationID); err != nil {
			continue // Skip erreur sur une conversation
		}

		// Marquer comme lue (ignorer erreurs individuelles)
		_ = s.conversationService.MarkConversationAsRead(conversationID, userID)
	}

	return nil
}

// ========== Statistics ==========

// GetMessagingStats récupère les statistiques de messagerie d'un utilisateur
func (s *MessagingService) GetMessagingStats(userID string) (*MessagingStatsResponse, error) {
	stats := &MessagingStatsResponse{}

	// Nombre total de conversations
	totalConversations, err := s.getTotalConversationsCount(userID)
	if err == nil {
		stats.TotalConversations = totalConversations
	}

	// Conversations actives
	activeConversations, err := s.getActiveConversationsCount(userID)
	if err == nil {
		stats.ActiveConversations = activeConversations
	}

	// Conversations non lues
	unreadConversations, err := s.conversationService.GetUnreadConversationsCount(userID)
	if err == nil {
		stats.UnreadConversations = unreadConversations
	}

	// Total messages non lus
	totalUnreadMessages, err := s.conversationService.GetTotalUnreadMessagesCount(userID)
	if err == nil {
		stats.TotalUnreadMessages = totalUnreadMessages
	}

	return stats, nil
}

// ========== Search ==========

// SearchInMessaging recherche dans conversations et messages
func (s *MessagingService) SearchInMessaging(userID, query string, searchType string, limit int) (*MessagingSearchResponse, error) {
	response := &MessagingSearchResponse{
		Query:         query,
		SearchType:    searchType,
		Conversations: []ConversationClassicDTO{},
		Messages:      []models.MessageClassicDTO{},
	}

	if len(query) < 2 {
		return response, nil
	}

	switch searchType {
	case "conversations", "all":
		conversations, err := s.conversationService.SearchConversations(userID, query, 1, limit)
		if err == nil {
			response.Conversations = conversations
		}
	}

	switch searchType {
	case "messages", "all":
		messages, err := s.messageService.SearchMessages(userID, query, limit)
		if err == nil {
			response.Messages = messages
		}
	}

	return response, nil
}

// ========== Helper Methods ==========

// getTotalConversationsCount compte le total des conversations d'un utilisateur
func (s *MessagingService) getTotalConversationsCount(userID string) (int64, error) {
	var count int64

	err := s.db.Model(&models.ConversationClassic{}).
		Joins("JOIN conversation_classic_participants ccp ON ccp.conversation_classic_id = conversation_classics.id").
		Where("ccp.user_id = ?", userID).
		Count(&count).Error

	return count, err
}

// getActiveConversationsCount compte les conversations actives d'un utilisateur
func (s *MessagingService) getActiveConversationsCount(userID string) (int64, error) {
	var count int64

	err := s.db.Model(&models.ConversationClassic{}).
		Joins("JOIN conversation_classic_participants ccp ON ccp.conversation_classic_id = conversation_classics.id").
		Where("ccp.user_id = ? AND conversation_classics.is_active = ?", userID, true).
		Count(&count).Error

	return count, err
}
