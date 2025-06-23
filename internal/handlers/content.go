package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

// CreateContentRequest représente les données pour créer un contenu
type CreateContentRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"` // image, video, text, gallery
	IsPremium   bool   `json:"is_premium"`
	IsPublished bool   `json:"is_published"`
}

// UpdateContentRequest représente les données pour mettre à jour un contenu
type UpdateContentRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	IsPremium   *bool   `json:"is_premium"`
	IsPublished *bool   `json:"is_published"`
}

// FeedItem représente un élément du fil d'actualité avec métadonnées étendues
type FeedItem struct {
	ID            uint                   `json:"id"`
	CreatorID     string                 `json:"creator_id"`
	Creator       map[string]interface{} `json:"creator"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Type          string                 `json:"type"`
	MediaURL      string                 `json:"media_url"`
	ThumbnailURL  string                 `json:"thumbnail_url"`
	ViewCount     int                    `json:"view_count"`
	LikesCount    int64                  `json:"likes_count"`
	CommentsCount int64                  `json:"comments_count"`
	IsLikedByUser bool                   `json:"is_liked_by_user"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// GetContents récupère tous les contenus publics avec pagination et filtres
func GetContents(w http.ResponseWriter, r *http.Request) {
	// Pagination
	page := 1
	pageSize := 10

	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
		if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 && s <= 50 {
			pageSize = s
		}
	}

	// Filtres
	contentType := r.URL.Query().Get("type")
	creatorID := r.URL.Query().Get("creator_id")

	// Construction de la requête
	query := database.GetDB().Model(&models.Content{}).
		Where("is_published = ? AND is_flagged = ?", true, false).
		Preload("Creator")

	// Appliquer les filtres
	if contentType != "" {
		query = query.Where("type = ?", contentType)
	}
	if creatorID != "" {
		query = query.Where("creator_id = ?", creatorID)
	}

	// Compter le total pour la pagination
	var total int64
	query.Count(&total)

	// Récupérer les contenus avec pagination
	var contents []models.Content
	offset := (page - 1) * pageSize
	result := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&contents)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des contenus")
		return
	}

	// Calculer les métadonnées de pagination
	totalPages := (int(total) + pageSize - 1) / pageSize

	response := map[string]interface{}{
		"contents": contents,
		"pagination": map[string]interface{}{
			"page":        page,
			"size":        pageSize,
			"total_items": total,
			"total_pages": totalPages,
		},
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetContent récupère un contenu spécifique par son ID
func GetContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var content models.Content
	result := database.GetDB().Preload("Creator").Preload("Comments.User").First(&content, "id = ?", id)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	// Vérifier si le contenu est publié et non signalé
	if !content.IsPublished || content.IsFlagged {
		// Vérifier si l'utilisateur est le créateur ou un admin
		userID, ok := r.Context().Value(middleware.UserIDKey).(string)
		if !ok || (userID != content.CreatorID) {
			userRole, _ := r.Context().Value(middleware.UserRoleKey).(string)
			if userRole != string(models.RoleAdmin) {
				respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
				return
			}
		}
	}

	// Incrémenter le compteur de vues
	database.GetDB().Model(&content).UpdateColumn("view_count", content.ViewCount+1)

	respondWithJSON(w, http.StatusOK, content)
}

