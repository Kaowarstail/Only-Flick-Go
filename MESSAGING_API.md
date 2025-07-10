# API Documentation - OnlyFlick Messaging System

## 🚀 Overview

Le système de messagerie d'OnlyFlick permet aux utilisateurs d'échanger des messages privés, avec support des **messages payants**. Les créateurs peuvent envoyer du contenu premium verrouillé que les utilisateurs peuvent débloquer moyennant paiement.

## 🔐 Authentication

Toutes les routes nécessitent un token JWT valide dans l'en-tête Authorization :
```
Authorization: Bearer <your-jwt-token>
```

## 📝 Endpoints

### Conversations

#### GET /api/v1/conversations
Récupère les conversations de l'utilisateur connecté

**Query Parameters:**
- `page` (int, optional): Numéro de page (défaut: 1)
- `limit` (int, optional): Nombre d'éléments par page (défaut: 20, max: 50)

**Response:**
```json
{
  "success": true,
  "data": {
    "conversations": [
      {
        "id": 1,
        "participants": [
          {
            "id": "user-1",
            "username": "alice_photographer",
            "profile_picture": "https://example.com/avatar.jpg"
          },
          {
            "id": "user-2", 
            "username": "bob_fitness"
          }
        ],
        "last_message": {
          "id": 15,
          "content": "Hello there!",
          "sender": {
            "id": "user-1",
            "username": "alice_photographer"
          },
          "is_paid": false,
          "created_at": "2025-07-10T14:30:00Z"
        },
        "unread_count": 3,
        "updated_at": "2025-07-10T14:30:00Z",
        "is_active": true
      }
    ],
    "total": 25,
    "unread_total": 8
  },
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 25,
    "total_pages": 2
  }
}
```

#### POST /api/v1/conversations
Créer ou récupérer une conversation avec un autre utilisateur

**Body:**
```json
{
  "other_user_id": "user-123"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "participant_1_id": "user-1",
    "participant_2_id": "user-123",
    "participants": [...],
    "created_at": "2025-07-10T14:30:00Z"
  },
  "message": "Conversation created successfully"
}
```

#### GET /api/v1/conversations/{id}/messages
Récupère les messages d'une conversation

**Query Parameters:**
- `page` (int, optional): Numéro de page (défaut: 1)
- `limit` (int, optional): Nombre d'éléments par page (défaut: 50, max: 100)

**Response:**
```json
{
  "success": true,
  "data": {
    "messages": [
      {
        "id": 42,
        "conversation_id": 1,
        "sender": {
          "id": "user-1",
          "username": "alice_photographer"
        },
        "content": "Check out my latest photos!",
        "media_url": null,
        "is_paid": false,
        "message_type": "text",
        "status": "read",
        "created_at": "2025-07-10T14:30:00Z",
        "read_at": "2025-07-10T14:32:00Z"
      },
      {
        "id": 43,
        "conversation_id": 1,
        "sender": {
          "id": "user-1", 
          "username": "alice_photographer"
        },
        "content": null,
        "media_url": null,
        "is_paid": true,
        "price": 9.99,
        "is_unlocked": false,
        "can_unlock": true,
        "message_type": "paid_text",
        "status": "sent",
        "created_at": "2025-07-10T14:35:00Z"
      }
    ],
    "total": 15,
    "page": 1,
    "limit": 50,
    "has_more": false
  }
}
```

#### PUT /api/v1/conversations/{id}/read
Marquer tous les messages d'une conversation comme lus

**Response:**
```json
{
  "success": true,
  "message": "Messages marked as read"
}
```

### Messages

#### POST /api/v1/messages
Envoyer un message gratuit

**Rate Limit:** 10 messages/minute

**Body:**
```json
{
  "conversation_id": 1,
  "content": "Hello! How are you?",
  "message_type": "text"
}
```

