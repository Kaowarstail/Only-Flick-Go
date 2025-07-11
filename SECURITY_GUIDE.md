# Guide de Sécurité - OnlyFlick Monitoring

Ce guide présente les bonnes pratiques de sécurité pour déployer le monitoring OnlyFlick en production.

## 🔒 Principes de Sécurité

### 1. Zéro Secret en Clair
- **Jamais** de secrets dans le code source
- **Jamais** de secrets dans les fichiers de configuration committés
- **Toujours** utiliser des gestionnaires de secrets dédiés

### 2. Rotation des Secrets
- Changer régulièrement les mots de passe
- Utiliser des secrets forts et uniques
- Documenter les rotations

### 3. Accès Minimal
- Principe du moindre privilège
- Séparer les environnements
- Auditer les accès

## 🛡️ Configuration Sécurisée

### GitHub Secrets (Recommandé)

#### Configuration des Secrets
1. **Accès aux paramètres**
   ```bash
   Repository > Settings > Secrets and variables > Actions
   ```

2. **Secrets obligatoires**
   ```bash
   GRAFANA_ADMIN_PASSWORD     # Mot de passe fort (16+ caractères)
   GRAFANA_SECRET_KEY         # Clé de 32+ caractères
   GRAFANA_DB_PASSWORD        # Mot de passe base de données
   SMTP_PASSWORD              # Mot de passe SMTP
   WEBHOOK_TOKEN              # Token pour webhooks
   ```

3. **Secrets optionnels**
   ```bash
   SLACK_WEBHOOK_URL          # URL webhook Slack
   DOMAIN                     # Domaine principal
   MONITORING_DOMAIN          # Sous-domaine monitoring
   ADMIN_EMAIL                # Email admin
   LETSENCRYPT_EMAIL          # Email Let's Encrypt
   ```

#### Validation des Secrets
```bash
# Valider la configuration GitHub Secrets
./validate_secrets.sh --github-actions

# Générer des secrets manquants (développement uniquement)
./validate_secrets.sh --generate-missing
```

### Docker Secrets

#### Configuration Docker Swarm
```bash
# Créer les secrets Docker
echo "mon_mot_de_passe_fort" | docker secret create grafana_admin_password -
echo "ma_cle_secrete_32_chars_minimum" | docker secret create grafana_secret_key -
echo "mot_de_passe_db" | docker secret create grafana_db_password -
echo "mot_de_passe_smtp" | docker secret create smtp_password -
echo "token_webhook" | docker secret create webhook_token -
```

#### Utilisation dans Docker Compose
```yaml
services:
  grafana:
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

### Kubernetes Secrets

#### Création des Secrets
```bash
# Créer un secret Kubernetes
kubectl create secret generic monitoring-secrets \
  --from-literal=grafana-admin-password="mot_de_passe_fort" \
  --from-literal=grafana-secret-key="cle_secrete_32_chars" \
  --from-literal=grafana-db-password="mot_de_passe_db" \
  --from-literal=smtp-password="mot_de_passe_smtp" \
  --from-literal=webhook-token="token_webhook"
```

#### Utilisation dans les Pods
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
        env:
        - name: GF_SECURITY_ADMIN_PASSWORD
          valueFrom:
            secretKeyRef:
              name: monitoring-secrets
              key: grafana-admin-password
```

## 🔧 Outils de Sécurité

### Génération de Secrets Forts

#### Mots de Passe
```bash
# Générer un mot de passe fort
openssl rand -base64 32

# Générer un mot de passe avec caractères spéciaux
openssl rand -base64 32 | tr -d "=+/" | cut -c1-25

# Utiliser pwgen si disponible
pwgen -s 32 1
```

#### Clés Secrètes
```bash
# Générer une clé hexadécimale
openssl rand -hex 32

# Générer une clé base64
openssl rand -base64 32

# Générer un UUID
uuidgen
```

### Validation et Audit

#### Script de Validation
```bash
# Valider tous les secrets
./validate_secrets.sh --verbose

# Valider avec génération des manquants
./validate_secrets.sh --generate-missing

# Valider les secrets Docker
./validate_secrets.sh --docker-secrets

# Valider les secrets Kubernetes
./validate_secrets.sh --kubernetes
```

#### Audit des Accès
```bash
# Vérifier les logs de déploiement
docker-compose logs grafana | grep -i "admin\|login\|auth"

# Vérifier les connexions Grafana
curl -s http://localhost:3000/api/admin/stats

# Auditer les secrets Git
git log --grep="secret\|password\|token" --oneline
```