// CreateContent crée un nouveau contenu
func CreateContent(w http.ResponseWriter, r *http.Request) {
	// Récupérer l'ID utilisateur du contexte
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Vérifier que l'utilisateur est un créateur
	var user models.User
	result := database.GetDB().First(&user, "id = ?", userID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	if user.Role != models.RoleCreator && user.Role != models.RoleAdmin {
		respondWithError(w, http.StatusForbidden, "Seuls les créateurs peuvent publier du contenu")
		return
	}

	var createRequest CreateContentRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&createRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}
	defer r.Body.Close()

	// Validation des données
	createRequest.Title = utils.SanitizeInput(createRequest.Title)
	createRequest.Description = utils.SanitizeInput(createRequest.Description)
	createRequest.Type = utils.SanitizeInput(createRequest.Type)

	if createRequest.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Le titre est requis")
		return
	}

	if len(createRequest.Title) > 200 {
		respondWithError(w, http.StatusBadRequest, "Le titre ne peut pas dépasser 200 caractères")
		return
	}

	if len(createRequest.Description) > 2000 {
		respondWithError(w, http.StatusBadRequest, "La description ne peut pas dépasser 2000 caractères")
		return
	}

	// Valider le type de contenu
	validTypes := []string{"image", "video", "text", "gallery"}
	isValidType := false
	for _, validType := range validTypes {
		if createRequest.Type == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		respondWithError(w, http.StatusBadRequest, "Type de contenu invalide")
		return
	}

	// Créer le contenu
	content := models.Content{
		CreatorID:   userID,
		Title:       createRequest.Title,
		Description: createRequest.Description,
		Type:        createRequest.Type,
		IsPremium:   createRequest.IsPremium,
		IsPublished: createRequest.IsPublished,
		ViewCount:   0,
		IsFlagged:   false,
	}

	result = database.GetDB().Create(&content)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la création du contenu")
		return
	}

	// Précharger les relations pour la réponse
	database.GetDB().Preload("Creator").First(&content, content.ID)

	respondWithJSON(w, http.StatusCreated, content)
}

// UpdateContent met à jour un contenu existant
func UpdateContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Récupérer l'ID utilisateur du contexte
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Récupérer le contenu
	var content models.Content
	result := database.GetDB().First(&content, "id = ?", id)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	// Vérifier les permissions
	if content.CreatorID != userID {
		userRole, _ := r.Context().Value(middleware.UserRoleKey).(string)
		if userRole != string(models.RoleAdmin) {
			respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à modifier ce contenu")
			return
		}
	}

	var updateRequest UpdateContentRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updateRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}
	defer r.Body.Close()

	// Appliquer les mises à jour
	if updateRequest.Title != nil {
		title := utils.SanitizeInput(*updateRequest.Title)
		if title == "" {
			respondWithError(w, http.StatusBadRequest, "Le titre ne peut pas être vide")
			return
		}
		if len(title) > 200 {
			respondWithError(w, http.StatusBadRequest, "Le titre ne peut pas dépasser 200 caractères")
			return
		}
		content.Title = title
	}

	if updateRequest.Description != nil {
		description := utils.SanitizeInput(*updateRequest.Description)
		if len(description) > 2000 {
			respondWithError(w, http.StatusBadRequest, "La description ne peut pas dépasser 2000 caractères")
			return
		}
		content.Description = description
	}

	if updateRequest.IsPremium != nil {
		content.IsPremium = *updateRequest.IsPremium
	}

	if updateRequest.IsPublished != nil {
		content.IsPublished = *updateRequest.IsPublished
	}

	// Sauvegarder les modifications
	result = database.GetDB().Save(&content)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour du contenu")
		return
	}

	// Précharger les relations pour la réponse
	database.GetDB().Preload("Creator").First(&content, content.ID)

	respondWithJSON(w, http.StatusOK, content)
}

// DeleteContent supprime un contenu
func DeleteContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Récupérer l'ID utilisateur du contexte
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Récupérer le contenu
	var content models.Content
	result := database.GetDB().First(&content, "id = ?", id)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	// Vérifier les permissions
	if content.CreatorID != userID {
		userRole, _ := r.Context().Value(middleware.UserRoleKey).(string)
		if userRole != string(models.RoleAdmin) {
			respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à supprimer ce contenu")
			return
		}
	}

	// Supprimer le contenu (soft delete grâce à GORM)
	result = database.GetDB().Delete(&content)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la suppression du contenu")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Contenu supprimé avec succès",
	})
}

