# Schéma de la Base de Données - OnlyFlick

Ce document présente le schéma de la base de données pour l'application OnlyFlick, incluant la description des tables, leurs champs et les relations entre elles.

## Vue d'ensemble

La base de données OnlyFlick est conçue pour prendre en charge un réseau social où des créateurs de contenu peuvent publier et monétiser leur contenu auprès d'abonnés payants. Le schéma comprend des tables pour les utilisateurs, le contenu, les abonnements, les interactions, les paiements et la modération.

## Tables Principales

### Users

La table centrale qui stocke tous les utilisateurs (créateurs, abonnés et administrateurs).

| Colonne          | Type          | Description                                      |
|------------------|---------------|--------------------------------------------------|
| id               | uint (PK)     | Identifiant unique                               |
| username         | string        | Nom d'utilisateur unique                         |
| email            | string        | Adresse email unique                             |
| password         | string        | Mot de passe hashé                               |
| first_name       | string        | Prénom                                           |
| last_name        | string        | Nom de famille                                   |
| role             | string enum   | Rôle: admin, creator, subscriber                 |
| biography        | text          | Biographie/À propos                              |
| profile_picture  | string        | URL de la photo de profil                        |
| is_active        | bool          | Statut d'activité du compte                      |
| is_banned        | bool          | Indique si l'utilisateur est banni               |
| ban_reason       | string        | Raison du bannissement                           |
| last_login       | timestamp     | Date et heure de la dernière connexion           |
| created_at       | timestamp     | Date et heure de création                        |
| updated_at       | timestamp     | Date et heure de dernière modification           |

### CreatorProfiles

Contient des informations supplémentaires pour les utilisateurs qui sont des créateurs.

| Colonne          | Type          | Description                                      |
|------------------|---------------|--------------------------------------------------|
| id               | uint (PK)     | Identifiant unique                               |
| user_id          | uint (FK)     | Référence à l'utilisateur                        |
| banner_image     | string        | URL de l'image de bannière                       |
| website_url      | string        | Site web du créateur                             |
| social_links     | string (JSON) | Liens vers les réseaux sociaux                   |
| created_at       | timestamp     | Date et heure de création                        |
| updated_at       | timestamp     | Date et heure de dernière modification           |

### Contents

Stocke tous les contenus publiés par les créateurs.

| Colonne          | Type          | Description                                      |
|------------------|---------------|--------------------------------------------------|
| id               | uint (PK)     | Identifiant unique                               |
| creator_id       | uint (FK)     | Référence au créateur                            |
| title            | string        | Titre du contenu                                 |
| description      | text          | Description du contenu                           |
| type             | string        | Type de contenu (image, vidéo, texte...)         |
| media_url        | string        | URL du média                                     |
| thumbnail_url    | string        | URL de la miniature                              |
| is_premium       | bool          | Indique si c'est du contenu premium              |
| is_published     | bool          | Indique si le contenu est publié                 |
| view_count       | int           | Nombre de vues                                   |
| is_flagged       | bool          | Indique si le contenu a été signalé             |
| created_at       | timestamp     | Date et heure de création                        |
| updated_at       | timestamp     | Date et heure de dernière modification           |
| deleted_at       | timestamp     | Date de suppression (soft delete)                |

### SubscriptionPlans

Définit les différents plans d'abonnement proposés par les créateurs.

| Colonne          | Type          | Description                                      |
|------------------|---------------|--------------------------------------------------|
| id               | uint (PK)     | Identifiant unique                               |
| creator_id       | uint (FK)     | Référence au créateur                            |
| name             | string        | Nom du plan                                      |
| description      | text          | Description du plan                              |
| price            | float         | Prix de l'abonnement                             |
| duration         | int           | Durée en jours                                   |
| is_active        | bool          | Indique si le plan est actif                     |
| created_at       | timestamp     | Date et heure de création                        |
| updated_at       | timestamp     | Date et heure de dernière modification           |

### Subscriptions

Enregistre les abonnements des utilisateurs aux plans des créateurs.

