# OnlyFlick WebSocket Implementation

## Vue d'ensemble

Ce projet implémente un système WebSocket temps réel complet pour OnlyFlick, permettant:

- **Messages instantanés** avec broadcasting temps réel
- **Indicateurs de frappe** (typing indicators)
- **Statut de présence** (online/offline)
- **Messages payants** avec notifications de déblocage
- **Gestion des conversations** avec mises à jour en temps réel
- **Rate limiting** et sécurité avancée

## Architecture

### Structure du Projet

```
internal/
├── websocket/
│   ├── hub.go          # Gestionnaire principal des connexions
│   ├── client.go       # Client WebSocket individuel
├── handlers/
│   ├── websocket.go    # Handler REST pour WebSocket
│   ├── message.go      # Handler messages avec intégration WebSocket
├── middleware/
│   ├── gin_middleware.go # Middlewares pour Gin
├── utils/
│   ├── rate_limiter.go      # Limitation de taux
│   ├── websocket_metrics.go # Métriques WebSocket
├── config/
│   └── websocket.go    # Configuration WebSocket
├── routes/
│   ├── websocket.go    # Routes WebSocket
│   └── gin_routes.go   # Routes Gin
models/
└── websocket.go        # Modèles WebSocket
```

### Composants Principaux

#### 1. Hub WebSocket (`internal/websocket/hub.go`)
- Gère toutes les connexions WebSocket
- Broadcasting intelligent par conversation/utilisateur
- Gestion de la présence utilisateur
- Nettoyage automatique des connexions

#### 2. Client WebSocket (`internal/websocket/client.go`)
- Représente une connexion WebSocket individuelle
- Gère les événements de frappe
- Heartbeat et détection de déconnexion
- Validation des permissions

#### 3. Intégration REST (`internal/handlers/websocket.go`)
- API REST complémentaire pour les informations WebSocket
- Métriques et monitoring
- Gestion des utilisateurs en ligne

## Fonctionnalités Implémentées

### Messages Temps Réel
- ✅ Broadcasting automatique des nouveaux messages
- ✅ Support des messages payants
- ✅ Notifications de déblocage
- ✅ Métadonnées complètes (sender, conversation)

### Indicateurs de Frappe
- ✅ Événements `user_typing` et `user_stopped_typing`
- ✅ Auto-stop après 3 secondes
- ✅ Validation des permissions par conversation

### Présence Utilisateur
- ✅ Statut online/offline automatique
- ✅ Dernière activité trackée
- ✅ Broadcasting aux contacts

### Sécurité
- ✅ Authentification JWT obligatoire
- ✅ Rate limiting (100 messages/minute)
- ✅ Validation des origines
- ✅ Permissions par conversation

### Performance
- ✅ Heartbeat/Pong pour maintenir les connexions
- ✅ Nettoyage automatique des connexions mortes
- ✅ Métriques complètes
- ✅ Limitation des connexions simultanées

## Configuration

### Variables d'Environnement

```bash
# WebSocket
WS_READ_BUFFER_SIZE=1024
WS_WRITE_BUFFER_SIZE=1024
WS_MESSAGE_RATE_LIMIT=100
WS_CONNECTION_LIMIT=1000
WS_ALLOWED_ORIGINS=https://onlyflick.app,http://localhost:3000
WS_ENABLE_TYPING=true
WS_ENABLE_PRESENCE=true
WS_TYPING_TIMEOUT=3s
WS_INACTIVITY_TIMEOUT=30m
WS_CLEANUP_INTERVAL=5m
```

### Configuration par Défaut

```go
config := WebSocketConfig{
    ReadBufferSize:         1024,
    WriteBufferSize:        1024,
    MessageRateLimit:       100,
    ConnectionLimit:        1000,
    EnableTypingIndicators: true,
    EnablePresence:         true,
    TypingTimeout:          3 * time.Second,
    InactivityTimeout:      30 * time.Minute,
    CleanupInterval:        5 * time.Minute,
}
```

## API WebSocket

### Connexion
```
ws://localhost:8080/api/v1/ws
```

### Événements Supportés

#### Entrants (Client → Serveur)
- `user_typing` - Indicateur de frappe
- `user_stopped_typing` - Arrêt de frappe
- `user_active_in_conversation` - Activité dans une conversation
- `heartbeat` - Battement de cœur

#### Sortants (Serveur → Client)
- `message_sent` - Nouveau message
- `paid_message_sent` - Message payant envoyé
- `paid_message_unlocked` - Message payant débloqué
- `user_typing` - Utilisateur en train d'écrire
- `user_stopped_typing` - Utilisateur a arrêté d'écrire
- `user_online` - Utilisateur en ligne
- `user_offline` - Utilisateur hors ligne
- `conversation_updated` - Conversation mise à jour
- `connection_established` - Connexion établie
- `error` - Erreur

## API REST Complémentaire

### Informations WebSocket
```http
GET /api/v1/ws/info
Authorization: Bearer JWT_TOKEN
```

### Utilisateurs En Ligne
```http
GET /api/v1/ws/online
Authorization: Bearer JWT_TOKEN
```