// GetCreatorContents récupère tous les contenus d'un créateur
func GetCreatorContents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	creatorID := vars["id"]

	// Pagination
	page := 1
	pageSize := 10

	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
		if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 && s <= 50 {
			pageSize = s
		}
	}

	// Vérifier si l'utilisateur connecté est le créateur lui-même
	userID, _ := r.Context().Value(middleware.UserIDKey).(string)
	isOwner := userID == creatorID

	// Construction de la requête
	query := database.GetDB().Model(&models.Content{}).Where("creator_id = ?", creatorID)

	// Si ce n'est pas le propriétaire, ne montrer que les contenus publiés et non signalés
	if !isOwner {
		query = query.Where("is_published = ? AND is_flagged = ?", true, false)
	}

	// Compter le total
	var total int64
	query.Count(&total)

	// Récupérer les contenus
	var contents []models.Content
	offset := (page - 1) * pageSize
	result := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&contents)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des contenus")
		return
	}

	// Calculer les métadonnées de pagination
	totalPages := (int(total) + pageSize - 1) / pageSize

	response := map[string]interface{}{
		"contents": contents,
		"pagination": map[string]interface{}{
			"page":        page,
			"size":        pageSize,
			"total_items": total,
			"total_pages": totalPages,
		},
	}

	respondWithJSON(w, http.StatusOK, response)
}

// SearchContents recherche des contenus
func SearchContents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondWithError(w, http.StatusBadRequest, "Paramètre de recherche 'q' requis")
		return
	}

	query = utils.SanitizeInput(query)

	// Pagination
	page := 1
	pageSize := 10

	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
		if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 && s <= 50 {
			pageSize = s
		}
	}

	// Recherche dans les titres et descriptions
	searchPattern := "%" + strings.ToLower(query) + "%"
	dbQuery := database.GetDB().Model(&models.Content{}).
		Where("is_published = ? AND is_flagged = ?", true, false).
		Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern).
		Preload("Creator")

	// Compter le total
	var total int64
	dbQuery.Count(&total)

	// Récupérer les contenus
	var contents []models.Content
	offset := (page - 1) * pageSize
	result := dbQuery.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&contents)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la recherche")
		return
	}

	// Calculer les métadonnées de pagination
	totalPages := (int(total) + pageSize - 1) / pageSize

	response := map[string]interface{}{
		"contents": contents,
		"pagination": map[string]interface{}{
			"page":        page,
			"size":        pageSize,
			"total_items": total,
			"total_pages": totalPages,
		},
		"query": query,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetTrendingContents récupère les contenus populaires
func GetTrendingContents(w http.ResponseWriter, r *http.Request) {
	// Pagination
	page := 1
	pageSize := 10

	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
		if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 && s <= 50 {
			pageSize = s
		}
	}

	// Récupérer les contenus les plus vus des 7 derniers jours
	// (ou tous si pas assez récents)
	query := database.GetDB().Model(&models.Content{}).
		Where("is_published = ? AND is_flagged = ?", true, false).
		Preload("Creator")

	// Compter le total
	var total int64
	query.Count(&total)

	// Récupérer les contenus triés par nombre de vues
	var contents []models.Content
	offset := (page - 1) * pageSize
	result := query.Offset(offset).Limit(pageSize).Order("view_count DESC, created_at DESC").Find(&contents)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des contenus populaires")
		return
	}

	// Calculer les métadonnées de pagination
	totalPages := (int(total) + pageSize - 1) / pageSize

	response := map[string]interface{}{
		"contents": contents,
		"pagination": map[string]interface{}{
			"page":        page,
			"size":        pageSize,
			"total_items": total,
			"total_pages": totalPages,
		},
	}

	respondWithJSON(w, http.StatusOK, response)
}

