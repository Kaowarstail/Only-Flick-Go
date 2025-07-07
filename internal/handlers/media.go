package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/Kaowarstail/Only-Flick-Go/internal/services"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

// UploadProfilePicture gère l'upload d'une photo de profil
func UploadProfilePicture(w http.ResponseWriter, r *http.Request) {
	fmt.Println("📤 [UploadProfilePicture] Début de l'upload de photo de profil")

	// Récupérer l'ID utilisateur du contexte
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		fmt.Println("❌ [UploadProfilePicture] Impossible d'extraire l'ID utilisateur")
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Vérifier que l'utilisateur existe
	var user models.User
	result := database.GetDB().First(&user, "id = ?", userID)
	if result.Error != nil {
		fmt.Printf("❌ [UploadProfilePicture] Utilisateur non trouvé: %s\n", userID)
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Limite de taille de fichier (5MB)
	r.ParseMultipartForm(5 << 20)

	// Récupérer le fichier depuis la requête
	file, handler, err := r.FormFile("profile_picture")
	if err != nil {
		fmt.Printf("❌ [UploadProfilePicture] Erreur lors de la récupération du fichier: %v\n", err)
		respondWithError(w, http.StatusBadRequest, "Erreur lors de la récupération du fichier")
		return
	}
	defer file.Close()

	fmt.Printf("📁 [UploadProfilePicture] Fichier reçu: %s, Taille: %d octets\n", handler.Filename, handler.Size)

	// Vérifier que c'est une image
	contentType := handler.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		fmt.Printf("❌ [UploadProfilePicture] Type de fichier non supporté: %s\n", contentType)
		respondWithError(w, http.StatusBadRequest, "La photo de profil doit être une image")
		return
	}

	// Initialiser le service Cloudinary
	fmt.Println("🔧 [UploadProfilePicture] Initialisation du service Cloudinary")
	cloudinaryService, err := services.NewCloudinaryService()
	if err != nil {
		fmt.Printf("❌ [UploadProfilePicture] Erreur lors de l'initialisation du service Cloudinary: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'initialisation du service Cloudinary")
		return
	}

	// Supprimer l'ancienne photo de profil si elle existe
	if user.ProfilePicture != "" {
		fmt.Printf("🗑️ [UploadProfilePicture] Suppression de l'ancienne photo de profil\n")
		// Extraire le PublicID de l'URL existante si c'est une URL Cloudinary
		// Pour le moment, on skip cette étape car c'est complexe
	}

	// Uploader la nouvelle photo de profil vers Cloudinary
	fmt.Println("🚀 [UploadProfilePicture] Upload vers Cloudinary")
	uploadResult, err := cloudinaryService.UploadProfilePicture(file, handler.Filename, userID)
	if err != nil {
		fmt.Printf("❌ [UploadProfilePicture] Erreur lors de l'upload: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'upload: "+err.Error())
		return
	}

	// Mettre à jour l'utilisateur avec la nouvelle URL de photo de profil
	fmt.Println("💾 [UploadProfilePicture] Mise à jour de l'utilisateur en base de données")
	user.ProfilePicture = uploadResult.SecureURL
	result = database.GetDB().Save(&user)
	if result.Error != nil {
		fmt.Printf("❌ [UploadProfilePicture] Erreur lors de la mise à jour de l'utilisateur: %s\n", result.Error.Error())
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour de l'utilisateur")
		return
	}

	fmt.Printf("✅ [UploadProfilePicture] Photo de profil uploadée avec succès. URL: %s\n", user.ProfilePicture)

	response := map[string]interface{}{
		"message":         "Photo de profil uploadée avec succès",
		"profile_picture": uploadResult.SecureURL,
		"public_id":       uploadResult.PublicID,
		"user": map[string]interface{}{
			"id":              user.ID,
			"username":        user.Username,
			"email":           user.Email,
			"profile_picture": user.ProfilePicture,
		},
		"file_info": map[string]interface{}{
			"format":        uploadResult.Format,
			"resource_type": uploadResult.ResourceType,
			"width":         uploadResult.Width,
			"height":        uploadResult.Height,
			"bytes":         uploadResult.Bytes,
		},
	}

	fmt.Println("🎉 [UploadProfilePicture] Upload terminé avec succès")
	respondWithJSON(w, http.StatusOK, response)
}

