# OnlyFlick API - Messagerie et Édition de Profil

## Vue d'ensemble

Cette implémentation ajoute les fonctionnalités de messagerie avec messages payants et d'édition de profil à l'API OnlyFlick existante. L'architecture suit les principes de l'API existante en utilisant PostgreSQL, JWT, et Gin.

## Nouvelles fonctionnalités

### 🔒 Messagerie
- **Conversations privées** entre utilisateurs
- **Messages payants** avec système de déblocage
- **Upload de médias** (images, vidéos, audio, documents)
- **Système de commission** 20% automatique sur les messages payants
- **Notifications** et statut de lecture des messages
- **Statistiques** de messagerie détaillées

### 👤 Édition de profil
- **Profils utilisateur** complets avec photos et informations
- **Liens sociaux** (Instagram, Twitter, TikTok, YouTube, etc.)
- **Statistiques utilisateur** (followers, messages, etc.)
- **Profils créateur** avec paramètres de monétisation
- **Gains et revenus** avec historique mensuel
- **Gestion des abonnements** et prix personnalisés

## Architecture

### Structure du projet

```
Only-Flick-Go/
├── cmd/api/main.go                              # Point d'entrée
├── internal/
│   ├── config/
│   │   └── messaging.go                         # Configuration messagerie
│   ├── database/
│   │   └── migrations/
│   │       └── 002_messaging_profile_tables.sql # Nouvelles tables
│   ├── handlers/
│   │   ├── conversation.go                      # Gestion conversations
│   │   ├── message.go                           # Gestion messages
│   │   ├── media.go                             # Gestion médias
│   │   ├── users.go                             # Extensions profil utilisateur
│   │   └── creators.go                          # Extensions profil créateur
│   ├── middleware/
│   │   └── messaging.go                         # Middleware messagerie
│   ├── routes/
│   │   └── messaging.go                         # Routes messagerie
│   ├── services/
│   │   ├── conversation_service.go              # Logique conversations
│   │   └── message_service.go                   # Logique messages
│   ├── utils/
│   │   └── context.go                           # Utilitaires contexte
│   └── validators/
│       └── messaging.go                         # Validation données
├── models/
│   └── messaging.go                             # Modèles messagerie
├── uploads/                                     # Stockage fichiers
├── API_EXAMPLES.md                              # Documentation API
├── test_api.sh                                  # Script de test
└── README_MESSAGING.md                          # Ce fichier
```

### Nouvelles tables de base de données

1. **conversations** - Conversations entre utilisateurs
2. **messages** - Messages avec support payant
3. **social_links** - Liens sociaux utilisateur
4. **user_stats** - Statistiques utilisateur
5. **paid_message_transactions** - Transactions messages payants
6. **creator_earnings** - Gains créateur
7. **monthly_earnings** - Gains mensuels
8. **media_files** - Fichiers uploadés

## Installation et configuration

### 1. Migration de la base de données

```bash
# Appliquer la migration
psql -U your_user -d only_flick_db -f internal/database/migrations/002_messaging_profile_tables.sql
```

### 2. Configuration environnement

```bash
# Variables d'environnement
export JWT_SECRET="your-secret-key-change-this-in-production"
export JWT_EXPIRATION_HOURS=24
export MAX_FILE_SIZE=52428800  # 50MB
export UPLOAD_PATH="uploads"
export MIN_PAID_MESSAGE_PRICE=0.99
export MAX_PAID_MESSAGE_PRICE=500.00
```

### 3. Créer les dossiers d'upload

```bash
mkdir -p uploads/messages
mkdir -p uploads/avatars
mkdir -p uploads/banners
mkdir -p uploads/thumbnails
```

### 4. Démarrer l'API

```bash
cd cmd/api
go run main.go
```

## Utilisation

### Authentification

Tous les endpoints nécessitent un token JWT dans l'en-tête :

```bash
Authorization: Bearer <your-jwt-token>
```

### Exemples d'utilisation

#### Envoyer un message

```bash
curl -X POST http://localhost:8080/api/v1/messaging/messages \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "conversation_id": "conv-id",
    "content": "Hello!",
    "message_type": "text"
  }'
```

#### Envoyer un message payant

```bash
curl -X POST http://localhost:8080/api/v1/messaging/messages \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "conversation_id": "conv-id",
    "content": "Exclusive content!",
    "message_type": "paid_text",
    "is_paid": true,
    "price": 5.99,
    "preview_text": "Preview of exclusive content..."
  }'
```

#### Uploader un fichier

