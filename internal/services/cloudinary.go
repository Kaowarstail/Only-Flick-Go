package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/config"
	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService struct {
	client *cloudinary.Cloudinary
}

type UploadResult struct {
	PublicID     string `json:"public_id"`
	SecureURL    string `json:"secure_url"`
	URL          string `json:"url"`
	Format       string `json:"format"`
	ResourceType string `json:"resource_type"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Bytes        int64  `json:"bytes"`
}

// NewCloudinaryService crée une nouvelle instance du service Cloudinary
func NewCloudinaryService() (*CloudinaryService, error) {
	fmt.Println("🔧 [Cloudinary] Initialisation du service Cloudinary")

	cfg := config.Get()

	// Vérifier les paramètres de configuration
	if cfg.Cloudinary.CloudName == "" {
		fmt.Println("❌ [Cloudinary] CloudName manquant dans la configuration")
		return nil, fmt.Errorf("CloudName manquant dans la configuration")
	}
	if cfg.Cloudinary.APIKey == "" {
		fmt.Println("❌ [Cloudinary] APIKey manquant dans la configuration")
		return nil, fmt.Errorf("APIKey manquant dans la configuration")
	}
	if cfg.Cloudinary.APISecret == "" {
		fmt.Println("❌ [Cloudinary] APISecret manquant dans la configuration")
		return nil, fmt.Errorf("APISecret manquant dans la configuration")
	}

	fmt.Printf("✅ [Cloudinary] Configuration chargée - CloudName: %s, APIKey: %s***\n", cfg.Cloudinary.CloudName, cfg.Cloudinary.APIKey[:4])

	// Créer le client Cloudinary
	cld, err := cloudinary.NewFromParams(
		cfg.Cloudinary.CloudName,
		cfg.Cloudinary.APIKey,
		cfg.Cloudinary.APISecret,
	)
	if err != nil {
		fmt.Printf("❌ [Cloudinary] Erreur lors de la création du client Cloudinary: %v\n", err)
		return nil, fmt.Errorf("erreur lors de la création du client Cloudinary: %v", err)
	}

	fmt.Println("✅ [Cloudinary] Client Cloudinary créé avec succès")
	return &CloudinaryService{
		client: cld,
	}, nil
}

// UploadImage uploade une image sur Cloudinary
func (cs *CloudinaryService) UploadImage(file multipart.File, filename string, contentID string) (*UploadResult, error) {
	fmt.Printf("📤 [Cloudinary] Début de l'upload d'image - Filename: %s, ContentID: %s\n", filename, contentID)

	// Générer un nom de fichier unique
	uuid, err := utils.GenerateUUID()
	if err != nil {
		fmt.Printf("❌ [Cloudinary] Erreur lors de la génération de l'UUID: %v\n", err)
		return nil, fmt.Errorf("erreur lors de la génération de l'UUID: %v", err)
	}
	fmt.Printf("✅ [Cloudinary] UUID généré: %s\n", uuid)

	// Créer le public ID avec un préfixe pour organiser les fichiers
	publicID := fmt.Sprintf("onlyflick/content/%s/%s", contentID, uuid)
	fmt.Printf("✅ [Cloudinary] Public ID créé: %s\n", publicID)

	// Préparer les options d'upload
	uploadParams := uploader.UploadParams{
		PublicID:     publicID,
		Folder:       "onlyflick/content",
		ResourceType: "image",
		// Transformation automatique pour optimiser la taille
		Transformation: "q_auto,f_auto",
		// Tags pour organiser les fichiers
		Tags: []string{"onlyflick", "content", "image"},
		// Context pour ajouter des métadonnées
		Context: map[string]string{
			"content_id":  contentID,
			"uploaded_at": time.Now().Format(time.RFC3339),
		},
	}
	fmt.Printf("✅ [Cloudinary] Paramètres d'upload préparés\n")

	// Uploader le fichier
	fmt.Printf("🚀 [Cloudinary] Début de l'upload vers Cloudinary...\n")
	result, err := cs.client.Upload.Upload(context.Background(), file, uploadParams)
	if err != nil {
		fmt.Printf("❌ [Cloudinary] Erreur lors de l'upload sur Cloudinary: %v\n", err)
		return nil, fmt.Errorf("erreur lors de l'upload sur Cloudinary: %v", err)
	}
	fmt.Printf("✅ [Cloudinary] Upload réussi! URL: %s\n", result.SecureURL)

	return &UploadResult{
		PublicID:     result.PublicID,
		SecureURL:    result.SecureURL,
		URL:          result.URL,
		Format:       result.Format,
		ResourceType: result.ResourceType,
		Width:        result.Width,
		Height:       result.Height,
		Bytes:        int64(result.Bytes),
	}, nil
}

// UploadVideo uploade une vidéo sur Cloudinary
func (cs *CloudinaryService) UploadVideo(file multipart.File, filename string, contentID string) (*UploadResult, error) {
	fmt.Printf("📤 [Cloudinary] Début de l'upload de vidéo - Filename: %s, ContentID: %s\n", filename, contentID)

	// Générer un nom de fichier unique
	uuid, err := utils.GenerateUUID()
	if err != nil {
		fmt.Printf("❌ [Cloudinary] Erreur lors de la génération de l'UUID: %v\n", err)
		return nil, fmt.Errorf("erreur lors de la génération de l'UUID: %v", err)
	}
	fmt.Printf("✅ [Cloudinary] UUID généré: %s\n", uuid)

	// Créer le public ID avec un préfixe pour organiser les fichiers
	publicID := fmt.Sprintf("onlyflick/content/%s/%s", contentID, uuid)
	fmt.Printf("✅ [Cloudinary] Public ID créé: %s\n", publicID)

	// Préparer les options d'upload pour vidéo
	uploadParams := uploader.UploadParams{
		PublicID:     publicID,
		Folder:       "onlyflick/content",
		ResourceType: "video",
		// Transformation automatique pour optimiser la qualité
		Transformation: "q_auto",
		// Tags pour organiser les fichiers
		Tags: []string{"onlyflick", "content", "video"},
		// Context pour ajouter des métadonnées
		Context: map[string]string{
			"content_id":  contentID,
			"uploaded_at": time.Now().Format(time.RFC3339),
		},
	}
	fmt.Printf("✅ [Cloudinary] Paramètres d'upload vidéo préparés\n")

	// Uploader le fichier
	fmt.Printf("🚀 [Cloudinary] Début de l'upload vidéo vers Cloudinary...\n")
	result, err := cs.client.Upload.Upload(context.Background(), file, uploadParams)
	if err != nil {
		fmt.Printf("❌ [Cloudinary] Erreur lors de l'upload vidéo sur Cloudinary: %v\n", err)
		return nil, fmt.Errorf("erreur lors de l'upload vidéo sur Cloudinary: %v", err)
	}
	fmt.Printf("✅ [Cloudinary] Upload vidéo réussi! URL: %s\n", result.SecureURL)

	return &UploadResult{
		PublicID:     result.PublicID,
		SecureURL:    result.SecureURL,
		URL:          result.URL,
		Format:       result.Format,
		ResourceType: result.ResourceType,
		Width:        result.Width,
		Height:       result.Height,
		Bytes:        int64(result.Bytes),
	}, nil
}

// UploadProfilePicture uploade une photo de profil avec des transformations spécifiques
func (cs *CloudinaryService) UploadProfilePicture(file multipart.File, filename string, userID string) (*UploadResult, error) {
	fmt.Printf("📤 [Cloudinary] Début de l'upload de photo de profil - Filename: %s, UserID: %s\n", filename, userID)

	// Générer un nom de fichier unique
	uuid, err := utils.GenerateUUID()
	if err != nil {
		fmt.Printf("❌ [Cloudinary] Erreur lors de la génération de l'UUID: %v\n", err)
		return nil, fmt.Errorf("erreur lors de la génération de l'UUID: %v", err)
	}
	fmt.Printf("✅ [Cloudinary] UUID généré: %s\n", uuid)

	// Créer le public ID avec un préfixe pour organiser les fichiers
	publicID := fmt.Sprintf("onlyflick/profiles/%s/%s", userID, uuid)
	fmt.Printf("✅ [Cloudinary] Public ID créé: %s\n", publicID)

	// Préparer les options d'upload pour photo de profil
	uploadParams := uploader.UploadParams{
		PublicID:     publicID,
		Folder:       "onlyflick/profiles",
		ResourceType: "image",
		// Transformation pour optimiser et redimensionner
		Transformation: "w_400,h_400,c_fill,g_face,q_auto,f_auto",
		// Tags pour organiser les fichiers
		Tags: []string{"onlyflick", "profile", "avatar"},
		// Context pour ajouter des métadonnées
		Context: map[string]string{
			"user_id":     userID,
			"uploaded_at": time.Now().Format(time.RFC3339),
		},
	}

	// Uploader le fichier
	result, err := cs.client.Upload.Upload(context.Background(), file, uploadParams)
	if err != nil {
		fmt.Printf("❌ [Cloudinary] Erreur lors de l'upload de la photo de profil: %v\n", err)
		return nil, fmt.Errorf("erreur lors de l'upload de la photo de profil: %v", err)
	}

	fmt.Printf("✅ [Cloudinary] Upload photo de profil réussi! URL: %s\n", result.SecureURL)

	return &UploadResult{
		PublicID:     result.PublicID,
		SecureURL:    result.SecureURL,
		URL:          result.URL,
		Format:       result.Format,
		ResourceType: result.ResourceType,
		Width:        result.Width,
		Height:       result.Height,
		Bytes:        int64(result.Bytes),
	}, nil
}

// UploadBannerImage uploade une bannière pour un créateur
func (cs *CloudinaryService) UploadBannerImage(file multipart.File, filename string, userID string) (*UploadResult, error) {
	fmt.Printf("📤 [Cloudinary] Début de l'upload de bannière - Filename: %s, UserID: %s\n", filename, userID)

	// Générer un nom de fichier unique
	uuid, err := utils.GenerateUUID()
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la génération de l'UUID: %v", err)
	}

	// Créer le public ID avec un préfixe pour organiser les fichiers
	publicID := fmt.Sprintf("onlyflick/banners/%s/%s", userID, uuid)

	// Préparer les options d'upload pour bannière
	uploadParams := uploader.UploadParams{
		PublicID:     publicID,
		Folder:       "onlyflick/banners",
		ResourceType: "image",
		// Transformation pour optimiser et redimensionner pour bannière
		Transformation: "w_1200,h_400,c_fill,q_auto,f_auto",
		// Tags pour organiser les fichiers
		Tags: []string{"onlyflick", "banner", "creator"},
		// Context pour ajouter des métadonnées
		Context: map[string]string{
			"user_id":     userID,
			"uploaded_at": time.Now().Format(time.RFC3339),
		},
	}

	// Uploader le fichier
	result, err := cs.client.Upload.Upload(context.Background(), file, uploadParams)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'upload de la bannière: %v", err)
	}

	return &UploadResult{
		PublicID:     result.PublicID,
		SecureURL:    result.SecureURL,
		URL:          result.URL,
		Format:       result.Format,
		ResourceType: result.ResourceType,
		Width:        result.Width,
		Height:       result.Height,
		Bytes:        int64(result.Bytes),
	}, nil
}

// GenerateThumbnail génère une miniature pour une image ou vidéo
func (cs *CloudinaryService) GenerateThumbnail(publicID string, resourceType string) (string, error) {
	var transformation string

	switch resourceType {
	case "image":
		// Transformation pour créer une miniature d'image
		transformation = "w_300,h_300,c_fill,q_auto,f_auto"
	case "video":
		// Transformation pour créer une miniature vidéo (première frame)
		transformation = "w_300,h_300,c_fill,q_auto,f_auto,so_0"
	default:
		return "", fmt.Errorf("type de ressource non supporté: %s", resourceType)
	}

	// Générer l'URL de la miniature
	asset, err := cs.client.Image(publicID)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la génération de l'asset: %v", err)
	}

	thumbnailURL, err := asset.String()
	if err != nil {
		return "", fmt.Errorf("erreur lors de la génération de l'URL de miniature: %v", err)
	}

	// Ajouter la transformation à l'URL
	thumbnailURL = strings.Replace(thumbnailURL, "/image/upload/", "/image/upload/"+transformation+"/", 1)

	return thumbnailURL, nil
}

// DeleteFile supprime un fichier de Cloudinary
func (cs *CloudinaryService) DeleteFile(publicID string, resourceType string) error {
	deleteParams := uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: resourceType,
	}

	_, err := cs.client.Upload.Destroy(context.Background(), deleteParams)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression du fichier Cloudinary: %v", err)
	}

	return nil
}

// ValidateFileType valide le type de fichier selon le type de contenu
func (cs *CloudinaryService) ValidateFileType(contentType string, mediaType string, filename string) bool {
	// Convertir en minuscules pour la comparaison
	contentType = strings.ToLower(contentType)
	filename = strings.ToLower(filename)

	switch mediaType {
	case "image":
		// Accepter tous les types d'images
		if strings.HasPrefix(contentType, "image/") {
			return true
		}
		// Si le Content-Type n'est pas détecté correctement (application/octet-stream),
		// vérifier l'extension du fichier
		if contentType == "application/octet-stream" || contentType == "" {
			imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".ico", ".tiff", ".tif"}
			for _, ext := range imageExtensions {
				if strings.HasSuffix(filename, ext) {
					return true
				}
			}
		}
		return false
	case "video":
		// Accepter tous les types de vidéos
		if strings.HasPrefix(contentType, "video/") {
			return true
		}
		// Si le Content-Type n'est pas détecté correctement (application/octet-stream),
		// vérifier l'extension du fichier
		if contentType == "application/octet-stream" || contentType == "" {
			videoExtensions := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".mkv", ".m4v", ".3gp"}
			for _, ext := range videoExtensions {
				if strings.HasSuffix(filename, ext) {
					return true
				}
			}
		}
		return false
	default:
		return false
	}
}

// GetOptimizedURL génère une URL optimisée pour l'affichage
func (cs *CloudinaryService) GetOptimizedURL(publicID string, width, height int) (string, error) {
	transformation := fmt.Sprintf("w_%d,h_%d,c_fill,q_auto,f_auto", width, height)

	asset, err := cs.client.Image(publicID)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la génération de l'asset: %v", err)
	}

	url, err := asset.String()
	if err != nil {
		return "", fmt.Errorf("erreur lors de la génération de l'URL optimisée: %v", err)
	}

	// Ajouter la transformation à l'URL
	url = strings.Replace(url, "/image/upload/", "/image/upload/"+transformation+"/", 1)

	return url, nil
}

// GetImageURL génère une URL d'image avec des transformations personnalisées
func (cs *CloudinaryService) GetImageURL(publicID string, transformations ...string) (string, error) {
	// Créer l'asset image
	asset, err := cs.client.Image(publicID)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la génération de l'asset image: %v", err)
	}

	// Générer l'URL de base
	url, err := asset.String()
	if err != nil {
		return "", fmt.Errorf("erreur lors de la génération de l'URL image: %v", err)
	}

	// Ajouter les transformations à l'URL
	if len(transformations) > 0 {
		transformation := strings.Join(transformations, ",")
		url = strings.Replace(url, "/image/upload/", "/image/upload/"+transformation+"/", 1)
	}

	return url, nil
}

// GetVideoURL génère une URL de vidéo avec des transformations personnalisées
func (cs *CloudinaryService) GetVideoURL(publicID string, transformations ...string) (string, error) {
	// Créer l'asset vidéo
	asset, err := cs.client.Video(publicID)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la génération de l'asset vidéo: %v", err)
	}

	// Générer l'URL de base
	url, err := asset.String()
	if err != nil {
		return "", fmt.Errorf("erreur lors de la génération de l'URL vidéo: %v", err)
	}

	// Ajouter les transformations à l'URL
	if len(transformations) > 0 {
		transformation := strings.Join(transformations, ",")
		url = strings.Replace(url, "/video/upload/", "/video/upload/"+transformation+"/", 1)
	}

	return url, nil
}

// GetThumbnailURL génère une URL de miniature pour n'importe quel type de média
func (cs *CloudinaryService) GetThumbnailURL(publicID string, resourceType string, width, height int) (string, error) {
	var url string
	var err error

	switch resourceType {
	case "image":
		transformation := fmt.Sprintf("w_%d,h_%d,c_fill,q_auto,f_auto", width, height)
		url, err = cs.GetImageURL(publicID, transformation)
	case "video":
		transformation := fmt.Sprintf("w_%d,h_%d,c_fill,q_auto,f_auto,so_0", width, height)
		url, err = cs.GetVideoURL(publicID, transformation)
	default:
		return "", fmt.Errorf("type de ressource non supporté: %s", resourceType)
	}

	if err != nil {
		return "", fmt.Errorf("erreur lors de la génération de l'URL de miniature: %v", err)
	}

	return url, nil
}
