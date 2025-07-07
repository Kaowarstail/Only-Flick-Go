package config

import (
	"os"
	"strconv"
)

// UploadConfig contient la configuration pour les uploads
type UploadConfig struct {
	MaxFileSize       int64
	MaxImageSize      int64
	MaxVideoSize      int64
	MaxAudioSize      int64
	MaxDocumentSize   int64
	AllowedImageTypes []string
	AllowedVideoTypes []string
	AllowedAudioTypes []string
	AllowedDocTypes   []string
	UploadPath        string
	ThumbnailPath     string
	ThumbnailSize     int
}

// GetUploadConfig retourne la configuration des uploads
func GetUploadConfig() *UploadConfig {
	maxFileSize := int64(50 * 1024 * 1024) // 50MB par défaut
	if size := os.Getenv("MAX_FILE_SIZE"); size != "" {
		if parsed, err := strconv.ParseInt(size, 10, 64); err == nil {
			maxFileSize = parsed
		}
	}

	uploadPath := "uploads"
	if path := os.Getenv("UPLOAD_PATH"); path != "" {
		uploadPath = path
	}

	return &UploadConfig{
		MaxFileSize:     maxFileSize,
		MaxImageSize:    10 * 1024 * 1024,  // 10MB
		MaxVideoSize:    100 * 1024 * 1024, // 100MB
		MaxAudioSize:    20 * 1024 * 1024,  // 20MB
		MaxDocumentSize: 10 * 1024 * 1024,  // 10MB
		AllowedImageTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
		},
		AllowedVideoTypes: []string{
			"video/mp4",
			"video/avi",
			"video/mov",
			"video/wmv",
			"video/webm",
		},
		AllowedAudioTypes: []string{
			"audio/mp3",
			"audio/wav",
			"audio/ogg",
			"audio/m4a",
			"audio/aac",
		},
		AllowedDocTypes: []string{
			"application/pdf",
			"text/plain",
		},
		UploadPath:     uploadPath,
		ThumbnailPath:  uploadPath + "/thumbnails",
		ThumbnailSize:  200,
	}
}

// MessagingConfig contient la configuration pour la messagerie
type MessagingConfig struct {
	MaxMessageLength        int
	MaxPreviewLength        int
	MinPaidMessagePrice     float64
	MaxPaidMessagePrice     float64
	PlatformCommissionRate  float64
	MaxMessagesPerMinute    int
	MaxConversationsPerUser int
	MessageRetentionDays    int
}

// GetMessagingConfig retourne la configuration de la messagerie
func GetMessagingConfig() *MessagingConfig {
	maxMessageLength := 5000
	if length := os.Getenv("MAX_MESSAGE_LENGTH"); length != "" {
		if parsed, err := strconv.Atoi(length); err == nil {
			maxMessageLength = parsed
		}
	}

	minPrice := 0.99
	if price := os.Getenv("MIN_PAID_MESSAGE_PRICE"); price != "" {
		if parsed, err := strconv.ParseFloat(price, 64); err == nil {
			minPrice = parsed
		}
	}

	maxPrice := 500.0
	if price := os.Getenv("MAX_PAID_MESSAGE_PRICE"); price != "" {
		if parsed, err := strconv.ParseFloat(price, 64); err == nil {
			maxPrice = parsed
		}
	}

	return &MessagingConfig{
		MaxMessageLength:        maxMessageLength,
		MaxPreviewLength:        100,
		MinPaidMessagePrice:     minPrice,
		MaxPaidMessagePrice:     maxPrice,
		PlatformCommissionRate:  0.20, // 20%
		MaxMessagesPerMinute:    30,
		MaxConversationsPerUser: 1000,
		MessageRetentionDays:    365,
	}
}

// SecurityConfig contient la configuration de sécurité
type SecurityConfig struct {
	JWTSecret                string
	JWTExpirationHours       int
	MaxLoginAttempts         int
	LoginLockoutMinutes      int
	RequireEmailVerification bool
	PasswordMinLength        int
	EnableRateLimiting       bool
	RateLimitRequests        int
	RateLimitWindow          int
}

// GetSecurityConfig retourne la configuration de sécurité
func GetSecurityConfig() *SecurityConfig {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-this-in-production"
	}

	jwtExpiration := 24 // 24 heures par défaut
	if hours := os.Getenv("JWT_EXPIRATION_HOURS"); hours != "" {
		if parsed, err := strconv.Atoi(hours); err == nil {
			jwtExpiration = parsed
		}
	}

	return &SecurityConfig{
		JWTSecret:                jwtSecret,
		JWTExpirationHours:       jwtExpiration,
		MaxLoginAttempts:         5,
		LoginLockoutMinutes:      30,
		RequireEmailVerification: true,
		PasswordMinLength:        8,
		EnableRateLimiting:       true,
		RateLimitRequests:        60,
		RateLimitWindow:          60, // 60 secondes
	}
}
