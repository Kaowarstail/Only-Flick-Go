package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Kaowarstail/Only-Flick-Go/config"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

var DB *gorm.DB

// Initialize initialise la connexion à la base de données et effectue les migrations
func Initialize() error {
	cfg := config.Get()

	// Configuration de GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	var err error

	// Vérifier si on utilise SQLite en développement
	if os.Getenv("DB_TYPE") == "sqlite" && os.Getenv("ENV") != "production" {
		// Utiliser SQLite pour le développement
		dbPath := "dev_database.db"
		log.Println("Using SQLite database:", dbPath)
		DB, err = gorm.Open(sqlite.Open(dbPath), gormConfig)
	} else {
		// Utiliser PostgreSQL pour la production
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.DBName,
			cfg.Database.SSLMode)

		log.Println("Using PostgreSQL database")
		DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("✅ Connexion à la base de données réussie")

	// Seulement faire les migrations si pas en production
	if os.Getenv("ENV") != "production" {
		log.Println("Running database migrations...")
		err = DB.AutoMigrate(
			&models.User{},
			&models.CreatorProfile{},
			&models.Content{},
			&models.SubscriptionPlan{},
			&models.Subscription{},
			&models.Comment{},
			&models.Like{},
			&models.Report{},
			&models.Message{},
			&models.Notification{},
			&models.Transaction{},
			&models.Payout{},
		)
		if err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
		log.Println("Database migrations completed successfully")
	} else {
		log.Println("🚀 Production mode: skipping migrations (tables already exist)")
	}

	return nil
}

// GetDB retourne l'instance de la connexion à la base de données
func GetDB() *gorm.DB {
	return DB
}
