package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

// GetUsers récupère tous les utilisateurs
func GetUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	result := database.GetDB().Find(&users)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des utilisateurs")
		return
	}

	// Convertir les utilisateurs en réponses
	var responses []models.UserResponse
	for _, user := range users {
		responses = append(responses, user.ToResponse())
	}

	respondWithJSON(w, http.StatusOK, responses)
}

// GetUser récupère un utilisateur par son ID
func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID utilisateur invalide")
		return
	}

	var user models.User
	result := database.GetDB().First(&user, id)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	respondWithJSON(w, http.StatusOK, user.ToResponse())
}

// CreateUser crée un nouvel utilisateur
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données d'utilisateur invalides")
		return
	}
	defer r.Body.Close()

	// Hachage du mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors du hachage du mot de passe")
		return
	}
	user.Password = string(hashedPassword)

	// Création de l'utilisateur dans la base de données
	result := database.GetDB().Create(&user)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la création de l'utilisateur")
		return
	}

	respondWithJSON(w, http.StatusCreated, user.ToResponse())
}

// UpdateUser met à jour un utilisateur existant
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID utilisateur invalide")
		return
	}

	var user models.User
	result := database.GetDB().First(&user, id)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Décode les nouvelles valeurs
	var updatedUser models.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updatedUser); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données d'utilisateur invalides")
		return
	}
	defer r.Body.Close()

	// Mise à jour des champs
	user.Username = updatedUser.Username
	user.Email = updatedUser.Email
	user.FirstName = updatedUser.FirstName
	user.LastName = updatedUser.LastName

	// Si un nouveau mot de passe est fourni, le hacher
	if updatedUser.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updatedUser.Password), bcrypt.DefaultCost)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Erreur lors du hachage du mot de passe")
			return
		}
		user.Password = string(hashedPassword)
	}

	// Sauvegarde des modifications
	database.GetDB().Save(&user)

	respondWithJSON(w, http.StatusOK, user.ToResponse())
}

// DeleteUser supprime un utilisateur
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID utilisateur invalide")
		return
	}

	var user models.User
	result := database.GetDB().First(&user, id)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	database.GetDB().Delete(&user)

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Utilisateur supprimé avec succès"})
}

// Fonctions utilitaires pour les réponses HTTP
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
