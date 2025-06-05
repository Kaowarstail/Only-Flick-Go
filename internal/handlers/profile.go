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
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID utilisateur invalide")
		return
	}

	// Vérifier si l'utilisateur connecté a le droit de modifier ce profil
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || uint(id) != userID {
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
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID utilisateur invalide")
		return
	}

	// Vérifier si l'utilisateur existe
	var user models.User
	result := database.GetDB().First(&user, id)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// TODO: Implémenter le modèle de relation "following" et récupérer la liste
	// Pour l'instant, retournons une liste vide
	following := []map[string]interface{}{}

	respondWithJSON(w, http.StatusOK, following)
}

// BlockUser bloque un utilisateur
func BlockUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID utilisateur invalide")
		return
	}
	targetID, err := strconv.Atoi(vars["targetId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID cible invalide")
		return
	}

	// On va utiliser targetID lors de l'implémentation complète
	// Pour l'instant, on le marque comme utilisé pour éviter l'erreur de compilation
	_ = targetID

	// Vérifier si l'utilisateur connecté est bien l'utilisateur qui fait la demande
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || uint(userID) != currentUserID {
		respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à effectuer cette action")
		return
	}

	// TODO: Implémenter le modèle de relation "blocked" et ajouter la relation
	// Pour l'instant, simulons une opération réussie

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Utilisateur bloqué avec succès",
	})
}

// UnblockUser débloque un utilisateur
func UnblockUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID utilisateur invalide")
		return
	}
	targetID, err := strconv.Atoi(vars["targetId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID cible invalide")
		return
	}

	// On va utiliser targetID lors de l'implémentation complète
	// Pour l'instant, on le marque comme utilisé pour éviter l'erreur de compilation
	_ = targetID

	// Vérifier si l'utilisateur connecté est bien l'utilisateur qui fait la demande
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || uint(userID) != currentUserID {
		respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à effectuer cette action")
		return
	}

	// TODO: Implémenter le modèle de relation "blocked" et supprimer la relation
	// Pour l'instant, simulons une opération réussie

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Utilisateur débloqué avec succès",
	})
}

// GetBlockedUsers récupère la liste des utilisateurs bloqués par un utilisateur
func GetBlockedUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID utilisateur invalide")
		return
	}

	// Vérifier si l'utilisateur connecté est bien l'utilisateur qui fait la demande
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || uint(id) != currentUserID {
		respondWithError(w, http.StatusForbidden, "Vous n'êtes pas autorisé à effectuer cette action")
		return
	}

	// TODO: Implémenter le modèle de relation "blocked" et récupérer la liste
	// Pour l'instant, retournons une liste vide
	blockedUsers := []map[string]interface{}{}

	respondWithJSON(w, http.StatusOK, blockedUsers)
}

// UpdateNotificationSettings met à jour les paramètres de notification d'un utilisateur
func UpdateNotificationSettings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID utilisateur invalide")
		return
	}

	// Vérifier si l'utilisateur connecté est bien l'utilisateur qui fait la demande
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || uint(id) != currentUserID {
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

	// TODO: Implémenter le modèle de notification_settings et mettre à jour les paramètres
	// Pour l'instant, simulons une opération réussie

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Paramètres de notification mis à jour avec succès",
	})
}

// GetFeed récupère le flux d'activités d'un utilisateur
func GetFeed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID utilisateur invalide")
		return
	}

	// Vérifier si l'utilisateur connecté est bien l'utilisateur qui fait la demande
	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || uint(id) != currentUserID {
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

	// TODO: Implémenter la logique de récupération du flux d'activités
	// Pour l'instant, retournons des données fictives
	feedItems := []map[string]interface{}{
		{
			"id":   1,
			"type": "content",
			"creator": map[string]interface{}{
				"id":              2,
				"username":        "creator1",
				"profile_picture": "/uploads/profiles/creator1.jpg",
			},
			"content": map[string]interface{}{
				"id":            101,
				"title":         "Nouveau contenu",
				"preview_image": "/uploads/content/preview101.jpg",
				"created_at":    "2023-06-01T12:00:00Z",
			},
		},
		{
			"id":   2,
			"type": "follow",
			"user": map[string]interface{}{
				"id":              3,
				"username":        "fan123",
				"profile_picture": "/uploads/profiles/fan123.jpg",
			},
			"action":     "a commencé à vous suivre",
			"created_at": "2023-06-02T15:30:00Z",
		},
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"items": feedItems,
		"pagination": map[string]int{
			"page":        page,
			"size":        pageSize,
			"total_items": 2,
			"total_pages": 1,
		},
	})
}
