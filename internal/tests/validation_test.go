package main

import (
	"fmt"
	"log"
	"testing"

	"github.com/Kaowarstail/Only-Flick-Go/internal/config"
	"github.com/Kaowarstail/Only-Flick-Go/internal/validators"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

// TestMessageValidation teste la validation des messages
func TestMessageValidation(t *testing.T) {
	validator := validators.NewMessageValidator()

	// Test message valide
	message := &models.Message{
		Content:     "Test message",
		MessageType: models.MessageTypeText,
		IsPaid:      false,
	}

	err := validator.ValidateMessage(message)
	if err != nil {
		t.Errorf("Message valide rejeté: %v", err)
	}

	// Test message payant valide
	paidMessage := &models.Message{
		Content:     "Paid content",
		MessageType: models.MessageTypePaidText,
		IsPaid:      true,
		Price:       5.99,
		PreviewText: "Preview text",
	}

	err = validator.ValidateMessage(paidMessage)
	if err != nil {
		t.Errorf("Message payant valide rejeté: %v", err)
	}

	// Test message sans contenu
	emptyMessage := &models.Message{
		MessageType: models.MessageTypeText,
		IsPaid:      false,
	}

	err = validator.ValidateMessage(emptyMessage)
	if err == nil {
		t.Error("Message vide accepté")
	}

	// Test prix trop bas
	lowPriceMessage := &models.Message{
		Content:     "Paid content",
		MessageType: models.MessageTypePaidText,
		IsPaid:      true,
		Price:       0.50,
		PreviewText: "Preview text",
	}

	err = validator.ValidateMessage(lowPriceMessage)
	if err == nil {
		t.Error("Prix trop bas accepté")
	}
}

// TestProfileValidation teste la validation des profils
func TestProfileValidation(t *testing.T) {
	validator := validators.NewProfileValidator()

	// Test username valide
	err := validator.ValidateUsername("validuser")
	if err != nil {
		t.Errorf("Username valide rejeté: %v", err)
	}

	// Test username trop court
	err = validator.ValidateUsername("ab")
	if err == nil {
		t.Error("Username trop court accepté")
	}

	// Test username trop long
	err = validator.ValidateUsername("verylongusernamethatexceedslimit")
	if err == nil {
		t.Error("Username trop long accepté")
	}

	// Test email valide
	err = validator.ValidateEmail("test@example.com")
	if err != nil {
		t.Errorf("Email valide rejeté: %v", err)
	}

	// Test email invalide
	err = validator.ValidateEmail("invalid-email")
	if err == nil {
		t.Error("Email invalide accepté")
	}

	// Test biographie valide
	err = validator.ValidateBiography("This is a valid biography")
	if err != nil {
		t.Errorf("Biographie valide rejetée: %v", err)
	}

	// Test biographie trop longue
	longBio := make([]byte, 1001)
	for i := range longBio {
		longBio[i] = 'a'
	}
	err = validator.ValidateBiography(string(longBio))
	if err == nil {
		t.Error("Biographie trop longue acceptée")
	}
}

// TestConfigLoading teste le chargement de la configuration
func TestConfigLoading(t *testing.T) {
	uploadConfig := config.GetUploadConfig()
	if uploadConfig.MaxFileSize <= 0 {
		t.Error("MaxFileSize doit être positif")
	}

	if len(uploadConfig.AllowedImageTypes) == 0 {
		t.Error("AllowedImageTypes ne doit pas être vide")
	}

	messagingConfig := config.GetMessagingConfig()
	if messagingConfig.MaxMessageLength <= 0 {
		t.Error("MaxMessageLength doit être positif")
	}

	if messagingConfig.MinPaidMessagePrice <= 0 {
		t.Error("MinPaidMessagePrice doit être positif")
	}

	if messagingConfig.PlatformCommissionRate <= 0 || messagingConfig.PlatformCommissionRate >= 1 {
		t.Error("PlatformCommissionRate doit être entre 0 et 1")
	}
}

// TestConversationValidation teste la validation des conversations
func TestConversationValidation(t *testing.T) {
	validator := validators.NewMessageValidator()

	// Test conversation valide
	err := validator.ValidateConversation("user1", "user2")
	if err != nil {
		t.Errorf("Conversation valide rejetée: %v", err)
	}

	// Test conversation avec même utilisateur
	err = validator.ValidateConversation("user1", "user1")
	if err == nil {
		t.Error("Conversation avec même utilisateur acceptée")
	}

	// Test conversation avec utilisateur vide
	err = validator.ValidateConversation("", "user2")
	if err == nil {
		t.Error("Conversation avec utilisateur vide acceptée")
	}
}

// TestSocialLinksValidation teste la validation des liens sociaux
func TestSocialLinksValidation(t *testing.T) {
	validator := validators.NewProfileValidator()

	// Test liens sociaux valides
	validLinks := map[string]interface{}{
		"instagram": "username",
		"twitter":   "username",
		"website":   "https://example.com",
	}

	err := validator.ValidateSocialLinks(validLinks)
	if err != nil {
		t.Errorf("Liens sociaux valides rejetés: %v", err)
	}

	// Test plateforme non autorisée
	invalidPlatform := map[string]interface{}{
		"facebook": "username",
	}

	err = validator.ValidateSocialLinks(invalidPlatform)
	if err == nil {
		t.Error("Plateforme non autorisée acceptée")
	}

	// Test URL invalide
	invalidURL := map[string]interface{}{
		"website": "invalid-url",
	}

	err = validator.ValidateSocialLinks(invalidURL)
	if err == nil {
		t.Error("URL invalide acceptée")
	}
}

// TestPriceValidation teste la validation des prix
func TestPriceValidation(t *testing.T) {
	validator := validators.NewProfileValidator()

	// Test prix valide
	err := validator.ValidatePrice(9.99)
	if err != nil {
		t.Errorf("Prix valide rejeté: %v", err)
	}

	// Test prix négatif
	err = validator.ValidatePrice(-1.00)
	if err == nil {
		t.Error("Prix négatif accepté")
	}

	// Test prix trop élevé
	err = validator.ValidatePrice(1001.00)
	if err == nil {
		t.Error("Prix trop élevé accepté")
	}
}

// BenchmarkMessageValidation benchmarque la validation des messages
func BenchmarkMessageValidation(b *testing.B) {
	validator := validators.NewMessageValidator()
	message := &models.Message{
		Content:     "Test message",
		MessageType: models.MessageTypeText,
		IsPaid:      false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateMessage(message)
	}
}

// BenchmarkProfileValidation benchmarque la validation des profils
func BenchmarkProfileValidation(b *testing.B) {
	validator := validators.NewProfileValidator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateUsername("testuser")
		validator.ValidateEmail("test@example.com")
		validator.ValidateBiography("Test biography")
	}
}

// Fonction principale pour exécuter les tests
func main() {
	fmt.Println("Exécution des tests OnlyFlick...")

	// Vous pouvez exécuter les tests individuellement ici
	// ou utiliser `go test` pour exécuter tous les tests

	testing.Main(func(pat, str string) (bool, error) {
		return true, nil
	}, []testing.InternalTest{
		{
			Name: "TestMessageValidation",
			F:    TestMessageValidation,
		},
		{
			Name: "TestProfileValidation",
			F:    TestProfileValidation,
		},
		{
			Name: "TestConfigLoading",
			F:    TestConfigLoading,
		},
		{
			Name: "TestConversationValidation",
			F:    TestConversationValidation,
		},
		{
			Name: "TestSocialLinksValidation",
			F:    TestSocialLinksValidation,
		},
		{
			Name: "TestPriceValidation",
			F:    TestPriceValidation,
		},
	}, []testing.InternalBenchmark{
		{
			Name: "BenchmarkMessageValidation",
			F:    BenchmarkMessageValidation,
		},
		{
			Name: "BenchmarkProfileValidation",
			F:    BenchmarkProfileValidation,
		},
	}, []testing.InternalExample{})

	log.Println("Tests terminés.")
}
