package services

import (
	"testing"

	internalmodels "github.com/Kaowarstail/Only-Flick-Go/internal/models"
	"github.com/Kaowarstail/Only-Flick-Go/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	// Try to connect to SQLite in memory
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		// Fallback: try regular file-based SQLite
		db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			panic("failed to connect database: " + err.Error())
		}
	}

	err = db.AutoMigrate(
		&models.User{},
		&internalmodels.Conversation{},
		&internalmodels.EnhancedMessage{},
		&internalmodels.PaidMessageTransaction{},
	)
	if err != nil {
		panic("failed to migrate database: " + err.Error())
	}

	return db
}

func TestConversationService_CreateOrGetConversation(t *testing.T) {
	db := setupTestDB()
	service := NewConversationService(db)
	// Créer des utilisateurs test
	user1 := models.User{ID: "user-1", Username: "user1", Email: "user1@test.com"}
	user2 := models.User{ID: "user-2", Username: "user2", Email: "user2@test.com"}
	db.Create(&user1)
	db.Create(&user2)

	// Test création nouvelle conversation
	conv, err := service.CreateOrGetConversation("user-1", "user-2")
	assert.NoError(t, err)
	assert.NotNil(t, conv)
	assert.Equal(t, "user-1", conv.Participant1ID)
	assert.Equal(t, "user-2", conv.Participant2ID)

	// Test récupération conversation existante
	conv2, err := service.CreateOrGetConversation("user-2", "user-1") // Ordre inversé
	assert.NoError(t, err)
	assert.Equal(t, conv.ID, conv2.ID) // Même conversation
}

func TestConversationService_IsParticipant(t *testing.T) {
	db := setupTestDB()
	service := NewConversationService(db)

	// Créer conversation test
	conv := internalmodels.Conversation{
		Participant1ID: "user-1",
		Participant2ID: "user-2",
	}
	db.Create(&conv)

	// Test participant valide
	isParticipant, err := service.IsParticipant(conv.ID, "user-1")
	assert.NoError(t, err)
	assert.True(t, isParticipant)

	// Test non-participant
	isParticipant, err = service.IsParticipant(conv.ID, "user-3")
	assert.NoError(t, err)
	assert.False(t, isParticipant)
}
