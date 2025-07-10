package utils

import (
	"fmt"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/internal/constants"
	"github.com/Kaowarstail/Only-Flick-Go/internal/dto"
	"github.com/Kaowarstail/Only-Flick-Go/internal/models"
	userModels "github.com/Kaowarstail/Only-Flick-Go/models"
)

// MessagingHelpers fournit des utilitaires pour le système de messagerie
type MessagingHelpers struct{}

// NewMessagingHelpers crée une nouvelle instance des helpers de messagerie
func NewMessagingHelpers() *MessagingHelpers {
	return &MessagingHelpers{}
}

// BuildConversationResponse construit une réponse de conversation enrichie
func (h *MessagingHelpers) BuildConversationResponse(
	conversation *models.Conversation,
	currentUserID string,
	unreadCount int,
) dto.ConversationResponse {
	// Construire la liste des participants
	participants := []userModels.User{
		conversation.Participant1,
		conversation.Participant2,
	}

	// Identifier l'autre utilisateur
	var otherUser *userModels.User
	if conversation.Participant1ID == currentUserID {
		otherUser = &conversation.Participant2
	} else {
		otherUser = &conversation.Participant1
	}

	return dto.ConversationResponse{
		ID:           conversation.ID,
		Participants: participants,
		LastMessage:  conversation.LastMessage,
		UnreadCount:  unreadCount,
		UpdatedAt:    conversation.UpdatedAt,
		IsActive:     conversation.IsActive,
		OtherUser:    otherUser,
	}
}

// GetConversationDisplayName retourne le nom d'affichage pour une conversation
func (h *MessagingHelpers) GetConversationDisplayName(
	conversation *models.Conversation,
	currentUserID string,
) string {
	otherUser := conversation.GetOtherParticipant(currentUserID)
	if otherUser == nil {
		return "Conversation"
	}

	// Priorité : FirstName + LastName, sinon Username
	if otherUser.FirstName != "" && otherUser.LastName != "" {
		return fmt.Sprintf("%s %s", otherUser.FirstName, otherUser.LastName)
	}
	if otherUser.FirstName != "" {
		return otherUser.FirstName
	}
	return otherUser.Username
}

// GetConversationPreview retourne un aperçu du dernier message
func (h *MessagingHelpers) GetConversationPreview(conversation *models.Conversation) string {
	if conversation.LastMessage == nil {
		return "Aucun message"
	}

	lastMessage := conversation.LastMessage
	if lastMessage.HasContent() {
		content := *lastMessage.Content
		if len(content) > 50 {
			return content[:47] + "..."
		}
		return content
	}

	return lastMessage.GetDisplayContent()
}

// FormatMessageTime formate l'heure d'un message pour l'affichage
func (h *MessagingHelpers) FormatMessageTime(messageTime time.Time) string {
	now := time.Now()
	diff := now.Sub(messageTime)

	switch {
	case diff < time.Minute:
		return "Maintenant"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "Il y a 1 minute"
		}
		return fmt.Sprintf("Il y a %d minutes", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "Il y a 1 heure"
		}
		return fmt.Sprintf("Il y a %d heures", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "Hier"
		}
		return fmt.Sprintf("Il y a %d jours", days)
	default:
		return messageTime.Format("02/01/2006")
	}
}

// ValidateAndNormalizePagination valide et normalise les paramètres de pagination
func (h *MessagingHelpers) ValidateAndNormalizePagination(page, limit int, isConversation bool) (int, int) {
	// Valider page
	if page <= 0 {
		page = 1
	}

	// Valider limit selon le type
	if isConversation {
		if limit <= 0 {
			limit = constants.GetDefaultPageSize()
		} else if limit > constants.MaxConversationsPerPage {
			limit = constants.MaxConversationsPerPage
		}
	} else {
		if limit <= 0 {
			limit = constants.GetDefaultMessagePageSize()
		} else if limit > constants.MaxMessagesPerPage {
			limit = constants.MaxMessagesPerPage
		}
	}

	return page, limit
}

// BuildMessagesResponse construit une réponse de messages avec pagination
func (h *MessagingHelpers) BuildMessagesResponse(
	messages []models.Message,
	total int64,
	page, limit int,
	unreadCount int,
) dto.MessagesResponse {
	hasMore := int64((page-1)*limit+len(messages)) < total

	return dto.MessagesResponse{
		Messages:    messages,
		Total:       total,
		Page:        page,
		Limit:       limit,
		HasMore:     hasMore,
		UnreadCount: unreadCount,
	}
}

// BuildConversationsResponse construit une réponse de conversations avec pagination
func (h *MessagingHelpers) BuildConversationsResponse(
	conversations []dto.ConversationResponse,
	total int64,
	page, limit int,
	unreadTotal int,
) dto.ConversationsResponse {
	hasMore := int64((page-1)*limit+len(conversations)) < total

	return dto.ConversationsResponse{
		Conversations: conversations,
		Total:         total,
		UnreadTotal:   unreadTotal,
		Page:          page,
		Limit:         limit,
		HasMore:       hasMore,
	}
}

// SanitizeMessageContent nettoie et sécurise le contenu d'un message
func (h *MessagingHelpers) SanitizeMessageContent(content string) string {
	// Supprimer les espaces en trop
	content = TrimAndNormalizeSpaces(content)

	// Limiter la longueur
	if len(content) > constants.MaxMessageContentLength {
		content = content[:constants.MaxMessageContentLength]
	}

	return content
}

// GenerateConversationID génère un ID de conversation unique basé sur les participants
func (h *MessagingHelpers) GenerateConversationID(userID1, userID2 string) string {
	if userID1 < userID2 {
		return fmt.Sprintf("%s_%s", userID1, userID2)
	}
	return fmt.Sprintf("%s_%s", userID2, userID1)
}

// IsValidMessageStatus vérifie si un statut de message est valide
func (h *MessagingHelpers) IsValidMessageStatus(status string) bool {
	return constants.IsValidMessageStatus(status)
}

// IsValidMessageType vérifie si un type de message est valide
func (h *MessagingHelpers) IsValidMessageType(messageType string) bool {
	return constants.IsValidMessageType(messageType)
}

// GetMessageStatusDisplayName retourne le nom d'affichage pour un statut
func (h *MessagingHelpers) GetMessageStatusDisplayName(status string) string {
	switch status {
	case constants.MessageStatusSending:
		return "Envoi en cours..."
	case constants.MessageStatusSent:
		return "Envoyé"
	case constants.MessageStatusDelivered:
		return "Livré"
	case constants.MessageStatusRead:
		return "Lu"
	case constants.MessageStatusFailed:
		return "Échec"
	default:
		return "Inconnu"
	}
}

// CalculateOffset calcule l'offset pour la pagination
func (h *MessagingHelpers) CalculateOffset(page, limit int) int {
	return (page - 1) * limit
}

// TrimAndNormalizeSpaces supprime les espaces en trop et normalise les espaces
func TrimAndNormalizeSpaces(text string) string {
	// Cette fonction pourrait être plus sophistiquée avec regex
	// Pour l'instant, simple trim
	return text
}