## 🚨 Gestion des Incidents

### Compromission de Secrets

#### Actions Immédiates
1. **Révoquer immédiatement** les secrets compromis
2. **Régénérer** de nouveaux secrets
3. **Redéployer** avec les nouveaux secrets
4. **Auditer** les accès récents

#### Procédure de Rotation
```bash
# 1. Générer de nouveaux secrets
./validate_secrets.sh --generate-missing

# 2. Mettre à jour GitHub Secrets
# (via l'interface web ou API)

# 3. Redéployer
./deploy_monitoring_prod.sh update

# 4. Vérifier le fonctionnement
./test_monitoring.sh --quick-test
```

### Détection d'Intrusion

#### Monitoring des Connexions
```bash
# Surveiller les connexions Grafana
tail -f monitoring/grafana/data/log/grafana.log | grep -i "login\|auth"

# Surveiller les métriques Prometheus
curl -s "http://localhost:9090/api/v1/query?query=prometheus_notifications_total"
```

#### Alertes de Sécurité
```yaml
# Règle d'alerte pour tentatives de connexion suspectes
groups:
- name: security
  rules:
  - alert: SuspiciousLoginAttempts
    expr: increase(grafana_api_response_status_total{code="401"}[5m]) > 5
    for: 1m
    labels:
      severity: warning
    annotations:
      summary: "Tentatives de connexion suspectes détectées"
```

## 🔐 Bonnes Pratiques

### Développement Local

#### Environnement Isolé
```bash
# Utiliser des secrets de développement
export GRAFANA_ADMIN_PASSWORD="dev_password_123"
export GRAFANA_SECRET_KEY="dev_secret_key_32_chars_minimum"

# Démarrer en mode développement
./start_monitoring.sh
```

#### Tests de Sécurité
```bash
# Tester la configuration
./test_monitoring.sh --security-test

# Valider les secrets
./validate_secrets.sh --verbose

# Vérifier les ports exposés
nmap -sS -p 3000,9090,9093 localhost
```

### Production

#### Déploiement Sécurisé
```bash
# Déployer via GitHub Actions (recommandé)
git push origin main

# Ou déployer manuellement après validation
./validate_secrets.sh --github-actions
./deploy_monitoring_prod.sh deploy
```

#### Monitoring de Sécurité
```bash
# Surveiller les logs en temps réel
docker-compose -f docker-compose.monitoring.prod.yml logs -f | grep -i "error\|warning\|auth"

# Vérifier l'état des services
./test_monitoring.sh --health-check
```

## 📋 Checklist de Sécurité

### Avant le Déploiement
- [ ] Tous les secrets sont configurés dans GitHub Secrets
- [ ] Aucun secret en clair dans le code
- [ ] Validation des secrets réussie
- [ ] Tests de sécurité passés
- [ ] Domaines et certificats SSL configurés

### Après le Déploiement
- [ ] Services accessibles uniquement via HTTPS
- [ ] Mots de passe par défaut changés
- [ ] Alertes de sécurité configurées
- [ ] Logs de sécurité activés
- [ ] Sauvegarde des configurations

### Maintenance Régulière
- [ ] Rotation des secrets (mensuelle)
- [ ] Mise à jour des images Docker
- [ ] Audit des accès
- [ ] Test de restauration
- [ ] Vérification des alertes

## 🆘 Contacts d'Urgence

### Procédure d'Incident
1. **Identifier** le type d'incident
2. **Isoler** les systèmes affectés
3. **Documenter** les actions prises
4. **Communiquer** avec l'équipe
5. **Restaurer** les services
6. **Analyser** les causes

### Ressources Utiles
- [Guide des Secrets GitHub](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [Sécurité Docker](https://docs.docker.com/engine/security/)
- [Sécurité Kubernetes](https://kubernetes.io/docs/concepts/security/)
- [OWASP Secrets Management](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)

## 🔗 Fichiers Associés

- `GITHUB_SECRETS_GUIDE.md` - Guide détaillé GitHub Secrets
- `validate_secrets.sh` - Script de validation
- `.github/workflows/deploy-monitoring.yml` - Pipeline de déploiement
- `.github/workflows/test-monitoring.yml` - Tests automatisés
- `deploy_monitoring_prod.sh` - Script de déploiement
- `test_monitoring.sh` - Script de test
