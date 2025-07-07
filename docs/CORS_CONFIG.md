# Configuration CORS

La configuration CORS est maintenant gérée via les variables d'environnement dans le fichier `.env`.

## Variables disponibles

### CORS_ALLOWED_ORIGINS
Liste des origines autorisées, séparées par des virgules.

**Exemple :**
```
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:49219,https://monapp.com
```

**Valeur par défaut :**
```
http://localhost:3000,http://localhost:49219,http://127.0.0.1:3000,http://127.0.0.1:49219
```

### CORS_ALLOW_CREDENTIALS
Autoriser l'envoi de cookies et d'en-têtes d'authentification.

**Valeurs possibles :** `true` ou `false`
**Valeur par défaut :** `true`

### CORS_MAX_AGE
Durée en secondes pour le cache des requêtes preflight.

**Valeur par défaut :** `86400` (24 heures)

## Comportement en développement

En mode développement, le middleware accepte automatiquement toutes les origines qui commencent par :
- `http://localhost`
- `http://127.0.0`

Cela permet de travailler avec Flutter Web qui utilise des ports dynamiques.

## Exemple de configuration

```env
# Pour le développement local
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:49219
CORS_ALLOW_CREDENTIALS=true
CORS_MAX_AGE=86400

# Pour la production
CORS_ALLOWED_ORIGINS=https://monapp.com,https://www.monapp.com
CORS_ALLOW_CREDENTIALS=true
CORS_MAX_AGE=3600
```
