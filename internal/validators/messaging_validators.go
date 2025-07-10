package validators

import (
	"regexp"
	"strings"

	"github.com/Kaowarstail/Only-Flick-Go/internal/constants"
	"github.com/Kaowarstail/Only-Flick-Go/internal/errors"
	"github.com/Kaowarstail/Only-Flick-Go/internal/models"
	"gorm.io/gorm"
)

// ValidateMessageContent valide le contenu d'un message
func ValidateMessageContent(message *models.MessageClassic) error {
	// Au moins contenu OU média requis
	hasContent := message.Content != nil && strings.TrimSpace(*message.Content) != ""
	hasMedia := message.MediaURL != nil && strings.TrimSpace(*message.MediaURL) != ""

	if !hasContent && !hasMedia {
		return errors.ErrInvalidMessageContent
	}

	// Validation longueur contenu
	if hasContent && len(*message.Content) > constants.MaxMessageContentLength {
		return errors.ErrMessageTooLong
	}

	// Si média présent, type requis
	if hasMedia && (message.MediaType == nil || strings.TrimSpace(*message.MediaType) == "") {
		return errors.ErrInvalidMediaType
	}

	// Validation type MIME
	if hasMedia && !constants.IsAllowedMimeType(*message.MediaType) {
		return errors.ErrInvalidMediaType
	}

	return nil
}

// ValidateConversationParticipants valide les participants d'une conversation
func ValidateConversationParticipants(conversation *models.ConversationClassic) error {
	// Au moins 2 participants pour conversation directe
	if conversation.Type == "direct" && len(conversation.Participants) != 2 {
		return errors.ErrInvalidParticipantCount
	}

	// Participants uniques
	userIDs := make(map[string]bool)
	for _, participant := range conversation.Participants {
		if userIDs[participant.ID] {
			return errors.ErrInvalidParticipantCount
		}
		userIDs[participant.ID] = true
	}

	// Vérifier auto-conversation
	if len(conversation.Participants) == 2 {
		if conversation.Participants[0].ID == conversation.Participants[1].ID {
			return errors.ErrSelfConversation
		}
	}

	return nil
}

// SanitizeMessageContent nettoie le contenu d'un message
func SanitizeMessageContent(content string) string {
	// Trim whitespace
	content = strings.TrimSpace(content)

	// Supprimer caractères de contrôle
	reg := regexp.MustCompile(`[\x00-\x1F\x7F]`)
	content = reg.ReplaceAllString(content, "")

	// Limiter lignes consécutives
	reg = regexp.MustCompile(`\n{3,}`)
	content = reg.ReplaceAllString(content, "\n\n")

	return content
}

// ValidateUUID vérifie si une string est un UUID valide
func ValidateUUID(uuid string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(uuid)
}

// ValidateConversationAccess vérifie si un utilisateur peut accéder à une conversation
func ValidateConversationAccess(db *gorm.DB, conversationID, userID string) error {
	var conversation models.ConversationClassic

	// Vérifier que la conversation existe et que l'utilisateur en fait partie
	err := db.Joins("JOIN conversation_classic_participants ccp ON ccp.conversation_classic_id = conversation_classics.id").
		Where("conversation_classics.id = ? AND ccp.user_id = ?", conversationID, userID).
		First(&conversation).Error

	if err != nil {
		if errors.IsNotFoundError(err) {
			return errors.ErrUnauthorizedAccess
		}
		return err
	}

	// Vérifier que la conversation est active
	if !conversation.IsActive {
		return errors.ErrConversationInactive
	}

	return nil
}
