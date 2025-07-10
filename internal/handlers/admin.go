package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/Kaowarstail/Only-Flick-Go/models"
	"github.com/gorilla/mux"
)

// Fonction utilitaire pour la pagination
func getPaginationParams(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return page, limit
}

// Structures pour les réponses admin
type AdminContentResponse struct {
	ID                  uint      `json:"id"`
	CreatorID           string    `json:"creator_id"`
	CreatorName         string    `json:"creator_name"`
	CreatorUsername     string    `json:"creator_username"`
	CreatorProfileImage string    `json:"creator_profile_image"`
	Title               string    `json:"title"`
	Description         string    `json:"description"`
	Type                string    `json:"type"`
	MediaURL            string    `json:"media_url"`
	ThumbnailURL        string    `json:"thumbnail_url"`
	CoverURL            string    `json:"cover_url"`
	PublicID            string    `json:"public_id"`
	IsPremium           bool      `json:"is_premium"`
	IsPublished         bool      `json:"is_published"`
	ViewCount           int       `json:"view_count"`
	IsFlagged           bool      `json:"is_flagged"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	LikesCount          int       `json:"likes_count"`
	CommentsCount       int       `json:"comments_count"`
	ReportsCount        int       `json:"reports_count"`
}

type AdminContentsResponse struct {
	Contents []AdminContentResponse `json:"contents"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	Limit    int                    `json:"limit"`
}

