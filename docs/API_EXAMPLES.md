# OnlyFlick API - Exemples d'appels pour la messagerie et l'édition de profil

Ce document contient des exemples d'appels API pour les nouvelles fonctionnalités de messagerie et d'édition de profil de OnlyFlick.

## Base URL
```
http://localhost:8080/api/v1
```

## Authentification

Tous les appels API nécessitent un token JWT dans l'en-tête Authorization :
```
Authorization: Bearer <your-jwt-token>
```

## 1. Messagerie (Conversations)

### Récupérer toutes les conversations de l'utilisateur
```bash
GET /api/v1/messaging/conversations
```

**Réponse :**
```json
{
  "success": true,
  "data": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "user1_id": "user1-id",
      "user2_id": "user2-id",
      "participants": [
        {
          "id": "user2-id",
          "username": "alice",
          "profile_picture": "/uploads/avatars/alice.jpg"
        }
      ],
      "last_message": {
        "id": "msg-id",
        "content": "Hello!",
        "created_at": "2025-01-08T10:30:00Z",
        "sender_id": "user2-id"
      },
      "unread_count": 2,
      "created_at": "2025-01-08T10:00:00Z",
      "updated_at": "2025-01-08T10:30:00Z"
    }
  ],
  "message": "Conversations récupérées avec succès"
}
```

### Créer une nouvelle conversation
```bash
POST /api/v1/messaging/conversations
Content-Type: application/json

{
  "participant_id": "user-id-to-chat-with"
}
```

**Réponse :**
```json
{
  "success": true,
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "user1_id": "current-user-id",
    "user2_id": "user-id-to-chat-with",
    "created_at": "2025-01-08T10:30:00Z",
    "updated_at": "2025-01-08T10:30:00Z",
    "is_active": true
  },
  "message": "Conversation créée avec succès"
}
```

### Récupérer les messages d'une conversation
```bash
GET /api/v1/messaging/conversations/{conversation_id}/messages?page=1&limit=20
```

**Réponse :**
```json
{
  "success": true,
  "data": [
    {
      "id": "msg-id",
      "conversation_id": "conv-id",
      "sender_id": "user-id",
      "sender": {
        "id": "user-id",
        "username": "alice",
        "profile_picture": "/uploads/avatars/alice.jpg"
      },
      "content": "Hello! How are you?",
      "message_type": "text",
      "is_paid": false,
      "status": "read",
      "created_at": "2025-01-08T10:30:00Z",
      "read_at": "2025-01-08T10:35:00Z"
    }
  ],
  "message": "Messages récupérés avec succès"
}
```

### Marquer une conversation comme lue
```bash
POST /api/v1/messaging/conversations/{conversation_id}/read
```

**Réponse :**
```json
{
  "success": true,
  "message": "Conversation marquée comme lue"
}
```

### Récupérer le nombre de messages non lus
```bash
GET /api/v1/messaging/unread-count
```

**Réponse :**
```json
{
  "success": true,
  "data": {
    "unread_count": 5
  },
  "message": "Nombre de messages non lus récupéré avec succès"
}
```

## 2. Messages

### Envoyer un message texte
```bash
POST /api/v1/messaging/messages
Content-Type: application/json

{
  "conversation_id": "conv-id",
  "content": "Hello! How are you doing?",
  "message_type": "text"
}
```

### Envoyer un message avec média
```bash
POST /api/v1/messaging/messages
Content-Type: application/json

{
  "conversation_id": "conv-id",
  "content": "Check out this photo!",
  "message_type": "image",
  "media_url": "/uploads/messages/image-id.jpg",
  "media_type": "image"
}
```

### Envoyer un message payant
```bash
POST /api/v1/messaging/messages
Content-Type: application/json

{
  "conversation_id": "conv-id",
  "content": "Exclusive content for you!",
  "message_type": "paid_text",
  "is_paid": true,
  "price": 5.99,
  "preview_text": "This is a preview of the exclusive content..."
}
```

**Réponse :**
```json
{
  "success": true,
  "data": {
    "id": "msg-id",
    "conversation_id": "conv-id",
    "sender_id": "current-user-id",
    "content": "Exclusive content for you!",
    "message_type": "paid_text",
    "is_paid": true,
    "price": 5.99,
    "preview_text": "This is a preview of the exclusive content...",
    "is_unlocked": false,
    "status": "sent",
    "created_at": "2025-01-08T10:30:00Z"
  },
  "message": "Message envoyé avec succès"
}
```

### Déverrouiller un message payant
```bash
POST /api/v1/messaging/messages/{message_id}/unlock
```