**Body (avec media):**
```json
{
  "conversation_id": 1,
  "media_url": "https://example.com/image.jpg",
  "media_type": "image",
  "message_type": "image"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 44,
    "conversation_id": 1,
    "sender_id": "user-2",
    "content": "Hello! How are you?",
    "is_paid": false,
    "message_type": "text",
    "status": "sent",
    "created_at": "2025-07-10T14:40:00Z"
  },
  "message": "Message sent successfully"
}
```

#### POST /api/v1/messages/paid
Envoyer un message payant

**Rate Limit:** 5 messages/minute

**Body:**
```json
{
  "conversation_id": 1,
  "content": "Exclusive behind-the-scenes content!",
  "price": 15.99,
  "message_type": "paid_text"
}
```

**Body (avec media payant):**
```json
{
  "conversation_id": 1,
  "media_url": "https://example.com/premium-video.mp4",
  "media_type": "video",
  "thumbnail_url": "https://example.com/thumbnail.jpg",
  "price": 25.99,
  "message_type": "paid_media"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 45,
    "conversation_id": 1,
    "sender_id": "user-1",
    "content": "Exclusive behind-the-scenes content!",
    "is_paid": true,
    "price": 15.99,
    "is_unlocked": false,
    "message_type": "paid_text",
    "status": "sent",
    "created_at": "2025-07-10T14:45:00Z"
  },
  "message": "Paid message sent successfully"
}
```

#### POST /api/v1/messages/{id}/unlock
Débloquer un message payant

**Response:**
```json
{
  "success": true,
  "message": "Message unlocked successfully"
}
```

#### GET /api/v1/messages/{id}
Récupérer un message spécifique avec contrôle d'accès

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 45,
    "conversation_id": 1,
    "sender": {
      "id": "user-1",
      "username": "alice_photographer"
    },
    "content": "Exclusive content here!",
    "is_paid": true,
    "price": 15.99,
    "is_unlocked": true,
    "can_unlock": false,
    "message_type": "paid_text",
    "created_at": "2025-07-10T14:45:00Z"
  }
}
```

## 🚫 Error Responses

### 400 Bad Request
```json
{
  "success": false,
  "error": {
    "message": "Invalid request data"
  }
}
```

### 401 Unauthorized
```json
{
  "success": false,
  "error": {
    "message": "User ID not found in context"
  }
}
```

### 403 Forbidden
```json
{
  "success": false,
  "error": {
    "message": "Access denied to this conversation"
  }
}
```

### 429 Too Many Requests
```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many messages. Please slow down."
  }
}
```

## 💰 Tarification des Messages Payants

- **Prix minimum:** 0.99€
- **Prix maximum:** 500.00€
- **Commission OnlyFlick:** 20% sur chaque transaction
- **Revenus créateur:** 80% du prix du message

## 🔄 Logique d'Affichage des Messages Payants

1. **Expéditeur:** Voit toujours le contenu complet
2. **Destinataire non-payeur:** Voit les métadonnées (prix, type) mais pas le contenu
3. **Destinataire payeur:** Voit le contenu complet après déverrouillage
4. **Thumbnail:** Peut être visible même pour les messages verrouillés

## 🚀 Exemples d'Utilisation

### Créer une conversation et envoyer un message
```bash
# 1. Créer conversation
curl -X POST http://localhost:8080/api/v1/conversations \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"other_user_id": "user-123"}'

# 2. Envoyer un message
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "conversation_id": 1,
    "content": "Hello there!",
    "message_type": "text"
  }'
```

### Envoyer un message payant
```bash
curl -X POST http://localhost:8080/api/v1/messages/paid \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "conversation_id": 1,
    "content": "Exclusive content!",
    "price": 9.99,
    "message_type": "paid_text"
  }'
```

### Débloquer un message payant
```bash
curl -X POST http://localhost:8080/api/v1/messages/45/unlock \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 🔧 Rate Limiting

- **Messages normaux:** 10 par minute
- **Messages payants:** 5 par minute
- **Récupération:** Pas de limite
- **Lecture:** Pas de limite

Les limites sont appliquées par utilisateur et se réinitialisent automatiquement.
