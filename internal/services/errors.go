package services

import (
	"errors"
	"fmt"
	"net/http"
)

// ServiceError représente une erreur de service avec code HTTP
type ServiceError struct {
	Code       string
	Message    string
	HTTPStatus int
	Internal   error
}

func (e ServiceError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Erreurs prédéfinies pour les services de messagerie
var (
	ErrUserNotFound          = ServiceError{"USER_NOT_FOUND", "User not found", http.StatusNotFound, nil}
	ErrConversationNotFound  = ServiceError{"CONVERSATION_NOT_FOUND", "Conversation not found", http.StatusNotFound, nil}
	ErrMessageNotFound       = ServiceError{"MESSAGE_NOT_FOUND", "Message not found", http.StatusNotFound, nil}
	ErrUnauthorizedAccess    = ServiceError{"UNAUTHORIZED_ACCESS", "Unauthorized access", http.StatusForbidden, nil}
	ErrInvalidRequest        = ServiceError{"INVALID_REQUEST", "Invalid request", http.StatusBadRequest, nil}
	ErrCannotMessageUser     = ServiceError{"CANNOT_MESSAGE_USER", "Cannot send message to this user", http.StatusForbidden, nil}
	ErrSelfConversation      = ServiceError{"SELF_CONVERSATION", "Cannot create conversation with yourself", http.StatusBadRequest, nil}
	ErrInvalidMessageContent = ServiceError{"INVALID_MESSAGE_CONTENT", "Invalid message content", http.StatusBadRequest, nil}
	ErrInvalidPagination     = ServiceError{"INVALID_PAGINATION", "Invalid pagination parameters", http.StatusBadRequest, nil}
	ErrInvalidUUID           = ServiceError{"INVALID_UUID", "Invalid UUID format", http.StatusBadRequest, nil}
	ErrMediaValidation       = ServiceError{"MEDIA_VALIDATION", "Media validation failed", http.StatusBadRequest, nil}
	ErrDatabaseOperation     = ServiceError{"DATABASE_ERROR", "Database operation failed", http.StatusInternalServerError, nil}
)

// NewServiceError crée une nouvelle erreur de service
func NewServiceError(code, message string, httpStatus int, internal error) ServiceError {
	return ServiceError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Internal:   internal,
	}
}

// WrapError encapsule une erreur interne
func WrapError(baseError ServiceError, internal error) ServiceError {
	return ServiceError{
		Code:       baseError.Code,
		Message:    baseError.Message,
		HTTPStatus: baseError.HTTPStatus,
		Internal:   internal,
	}
}

// IsServiceError vérifie si une erreur est une ServiceError
func IsServiceError(err error) (ServiceError, bool) {
	var serviceErr ServiceError
	if errors.As(err, &serviceErr) {
		return serviceErr, true
	}
	return serviceErr, false
}

// GetHTTPStatus extrait le code de statut HTTP d'une erreur
func GetHTTPStatus(err error) int {
	if serviceErr, ok := IsServiceError(err); ok {
		return serviceErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// FormatErrorResponse formate une erreur pour la réponse API
func FormatErrorResponse(err error) map[string]interface{} {
	if serviceErr, ok := IsServiceError(err); ok {
		response := map[string]interface{}{
			"error":   true,
			"code":    serviceErr.Code,
			"message": serviceErr.Message,
		}

		// En mode développement, inclure l'erreur interne
		if serviceErr.Internal != nil {
			response["internal"] = serviceErr.Internal.Error()
		}

		return response
	}

	// Erreur générique
	return map[string]interface{}{
		"error":   true,
		"code":    "INTERNAL_ERROR",
		"message": "An internal error occurred",
	}
}

// ValidationError représente une erreur de validation
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors représente plusieurs erreurs de validation
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (v ValidationErrors) Error() string {
	if len(v.Errors) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s", v.Errors[0].Message)
}

// NewValidationError crée une nouvelle erreur de validation
func NewValidationError(field, message string) ValidationErrors {
	return ValidationErrors{
		Errors: []ValidationError{
			{Field: field, Message: message},
		},
	}
}

// AddValidationError ajoute une erreur de validation
func (v *ValidationErrors) AddError(field, message string) {
	v.Errors = append(v.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// HasErrors vérifie s'il y a des erreurs de validation
func (v ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}

// ToServiceError convertit en ServiceError
func (v ValidationErrors) ToServiceError() ServiceError {
	return ServiceError{
		Code:       "VALIDATION_ERROR",
		Message:    v.Error(),
		HTTPStatus: http.StatusBadRequest,
		Internal:   nil,
	}
}
