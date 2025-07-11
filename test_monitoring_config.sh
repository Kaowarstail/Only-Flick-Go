#!/bin/bash

# Script de test simple pour les configurations de monitoring
# Ne nécessite pas Docker (pour les tests rapides)

set -e

echo "🔍 Test des configurations de monitoring (sans Docker)..."

# Test des fichiers de configuration Prometheus
echo "📊 Test des fichiers Prometheus..."
if [ -f "monitoring/prometheus/prometheus.yml" ]; then
    echo "✅ prometheus.yml trouvé"
    
    # Vérifier la syntaxe YAML basique
    if command -v yamllint >/dev/null 2>&1; then
        yamllint -d relaxed monitoring/prometheus/prometheus.yml && echo "✅ YAML valide"
    else
        echo "⚠️ yamllint non installé, vérification YAML basique seulement"
        if python3 -c "import yaml; yaml.safe_load(open('monitoring/prometheus/prometheus.yml'))" 2>/dev/null; then
            echo "✅ YAML valide"
        else
            echo "❌ YAML invalide"
            exit 1
        fi
    fi
else
    echo "❌ prometheus.yml manquant"
    exit 1
fi

# Test des fichiers de configuration Alertmanager
echo "🚨 Test des fichiers Alertmanager..."
if [ -f "monitoring/alertmanager/alertmanager.yml" ]; then
    echo "✅ alertmanager.yml trouvé"
    
    # Vérifier la syntaxe YAML basique
    if command -v yamllint >/dev/null 2>&1; then
        yamllint -d relaxed monitoring/alertmanager/alertmanager.yml && echo "✅ YAML valide"
    else
        if python3 -c "import yaml; yaml.safe_load(open('monitoring/alertmanager/alertmanager.yml'))" 2>/dev/null; then
            echo "✅ YAML valide"
        else
            echo "❌ YAML invalide"
            exit 1
        fi
    fi
else
    echo "❌ alertmanager.yml manquant"
    exit 1
fi

# Test des dashboards Grafana
echo "📈 Test des dashboards Grafana..."
dashboard_count=0
for dashboard in monitoring/grafana/dashboards/*.json; do
    if [ -f "$dashboard" ]; then
        echo "Validation de $dashboard..."
        
        # Vérifier que c'est un JSON valide
        if command -v jq >/dev/null 2>&1; then
            if jq . "$dashboard" > /dev/null 2>&1; then
                echo "✅ Dashboard valide: $dashboard"
                dashboard_count=$((dashboard_count + 1))
            else
                echo "❌ Dashboard JSON invalide: $dashboard"
                exit 1
            fi
        elif python3 -c "import json; json.load(open('$dashboard'))" 2>/dev/null; then
            echo "✅ Dashboard valide: $dashboard"
            dashboard_count=$((dashboard_count + 1))
        else
            echo "❌ Dashboard JSON invalide: $dashboard"
            exit 1
        fi
    fi
done

if [ $dashboard_count -eq 0 ]; then
    echo "⚠️ Aucun dashboard trouvé"
else
    echo "✅ $dashboard_count dashboards validés"
fi

# Test des fichiers Docker Compose
echo "🐳 Test des fichiers Docker Compose..."
if [ -f "docker-compose.monitoring.yml" ]; then
    echo "✅ docker-compose.monitoring.yml trouvé"
else
    echo "❌ docker-compose.monitoring.yml manquant"
    exit 1
fi

if [ -f "docker-compose.monitoring.prod.yml" ]; then
    echo "✅ docker-compose.monitoring.prod.yml trouvé"
else
    echo "❌ docker-compose.monitoring.prod.yml manquant"
    exit 1
fi

echo "🎉 Tous les tests de configuration ont réussi !"
echo "💡 Pour tester avec Docker, utilisez: ./test_docker_compose_syntax.sh"
