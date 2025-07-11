# GitHub Secrets Configuration pour OnlyFlick Monitoring

Ce guide explique comment configurer les secrets GitHub pour déployer le monitoring OnlyFlick de manière sécurisée en production.

## Configuration des Secrets GitHub

### 1. Accès aux Secrets GitHub

1. Allez dans votre repository GitHub
2. Cliquez sur **Settings** > **Secrets and variables** > **Actions**
3. Cliquez sur **New repository secret**

### 2. Secrets à configurer

#### Secrets obligatoires :

```
GRAFANA_ADMIN_PASSWORD
GRAFANA_SECRET_KEY
GRAFANA_DB_PASSWORD
SMTP_PASSWORD
WEBHOOK_TOKEN
```

#### Secrets optionnels :

```
SLACK_WEBHOOK_URL
LETSENCRYPT_EMAIL
DOMAIN
MONITORING_DOMAIN
API_HOST
ADMIN_EMAIL
SMTP_HOST
SMTP_FROM
SMTP_USERNAME
```

### 3. Valeurs recommandées

#### GRAFANA_ADMIN_PASSWORD
```
Mot de passe fort pour l'admin Grafana
Exemple : Un mot de passe généré de 16+ caractères
```

#### GRAFANA_SECRET_KEY
```
Clé secrète pour signer les cookies Grafana
Doit faire au minimum 32 caractères
Exemple : $(openssl rand -hex 32)
```

#### GRAFANA_DB_PASSWORD
```
Mot de passe pour la base de données PostgreSQL de Grafana
Exemple : Un mot de passe généré de 16+ caractères
```

#### SMTP_PASSWORD
```
Mot de passe pour l'envoi d'emails d'alertes
Obtenez-le auprès de votre fournisseur SMTP
```

#### WEBHOOK_TOKEN
```
Token pour sécuriser les webhooks d'alertes
Exemple : $(openssl rand -hex 32)
```

#### SLACK_WEBHOOK_URL (Optionnel)
```
URL du webhook Slack pour les notifications
Format : https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
```

## Workflow GitHub Actions

### 1. Structure du workflow

Créez le fichier `.github/workflows/deploy-monitoring.yml` :

```yaml
name: Deploy Monitoring

on:
  push:
    branches: [ main ]
    paths:
      - 'monitoring/**'
      - 'docker-compose.monitoring.prod.yml'
      - '.github/workflows/deploy-monitoring.yml'
  workflow_dispatch:

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2
    
    - name: Create environment file
      run: |
        cat > .env.monitoring.prod << EOF
        GRAFANA_ADMIN_USER=admin
        GRAFANA_ADMIN_PASSWORD=${{ secrets.GRAFANA_ADMIN_PASSWORD }}
        GRAFANA_SECRET_KEY=${{ secrets.GRAFANA_SECRET_KEY }}
        GRAFANA_DB_NAME=grafana
        GRAFANA_DB_USER=grafana
        GRAFANA_DB_PASSWORD=${{ secrets.GRAFANA_DB_PASSWORD }}
        SMTP_HOST=${{ secrets.SMTP_HOST || 'smtp.gmail.com' }}
        SMTP_PORT=587
        SMTP_FROM=${{ secrets.SMTP_FROM || 'monitoring@example.com' }}
        SMTP_USERNAME=${{ secrets.SMTP_USERNAME || secrets.SMTP_FROM }}
        SMTP_PASSWORD=${{ secrets.SMTP_PASSWORD }}
        ADMIN_EMAIL=${{ secrets.ADMIN_EMAIL || 'admin@example.com' }}
        SLACK_WEBHOOK_URL=${{ secrets.SLACK_WEBHOOK_URL }}
        WEBHOOK_TOKEN=${{ secrets.WEBHOOK_TOKEN }}
        LETSENCRYPT_EMAIL=${{ secrets.LETSENCRYPT_EMAIL || secrets.ADMIN_EMAIL }}
        DOMAIN=${{ secrets.DOMAIN || 'example.com' }}
        MONITORING_DOMAIN=${{ secrets.MONITORING_DOMAIN || 'monitoring.example.com' }}
        API_HOST=${{ secrets.API_HOST || 'api.example.com' }}
        API_PORT=8080
        EOF
    
    - name: Deploy to production
      run: |
        # Copier les fichiers vers le serveur de production
        # Exemple avec rsync ou scp
        ./deploy_monitoring_prod.sh deploy
```

### 2. Workflow pour les tests

Créez le fichier `.github/workflows/test-monitoring.yml` :

