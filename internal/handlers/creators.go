package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

// BecomeCreatorRequest représente les données de demande pour devenir créateur
type BecomeCreatorRequest struct {
	Biography   string                 `json:"biography"`
	Categories  []string               `json:"categories"`
	WebsiteURL  string                 `json:"website_url"`
	SocialLinks map[string]interface{} `json:"social_links"`
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
	result := database.GetDB().Where("id = ?", userID).First(&user)
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

	// Créer le profil créateur avec les informations supplémentaires
	socialLinksJSON := ""
	if req.SocialLinks != nil && len(req.SocialLinks) > 0 {
		socialLinksBytes, err := json.Marshal(req.SocialLinks)
		if err == nil {
			socialLinksJSON = string(socialLinksBytes)
		}
	}

	creatorProfile := models.CreatorProfile{
		UserID:      userID,
		WebsiteURL:  req.WebsiteURL,
		SocialLinks: socialLinksJSON,
	}

	result = database.GetDB().Create(&creatorProfile)
	if result.Error != nil {
		// Si la création du profil échoue, on continue quand même
		// car l'utilisateur est maintenant un créateur
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
	id := vars["id"]

	// Extraire l'ID utilisateur du contexte
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Vérifier si l'utilisateur authentifié est bien le créateur
	if userID != id {
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

// getUserIDFromContext récupère l'ID utilisateur depuis le contexte
func getUserIDFromContext(ctx context.Context) string {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return ""
	}
	return userID
}

// UpdateCreatorProfile met à jour le profil d'un créateur
func UpdateCreatorProfile(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	var updateData struct {
		Biography           *string                `json:"biography"`
		Categories          []string               `json:"categories"`
		WebsiteURL          *string                `json:"website_url"`
		SocialLinks         map[string]interface{} `json:"social_links"`
		SubscriptionPrice   *float64               `json:"subscription_price"`
		MessagePrice        *float64               `json:"message_price"`
		CustomTipAmounts    []float64              `json:"custom_tip_amounts"`
		AcceptCustomTips    *bool                  `json:"accept_custom_tips"`
		AcceptMessaging     *bool                  `json:"accept_messaging"`
		AcceptSubscriptions *bool                  `json:"accept_subscriptions"`
		ProfilePicture      *string                `json:"profile_picture"`
		CoverPicture        *string                `json:"cover_picture"`
		Location            *string                `json:"location"`
		IsVerified          *bool                  `json:"is_verified"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updateData); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}
	defer r.Body.Close()

	// Récupérer l'utilisateur créateur
	var user models.User
	if err := database.GetDB().First(&user, "id = ? AND role = ?", userID, models.RoleCreator).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Créateur non trouvé")
		return
	}

	// Mise à jour des champs de base
	if updateData.Biography != nil {
		user.Biography = *updateData.Biography
	}
	if updateData.WebsiteURL != nil {
		user.Website = updateData.WebsiteURL
	}
	if updateData.SubscriptionPrice != nil {
		user.SubscriptionPrice = updateData.SubscriptionPrice
	}
	if updateData.MessagePrice != nil {
		user.MessagePrice = updateData.MessagePrice
	}
	if updateData.CustomTipAmounts != nil {
		user.CustomTipAmounts = updateData.CustomTipAmounts
	}
	if updateData.AcceptCustomTips != nil {
		user.AcceptCustomTips = updateData.AcceptCustomTips
	}
	if updateData.AcceptMessaging != nil {
		user.AcceptMessaging = updateData.AcceptMessaging
	}
	if updateData.AcceptSubscriptions != nil {
		user.AcceptSubscriptions = updateData.AcceptSubscriptions
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
	if updateData.IsVerified != nil {
		user.IsVerified = updateData.IsVerified
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
		Message: "Profil créateur mis à jour avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetCreatorEarnings récupère les gains d'un créateur
func GetCreatorEarnings(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Vérifier que l'utilisateur est un créateur
	var user models.User
	if err := database.GetDB().First(&user, "id = ? AND role = ?", userID, models.RoleCreator).Error; err != nil {
		respondWithError(w, http.StatusForbidden, "Accès réservé aux créateurs")
		return
	}

	var earnings models.CreatorEarnings
	if err := database.GetDB().First(&earnings, "creator_id = ?", userID).Error; err != nil {
		// Créer un enregistrement de gains par défaut
		earnings = models.CreatorEarnings{
			CreatorID: userID,
		}
		database.GetDB().Create(&earnings)
	}

	response := models.APIResponse{
		Success: true,
		Data:    earnings,
		Message: "Gains récupérés avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetCreatorMonthlyEarnings récupère les gains mensuels d'un créateur
func GetCreatorMonthlyEarnings(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Vérifier que l'utilisateur est un créateur
	var user models.User
	if err := database.GetDB().First(&user, "id = ? AND role = ?", userID, models.RoleCreator).Error; err != nil {
		respondWithError(w, http.StatusForbidden, "Accès réservé aux créateurs")
		return
	}

	var monthlyEarnings []models.MonthlyEarning
	if err := database.GetDB().Where("creator_id = ?", userID).Order("year DESC, month DESC").Find(&monthlyEarnings).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des gains mensuels")
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    monthlyEarnings,
		Message: "Gains mensuels récupérés avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetCreatorStats récupère les statistiques détaillées d'un créateur
func GetCreatorStats(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Vérifier que l'utilisateur est un créateur
	var user models.User
	if err := database.GetDB().First(&user, "id = ? AND role = ?", userID, models.RoleCreator).Error; err != nil {
		respondWithError(w, http.StatusForbidden, "Accès réservé aux créateurs")
		return
	}

	// Récupérer les statistiques de base
	var stats models.UserStats
	if err := database.GetDB().First(&stats, "user_id = ?", userID).Error; err != nil {
		stats = models.UserStats{UserID: userID}
		database.GetDB().Create(&stats)
	}

	// Récupérer les gains
	var earnings models.CreatorEarnings
	if err := database.GetDB().First(&earnings, "creator_id = ?", userID).Error; err != nil {
		earnings = models.CreatorEarnings{CreatorID: userID}
		database.GetDB().Create(&earnings)
	}

	// Statistiques détaillées
	var detailedStats struct {
		models.UserStats
		models.CreatorEarnings
		MessageStats struct {
			TotalSent        int64   `json:"total_sent"`
			PaidMessagesSent int64   `json:"paid_messages_sent"`
			MessageRevenue   float64 `json:"message_revenue"`
		} `json:"message_stats"`
		ConversationStats struct {
			ActiveConversations int64 `json:"active_conversations"`
			TotalConversations  int64 `json:"total_conversations"`
		} `json:"conversation_stats"`
	}

	detailedStats.UserStats = stats
	detailedStats.CreatorEarnings = earnings

	db := database.GetDB()

	// Statistiques de messages
	db.Model(&models.Message{}).Where("sender_id = ?", userID).Count(&detailedStats.MessageStats.TotalSent)
	db.Model(&models.Message{}).Where("sender_id = ? AND is_paid = true", userID).Count(&detailedStats.MessageStats.PaidMessagesSent)
	db.Model(&models.PaidMessageTransaction{}).Where("creator_id = ?", userID).Select("COALESCE(SUM(creator_amount), 0)").Scan(&detailedStats.MessageStats.MessageRevenue)

	// Statistiques de conversations
	db.Model(&models.Conversation{}).Where("(user1_id = ? OR user2_id = ?) AND last_message_at > NOW() - INTERVAL '7 days'", userID, userID).Count(&detailedStats.ConversationStats.ActiveConversations)
	db.Model(&models.Conversation{}).Where("user1_id = ? OR user2_id = ?", userID, userID).Count(&detailedStats.ConversationStats.TotalConversations)

	response := models.APIResponse{
		Success: true,
		Data:    detailedStats,
		Message: "Statistiques détaillées récupérées avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}
