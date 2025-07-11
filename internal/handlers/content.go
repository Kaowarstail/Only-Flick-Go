package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/Kaowarstail/Only-Flick-Go/internal/services"
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

// CreateContentWithMedia crée un nouveau contenu et upload le média en une seule requête
func CreateContentWithMedia(w http.ResponseWriter, r *http.Request) {
	fmt.Println("📤 [CreateContentWithMedia] Début de la création de contenu avec média")

	// Récupérer l'ID utilisateur depuis le contexte JWT
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Vérifier que l'utilisateur est un créateur
	var user models.User
	result := database.GetDB().First(&user, "id = ? AND role = ?", userID, models.RoleCreator)
	if result.Error != nil {
		respondWithError(w, http.StatusForbidden, "Seuls les créateurs peuvent publier du contenu")
		return
	}

	// Limite de taille de fichier (20MB pour les vidéos)
	err := r.ParseMultipartForm(20 << 20)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Erreur lors du parsing de la requête multipart")
		return
	}

	// Récupérer les champs du formulaire
	title := r.FormValue("title")
	description := r.FormValue("description")
	contentType := r.FormValue("type") // "image" ou "video"
	isPremiumStr := r.FormValue("is_premium")
	isPublishedStr := r.FormValue("is_published")

	// Validations
	if title == "" {
		respondWithError(w, http.StatusBadRequest, "Le titre est requis")
		return
	}
	if contentType == "" {
		contentType = "image" // Par défaut
	}
	if contentType != "image" && contentType != "video" {
		respondWithError(w, http.StatusBadRequest, "Le type doit être 'image' ou 'video'")
		return
	}

	// Conversion des booléens
	isPremium := isPremiumStr == "true"
	isPublished := isPublishedStr == "true"

	// 1. Créer le contenu d'abord
	content := models.Content{
		CreatorID:   userID,
		Title:       title,
		Description: description,
		Type:        contentType,
		IsPremium:   isPremium,
		IsPublished: isPublished,
	}

	result = database.GetDB().Create(&content)
	if result.Error != nil {
		fmt.Printf("❌ [CreateContentWithMedia] Erreur lors de la création du contenu: %v\n", result.Error)
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la création du contenu")
		return
	}

	fmt.Printf("✅ [CreateContentWithMedia] Contenu créé avec ID: %s\n", content.ID)

	// 2. Récupérer le fichier média
	file, handler, err := r.FormFile("media")
	if err != nil {
		fmt.Printf("❌ [CreateContentWithMedia] Erreur lors de la récupération du fichier: %v\n", err)
		respondWithError(w, http.StatusBadRequest, "Fichier média requis")
		return
	}
	defer file.Close()

	fmt.Printf("📁 [CreateContentWithMedia] Fichier reçu: %s, Taille: %d octets\n", handler.Filename, handler.Size)

	// 3. Initialiser le service Cloudinary
	cloudinaryService, err := services.NewCloudinaryService()
	if err != nil {
		fmt.Printf("❌ [CreateContentWithMedia] Erreur lors de l'initialisation du service Cloudinary: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'initialisation du service Cloudinary")
		return
	}

	// 4. Valider le type de fichier
	contentTypeHeader := handler.Header.Get("Content-Type")
	if !cloudinaryService.ValidateFileType(contentTypeHeader, contentType, handler.Filename) {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Type de fichier non supporté pour le type '%s'", contentType))
		return
	}

	// 5. Upload vers Cloudinary
	var uploadResult *services.UploadResult
	contentIDStr := strconv.Itoa(int(content.ID))
	if contentType == "image" {
		uploadResult, err = cloudinaryService.UploadImage(file, handler.Filename, contentIDStr)
	} else if contentType == "video" {
		uploadResult, err = cloudinaryService.UploadVideo(file, handler.Filename, contentIDStr)
	}

	if err != nil {
		fmt.Printf("❌ [CreateContentWithMedia] Erreur lors de l'upload: %v\n", err)
		// Supprimer le contenu créé en cas d'échec d'upload
		database.GetDB().Delete(&content)
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'upload du fichier: "+err.Error())
		return
	}

	// 6. Générer une miniature automatiquement
	thumbnailURL, err := cloudinaryService.GenerateThumbnail(uploadResult.PublicID, uploadResult.ResourceType)
	if err != nil {
		fmt.Printf("⚠️ [CreateContentWithMedia] Erreur lors de la génération de la miniature: %v\n", err)
		// Continuer même si la miniature échoue
		thumbnailURL = uploadResult.SecureURL
	}

	// 7. Mettre à jour le contenu avec les URLs
	content.MediaURL = uploadResult.SecureURL
	content.ThumbnailURL = thumbnailURL
	content.PublicID = uploadResult.PublicID

	result = database.GetDB().Save(&content)
	if result.Error != nil {
		fmt.Printf("❌ [CreateContentWithMedia] Erreur lors de la mise à jour du contenu: %v\n", result.Error)
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la sauvegarde des URLs")
		return
	}

	fmt.Printf("✅ [CreateContentWithMedia] Contenu créé et uploadé avec succès. URL: %s\n", content.MediaURL)

	// 8. Construire la réponse
	response := map[string]interface{}{
		"message": "Contenu créé et média uploadé avec succès",
		"content": map[string]interface{}{
			"id":            content.ID,
			"title":         content.Title,
			"description":   content.Description,
			"type":          content.Type,
			"media_url":     content.MediaURL,
			"thumbnail_url": content.ThumbnailURL,
			"public_id":     content.PublicID,
			"is_premium":    content.IsPremium,
			"is_published":  content.IsPublished,
			"created_at":    content.CreatedAt,
		},
		"upload_info": map[string]interface{}{
			"format":        uploadResult.Format,
			"resource_type": uploadResult.ResourceType,
			"width":         uploadResult.Width,
			"height":        uploadResult.Height,
			"bytes":         uploadResult.Bytes,
		},
	}

	fmt.Println("🎉 [CreateContentWithMedia] Processus terminé avec succès")
	respondWithJSON(w, http.StatusCreated, response)
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

	// Vérifier que l'utilisateur est un créateur, subscriber ou admin
	var user models.User
	result := database.GetDB().First(&user, "id = ?", userID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	if user.Role != models.RoleCreator && user.Role != models.RoleAdmin && user.Role != models.RoleSubscriber {
		respondWithError(w, http.StatusForbidden, "Seuls les créateurs et abonnés peuvent publier du contenu")
		return
	}

	var createRequest CreateContentRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&createRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}
	defer r.Body.Close()

	// Les subscribers peuvent seulement publier du contenu gratuit
	if user.Role == models.RoleSubscriber && createRequest.IsPremium {
		respondWithError(w, http.StatusForbidden, "Les abonnés ne peuvent publier que du contenu gratuit")
		return
	}

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
		services.RecordError("database", "error")
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la création du contenu")
		return
	}

	// Enregistrer la métrique de création de contenu
	services.RecordContentCreation(content.Type, content.CreatorID)

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

	// Supprimer le fichier de Cloudinary si un PublicID existe
	if content.PublicID != "" {
		cloudinaryService, err := services.NewCloudinaryService()
		if err == nil {
			// Déterminer le type de ressource basé sur le type de contenu
			resourceType := "image"
			if content.Type == "video" {
				resourceType = "video"
			}

			err = cloudinaryService.DeleteFile(content.PublicID, resourceType)
			if err != nil {
				fmt.Printf("⚠️ [DeleteContent] Erreur lors de la suppression du fichier Cloudinary: %v\n", err)
				// Ne pas faire échouer la suppression du contenu si la suppression Cloudinary échoue
			} else {
				fmt.Printf("✅ [DeleteContent] Fichier Cloudinary supprimé: %s\n", content.PublicID)
			}
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Contenu supprimé avec succès",
	})
}