**Réponse :**
```json
{
  "success": true,
  "data": {
    "id": "msg-id",
    "content": "Exclusive content for you!",
    "is_unlocked": true,
    "unlocked_at": "2025-01-08T10:35:00Z",
    "unlocked_by": "current-user-id"
  },
  "message": "Message débloqué avec succès"
}
```

### Récupérer l'aperçu d'un message payant
```bash
GET /api/v1/messaging/messages/{message_id}/preview
```

**Réponse :**
```json
{
  "success": true,
  "data": {
    "id": "msg-id",
    "preview_text": "This is a preview of the exclusive content...",
    "price": 5.99,
    "is_unlocked": false,
    "sender": {
      "id": "sender-id",
      "username": "creator",
      "profile_picture": "/uploads/avatars/creator.jpg"
    }
  },
  "message": "Aperçu récupéré avec succès"
}
```

### Marquer un message comme lu
```bash
POST /api/v1/messaging/messages/{message_id}/read
```

### Supprimer un message
```bash
DELETE /api/v1/messaging/messages/{message_id}
```

### Récupérer les statistiques de messages
```bash
GET /api/v1/messaging/messages/stats
```

**Réponse :**
```json
{
  "success": true,
  "data": {
    "total_messages": 150,
    "paid_messages": 25,
    "unlocked_messages": 10,
    "total_earnings": 149.75,
    "total_spent": 59.90,
    "conversation_count": 15
  },
  "message": "Statistiques récupérées avec succès"
}
```

## 3. Upload de médias

### Uploader un fichier
```bash
POST /api/v1/messaging/media/upload
Content-Type: multipart/form-data

file: [binary data]
```

**Réponse :**
```json
{
  "success": true,
  "data": {
    "file_id": "file-id",
    "file_url": "/uploads/messages/file-id.jpg",
    "file_name": "photo.jpg",
    "file_size": 1024000,
    "media_type": "image",
    "mime_type": "image/jpeg"
  },
  "message": "Fichier uploadé avec succès"
}
```

### Récupérer les métadonnées d'un fichier
```bash
GET /api/v1/messaging/media/{file_id}
```

### Récupérer tous les fichiers de l'utilisateur
```bash
GET /api/v1/messaging/media
```

### Servir un fichier
```bash
GET /uploads/messages/{filename}
```

### Supprimer un fichier
```bash
DELETE /api/v1/messaging/media/{file_id}
```

## 4. Édition de profil utilisateur

### Mettre à jour le profil utilisateur
```bash
PUT /api/v1/profiles/users/me
Content-Type: application/json

{
  "first_name": "John",
  "last_name": "Doe",
  "biography": "Content creator and photographer",
  "location": "Paris, France",
  "website": "https://johndoe.com",
  "birthday": "1990-01-15",
  "gender": "male",
  "social_links": {
    "instagram": "johndoe",
    "twitter": "johndoe",
    "youtube": "johndoe",
    "website": "https://johndoe.com"
  }
}
```

**Réponse :**
```json
{
  "success": true,
  "data": {
    "id": "user-id",
    "username": "johndoe",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "biography": "Content creator and photographer",
    "location": "Paris, France",
    "website": "https://johndoe.com",
    "profile_picture": "/uploads/avatars/johndoe.jpg",
    "updated_at": "2025-01-08T10:30:00Z"
  },
  "message": "Profil mis à jour avec succès"
}
```

### Récupérer un profil utilisateur
```bash
GET /api/v1/profiles/users/{user_id}
```

**Réponse :**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "user-id",
      "username": "johndoe",
      "first_name": "John",
      "last_name": "Doe",
      "biography": "Content creator and photographer",
      "profile_picture": "/uploads/avatars/johndoe.jpg",
      "location": "Paris, France",
      "website": "https://johndoe.com"
    },
    "social_links": {
      "instagram": "johndoe",
      "twitter": "johndoe",
      "youtube": "johndoe",
      "website": "https://johndoe.com"
    },
    "stats": {
      "user_id": "user-id",
      "followers_count": 150,
      "following_count": 75,
      "posts_count": 25,
      "total_messages_sent": 500,
      "last_active_at": "2025-01-08T10:30:00Z"
    }
  },
  "message": "Profil récupéré avec succès"
}
```

### Mettre à jour les liens sociaux
```bash
PUT /api/v1/profiles/users/me/social-links
Content-Type: application/json

