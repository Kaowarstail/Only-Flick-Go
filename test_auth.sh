#!/bin/bash

# Script pour exécuter les tests d'authentification
# Usage: ./test_auth.sh [option]

echo "🧪 Tests d'Authentification - OnlyFlick API"
echo "=========================================="

case "$1" in
    "all")
        echo "🔍 Exécution de TOUS les tests d'authentification..."
        go test -v -run "Test.*Auth|TestLogin|TestRegister|TestLogout|TestRefresh|TestGenerate|TestGetCurrentUser|TestRequestPasswordReset|TestFullAuthFlow" ./internal/handlers/ -count=1
        ;;
    "core")
        echo "🔍 Exécution des tests PRINCIPAUX d'authentification..."
        go test -v -run "TestLogin_Success|TestRegister_Success|TestLogout_Success|TestGenerateJWT_Success|TestFullAuthFlow" ./internal/handlers/ -count=1
        ;;
    "security")
        echo "🛡️ Exécution des tests de SÉCURITÉ..."
        go test -v -run "TestLogin_SQLInjectionPrevention|TestRegister_XSSPrevention|TestLogin_BannedUser" ./internal/handlers/ -count=1
        ;;
    "login")
        echo "🔐 Exécution des tests de CONNEXION..."
        go test -v -run "TestLogin_.*" ./internal/handlers/ -count=1
        ;;
    "register")
        echo "📝 Exécution des tests d'INSCRIPTION..."
        go test -v -run "TestRegister_.*" ./internal/handlers/ -count=1
        ;;
    "jwt")
        echo "🎫 Exécution des tests de JWT..."
        go test -v -run "TestGenerateJWT_.*|TestRefreshToken_.*" ./internal/handlers/ -count=1
        ;;
    "flow")
        echo "🔗 Exécution du test de flux COMPLET..."
        go test -v -run "TestFullAuthFlow" ./internal/handlers/ -count=1
        ;;
    "coverage")
        echo "📊 Exécution avec COUVERTURE DE CODE..."
        go test -v -cover -run "Test.*Auth|TestLogin|TestRegister|TestLogout|TestRefresh|TestGenerate" ./internal/handlers/ -count=1
        ;;
    *)
        echo "❓ Usage: $0 [option]"
        echo ""
        echo "Options disponibles:"
        echo "  all      - Tous les tests d'authentification"
        echo "  core     - Tests principaux (connexion, inscription, déconnexion)"
        echo "  security - Tests de sécurité (SQL injection, XSS, utilisateurs bannis)"
        echo "  login    - Tests de connexion uniquement"
        echo "  register - Tests d'inscription uniquement"
        echo "  jwt      - Tests de gestion des tokens JWT"
        echo "  flow     - Test du flux d'authentification complet"
        echo "  coverage - Tous les tests avec couverture de code"
        echo ""
        echo "Exemples:"
        echo "  $0 core      # Tests principaux"
        echo "  $0 security  # Tests de sécurité"
        echo "  $0 coverage  # Avec couverture"
        ;;
esac
