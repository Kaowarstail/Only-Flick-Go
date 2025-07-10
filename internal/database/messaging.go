package database

import (
	"github.com/Kaowarstail/Only-Flick-Go/internal/models"
	"gorm.io/gorm"
)

// RunMessagingMigrations effectue les migrations pour le système de messagerie
func RunMessagingMigrations(db *gorm.DB) error {
	// Auto-migrate les nouveaux models
	err := db.AutoMigrate(
		&models.Conversation{},
		&models.EnhancedMessage{},
		&models.PaidMessageTransaction{},
	)
	if err != nil {
		return err
	}
	// Créer les index personnalisés
	conv := models.Conversation{}
	if err := conv.CreateIndexes(db); err != nil {
		return err
	}

	msg := models.EnhancedMessage{}
	if err := msg.CreateIndexes(db); err != nil {
		return err
	}

	return nil
}
