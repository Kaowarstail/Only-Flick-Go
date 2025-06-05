package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Kaowarstail/Only-Flick-Go/config"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

var DB *gorm.DB

// Initialize initialise la connexion à la base de données et effectue les migrations
func Initialize() error {
	cfg := config.Get()

	// Construction de la chaîne de connexion
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode)

	// Configuration de GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Connexion à la base de données
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Connected to database successfully")

	// Migration automatique des schémas
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
	return nil
}

// GetDB retourne l'instance de la connexion à la base de données
func GetDB() *gorm.DB {
	return DB
}