| Colonne          | Type          | Description                                      |
|------------------|---------------|--------------------------------------------------|
| id               | uint (PK)     | Identifiant unique                               |
| subscriber_id    | uint (FK)     | Référence à l'abonné                             |
| creator_id       | uint (FK)     | Référence au créateur                            |
| plan_id          | uint (FK)     | Référence au plan d'abonnement                   |
| start_date       | timestamp     | Date de début de l'abonnement                    |
| end_date         | timestamp     | Date de fin de l'abonnement                      |
| is_active        | bool          | Statut d'activité de l'abonnement                |
| auto_renew       | bool          | Renouvellement automatique                       |
| payment_status   | string        | Statut du paiement                               |
| transaction_id   | string        | Identifiant de la transaction                    |
| created_at       | timestamp     | Date et heure de création                        |
| updated_at       | timestamp     | Date et heure de dernière modification           |

## Tables d'Interactions

### Comments

Stocke les commentaires laissés par les utilisateurs sur les contenus.

| Colonne          | Type          | Description                                      |
|------------------|---------------|--------------------------------------------------|
| id               | uint (PK)     | Identifiant unique                               |
| content_id       | uint (FK)     | Référence au contenu                             |
| user_id          | uint (FK)     | Référence à l'utilisateur                        |
| text             | text          | Texte du commentaire                             |
| is_hidden        | bool          | Indique si le commentaire est masqué             |
| created_at       | timestamp     | Date et heure de création                        |
| updated_at       | timestamp     | Date et heure de dernière modification           |
| deleted_at       | timestamp     | Date de suppression (soft delete)                |

### Likes

Enregistre les "j'aime" des utilisateurs sur les contenus.

| Colonne          | Type          | Description                                      |
|------------------|---------------|--------------------------------------------------|
| id               | uint (PK)     | Identifiant unique                               |
| content_id       | uint (FK)     | Référence au contenu                             |
| user_id          | uint (FK)     | Référence à l'utilisateur                        |
| created_at       | timestamp     | Date et heure de création                        |

### Messages

Stocke les messages privés échangés entre les utilisateurs.

| Colonne          | Type          | Description                                      |
|------------------|---------------|--------------------------------------------------|
| id               | uint (PK)     | Identifiant unique                               |
| sender_id        | uint (FK)     | Référence à l'expéditeur                         |
| recipient_id     | uint (FK)     | Référence au destinataire                        |
| content          | text          | Contenu du message                               |
| is_read          | bool          | Indique si le message a été lu                   |
| read_at          | timestamp     | Date de lecture                                  |
| created_at       | timestamp     | Date et heure de création                        |
| updated_at       | timestamp     | Date et heure de dernière modification           |

### Notifications

Stocke les notifications envoyées aux utilisateurs.

| Colonne          | Type          | Description                                      |
|------------------|---------------|--------------------------------------------------|
| id               | uint (PK)     | Identifiant unique                               |
| user_id          | uint (FK)     | Référence à l'utilisateur                        |
| type             | string        | Type de notification                             |
| message          | string        | Message de la notification                       |
| is_read          | bool          | Indique si la notification a été lue             |
| read_at          | timestamp     | Date de lecture                                  |
| related_id       | uint          | ID de l'entité liée                              |
| created_at       | timestamp     | Date et heure de création                        |

## Tables de Modération et Signalement

### Reports

Stocke les signalements de contenu inapproprié.

| Colonne          | Type          | Description                                      |
|------------------|---------------|--------------------------------------------------|
| id               | uint (PK)     | Identifiant unique                               |
| content_id       | uint (FK)     | Référence au contenu signalé                     |
| reporter_id      | uint (FK)     | Référence à l'utilisateur qui signale            |
| reason           | text          | Raison du signalement                            |
| status           | string        | Statut: pending, reviewed, dismissed             |
| reviewed_by      | uint (FK)     | Référence à l'admin qui a traité le signalement  |
| reviewed_at      | timestamp     | Date de traitement du signalement                |
| created_at       | timestamp     | Date et heure de création                        |
| updated_at       | timestamp     | Date et heure de dernière modification           |

