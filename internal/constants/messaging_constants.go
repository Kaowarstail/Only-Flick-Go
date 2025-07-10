package constants

import "strings"

const (
	// Limites
	MaxMessageContentLength = 5000
	MaxMediaFileSize        = 50 * 1024 * 1024 // 50MB

	// Types MIME autorisés
	AllowedImageTypes = "image/jpeg,image/png,image/webp,image/gif"
	AllowedVideoTypes = "video/mp4,video/webm,video/quicktime"
	AllowedAudioTypes = "audio/mpeg,audio/wav,audio/ogg"

	// Pagination
	DefaultMessagesPerPage      = 50
	DefaultConversationsPerPage = 20
	MaxMessagesPerPage          = 100
	MaxConversationsPerPage     = 50
)

// IsAllowedMimeType vérifie si un type MIME est autorisé
func IsAllowedMimeType(mimeType string) bool {
	allowed := AllowedImageTypes + "," + AllowedVideoTypes + "," + AllowedAudioTypes
	for _, allowedType := range strings.Split(allowed, ",") {
		if allowedType == mimeType {
			return true
		}
	}
	return false
}
