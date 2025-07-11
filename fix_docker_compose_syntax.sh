#!/bin/bash

# Script pour corriger docker-compose vers docker compose dans tous les fichiers

set -e

echo "🔧 Correction de docker-compose vers docker compose..."

# Fichiers à corriger
files_to_fix=(
    "deploy_monitoring_prod.sh"
    "deploy_monitoring_secure.sh"
    "start_monitoring.sh"
    "test_monitoring.sh"
    ".github/workflows/test-monitoring.yml"
    ".github/workflows/deploy-monitoring.yml"
    "GITHUB_ACTIONS_TROUBLESHOOTING.md"
    "DOCKER_COMPOSE_PROD_GUIDE.md"
    "MONITORING.md"
    "MONITORING_QUICK_START.md"
)

# Fonction pour corriger un fichier
fix_file() {
    local file="$1"
    if [ -f "$file" ]; then
        echo "Correction de $file..."
        
        # Sauvegarder le fichier original
        cp "$file" "$file.bak"
        
        # Remplacer docker-compose par docker compose
        # Mais garder les noms de fichiers comme docker-compose.yml
        sed -i 's/docker-compose -f/docker compose -f/g' "$file"
        sed -i 's/docker-compose --/docker compose --/g' "$file"
        sed -i 's/docker-compose up/docker compose up/g' "$file"
        sed -i 's/docker-compose down/docker compose down/g' "$file"
        sed -i 's/docker-compose ps/docker compose ps/g' "$file"
        sed -i 's/docker-compose logs/docker compose logs/g' "$file"
        sed -i 's/docker-compose pull/docker compose pull/g' "$file"
        sed -i 's/docker-compose restart/docker compose restart/g' "$file"
        sed -i 's/docker-compose config/docker compose config/g' "$file"
        sed -i 's/docker-compose version/docker compose version/g' "$file"
        sed -i 's/docker-compose exec/docker compose exec/g' "$file"
        sed -i 's/command -v docker-compose/docker compose version/g' "$file"
        
        # Vérifications spécifiques
        sed -i 's/if ! command -v docker compose/if ! docker compose version/g' "$file"
        sed -i 's/Docker Compose n'\''est pas installé/Docker Compose n'\''est pas disponible/g' "$file"
        
        echo "✅ $file corrigé"
    else
        echo "⚠️ $file non trouvé"
    fi
}

# Corriger tous les fichiers
for file in "${files_to_fix[@]}"; do
    fix_file "$file"
done

echo "✅ Correction terminée"
echo ""
echo "📋 Résumé des changements:"
echo "- docker-compose -f → docker compose -f"
echo "- docker-compose up → docker compose up"
echo "- docker-compose down → docker compose down"
echo "- Etc."
echo ""
echo "🗂️ Fichiers de sauvegarde créés avec extension .bak"
echo "💡 Pour restaurer: mv fichier.bak fichier"
