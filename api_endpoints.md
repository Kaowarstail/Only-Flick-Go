# API Endpoints - OnlyFlick

Ce document détaille tous les endpoints de l'API Rest OnlyFlick, organisés par domaine fonctionnel.

## Table des matières
1. [Authentification](#authentification)
2. [Gestion des Utilisateurs](#gestion-des-utilisateurs)
3. [Gestion des Créateurs](#gestion-des-créateurs)
4. [Gestion du Contenu](#gestion-du-contenu)
5. [Abonnements](#abonnements)
6. [Interactions Sociales](#interactions-sociales)
7. [Modération](#moderation)
8. [Notifications](#notifications)
9. [Paiements](#paiements)

## Base URL

Tous les endpoints commencent par la base : `/api/v1`

## Authentification

| Méthode | Endpoint             | Description                                      | Accès         |
|---------|----------------------|--------------------------------------------------|---------------|
| POST    | `/auth/register`     | Inscription d'un nouvel utilisateur              | Public        |
| POST    | `/auth/login`        | Connexion d'un utilisateur                       | Public        |
| POST    | `/auth/logout`       | Déconnexion de l'utilisateur courant             | Authentifié   |
| POST    | `/auth/refresh-token`| Rafraîchir le token d'authentification           | Authentifié   |
| POST    | `/auth/reset-password`| Demander une réinitialisation de mot de passe   | Public        |
| PUT     | `/auth/reset-password/:token`| Réinitialiser le mot de passe avec token | Public        |
| GET     | `/auth/me`           | Obtenir les informations de l'utilisateur courant| Authentifié   |

## Gestion des Utilisateurs

| Méthode | Endpoint                 | Description                                 | Accès                    |
|---------|--------------------------|---------------------------------------------|--------------------------|
| GET     | `/users`                 | Liste des utilisateurs (avec filtres)       | Admin                    |
| GET     | `/users/:id`             | Détails d'un utilisateur                    | Authentifié (restrictions)|
| PUT     | `/users/:id`             | Mise à jour d'un utilisateur                | Propriétaire/Admin       |
| DELETE  | `/users/:id`             | Suppression d'un utilisateur                | Propriétaire/Admin       |
| PUT     | `/users/:id/profile-pic` | Téléchargement d'une photo de profil        | Propriétaire/Admin       |
| GET     | `/users/:id/following`   | Liste des créateurs suivis                  | Propriétaire/Admin       |
| POST    | `/users/:id/block/:targetId` | Bloquer un utilisateur                  | Propriétaire             |
| DELETE  | `/users/:id/block/:targetId` | Débloquer un utilisateur                | Propriétaire             |
| GET     | `/users/:id/blocked`     | Liste des utilisateurs bloqués              | Propriétaire             |

## Gestion des Créateurs

| Méthode | Endpoint                   | Description                               | Accès                    |
|---------|-----------------------------|-------------------------------------------|--------------------------|
| POST    | `/creators`                 | Demander le statut de créateur            | Abonné                   |
| GET     | `/creators`                 | Liste des créateurs (avec filtres)        | Public                   |
| GET     | `/creators/:id`             | Profil et statistiques d'un créateur      | Public (avec restrictions)|
| PUT     | `/creators/:id`             | Mise à jour d'un profil créateur          | Propriétaire             |
| PUT     | `/creators/:id/banner`      | Téléchargement d'une bannière             | Propriétaire             |
| GET     | `/creators/:id/subscribers` | Liste des abonnés (pour le créateur)      | Propriétaire             |
| GET     | `/creators/:id/stats`       | Statistiques détaillées pour le créateur  | Propriétaire             |
| GET     | `/creators/featured`        | Liste des créateurs mis en avant          | Public                   |
| GET     | `/creators/search`          | Recherche de créateurs                    | Public                   |

## Gestion du Contenu

| Méthode | Endpoint                     | Description                             | Accès                    |
|---------|------------------------------|-----------------------------------------|--------------------------|
| POST    | `/contents`                  | Création d'un nouveau contenu           | Créateur                 |
| GET     | `/contents`                  | Liste des contenus (avec filtres)       | Public (avec restrictions)|
| GET     | `/contents/:id`              | Détails d'un contenu spécifique         | Public/Abonnés selon contenu |
| PUT     | `/contents/:id`              | Mise à jour d'un contenu                | Propriétaire/Admin       |
| DELETE  | `/contents/:id`              | Suppression d'un contenu                | Propriétaire/Admin       |
| POST    | `/contents/:id/media`        | Téléchargement des médias du contenu    | Créateur                 |
| POST    | `/contents/:id/thumbnail`    | Téléchargement d'une miniature          | Créateur                 |
| GET     | `/contents/search`           | Recherche de contenus                   | Public (avec restrictions)|
| GET     | `/contents/trending`         | Contenus tendances                      | Public                   |
| GET     | `/creators/:id/contents`     | Liste des contenus d'un créateur        | Public (avec restrictions)|
| GET     | `/users/:id/feed`            | Fil d'actualité personnalisé            | Authentifié              |

## Abonnements

| Méthode | Endpoint                        | Description                          | Accès                    |
|---------|---------------------------------|--------------------------------------|--------------------------|
| POST    | `/subscription-plans`           | Création d'un plan d'abonnement      | Créateur                 |
| GET     | `/subscription-plans`           | Liste des plans d'abonnement         | Public                   |
| GET     | `/subscription-plans/:id`       | Détails d'un plan d'abonnement       | Public                   |
| PUT     | `/subscription-plans/:id`       | Modification d'un plan d'abonnement  | Propriétaire/Admin       |
| DELETE  | `/subscription-plans/:id`       | Suppression d'un plan d'abonnement   | Propriétaire/Admin       |
| GET     | `/creators/:id/subscription-plans` | Plans d'abonnement d'un créateur  | Public                   |
| POST    | `/subscriptions`                | Création d'un abonnement             | Authentifié              |
| GET     | `/subscriptions`                | Liste des abonnements de l'utilisateur | Authentifié            |
| GET     | `/subscriptions/:id`            | Détails d'un abonnement              | Propriétaire/Admin       |
| PUT     | `/subscriptions/:id`            | Modification d'un abonnement         | Propriétaire/Admin       |
| DELETE  | `/subscriptions/:id`            | Annulation d'un abonnement           | Propriétaire/Admin       |
| PUT     | `/subscriptions/:id/renew`      | Renouvellement d'un abonnement       | Propriétaire             |

## Interactions Sociales

| Méthode | Endpoint                      | Description                          | Accès                    |
|---------|-------------------------------|--------------------------------------|--------------------------|
| POST    | `/contents/:id/comments`      | Ajout d'un commentaire               | Authentifié              |
| GET     | `/contents/:id/comments`      | Liste des commentaires sur un contenu| Public (selon contenu)   |
| PUT     | `/comments/:id`               | Modification d'un commentaire        | Propriétaire/Admin       |
| DELETE  | `/comments/:id`               | Suppression d'un commentaire         | Propriétaire/Admin/Créateur |
| POST    | `/contents/:id/likes`         | Liker un contenu                     | Authentifié              |
| DELETE  | `/contents/:id/likes`         | Retirer un like                      | Authentifié              |
| GET     | `/contents/:id/likes`         | Liste des likes d'un contenu         | Propriétaire/Admin       |
| POST    | `/messages`                   | Envoi d'un message                   | Authentifié              |
| GET     | `/messages`                   | Liste des conversations              | Authentifié              |
| GET     | `/messages/:userId`           | Messages avec un utilisateur         | Authentifié              |
| PUT     | `/messages/:id/read`          | Marquer un message comme lu          | Destinataire             |
| DELETE  | `/messages/:id`               | Supprimer un message                 | Propriétaire/Admin       |

## Modération

| Méthode | Endpoint                      | Description                          | Accès                    |
|---------|-------------------------------|--------------------------------------|--------------------------|
| POST    | `/reports`                    | Signaler un contenu                  | Authentifié              |
| GET     | `/reports`                    | Liste des signalements               | Admin                    |
| GET     | `/reports/:id`                | Détails d'un signalement             | Admin                    |
| PUT     | `/reports/:id`                | Traiter un signalement               | Admin                    |
| PUT     | `/users/:id/ban`              | Bannir un utilisateur                | Admin                    |
| PUT     | `/users/:id/unban`            | Réintégrer un utilisateur            | Admin                    |
| GET     | `/audit-logs`                 | Journaux d'audit des actions admin   | Admin                    |

## Notifications

| Méthode | Endpoint                       | Description                         | Accès                    |
|---------|--------------------------------|-------------------------------------|--------------------------|
| GET     | `/notifications`               | Liste des notifications             | Authentifié              |
| PUT     | `/notifications/:id/read`      | Marquer une notification comme lue  | Propriétaire             |
| PUT     | `/notifications/read-all`      | Marquer toutes les notifs comme lues| Authentifié              |
| GET     | `/notifications/unread-count`  | Nombre de notifications non lues    | Authentifié              |
| PUT     | `/users/:id/notification-settings` | Paramètres de notification      | Propriétaire             |

## Paiements

| Méthode | Endpoint                        | Description                        | Accès                    |
|---------|---------------------------------|------------------------------------|--------------------------|
| GET     | `/payments/methods`             | Méthodes de paiement de l'utilisateur | Authentifié           |
| POST    | `/payments/methods`             | Ajout d'une méthode de paiement    | Authentifié              |
| DELETE  | `/payments/methods/:id`         | Suppression d'une méthode de paiement | Propriétaire          |
| GET     | `/transactions`                 | Historique des transactions        | Authentifié              |
| GET     | `/transactions/:id`             | Détails d'une transaction          | Propriétaire/Admin       |
| GET     | `/creators/:id/earnings`        | Revenus d'un créateur              | Propriétaire             |
| POST    | `/payouts/request`              | Demande de versement               | Créateur                 |
| GET     | `/payouts`                      | Historique des versements          | Créateur                 |
| GET     | `/payouts/:id`                  | Détails d'un versement             | Propriétaire/Admin       |

## Légende d'accès

- **Public** : Accessible sans authentification
- **Authentifié** : Nécessite d'être connecté
- **Propriétaire** : Utilisateur propriétaire de la ressource ou de son compte
- **Créateur** : Utilisateur avec le rôle de créateur
- **Abonné** : Utilisateur avec le rôle d'abonné
- **Admin** : Utilisateur avec le rôle d'administrateur

## Codes d'état HTTP

- **200 OK** : Requête réussie
- **201 Created** : Ressource créée avec succès
- **204 No Content** : Requête réussie sans contenu à renvoyer (ex: suppression)
- **400 Bad Request** : Erreur dans les données fournies
- **401 Unauthorized** : Authentification nécessaire ou échouée
- **403 Forbidden** : Authentifié mais pas autorisé à accéder à la ressource
- **404 Not Found** : Ressource non trouvée
- **409 Conflict** : Conflit avec l'état actuel de la ressource
- **429 Too Many Requests** : Trop de requêtes (rate limiting)
- **500 Internal Server Error** : Erreur serveur

## Format des réponses

Toutes les réponses suivent le format standard suivant :

```json
{
  "success": true,
  "data": { /* données de la réponse */ },
  "message": "Description optionnelle du résultat",
  "meta": {
    "pagination": {
      "total": 100,
      "page": 1,
      "per_page": 20,
      "pages": 5
    }
  }
}
```

En cas d'erreur :

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Description de l'erreur"
  }
}
```

## Pagination

Les endpoints retournant des listes supportent la pagination via les paramètres suivants :

- `page` : Numéro de page (défaut: 1)
- `per_page` : Nombre d'éléments par page (défaut: 20, max: 100)

## Filtrage et tri

Les endpoints de liste supportent les paramètres suivants :

- `sort` : Champ de tri (ex: `created_at`)
- `order` : Ordre de tri (`asc` ou `desc`, défaut: `desc`)
- Paramètres spécifiques de filtrage selon l'endpoint
