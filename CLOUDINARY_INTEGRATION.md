# Configuration Cloudinary - OnlyFlick

## Vue d'ensemble

OnlyFlick utilise Cloudinary pour gérer le stockage et l'optimisation des médias (images, vidéos, photos de profil, bannières). Cette intégration permet :

- **Stockage sécurisé** : Tous les médias sont stockés sur Cloudinary avec des URLs sécurisées
- **Optimisation automatique** : Les images et vidéos sont automatiquement optimisées (qualité, format, taille)
- **Transformations en temps réel** : Génération d'URLs optimisées selon les besoins
- **Gestion des miniatures** : Génération automatique de miniatures pour les vidéos
- **Organisation** : Les fichiers sont organisés par dossiers (content, profiles, banners)

## Configuration

### Variables d'environnement requises

```bash
# Configuration Cloudinary
export CLOUDINARY_CLOUD_NAME="your-cloud-name"
export CLOUDINARY_API_KEY="your-api-key"
export CLOUDINARY_API_SECRET="your-api-secret"
```

### Structure des dossiers Cloudinary

```
onlyflick/
├── content/
│   ├── {content_id}/
│   │   ├── {uuid} (fichier média)
│   │   └── {uuid} (miniature)
├── profiles/
│   ├── {user_id}/
│   │   └── {uuid} (photo de profil)
└── banners/
    ├── {user_id}/
        └── {uuid} (bannière)
```

## Endpoints API

### Upload de média pour un contenu

```http
POST /api/contents/{id}/media
Content-Type: multipart/form-data
Authorization: Bearer {jwt_token}

{
  "media": [fichier]
}
```

**Réponse :**
```json
{
  "message": "Média uploadé avec succès",
  "media_url": "https://res.cloudinary.com/...",
  "thumbnail_url": "https://res.cloudinary.com/...",
  "public_id": "onlyflick/content/123/uuid",
  "content": { ... },
  "file_info": {
    "format": "jpg",
    "resource_type": "image",
    "width": 1920,
    "height": 1080,
    "bytes": 245760
  }
}
```

### Upload de miniature pour un contenu

```http
POST /api/contents/{id}/thumbnail
Content-Type: multipart/form-data
Authorization: Bearer {jwt_token}

{
  "thumbnail": [fichier]
}
```

### Upload de photo de profil

```http
POST /api/users/{id}/profile-picture
Content-Type: multipart/form-data
Authorization: Bearer {jwt_token}

{
  "profile_picture": [fichier]
}
```

### Upload de bannière de créateur

```http
POST /api/creators/{id}/banner
Content-Type: multipart/form-data
Authorization: Bearer {jwt_token}

{
  "banner_image": [fichier]
}
```

### Génération d'URL optimisée

```http
GET /api/contents/{id}/optimized-url?w=800&h=600&q=80&f=webp
Authorization: Bearer {jwt_token}
```

**Paramètres :**
- `w` : largeur
- `h` : hauteur
- `q` : qualité (auto, 1-100)
- `f` : format (auto, webp, jpg, png)

## Modèles de données

### Content (mise à jour)

```go
type Content struct {
    ID           uint           `json:"id"`
    CreatorID    string         `json:"creator_id"`
    Title        string         `json:"title"`
    Description  string         `json:"description"`
    Type         string         `json:"type"` // image, video, text
    MediaURL     string         `json:"media_url"`      // URL Cloudinary
    ThumbnailURL string         `json:"thumbnail_url"`  // URL miniature
    CoverURL     string         `json:"cover_url"`
    PublicID     string         `json:"public_id"`      // ID Cloudinary pour suppression
    IsPremium    bool           `json:"is_premium"`
    IsPublished  bool           `json:"is_published"`
    ViewCount    int            `json:"view_count"`
    IsFlagged    bool           `json:"is_flagged"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
    DeletedAt    gorm.DeletedAt `json:"-"`
}
```

### User (mise à jour)

```go
type User struct {
    ID             string    `json:"id"`
    Username       string    `json:"username"`
    Email          string    `json:"email"`
    Password       string    `json:"-"`
    FirstName      string    `json:"first_name"`
    LastName       string    `json:"last_name"`
    Role           UserRole  `json:"role"`
    Biography      string    `json:"biography"`
    ProfilePicture string    `json:"profile_picture"` // URL Cloudinary
    // ... autres champs
}
```

### CreatorProfile (mise à jour)

```go
type CreatorProfile struct {
    ID          uint      `json:"id"`
    UserID      string    `json:"user_id"`
    BannerImage string    `json:"banner_image"` // URL Cloudinary
    WebsiteURL  string    `json:"website_url"`
    SocialLinks string    `json:"social_links"`
    // ... autres champs
}
```

## Transformations automatiques

### Photos de profil
- Redimensionnement : 300x300 pixels
- Recadrage : carré centré sur le visage
- Optimisation : qualité et format automatiques

### Bannières
- Redimensionnement : 1200x400 pixels
- Recadrage : remplissage intelligent
- Optimisation : qualité et format automatiques

### Contenus média
- Images : optimisation automatique (qualité, format)
- Vidéos : optimisation automatique avec génération de miniatures
- Miniatures : 300x300 pixels pour les grilles

## Sécurité

### Validation des fichiers
- Types MIME autorisés : image/*, video/*
- Extensions autorisées : .jpg, .jpeg, .png, .gif, .webp, .mp4, .avi, .mov, .wmv, .flv, .webm, .mkv
- Taille maximale : 5MB pour les images, 10MB pour les bannières, 100MB pour les vidéos

### Permissions
- Seul le propriétaire peut uploader des médias pour son contenu
- Seul l'utilisateur peut modifier sa photo de profil
- Seul le créateur peut modifier sa bannière

## Tests

### Test de configuration

```bash
# Tester la configuration Cloudinary
./scripts/test_cloudinary.sh
```

### Test d'upload avec curl

```bash
# Upload d'une image
curl -X POST http://localhost:8080/api/contents/1/media \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "media=@/path/to/image.jpg"

# Upload d'une photo de profil
curl -X POST http://localhost:8080/api/users/USER_ID/profile-picture \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "profile_picture=@/path/to/profile.jpg"
```

## Gestion des erreurs

### Erreurs communes
- `CloudName manquant` : Vérifiez la variable CLOUDINARY_CLOUD_NAME
- `APIKey manquant` : Vérifiez la variable CLOUDINARY_API_KEY
- `APISecret manquant` : Vérifiez la variable CLOUDINARY_API_SECRET
- `Erreur d'upload` : Vérifiez la connectivité Internet et les permissions Cloudinary

### Logs
Les logs détaillés sont disponibles dans la console avec le préfixe `[Cloudinary]`.

## Maintenance

### Nettoyage des fichiers
Lors de la suppression d'un contenu, les fichiers Cloudinary sont automatiquement supprimés via le `PublicID`.

### Monitoring
Surveillez l'utilisation de Cloudinary via le dashboard Cloudinary pour :
- Bande passante utilisée
- Transformations effectuées
- Stockage utilisé
- Quotas et limites

## Migration

Pour migrer les contenus existants vers Cloudinary :

1. Utilisez l'endpoint `/api/contents/{id}/migrate` (non implémenté)
2. Ou uploadez manuellement les nouveaux médias via les endpoints d'upload
3. Les anciens contenus continuent de fonctionner avec leurs URLs existantes
