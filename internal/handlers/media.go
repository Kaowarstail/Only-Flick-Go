package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
	"github.com/Kaowarstail/Only-Flick-Go/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// MediaHandler gère les requêtes liées aux médias
type MediaHandler struct {
	uploadPath string
	maxSize    int64
}

// NewMediaHandler crée un nouveau handler pour les médias
func NewMediaHandler(uploadPath string, maxSize int64) *MediaHandler {
	return &MediaHandler{
		uploadPath: uploadPath,
		maxSize:    maxSize,
	}
}

// getUserIDFromContext récupère l'ID utilisateur depuis le contexte
func getUserIDFromContext(ctx context.Context) string {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return ""
	}
	return userID
}

// UploadMedia gère l'upload d'un fichier média
func (h *MediaHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Limiter la taille du fichier
	r.Body = http.MaxBytesReader(w, r.Body, h.maxSize)

	// Parser le multipart form
	if err := r.ParseMultipartForm(h.maxSize); err != nil {
		respondWithError(w, http.StatusBadRequest, "Fichier trop volumineux")
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Fichier requis")
		return
	}
	defer file.Close()

	// Vérifier le type de fichier
	mediaType := h.getMediaType(handler.Filename)
	if mediaType == "" {
		respondWithError(w, http.StatusBadRequest, "Type de fichier non supporté")
		return
	}

	// Générer un nom de fichier unique
	fileID := uuid.New().String()
	fileExt := filepath.Ext(handler.Filename)
	fileName := fmt.Sprintf("%s%s", fileID, fileExt)

	// Créer le dossier de destination s'il n'existe pas
	uploadDir := filepath.Join(h.uploadPath, "messages")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la création du dossier")
		return
	}

	// Créer le fichier de destination
	filePath := filepath.Join(uploadDir, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la création du fichier")
		return
	}
	defer dst.Close()

	// Copier le fichier
	if _, err := io.Copy(dst, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la sauvegarde du fichier")
		return
	}

	// Construire l'URL du fichier
	fileURL := fmt.Sprintf("/uploads/messages/%s", fileName)

	// Sauvegarder les informations du fichier en base de données
	mediaFile := models.MediaFile{
		ID:           fileID,
		UserID:       userID,
		FileName:     handler.Filename,
		FilePath:     filePath,
		FileURL:      fileURL,
		FileSize:     handler.Size,
		MediaType:    models.MediaType(mediaType),
		MimeType:     handler.Header.Get("Content-Type"),
		UploadedAt:   time.Now(),
	}

	if err := database.GetDB().Create(&mediaFile).Error; err != nil {
		// Supprimer le fichier si la sauvegarde en base échoue
		os.Remove(filePath)
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'enregistrement du fichier")
		return
	}

	response := models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"file_id":    fileID,
			"file_url":   fileURL,
			"file_name":  handler.Filename,
			"file_size":  handler.Size,
			"media_type": mediaType,
			"mime_type":  handler.Header.Get("Content-Type"),
		},
		Message: "Fichier uploadé avec succès",
	}

	respondWithJSON(w, http.StatusCreated, response)
}

// GetMedia récupère les métadonnées d'un fichier média
func (h *MediaHandler) GetMedia(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]

	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	var mediaFile models.MediaFile
	if err := database.GetDB().First(&mediaFile, "id = ?", fileID).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Fichier non trouvé")
		return
	}

	// Vérifier les permissions (propriétaire du fichier ou participant à la conversation)
	if mediaFile.UserID != userID {
		// Vérifier si l'utilisateur a accès via une conversation
		var message models.Message
		if err := database.GetDB().First(&message, "media_url LIKE ?", "%"+fileID+"%").Error; err != nil {
			respondWithError(w, http.StatusForbidden, "Accès non autorisé")
			return
		}

		// Vérifier si l'utilisateur est participant à la conversation
		var conversation models.Conversation
		if err := database.GetDB().First(&conversation, "id = ? AND (user1_id = ? OR user2_id = ?)", 
			message.ConversationID, userID, userID).Error; err != nil {
			respondWithError(w, http.StatusForbidden, "Accès non autorisé")
			return
		}
	}

	response := models.APIResponse{
		Success: true,
		Data:    mediaFile,
		Message: "Métadonnées récupérées avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// ServeMedia sert le fichier média
func (h *MediaHandler) ServeMedia(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileName := vars["filename"]

	// Construire le chemin du fichier
	filePath := filepath.Join(h.uploadPath, "messages", fileName)

	// Vérifier que le fichier existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		respondWithError(w, http.StatusNotFound, "Fichier non trouvé")
		return
	}

	// Servir le fichier
	http.ServeFile(w, r, filePath)
}