// GetCreatorContents récupère tous les contenus d'un créateur organisés par statut premium
func GetCreatorContents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	creatorID := vars["id"]

	// Vérifier que le créateur existe
	var creator models.User
	result := database.GetDB().First(&creator, "id = ? AND role = ?", creatorID, models.RoleCreator)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Créateur non trouvé")
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

	if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
		if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 && s <= 50 {
			pageSize = s
		}
	}

	// Vérifier si l'utilisateur connecté est le créateur lui-même
	userID, _ := r.Context().Value(middleware.UserIDKey).(string)
	isOwner := userID == creatorID

	// Construire les conditions de base
	baseConditions := "creator_id = ?"
	baseArgs := []interface{}{creatorID}

	// Si ce n'est pas le propriétaire, ne montrer que les contenus publiés et non signalés
	if !isOwner {
		baseConditions += " AND is_published = ? AND is_flagged = ?"
		baseArgs = append(baseArgs, true, false)
	}

	// Récupérer les contenus gratuits
	var freeContents []models.Content
	freeConditions := baseConditions + " AND is_premium = ?"
	freeArgs := append(baseArgs, false)

	freeQuery := database.GetDB().Model(&models.Content{}).Where(freeConditions, freeArgs...)
	var totalFree int64
	freeQuery.Count(&totalFree)

	offset := (page - 1) * pageSize
	freeQuery.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&freeContents)

	// Récupérer les contenus premium
	var premiumContents []models.Content
	premiumConditions := baseConditions + " AND is_premium = ?"
	premiumArgs := append(baseArgs, true)

	premiumQuery := database.GetDB().Model(&models.Content{}).Where(premiumConditions, premiumArgs...)
	var totalPremium int64
	premiumQuery.Count(&totalPremium)

	premiumQuery.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&premiumContents)

	// Calculer les métadonnées de pagination
	totalFreePages := (int(totalFree) + pageSize - 1) / pageSize
	totalPremiumPages := (int(totalPremium) + pageSize - 1) / pageSize

	response := map[string]interface{}{
		"creator": map[string]interface{}{
			"id":       creator.ID,
			"username": creator.Username,
			"email":    creator.Email,
		},
		"free_content": map[string]interface{}{
			"contents": freeContents,
			"pagination": map[string]interface{}{
				"page":        page,
				"size":        pageSize,
				"total_items": totalFree,
				"total_pages": totalFreePages,
			},
		},
		"premium_content": map[string]interface{}{
			"contents": premiumContents,
			"pagination": map[string]interface{}{
				"page":        page,
				"size":        pageSize,
				"total_items": totalPremium,
				"total_pages": totalPremiumPages,
			},
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

// UploadContentMedia uploade un média pour un contenu existant
func UploadContentMedia(w http.ResponseWriter, r *http.Request) {
	fmt.Println("📤 [UploadContentMedia] Début de l'upload de média")

	// Récupérer l'ID du contenu depuis l'URL
	vars := mux.Vars(r)
	contentID := vars["id"]
	if contentID == "" {
		respondWithError(w, http.StatusBadRequest, "ID du contenu requis")
		return
	}

	// Récupérer l'ID utilisateur depuis le contexte JWT
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Vérifier que le contenu existe et appartient à l'utilisateur
	var content models.Content
	result := database.GetDB().First(&content, "id = ? AND creator_id = ?", contentID, userID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé ou accès refusé")
		return
	}

	// Limite de taille de fichier (20MB)
	r.ParseMultipartForm(20 << 20)

	// Récupérer le fichier depuis la requête
	file, handler, err := r.FormFile("media")
	if err != nil {
		fmt.Printf("❌ [UploadContentMedia] Erreur lors de la récupération du fichier: %v\n", err)
		respondWithError(w, http.StatusBadRequest, "Erreur lors de la récupération du fichier")
		return
	}
	defer file.Close()

	fmt.Printf("📁 [UploadContentMedia] Fichier reçu: %s, Taille: %d octets\n", handler.Filename, handler.Size)

	// Initialiser le service Cloudinary
	fmt.Println("🔧 [UploadContentMedia] Initialisation du service Cloudinary")
	cloudinaryService, err := services.NewCloudinaryService()
	if err != nil {
		fmt.Printf("❌ [UploadContentMedia] Erreur lors de l'initialisation du service Cloudinary: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'initialisation du service Cloudinary")
		return
	}

	// Valider le type de fichier
	contentTypeHeader := handler.Header.Get("Content-Type")
	if !cloudinaryService.ValidateFileType(contentTypeHeader, content.Type, handler.Filename) {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Type de fichier non supporté pour le type '%s'", content.Type))
		return
	}

	// Supprimer l'ancien média s'il existe
	if content.PublicID != "" {
		fmt.Printf("🗑️ [UploadContentMedia] Suppression de l'ancien média\n")
		err = cloudinaryService.DeleteFile(content.PublicID, content.Type)
		if err != nil {
			fmt.Printf("⚠️ [UploadContentMedia] Erreur lors de la suppression de l'ancien média: %v\n", err)
		}
	}

	// Upload vers Cloudinary
	fmt.Println("🚀 [UploadContentMedia] Upload vers Cloudinary")
	var uploadResult *services.UploadResult
	contentIDStr := strconv.Itoa(int(content.ID))
	if content.Type == "image" {
		uploadResult, err = cloudinaryService.UploadImage(file, handler.Filename, contentIDStr)
	} else if content.Type == "video" {
		uploadResult, err = cloudinaryService.UploadVideo(file, handler.Filename, contentIDStr)
	} else {
		respondWithError(w, http.StatusBadRequest, "Type de contenu non supporté")
		return
	}

	if err != nil {
		fmt.Printf("❌ [UploadContentMedia] Erreur lors de l'upload: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'upload: "+err.Error())
		return
	}

	// Générer une miniature
	thumbnailURL, err := cloudinaryService.GenerateThumbnail(uploadResult.PublicID, uploadResult.ResourceType)
	if err != nil {
		fmt.Printf("⚠️ [UploadContentMedia] Erreur lors de la génération de la miniature: %v\n", err)
		thumbnailURL = uploadResult.SecureURL
	}

	// Mettre à jour le contenu avec les nouvelles URLs
	fmt.Println("💾 [UploadContentMedia] Mise à jour du contenu en base de données")
	content.MediaURL = uploadResult.SecureURL
	content.ThumbnailURL = thumbnailURL
	content.PublicID = uploadResult.PublicID

	result = database.GetDB().Save(&content)
	if result.Error != nil {
		fmt.Printf("❌ [UploadContentMedia] Erreur lors de la mise à jour du contenu: %s\n", result.Error.Error())
		// Optionnel: supprimer l'image de Cloudinary en cas d'erreur
		cloudinaryService.DeleteFile(uploadResult.PublicID, "image")
		database.GetDB().Delete(&content)
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour du contenu")
		return
	}

	fmt.Printf("✅ [UploadContentMedia] Média uploadé avec succès. URL: %s\n", content.MediaURL)

	response := map[string]interface{}{
		"message":       "Média uploadé avec succès",
		"media_url":     uploadResult.SecureURL,
		"thumbnail_url": thumbnailURL,
		"public_id":     uploadResult.PublicID,
		"content": map[string]interface{}{
			"id":            content.ID,
			"title":         content.Title,
			"media_url":     content.MediaURL,
			"thumbnail_url": content.ThumbnailURL,
		},
		"file_info": map[string]interface{}{
			"format":        uploadResult.Format,
			"resource_type": uploadResult.ResourceType,
			"width":         uploadResult.Width,
			"height":        uploadResult.Height,
			"bytes":         uploadResult.Bytes,
		},
	}

	fmt.Println("🎉 [UploadContentMedia] Upload terminé avec succès")
	respondWithJSON(w, http.StatusOK, response)
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

	// Initialiser le service Cloudinary
	cloudinaryService, err := services.NewCloudinaryService()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'initialisation du service Cloudinary")
		return
	}

	// Uploader la miniature vers Cloudinary
	uploadResult, err := cloudinaryService.UploadImage(file, handler.Filename, fmt.Sprintf("%d", content.ID))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'upload: "+err.Error())
		return
	}

	thumbnailURL := uploadResult.SecureURL

	// Mettre à jour l'URL de la miniature et le public ID
	content.ThumbnailURL = thumbnailURL
	content.PublicID = uploadResult.PublicID
	result = database.GetDB().Save(&content)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour du contenu")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":       "Miniature uploadée avec succès",
		"thumbnail_url": thumbnailURL,
		"public_id":     uploadResult.PublicID,
		"file_info": map[string]interface{}{
			"format":        uploadResult.Format,
			"resource_type": uploadResult.ResourceType,
			"width":         uploadResult.Width,
			"height":        uploadResult.Height,
			"bytes":         uploadResult.Bytes,
		},
	})
}

