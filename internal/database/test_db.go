package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDB *gorm.DB

// InitTestDB initialise une base de données SQLite en mémoire pour les tests
func InitTestDB() {
	var err error

	// Utiliser SQLite en mémoire pour les tests
	testDB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Pas de logs pendant les tests
	})

	if err != nil {
		panic("Failed to create test database: " + err.Error())
	}

	// Remplacer la DB globale par la DB de test
	DB = testDB
}

// CloseDB ferme la connexion à la base de données de test
func CloseDB() {
	if testDB != nil {
		sqlDB, err := testDB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// CleanTestDB nettoie la base de données de test
func CleanTestDB() {
	if testDB != nil {
		// Supprimer toutes les tables pour nettoyer
		testDB.Migrator().DropTable(
			"users", "creator_profiles", "contents", "subscription_plans",
			"subscriptions", "comments", "likes", "reports", "messages",
			"notifications", "transactions", "payouts", "password_reset_tokens",
			"email_verification_tokens", "jwt_blacklist_tokens",
		)
	}
}
