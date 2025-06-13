package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
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

	var responses []models.UserResponse
	for _, user := range users {
		responses = append(responses, user.ToResponse())
	}

	respondWithJSON(w, http.StatusOK, responses)
}

// GetUser récupère un utilisateur par son ID
func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var user models.User
	result := database.GetDB().First(&user, "id = ?", id)
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors du hachage du mot de passe")
		return
	}
	user.Password = string(hashedPassword)

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
	id := vars["id"]

	var user models.User
	result := database.GetDB().First(&user, "id = ?", id)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	if userID != user.ID {
		userRole, _ := r.Context().Value(middleware.UserRoleKey).(string)
		if userRole != string(models.RoleAdmin) {
			respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à modifier cet utilisateur")
			return
		}
	}

	var updateData struct {
		FirstName      *string `json:"first_name"`
		LastName       *string `json:"last_name"`
		Email          *string `json:"email"`
		Biography      *string `json:"biography"`
		ProfilePicture *string `json:"profile_picture"`
		Password       *string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updateData); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}
	defer r.Body.Close()

	if updateData.FirstName != nil {
		user.FirstName = *updateData.FirstName
	}
	if updateData.LastName != nil {
		user.LastName = *updateData.LastName
	}
	if updateData.Email != nil {
		user.Email = *updateData.Email
	}
	if updateData.Biography != nil {
		user.Biography = *updateData.Biography
	}
	if updateData.ProfilePicture != nil {
		user.ProfilePicture = *updateData.ProfilePicture
	}
	if updateData.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*updateData.Password), bcrypt.DefaultCost)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Erreur lors du hachage du mot de passe")
			return
		}
		user.Password = string(hashedPassword)
	}

	result = database.GetDB().Save(&user)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour de l'utilisateur")
		return
	}

	respondWithJSON(w, http.StatusOK, user.ToResponse())
}

// DeleteUser supprime un utilisateur
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var user models.User
	result := database.GetDB().First(&user, "id = ?", id)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	if userID != user.ID {
		userRole, _ := r.Context().Value(middleware.UserRoleKey).(string)
		if userRole != string(models.RoleAdmin) {
			respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à supprimer cet utilisateur")
			return
		}
	}

	result = database.GetDB().Delete(&user)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la suppression de l'utilisateur")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Utilisateur supprimé avec succès"})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
