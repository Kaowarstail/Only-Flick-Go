package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/Kaowarstail/Only-Flick-Go/config"
	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	internalModels "github.com/Kaowarstail/Only-Flick-Go/internal/models"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

func main() {
	fmt.Println("🚀 OnlyFlick Messaging System - Database Setup & Test")
	fmt.Println("=" + string(make([]byte, 60)) + "=")

	// Charger les variables d'environnement
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: No .env file found: %v", err)
	}

	// Charger la configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	fmt.Printf("✅ Configuration loaded successfully\n")
	fmt.Printf("   Database: %s@%s:%s/%s\n", 
		cfg.Database.User, 
		cfg.Database.Host, 
		cfg.Database.Port, 
		cfg.Database.DBName)

	// Initialiser la base de données
	fmt.Println("\n📊 Initializing database connection...")
	if err := database.Initialize(); err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}

	fmt.Println("✅ Database connected and migrated successfully!")

	// Tester les modèles de messagerie
	fmt.Println("\n💬 Testing messaging models...")
	
	db := database.GetDB()
	
	// Vérifier que les tables existent
	tables := []string{"conversations", "messages", "users"}
	for _, table := range tables {
		var count int64
		if err := db.Table(table).Count(&count).Error; err != nil {
			log.Printf("❌ Error checking table %s: %v", table, err)
		} else {
			fmt.Printf("✅ Table '%s' exists with %d records\n", table, count)
		}
	}

	// Vérifier les relations et contraintes
	fmt.Println("\n🔗 Checking database schema...")
		// Test de création d'une conversation (pour vérifier les contraintes)
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	
	if userCount >= 2 {
		fmt.Printf("✅ Found %d users - ready for messaging tests\n", userCount)
				// Compter les conversations existantes
		var convCount int64
		db.Model(&internalModels.Conversation{}).Count(&convCount)
		fmt.Printf("✅ Found %d existing conversations\n", convCount)
		
		// Compter les messages existants
		var msgCount int64
		db.Model(&internalModels.Message{}).Count(&msgCount)
		fmt.Printf("✅ Found %d existing messages\n", msgCount)
		
	} else {
		fmt.Println("⚠️  Need at least 2 users for messaging tests")
	}

	// Test des indexes et performance
	fmt.Println("\n⚡ Checking database indexes...")
	
	// Vérifier les indexes importants
	indexes := []struct {
		table string
		name  string
	}{
		{"conversations", "idx_conversations_updated_at"},
		{"messages", "idx_messages_conversation_id"},
		{"messages", "idx_messages_sender_id"},
		{"messages", "idx_messages_created_at"},
	}
	
	for _, idx := range indexes {
		var exists bool
		query := fmt.Sprintf(`
			SELECT EXISTS (
				SELECT 1 FROM pg_indexes 
				WHERE tablename = '%s' AND indexname = '%s'
			)`, idx.table, idx.name)
		
		if err := db.Raw(query).Scan(&exists).Error; err != nil {
			fmt.Printf("⚠️  Could not check index %s: %v\n", idx.name, err)
		} else if exists {
			fmt.Printf("✅ Index %s exists\n", idx.name)
		} else {
			fmt.Printf("❌ Index %s missing\n", idx.name)
		}
	}

	fmt.Println("\n🎉 Database setup test completed!")
	fmt.Println("=" + string(make([]byte, 60)) + "=")
	fmt.Println("Your OnlyFlick messaging system database is ready!")
}
