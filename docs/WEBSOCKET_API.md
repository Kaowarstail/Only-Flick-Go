# OnlyFlick WebSocket API Documentation

## Introduction

OnlyFlick utilise des WebSockets pour la communication temps réel, permettant:
- Messages instantanés
- Indicateurs de frappe (typing indicators)
- Statut de présence (online/offline)
- Notifications en temps réel
- Synchronisation des conversations

## Connexion WebSocket

### Endpoint
```
ws://localhost:8080/api/v1/ws
```

### Authentification
La connexion WebSocket nécessite un token JWT valide dans les headers:
```javascript
const headers = {
    'Authorization': 'Bearer YOUR_JWT_TOKEN'
};
```

### Exemple de connexion JavaScript
```javascript
const token = 'your-jwt-token';
const ws = new WebSocket('ws://localhost:8080/api/v1/ws', [], {
    headers: {
        'Authorization': `Bearer ${token}`
    }
});

ws.onopen = function(event) {
    console.log('WebSocket connected');
};

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    handleWebSocketMessage(message);
};

ws.onclose = function(event) {
    console.log('WebSocket disconnected');
};
```

## Format des Messages

Tous les messages WebSocket suivent ce format JSON:
```json
{
    "type": "event_type",
    "data": {},
    "timestamp": "2024-01-01T00:00:00Z",
    "user_id": "user_id",
    "conversation_id": "conversation_id"
}
```

## Types d'Événements

### Messages

#### `message_sent`
Nouveau message envoyé dans une conversation
```json
{
    "type": "message_sent",
    "data": {
        "message": {
            "id": "msg_123",
            "conversation_id": "conv_456",
            "sender_id": "user_789",
            "content": "Hello!",
            "message_type": "text",
            "is_paid": false,
            "created_at": "2024-01-01T00:00:00Z"
        },
        "conversation": {
            "id": "conv_456",
            "participant_1_id": "user_789",
            "participant_2_id": "user_101"
        },
        "sender": {
            "id": "user_789",
            "username": "john_doe"
        }
    },
    "timestamp": "2024-01-01T00:00:00Z",
    "user_id": "user_789",
    "conversation_id": "conv_456"
}
```

#### `paid_message_sent`
Message payant envoyé
```json
{
    "type": "paid_message_sent",
    "data": {
        "message": {
            "id": "msg_123",
            "conversation_id": "conv_456",
            "sender_id": "user_789",
            "content": "Premium content",
            "is_paid": true,
            "price": 5.99,
            "preview_text": "This is a premium message..."
        }
    }
}
```

#### `paid_message_unlocked`
Message payant débloqué
```json
{
    "type": "paid_message_unlocked",
    "data": {
        "message_id": "msg_123",
        "unlocked_by": "user_101",
        "message": {
            "id": "msg_123",
            "content": "Full premium content revealed"
        },
        "transaction": {
            "id": "tx_456",
            "amount": 5.99,
            "status": "completed"
        }
    }
}
```

### Indicateurs de Frappe

#### `user_typing`
Utilisateur en train d'écrire
```json
{
    "type": "user_typing",
    "data": {
        "user_id": "user_789",
        "username": "john_doe",
        "conversation_id": "conv_456",
        "is_typing": true
    }
}
```

#### `user_stopped_typing`
Utilisateur a arrêté d'écrire
```json
{
    "type": "user_stopped_typing",
    "data": {
        "user_id": "user_789",
        "username": "john_doe",
        "conversation_id": "conv_456",
        "is_typing": false
    }
}
```

### Statut de Présence

#### `user_online`
Utilisateur en ligne
```json
{
    "type": "user_online",
    "data": {
        "user_id": "user_789",
        "username": "john_doe",
        "is_online": true,
        "last_active_at": "2024-01-01T00:00:00Z"
    }
}
```

#### `user_offline`
Utilisateur hors ligne
```json
{
    "type": "user_offline",
    "data": {
        "user_id": "user_789",
        "username": "john_doe",
        "is_online": false,
        "last_active_at": "2024-01-01T00:00:00Z"
    }
}
```

### Conversations

#### `conversation_updated`
Conversation mise à jour
```json
{
    "type": "conversation_updated",
    "data": {
        "conversation": {
            "id": "conv_456",
            "participant_1_id": "user_789",
            "participant_2_id": "user_101",
            "updated_at": "2024-01-01T00:00:00Z"
        },
        "last_message": {
            "id": "msg_123",
            "content": "Latest message",
            "created_at": "2024-01-01T00:00:00Z"
        },
        "unread_count": 3
    }
}
```

### Événements Système

#### `connection_established`
Connexion établie avec succès
```json
{
    "type": "connection_established",
    "data": {
        "user_id": "user_789",
        "server_time": "2024-01-01T00:00:00Z",
        "connection_id": "user_789_20240101000000",
        "capabilities": [
            "messaging",
            "typing",
            "presence",
            "typing_indicators",
            "user_presence"
        ]
    }
}
```

