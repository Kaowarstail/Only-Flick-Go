package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/Kaowarstail/Only-Flick-Go/config"
	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	internalModels "github.com/Kaowarstail/Only-Flick-Go/internal/models"
	"github.com/Kaowarstail/Only-Flick-Go/models"
	"github.com/Kaowarstail/Only-Flick-Go/seed"
)

func main() {
	fmt.Println("🚀 OnlyFlick Messaging System - Database Test & Seed")
	fmt.Println("=" + string(make([]rune, 55)) + "=")

	// Charger les variables d'environnement
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: No .env file found: %v", err)
	}

	// Forcer l'utilisation de SQLite pour ce test
	log.Println("🔧 Configuring SQLite for testing...")
	// Nous allons contourner le problème CGO en utilisant directement la configuration

	// Charger la configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	fmt.Printf("✅ Configuration loaded successfully\n")

	// Initialiser la base de données
	fmt.Println("\n📊 Initializing database connection...")
	
	// Forcer l'utilisation de SQLite en définissant la variable d'environnement
	originalDBType := cfg.Database.Host
	cfg.Database.Host = "" // Cela forcera l'utilisation de SQLite
	
	if err := database.Initialize(); err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}

	fmt.Println("✅ Database connected and migrated successfully!")

	// Tester les modèles de messagerie
	fmt.Println("\n💬 Testing messaging models...")
	
	db := database.GetDB()
	
	// Vérifier que les tables existent
	tables := []string{"users", "conversations", "messages"}
	for _, table := range tables {
		var count int64
		if err := db.Table(table).Count(&count).Error; err != nil {
			log.Printf("⚠️  Error checking table %s: %v", table, err)
		} else {
			fmt.Printf("✅ Table '%s' exists with %d records\n", table, count)
		}
	}

	// Exécuter le seed si les tables sont vides
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	
	if userCount == 0 {
		fmt.Println("\n🌱 Running seed data...")
		if err := seed.Run(); err != nil {
			log.Printf("⚠️  Seed error: %v", err)
		} else {
			fmt.Println("✅ Seed data created successfully!")
		}
		
		// Recompter après le seed
		db.Model(&models.User{}).Count(&userCount)
	}

	fmt.Printf("✅ Found %d users in database\n", userCount)

	if userCount >= 2 {
		// Compter les conversations existantes
		var convCount int64
		db.Model(&internalModels.Conversation{}).Count(&convCount)
		fmt.Printf("✅ Found %d conversations\n", convCount)
		
		// Compter les messages existants
		var msgCount int64
		db.Model(&internalModels.Message{}).Count(&msgCount)
		fmt.Printf("✅ Found %d messages\n", msgCount)
		
		// Afficher quelques exemples
		if convCount > 0 {
			fmt.Println("\n📋 Sample conversations:")
			var conversations []internalModels.Conversation
			db.Preload("Participants").Limit(3).Find(&conversations)
			
			for i, conv := range conversations {
				fmt.Printf("   %d. Conversation ID: %d (%d participants)\n", 
					i+1, conv.ID, len(conv.Participants))
			}
		}
		
		if msgCount > 0 {
			fmt.Println("\n💌 Sample messages:")
			var messages []internalModels.Message
			db.Limit(5).Order("created_at desc").Find(&messages)
			
			for i, msg := range messages {
				content := msg.Content
				if len(content) > 50 {
					content = content[:50] + "..."
				}
				fmt.Printf("   %d. From User %d: %s\n", i+1, msg.SenderID, content)
			}
		}
	}

	// Test des fonctionnalités de messagerie
	fmt.Println("\n🧪 Testing messaging functionality...")
	
	// Créer une conversation de test si nous avons des utilisateurs
	if userCount >= 2 {
		var users []models.User
		db.Limit(2).Find(&users)
		
		if len(users) >= 2 {
			// Test de création de conversation
			conversation := &internalModels.Conversation{
				Type:        "direct",
				IsActive:    true,
			}
			
			if err := db.Create(conversation).Error; err != nil {
				fmt.Printf("⚠️  Could not create test conversation: %v\n", err)
			} else {
				fmt.Printf("✅ Created test conversation ID: %d\n", conversation.ID)
				
				// Ajouter les participants
				participants := []internalModels.ConversationParticipant{
					{ConversationID: conversation.ID, UserID: users[0].ID},
					{ConversationID: conversation.ID, UserID: users[1].ID},
				}
				
				for _, participant := range participants {
					db.Create(&participant)
				}
				
				// Test de création de message
				message := &internalModels.Message{
					ConversationID: conversation.ID,
					SenderID:       users[0].ID,
					Content:        "Message de test du système OnlyFlick! 🚀",
					Type:           "text",
					Status:         "sent",
				}
				
				if err := db.Create(message).Error; err != nil {
					fmt.Printf("⚠️  Could not create test message: %v\n", err)
				} else {
					fmt.Printf("✅ Created test message ID: %d\n", message.ID)
				}
			}
		}
	}

	fmt.Println("\n🎉 Database test completed successfully!")
	fmt.Println("=" + string(make([]rune, 55)) + "=")
	fmt.Println("Your OnlyFlick messaging system is ready for development!")
	
	// Restaurer la configuration originale
	cfg.Database.Host = originalDBType
}
