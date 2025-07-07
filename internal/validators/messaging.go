package validators

import (
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/Kaowarstail/Only-Flick-Go/internal/config"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

// MessageValidator valide les messages
type MessageValidator struct {
	config *config.MessagingConfig
}

// NewMessageValidator crée un nouveau validateur de messages
func NewMessageValidator() *MessageValidator {
	return &MessageValidator{
		config: config.GetMessagingConfig(),
	}
}

// ValidateMessage valide un message
func (v *MessageValidator) ValidateMessage(message *models.Message) error {
	// Vérifier que le message a un contenu ou un média
	if message.Content == "" && message.MediaURL == "" {
		return errors.New("le message doit avoir un contenu ou un média")
	}

	// Vérifier la longueur du contenu
	if len(message.Content) > v.config.MaxMessageLength {
		return fmt.Errorf("le contenu du message ne peut pas dépasser %d caractères", v.config.MaxMessageLength)
	}

	// Vérifier le prix des messages payants
	if message.IsPaid {
		if message.Price < v.config.MinPaidMessagePrice {
			return fmt.Errorf("le prix minimum pour un message payant est de %.2f", v.config.MinPaidMessagePrice)
		}
		if message.Price > v.config.MaxPaidMessagePrice {
			return fmt.Errorf("le prix maximum pour un message payant est de %.2f", v.config.MaxPaidMessagePrice)
		}
		if message.PreviewText == "" {
			return errors.New("un aperçu est requis pour les messages payants")
		}
		if len(message.PreviewText) > v.config.MaxPreviewLength {
			return fmt.Errorf("l'aperçu ne peut pas dépasser %d caractères", v.config.MaxPreviewLength)
		}
	}

	// Vérifier le type de message
	validTypes := []models.MessageType{
		models.MessageTypeText,
		models.MessageTypeImage,
		models.MessageTypeVideo,
		models.MessageTypeAudio,
		models.MessageTypeDocument,
		models.MessageTypePaidText,
		models.MessageTypePaidMedia,
	}
	
	valid := false
	for _, validType := range validTypes {
		if message.MessageType == validType {
			valid = true
			break
		}
	}
	
	if !valid {
		return errors.New("type de message invalide")
	}

	return nil
}

// ValidateConversation valide une conversation
func (v *MessageValidator) ValidateConversation(user1ID, user2ID string) error {
	if user1ID == "" || user2ID == "" {
		return errors.New("les IDs des participants sont requis")
	}

	if user1ID == user2ID {
		return errors.New("un utilisateur ne peut pas créer une conversation avec lui-même")
	}

	return nil
}

// FileValidator valide les fichiers uploadés
type FileValidator struct {
	config *config.UploadConfig
}

// NewFileValidator crée un nouveau validateur de fichiers
func NewFileValidator() *FileValidator {
	return &FileValidator{
		config: config.GetUploadConfig(),
	}
}

// ValidateFile valide un fichier uploadé
func (v *FileValidator) ValidateFile(fileHeader *multipart.FileHeader) error {
	// Vérifier la taille du fichier
	if fileHeader.Size > v.config.MaxFileSize {
		return fmt.Errorf("la taille du fichier ne peut pas dépasser %d MB", v.config.MaxFileSize/(1024*1024))
	}

	// Vérifier le type de fichier
	mediaType := v.GetMediaType(fileHeader)
	if mediaType == "" {
		return errors.New("type de fichier non supporté")
	}

	// Vérifier la taille selon le type de média
	switch mediaType {
	case "image":
		if fileHeader.Size > v.config.MaxImageSize {
			return fmt.Errorf("la taille de l'image ne peut pas dépasser %d MB", v.config.MaxImageSize/(1024*1024))
		}
		if !v.isAllowedType(fileHeader.Header.Get("Content-Type"), v.config.AllowedImageTypes) {
			return errors.New("type d'image non supporté")
		}
	case "video":
		if fileHeader.Size > v.config.MaxVideoSize {
			return fmt.Errorf("la taille de la vidéo ne peut pas dépasser %d MB", v.config.MaxVideoSize/(1024*1024))
		}
		if !v.isAllowedType(fileHeader.Header.Get("Content-Type"), v.config.AllowedVideoTypes) {
			return errors.New("type de vidéo non supporté")
		}
	case "audio":
		if fileHeader.Size > v.config.MaxAudioSize {
			return fmt.Errorf("la taille de l'audio ne peut pas dépasser %d MB", v.config.MaxAudioSize/(1024*1024))
		}
		if !v.isAllowedType(fileHeader.Header.Get("Content-Type"), v.config.AllowedAudioTypes) {
			return errors.New("type d'audio non supporté")
		}
	case "document":
		if fileHeader.Size > v.config.MaxDocumentSize {
			return fmt.Errorf("la taille du document ne peut pas dépasser %d MB", v.config.MaxDocumentSize/(1024*1024))
		}
		if !v.isAllowedType(fileHeader.Header.Get("Content-Type"), v.config.AllowedDocTypes) {
			return errors.New("type de document non supporté")
		}
	}

	return nil
}

// GetMediaType détermine le type de média basé sur l'extension du fichier
func (v *FileValidator) GetMediaType(fileHeader *multipart.FileHeader) string {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return "image"
	case ".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm":
		return "video"
	case ".mp3", ".wav", ".ogg", ".m4a", ".aac":
		return "audio"
	case ".pdf", ".txt":
		return "document"
	default:
		return ""
	}
}

