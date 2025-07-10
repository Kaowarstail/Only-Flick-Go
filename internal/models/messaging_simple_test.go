package models

import (
	"testing"
)

func TestUUIDGeneration(t *testing.T) {
	uuid1, err := generateUUID()
	if err != nil {
		t.Fatalf("Failed to generate UUID: %v", err)
	}

	uuid2, err := generateUUID()
	if err != nil {
		t.Fatalf("Failed to generate second UUID: %v", err)
	}

	if uuid1 == uuid2 {
		t.Error("Generated UUIDs should be unique")
	}

	// Vérifier le format UUID
	if len(uuid1) != 36 {
		t.Errorf("Expected UUID length 36, got %d", len(uuid1))
	}

	// Vérifier qu'il contient des tirets aux bonnes positions
	if uuid1[8] != '-' || uuid1[13] != '-' || uuid1[18] != '-' || uuid1[23] != '-' {
		t.Errorf("UUID format incorrect: %s", uuid1)
	}
}

func TestMessageValidation(t *testing.T) {
	// Test validation avec contenu valide
	content := "Valid message content"
	message := MessageClassic{
		Content:     &content,
		MessageType: MessageTypeText,
	}

	err := message.ValidateContent()
	if err != nil {
		t.Errorf("Expected no error for valid content, got: %v", err)
	}

	// Test validation sans contenu ni média
	emptyMessage := MessageClassic{}
	err = emptyMessage.ValidateContent()
	if err == nil {
		t.Error("Expected error for empty message")
	}

	// Test validation avec contenu trop long
	longContent := make([]byte, 6000)
	for i := range longContent {
		longContent[i] = 'a'
	}
	longContentStr := string(longContent)
	longMessage := MessageClassic{
		Content: &longContentStr,
	}

	err = longMessage.ValidateContent()
	if err == nil {
		t.Error("Expected error for content too long")
	}
}

func TestMessageTypeConstants(t *testing.T) {
	// Test des constantes de type de message
	if MessageTypeText != "text" {
		t.Errorf("Expected MessageTypeText to be 'text', got '%s'", MessageTypeText)
	}

	if MessageTypeImage != "image" {
		t.Errorf("Expected MessageTypeImage to be 'image', got '%s'", MessageTypeImage)
	}

	if MessageTypeVideo != "video" {
		t.Errorf("Expected MessageTypeVideo to be 'video', got '%s'", MessageTypeVideo)
	}

	if MessageTypeAudio != "audio" {
		t.Errorf("Expected MessageTypeAudio to be 'audio', got '%s'", MessageTypeAudio)
	}
}

func TestMessageStatusConstants(t *testing.T) {
	// Test des constantes de statut de message
	if MessageStatusSending != "sending" {
		t.Errorf("Expected MessageStatusSending to be 'sending', got '%s'", MessageStatusSending)
	}

	if MessageStatusSent != "sent" {
		t.Errorf("Expected MessageStatusSent to be 'sent', got '%s'", MessageStatusSent)
	}

	if MessageStatusDelivered != "delivered" {
		t.Errorf("Expected MessageStatusDelivered to be 'delivered', got '%s'", MessageStatusDelivered)
	}

	if MessageStatusRead != "read" {
		t.Errorf("Expected MessageStatusRead to be 'read', got '%s'", MessageStatusRead)
	}

	if MessageStatusFailed != "failed" {
		t.Errorf("Expected MessageStatusFailed to be 'failed', got '%s'", MessageStatusFailed)
	}
}

func TestGetDisplayContent(t *testing.T) {
	// Test avec contenu texte
	content := "Hello World"
	message := MessageClassic{
		Content:     &content,
		MessageType: MessageTypeText,
	}

	display := message.GetDisplayContent()
	if display != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", display)
	}

	// Test avec image sans contenu
	imageMessage := MessageClassic{
		MessageType: MessageTypeImage,
		MediaURL:    stringPtr("http://example.com/image.jpg"),
	}

	display = imageMessage.GetDisplayContent()
	if display != "📸 Image" {
		t.Errorf("Expected '📸 Image', got '%s'", display)
	}

	// Test avec vidéo sans contenu
	videoMessage := MessageClassic{
		MessageType: MessageTypeVideo,
		MediaURL:    stringPtr("http://example.com/video.mp4"),
	}

	display = videoMessage.GetDisplayContent()
	if display != "🎥 Vidéo" {
		t.Errorf("Expected '🎥 Vidéo', got '%s'", display)
	}

	// Test avec audio sans contenu
	audioMessage := MessageClassic{
		MessageType: MessageTypeAudio,
		MediaURL:    stringPtr("http://example.com/audio.mp3"),
	}

	display = audioMessage.GetDisplayContent()
	if display != "🎵 Audio" {
		t.Errorf("Expected '🎵 Audio', got '%s'", display)
	}
}

func TestIsMediaMessage(t *testing.T) {
	// Test avec média
	message := MessageClassic{
		MediaURL: stringPtr("http://example.com/image.jpg"),
	}

	if !message.IsMediaMessage() {
		t.Error("Expected message with MediaURL to be a media message")
	}

	// Test sans média
	content := "Just text"
	textMessage := MessageClassic{
		Content: &content,
	}

	if textMessage.IsMediaMessage() {
		t.Error("Expected text-only message to not be a media message")
	}

	// Test avec MediaURL vide
	emptyURL := ""
	emptyMessage := MessageClassic{
		MediaURL: &emptyURL,
	}

	if emptyMessage.IsMediaMessage() {
		t.Error("Expected message with empty MediaURL to not be a media message")
	}
}

// Helper function pour créer un pointeur string
func stringPtr(s string) *string {
	return &s
}