```bash
curl -X POST http://localhost:8080/api/v1/messaging/media/upload \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -F "file=@photo.jpg"
```

#### Mettre à jour le profil

```bash
curl -X PUT http://localhost:8080/api/v1/profiles/users/me \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "biography": "Content creator",
    "location": "Paris",
    "social_links": {
      "instagram": "john",
      "twitter": "john"
    }
  }'
```

## Tests

### Test automatique

```bash
# Rendre le script exécutable
chmod +x test_api.sh

# Exécuter les tests
./test_api.sh -t "your-jwt-token"
```

### Test manuel

Consultez `API_EXAMPLES.md` pour des exemples détaillés d'appels API.

## Sécurité

### Authentification et autorisation
- **JWT obligatoire** pour tous les endpoints
- **Validation des permissions** sur les conversations
- **Vérification de propriété** des fichiers
- **Rate limiting** pour éviter le spam

### Validation des données
- **Validation stricte** des messages et fichiers
- **Sanitization** des entrées utilisateur
- **Contrôle des types** de fichiers autorisés
- **Limites de taille** pour les uploads

### Transactions financières
- **Commission automatique** 20% sur messages payants
- **Vérification des doublons** de paiement
- **Logging des transactions** pour audit
- **Gestion des erreurs** de paiement

## Monétisation

### Messages payants
- Prix minimum : 0.99€
- Prix maximum : 500€
- Commission plateforme : 20%
- Gain créateur : 80%

### Calcul automatique
```sql
-- Trigger automatique pour calculer les frais
CREATE TRIGGER calculate_transaction_fees 
BEFORE INSERT ON paid_message_transactions
FOR EACH ROW EXECUTE FUNCTION calculate_platform_fee();
```

## Performance

### Optimisations base de données
- **Index optimisés** sur les requêtes fréquentes
- **Pagination** obligatoire pour les listes
- **Connection pooling** PostgreSQL
- **Triggers** pour mise à jour automatique

### Gestion des fichiers
- **Stockage local** (extensible vers S3)
- **Génération de thumbnails** pour images
- **Validation des types** MIME
- **Limitation de taille** par type de fichier

## Monitoring

### Logs
- **Journalisation** des actions importantes
- **Tracking** des erreurs et exceptions
- **Audit** des transactions financières
- **Métriques** d'utilisation

### Métriques
- Nombre de messages envoyés
- Volume de transactions
- Taux de conversion messages payants
- Utilisation du stockage

## Limites et contraintes

### Messages
- Longueur maximale : 5000 caractères
- Aperçu payant : 100 caractères
- Rate limit : 30 messages/minute
- Rétention : 365 jours

### Fichiers
- Taille maximale globale : 50MB
- Images : 10MB max
- Vidéos : 100MB max
- Audio : 20MB max
- Documents : 10MB max

### Profils
- Biographie : 1000 caractères max
- Username : 3-30 caractères
- Liens sociaux : 5 plateformes max

## Extensibilité

### Webhooks
Préparation pour webhooks futurs :
- Nouveaux messages reçus
- Messages payants débloqués
- Nouveaux abonnés
- Seuils de gains atteints

### Intégrations
- Systèmes de paiement (Stripe, PayPal)
- Stockage cloud (AWS S3, Google Cloud)
- Services de transcoding vidéo
- Notifications push

## Maintenance

### Migrations futures
- Système de migration versionnées
- Rollback automatique en cas d'erreur
- Backup automatique avant migration
- Tests d'intégrité post-migration

### Nettoyage
- Suppression des fichiers orphelins
- Archivage des anciennes conversations
- Purge des données expirées
- Optimisation des index

## Support

### Documentation
- **API_EXAMPLES.md** : Exemples d'appels API
- **test_api.sh** : Script de test automatique
- **README_MESSAGING.md** : Documentation complète

### Débogage
- Logs détaillés pour chaque opération
- Codes d'erreur spécifiques
- Messages d'erreur explicites
- Validation côté client et serveur

## Roadmap

### Phase 1 ✅
- [x] Messagerie de base
- [x] Messages payants
- [x] Upload de médias
- [x] Édition de profil

### Phase 2 🔄
- [ ] Notifications push
- [ ] Webhooks
- [ ] Intégration paiements
- [ ] Modération automatique

### Phase 3 📋
- [ ] Groupes de discussion
- [ ] Messages temporaires
- [ ] Streaming en direct
- [ ] Marketplace de contenus

---

Cette implémentation fournit une base solide et extensible pour les fonctionnalités de messagerie et d'édition de profil d'OnlyFlick, avec une architecture robuste et sécurisée prête pour la production.
