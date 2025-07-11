#!/bin/bash

# Script pour exécuter les tests uniquement sur les packages qui en contiennent
# Ceci évite les messages "no test files" et "FAIL" sur les packages sans tests

set -e

echo "🔍 Recherche des packages avec des fichiers de test..."

# Trouver tous les répertoires contenant des fichiers *_test.go
TEST_DIRS=$(find . -name "*_test.go" -type f | xargs -I {} dirname {} | sort -u)

if [ -z "$TEST_DIRS" ]; then
    echo "❌ Aucun fichier de test trouvé!"
    exit 1
fi

echo "✅ Fichiers de test trouvés dans les répertoires suivants:"
for dir in $TEST_DIRS; do
    echo "  - $dir"
done

echo ""
echo "🧪 Exécution des tests..."

# Convertir les chemins de répertoires en packages Go
TEST_PACKAGES=""
for dir in $TEST_DIRS; do
    # Convertir ./internal/handlers en ./internal/handlers/...
    if [ "$dir" = "." ]; then
        TEST_PACKAGES="$TEST_PACKAGES ."
    else
        # Enlever le ./ du début et ajouter /...
        PKG=${dir#./}
        TEST_PACKAGES="$TEST_PACKAGES ./$PKG"
    fi
done

echo "📦 Packages à tester: $TEST_PACKAGES"
echo ""

# Exécuter les tests uniquement sur les packages qui en contiennent
go test -v -race -buildvcs $TEST_PACKAGES

echo ""
echo "✅ Tous les tests ont été exécutés avec succès!"
