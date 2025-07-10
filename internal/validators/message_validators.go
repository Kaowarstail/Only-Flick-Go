package validators

import (
	"errors"
	"strings"

	"github.com/Kaowarstail/Only-Flick-Go/internal/constants"
)

// MessageValidator fournit des méthodes de validation pour les messages
type MessageValidator struct{}

// NewMessageValidator crée une nouvelle instance de MessageValidator
func NewMessageValidator() *MessageValidator {
	return &MessageValidator{}
}

// ValidateContent valide le contenu d'un message
func (v *MessageValidator) ValidateContent(content *string) error {
	if content == nil {
		return nil // Content peut être nil si média présent
	}

	contentStr := strings.TrimSpace(*content)

	if len(contentStr) == 0 {
		return nil // Content vide ok si média présent
	}

	if len(contentStr) > constants.MaxMessageContentLength {
		return errors.New("message content too long")
	}

	// Vérifier contenu inapproprié (basique)
	if v.containsInappropriateContent(contentStr) {
		return errors.New("message contains inappropriate content")
	}

	return nil
}

// ValidateMediaURL valide l'URL média et son type
func (v *MessageValidator) ValidateMediaURL(mediaURL *string, mediaType *string) error {
	if mediaURL == nil {
		return nil
	}

	urlStr := strings.TrimSpace(*mediaURL)
	if len(urlStr) == 0 {
		return nil
	}

	// Vérifier format URL basique
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return errors.New("invalid media URL format")
	}

	// Media type requis si URL fournie
	if mediaType == nil || strings.TrimSpace(*mediaType) == "" {
		return errors.New("media_type required when media_url provided")
	}

	// Vérifier que le media type est autorisé
	if !constants.IsAllowedMediaType(*mediaType) {
		return errors.New("unsupported media type")
	}

	return nil
}

// ValidateMessageComplete valide un message complet
func (v *MessageValidator) ValidateMessageComplete(content *string, mediaURL *string, mediaType *string) error {
	// Au moins contenu OU média requis
	hasContent := content != nil && strings.TrimSpace(*content) != ""
	hasMedia := mediaURL != nil && strings.TrimSpace(*mediaURL) != ""

	if !hasContent && !hasMedia {
		return errors.New("message must have content or media")
	}

	// Valider contenu si présent
	if err := v.ValidateContent(content); err != nil {
		return err
	}

	// Valider média si présent
	if err := v.ValidateMediaURL(mediaURL, mediaType); err != nil {
		return err
	}

	return nil
}

// ValidateMessageType valide le type de message
func (v *MessageValidator) ValidateMessageType(messageType string) error {
	if !constants.IsValidMessageType(messageType) {
		return errors.New("invalid message type")
	}
	return nil
}

// ValidateConversationID valide l'ID de conversation
func (v *MessageValidator) ValidateConversationID(conversationID uint) error {
	if conversationID == 0 {
		return errors.New("conversation_id is required")
	}
	return nil
}

// ValidateUserID valide l'ID utilisateur
func (v *MessageValidator) ValidateUserID(userID string) error {
	if strings.TrimSpace(userID) == "" {
		return errors.New("user_id is required")
	}
	return nil
}

// ValidatePagination valide les paramètres de pagination
func (v *MessageValidator) ValidatePagination(page, limit int) (int, int, error) {
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = constants.GetDefaultMessagePageSize()
	} else if limit > constants.MaxMessagesPerPage {
		limit = constants.MaxMessagesPerPage
	}

	return page, limit, nil
}

// ValidateSearchQuery valide une requête de recherche
func (v *MessageValidator) ValidateSearchQuery(query string) error {
	query = strings.TrimSpace(query)
	if len(query) == 0 {
		return errors.New("search query cannot be empty")
	}
	if len(query) < 2 {
		return errors.New("search query must be at least 2 characters")
	}
	if len(query) > 100 {
		return errors.New("search query too long (max 100 characters)")
	}
	return nil
}

// containsInappropriateContent vérifie si le contenu contient des mots inappropriés
func (v *MessageValidator) containsInappropriateContent(content string) bool {
	// Liste simple de mots interdits (à étendre selon besoins)
	forbiddenWords := []string{
		// Ajouter mots inappropriés selon contexte
		// Pour l'exemple, liste vide
	}

	contentLower := strings.ToLower(content)
	for _, word := range forbiddenWords {
		if strings.Contains(contentLower, strings.ToLower(word)) {
			return true
		}
	}

	return false
}

// ValidateMediaSize valide la taille d'un fichier média
func (v *MessageValidator) ValidateMediaSize(size int64) error {
	if size > constants.MaxMediaFileSize {
		return errors.New("media file too large (max 50MB)")
	}
	return nil
}

// ValidateConversationParticipants valide que deux utilisateurs peuvent avoir une conversation
func (v *MessageValidator) ValidateConversationParticipants(userID1, userID2 string) error {
	if userID1 == userID2 {
		return errors.New("cannot create conversation with yourself")
	}

	if strings.TrimSpace(userID1) == "" || strings.TrimSpace(userID2) == "" {
		return errors.New("both participant IDs are required")
	}

	return nil
}
