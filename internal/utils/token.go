package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

// GenerateSecureToken génère un token sécurisé aléatoirement
func GenerateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HashToken crée un hash SHA-256 d'un token
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CreatePasswordResetToken crée un nouveau token de réinitialisation de mot de passe
func CreatePasswordResetToken(userID string) (string, error) {
	// Générer un token sécurisé
	token, err := GenerateSecureToken()
	if err != nil {
		return "", err
	}

	// Invalider les anciens tokens pour cet utilisateur
	database.GetDB().Model(&models.PasswordResetToken{}).
		Where("user_id = ? AND is_used = false", userID).
		Update("is_used", true)

	// Créer le nouveau token
	resetToken := models.PasswordResetToken{
		UserID:    userID,
		Token:     HashToken(token),
		ExpiresAt: time.Now().Add(time.Hour * 1), // Expire dans 1 heure
		IsUsed:    false,
	}

	if err := database.GetDB().Create(&resetToken).Error; err != nil {
		return "", err
	}

	return token, nil
}

// ValidatePasswordResetToken valide un token de réinitialisation de mot de passe
func ValidatePasswordResetToken(token string) (*models.PasswordResetToken, error) {
	hashedToken := HashToken(token)

	var resetToken models.PasswordResetToken
	err := database.GetDB().Where("token = ? AND is_used = false", hashedToken).First(&resetToken).Error
	if err != nil {
		return nil, err
	}

	if resetToken.IsExpired() {
		return nil, ErrTokenExpired
	}

	return &resetToken, nil
}

// UsePasswordResetToken marque un token de réinitialisation comme utilisé
func UsePasswordResetToken(tokenID uint) error {
	return database.GetDB().Model(&models.PasswordResetToken{}).
		Where("id = ?", tokenID).
		Update("is_used", true).Error
}

// CreateEmailVerificationToken crée un nouveau token de vérification d'email
func CreateEmailVerificationToken(userID string) (string, error) {
	// Générer un token sécurisé
	token, err := GenerateSecureToken()
	if err != nil {
		return "", err
	}

	// Invalider les anciens tokens pour cet utilisateur
	database.GetDB().Model(&models.EmailVerificationToken{}).
		Where("user_id = ? AND is_used = false", userID).
		Update("is_used", true)

	// Créer le nouveau token
	verificationToken := models.EmailVerificationToken{
		UserID:    userID,
		Token:     HashToken(token),
		ExpiresAt: time.Now().Add(time.Hour * 24), // Expire dans 24 heures
		IsUsed:    false,
	}

	if err := database.GetDB().Create(&verificationToken).Error; err != nil {
		return "", err
	}

	return token, nil
}

// ValidateEmailVerificationToken valide un token de vérification d'email
func ValidateEmailVerificationToken(token string) (*models.EmailVerificationToken, error) {
	hashedToken := HashToken(token)

	var verificationToken models.EmailVerificationToken
	err := database.GetDB().Where("token = ? AND is_used = false", hashedToken).
		Preload("User").First(&verificationToken).Error
	if err != nil {
		return nil, err
	}

	if verificationToken.IsExpired() {
		return nil, ErrTokenExpired
	}

	return &verificationToken, nil
}

// UseEmailVerificationToken marque un token de vérification d'email comme utilisé
func UseEmailVerificationToken(tokenID uint) error {
	return database.GetDB().Model(&models.EmailVerificationToken{}).
		Where("id = ?", tokenID).
		Update("is_used", true).Error
}

// BlacklistJWTToken ajoute un token JWT à la liste noire
func BlacklistJWTToken(tokenString string, expiresAt time.Time) error {
	hashedToken := HashToken(tokenString)

	blacklistedToken := models.BlacklistedToken{
		Token:     hashedToken,
		ExpiresAt: expiresAt,
	}

	return database.GetDB().Create(&blacklistedToken).Error
}

// IsJWTTokenBlacklisted vérifie si un token JWT est en liste noire
func IsJWTTokenBlacklisted(tokenString string) bool {
	hashedToken := HashToken(tokenString)

	var count int64
	database.GetDB().Model(&models.BlacklistedToken{}).
		Where("token = ? AND expires_at > ?", hashedToken, time.Now()).
		Count(&count)

	return count > 0
}

// CleanupExpiredTokens supprime les tokens expirés
func CleanupExpiredTokens() error {
	now := time.Now()

	// Supprimer les tokens de réinitialisation de mot de passe expirés
	if err := database.GetDB().Where("expires_at < ?", now).Delete(&models.PasswordResetToken{}).Error; err != nil {
		return err
	}

	// Supprimer les tokens de vérification d'email expirés
	if err := database.GetDB().Where("expires_at < ?", now).Delete(&models.EmailVerificationToken{}).Error; err != nil {
		return err
	}

	// Supprimer les tokens JWT en liste noire expirés
	return database.GetDB().Where("expires_at < ?", now).Delete(&models.BlacklistedToken{}).Error
}
