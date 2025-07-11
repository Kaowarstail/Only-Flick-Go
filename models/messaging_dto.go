package models

import "errors"

// ========== Request DTOs ==========

// CreateConversationRequest représente une requête de création de conversation
type CreateConversationRequest struct {
	OtherUserID string `json:"other_user_id" validate:"required,uuid"`
}

// Validation
func (r *CreateConversationRequest) Validate() error {
	if r.OtherUserID == "" {
		return errors.New("other_user_id is required")
	}
	return nil
}

// MarkAsReadRequest représente une requête de marquage comme lu
type MarkAsReadRequest struct {
	ConversationID string `json:"conversation_id" validate:"required,uuid"`
}

// Validation
func (r *MarkAsReadRequest) Validate() error {
	if r.ConversationID == "" {
		return errors.New("conversation_id is required")
	}
	return nil
}

// MessagingSearchRequest représente une requête de recherche
type MessagingSearchRequest struct {
	Query      string `json:"query" validate:"required,min=2"`
	SearchType string `json:"search_type" validate:"oneof=conversations messages all"`
	Limit      int    `json:"limit" validate:"min=1,max=100"`
}

// Validation
func (r *MessagingSearchRequest) Validate() error {
	if len(r.Query) < 2 {
		return errors.New("query must be at least 2 characters")
	}

	validTypes := map[string]bool{
		"conversations": true,
		"messages":      true,
		"all":           true,
	}

	if !validTypes[r.SearchType] {
		return errors.New("invalid search type")
	}

	if r.Limit < 1 || r.Limit > 100 {
		r.Limit = 20 // Default
	}

	return nil
}

// SendMessageRequest est défini dans services pour éviter import circulaire
// mais nous gardons une version model pour validation
type SendMessageRequestModel struct {
	ConversationID string      `json:"conversation_id" validate:"required"`
	Content        *string     `json:"content,omitempty"`
	MediaURL       *string     `json:"media_url,omitempty"`
	MediaType      *string     `json:"media_type,omitempty"`
	MessageType    MessageType `json:"message_type" validate:"required"`
}

// Validation
func (r *SendMessageRequestModel) Validate() error {
	if r.ConversationID == "" {
		return errors.New("conversation_id is required")
	}

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

	// Validation longueur contenu
	if hasContent && len(*r.Content) > 5000 {
		return errors.New("message content too long (max 5000 characters)")
	}

	return nil
}

// ========== Response DTOs ==========

// ConversationStatsDTO représente les statistiques de conversations
type ConversationStatsDTO struct {
	TotalConversations  int64 `json:"total_conversations"`
	ActiveConversations int64 `json:"active_conversations"`
	UnreadConversations int64 `json:"unread_conversations"`
	TotalUnreadMessages int64 `json:"total_unread_messages"`
}

// MessagingSearchResultDTO représente un résultat de recherche unifié
type MessagingSearchResultDTO struct {
	Type         string      `json:"type"` // "conversation" or "message"
	Conversation interface{} `json:"conversation,omitempty"`
	Message      interface{} `json:"message,omitempty"`
	Relevance    float64     `json:"relevance"` // Score de pertinence
}

// PaginationMetaDTO contient des métadonnées de pagination enrichies
type PaginationMetaDTO struct {
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	Total       int64 `json:"total"`
	TotalPages  int   `json:"total_pages"`
	HasMore     bool  `json:"has_more"`
	HasPrevious bool  `json:"has_previous"`
}

// NewPaginationMeta crée une nouvelle instance avec calculs automatiques
func NewPaginationMeta(page, limit int, total int64) PaginationMetaDTO {
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}

	return PaginationMetaDTO{
		Page:        page,
		Limit:       limit,
		Total:       total,
		TotalPages:  totalPages,
		HasMore:     page < totalPages,
		HasPrevious: page > 1,
	}
}

// ========== Error Types ==========

// MessagingError représente une erreur spécifique au système de messagerie
type MessagingError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e MessagingError) Error() string {
	return e.Message
}

// Erreurs prédéfinies
var (
	ErrConversationNotFound  = MessagingError{"CONVERSATION_NOT_FOUND", "Conversation not found", ""}
	ErrMessageNotFound       = MessagingError{"MESSAGE_NOT_FOUND", "Message not found", ""}
	ErrUnauthorizedAccess    = MessagingError{"UNAUTHORIZED_ACCESS", "Unauthorized access to resource", ""}
	ErrInvalidMessageContent = MessagingError{"INVALID_MESSAGE_CONTENT", "Invalid message content", "content"}
	ErrUserBlocked           = MessagingError{"USER_BLOCKED", "Cannot message blocked user", ""}
	ErrSelfMessage           = MessagingError{"SELF_MESSAGE", "Cannot message yourself", ""}
	ErrInvalidMediaType      = MessagingError{"INVALID_MEDIA_TYPE", "Unsupported media type", "media_type"}
	ErrMessageTooLong        = MessagingError{"MESSAGE_TOO_LONG", "Message content too long", "content"}
	ErrInvalidPagination     = MessagingError{"INVALID_PAGINATION", "Invalid pagination parameters", ""}
	ErrInvalidUUID           = MessagingError{"INVALID_UUID", "Invalid UUID format", ""}
)

// ========== Constants ==========

// Constantes pour la messagerie
const (
	// Limites
	MaxMessageLength       = 5000
	MaxSearchQueryLength   = 100
	MinSearchQueryLength   = 2
	MaxPaginationLimit     = 100
	DefaultPaginationLimit = 20

	// Types de recherche
	SearchTypeConversations = "conversations"
	SearchTypeMessages      = "messages"
	SearchTypeAll           = "all"

	// Types MIME autorisés
	MimeTypeImageJPEG = "image/jpeg"
	MimeTypeImagePNG  = "image/png"
	MimeTypeImageGIF  = "image/gif"
	MimeTypeImageWEBP = "image/webp"
	MimeTypeVideoMP4  = "video/mp4"
	MimeTypeVideoWEBM = "video/webm"
	MimeTypeAudioMP3  = "audio/mp3"
	MimeTypeAudioWAV  = "audio/wav"
	MimeTypeAudioOGG  = "audio/ogg"
)

// AllowedMimeTypes liste des types MIME autorisés
var AllowedMimeTypes = map[string]bool{
	MimeTypeImageJPEG: true,
	MimeTypeImagePNG:  true,
	MimeTypeImageGIF:  true,
	MimeTypeImageWEBP: true,
	MimeTypeVideoMP4:  true,
	MimeTypeVideoWEBM: true,
	MimeTypeAudioMP3:  true,
	MimeTypeAudioWAV:  true,
	MimeTypeAudioOGG:  true,
}

// IsAllowedMimeType vérifie si un type MIME est autorisé
func IsAllowedMimeType(mimeType string) bool {
	return AllowedMimeTypes[mimeType]
}

// GetMimeTypeCategory retourne la catégorie d'un type MIME
func GetMimeTypeCategory(mimeType string) string {
	switch {
	case mimeType[:5] == "image":
		return "image"
	case mimeType[:5] == "video":
		return "video"
	case mimeType[:5] == "audio":
		return "audio"
	default:
		return "unknown"
	}
}
