package main

import (
	"log"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/models"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Initialize database connection
	err := database.Initialize()
	if err != nil {
		log.Fatal("Erreur d'initialisation de la base de données:", err)
	}
	db := database.GetDB()

	// Hash the admin password
	password := "admin123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Erreur lors du hachage du mot de passe:", err)
	}

	// Update the admin user password
	var adminUser models.User
	result := db.Where("username = ? OR email = ?", "admin", "admin@onlyflick.com").First(&adminUser)
	if result.Error != nil {
		log.Fatal("Utilisateur admin non trouvé:", result.Error)
	}

	// Update password
	result = db.Model(&adminUser).Update("password", string(hashedPassword))
	if result.Error != nil {
		log.Fatal("Erreur lors de la mise à jour du mot de passe:", result.Error)
	}

	log.Printf("✅ Mot de passe de l'utilisateur admin (ID: %s) mis à jour avec succès", adminUser.ID)
}
