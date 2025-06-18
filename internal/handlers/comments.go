package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

// CreateCommentRequest représente les données pour créer un commentaire
type CreateCommentRequest struct {
	Content string `json:"content"`
}

// UpdateCommentRequest représente les données pour mettre à jour un commentaire
type UpdateCommentRequest struct {
	Content string `json:"content"`
}

// GetContentComments récupère tous les commentaires d'un contenu
func GetContentComments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentID := vars["id"]

	// Vérifier que le contenu existe et est accessible
	var content models.Content
	result := database.GetDB().First(&content, "id = ?", contentID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	// Vérifier si le contenu est publié (sauf si c'est le créateur)
	if !content.IsPublished || content.IsFlagged {
		userID, _ := r.Context().Value(middleware.UserIDKey).(string)
		if userID != content.CreatorID {
			userRole, _ := r.Context().Value(middleware.UserRoleKey).(string)
			if userRole != string(models.RoleAdmin) {
				respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
				return
			}
		}
	}

	// Pagination
	page := 1
	pageSize := 20

	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
		if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}

	// Récupérer les commentaires
	var comments []models.Comment
	var total int64

	query := database.GetDB().Model(&models.Comment{}).
		Where("content_id = ?", contentID).
		Preload("User")

	// Compter le total
	query.Count(&total)

	// Récupérer avec pagination
	offset := (page - 1) * pageSize
	result = query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&comments)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des commentaires")
		return
	}

	// Calculer les métadonnées de pagination
	totalPages := (int(total) + pageSize - 1) / pageSize

	response := map[string]interface{}{
		"comments": comments,
		"pagination": map[string]interface{}{
			"page":        page,
			"size":        pageSize,
			"total_items": total,
			"total_pages": totalPages,
		},
	}

	respondWithJSON(w, http.StatusOK, response)
}

// AddComment ajoute un commentaire à un contenu
func AddComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentID := vars["id"]

	// Récupérer l'ID utilisateur du contexte
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Vérifier que le contenu existe et est accessible
	var content models.Content
	result := database.GetDB().First(&content, "id = ?", contentID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	// Vérifier si le contenu est publié
	if !content.IsPublished || content.IsFlagged {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	var commentRequest CreateCommentRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&commentRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}
	defer r.Body.Close()

	// Validation
	commentRequest.Content = utils.SanitizeInput(commentRequest.Content)
	if commentRequest.Content == "" {
		respondWithError(w, http.StatusBadRequest, "Le contenu du commentaire est requis")
		return
	}

	if len(commentRequest.Content) > 1000 {
		respondWithError(w, http.StatusBadRequest, "Le commentaire ne peut pas dépasser 1000 caractères")
		return
	}

	// Créer le commentaire
	contentIDUint, _ := strconv.ParseUint(contentID, 10, 32)
	comment := models.Comment{
		ContentID: uint(contentIDUint),
		UserID:    userID,
		Text:      commentRequest.Content,
	}

	result = database.GetDB().Create(&comment)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la création du commentaire")
		return
	}

	// Précharger l'utilisateur pour la réponse
	database.GetDB().Preload("User").First(&comment, comment.ID)

	respondWithJSON(w, http.StatusCreated, comment)
}

// UpdateComment met à jour un commentaire
func UpdateComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["id"]

	// Récupérer l'ID utilisateur du contexte
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Récupérer le commentaire
	var comment models.Comment
	result := database.GetDB().Preload("User").First(&comment, "id = ?", commentID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Commentaire non trouvé")
		return
	}

	// Vérifier les permissions
	if comment.UserID != userID {
		userRole, _ := r.Context().Value(middleware.UserRoleKey).(string)
		if userRole != string(models.RoleAdmin) {
			respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à modifier ce commentaire")
			return
		}
	}

	var updateRequest UpdateCommentRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updateRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}
	defer r.Body.Close()

	// Validation
	updateRequest.Content = utils.SanitizeInput(updateRequest.Content)
	if updateRequest.Content == "" {
		respondWithError(w, http.StatusBadRequest, "Le contenu du commentaire est requis")
		return
	}

	if len(updateRequest.Content) > 1000 {
		respondWithError(w, http.StatusBadRequest, "Le commentaire ne peut pas dépasser 1000 caractères")
		return
	}

	// Mettre à jour le commentaire
	comment.Text = updateRequest.Content
	result = database.GetDB().Save(&comment)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour du commentaire")
		return
	}

	respondWithJSON(w, http.StatusOK, comment)
}

