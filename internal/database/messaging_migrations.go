package database

import (
	"fmt"

	"github.com/Kaowarstail/Only-Flick-Go/internal/models"
	"gorm.io/gorm"
)

// RunMessagingMigrations exécute les migrations pour le système de messagerie
func RunMessagingMigrations(db *gorm.DB) error {	// Auto-migrate les nouveaux models
	err := db.AutoMigrate(
		&models.Conversation{},
		&models.Message{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate messaging models: %w", err)
	}

	// Créer index unique pour conversations
	conv := models.Conversation{}
	if err := conv.CreateIndexes(db); err != nil {
		return fmt.Errorf("failed to create conversation indexes: %w", err)
	}

	// Index additionnels pour performance
	if err := createMessagingIndexes(db); err != nil {
		return fmt.Errorf("failed to create messaging indexes: %w", err)
	}

	return nil
}

// createMessagingIndexes crée les index pour optimiser les performances
func createMessagingIndexes(db *gorm.DB) error {
	indexes := []string{
		// Index pour récupération rapide des messages d'une conversation
		`CREATE INDEX IF NOT EXISTS idx_messages_conversation_created 
		 ON messages(conversation_id, created_at DESC)`,

		// Index pour messages non lus par conversation
		`CREATE INDEX IF NOT EXISTS idx_messages_unread 
		 ON messages(conversation_id, sender_id) WHERE read_at IS NULL`,

		// Index pour recherche dans le contenu des messages (PostgreSQL)
		`CREATE INDEX IF NOT EXISTS idx_messages_content_search 
		 ON messages USING gin(to_tsvector('french', content)) 
		 WHERE content IS NOT NULL`,

		// Index pour conversations actives d'un utilisateur
		`CREATE INDEX IF NOT EXISTS idx_conversations_user_active 
		 ON conversations(participant_1_id, updated_at DESC) 
		 WHERE is_active = true`,

		// Index pour conversations actives de l'autre utilisateur
		`CREATE INDEX IF NOT EXISTS idx_conversations_user2_active 
		 ON conversations(participant_2_id, updated_at DESC) 
		 WHERE is_active = true`,

		// Index pour messages par type et date
		`CREATE INDEX IF NOT EXISTS idx_messages_type_date 
		 ON messages(message_type, created_at DESC)`,

		// Index pour statut des messages
		`CREATE INDEX IF NOT EXISTS idx_messages_status 
		 ON messages(status, created_at DESC)`,

		// Index pour média messages
		`CREATE INDEX IF NOT EXISTS idx_messages_media 
		 ON messages(message_type, media_url) 
		 WHERE media_url IS NOT NULL`,
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			// Ignorer les erreurs pour les index qui existent déjà ou syntax SQL spécifique
			fmt.Printf("Warning: Failed to create index: %v\n", err)
		}
	}

	return nil
}

// createSQLiteIndexes crée des index compatibles SQLite
func createSQLiteIndexes(db *gorm.DB) error {
	indexes := []string{
		// Index pour récupération rapide des messages d'une conversation
		`CREATE INDEX IF NOT EXISTS idx_messages_conversation_created 
		 ON messages(conversation_id, created_at DESC)`,

		// Index pour messages non lus
		`CREATE INDEX IF NOT EXISTS idx_messages_unread 
		 ON messages(conversation_id, sender_id, read_at)`,

		// Index pour conversations actives d'un utilisateur
		`CREATE INDEX IF NOT EXISTS idx_conversations_user_active 
		 ON conversations(participant_1_id, updated_at DESC, is_active)`,

		// Index pour conversations actives de l'autre utilisateur
		`CREATE INDEX IF NOT EXISTS idx_conversations_user2_active 
		 ON conversations(participant_2_id, updated_at DESC, is_active)`,

		// Index pour messages par type
		`CREATE INDEX IF NOT EXISTS idx_messages_type_date 
		 ON messages(message_type, created_at DESC)`,

		// Index pour statut des messages
		`CREATE INDEX IF NOT EXISTS idx_messages_status 
		 ON messages(status, created_at DESC)`,

		// Index pour recherche basique dans le contenu (SQLite)
		`CREATE INDEX IF NOT EXISTS idx_messages_content 
		 ON messages(content)`,
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			fmt.Printf("Warning: Failed to create SQLite index: %v\n", err)
		}
	}

	return nil
}

// RollbackMessagingMigrations annule les migrations messagerie (pour dev/test)
func RollbackMessagingMigrations(db *gorm.DB) error {
	return db.Migrator().DropTable(
		&models.Message{},
		&models.Conversation{},
	)
}

// InitializeMessagingSchema initialise le schéma de messagerie complet
func InitializeMessagingSchema(db *gorm.DB) error {
	// Exécuter les migrations
	if err := RunMessagingMigrations(db); err != nil {
		return fmt.Errorf("failed to run messaging migrations: %w", err)
	}

	// Créer des triggers pour maintenir la cohérence (si supporté)
	if err := createMessagingTriggers(db); err != nil {
		fmt.Printf("Warning: Failed to create messaging triggers: %v\n", err)
	}

	return nil
}

// createMessagingTriggers crée des triggers pour maintenir la cohérence des données
func createMessagingTriggers(db *gorm.DB) error {
	triggers := []string{
		// Trigger pour mettre à jour last_message_id dans conversations
		`CREATE OR REPLACE FUNCTION update_conversation_last_message()
		RETURNS TRIGGER AS $$
		BEGIN
			UPDATE conversations 
			SET last_message_id = NEW.id, updated_at = NEW.created_at
			WHERE id = NEW.conversation_id;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;`,

		`CREATE TRIGGER trigger_update_conversation_last_message
		AFTER INSERT ON messages
		FOR EACH ROW
		EXECUTE FUNCTION update_conversation_last_message();`,
	}

	// Ces triggers sont spécifiques à PostgreSQL
	for _, triggerSQL := range triggers {
		if err := db.Exec(triggerSQL).Error; err != nil {
			// Ignorer les erreurs pour d'autres SGBD
			fmt.Printf("Warning: Trigger creation failed (might not be PostgreSQL): %v\n", err)
		}
	}

	return nil
}

// GetMessagingStats retourne les statistiques du système de messagerie
func GetMessagingStats(db *gorm.DB) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Nombre total de conversations
	var totalConversations int64
	if err := db.Model(&models.Conversation{}).Count(&totalConversations).Error; err != nil {
		return nil, err
	}
	stats["total_conversations"] = totalConversations

	// Nombre de conversations actives
	var activeConversations int64
	if err := db.Model(&models.Conversation{}).Where("is_active = ?", true).Count(&activeConversations).Error; err != nil {
		return nil, err
	}
	stats["active_conversations"] = activeConversations

	// Nombre total de messages
	var totalMessages int64
	if err := db.Model(&models.Message{}).Count(&totalMessages).Error; err != nil {
		return nil, err
	}
	stats["total_messages"] = totalMessages

	// Nombre de messages non lus
	var unreadMessages int64
	if err := db.Model(&models.Message{}).Where("read_at IS NULL").Count(&unreadMessages).Error; err != nil {
		return nil, err
	}
	stats["unread_messages"] = unreadMessages

	return stats, nil
}
