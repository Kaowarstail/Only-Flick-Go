# Configuration du Système de Messagerie OnlyFlick Go

## Vue d'ensemble

Ce document décrit l'intégration complète du système de messagerie OnlyFlick côté Go backend, incluant les services, handlers, routes et configuration de base de données.

## Structure du Système

### 1. Architecture des Services

```
internal/services/
├── conversation_service.go      # Gestion des conversations
├── message_service.go          # Gestion des messages
├── messaging_service.go        # Service orchestrateur
├── messaging_service_manager.go # Gestionnaire de dépendances
└── errors.go                   # Gestion des erreurs
```

### 2. Modèles et DTOs

```
models/
├── messaging.go                # Modèles GORM principales
└── messaging_dto.go           # DTOs pour l'API
```

### 3. Utilitaires

```
internal/utils/
├── messaging_utils.go         # Utilitaires spécifiques
└── validation.go             # Validation étendue
```

### 4. Handlers et Routes

```
internal/handlers/
└── messaging_handler.go       # Handlers HTTP

internal/routes/
└── messaging_routes.go        # Configuration des routes
```

## Configuration de Base de Données

### Migration

Exécuter la migration SQL :
```bash
psql -d onlyflick_db -f migrations/003_create_messaging_system.sql
```

### Tables Créées

- `conversations_classic` - Conversations principales
- `conversation_classic_participants` - Participants aux conversations
- `messages_classic` - Messages
- `conversation_classic_read_status` - Statuts de lecture

## Intégration dans main.go

### 1. Import des Routes

```go
import (
    "github.com/Kaowarstail/Only-Flick-Go/internal/routes"
)
```

### 2. Configuration des Routes

Dans votre fonction de configuration des routes, ajouter :

```go
// Setup messaging routes
routes.SetupMessagingRoutes(router, db)
```

### 3. Exemple d'intégration complète

```go
func setupRoutes(router *mux.Router, db *gorm.DB) {
    // Routes existantes...
    setupAuthRoutes(router, db)
    setupUserRoutes(router, db)
    setupContentRoutes(router, db)
    
    // Nouvelle intégration messagerie
    routes.SetupMessagingRoutes(router, db)
    
    // Autres routes...
}
```

## Endpoints API Disponibles

### Conversations
- `GET /api/conversations` - Liste des conversations
- `POST /api/conversations` - Créer/récupérer conversation
- `GET /api/conversations/{id}/messages` - Messages d'une conversation
- `PUT /api/conversations/{id}/read` - Marquer comme lu
- `PUT /api/conversations/read-all` - Marquer tout comme lu

### Messages
- `POST /api/messages` - Envoyer un message

### Dashboard & Stats
- `GET /api/messaging/dashboard` - Tableau de bord
- `GET /api/messaging/stats` - Statistiques

### Recherche
- `GET /api/messaging/search` - Recherche globale

### Opérations Avancées
- `POST /api/messaging/start` - Démarrer conversation + message

## Utilisation des Services

### Exemple d'utilisation directe

```go
package main

import (
    "github.com/Kaowarstail/Only-Flick-Go/internal/services"
    "gorm.io/gorm"
)

func exampleUsage(db *gorm.DB) {
    // Créer le gestionnaire de services
    manager := services.NewMessagingServiceManager(db)
    
    // Récupérer le service principal
    messagingService := manager.GetMessagingService()
    
    // Utiliser les services
    conversations, err := messagingService.GetUserConversations("user-id", 1, 20)
    if err != nil {
        // Gérer l'erreur
    }
    
    // Envoyer un message
    request := &services.SendMessageRequest{
        ConversationID: "conv-id",
        Content:        stringPtr("Hello World!"),
        MessageType:    "text",
    }
    
    message, err := messagingService.SendMessage(request, "user-id")
    if err != nil {
        // Gérer l'erreur
    }
}

func stringPtr(s string) *string {
    return &s
}
```

## Configuration de l'Environnement

### Variables d'Environnement Recommandées

```env
# Base de données
DB_HOST=localhost
DB_PORT=5432
DB_NAME=onlyflick_db
DB_USER=your_user
DB_PASSWORD=your_password

# Messagerie
MESSAGING_MAX_MESSAGE_LENGTH=5000
MESSAGING_MAX_FILE_SIZE=50MB
MESSAGING_ALLOWED_MEDIA_TYPES=image/jpeg,image/png,video/mp4

# CORS (si nécessaire)
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://onlyflick.com
```

## Validation et Sécurité

### Authentification
- Toutes les routes nécessitent un JWT valide
- L'ID utilisateur est extrait automatiquement du token

### Validation des Données
- Validation automatique des DTOs
- Nettoyage du contenu des messages
- Validation des types de fichiers média

### Permissions
- Les utilisateurs ne peuvent accéder qu'à leurs conversations
- Vérification automatique des permissions dans les services

## Tests et Débogage

### Logs de Débogage

Les services incluent des logs détaillés. Pour activer :

```go
import "log"

// Dans votre configuration
log.SetLevel(log.DebugLevel)
```

### Tests d'API avec curl

```bash
# Récupérer les conversations
curl -H "Authorization: Bearer YOUR_JWT" \
     http://localhost:8080/api/conversations

# Envoyer un message
curl -X POST \
     -H "Authorization: Bearer YOUR_JWT" \
     -H "Content-Type: application/json" \
     -d '{"conversation_id":"conv-id","content":"Hello","message_type":"text"}' \
     http://localhost:8080/api/messages
```

## Performance et Optimisation

### Index de Base de Données
- Index optimisés pour les requêtes fréquentes
- Index de recherche full-text sur le contenu
- Index composites pour les requêtes paginées

### Pagination
- Pagination par défaut : 20 éléments
- Limite maximale : 100 éléments
- Curseur basé sur created_at pour la performance

### Cache (Recommandations futures)
- Cache Redis pour les conversations actives
- Cache des statuts de lecture
- Cache des compteurs de messages non lus

## Maintenance

### Nettoyage Automatique
- Fonction SQL `cleanup_old_messaging_data()` disponible
- Suppression des messages soft-deleted anciens
- Peut être exécutée via cron

### Monitoring
- Surveiller les tables de messages pour la croissance
- Surveiller les performances des requêtes de recherche
- Alertes sur les erreurs de service

## Extensibilité Future

Le système est conçu pour supporter facilement :
- Messages de groupe étendus
- Réactions aux messages
- Messages vocaux
- Chiffrement end-to-end
- Notifications push
- Synchronisation temps réel (WebSockets)

## Support et Documentation

- Code entièrement documenté en français
- Erreurs avec codes HTTP appropriés
- Logs détaillés pour le débogage
- Architecture modulaire pour les extensions

## Checklist d'Intégration

- [ ] Migration de base de données exécutée
- [ ] Routes configurées dans main.go
- [ ] Variables d'environnement configurées
- [ ] Tests d'API effectués
- [ ] Authentification JWT fonctionnelle
- [ ] Permissions vérifiées
- [ ] Logs de débogage configurés
- [ ] Documentation équipe mise à jour