## Tables Financières

### Transactions

Enregistre toutes les transactions financières sur la plateforme.

| Colonne          | Type          | Description                                      |
|------------------|---------------|--------------------------------------------------|
| id               | uint (PK)     | Identifiant unique                               |
| user_id          | uint (FK)     | Référence à l'utilisateur                        |
| subscription_id  | uint (FK)     | Référence à l'abonnement (si applicable)         |
| amount           | float         | Montant de la transaction                        |
| currency         | string        | Devise                                           |
| status           | string        | Statut: success, pending, failed                 |
| payment_method   | string        | Méthode de paiement                              |
| payment_id       | string        | ID externe du système de paiement                |
| description      | string        | Description de la transaction                    |
| created_at       | timestamp     | Date et heure de création                        |
| updated_at       | timestamp     | Date et heure de dernière modification           |

### Payouts

Enregistre les versements aux créateurs.

| Colonne          | Type          | Description                                      |
|------------------|---------------|--------------------------------------------------|
| id               | uint (PK)     | Identifiant unique                               |
| creator_id       | uint (FK)     | Référence au créateur                            |
| amount           | float         | Montant du versement                             |
| currency         | string        | Devise                                           |
| status           | string        | Statut: pending, processed, failed               |
| payment_method   | string        | Méthode de paiement                              |
| reference        | string        | Référence du versement                           |
| processed_at     | timestamp     | Date de traitement                               |
| created_at       | timestamp     | Date et heure de création                        |
| updated_at       | timestamp     | Date et heure de dernière modification           |

## Diagramme des Relations

```
┌─────────┐       ┌─────────────────┐       ┌─────────┐
│  Users  │───1:N─┤CreatorProfiles  │       │ Reports │
└─────────┘       └─────────────────┘       └────┬────┘
     │                                           │
     │                                           │
     │1:N         ┌─────────┐                    │
     └────────────┤Contents │←───────────────────┘
     │            └────┬────┘
     │                 │
     │                 │
     │      ┌─────┐    │     ┌─────────┐
     │      │Likes│    │     │Comments │
     │      └──┬──┘    │     └───┬─────┘
     │         │       │         │
     └─────────┴───────┴─────────┘
     │
     │            ┌─────────────────┐      ┌─────────────┐
     │            │SubscriptionPlans│──1:N─┤Subscriptions│
     └────────────┤                 │      │             │
     │            └─────────────────┘      └─────┬───────┘
     │                                           │
     │                                           │
     │                      ┌─────────────┐      │
     └──────────────────────┤Transactions │←─────┘
     │                      └─────────────┘
     │
     │                      ┌─────────┐
     └─────────────────────→│ Payouts │
     │                      └─────────┘
     │
     │                      ┌─────────────┐
     │                      │Notifications│
     └─────────────────────→│            │
     │                      └─────────────┘
     │
     │        ┌─────────┐
     └────────┤Messages │
              └─────────┘
```

## Clés et Contraintes

- **Clés primaires** : Chaque table dispose d'un ID auto-incrémenté comme clé primaire.
- **Clés étrangères** : Les relations entre tables sont maintenues par des clés étrangères.
- **Contraintes d'unicité** : 
  - `username` et `email` dans la table Users
  - `user_id` dans la table CreatorProfiles
- **Index** :
  - Index sur les champs couramment utilisés pour les requêtes comme `creator_id`, `subscriber_id`, etc.
  - Index sur `deleted_at` pour les tables avec soft delete

## Considérations Techniques

- **Soft Delete** : Certaines tables (Contents, Comments) utilisent le soft delete pour préserver l'historique.
- **Timestamps automatiques** : Toutes les tables incluent `created_at` et `updated_at` gérés automatiquement.
- **Types de données** :
  - Les URLs sont stockées sous forme de chaînes de caractères
  - Les montants financiers sont stockés comme des nombres à virgule flottante
  - Les énumérations comme les rôles utilisent des chaînes de caractères pour la lisibilité