#### `heartbeat`
Battement de cœur pour maintenir la connexion
```json
{
    "type": "heartbeat",
    "data": {
        "server_time": "2024-01-01T00:00:00Z",
        "client_time": "2024-01-01T00:00:00Z"
    }
}
```

#### `error`
Erreur WebSocket
```json
{
    "type": "error",
    "data": {
        "code": "RATE_LIMIT_EXCEEDED",
        "message": "Too many messages, please slow down",
        "details": "Maximum 100 messages per minute"
    }
}
```

## Envoi de Messages

### Indicateur de Frappe
```javascript
ws.send(JSON.stringify({
    type: 'user_typing',
    data: {
        conversation_id: 'conv_456',
        is_typing: true
    }
}));
```

### Arrêt de Frappe
```javascript
ws.send(JSON.stringify({
    type: 'user_stopped_typing',
    data: {
        conversation_id: 'conv_456',
        is_typing: false
    }
}));
```

### Activité dans une Conversation
```javascript
ws.send(JSON.stringify({
    type: 'user_active_in_conversation',
    data: {
        conversation_id: 'conv_456',
        is_active: true
    }
}));
```

### Heartbeat
```javascript
ws.send(JSON.stringify({
    type: 'heartbeat',
    data: {
        client_time: new Date().toISOString()
    }
}));
```

## Gestion des Erreurs

### Codes d'Erreur Courants

- `RATE_LIMIT_EXCEEDED`: Trop de messages envoyés
- `INVALID_MESSAGE_FORMAT`: Format de message invalide
- `UNKNOWN_EVENT_TYPE`: Type d'événement inconnu
- `ACCESS_DENIED`: Accès refusé à la conversation
- `INVALID_TYPING_EVENT`: Données d'événement de frappe invalides
- `CONNECTION_LIMIT_EXCEEDED`: Limite de connexions atteinte

### Gestion des Déconnexions

```javascript
ws.onclose = function(event) {
    console.log('WebSocket disconnected:', event.code, event.reason);
    
    // Reconnexion automatique
    setTimeout(function() {
        connectWebSocket();
    }, 5000);
};
```

## Limites et Contraintes

### Rate Limiting
- **Messages**: 100 messages par minute par utilisateur
- **Connexions**: 1000 connexions simultanées maximum
- **Typing**: Auto-stop après 3 secondes d'inactivité

### Timeouts
- **Heartbeat**: Toutes les 54 secondes
- **Inactivité**: 30 minutes
- **Écriture**: 10 secondes
- **Lecture**: 60 secondes

### Sécurité
- Authentification JWT obligatoire
- Vérification des origines en production
- Validation des permissions pour les conversations
- Protection contre le spam

## API REST Complémentaire

### Informations WebSocket
```
GET /api/v1/ws/info
```

### Utilisateurs En Ligne
```
GET /api/v1/ws/online
```

### Statut Utilisateur
```
GET /api/v1/ws/user/{user_id}/status
```

### Métriques (Admin)
```
GET /api/v1/ws/admin/metrics
```

### Health Check
```
GET /api/v1/ws/health
```

## Exemple d'Implémentation Flutter

```dart
import 'package:web_socket_channel/web_socket_channel.dart';

class WebSocketService {
  WebSocketChannel? _channel;
  
  void connect(String token) {
    final uri = Uri.parse('ws://localhost:8080/api/v1/ws');
    _channel = WebSocketChannel.connect(
      uri,
      headers: {'Authorization': 'Bearer $token'},
    );
    
    _channel!.stream.listen((message) {
      final data = jsonDecode(message);
      _handleMessage(data);
    });
  }
  
  void sendTyping(String conversationId, bool isTyping) {
    _channel!.sink.add(jsonEncode({
      'type': isTyping ? 'user_typing' : 'user_stopped_typing',
      'data': {
        'conversation_id': conversationId,
        'is_typing': isTyping,
      },
    }));
  }
  
  void setActiveConversation(String conversationId) {
    _channel!.sink.add(jsonEncode({
      'type': 'user_active_in_conversation',
      'data': {
        'conversation_id': conversationId,
        'is_active': true,
      },
    }));
  }
  
  void _handleMessage(Map<String, dynamic> data) {
    switch (data['type']) {
      case 'message_sent':
        _handleNewMessage(data['data']);
        break;
      case 'user_typing':
        _handleTyping(data['data']);
        break;
      case 'user_online':
        _handleUserOnline(data['data']);
        break;
      case 'paid_message_unlocked':
        _handlePaidMessageUnlocked(data['data']);
        break;
    }
  }
  
  void dispose() {
    _channel?.sink.close();
  }
}
```

## Test de Connexion

Utiliser l'outil de test inclus:
```bash
cd cmd/websocket-test
go run main.go -token YOUR_JWT_TOKEN
```

## Monitoring et Métriques

Les métriques WebSocket incluent:
- Connexions actives
- Messages envoyés/reçus
- Erreurs
- Événements de frappe
- Statuts utilisateur
- Rate limiting

Accès aux métriques (admin uniquement):
```bash
curl -H "Authorization: Bearer ADMIN_TOKEN" \
     http://localhost:8080/api/v1/ws/admin/metrics
```
