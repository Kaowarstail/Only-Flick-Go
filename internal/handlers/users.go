package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
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

// getUserIDFromContext récupère l'ID utilisateur depuis le contexte
func getUserIDFromContext(ctx context.Context) string {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return ""
	}
	return userID
}

// UpdateUserProfile met à jour le profil d'un utilisateur avec les nouvelles données
func UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	var updateData struct {
		FirstName      *string                `json:"first_name"`
		LastName       *string                `json:"last_name"`
		Email          *string                `json:"email"`
		Biography      *string                `json:"biography"`
		ProfilePicture *string                `json:"profile_picture"`
		CoverPicture   *string                `json:"cover_picture"`
		Location       *string                `json:"location"`
		Website        *string                `json:"website"`
		Birthday       *string                `json:"birthday"`
		Gender         *string                `json:"gender"`
		SocialLinks    map[string]interface{} `json:"social_links"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updateData); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}
	defer r.Body.Close()

	// Récupérer l'utilisateur actuel
	var user models.User
	if err := database.GetDB().First(&user, "id = ?", userID).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Mise à jour des champs de base
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
	if updateData.CoverPicture != nil {
		user.CoverPicture = updateData.CoverPicture
	}
	if updateData.Location != nil {
		user.Location = updateData.Location
	}
	if updateData.Website != nil {
		user.Website = updateData.Website
	}
	if updateData.Birthday != nil {
		user.Birthday = updateData.Birthday
	}
	if updateData.Gender != nil {
		user.Gender = updateData.Gender
	}

	// Sauvegarder les modifications
	if err := database.GetDB().Save(&user).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour du profil")
		return
	}

	// Gérer les liens sociaux
	if updateData.SocialLinks != nil {
		var socialLinks models.SocialLinks
		if err := database.GetDB().First(&socialLinks, "user_id = ?", userID).Error; err != nil {
			// Créer si n'existe pas
			socialLinks = models.SocialLinks{
				UserID: userID,
				Links:  updateData.SocialLinks,
			}
			database.GetDB().Create(&socialLinks)
		} else {
			// Mettre à jour
			socialLinks.Links = updateData.SocialLinks
			database.GetDB().Save(&socialLinks)
		}
	}

	response := models.APIResponse{
		Success: true,
		Data:    user.ToResponse(),
		Message: "Profil mis à jour avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetUserProfile récupère le profil complet d'un utilisateur
func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var user models.User
	if err := database.GetDB().First(&user, "id = ?", userID).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Récupérer les liens sociaux
	var socialLinks models.SocialLinks
	database.GetDB().First(&socialLinks, "user_id = ?", userID)

	// Récupérer les statistiques
	var stats models.UserStats
	database.GetDB().First(&stats, "user_id = ?", userID)

	profile := map[string]interface{}{
		"user":         user.ToResponse(),
		"social_links": socialLinks.Links,
		"stats":        stats,
	}

	response := models.APIResponse{
		Success: true,
		Data:    profile,
		Message: "Profil récupéré avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// UpdateUserSocialLinks met à jour les liens sociaux d'un utilisateur
func UpdateUserSocialLinks(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	var request struct {
		Links map[string]interface{} `json:"links"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}
	defer r.Body.Close()

	var socialLinks models.SocialLinks
	if err := database.GetDB().First(&socialLinks, "user_id = ?", userID).Error; err != nil {
		// Créer si n'existe pas
		socialLinks = models.SocialLinks{
			UserID: userID,
			Links:  request.Links,
		}
		if err := database.GetDB().Create(&socialLinks).Error; err != nil {
			respondWithError(w, http.StatusInternalServerError, "Erreur lors de la création des liens sociaux")
			return
		}
	} else {
		// Mettre à jour
		socialLinks.Links = request.Links
		if err := database.GetDB().Save(&socialLinks).Error; err != nil {
			respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour des liens sociaux")
			return
		}
	}

	response := models.APIResponse{
		Success: true,
		Data:    socialLinks,
		Message: "Liens sociaux mis à jour avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetUserStats récupère les statistiques d'un utilisateur
func GetUserStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var stats models.UserStats
	if err := database.GetDB().First(&stats, "user_id = ?", userID).Error; err != nil {
		// Créer des statistiques par défaut si elles n'existent pas
		stats = models.UserStats{
			UserID: userID,
		}
		database.GetDB().Create(&stats)
	}

	response := models.APIResponse{
		Success: true,
		Data:    stats,
		Message: "Statistiques récupérées avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}
