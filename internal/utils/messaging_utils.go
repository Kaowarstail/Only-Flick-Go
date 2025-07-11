package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// TruncateString tronque une string à une longueur donnée
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	if maxLength <= 3 {
		return s[:maxLength]
	}

	return s[:maxLength-3] + "..."
}

// GenerateConversationTitle génère un titre pour une conversation
func GenerateConversationTitle(participants []string, currentUserID string) string {
	var otherParticipants []string

	for _, participant := range participants {
		if participant != currentUserID {
			otherParticipants = append(otherParticipants, participant)
		}
	}

	if len(otherParticipants) == 0 {
		return "Conversation"
	}

	if len(otherParticipants) == 1 {
		return otherParticipants[0]
	}

	if len(otherParticipants) <= 3 {
		return strings.Join(otherParticipants, ", ")
	}

	return strings.Join(otherParticipants[:2], ", ") + " et " +
		fmt.Sprintf("%d autres", len(otherParticipants)-2)
}

// ExtractMentions extrait les mentions (@username) d'un message
func ExtractMentions(content string) []string {
	mentionRegex := regexp.MustCompile(`@([a-zA-Z0-9_]+)`)
	matches := mentionRegex.FindAllStringSubmatch(content, -1)

	var mentions []string
	for _, match := range matches {
		if len(match) > 1 {
			mentions = append(mentions, match[1])
		}
	}

	return mentions
}

// CleanMessageContent nettoie et valide le contenu d'un message
func CleanMessageContent(content string) string {
	if content == "" {
		return ""
	}

	// Nettoyer le contenu
	cleaned := SanitizeMessageContent(content)

	// Supprimer les espaces multiples
	spaceRegex := regexp.MustCompile(`\s+`)
	cleaned = spaceRegex.ReplaceAllString(cleaned, " ")

	// Supprimer les lignes vides multiples
	lineRegex := regexp.MustCompile(`\n\s*\n\s*\n`)
	cleaned = lineRegex.ReplaceAllString(cleaned, "\n\n")

	return strings.TrimSpace(cleaned)
}

// IsValidMediaURL vérifie si une URL de média est valide
func IsValidMediaURL(mediaURL string) bool {
	if mediaURL == "" {
		return false
	}

	// Parser l'URL
	parsedURL, err := url.Parse(mediaURL)
	if err != nil {
		return false
	}

	// Vérifier le schéma
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	// Vérifier qu'il y a un host
	if parsedURL.Host == "" {
		return false
	}

	return true
}

// GetFileExtensionFromMimeType retourne l'extension de fichier pour un type MIME
func GetFileExtensionFromMimeType(mimeType string) string {
	extensions := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
		"image/webp": ".webp",
		"video/mp4":  ".mp4",
		"video/webm": ".webm",
		"audio/mp3":  ".mp3",
		"audio/wav":  ".wav",
		"audio/ogg":  ".ogg",
	}

	if ext, exists := extensions[mimeType]; exists {
		return ext
	}

	return ""
}

// FormatFileSize formate une taille de fichier en octets vers une chaîne lisible
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// IsEmptyMessage vérifie si un message est considéré comme vide
func IsEmptyMessage(content *string, mediaURL *string) bool {
	hasContent := content != nil && strings.TrimSpace(*content) != ""
	hasMedia := mediaURL != nil && strings.TrimSpace(*mediaURL) != ""

	return !hasContent && !hasMedia
}

// NormalizeSearchQuery normalise une requête de recherche
func NormalizeSearchQuery(query string) string {
	// Supprimer les espaces en début/fin
	query = strings.TrimSpace(query)

	// Convertir en minuscules
	query = strings.ToLower(query)

	// Supprimer les caractères spéciaux pour la recherche
	query = regexp.MustCompile(`[^\w\s\-_]`).ReplaceAllString(query, "")

	// Supprimer les espaces multiples
	query = regexp.MustCompile(`\s+`).ReplaceAllString(query, " ")

	return query
}

// BuildSearchPattern construit un pattern de recherche SQL LIKE
func BuildSearchPattern(query string) string {
	// Échapper les caractères spéciaux SQL LIKE
	query = strings.ReplaceAll(query, "%", `\%`)
	query = strings.ReplaceAll(query, "_", `\_`)

	// Ajouter des wildcards
	return "%" + query + "%"
}

// ValidateConversationType valide un type de conversation
func ValidateConversationType(conversationType string) bool {
	validTypes := map[string]bool{
		"direct": true,
		"group":  true,
	}

	return validTypes[conversationType]
}

// ValidateMessageType valide un type de message
func ValidateMessageType(messageType string) bool {
	validTypes := map[string]bool{
		"text":  true,
		"image": true,
		"video": true,
		"audio": true,
		"file":  true,
	}

	return validTypes[messageType]
}

// ValidateMessageStatus valide un statut de message
func ValidateMessageStatus(status string) bool {
	validStatuses := map[string]bool{
		"sending":   true,
		"sent":      true,
		"delivered": true,
		"read":      true,
		"failed":    true,
	}

	return validStatuses[status]
}

// GetMimeTypeFromExtension retourne le type MIME pour une extension de fichier
func GetMimeTypeFromExtension(extension string) string {
	extension = strings.ToLower(extension)
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".mp3":  "audio/mp3",
		".wav":  "audio/wav",
		".ogg":  "audio/ogg",
	}

	if mimeType, exists := mimeTypes[extension]; exists {
		return mimeType
	}

	return ""
}