// isAllowedType vérifie si le type MIME est autorisé
func (v *FileValidator) isAllowedType(mimeType string, allowedTypes []string) bool {
	for _, allowedType := range allowedTypes {
		if mimeType == allowedType {
			return true
		}
	}
	return false
}

// ValidateFileName valide le nom du fichier
func (v *FileValidator) ValidateFileName(filename string) error {
	if filename == "" {
		return errors.New("nom de fichier requis")
	}

	// Vérifier les caractères dangereux
	dangerousChars := []string{"..", "/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range dangerousChars {
		if strings.Contains(filename, char) {
			return errors.New("nom de fichier contient des caractères non autorisés")
		}
	}

	// Vérifier la longueur du nom de fichier
	if len(filename) > 255 {
		return errors.New("nom de fichier trop long")
	}

	return nil
}

// ProfileValidator valide les données de profil
type ProfileValidator struct{}

// NewProfileValidator crée un nouveau validateur de profil
func NewProfileValidator() *ProfileValidator {
	return &ProfileValidator{}
}

// ValidateUsername valide un nom d'utilisateur
func (v *ProfileValidator) ValidateUsername(username string) error {
	if len(username) < 3 {
		return errors.New("le nom d'utilisateur doit avoir au moins 3 caractères")
	}
	if len(username) > 30 {
		return errors.New("le nom d'utilisateur ne peut pas dépasser 30 caractères")
	}
	
	// Vérifier les caractères autorisés (lettres, chiffres, underscore, tiret)
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || char == '_' || char == '-') {
			return errors.New("le nom d'utilisateur ne peut contenir que des lettres, chiffres, _ et -")
		}
	}

	return nil
}

// ValidateEmail valide une adresse email
func (v *ProfileValidator) ValidateEmail(email string) error {
	if email == "" {
		return errors.New("adresse email requise")
	}

	// Validation basique de l'email
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return errors.New("adresse email invalide")
	}

	if len(email) > 254 {
		return errors.New("adresse email trop longue")
	}

	return nil
}

// ValidateBiography valide une biographie
func (v *ProfileValidator) ValidateBiography(bio string) error {
	if len(bio) > 1000 {
		return errors.New("la biographie ne peut pas dépasser 1000 caractères")
	}
	return nil
}

// ValidatePrice valide un prix
func (v *ProfileValidator) ValidatePrice(price float64) error {
	if price < 0 {
		return errors.New("le prix ne peut pas être négatif")
	}
	if price > 1000 {
		return errors.New("le prix ne peut pas dépasser 1000")
	}
	return nil
}

// ValidateSocialLinks valide les liens sociaux
func (v *ProfileValidator) ValidateSocialLinks(links map[string]interface{}) error {
	allowedPlatforms := []string{"instagram", "twitter", "tiktok", "youtube", "website", "onlyfans"}
	
	for platform, link := range links {
		// Vérifier que la plateforme est autorisée
		allowed := false
		for _, allowedPlatform := range allowedPlatforms {
			if platform == allowedPlatform {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("plateforme non autorisée: %s", platform)
		}

		// Vérifier que le lien est une chaîne
		linkStr, ok := link.(string)
		if !ok {
			return fmt.Errorf("lien invalide pour %s", platform)
		}

		// Vérifier la longueur du lien
		if len(linkStr) > 255 {
			return fmt.Errorf("lien trop long pour %s", platform)
		}

		// Validation basique d'URL pour le site web
		if platform == "website" && linkStr != "" {
			if !strings.HasPrefix(linkStr, "http://") && !strings.HasPrefix(linkStr, "https://") {
				return errors.New("l'URL du site web doit commencer par http:// ou https://")
			}
		}
	}

	return nil
}
