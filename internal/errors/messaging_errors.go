package errors

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Erreurs métier spécifiques
var (
	ErrConversationNotFound    = errors.New("conversation not found")
	ErrMessageNotFound         = errors.New("message not found")
	ErrUnauthorizedAccess      = errors.New("unauthorized access to conversation")
	ErrCannotMessageUser       = errors.New("cannot send message to this user")
	ErrConversationInactive    = errors.New("conversation is inactive")
	ErrDuplicateConversation   = errors.New("conversation already exists")
	ErrInvalidParticipantCount = errors.New("invalid participant count")
	ErrSelfConversation        = errors.New("cannot create conversation with yourself")
	ErrInvalidMessageContent   = errors.New("message must have content or media")
	ErrMessageTooLong          = errors.New("message content too long")
	ErrInvalidMediaType        = errors.New("invalid media type")
)

// MessagingError erreur typée pour le système de messagerie
type MessagingError struct {
	Code       string
	Message    string
	UserID     string
	ResourceID string
}

func (e MessagingError) Error() string {
	return fmt.Sprintf("messaging error [%s]: %s (user:%s, resource:%s)",
		e.Code, e.Message, e.UserID, e.ResourceID)
}

// NewMessagingError crée une erreur de messagerie
func NewMessagingError(code, message, userID, resourceID string) MessagingError {
	return MessagingError{
		Code:       code,
		Message:    message,
		UserID:     userID,
		ResourceID: resourceID,
	}
}

// IsNotFoundError vérifie si l'erreur est "not found"
func IsNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound) ||
		errors.Is(err, ErrConversationNotFound) ||
		errors.Is(err, ErrMessageNotFound)
}

// IsAuthorizationError vérifie si l'erreur est d'autorisation
func IsAuthorizationError(err error) bool {
	return errors.Is(err, ErrUnauthorizedAccess) ||
		errors.Is(err, ErrCannotMessageUser)
}

// IsValidationError vérifie si l'erreur est de validation
func IsValidationError(err error) bool {
	return errors.Is(err, ErrInvalidMessageContent) ||
		errors.Is(err, ErrMessageTooLong) ||
		errors.Is(err, ErrInvalidMediaType) ||
		errors.Is(err, ErrInvalidParticipantCount)
}