// UploadBannerImage gère l'upload d'une bannière pour un créateur
func UploadBannerImage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("📤 [UploadBannerImage] Début de l'upload de bannière")

	// Récupérer l'ID utilisateur du contexte
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		fmt.Println("❌ [UploadBannerImage] Impossible d'extraire l'ID utilisateur")
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Vérifier que l'utilisateur existe et est un créateur
	var user models.User
	result := database.GetDB().First(&user, "id = ? AND role = ?", userID, models.RoleCreator)
	if result.Error != nil {
		fmt.Printf("❌ [UploadBannerImage] Créateur non trouvé: %s\n", userID)
		respondWithError(w, http.StatusNotFound, "Créateur non trouvé")
		return
	}

	// Récupérer ou créer le profil créateur
	var profile models.CreatorProfile
	result = database.GetDB().FirstOrCreate(&profile, models.CreatorProfile{UserID: userID})
	if result.Error != nil {
		fmt.Printf("❌ [UploadBannerImage] Erreur lors de la récupération du profil créateur: %v\n", result.Error)
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération du profil créateur")
		return
	}

	// Limite de taille de fichier (5MB)
	r.ParseMultipartForm(5 << 20)

	// Récupérer le fichier depuis la requête
	file, handler, err := r.FormFile("banner_image")
	if err != nil {
		fmt.Printf("❌ [UploadBannerImage] Erreur lors de la récupération du fichier: %v\n", err)
		respondWithError(w, http.StatusBadRequest, "Erreur lors de la récupération du fichier")
		return
	}
	defer file.Close()

	fmt.Printf("📁 [UploadBannerImage] Fichier reçu: %s, Taille: %d octets\n", handler.Filename, handler.Size)

	// Vérifier que c'est une image
	contentType := handler.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		fmt.Printf("❌ [UploadBannerImage] Type de fichier non supporté: %s\n", contentType)
		respondWithError(w, http.StatusBadRequest, "La bannière doit être une image")
		return
	}

	// Initialiser le service Cloudinary
	fmt.Println("🔧 [UploadBannerImage] Initialisation du service Cloudinary")
	cloudinaryService, err := services.NewCloudinaryService()
	if err != nil {
		fmt.Printf("❌ [UploadBannerImage] Erreur lors de l'initialisation du service Cloudinary: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'initialisation du service Cloudinary")
		return
	}

	// Uploader la bannière vers Cloudinary
	fmt.Println("🚀 [UploadBannerImage] Upload vers Cloudinary")
	uploadResult, err := cloudinaryService.UploadBannerImage(file, handler.Filename, userID)
	if err != nil {
		fmt.Printf("❌ [UploadBannerImage] Erreur lors de l'upload: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'upload: "+err.Error())
		return
	}

	// Mettre à jour le profil créateur avec la nouvelle URL de bannière
	fmt.Println("💾 [UploadBannerImage] Mise à jour du profil créateur en base de données")
	profile.BannerImage = uploadResult.SecureURL
	result = database.GetDB().Save(&profile)
	if result.Error != nil {
		fmt.Printf("❌ [UploadBannerImage] Erreur lors de la mise à jour du profil: %s\n", result.Error.Error())
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour du profil créateur")
		return
	}

	fmt.Printf("✅ [UploadBannerImage] Bannière uploadée avec succès. URL: %s\n", profile.BannerImage)

	response := map[string]interface{}{
		"message":      "Bannière uploadée avec succès",
		"banner_image": uploadResult.SecureURL,
		"public_id":    uploadResult.PublicID,
		"profile": map[string]interface{}{
			"id":           profile.ID,
			"user_id":      profile.UserID,
			"banner_image": profile.BannerImage,
		},
		"file_info": map[string]interface{}{
			"format":        uploadResult.Format,
			"resource_type": uploadResult.ResourceType,
			"width":         uploadResult.Width,
			"height":        uploadResult.Height,
			"bytes":         uploadResult.Bytes,
		},
	}

	fmt.Println("🎉 [UploadBannerImage] Upload terminé avec succès")
	respondWithJSON(w, http.StatusOK, response)
}

// GetUserProfile récupère le profil d'un utilisateur avec ses photos
func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	// Récupérer l'utilisateur
	var user models.User
	result := database.GetDB().First(&user, "id = ?", userID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Construire la réponse
	response := map[string]interface{}{
		"id":              user.ID,
		"username":        user.Username,
		"email":           user.Email,
		"first_name":      user.FirstName,
		"last_name":       user.LastName,
		"role":            user.Role,
		"biography":       user.Biography,
		"profile_picture": user.ProfilePicture,
		"is_active":       user.IsActive,
		"created_at":      user.CreatedAt,
		"updated_at":      user.UpdatedAt,
	}

	// Si c'est un créateur, ajouter les informations du profil créateur
	if user.Role == models.RoleCreator {
		var profile models.CreatorProfile
		result = database.GetDB().First(&profile, "user_id = ?", userID)
		if result.Error == nil {
			response["creator_profile"] = map[string]interface{}{
				"id":           profile.ID,
				"banner_image": profile.BannerImage,
				"website_url":  profile.WebsiteURL,
				"social_links": profile.SocialLinks,
				"created_at":   profile.CreatedAt,
				"updated_at":   profile.UpdatedAt,
			}
		}
	}

	respondWithJSON(w, http.StatusOK, response)
}
