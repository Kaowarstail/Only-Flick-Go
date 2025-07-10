package database

import (
	"github.com/Kaowarstail/Only-Flick-Go/internal/models"
	"gorm.io/gorm"
)

// RunMessagingMigrations effectue les migrations pour le système de messagerie
func RunMessagingMigrations(db *gorm.DB) error {
	// Auto-migrate les modèles de messagerie classique uniquement
	err := db.AutoMigrate(
		&models.ConversationClassic{},
		&models.MessageClassic{},
		&models.ConversationClassicReadStatus{},
		&models.MessageClassicReaction{},
	)
	if err != nil {
		return err
	}

	// Créer les index personnalisés pour les nouveaux models
	if err := createMessagingClassicIndexes(db); err != nil {
		return err
	}

	return nil
}

// createMessagingClassicIndexes crée les index personnalisés pour les models classiques
func createMessagingClassicIndexes(db *gorm.DB) error {
	indexes := []string{
		// Index unique pour conversation_classic_read_statuses
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_conversation_classic_read_status_unique 
		 ON conversation_classic_read_statuses(conversation_id, user_id) 
		 WHERE deleted_at IS NULL`,

		// Index unique pour message_classic_reactions
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_message_classic_reaction_unique 
		 ON message_classic_reactions(message_id, user_id) 
		 WHERE deleted_at IS NULL`,

		// Index composite pour performance
		`CREATE INDEX IF NOT EXISTS idx_message_classics_conversation_created 
		 ON message_classics(conversation_id, created_at DESC) 
		 WHERE deleted_at IS NULL`,

		// Index pour comptage non lus
		`CREATE INDEX IF NOT EXISTS idx_message_classics_unread 
		 ON message_classics(conversation_id, sender_id, created_at) 
		 WHERE deleted_at IS NULL`,
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			// Log mais ne pas échouer si index existe déjà
			continue
		}
	}

	return nil
}
