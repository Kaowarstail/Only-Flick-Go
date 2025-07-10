package constants

// Message Types
const (
	MessageTypeText  = "text"
	MessageTypeImage = "image"
	MessageTypeVideo = "video"
	MessageTypeAudio = "audio"
)

// Message Status
const (
	MessageStatusSending   = "sending"
	MessageStatusSent      = "sent"
	MessageStatusDelivered = "delivered"
	MessageStatusRead      = "read"
	MessageStatusFailed    = "failed"
)

// Business Rules
const (
	MaxMessageContentLength = 5000
	MaxConversationsPerPage = 50
	MaxMessagesPerPage      = 100
	MaxMediaFileSize        = 50 * 1024 * 1024 // 50MB
	DefaultPageSize         = 20
	DefaultMessagePageSize  = 50
)

// Media Types autorisés
var AllowedMediaTypes = []string{
	"image/jpeg",
	"image/png",
	"image/webp",
	"image/gif",
	"video/mp4",
	"video/webm",
	"video/quicktime",
	"audio/mp3",
	"audio/wav",
	"audio/ogg",
	"audio/mpeg",
}

// Helper functions

// IsValidMessageType vérifie si le type de message est valide
func IsValidMessageType(messageType string) bool {
	validTypes := []string{
		MessageTypeText,
		MessageTypeImage,
		MessageTypeVideo,
		MessageTypeAudio,
	}

	for _, validType := range validTypes {
		if messageType == validType {
			return true
		}
	}
	return false
}

// IsValidMessageStatus vérifie si le statut de message est valide
func IsValidMessageStatus(status string) bool {
	validStatuses := []string{
		MessageStatusSending,
		MessageStatusSent,
		MessageStatusDelivered,
		MessageStatusRead,
		MessageStatusFailed,
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// IsMediaType vérifie si le type de message est un média
func IsMediaType(messageType string) bool {
	return messageType == MessageTypeImage ||
		messageType == MessageTypeVideo ||
		messageType == MessageTypeAudio
}

// GetMessageTypeFromMimeType détermine le type de message depuis le MIME type
func GetMessageTypeFromMimeType(mimeType string) string {
	switch {
	case isImageMimeType(mimeType):
		return MessageTypeImage
	case isVideoMimeType(mimeType):
		return MessageTypeVideo
	case isAudioMimeType(mimeType):
		return MessageTypeAudio
	default:
		return MessageTypeText
	}
}

// isImageMimeType vérifie si le MIME type est une image
func isImageMimeType(mimeType string) bool {
	imageMimes := []string{"image/jpeg", "image/png", "image/webp", "image/gif"}
	for _, mime := range imageMimes {
		if mime == mimeType {
			return true
		}
	}
	return false
}

// isVideoMimeType vérifie si le MIME type est une vidéo
func isVideoMimeType(mimeType string) bool {
	videoMimes := []string{"video/mp4", "video/webm", "video/quicktime"}
	for _, mime := range videoMimes {
		if mime == mimeType {
			return true
		}
	}
	return false
}

// isAudioMimeType vérifie si le MIME type est un audio
func isAudioMimeType(mimeType string) bool {
	audioMimes := []string{"audio/mp3", "audio/wav", "audio/ogg", "audio/mpeg"}
	for _, mime := range audioMimes {
		if mime == mimeType {
			return true
		}
	}
	return false
}

// IsAllowedMediaType vérifie si le MIME type est autorisé
func IsAllowedMediaType(mimeType string) bool {
	for _, allowedType := range AllowedMediaTypes {
		if allowedType == mimeType {
			return true
		}
	}
	return false
}

// GetDefaultPageSize retourne la taille de page par défaut
func GetDefaultPageSize() int {
	return DefaultPageSize
}

// GetDefaultMessagePageSize retourne la taille de page par défaut pour les messages
func GetDefaultMessagePageSize() int {
	return DefaultMessagePageSize
}

// ValidatePageSize valide et normalise la taille de page
func ValidatePageSize(size int, maxSize int) int {
	if size <= 0 {
		return DefaultPageSize
	}
	if size > maxSize {
		return maxSize
	}
	return size
}