// DeleteMedia supprime un fichier média
func (h *MediaHandler) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]

	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	var mediaFile models.MediaFile
	if err := database.GetDB().First(&mediaFile, "id = ? AND user_id = ?", fileID, userID).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Fichier non trouvé ou non autorisé")
		return
	}

	// Supprimer le fichier du système de fichiers
	if err := os.Remove(mediaFile.FilePath); err != nil && !os.IsNotExist(err) {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la suppression du fichier")
		return
	}

	// Supprimer l'enregistrement de la base de données
	if err := database.GetDB().Delete(&mediaFile).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la suppression en base de données")
		return
	}

	response := models.APIResponse{
		Success: true,
		Message: "Fichier supprimé avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetUserMediaFiles récupère tous les fichiers média d'un utilisateur
func (h *MediaHandler) GetUserMediaFiles(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	var mediaFiles []models.MediaFile
	if err := database.GetDB().Where("user_id = ?", userID).Order("uploaded_at DESC").Find(&mediaFiles).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des fichiers")
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    mediaFiles,
		Message: "Fichiers récupérés avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// getMediaType détermine le type de média basé sur l'extension du fichier
func (h *MediaHandler) getMediaType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return "image"
	case ".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm":
		return "video"
	case ".mp3", ".wav", ".ogg", ".m4a", ".aac":
		return "audio"
	case ".pdf":
		return "document"
	default:
		return ""
	}
}

// GetMediaStats récupère les statistiques d'utilisation des médias
func (h *MediaHandler) GetMediaStats(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	var stats struct {
		TotalFiles    int64   `json:"total_files"`
		TotalSize     int64   `json:"total_size"`
		ImageCount    int64   `json:"image_count"`
		VideoCount    int64   `json:"video_count"`
		AudioCount    int64   `json:"audio_count"`
		DocumentCount int64   `json:"document_count"`
		AverageSize   float64 `json:"average_size"`
	}

	db := database.GetDB()

	// Total des fichiers
	db.Model(&models.MediaFile{}).Where("user_id = ?", userID).Count(&stats.TotalFiles)

	// Taille totale
	db.Model(&models.MediaFile{}).Where("user_id = ?", userID).Select("COALESCE(SUM(file_size), 0)").Scan(&stats.TotalSize)

	// Comptage par type
	db.Model(&models.MediaFile{}).Where("user_id = ? AND media_type = ?", userID, "image").Count(&stats.ImageCount)
	db.Model(&models.MediaFile{}).Where("user_id = ? AND media_type = ?", userID, "video").Count(&stats.VideoCount)
	db.Model(&models.MediaFile{}).Where("user_id = ? AND media_type = ?", userID, "audio").Count(&stats.AudioCount)
	db.Model(&models.MediaFile{}).Where("user_id = ? AND media_type = ?", userID, "document").Count(&stats.DocumentCount)

	// Taille moyenne
	if stats.TotalFiles > 0 {
		stats.AverageSize = float64(stats.TotalSize) / float64(stats.TotalFiles)
	}

	response := models.APIResponse{
		Success: true,
		Data:    stats,
		Message: "Statistiques récupérées avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}