// GetOptimizedMediaURL génère une URL optimisée pour un contenu
func GetOptimizedMediaURL(w http.ResponseWriter, r *http.Request) {
	// Récupérer l'ID du contenu depuis l'URL
	vars := mux.Vars(r)
	contentIDStr, exists := vars["id"]
	if !exists {
		respondWithError(w, http.StatusBadRequest, "ID du contenu manquant")
		return
	}

	contentID, err := strconv.Atoi(contentIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID du contenu invalide")
		return
	}

	// Récupérer les paramètres de transformation depuis la query string
	width := r.URL.Query().Get("w")
	height := r.URL.Query().Get("h")
	quality := r.URL.Query().Get("q")
	format := r.URL.Query().Get("f")

	// Récupérer le contenu
	var content models.Content
	result := database.GetDB().First(&content, "id = ?", contentID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	// Vérifier si le contenu a un PublicID Cloudinary
	if content.PublicID == "" {
		// Retourner l'URL standard si pas de PublicID
		respondWithJSON(w, http.StatusOK, map[string]string{
			"media_url":     content.MediaURL,
			"thumbnail_url": content.ThumbnailURL,
		})
		return
	}

	// Initialiser le service Cloudinary
	cloudinaryService, err := services.NewCloudinaryService()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'initialisation du service Cloudinary")
		return
	}

	// Construire les transformations
	var transformations []string

	if width != "" && height != "" {
		transformations = append(transformations, fmt.Sprintf("w_%s,h_%s,c_fill", width, height))
	}

	if quality != "" {
		transformations = append(transformations, fmt.Sprintf("q_%s", quality))
	} else {
		transformations = append(transformations, "q_auto")
	}

	if format != "" {
		transformations = append(transformations, fmt.Sprintf("f_%s", format))
	} else {
		transformations = append(transformations, "f_auto")
	}

	// Générer les URLs optimisées
	var optimizedURL, thumbnailURL string

	if content.Type == "video" {
		optimizedURL, err = cloudinaryService.GetVideoURL(content.PublicID, transformations...)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Erreur lors de la génération de l'URL vidéo")
			return
		}

		// Générer une miniature pour la vidéo
		thumbnailURL, err = cloudinaryService.GetThumbnailURL(content.PublicID, "video", 300, 300)
		if err != nil {
			thumbnailURL = content.ThumbnailURL // Fallback
		}
	} else {
		optimizedURL, err = cloudinaryService.GetImageURL(content.PublicID, transformations...)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Erreur lors de la génération de l'URL image")
			return
		}

		// Générer une miniature pour l'image
		thumbnailURL, err = cloudinaryService.GetThumbnailURL(content.PublicID, "image", 300, 300)
		if err != nil {
			thumbnailURL = content.ThumbnailURL // Fallback
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"media_url":       optimizedURL,
		"thumbnail_url":   thumbnailURL,
		"public_id":       content.PublicID,
		"transformations": transformations,
	})
}