### Statut Utilisateur
```http
GET /api/v1/ws/user/{user_id}/status
Authorization: Bearer JWT_TOKEN
```

### Métriques (Admin)
```http
GET /api/v1/ws/admin/metrics
Authorization: Bearer ADMIN_JWT_TOKEN
```

## Intégration avec l'API REST

### Handler Message Modifié
```go
// SendMessage avec broadcasting WebSocket
func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
    // ... envoi du message ...
    
    // Broadcasting WebSocket
    if h.hub != nil && conversation != nil && sender != nil {
        h.hub.BroadcastMessageSent(message, conversation, sender)
    }
}
```

### Handler Paid Message
```go
// UnlockPaidMessage avec notification WebSocket
func (h *MessageHandler) UnlockPaidMessage(w http.ResponseWriter, r *http.Request) {
    // ... déblocage du message ...
    
    // Broadcasting WebSocket
    if h.hub != nil {
        h.hub.BroadcastPaidMessageUnlocked(message, userID, transaction)
    }
}
```

## Tests

### Client de Test Go
```bash
cd cmd/websocket-test
go run main.go -token YOUR_JWT_TOKEN
```

### Script de Test Bash
```bash
./test_websocket.sh YOUR_JWT_TOKEN
```

### Test Manuel avec wscat
```bash
npm install -g wscat
wscat -c ws://localhost:8080/api/v1/ws -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Métriques et Monitoring

### Métriques Collectées
- Connexions actives/ouvertes/fermées
- Messages envoyés/reçus
- Événements de frappe
- Statuts utilisateur
- Rate limiting
- Erreurs
- Performance (latence, taille des messages)

### Endpoints de Monitoring
```http
GET /api/v1/ws/health     # Health check
GET /api/v1/ws/admin/metrics  # Métriques complètes (admin)
```

## Exemple d'Utilisation Flutter

```dart
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
      handleWebSocketMessage(data);
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
  
  void handleWebSocketMessage(Map<String, dynamic> data) {
    switch (data['type']) {
      case 'message_sent':
        // Ajouter le message à l'interface
        break;
      case 'user_typing':
        // Afficher "John is typing..."
        break;
      case 'user_online':
        // Mettre à jour le statut utilisateur
        break;
      case 'paid_message_unlocked':
        // Mettre à jour le message débloqué
        break;
    }
  }
}
```

## Déploiement

### Préparation
1. Compiler l'application avec support WebSocket
2. Configurer les variables d'environnement
3. Ajuster les limits de connexion selon les besoins

### Nginx (Optionnel)
```nginx
location /api/v1/ws {
    proxy_pass http://127.0.0.1:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_read_timeout 86400;
}
```

## Limitations et Considérations

### Limitations Actuelles
- **Connexions simultanées**: 1000 par défaut
- **Rate limiting**: 100 messages/minute par utilisateur
- **Taille des messages**: 512 bytes maximum
- **Timeout inactivité**: 30 minutes

### Considérations de Performance
- **Mémoire**: ~1KB par connexion WebSocket
- **CPU**: Broadcasting optimal avec channels Go
- **Réseau**: Heartbeat toutes les 54 secondes

### Scaling
- Possibilité d'ajouter Redis pour le clustering
- Load balancing avec sticky sessions
- Monitoring des métriques en production

## Prochaines Améliorations

### Court Terme
- [ ] Intégration avec l'authentification JWT existante
- [ ] Tests d'intégration complets
- [ ] Logging structuré
- [ ] Configuration via fichier

### Long Terme
- [ ] Clustering avec Redis
- [ ] Notifications push mobiles
- [ ] Persistence des messages hors ligne
- [ ] Analytics temps réel

## Troubleshooting

### Problèmes Courants

1. **Connexion refusée**
   - Vérifier le token JWT
   - Vérifier l'origine (CORS)

2. **Messages non reçus**
   - Vérifier les permissions de conversation
   - Vérifier le rate limiting

3. **Déconnexions fréquentes**
   - Vérifier le heartbeat
   - Vérifier la stabilité réseau

### Debugging
```bash
# Logs WebSocket
tail -f /var/log/onlyflick/websocket.log

# Métriques en temps réel
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
     http://localhost:8080/api/v1/ws/admin/metrics
```

## Sécurité

### Mesures Implémentées
- ✅ Authentification JWT obligatoire
- ✅ Rate limiting anti-spam
- ✅ Validation des origines
- ✅ Permissions par conversation
- ✅ Sanitization des données

### Recommandations Production
- Utiliser HTTPS/WSS uniquement
- Configurer les origines autorisées
- Monitorer les tentatives d'abus
- Limits de ressources par utilisateur

## Support

Pour toute question ou problème:
1. Consulter la documentation dans `docs/WEBSOCKET_API.md`
2. Exécuter les tests avec `./test_websocket.sh`
3. Vérifier les logs d'erreur
4. Consulter les métriques de performance

L'implémentation WebSocket est maintenant complète et prête pour l'intégration avec le frontend Flutter ! 🚀
