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

// Type pour les paramètres de mise à jour des notifications
type NotificationSettings struct {
	EmailNotifications   bool `json:"email_notifications"`
	PushNotifications    bool `json:"push_notifications"`
	MessageNotifications bool `json:"message_notifications"`
	CommentNotifications bool `json:"comment_notifications"`
	LikeNotifications    bool `json:"like_notifications"`
	FollowNotifications  bool `json:"follow_notifications"`
	ContentNotifications bool `json:"content_notifications"`
}

// UploadProfilePicture gère le téléchargement d'une photo de profil
func UploadProfilePicture(w http.ResponseWriter, r *http.Request) {
	// Extraire l'ID utilisateur depuis l'URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Vérifier si l'utilisateur connecté a le droit de modifier ce profil
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || id != userID {
		respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à modifier ce profil")
		return
	}

	// Limite de taille du fichier à 5MB
	r.ParseMultipartForm(5 << 20)

	// Récupérer le fichier depuis la requête
	file, handler, err := r.FormFile("profile_picture")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Erreur lors de la récupération du fichier")
		return
	}
	defer file.Close()

	// Vérifier le type MIME (devrait être une image)
	if handler.Header.Get("Content-Type") != "image/jpeg" &&
		handler.Header.Get("Content-Type") != "image/png" &&
		handler.Header.Get("Content-Type") != "image/gif" {
		respondWithError(w, http.StatusBadRequest, "Le fichier doit être une image (JPEG, PNG ou GIF)")
		return
	}

	// TODO: Implémenter le stockage réel du fichier (S3, système de fichiers, etc.)
	// Pour l'instant, supposons que nous avons stocké le fichier et obtenu une URL

	// URL fictive pour la démonstration
	profilePictureURL := "/uploads/profiles/" + handler.Filename

	// Mettre à jour l'URL de la photo de profil dans la base de données
	var user models.User
	result := database.GetDB().First(&user, id)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	user.ProfilePicture = profilePictureURL
	result = database.GetDB().Save(&user)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour de la photo de profil")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Photo de profil mise à jour avec succès",
		"url":     profilePictureURL,
	})
}

// GetFollowing récupère la liste des utilisateurs suivis par un utilisateur
func GetFollowing(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "ID utilisateur requis")
		return
	}

	// Vérifier si l'utilisateur existe
	var user models.User
	result := database.GetDB().First(&user, "id = ?", id)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Récupérer les utilisateurs suivis
	var follows []models.Follow
	result = database.GetDB().Where("follower_id = ?", id).Preload("Followed").Find(&follows)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des utilisateurs suivis")
		return
	}

	var following []models.UserResponse
	for _, follow := range follows {
		following = append(following, follow.Followed.ToResponse())
	}

	respondWithJSON(w, http.StatusOK, following)
}

// BlockUser bloque un utilisateur
func BlockUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	targetID := vars["targetId"]

	// Vérifier si l'utilisateur connecté est bien l'utilisateur qui fait la demande
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || userID != currentUserID {
		respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à effectuer cette action")
		return
	}

	// Vérifier que l'utilisateur ne tente pas de se bloquer lui-même
	if userID == targetID {
		respondWithError(w, http.StatusBadRequest, "Vous ne pouvez pas vous bloquer vous-même")
		return
	}

	// Vérifier si l'utilisateur cible existe
	var targetUser models.User
	result := database.GetDB().First(&targetUser, "id = ?", targetID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur cible non trouvé")
		return
	}

	// Vérifier si l'utilisateur n'est pas déjà bloqué
	var existingBlock models.Block
	result = database.GetDB().Where("blocker_id = ? AND blocked_id = ?", userID, targetID).First(&existingBlock)
	if result.Error == nil {
		respondWithError(w, http.StatusConflict, "Utilisateur déjà bloqué")
		return
	}

	// Créer le blocage
	block := models.Block{
		BlockerID: userID,
		BlockedID: targetID,
	}

	result = database.GetDB().Create(&block)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors du blocage de l'utilisateur")
		return
	}

	// Supprimer le suivi mutuel s'il existe
	database.GetDB().Where("(follower_id = ? AND followed_id = ?) OR (follower_id = ? AND followed_id = ?)", 
		userID, targetID, targetID, userID).Delete(&models.Follow{})

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Utilisateur bloqué avec succès",
	})
}

// UnblockUser débloque un utilisateur
func UnblockUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	targetID := vars["targetId"]

	// Vérifier si l'utilisateur connecté est bien l'utilisateur qui fait la demande
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || userID != currentUserID {
		respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à effectuer cette action")
		return
	}

	// Vérifier si le blocage existe
	var block models.Block
	result := database.GetDB().Where("blocker_id = ? AND blocked_id = ?", userID, targetID).First(&block)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Blocage non trouvé")
		return
	}

	// Supprimer le blocage
	result = database.GetDB().Delete(&block)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors du déblocage de l'utilisateur")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Utilisateur débloqué avec succès",
	})
}

// GetBlockedUsers récupère la liste des utilisateurs bloqués par un utilisateur
func GetBlockedUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Vérifier si l'utilisateur connecté est bien l'utilisateur qui fait la demande
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || id != currentUserID {
		respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à effectuer cette action")
		return
	}

	// Récupérer les utilisateurs bloqués
	var blocks []models.Block
	result := database.GetDB().Where("blocker_id = ?", id).Preload("Blocked").Find(&blocks)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des utilisateurs bloqués")
		return
	}

	var blockedUsers []models.UserResponse
	for _, block := range blocks {
		blockedUsers = append(blockedUsers, block.Blocked.ToResponse())
	}

	respondWithJSON(w, http.StatusOK, blockedUsers)
}