// MigrateContentToCloudinary migre un contenu existant vers Cloudinary
func MigrateContentToCloudinary(w http.ResponseWriter, r *http.Request) {
	// Récupérer l'ID du contenu depuis l'URL
	vars := mux.Vars(r)
	contentIDStr, exists := vars["id"]
	if !exists {
		respondWithError(w, http.StatusBadRequest, "ID du contenu manquant")
		return
	}

	contentID, err := strconv.Atoi(contentIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID du contenu invalide")
		return
	}

	// Récupérer l'ID utilisateur du contexte
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Récupérer le contenu
	var content models.Content
	result := database.GetDB().First(&content, "id = ? AND creator_id = ?", contentID, userID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé ou non autorisé")
		return
	}

	// Vérifier si le contenu a déjà un PublicID Cloudinary
	if content.PublicID != "" {
		respondWithError(w, http.StatusBadRequest, "Le contenu est déjà migré vers Cloudinary")
		return
	}

	// Vérifier si le contenu a une URL média
	if content.MediaURL == "" {
		respondWithError(w, http.StatusBadRequest, "Le contenu n'a pas d'URL média à migrer")
		return
	}

	// TODO: Implémenter la migration depuis une URL existante
	// Cela nécessiterait de télécharger le fichier depuis l'URL existante
	// et de l'uploader vers Cloudinary

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Migration non implémentée",
		"content": content,
	})
}

