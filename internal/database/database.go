package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

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

	// Connexion à la base de données
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Connected to database successfully")

	// Migration automatique des schémas
	log.Println("Running database migrations...")
	err = DB.AutoMigrate(&models.User{})
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// GetDB retourne l'instance de la connexion à la base de données
func GetDB() *gorm.DB {
	return DB
}
