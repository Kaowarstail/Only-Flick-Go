# Monitoring OnlyFlick - Guide Complet

## Vue d'ensemble

Le monitoring OnlyFlick utilise une stack complète avec Prometheus, Grafana, et Alertmanager pour surveiller votre backend Go. Cette configuration permet de :

- **Surveiller les performances** de l'API
- **Analyser les métriques business** (utilisateurs, contenus, paiements)
- **Monitorer l'infrastructure** (CPU, mémoire, disque)
- **Recevoir des alertes** en cas de problème
- **Visualiser les données** via des dashboards Grafana

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   OnlyFlick API │───▶│   Prometheus    │───▶│     Grafana     │
│   (Port 8080)   │    │   (Port 9090)   │    │   (Port 3000)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │  Alertmanager   │
                       │   (Port 9093)   │
                       └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │ Node Exporter   │
                       │   (Port 9100)   │
                       └─────────────────┘
```

## Installation et Configuration

### 1. Démarrage rapide

```bash
# Démarrer le monitoring
./start_monitoring.sh start

# Afficher le statut
./start_monitoring.sh status

# Voir les logs
./start_monitoring.sh logs
```

### 2. Accès aux interfaces

- **Grafana** : http://localhost:3000
  - Login: `admin`
  - Mot de passe: `admin123`
- **Prometheus** : http://localhost:9090
- **Alertmanager** : http://localhost:9093

### 3. Dashboards disponibles

1. **OnlyFlick API - Monitoring Principal**
   - Taux de requêtes HTTP
   - Temps de réponse (percentiles)
   - Taux d'erreur
   - Utilisation mémoire
   - Connexions base de données

2. **OnlyFlick - Métriques Business**
   - Utilisateurs actifs
   - Abonnements actifs
   - Créations de contenu
   - Nouvelles inscriptions
   - Revenus

3. **OnlyFlick - Métriques Infrastructure**
   - Utilisation CPU
   - Utilisation mémoire système
   - Utilisation disque
   - Trafic réseau
   - Charge système

## Métriques Collectées

### Métriques HTTP
- `http_requests_total` : Nombre total de requêtes HTTP
- `http_request_duration_seconds` : Durée des requêtes HTTP

### Métriques Utilisateurs
- `active_users_total` : Nombre d'utilisateurs actifs
- `users_registered_total` : Nombre total d'utilisateurs inscrits

### Métriques Contenu
- `content_created_total` : Nombre total de contenus créés
- `content_views_total` : Nombre total de vues de contenu

### Métriques Paiements
- `payment_attempts_total` : Nombre total de tentatives de paiement
- `payment_amount_euros` : Montants des paiements en euros

### Métriques Cloudinary
- `cloudinary_upload_attempts_total` : Nombre total d'uploads
- `cloudinary_upload_success_total` : Nombre d'uploads réussis
- `cloudinary_upload_duration_seconds` : Durée des uploads

### Métriques Base de Données
- `database_connections_active` : Nombre de connexions actives
- `database_query_duration_seconds` : Durée des requêtes

## Alertes Configurées

### Alertes Critiques
- **OnlyFlickAPIDown** : API non disponible (> 1 min)
- **HighErrorRate** : Taux d'erreur > 10% (> 1 min)

### Alertes Warning
- **HighResponseTime** : Temps de réponse P95 > 1s (> 2 min)
- **HighMemoryUsage** : Utilisation mémoire > 80% (> 5 min)
- **HighGoroutineCount** : Nombre de goroutines > 1000 (> 5 min)
- **LowUploadSuccessRate** : Taux de succès upload < 95% (> 2 min)

## Configuration Alertmanager

Les alertes sont configurées pour :
- Envoyer des notifications webhook vers l'API OnlyFlick
- Envoyer des emails (configurez votre SMTP dans `alertmanager.yml`)
- Grouper les alertes par nom
- Répéter les alertes toutes les heures

## Maintenance

### Commandes utiles

```bash
# Redémarrer le monitoring
./start_monitoring.sh restart

# Voir le statut des services
./start_monitoring.sh status

# Nettoyer les volumes (supprime les données)
./start_monitoring.sh cleanup

# Afficher les informations d'accès
./start_monitoring.sh info
```

### Mise à jour des dashboards

Les dashboards sont automatiquement provisionnés depuis le dossier :
```
monitoring/grafana/dashboards/
```

Pour modifier un dashboard :
1. Éditez le fichier JSON correspondant
2. Redémarrez Grafana : `./start_monitoring.sh restart`

### Ajout de nouvelles métriques

1. Ajoutez vos métriques dans `internal/services/metrics.go`
2. Utilisez les métriques dans votre code avec les fonctions `Record*`
3. Mettez à jour les dashboards Grafana si nécessaire

## Dépannage

### Problèmes courants

1. **Prometheus ne peut pas scraper l'API**
   - Vérifiez que l'API OnlyFlick est démarrée sur le port 8080
   - Vérifiez que l'endpoint `/metrics` est accessible

2. **Grafana ne peut pas se connecter à Prometheus**
   - Vérifiez que Prometheus est démarré
   - Vérifiez la configuration de la datasource

3. **Dashboards vides**
   - Vérifiez que les métriques sont bien envoyées par l'API
   - Vérifiez les requêtes PromQL dans Grafana

### Logs de diagnostic

```bash
# Voir les logs de tous les services
./start_monitoring.sh logs

# Voir les logs d'un service spécifique
docker-compose -f docker-compose.monitoring.yml logs prometheus
docker-compose -f docker-compose.monitoring.yml logs grafana
```

## Optimisations

### Performance

- **Rétention des données** : Configurée à 200h dans Prometheus
- **Intervalle de scraping** : 15s par défaut
- **Buckets des histogrammes** : Optimisés pour les temps de réponse web

### Sécurité

- **Accès restreint** : Grafana avec authentification
- **Réseau isolé** : Services dans un réseau Docker dédié
- **Volumes persistants** : Données sauvegardées entre les redémarrages

## Intégration avec CI/CD

Pour intégrer le monitoring dans votre pipeline :

```yaml
# Exemple pour GitHub Actions
- name: Start Monitoring
  run: ./start_monitoring.sh start

- name: Wait for services
  run: sleep 30

- name: Check monitoring health
  run: |
    curl -f http://localhost:9090/-/healthy
    curl -f http://localhost:3000/api/health
```

## Métriques Personnalisées

Exemple d'ajout d'une nouvelle métrique :

```go
// Dans internal/services/metrics.go
var CustomMetric = promauto.NewCounter(
    prometheus.CounterOpts{
        Name: "custom_metric_total",
        Help: "Description de votre métrique",
    },
)

// Dans votre code
func MyFunction() {
    // Votre logique
    services.CustomMetric.Inc()
}
```

## Support

Pour toute question ou problème :
1. Consultez les logs des services
2. Vérifiez la configuration Prometheus
3. Testez les requêtes PromQL dans Prometheus
4. Consultez la documentation Grafana

---

**Note** : Ce monitoring est conçu pour un environnement de développement/test. Pour la production, considérez l'ajout de :
- Authentification renforcée
- Sauvegarde des données
- Monitoring externe des services
- Alertes par SMS/Slack