// UploadContentMedia gère l'upload de fichiers média pour un contenu
func UploadContentMedia(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Récupérer l'ID utilisateur du contexte
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Récupérer le contenu
	var content models.Content
	result := database.GetDB().First(&content, "id = ?", id)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	// Vérifier les permissions
	if content.CreatorID != userID {
		respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à modifier ce contenu")
		return
	}

	// Limite de taille de fichier (50MB)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		respondWithError(w, http.StatusBadRequest, "Erreur lors de l'analyse du formulaire multipart")
		return
	}

	// Récupérer le fichier depuis la requête
	file, handler, err := r.FormFile("media")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Erreur lors de la récupération du fichier")
		return
	}
	defer file.Close()

	// Vérifier le type MIME selon le type de contenu
	contentType := handler.Header.Get("Content-Type")
	switch content.Type {
	case "image":
		if !strings.HasPrefix(contentType, "image/") {
			respondWithError(w, http.StatusBadRequest, "Le fichier doit être une image")
			return
		}
	case "video":
		if !strings.HasPrefix(contentType, "video/") {
			respondWithError(w, http.StatusBadRequest, "Le fichier doit être une vidéo")
			return
		}
	default:
		respondWithError(w, http.StatusBadRequest, "Type de contenu ne supportant pas l'upload de média")
		return
	}

	// TODO: Implémenter le stockage réel du fichier (S3, système de fichiers, etc.)
	// Pour l'instant, on simule avec une URL fictive
	mediaURL := "/uploads/content/" + handler.Filename

	// Mettre à jour l'URL du média dans la base de données
	content.MediaURL = mediaURL
	result = database.GetDB().Save(&content)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour du contenu")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message":   "Fichier média uploadé avec succès",
		"media_url": mediaURL,
	})
}

// UploadContentThumbnail gère l'upload de la miniature d'un contenu
func UploadContentThumbnail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Récupérer l'ID utilisateur du contexte
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Récupérer le contenu
	var content models.Content
	result := database.GetDB().First(&content, "id = ?", id)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	// Vérifier les permissions
	if content.CreatorID != userID {
		respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à modifier ce contenu")
		return
	}

	// Limite de taille de fichier (5MB)
	r.ParseMultipartForm(5 << 20)

	// Récupérer le fichier depuis la requête
	file, handler, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Erreur lors de la récupération du fichier")
		return
	}
	defer file.Close()

	// Vérifier que c'est une image
	contentType := handler.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		respondWithError(w, http.StatusBadRequest, "La miniature doit être une image")
		return
	}

	// TODO: Implémenter le stockage réel du fichier
	thumbnailURL := "/uploads/thumbnails/" + handler.Filename

	// Mettre à jour l'URL de la miniature
	content.ThumbnailURL = thumbnailURL
	result = database.GetDB().Save(&content)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour du contenu")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message":       "Miniature uploadée avec succès",
		"thumbnail_url": thumbnailURL,
	})
}

// GetPublicFeed récupère le fil d'actualité public avec contenus non-premium uniquement
func GetPublicFeed(w http.ResponseWriter, r *http.Request) {
	// Pagination
	page := 1
	pageSize := 10

	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
		if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 && s <= 20 {
			pageSize = s
		}
	}

	// Récupérer l'ID utilisateur du contexte (optionnel pour le fil public)
	userID, _ := r.Context().Value(middleware.UserIDKey).(string)

	// Construire la requête pour récupérer les contenus publics (non-premium)
	query := database.GetDB().Model(&models.Content{}).
		Where("is_published = ? AND is_flagged = ? AND is_premium = ?", true, false, false).
		Preload("Creator")

	// Compter le total pour la pagination
	var total int64
	query.Count(&total)

	// Récupérer les contenus avec pagination
	var contents []models.Content
	offset := (page - 1) * pageSize
	result := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&contents)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération du fil d'actualité")
		return
	}

	// Transformer les contenus en FeedItems avec métadonnées étendues
	var feedItems []FeedItem
	for _, content := range contents {
		// Compter les likes
		var likesCount int64
		database.GetDB().Model(&models.Like{}).Where("content_id = ?", content.ID).Count(&likesCount)

		// Compter les commentaires (non masqués)
		var commentsCount int64
		database.GetDB().Model(&models.Comment{}).Where("content_id = ? AND is_hidden = ?", content.ID, false).Count(&commentsCount)

		// Vérifier si l'utilisateur connecté a liké ce contenu
		isLikedByUser := false
		if userID != "" {
			var existingLike models.Like
			likeResult := database.GetDB().Where("content_id = ? AND user_id = ?", content.ID, userID).First(&existingLike)
			isLikedByUser = likeResult.Error == nil
		}

		// Construire les informations du créateur
		creator := map[string]interface{}{
			"id":              content.Creator.ID,
			"username":        content.Creator.Username,
			"profile_picture": content.Creator.ProfilePicture,
			"role":            content.Creator.Role,
		}

		feedItem := FeedItem{
			ID:            content.ID,
			CreatorID:     content.CreatorID,
			Creator:       creator,
			Title:         content.Title,
			Description:   content.Description,
			Type:          content.Type,
			MediaURL:      content.MediaURL,
			ThumbnailURL:  content.ThumbnailURL,
			ViewCount:     content.ViewCount,
			LikesCount:    likesCount,
			CommentsCount: commentsCount,
			IsLikedByUser: isLikedByUser,
			CreatedAt:     content.CreatedAt,
			UpdatedAt:     content.UpdatedAt,
		}

		feedItems = append(feedItems, feedItem)
	}

	// Calculer les métadonnées de pagination
	totalPages := (int(total) + pageSize - 1) / pageSize

	response := map[string]interface{}{
		"feed": feedItems,
		"pagination": map[string]interface{}{
			"page":        page,
			"size":        pageSize,
			"total_items": total,
			"total_pages": totalPages,
		},
	}

	respondWithJSON(w, http.StatusOK, response)
}

