package utils

import (
	"testing"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

// setupTestDB initialise la base de données de test
func setupTestDB() {
	database.InitTestDB()
	db := database.GetDB()
	db.AutoMigrate(&models.PasswordResetToken{}, &models.EmailVerificationToken{}, &models.BlacklistedToken{})
}

// TestGenerateSecureToken teste la génération de tokens sécurisés
func TestGenerateSecureToken(t *testing.T) {
	token, err := GenerateSecureToken()
	if err != nil {
		t.Fatalf("GenerateSecureToken failed: %v", err)
	}

	if token == "" {
		t.Error("Generated token should not be empty")
	}

	if len(token) != 64 { // 32 bytes * 2 (hex encoding)
		t.Errorf("Expected token length 64, got %d", len(token))
	}

	// Vérifier que deux tokens générés sont différents
	token2, err := GenerateSecureToken()
	if err != nil {
		t.Fatalf("GenerateSecureToken failed on second call: %v", err)
	}

	if token == token2 {
		t.Error("Two generated tokens should be different")
	}
}

// TestHashToken teste le hachage de tokens
func TestHashToken(t *testing.T) {
	token := "test-token-123"
	hash := HashToken(token)

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == token {
		t.Error("Hash should be different from original token")
	}

	if len(hash) != 64 { // SHA-256 hex = 64 caractères
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}

	// Vérifier que le même token produit toujours le même hash
	hash2 := HashToken(token)
	if hash != hash2 {
		t.Error("Same token should always produce same hash")
	}

	// Vérifier que des tokens différents produisent des hashs différents
	differentToken := "different-token-456"
	differentHash := HashToken(differentToken)
	if hash == differentHash {
		t.Error("Different tokens should produce different hashes")
	}
}

// TestCreatePasswordResetToken teste la création de tokens de réinitialisation
func TestCreatePasswordResetToken(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	userID := "test-user-id"

	token, err := CreatePasswordResetToken(userID)
	if err != nil {
		t.Fatalf("CreatePasswordResetToken failed: %v", err)
	}

	if token == "" {
		t.Error("Generated token should not be empty")
	}

	// Vérifier que le token a été sauvegardé en base
	var resetToken models.PasswordResetToken
	err = database.GetDB().Where("user_id = ? AND is_used = false", userID).First(&resetToken).Error
	if err != nil {
		t.Errorf("Token should be saved in database: %v", err)
	}

	// Vérifier que le token est haché en base
	if resetToken.Token == token {
		t.Error("Token should be hashed in database")
	}

	// Vérifier que le token expire dans le futur
	if resetToken.ExpiresAt.Before(time.Now()) {
		t.Error("Token should expire in the future")
	}
}

// TestCreatePasswordResetTokenInvalidatesOld teste que les anciens tokens sont invalidés
func TestCreatePasswordResetTokenInvalidatesOld(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	userID := "test-user-id"

	// Créer un premier token
	token1, err := CreatePasswordResetToken(userID)
	if err != nil {
		t.Fatalf("CreatePasswordResetToken failed: %v", err)
	}

	// Créer un second token
	token2, err := CreatePasswordResetToken(userID)
	if err != nil {
		t.Fatalf("CreatePasswordResetToken failed: %v", err)
	}

	// Vérifier que les tokens sont différents
	if token1 == token2 {
		t.Error("Two tokens should be different")
	}

	// Vérifier qu'il n'y a qu'un seul token valide
	var validTokens []models.PasswordResetToken
	database.GetDB().Where("user_id = ? AND is_used = false", userID).Find(&validTokens)

	if len(validTokens) != 1 {
		t.Errorf("Expected 1 valid token, got %d", len(validTokens))
	}
}

// TestValidatePasswordResetToken teste la validation de tokens
func TestValidatePasswordResetToken(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	userID := "test-user-id"

	// Créer un token
	token, err := CreatePasswordResetToken(userID)
	if err != nil {
		t.Fatalf("CreatePasswordResetToken failed: %v", err)
	}

	// Valider le token
	resetToken, err := ValidatePasswordResetToken(token)
	if err != nil {
		t.Fatalf("ValidatePasswordResetToken failed: %v", err)
	}

	if resetToken.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, resetToken.UserID)
	}

	if resetToken.IsUsed {
		t.Error("Token should not be used yet")
	}
}

// TestValidatePasswordResetTokenInvalid teste la validation avec un token invalide
func TestValidatePasswordResetTokenInvalid(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	// Tenter de valider un token inexistant
	_, err := ValidatePasswordResetToken("invalid-token")
	if err == nil {
		t.Error("ValidatePasswordResetToken should fail with invalid token")
	}
}

// TestUsePasswordResetToken teste l'utilisation d'un token
func TestUsePasswordResetToken(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	userID := "test-user-id"

	// Créer et valider un token
	token, _ := CreatePasswordResetToken(userID)
	resetToken, _ := ValidatePasswordResetToken(token)

	// Marquer le token comme utilisé
	err := UsePasswordResetToken(resetToken.ID)
	if err != nil {
		t.Fatalf("UsePasswordResetToken failed: %v", err)
	}

	// Vérifier que le token est marqué comme utilisé
	var usedToken models.PasswordResetToken
	database.GetDB().First(&usedToken, resetToken.ID)

	if !usedToken.IsUsed {
		t.Error("Token should be marked as used")
	}

	// Vérifier qu'on ne peut plus valider le token
	_, err = ValidatePasswordResetToken(token)
	if err == nil {
		t.Error("Used token should not validate")
	}
}

