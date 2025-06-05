package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

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

	// Vérifier si l'utilisateur connecté a le droit de modifier cet utilisateur
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || userID != user.ID {
		// Vérifier si c'est un admin
		userRole, _ := r.Context().Value(middleware.UserRoleKey).(string)
		if userRole != string(models.RoleAdmin) {
			respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à modifier cet utilisateur")
			return
		}
	}

	// Structure pour contenir les données à mettre à jour
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

	// Mettre à jour uniquement les champs fournis
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

	// Enregistrer les modifications
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

	// Vérifier si l'utilisateur connecté a le droit de supprimer cet utilisateur
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || userID != user.ID {
		// Vérifier si c'est un admin
		userRole, _ := r.Context().Value(middleware.UserRoleKey).(string)
		if userRole != string(models.RoleAdmin) {
			respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à supprimer cet utilisateur")
			return
		}
	}

	// Supprimer l'utilisateur (soft delete si GORM est configuré avec DeletedAt)
	result = database.GetDB().Delete(&user)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la suppression de l'utilisateur")
		return
	}

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
