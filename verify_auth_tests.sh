#!/bin/bash

echo "🎯 Vérification Finale - Tests d'Authentification OnlyFlick"
echo "==========================================================="

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo ""
echo -e "${BLUE}📋 Exécution des tests d'authentification...${NC}"
echo ""

# Exécuter les tests et capturer le résultat
cd "$(dirname "$0")"
TEST_OUTPUT=$(go test ./internal/handlers/ -v 2>&1)
TEST_RESULT=$?

echo "$TEST_OUTPUT"

echo ""
echo "==========================================================="

if [ $TEST_RESULT -eq 0 ]; then
    echo -e "${GREEN}✅ SUCCÈS : Tous les tests d'authentification passent !${NC}"
    
    # Compter le nombre de tests qui passent
    PASSED_TESTS=$(echo "$TEST_OUTPUT" | grep "PASS:" | wc -l | tr -d ' ')
    echo -e "${GREEN}📊 Nombre de tests passants : $PASSED_TESTS${NC}"
    
    echo ""
    echo -e "${BLUE}🎉 Résumé des fonctionnalités testées :${NC}"
    echo "   • Connexion (avec email/username)"
    echo "   • Inscription avec validation"
    echo "   • Déconnexion sécurisée"
    echo "   • Rafraîchissement de tokens"
    echo "   • Informations utilisateur"
    echo "   • Réinitialisation mot de passe"
    echo "   • Génération JWT"
    echo "   • Sécurité (SQL injection, XSS)"
    echo "   • Flux d'authentification complet"
    
    echo ""
    echo -e "${BLUE}📚 Documentation disponible :${NC}"
    echo "   • TESTS_AUTHENTIFICATION_FINAL.md"
    echo "   • test_auth.sh (script de test avec options)"
    
    echo ""
    echo -e "${GREEN}🚀 Prêt pour la production / CI/CD !${NC}"
    
else
    echo -e "${RED}❌ ÉCHEC : Certains tests d'authentification échouent${NC}"
    echo ""
    echo -e "${YELLOW}🔧 Actions recommandées :${NC}"
    echo "   1. Vérifier les logs ci-dessus"
    echo "   2. Corriger les tests défaillants"
    echo "   3. Relancer ce script"
    echo ""
    exit 1
fi

echo ""
echo -e "${BLUE}💡 Commandes utiles :${NC}"
echo "   # Tous les tests"
echo "   ./test_auth.sh all"
echo ""
echo "   # Tests principaux seulement"
echo "   ./test_auth.sh core"
echo ""
echo "   # Avec couverture de code"
echo "   ./test_auth.sh coverage"
echo ""
echo "   # Test spécifique"
echo "   go test ./internal/handlers/ -run TestLogin_Success -v"

echo ""
echo "==========================================================="
echo -e "${GREEN}🎯 Vérification terminée avec succès !${NC}"
