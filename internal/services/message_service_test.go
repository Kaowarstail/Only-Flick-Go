package services

import (
	"testing"

	"github.com/Kaowarstail/Only-Flick-Go/internal/dto"
	internalmodels "github.com/Kaowarstail/Only-Flick-Go/internal/models"
	"github.com/Kaowarstail/Only-Flick-Go/models"
	"github.com/stretchr/testify/assert"
)

func TestMessageService_SendMessage(t *testing.T) {
	db := setupTestDB()
	service := NewMessageService(db)
	conversationService := NewConversationService(db)

	// Setup
	user1 := models.User{ID: "user-1", Username: "user1", Email: "user1@test.com"}
	user2 := models.User{ID: "user-2", Username: "user2", Email: "user2@test.com"}
	db.Create(&user1)
	db.Create(&user2)

	conv, _ := conversationService.CreateOrGetConversation("user-1", "user-2")

	// Test envoi message
	content := "Hello World"
	req := dto.SendMessageRequest{
		ConversationID: conv.ID,
		Content:        &content,
		MessageType:    "text",
	}

	message, err := service.SendMessage(req, "user-1")
	assert.NoError(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "Hello World", *message.Content)
	assert.Equal(t, "user-1", message.SenderID)
	assert.False(t, message.IsPaid)
}

func TestMessageService_SendPaidMessage(t *testing.T) {
	db := setupTestDB()
	service := NewMessageService(db)
	conversationService := NewConversationService(db)

	// Setup
	user1 := models.User{ID: "user-1", Username: "user1", Email: "user1@test.com"}
	user2 := models.User{ID: "user-2", Username: "user2", Email: "user2@test.com"}
	db.Create(&user1)
	db.Create(&user2)

	conv, _ := conversationService.CreateOrGetConversation("user-1", "user-2")

	// Test envoi message payant
	content := "Premium Content"
	req := dto.SendPaidMessageRequest{
		ConversationID: conv.ID,
		Content:        &content,
		Price:          9.99,
		MessageType:    "paid_text",
	}

	message, err := service.SendPaidMessage(req, "user-1")
	assert.NoError(t, err)
	assert.NotNil(t, message)
	assert.True(t, message.IsPaid)
	assert.Equal(t, 9.99, *message.Price)
	assert.False(t, message.IsUnlocked)
}

func TestMessageService_UnlockPaidMessage(t *testing.T) {
	db := setupTestDB()
	service := NewMessageService(db)
	conversationService := NewConversationService(db)

	// Setup
	user1 := models.User{ID: "user-1", Username: "user1", Email: "user1@test.com"}
	user2 := models.User{ID: "user-2", Username: "user2", Email: "user2@test.com"}
	db.Create(&user1)
	db.Create(&user2)

	conv, _ := conversationService.CreateOrGetConversation("user-1", "user-2")

	// Créer message payant
	price := 5.00
	message := internalmodels.EnhancedMessage{
		ConversationID: conv.ID,
		SenderID:       "user-1",
		Content:        stringPtr("Paid content"),
		IsPaid:         true,
		Price:          &price,
		IsUnlocked:     false,
		MessageType:    "paid_text",
	}
	db.Create(&message)

	// Test déblocage
	err := service.UnlockPaidMessage(message.ID, "user-2")
	assert.NoError(t, err)

	// Vérifier déblocage
	var updatedMessage internalmodels.EnhancedMessage
	db.First(&updatedMessage, message.ID)
	assert.True(t, updatedMessage.IsUnlocked)
	assert.Equal(t, "user-2", *updatedMessage.UnlockedBy)

	// Vérifier transaction créée
	var transaction internalmodels.PaidMessageTransaction
	err = db.Where("message_id = ?", message.ID).First(&transaction).Error
	assert.NoError(t, err)
	assert.Equal(t, "user-2", transaction.BuyerID)
	assert.Equal(t, "user-1", transaction.SellerID)
	assert.Equal(t, 5.00, transaction.Amount)
}

func stringPtr(s string) *string {
	return &s
}