// UpdateNotificationSettings met à jour les paramètres de notification d'un utilisateur
func UpdateNotificationSettings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "ID utilisateur requis")
		return
	}

	// Vérifier si l'utilisateur connecté est bien l'utilisateur qui fait la demande
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || id != currentUserID {
		respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à effectuer cette action")
		return
	}

	var settings NotificationSettings
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&settings); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}
	defer r.Body.Close()

	// Récupérer ou créer les paramètres de notification
	var notificationSettings models.NotificationSettings
	result := database.GetDB().Where("user_id = ?", id).First(&notificationSettings)
	if result.Error != nil {
		// Créer les paramètres s'ils n'existent pas
		notificationSettings = models.NotificationSettings{
			UserID: id,
		}
	}

	// Mettre à jour les paramètres
	notificationSettings.EmailNotifications = settings.EmailNotifications
	notificationSettings.PushNotifications = settings.PushNotifications
	notificationSettings.MessageNotifications = settings.MessageNotifications
	notificationSettings.CommentNotifications = settings.CommentNotifications
	notificationSettings.LikeNotifications = settings.LikeNotifications
	notificationSettings.FollowNotifications = settings.FollowNotifications
	notificationSettings.ContentNotifications = settings.ContentNotifications

	// Sauvegarder les paramètres
	result = database.GetDB().Save(&notificationSettings)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour des paramètres de notification")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Paramètres de notification mis à jour avec succès",
	})
}

// GetFeed récupère le flux d'activités d'un utilisateur
func GetFeed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "ID utilisateur requis")
		return
	}

	// Vérifier si l'utilisateur connecté est bien l'utilisateur qui fait la demande
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || id != currentUserID {
		respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à effectuer cette action")
		return
	}

	// Pagination
	page := 1
	pageSize := 10

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

	offset := (page - 1) * pageSize

	// Récupérer les IDs des créateurs suivis
	var follows []models.Follow
	database.GetDB().Where("follower_id = ?", id).Find(&follows)

	var followedIDs []string
	for _, follow := range follows {
		followedIDs = append(followedIDs, follow.FollowedID)
	}

	// Récupérer les contenus récents des créateurs suivis
	var contents []models.Content
	query := database.GetDB().Where("creator_id IN ?", followedIDs).
		Preload("Creator").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset)
	
	result := query.Find(&contents)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération du flux")
		return
	}

	// Construire le feed
	var feedItems []map[string]interface{}
	for _, content := range contents {
		feedItem := map[string]interface{}{
			"id":   content.ID,
			"type": "content",
			"creator": map[string]interface{}{
				"id":              content.Creator.ID,
				"username":        content.Creator.Username,
				"profile_picture": content.Creator.ProfilePicture,
			},
			"content": map[string]interface{}{
				"id":            content.ID,
				"title":         content.Title,
				"description":   content.Description,
				"thumbnail_url": content.ThumbnailURL,
				"media_type":    content.Type,
				"created_at":    content.CreatedAt,
			},
		}
		feedItems = append(feedItems, feedItem)
	}

	// Compter le total pour la pagination
	var totalItems int64
	database.GetDB().Model(&models.Content{}).Where("creator_id IN ?", followedIDs).Count(&totalItems)
	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"items": feedItems,
		"pagination": map[string]int{
			"page":        page,
			"size":        pageSize,
			"total_items": int(totalItems),
			"total_pages": totalPages,
		},
	})
}

// FollowUser permet à un utilisateur de suivre un autre utilisateur
func FollowUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	targetID := vars["targetId"]

	// Vérifier si l'utilisateur connecté est bien l'utilisateur qui fait la demande
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || userID != currentUserID {
		respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à effectuer cette action")
		return
	}

	// Vérifier que l'utilisateur ne tente pas de se suivre lui-même
	if userID == targetID {
		respondWithError(w, http.StatusBadRequest, "Vous ne pouvez pas vous suivre vous-même")
		return
	}

	// Vérifier si l'utilisateur cible existe
	var targetUser models.User
	result := database.GetDB().First(&targetUser, "id = ?", targetID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur cible non trouvé")
		return
	}

	// Vérifier si l'utilisateur n'est pas bloqué
	var block models.Block
	result = database.GetDB().Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)", 
		userID, targetID, targetID, userID).First(&block)
	if result.Error == nil {
		respondWithError(w, http.StatusForbidden, "Impossible de suivre cet utilisateur")
		return
	}

	// Vérifier si l'utilisateur n'est pas déjà suivi
	var existingFollow models.Follow
	result = database.GetDB().Where("follower_id = ? AND followed_id = ?", userID, targetID).First(&existingFollow)
	if result.Error == nil {
		respondWithError(w, http.StatusConflict, "Vous suivez déjà cet utilisateur")
		return
	}

	// Créer le suivi
	follow := models.Follow{
		FollowerID: userID,
		FollowedID: targetID,
	}

	result = database.GetDB().Create(&follow)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors du suivi de l'utilisateur")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{
		"message": "Utilisateur suivi avec succès",
	})
}

// UnfollowUser permet à un utilisateur de ne plus suivre un autre utilisateur
func UnfollowUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	targetID := vars["targetId"]

	// Vérifier si l'utilisateur connecté est bien l'utilisateur qui fait la demande
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || userID != currentUserID {
		respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à effectuer cette action")
		return
	}

	// Vérifier si le suivi existe
	var follow models.Follow
	result := database.GetDB().Where("follower_id = ? AND followed_id = ?", userID, targetID).First(&follow)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Suivi non trouvé")
		return
	}

	// Supprimer le suivi
	result = database.GetDB().Delete(&follow)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'arrêt du suivi")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Suivi arrêté avec succès",
	})
}