// DeleteComment supprime un commentaire
func DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["id"]

	// Récupérer l'ID utilisateur du contexte
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Récupérer le commentaire
	var comment models.Comment
	result := database.GetDB().First(&comment, "id = ?", commentID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Commentaire non trouvé")
		return
	}

	// Vérifier les permissions
	if comment.UserID != userID {
		userRole, _ := r.Context().Value(middleware.UserRoleKey).(string)
		if userRole != string(models.RoleAdmin) {
			// Vérifier si c'est le créateur du contenu
			var content models.Content
			database.GetDB().First(&content, comment.ContentID)
			if content.CreatorID != userID {
				respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à supprimer ce commentaire")
				return
			}
		}
	}

	// Supprimer le commentaire
	result = database.GetDB().Delete(&comment)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la suppression du commentaire")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Commentaire supprimé avec succès",
	})
}

// LikeContent ajoute un like à un contenu
func LikeContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentID := vars["id"]

	// Récupérer l'ID utilisateur du contexte
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Vérifier que le contenu existe et est accessible
	var content models.Content
	result := database.GetDB().First(&content, "id = ?", contentID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	// Vérifier si le contenu est publié
	if !content.IsPublished || content.IsFlagged {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	// Vérifier si l'utilisateur a déjà liké ce contenu
	var existingLike models.Like
	contentIDUint, _ := strconv.ParseUint(contentID, 10, 32)
	result = database.GetDB().Where("content_id = ? AND user_id = ?", contentIDUint, userID).First(&existingLike)
	if result.Error == nil {
		respondWithError(w, http.StatusConflict, "Vous avez déjà liké ce contenu")
		return
	}

	// Créer le like
	like := models.Like{
		ContentID: uint(contentIDUint),
		UserID:    userID,
	}

	result = database.GetDB().Create(&like)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'ajout du like")
		return
	}

	// Compter le nombre total de likes pour ce contenu
	var likeCount int64
	database.GetDB().Model(&models.Like{}).Where("content_id = ?", contentIDUint).Count(&likeCount)

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message":    "Contenu liké avec succès",
		"like_count": likeCount,
	})
}

// UnlikeContent supprime un like d'un contenu
func UnlikeContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentID := vars["id"]

	// Récupérer l'ID utilisateur du contexte
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Vérifier que le contenu existe
	var content models.Content
	result := database.GetDB().First(&content, "id = ?", contentID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	// Chercher le like existant
	var like models.Like
	contentIDUint, _ := strconv.ParseUint(contentID, 10, 32)
	result = database.GetDB().Where("content_id = ? AND user_id = ?", contentIDUint, userID).First(&like)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Like non trouvé")
		return
	}

	// Supprimer le like
	result = database.GetDB().Delete(&like)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la suppression du like")
		return
	}

	// Compter le nombre total de likes pour ce contenu
	var likeCount int64
	database.GetDB().Model(&models.Like{}).Where("content_id = ?", contentIDUint).Count(&likeCount)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "Like supprimé avec succès",
		"like_count": likeCount,
	})
}

// GetContentLikes récupère la liste des utilisateurs qui ont liké un contenu
func GetContentLikes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentID := vars["id"]

	// Vérifier que le contenu existe et est accessible
	var content models.Content
	result := database.GetDB().First(&content, "id = ?", contentID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	// Vérifier si le contenu est publié
	if !content.IsPublished || content.IsFlagged {
		userID, _ := r.Context().Value(middleware.UserIDKey).(string)
		if userID != content.CreatorID {
			userRole, _ := r.Context().Value(middleware.UserRoleKey).(string)
			if userRole != string(models.RoleAdmin) {
				respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
				return
			}
		}
	}

	// Pagination
	page := 1
	pageSize := 20

	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
		if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}

	// Récupérer les likes avec les informations utilisateur
	var likes []models.Like
	var total int64

	contentIDUint, _ := strconv.ParseUint(contentID, 10, 32)
	query := database.GetDB().Model(&models.Like{}).
		Where("content_id = ?", contentIDUint).
		Preload("User")

	// Compter le total
	query.Count(&total)

	// Récupérer avec pagination
	offset := (page - 1) * pageSize
	result = query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&likes)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des likes")
		return
	}

	// Calculer les métadonnées de pagination
	totalPages := (int(total) + pageSize - 1) / pageSize

	response := map[string]interface{}{
		"likes": likes,
		"pagination": map[string]interface{}{
			"page":        page,
			"size":        pageSize,
			"total_items": total,
			"total_pages": totalPages,
		},
	}

	respondWithJSON(w, http.StatusOK, response)
}
