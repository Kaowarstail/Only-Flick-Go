#!/bin/bash

# Script de test pour vérifier la syntaxe Docker Compose
# Ce script reproduit les tests du workflow GitHub Actions

set -e

echo "🔍 Test des fichiers Docker Compose..."

# Vérifier que les fichiers existent
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

# Vérifier la syntaxe avec la nouvelle syntaxe Docker Compose
echo "🔧 Validation de la syntaxe Docker Compose..."
docker compose -f docker-compose.monitoring.yml config > /dev/null
echo "✅ Configuration Docker Compose dev valide"

docker compose -f docker-compose.monitoring.prod.yml config > /dev/null
echo "✅ Configuration Docker Compose prod valide"

echo "🎉 Tous les tests ont réussi !"
