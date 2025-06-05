package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

// BecomeCreatorRequest représente les données de demande pour devenir créateur
type BecomeCreatorRequest struct {
	Biography  string   `json:"biography"`
	Categories []string `json:"categories"`
}

// GetCreators récupère tous les créateurs
func GetCreators(w http.ResponseWriter, r *http.Request) {
	var users []models.User

	// Pagination
	page := 1
	pageSize := 10
	offset := (page - 1) * pageSize

	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
			offset = (page - 1) * pageSize
		}
	}

	if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
		if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 && s <= 100 {
			pageSize = s
			offset = (page - 1) * pageSize
		}
	}

	// Récupérer les créateurs de la base de données
	result := database.GetDB().Where("role = ?", models.RoleCreator).
		Offset(offset).
		Limit(pageSize).
		Find(&users)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des créateurs")
		return
	}

	// Compter le nombre total de créateurs pour la pagination
	var total int64
	database.GetDB().Model(&models.User{}).Where("role = ?", models.RoleCreator).Count(&total)

	// Convertir les utilisateurs en réponses
	var creators []models.UserResponse
	for _, user := range users {
		creators = append(creators, user.ToResponse())
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"creators": creators,
		"pagination": map[string]interface{}{
			"page":  page,
			"size":  pageSize,
			"total": total,
			"pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetCreator récupère un créateur par son ID
func GetCreator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID créateur invalide")
		return
	}

	var user models.User
	result := database.GetDB().Where("id = ? AND role = ?", id, models.RoleCreator).First(&user)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Créateur non trouvé")
		return
	}

	respondWithJSON(w, http.StatusOK, user.ToResponse())
}

// GetFeaturedCreators récupère les créateurs mis en avant
func GetFeaturedCreators(w http.ResponseWriter, r *http.Request) {
	var users []models.User

	// Limite par défaut pour les créateurs en vedette
	limit := 5
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 20 {
			limit = l
		}
	}

	// Pour une implémentation réelle, nous aurions besoin d'un champ "featured" dans le modèle User
	// ou d'une table séparée pour les créateurs en vedette.
	// Pour cet exemple, nous allons simplement renvoyer les N premiers créateurs.
	result := database.GetDB().
		Where("role = ?", models.RoleCreator).
		Limit(limit).
		Order("id DESC"). // Dans un cas réel, nous pourrions ordonner par popularité ou autre métrique
		Find(&users)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des créateurs en vedette")
		return
	}

	// Convertir les utilisateurs en réponses
	var creators []models.UserResponse
	for _, user := range users {
		creators = append(creators, user.ToResponse())
	}

	respondWithJSON(w, http.StatusOK, creators)
}

// SearchCreators recherche des créateurs
func SearchCreators(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondWithError(w, http.StatusBadRequest, "Paramètre de recherche 'q' requis")
		return
	}

	// Pagination
	page := 1
	pageSize := 10
	offset := (page - 1) * pageSize

	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
			offset = (page - 1) * pageSize
		}
	}

	if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
		if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 && s <= 100 {
			pageSize = s
			offset = (page - 1) * pageSize
		}
	}

	var users []models.User

	// Recherche dans les champs username, first_name, last_name, biography
	result := database.GetDB().
		Where("role = ? AND (username LIKE ? OR first_name LIKE ? OR last_name LIKE ? OR biography LIKE ?)",
			models.RoleCreator,
			"%"+query+"%",
			"%"+query+"%",
			"%"+query+"%",
			"%"+query+"%").
		Offset(offset).
		Limit(pageSize).
		Find(&users)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la recherche de créateurs")
		return
	}

	// Compter le nombre total de résultats pour la pagination
	var total int64
	database.GetDB().Model(&models.User{}).
		Where("role = ? AND (username LIKE ? OR first_name LIKE ? OR last_name LIKE ? OR biography LIKE ?)",
			models.RoleCreator,
			"%"+query+"%",
			"%"+query+"%",
			"%"+query+"%",
			"%"+query+"%").
		Count(&total)

	// Convertir les utilisateurs en réponses
	var creators []models.UserResponse
	for _, user := range users {
		creators = append(creators, user.ToResponse())
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"creators": creators,
		"pagination": map[string]interface{}{
			"page":  page,
			"size":  pageSize,
			"total": total,
			"pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// BecomeCreator permet à un utilisateur de devenir créateur
func BecomeCreator(w http.ResponseWriter, r *http.Request) {
	// Extraire l'ID utilisateur du contexte
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	result := database.GetDB().First(&user, userID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Vérifier si l'utilisateur est déjà un créateur
	if user.Role == models.RoleCreator {
		respondWithError(w, http.StatusConflict, "Vous êtes déjà un créateur")
		return
	}

	// Décoder la requête
	var req BecomeCreatorRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données de demande invalides")
		return
	}
	defer r.Body.Close()

	// Mettre à jour le rôle et la biographie de l'utilisateur
	user.Role = models.RoleCreator
	if req.Biography != "" {
		user.Biography = req.Biography
	}

	// Sauvegarder les modifications
	result = database.GetDB().Save(&user)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour du statut de créateur")
		return
	}

	// TODO: Gérer les catégories du créateur (nécessite un modèle supplémentaire)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Vous êtes maintenant un créateur",
		"user":    user.ToResponse(),
	})
}

// UpdateCreator met à jour les informations d'un créateur
func UpdateCreator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID créateur invalide")
		return
	}

	// Extraire l'ID utilisateur du contexte
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Vérifier si l'utilisateur authentifié est bien le créateur
	if userID != uint(id) {
		// Vérifier si c'est un admin
		userRole, _ := r.Context().Value(middleware.UserRoleKey).(string)
		if userRole != string(models.RoleAdmin) {
			respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à modifier ce profil de créateur")
			return
		}
	}

	// Récupérer le créateur
	var user models.User
	result := database.GetDB().Where("id = ? AND role = ?", id, models.RoleCreator).First(&user)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Créateur non trouvé")
		return
	}

	// Structure pour contenir les données à mettre à jour
	var updateData struct {
		Biography  *string  `json:"biography"`
		Categories []string `json:"categories,omitempty"`
	}

	// Décoder la requête
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updateData); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données de mise à jour invalides")
		return
	}
	defer r.Body.Close()

	// Mettre à jour la biographie si fournie
	if updateData.Biography != nil {
		user.Biography = *updateData.Biography
	}

	// Sauvegarder les modifications
	result = database.GetDB().Save(&user)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour du profil créateur")
		return
	}

	// TODO: Gérer les catégories du créateur (nécessite un modèle supplémentaire)

	respondWithJSON(w, http.StatusOK, user.ToResponse())
}