type AdminUserResponse struct {
	ID              string    `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	Role            string    `json:"role"`
	IsActive        bool      `json:"is_active"`
	IsBanned        bool      `json:"is_banned"`
	BanReason       string    `json:"ban_reason"`
	IsEmailVerified bool      `json:"is_email_verified"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AdminUsersResponse struct {
	Users []AdminUserResponse `json:"users"`
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Limit int                 `json:"limit"`
}

type ReportResponse struct {
	ID           uint      `json:"id"`
	ContentID    uint      `json:"content_id"`
	ContentTitle string    `json:"content_title"`
	ReporterID   string    `json:"reporter_id"`
	ReporterName string    `json:"reporter_name"`
	Reason       string    `json:"reason"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type AdminContentDetails struct {
	AdminContentResponse
	Comments []AdminCommentResponse `json:"comments"`
	Reports  []AdminReportResponse  `json:"reports"`
	Stats    map[string]interface{} `json:"stats"`
}

type AdminCommentResponse struct {
	ID        uint      `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	IsFlagged bool      `json:"is_flagged"`
	CreatedAt time.Time `json:"created_at"`
}

type AdminReportResponse struct {
	ID           uint      `json:"id"`
	ReporterID   string    `json:"reporter_id"`
	ReporterName string    `json:"reporter_name"`
	Reason       string    `json:"reason"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// GetDashboardStats récupère les statistiques du dashboard admin
func GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7)
	monthAgo := now.AddDate(0, -1, 0)

	// 1. Statistiques générales
	var totalUsers, totalCreators, totalSubscribers int64
	db.Model(&models.User{}).Count(&totalUsers)
	db.Model(&models.User{}).Where("role = ?", models.RoleCreator).Count(&totalCreators)
	db.Model(&models.User{}).Where("role = ?", models.RoleSubscriber).Count(&totalSubscribers)

	// 2. Revenus
	var totalRevenue, monthlyRevenue, weeklyRevenue float64
	db.Model(&models.Transaction{}).Where("status = ?", "success").Select("COALESCE(SUM(amount), 0)").Scan(&totalRevenue)
	db.Model(&models.Transaction{}).Where("status = ? AND created_at >= ?", "success", monthAgo).Select("COALESCE(SUM(amount), 0)").Scan(&monthlyRevenue)
	db.Model(&models.Transaction{}).Where("status = ? AND created_at >= ?", "success", weekAgo).Select("COALESCE(SUM(amount), 0)").Scan(&weeklyRevenue)

	// 3. Contenus
	var totalContents int64
	db.Model(&models.Content{}).Count(&totalContents)

	// 4. Nouveaux utilisateurs
	var newUsersWeek, newUsersMonth int64
	db.Model(&models.User{}).Where("created_at >= ?", weekAgo).Count(&newUsersWeek)
	db.Model(&models.User{}).Where("created_at >= ?", monthAgo).Count(&newUsersMonth)

	// 5. Signalements
	var totalReports, pendingReports, resolvedReports, reportsToday, reportsWeek int64
	db.Model(&models.Report{}).Count(&totalReports)
	db.Model(&models.Report{}).Where("status = ?", "pending").Count(&pendingReports)
	db.Model(&models.Report{}).Where("status = ?", "resolved").Count(&resolvedReports)
	db.Model(&models.Report{}).Where("created_at >= ?", now.Truncate(24*time.Hour)).Count(&reportsToday)
	db.Model(&models.Report{}).Where("created_at >= ?", weekAgo).Count(&reportsWeek)

	// 6. Statistiques de contenu par période
	var contentsToday, contentsWeek, contentsMonth int64
	var freeContents, premiumContents int64
	db.Model(&models.Content{}).Where("created_at >= ?", now.Truncate(24*time.Hour)).Count(&contentsToday)
	db.Model(&models.Content{}).Where("created_at >= ?", weekAgo).Count(&contentsWeek)
	db.Model(&models.Content{}).Where("created_at >= ?", monthAgo).Count(&contentsMonth)
	db.Model(&models.Content{}).Where("is_premium = ?", false).Count(&freeContents)
	db.Model(&models.Content{}).Where("is_premium = ?", true).Count(&premiumContents)

	// Construire la réponse complète avec toutes les stats
	response := map[string]interface{}{
		"overview": models.DashboardStats{
			TotalUsers:       int(totalUsers),
			TotalCreators:    int(totalCreators),
			TotalSubscribers: int(totalSubscribers),
			TotalRevenue:     totalRevenue,
			MonthlyRevenue:   monthlyRevenue,
			WeeklyRevenue:    weeklyRevenue,
			TotalContents:    int(totalContents),
			NewUsersWeek:     int(newUsersWeek),
			NewUsersMonth:    int(newUsersMonth),
			PendingReports:   int(pendingReports),
		},
		"content_stats": map[string]interface{}{
			"total_contents":   int(totalContents),
			"free_contents":    int(freeContents),
			"premium_contents": int(premiumContents),
			"contents_today":   int(contentsToday),
			"contents_week":    int(contentsWeek),
			"contents_month":   int(contentsMonth),
		},
		"report_stats": map[string]interface{}{
			"total_reports":    int(totalReports),
			"pending_reports":  int(pendingReports),
			"resolved_reports": int(resolvedReports),
			"reports_today":    int(reportsToday),
			"reports_week":     int(reportsWeek),
		},
		"generated_at": now,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetAdminUsers récupère la liste des utilisateurs pour l'admin
func GetAdminUsers(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()

	// Pagination
	page, limit := getPaginationParams(r)
	offset := (page - 1) * limit

	// Filtres
	search := r.URL.Query().Get("search")
	role := r.URL.Query().Get("role")
	status := r.URL.Query().Get("status")

	// Construire la requête
	query := db.Model(&models.User{})

	if search != "" {
		query = query.Where("username ILIKE ? OR email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if role != "" {
		query = query.Where("role = ?", role)
	}

	if status == "active" {
		query = query.Where("is_active = ? AND is_banned = ?", true, false)
	} else if status == "banned" {
		query = query.Where("is_banned = ?", true)
	} else if status == "inactive" {
		query = query.Where("is_active = ?", false)
	}

	// Compter le total
	var total int64
	query.Count(&total)

	// Récupérer les utilisateurs avec pagination
	var users []models.User
	query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users)

	// Convertir en réponse
	var adminUsers []AdminUserResponse
	for _, user := range users {
		adminUsers = append(adminUsers, AdminUserResponse{
			ID:              user.ID,
			Username:        user.Username,
			Email:           user.Email,
			FirstName:       user.FirstName,
			LastName:        user.LastName,
			Role:            string(user.Role),
			IsActive:        user.IsActive,
			IsBanned:        user.IsBanned,
			BanReason:       user.BanReason,
			IsEmailVerified: user.IsEmailVerified,
			CreatedAt:       user.CreatedAt,
			UpdatedAt:       user.UpdatedAt,
		})
	}

	response := AdminUsersResponse{
		Users: adminUsers,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetAdminContents récupère la liste des contenus pour l'admin
func GetAdminContents(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()

	// Pagination
	page, limit := getPaginationParams(r)
	offset := (page - 1) * limit

	// Filtres
	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")
	contentType := r.URL.Query().Get("type")
	creatorID := r.URL.Query().Get("creator_id")

	// Construire la requête
	query := db.Model(&models.Content{}).Preload("Creator")

	if search != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if status == "published" {
		query = query.Where("is_published = ?", true)
	} else if status == "unpublished" {
		query = query.Where("is_published = ?", false)
	} else if status == "flagged" {
		query = query.Where("is_flagged = ?", true)
	}

	if contentType != "" {
		query = query.Where("type = ?", contentType)
	}

	if creatorID != "" {
		query = query.Where("creator_id = ?", creatorID)
	}

	// Compter le total
	var total int64
	query.Count(&total)

	// Récupérer les contenus avec pagination
	var contents []models.Content
	query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&contents)

	// Convertir en réponse
	var adminContents []AdminContentResponse
	for _, content := range contents {
		// Compter les likes, commentaires et signalements
		var likesCount, commentsCount, reportsCount int64
		db.Model(&models.Like{}).Where("content_id = ?", content.ID).Count(&likesCount)
		db.Model(&models.Comment{}).Where("content_id = ?", content.ID).Count(&commentsCount)
		db.Model(&models.Report{}).Where("content_id = ?", content.ID).Count(&reportsCount)

		// Générer le nom complet du créateur
		creatorName := content.Creator.FirstName + " " + content.Creator.LastName
		if creatorName == " " {
			creatorName = content.Creator.Username
		}

		adminContents = append(adminContents, AdminContentResponse{
			ID:                  content.ID,
			CreatorID:           content.CreatorID,
			CreatorName:         creatorName,
			CreatorUsername:     content.Creator.Username,
			CreatorProfileImage: content.Creator.ProfilePicture,
			Title:               content.Title,
			Description:         content.Description,
			Type:                content.Type,
			MediaURL:            content.MediaURL,
			ThumbnailURL:        content.ThumbnailURL,
			CoverURL:            content.CoverURL,
			PublicID:            content.PublicID,
			IsPremium:           content.IsPremium,
			IsPublished:         content.IsPublished,
			ViewCount:           content.ViewCount,
			IsFlagged:           content.IsFlagged,
			CreatedAt:           content.CreatedAt,
			UpdatedAt:           content.UpdatedAt,
			LikesCount:          int(likesCount),
			CommentsCount:       int(commentsCount),
			ReportsCount:        int(reportsCount),
		})
	}

	response := AdminContentsResponse{
		Contents: adminContents,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetRecentReports récupère les signalements récents
func GetRecentReports(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()

	// Pagination
	page, limit := getPaginationParams(r)
	offset := (page - 1) * limit

	// Récupérer les signalements avec les relations
	var reports []models.Report
	db.Preload("Content").Preload("Reporter").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&reports)

	// Convertir en réponse
	var reportResponses []ReportResponse
	for _, report := range reports {
		reporterName := report.Reporter.FirstName + " " + report.Reporter.LastName
		if reporterName == " " {
			reporterName = report.Reporter.Username
		}

		reportResponses = append(reportResponses, ReportResponse{
			ID:           report.ID,
			ContentID:    report.ContentID,
			ContentTitle: report.Content.Title,
			ReporterID:   report.ReporterID,
			ReporterName: reporterName,
			Reason:       report.Reason,
			Status:       report.Status,
			CreatedAt:    report.CreatedAt,
		})
	}

	respondWithJSON(w, http.StatusOK, reportResponses)
}

// UpdateUserRole met à jour le rôle d'un utilisateur
func UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}

	// Vérifier que le rôle est valide
	if requestBody.Role != string(models.RoleAdmin) &&
		requestBody.Role != string(models.RoleCreator) &&
		requestBody.Role != string(models.RoleSubscriber) {
		respondWithError(w, http.StatusBadRequest, "Rôle invalide")
		return
	}

	db := database.GetDB()
	result := db.Model(&models.User{}).Where("id = ?", requestBody.UserID).Update("role", requestBody.Role)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour")
		return
	}

	if result.RowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Rôle mis à jour avec succès"})
}

// BanUser bannit un utilisateur
func BanUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var requestBody struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}

	db := database.GetDB()
	now := time.Now()

	updates := map[string]interface{}{
		"is_banned":  true,
		"ban_reason": requestBody.Reason,
		"banned_at":  &now,
	}

	result := db.Model(&models.User{}).Where("id = ?", userID).Updates(updates)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors du bannissement")
		return
	}

	if result.RowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Utilisateur banni avec succès"})
}

// UnbanUser débannit un utilisateur
func UnbanUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	db := database.GetDB()

	updates := map[string]interface{}{
		"is_banned":  false,
		"ban_reason": "",
		"banned_at":  nil,
	}

	result := db.Model(&models.User{}).Where("id = ?", userID).Updates(updates)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors du débannissement")
		return
	}

	if result.RowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Utilisateur débanni avec succès"})
}

// UpdateReportStatus met à jour le statut d'un signalement
func UpdateReportStatus(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		ReportID uint   `json:"report_id"`
		Status   string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}

	// Vérifier que le statut est valide
	if requestBody.Status != "pending" && requestBody.Status != "reviewed" && requestBody.Status != "dismissed" {
		respondWithError(w, http.StatusBadRequest, "Statut invalide")
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Non autorisé")
		return
	}

	db := database.GetDB()
	now := time.Now()

	updates := map[string]interface{}{
		"status":      requestBody.Status,
		"reviewed_by": userID,
		"reviewed_at": &now,
	}

	result := db.Model(&models.Report{}).Where("id = ?", requestBody.ReportID).Updates(updates)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour")
		return
	}

	if result.RowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Signalement non trouvé")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Statut mis à jour avec succès"})
}

// UpdateUserStatus met à jour le statut d'un utilisateur (ban/unban)
func UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		UserID   string `json:"user_id"`
		IsBanned bool   `json:"is_banned"`
		Reason   string `json:"reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}

	db := database.GetDB()

	var updates map[string]interface{}

	if requestBody.IsBanned {
		now := time.Now()
		updates = map[string]interface{}{
			"is_banned":  true,
			"ban_reason": requestBody.Reason,
			"banned_at":  &now,
		}
	} else {
		updates = map[string]interface{}{
			"is_banned":  false,
			"ban_reason": "",
			"banned_at":  nil,
		}
	}

	result := db.Model(&models.User{}).Where("id = ?", requestBody.UserID).Updates(updates)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour du statut")
		return
	}

	if result.RowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	var message string
	if requestBody.IsBanned {
		message = "Utilisateur banni avec succès"
	} else {
		message = "Utilisateur débanni avec succès"
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": message})
}

// DeleteUser supprime un utilisateur
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var requestBody struct {
		Reason string `json:"reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		// Si pas de body, on continue quand même
		requestBody.Reason = ""
	}

	db := database.GetDB()
	result := db.Delete(&models.User{}, "id = ?", userID)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la suppression")
		return
	}

	if result.RowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Utilisateur supprimé avec succès"})
}

// GetUserDetails récupère les détails d'un utilisateur pour l'admin
func GetUserDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		respondWithError(w, http.StatusBadRequest, "ID utilisateur requis")
		return
	}

	db := database.GetDB()
	var user models.User
	result := db.First(&user, "id = ?", userID)

	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Convertir en réponse admin
	userResponse := AdminUserResponse{
		ID:              user.ID,
		Username:        user.Username,
		Email:           user.Email,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		Role:            string(user.Role),
		IsActive:        user.IsActive,
		IsBanned:        user.IsBanned,
		BanReason:       user.BanReason,
		IsEmailVerified: user.IsEmailVerified,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}

	respondWithJSON(w, http.StatusOK, userResponse)
}

// GetAdminContentDetails récupère les détails d'un contenu pour l'admin
func GetAdminContentDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentID := vars["id"]

	if contentID == "" {
		respondWithError(w, http.StatusBadRequest, "ID contenu requis")
		return
	}

	db := database.GetDB()
	var content models.Content
	result := db.Preload("Creator").Preload("Comments.User").First(&content, "id = ?", contentID)

	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	// Compter les likes, commentaires et signalements
	var likesCount, commentsCount, reportsCount int64
	db.Model(&models.Like{}).Where("content_id = ?", content.ID).Count(&likesCount)
	db.Model(&models.Comment{}).Where("content_id = ?", content.ID).Count(&commentsCount)
	db.Model(&models.Report{}).Where("content_id = ?", content.ID).Count(&reportsCount)

	// Générer le nom complet du créateur
	creatorName := content.Creator.FirstName + " " + content.Creator.LastName
	if creatorName == " " {
		creatorName = content.Creator.Username
	}

	// Récupérer les commentaires récents
	var comments []models.Comment
	db.Preload("User").Where("content_id = ?", content.ID).Order("created_at DESC").Limit(10).Find(&comments)

	// Récupérer les signalements
	var reports []models.Report
	db.Preload("Reporter").Where("content_id = ?", content.ID).Order("created_at DESC").Find(&reports)

	// Convertir les commentaires
	var commentResponses []AdminCommentResponse
	for _, comment := range comments {
		commentResponses = append(commentResponses, AdminCommentResponse{
			ID:        comment.ID,
			UserID:    comment.UserID,
			Username:  comment.User.Username,
			Content:   comment.Text,
			IsFlagged: comment.IsHidden,
			CreatedAt: comment.CreatedAt,
		})
	}

	// Convertir les signalements
	var reportResponses []AdminReportResponse
	for _, report := range reports {
		reporterName := report.Reporter.FirstName + " " + report.Reporter.LastName
		if reporterName == " " {
			reporterName = report.Reporter.Username
		}

		reportResponses = append(reportResponses, AdminReportResponse{
			ID:           report.ID,
			ReporterID:   report.ReporterID,
			ReporterName: reporterName,
			Reason:       report.Reason,
			Status:       report.Status,
			CreatedAt:    report.CreatedAt,
		})
	}

	// Construire la réponse détaillée
	contentDetails := AdminContentDetails{
		AdminContentResponse: AdminContentResponse{
			ID:                  content.ID,
			CreatorID:           content.CreatorID,
			CreatorName:         creatorName,
			CreatorUsername:     content.Creator.Username,
			CreatorProfileImage: content.Creator.ProfilePicture,
			Title:               content.Title,
			Description:         content.Description,
			Type:                content.Type,
			MediaURL:            content.MediaURL,
			ThumbnailURL:        content.ThumbnailURL,
			CoverURL:            content.CoverURL,
			PublicID:            content.PublicID,
			IsPremium:           content.IsPremium,
			IsPublished:         content.IsPublished,
			ViewCount:           content.ViewCount,
			IsFlagged:           content.IsFlagged,
			CreatedAt:           content.CreatedAt,
			UpdatedAt:           content.UpdatedAt,
			LikesCount:          int(likesCount),
			CommentsCount:       int(commentsCount),
			ReportsCount:        int(reportsCount),
		},
		Comments: commentResponses,
		Reports:  reportResponses,
		Stats: map[string]interface{}{
			"engagement_rate": float64(likesCount+commentsCount) / float64(content.ViewCount+1) * 100,
		},
	}

	respondWithJSON(w, http.StatusOK, contentDetails)
}

// UpdateAdminContent met à jour un contenu depuis l'interface admin
func UpdateAdminContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentID := vars["id"]

	if contentID == "" {
		respondWithError(w, http.StatusBadRequest, "ID contenu requis")
		return
	}

	var requestBody struct {
		Title       *string `json:"title,omitempty"`
		Description *string `json:"description,omitempty"`
		IsPremium   *bool   `json:"is_premium,omitempty"`
		IsPublished *bool   `json:"is_published,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}

	db := database.GetDB()
	var content models.Content
	result := db.First(&content, "id = ?", contentID)

	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	// Appliquer les mises à jour
	updates := map[string]interface{}{}

	if requestBody.Title != nil {
		updates["title"] = *requestBody.Title
	}
	if requestBody.Description != nil {
		updates["description"] = *requestBody.Description
	}
	if requestBody.IsPremium != nil {
		updates["is_premium"] = *requestBody.IsPremium
	}
	if requestBody.IsPublished != nil {
		updates["is_published"] = *requestBody.IsPublished
	}

	if len(updates) > 0 {
		result = db.Model(&content).Updates(updates)
		if result.Error != nil {
			respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour")
			return
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Contenu mis à jour avec succès"})
}

// DeleteAdminContent supprime un contenu depuis l'interface admin
func DeleteAdminContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentID := vars["id"]

	if contentID == "" {
		respondWithError(w, http.StatusBadRequest, "ID contenu requis")
		return
	}

	var requestBody struct {
		Reason string `json:"reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		// Si pas de body, on continue quand même
		requestBody.Reason = ""
	}

	db := database.GetDB()
	var content models.Content
	result := db.First(&content, "id = ?", contentID)

	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	// Supprimer le contenu (soft delete)
	result = db.Delete(&content)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la suppression")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Contenu supprimé avec succès"})
}

// UpdateContentStatus met à jour le statut d'un contenu (approuver/rejeter)
func UpdateContentStatus(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		ContentID string `json:"content_id"`
		Status    string `json:"status"`
		Reason    string `json:"reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}

	// Vérifier que le statut est valide
	if requestBody.Status != "approved" && requestBody.Status != "rejected" && requestBody.Status != "flagged" {
		respondWithError(w, http.StatusBadRequest, "Statut invalide")
		return
	}

	db := database.GetDB()

	updates := map[string]interface{}{}

	switch requestBody.Status {
	case "approved":
		updates["is_published"] = true
		updates["is_flagged"] = false
	case "rejected":
		updates["is_published"] = false
		updates["is_flagged"] = false
	case "flagged":
		updates["is_flagged"] = true
		updates["is_published"] = false
	}

	result := db.Model(&models.Content{}).Where("id = ?", requestBody.ContentID).Updates(updates)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour")
		return
	}

	if result.RowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Statut mis à jour avec succès"})
}

// FlagContent marque/démarque un contenu comme inapproprié
func FlagContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentID := vars["id"]

	if contentID == "" {
		respondWithError(w, http.StatusBadRequest, "ID contenu requis")
		return
	}

	var requestBody struct {
		IsFlagged bool   `json:"is_flagged"`
		Reason    string `json:"reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}

	db := database.GetDB()

	updates := map[string]interface{}{
		"is_flagged": requestBody.IsFlagged,
	}

	// Si on flagge le contenu, on le dépublie aussi
	if requestBody.IsFlagged {
		updates["is_published"] = false
	}

	result := db.Model(&models.Content{}).Where("id = ?", contentID).Updates(updates)

	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour")
		return
	}

	if result.RowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Contenu non trouvé")
		return
	}

	message := "Contenu déflaggé avec succès"
	if requestBody.IsFlagged {
		message = "Contenu flaggé avec succès"
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": message})
}