// LikeContent permet à un utilisateur de liker un contenu
func LikeContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID de contenu invalide")
		return
	}

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

	// Vérifier si l'utilisateur a déjà liké ce contenu
	var existingLike models.Like
	likeResult := database.GetDB().Where("content_id = ? AND user_id = ?", contentID, userID).First(&existingLike)
	if likeResult.Error == nil {
		respondWithError(w, http.StatusConflict, "Vous avez déjà liké ce contenu")
		return
	}

	// Créer le like
	like := models.Like{
		ContentID: uint(contentID),
		UserID:    userID,
	}

	result = database.GetDB().Create(&like)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'ajout du like")
		return
	}

	// Compter le nouveau nombre de likes
	var likesCount int64
	database.GetDB().Model(&models.Like{}).Where("content_id = ?", contentID).Count(&likesCount)

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message":     "Contenu liké avec succès",
		"likes_count": likesCount,
	})
}

// UnlikeContent permet à un utilisateur de retirer son like d'un contenu
func UnlikeContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID de contenu invalide")
		return
	}

	// Récupérer l'ID utilisateur du contexte
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Vérifier que le like existe
	var like models.Like
	result := database.GetDB().Where("content_id = ? AND user_id = ?", contentID, userID).First(&like)
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

	// Compter le nouveau nombre de likes
	var likesCount int64
	database.GetDB().Model(&models.Like{}).Where("content_id = ?", contentID).Count(&likesCount)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":     "Like retiré avec succès",
		"likes_count": likesCount,
	})
}

// GetContentLikes récupère la liste des utilisateurs qui ont liké un contenu
func GetContentLikes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID de contenu invalide")
		return
	}

	// Vérifier que le contenu existe
	var content models.Content
	result := database.GetDB().First(&content, "id = ?", contentID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	// Pagination
	page := 1
	pageSize := 20

	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	// Récupérer les likes avec informations des utilisateurs
	var likes []models.Like
	offset := (page - 1) * pageSize
	result = database.GetDB().
		Where("content_id = ?", contentID).
		Preload("User").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&likes)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des likes")
		return
	}

	// Compter le total
	var total int64
	database.GetDB().Model(&models.Like{}).Where("content_id = ?", contentID).Count(&total)

	// Transformer en réponse avec données utilisateur
	var likeUsers []map[string]interface{}
	for _, like := range likes {
		likeUser := map[string]interface{}{
			"user_id":         like.UserID,
			"username":        like.User.Username,
			"profile_picture": like.User.ProfilePicture,
			"created_at":      like.CreatedAt,
		}
		likeUsers = append(likeUsers, likeUser)
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	response := map[string]interface{}{
		"likes": likeUsers,
		"pagination": map[string]interface{}{
			"page":        page,
			"size":        pageSize,
			"total_items": total,
			"total_pages": totalPages,
		},
	}

	respondWithJSON(w, http.StatusOK, response)
}