// UploadContentImage gère l'upload d'une image pour un contenu et la création du contenu
func UploadContentImage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("📤 [UploadContentImage] Début de l'upload d'image pour contenu")

	// Récupérer l'ID utilisateur du contexte (créateur)
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		fmt.Println("❌ [UploadContentImage] Impossible d'extraire l'ID utilisateur")
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Vérifier que l'utilisateur est un créateur
	var user models.User
	result := database.GetDB().First(&user, "id = ? AND role = ?", userID, models.RoleCreator)
	if result.Error != nil {
		fmt.Printf("❌ [UploadContentImage] Créateur non trouvé: %s\n", userID)
		respondWithError(w, http.StatusForbidden, "Seuls les créateurs peuvent publier du contenu")
		return
	}

	// Limite de taille de fichier (10MB pour les images)
	r.ParseMultipartForm(10 << 20)

	// Récupérer les métadonnées du contenu depuis le formulaire
	title := r.FormValue("title")
	description := r.FormValue("description")
	isPremiumStr := r.FormValue("is_premium")
	isPublishedStr := r.FormValue("is_published")

	// Validation des champs obligatoires
	if title == "" {
		respondWithError(w, http.StatusBadRequest, "Le titre est obligatoire")
		return
	}

	// Conversion des booléens
	isPremium, _ := strconv.ParseBool(isPremiumStr)
	isPublished, _ := strconv.ParseBool(isPublishedStr)

	// Récupérer le fichier image depuis la requête
	file, handler, err := r.FormFile("image")
	if err != nil {
		fmt.Printf("❌ [UploadContentImage] Erreur lors de la récupération du fichier: %v\n", err)
		respondWithError(w, http.StatusBadRequest, "Erreur lors de la récupération du fichier image")
		return
	}
	defer file.Close()

	fmt.Printf("📁 [UploadContentImage] Fichier reçu: %s, Taille: %d octets\n", handler.Filename, handler.Size)

	// Vérifier que c'est bien une image
	contentType := handler.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		fmt.Printf("❌ [UploadContentImage] Type de fichier non supporté: %s\n", contentType)
		respondWithError(w, http.StatusBadRequest, "Le fichier doit être une image")
		return
	}

	// Créer d'abord l'entité Content en base pour avoir un ID
	content := models.Content{
		CreatorID:   userID,
		Title:       title,
		Description: description,
		Type:        "image",
		IsPremium:   isPremium,
		IsPublished: isPublished,
	}

	// Sauvegarder le contenu (sans l'URL pour l'instant)
	result = database.GetDB().Create(&content)
	if result.Error != nil {
		fmt.Printf("❌ [UploadContentImage] Erreur lors de la création du contenu: %v\n", result.Error)
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la création du contenu")
		return
	}

	fmt.Printf("✅ [UploadContentImage] Contenu créé avec ID: %s\n", content.ID)

	// Initialiser le service Cloudinary
	fmt.Println("🔧 [UploadContentImage] Initialisation du service Cloudinary")
	cloudinaryService, err := services.NewCloudinaryService()
	if err != nil {
		fmt.Printf("❌ [UploadContentImage] Erreur lors de l'initialisation du service Cloudinary: %v\n", err)
		// Supprimer le contenu créé en cas d'erreur
		database.GetDB().Delete(&content)
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'initialisation du service Cloudinary")
		return
	}

	// Uploader l'image vers Cloudinary
	fmt.Println("🚀 [UploadContentImage] Upload vers Cloudinary")
	contentIDStr := fmt.Sprintf("%d", content.ID)
	uploadResult, err := cloudinaryService.UploadImage(file, handler.Filename, contentIDStr)
	if err != nil {
		fmt.Printf("❌ [UploadContentImage] Erreur lors de l'upload: %v\n", err)
		// Supprimer le contenu créé en cas d'erreur
		database.GetDB().Delete(&content)
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'upload vers Cloudinary: "+err.Error())
		return
	}

	// Mettre à jour le contenu avec l'URL de l'image
	fmt.Println("💾 [UploadContentImage] Mise à jour du contenu avec l'URL Cloudinary")
	content.MediaURL = uploadResult.SecureURL
	content.ThumbnailURL = uploadResult.SecureURL // Pour les images, on peut utiliser la même URL
	content.PublicID = uploadResult.PublicID      // Stocker le PublicID pour la suppression future

	// Sauvegarder les modifications
	result = database.GetDB().Save(&content)
	if result.Error != nil {
		fmt.Printf("❌ [UploadContentImage] Erreur lors de la mise à jour du contenu: %s\n", result.Error.Error())
		// Optionnel: supprimer l'image de Cloudinary en cas d'erreur
		cloudinaryService.DeleteFile(uploadResult.PublicID, "image")
		database.GetDB().Delete(&content)
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour du contenu")
		return
	}

	fmt.Printf("✅ [UploadContentImage] Contenu créé avec succès. URL: %s\n", content.MediaURL)

	// Réponse complète avec toutes les informations
	response := map[string]interface{}{
		"message":    "Contenu créé et image uploadée avec succès",
		"content_id": content.ID,
		"media_url":  uploadResult.SecureURL,
		"public_id":  uploadResult.PublicID,
		"content": map[string]interface{}{
			"id":            content.ID,
			"title":         content.Title,
			"description":   content.Description,
			"type":          content.Type,
			"media_url":     content.MediaURL,
			"thumbnail_url": content.ThumbnailURL,
			"is_premium":    content.IsPremium,
			"is_published":  content.IsPublished,
			"creator_id":    content.CreatorID,
			"created_at":    content.CreatedAt,
		},
		"cloudinary_info": map[string]interface{}{
			"format":        uploadResult.Format,
			"resource_type": uploadResult.ResourceType,
			"width":         uploadResult.Width,
			"height":        uploadResult.Height,
			"bytes":         uploadResult.Bytes,
		},
	}

	fmt.Println("🎉 [UploadContentImage] Upload terminé avec succès")
	respondWithJSON(w, http.StatusCreated, response)
}