```yaml
name: Test Monitoring

on:
  pull_request:
    paths:
      - 'monitoring/**'
      - 'docker-compose.monitoring.yml'
      - 'test_monitoring.sh'
  workflow_dispatch:

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2
    
    - name: Test monitoring stack
      run: |
        # Utiliser des valeurs de test (non-sensibles)
        export GRAFANA_ADMIN_PASSWORD=test123
        export GRAFANA_SECRET_KEY=test_secret_key_32_chars_minimum
        export GRAFANA_DB_PASSWORD=test123
        export SMTP_PASSWORD=test123
        export WEBHOOK_TOKEN=test_webhook_token
        
        ./test_monitoring.sh
```

## Déploiement avec Docker Swarm

### 1. Utilisation des Docker Secrets

```yaml
version: '3.8'

services:
  grafana:
    image: grafana/grafana:latest
    secrets:
      - grafana_admin_password
      - grafana_secret_key
    environment:
      - GF_SECURITY_ADMIN_PASSWORD_FILE=/run/secrets/grafana_admin_password
      - GF_SECURITY_SECRET_KEY_FILE=/run/secrets/grafana_secret_key

secrets:
  grafana_admin_password:
    external: true
  grafana_secret_key:
    external: true
```

### 2. Script de création des secrets

```bash
#!/bin/bash
# create_docker_secrets.sh

# Créer les secrets Docker
echo "$GRAFANA_ADMIN_PASSWORD" | docker secret create grafana_admin_password -
echo "$GRAFANA_SECRET_KEY" | docker secret create grafana_secret_key -
echo "$GRAFANA_DB_PASSWORD" | docker secret create grafana_db_password -
echo "$SMTP_PASSWORD" | docker secret create smtp_password -
echo "$WEBHOOK_TOKEN" | docker secret create webhook_token -
```

## Déploiement avec Kubernetes

### 1. Création des secrets Kubernetes

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: monitoring-secrets
type: Opaque
data:
  grafana-admin-password: # base64 encoded
  grafana-secret-key: # base64 encoded
  grafana-db-password: # base64 encoded
  smtp-password: # base64 encoded
  webhook-token: # base64 encoded
```

### 2. Utilisation dans les pods

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
spec:
  template:
    spec:
      containers:
      - name: grafana
        image: grafana/grafana:latest
        env:
        - name: GF_SECURITY_ADMIN_PASSWORD
          valueFrom:
            secretKeyRef:
              name: monitoring-secrets
              key: grafana-admin-password
```

## Sécurité et Bonnes Pratiques

### 1. Rotation des secrets

- Changez régulièrement les mots de passe
- Utilisez des mots de passe forts et uniques
- Documentez les rotations dans un journal sécurisé

### 2. Audit et monitoring

```bash
# Audit des accès aux secrets
git log --grep="secret" --oneline
```

### 3. Variables d'environnement sensibles

❌ **Ne jamais faire :**
```bash
# Ne jamais commiter de secrets en clair
GRAFANA_ADMIN_PASSWORD=motdepasse123
```

✅ **Bonne pratique :**
```bash
# Utiliser des placeholders
GRAFANA_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD:-CHANGE_ME}
```

### 4. Validation des secrets

```bash
#!/bin/bash
# validate_secrets.sh

required_secrets=(
    "GRAFANA_ADMIN_PASSWORD"
    "GRAFANA_SECRET_KEY"
    "GRAFANA_DB_PASSWORD"
    "SMTP_PASSWORD"
    "WEBHOOK_TOKEN"
)

for secret in "${required_secrets[@]}"; do
    if [ -z "${!secret}" ]; then
        echo "Erreur: Secret $secret non défini"
        exit 1
    fi
done

echo "Tous les secrets requis sont définis"
```

## Commandes utiles

### Générer des secrets sécurisés

```bash
# Générer un mot de passe fort
openssl rand -base64 32

# Générer une clé secrète
openssl rand -hex 32

# Générer un UUID
uuidgen
```

### Vérifier les secrets GitHub

```bash
# Utiliser l'API GitHub (nécessite un token)
curl -H "Authorization: token $GITHUB_TOKEN" \
     https://api.github.com/repos/USERNAME/REPO/actions/secrets
```

## Dépannage

### 1. Secrets non disponibles

```bash
# Vérifier que les secrets sont bien définis
echo "Secrets disponibles:"
env | grep -E "(GRAFANA|SMTP|WEBHOOK)" | sed 's/=.*/=***/'
```

### 2. Erreurs de déploiement

```bash
# Vérifier les logs de déploiement
docker-compose -f docker-compose.monitoring.prod.yml logs

# Vérifier les secrets Docker
docker secret ls
```

### 3. Test des notifications

```bash
# Tester l'envoi d'emails
./test_monitoring.sh --test-email

# Tester les webhooks
curl -X POST "$WEBHOOK_URL" -H "Authorization: Bearer $WEBHOOK_TOKEN" \
     -d '{"test": "notification"}'
```

## Ressources

- [GitHub Secrets Documentation](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [Docker Secrets](https://docs.docker.com/engine/swarm/secrets/)
- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
- [Best Practices for Secrets Management](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)
