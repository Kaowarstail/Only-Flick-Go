package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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
	Type        string `json:"type"` // "image", "video", "text", "gallery"
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
	r.ParseMultipartForm(50 << 20)

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