// TestCreateEmailVerificationToken teste la création de tokens de vérification d'email
func TestCreateEmailVerificationToken(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	userID := "test-user-id"

	token, err := CreateEmailVerificationToken(userID)
	if err != nil {
		t.Fatalf("CreateEmailVerificationToken failed: %v", err)
	}

	if token == "" {
		t.Error("Generated token should not be empty")
	}

	// Vérifier que le token a été sauvegardé en base
	var verificationToken models.EmailVerificationToken
	err = database.GetDB().Where("user_id = ? AND is_used = false", userID).First(&verificationToken).Error
	if err != nil {
		t.Errorf("Token should be saved in database: %v", err)
	}

	// Vérifier que le token expire dans le futur
	if verificationToken.ExpiresAt.Before(time.Now()) {
		t.Error("Token should expire in the future")
	}
}

// TestValidateEmailVerificationToken teste la validation de tokens de vérification
func TestValidateEmailVerificationToken(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	userID := "test-user-id"

	// Créer un token
	token, err := CreateEmailVerificationToken(userID)
	if err != nil {
		t.Fatalf("CreateEmailVerificationToken failed: %v", err)
	}

	// Valider le token
	verificationToken, err := ValidateEmailVerificationToken(token)
	if err != nil {
		t.Fatalf("ValidateEmailVerificationToken failed: %v", err)
	}

	if verificationToken.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, verificationToken.UserID)
	}

	if verificationToken.IsUsed {
		t.Error("Token should not be used yet")
	}
}

// TestUseEmailVerificationToken teste l'utilisation d'un token de vérification
func TestUseEmailVerificationToken(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	userID := "test-user-id"

	// Créer et valider un token
	token, _ := CreateEmailVerificationToken(userID)
	verificationToken, _ := ValidateEmailVerificationToken(token)

	// Marquer le token comme utilisé
	err := UseEmailVerificationToken(verificationToken.ID)
	if err != nil {
		t.Fatalf("UseEmailVerificationToken failed: %v", err)
	}

	// Vérifier que le token est marqué comme utilisé
	var usedToken models.EmailVerificationToken
	database.GetDB().First(&usedToken, verificationToken.ID)

	if !usedToken.IsUsed {
		t.Error("Token should be marked as used")
	}
}

// TestBlacklistJWTToken teste l'ajout de tokens JWT en liste noire
func TestBlacklistJWTToken(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	tokenString := "test.jwt.token"
	expiresAt := time.Now().Add(time.Hour)

	err := BlacklistJWTToken(tokenString, expiresAt)
	if err != nil {
		t.Fatalf("BlacklistJWTToken failed: %v", err)
	}

	// Vérifier que le token est en liste noire
	if !IsJWTTokenBlacklisted(tokenString) {
		t.Error("Token should be blacklisted")
	}
}

// TestIsJWTTokenBlacklisted teste la vérification de liste noire
func TestIsJWTTokenBlacklisted(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	tokenString := "test.jwt.token"

	// Vérifier qu'un token non blacklisté retourne false
	if IsJWTTokenBlacklisted(tokenString) {
		t.Error("Non-blacklisted token should return false")
	}

	// Ajouter le token à la liste noire
	expiresAt := time.Now().Add(time.Hour)
	BlacklistJWTToken(tokenString, expiresAt)

	// Vérifier qu'il retourne maintenant true
	if !IsJWTTokenBlacklisted(tokenString) {
		t.Error("Blacklisted token should return true")
	}
}

// TestCleanupExpiredTokens teste le nettoyage des tokens expirés
func TestCleanupExpiredTokens(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	userID := "test-user-id"

	// Créer un token de réinitialisation expiré
	expiredResetToken := &models.PasswordResetToken{
		UserID:    userID,
		Token:     HashToken("expired-reset-token"),
		ExpiresAt: time.Now().Add(-time.Hour), // Expiré il y a 1 heure
		IsUsed:    false,
	}
	database.GetDB().Create(expiredResetToken)

	// Créer un token de vérification expiré
	expiredVerificationToken := &models.EmailVerificationToken{
		UserID:    userID,
		Token:     HashToken("expired-verification-token"),
		ExpiresAt: time.Now().Add(-time.Hour), // Expiré il y a 1 heure
		IsUsed:    false,
	}
	database.GetDB().Create(expiredVerificationToken)

	// Créer un token JWT blacklisté expiré
	expiredJWTToken := &models.BlacklistedToken{
		Token:     HashToken("expired.jwt.token"),
		ExpiresAt: time.Now().Add(-time.Hour), // Expiré il y a 1 heure
	}
	database.GetDB().Create(expiredJWTToken)

	// Exécuter le nettoyage
	err := CleanupExpiredTokens()
	if err != nil {
		t.Fatalf("CleanupExpiredTokens failed: %v", err)
	}

	// Vérifier que les tokens expirés ont été supprimés
	var resetCount, verificationCount, jwtCount int64

	database.GetDB().Model(&models.PasswordResetToken{}).Where("expires_at < ?", time.Now()).Count(&resetCount)
	database.GetDB().Model(&models.EmailVerificationToken{}).Where("expires_at < ?", time.Now()).Count(&verificationCount)
	database.GetDB().Model(&models.BlacklistedToken{}).Where("expires_at < ?", time.Now()).Count(&jwtCount)

	if resetCount > 0 {
		t.Errorf("Expected 0 expired reset tokens, got %d", resetCount)
	}

	if verificationCount > 0 {
		t.Errorf("Expected 0 expired verification tokens, got %d", verificationCount)
	}

	if jwtCount > 0 {
		t.Errorf("Expected 0 expired JWT tokens, got %d", jwtCount)
	}
}
