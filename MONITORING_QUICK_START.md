# 🎯 Monitoring OnlyFlick - Configuration Complète

## 🚀 Démarrage Rapide

```bash
# 1. Démarrer le monitoring
./start_monitoring.sh start

# 2. Démarrer votre API OnlyFlick
go run cmd/api/main.go

# 3. Tester le monitoring
./test_monitoring.sh
```

## 📊 Dashboards Disponibles

### 1. Vue d'Ensemble
- **URL**: http://localhost:3000/d/onlyflick-overview
- **Contenu**: Indicateurs clés, statut de l'API, métriques business principales

### 2. Monitoring API
- **URL**: http://localhost:3000/d/onlyflick-main
- **Contenu**: Performances HTTP, temps de réponse, erreurs, mémoire

### 3. Métriques Business
- **URL**: http://localhost:3000/d/onlyflick-business
- **Contenu**: Utilisateurs, contenus, abonnements, paiements

### 4. Infrastructure
- **URL**: http://localhost:3000/d/onlyflick-infrastructure
- **Contenu**: CPU, mémoire, disque, réseau

## 🔧 Configuration

### Services Déployés
- **Prometheus** (9090) : Collecte des métriques
- **Grafana** (3000) : Visualisation des données
- **Alertmanager** (9093) : Gestion des alertes
- **Node Exporter** (9100) : Métriques système

### Métriques Collectées
- **HTTP** : Requêtes, temps de réponse, erreurs
- **Business** : Utilisateurs, contenus, abonnements
- **Infrastructure** : CPU, mémoire, disque
- **Application** : Goroutines, mémoire Go
- **Base de données** : Connexions, temps de requête

## 📈 Fonctionnalités Avancées

### Mise à jour Automatique
- Les métriques sont mises à jour toutes les 30 secondes
- Synchronisation avec la base de données
- Calcul automatique des utilisateurs actifs

### Alertes Configurées
- ✅ API Down (critique)
- ⚠️ Temps de réponse élevé
- ⚠️ Taux d'erreur élevé
- ⚠️ Utilisation mémoire élevée

## 🛠️ Commandes Utiles

```bash
# Gestion du monitoring
./start_monitoring.sh start     # Démarrer
./start_monitoring.sh stop      # Arrêter
./start_monitoring.sh restart   # Redémarrer
./start_monitoring.sh status    # Statut
./start_monitoring.sh logs      # Logs

# Tests
./test_monitoring.sh test       # Test complet
./test_monitoring.sh traffic    # Générer du trafic
./test_monitoring.sh load       # Test de charge
```

## 🔐 Accès

### Grafana
- **URL**: http://localhost:3000
- **Login**: admin
- **Password**: admin123

### Prometheus
- **URL**: http://localhost:9090

### Alertmanager
- **URL**: http://localhost:9093

## 📋 Checklist de Démarrage

- [ ] Docker et Docker Compose installés
- [ ] Ports 3000, 9090, 9093, 9100 disponibles
- [ ] API OnlyFlick démarrée sur le port 8080
- [ ] Monitoring lancé avec `./start_monitoring.sh start`
- [ ] Dashboards accessibles dans Grafana
- [ ] Tests passés avec `./test_monitoring.sh`

## 🐛 Dépannage

### Problème : Grafana ne démarre pas
```bash
# Vérifier les logs
./start_monitoring.sh logs

# Vérifier les ports
netstat -tlnp | grep :3000
```

### Problème : Pas de métriques
```bash
# Vérifier que l'API est accessible
curl http://localhost:8080/metrics

# Vérifier la configuration Prometheus
curl http://localhost:9090/targets
```

### Problème : Dashboards vides
1. Vérifier que Prometheus collecte les données
2. Vérifier les requêtes dans Grafana
3. Générer du trafic avec `./test_monitoring.sh traffic`

## 🔧 Personnalisation

### Ajouter une métrique
```go
// Dans internal/services/metrics.go
var MyMetric = promauto.NewCounter(
    prometheus.CounterOpts{
        Name: "my_metric_total",
        Help: "Description de ma métrique",
    },
)

// Dans votre code
services.MyMetric.Inc()
```

### Modifier un dashboard
1. Éditer le fichier JSON dans `monitoring/grafana/dashboards/`
2. Redémarrer Grafana : `./start_monitoring.sh restart`

## 📚 Documentation

- [Guide complet](MONITORING.md)
- [Configuration Prometheus](monitoring/prometheus/prometheus.yml)
- [Alertes](monitoring/prometheus/alert.rules.yml)
- [Dashboards](monitoring/grafana/dashboards/)

## 🎉 Félicitations !

Vous avez maintenant un monitoring complet pour votre API OnlyFlick avec :
- ✅ Métriques en temps réel
- ✅ Dashboards visuels
- ✅ Alertes automatiques
- ✅ Tests automatisés
- ✅ Documentation complète

Votre stack de monitoring est prête pour la production ! 🚀