{
  "links": {
    "instagram": "johndoe",
    "twitter": "johndoe",
    "tiktok": "johndoe",
    "youtube": "johndoe",
    "website": "https://johndoe.com"
  }
}
```

### Récupérer les statistiques utilisateur
```bash
GET /api/v1/profiles/users/{user_id}/stats
```

## 5. Gestion de profil créateur

### Mettre à jour le profil créateur
```bash
PUT /api/v1/profiles/creators/me
Content-Type: application/json

{
  "biography": "Professional photographer and content creator",
  "subscription_price": 9.99,
  "message_price": 2.99,
  "custom_tip_amounts": [1.99, 4.99, 9.99, 19.99],
  "accept_custom_tips": true,
  "accept_messaging": true,
  "accept_subscriptions": true,
  "social_links": {
    "instagram": "creator",
    "twitter": "creator",
    "onlyfans": "creator"
  }
}
```

### Récupérer les gains du créateur
```bash
GET /api/v1/profiles/creators/me/earnings
```

**Réponse :**
```json
{
  "success": true,
  "data": {
    "creator_id": "creator-id",
    "total_earnings": 2499.75,
    "total_messages_earnings": 599.50,
    "total_subscriptions_earnings": 1800.25,
    "total_tips_earnings": 100.00,
    "current_month_earnings": 350.25,
    "last_month_earnings": 425.50,
    "pending_earnings": 150.75,
    "withdrawn_earnings": 2349.00,
    "updated_at": "2025-01-08T10:30:00Z"
  },
  "message": "Gains récupérés avec succès"
}
```

### Récupérer les gains mensuels
```bash
GET /api/v1/profiles/creators/me/monthly-earnings
```

**Réponse :**
```json
{
  "success": true,
  "data": [
    {
      "id": "earning-id",
      "creator_id": "creator-id",
      "year": 2025,
      "month": 1,
      "total_earnings": 350.25,
      "messages_earnings": 150.50,
      "subscriptions_earnings": 180.75,
      "tips_earnings": 19.00,
      "transactions_count": 45,
      "created_at": "2025-01-01T00:00:00Z"
    }
  ],
  "message": "Gains mensuels récupérés avec succès"
}
```

### Récupérer les statistiques détaillées du créateur
```bash
GET /api/v1/profiles/creators/me/stats
```

**Réponse :**
```json
{
  "success": true,
  "data": {
    "followers_count": 1250,
    "following_count": 150,
    "posts_count": 125,
    "total_messages_sent": 2500,
    "total_earnings": 2499.75,
    "current_month_earnings": 350.25,
    "message_stats": {
      "total_sent": 2500,
      "paid_messages_sent": 150,
      "message_revenue": 599.50
    },
    "conversation_stats": {
      "active_conversations": 25,
      "total_conversations": 150
    },
    "last_active_at": "2025-01-08T10:30:00Z"
  },
  "message": "Statistiques détaillées récupérées avec succès"
}
```

## 6. Gestion des erreurs

### Erreur d'authentification
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Token d'authentification requis"
  }
}
```

### Erreur de validation
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Le prix minimum pour un message payant est de 0.99"
  }
}
```

### Erreur de permissions
```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "Accès non autorisé à cette conversation"
  }
}
```

### Erreur de ressource non trouvée
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Conversation non trouvée"
  }
}
```

## 7. Codes de statut HTTP

- `200 OK` : Requête réussie
- `201 Created` : Ressource créée avec succès
- `400 Bad Request` : Données invalides
- `401 Unauthorized` : Authentification requise
- `403 Forbidden` : Permissions insuffisantes
- `404 Not Found` : Ressource non trouvée
- `413 Payload Too Large` : Fichier trop volumineux
- `429 Too Many Requests` : Limite de taux dépassée
- `500 Internal Server Error` : Erreur serveur

## 8. Limites et contraintes

### Messages
- Longueur maximale du contenu : 5000 caractères
- Prix minimum message payant : 0.99€
- Prix maximum message payant : 500€
- Messages par minute : 30 maximum

### Fichiers
- Taille maximale image : 10MB
- Taille maximale vidéo : 100MB
- Taille maximale audio : 20MB
- Taille maximale document : 10MB
- Types supportés : JPEG, PNG, GIF, WebP, MP4, AVI, MOV, MP3, WAV, PDF

### Profil
- Longueur maximale biographie : 1000 caractères
- Longueur maximale nom d'utilisateur : 30 caractères
- Longueur minimale nom d'utilisateur : 3 caractères

## 9. Webhooks (optionnel)

Pour les intégrations avancées, OnlyFlick peut envoyer des webhooks pour :
- Nouveaux messages reçus
- Messages payants débloqués
- Nouveaux abonnés
- Nouveaux gains

Configuration via l'interface administrateur.
