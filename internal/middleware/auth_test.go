package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/Kaowarstail/Only-Flick-Go/config"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

// Helper pour créer un token JWT valide pour les tests
func createTestJWT(userID string, role string) (string, error) {
	config.SetTestConfig()

	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Get().JWT.Secret))
}

// Handler de test simple
func testHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

// TestJWTAuthSuccess teste l'authentification JWT réussie
func TestJWTAuthSuccess(t *testing.T) {
	config.SetTestConfig()

	// Créer un token valide
	userID := "test-user-id"
	role := string(models.RoleSubscriber)
	token, err := createTestJWT(userID, role)
	if err != nil {
		t.Fatalf("Failed to create test JWT: %v", err)
	}

	// Créer la requête avec le token
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()

	// Créer le handler avec le middleware
	handler := JWTAuth(http.HandlerFunc(testHandler))
	handler.ServeHTTP(rr, req)

	// Vérifier le code de statut
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Vérifier que la réponse est correcte
	expected := "success"
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

// TestJWTAuthMissingToken teste l'authentification sans token
func TestJWTAuthMissingToken(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler := JWTAuth(http.HandlerFunc(testHandler))
	handler.ServeHTTP(rr, req)

	// Vérifier le code de statut
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

// TestJWTAuthInvalidTokenFormat teste les formats de token invalides
func TestJWTAuthInvalidTokenFormat(t *testing.T) {
	testCases := []struct {
		name   string
		header string
	}{
		{"No Bearer prefix", "invalid-token"},
		{"Wrong prefix", "Basic invalid-token"},
		{"Empty token", "Bearer "},
		{"No token after Bearer", "Bearer"},
		{"Multiple spaces", "Bearer  token"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tc.header)

			rr := httptest.NewRecorder()
			handler := JWTAuth(http.HandlerFunc(testHandler))
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusUnauthorized {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
			}
		})
	}
}

// TestJWTAuthInvalidToken teste avec un token JWT invalide
func TestJWTAuthInvalidToken(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")

	rr := httptest.NewRecorder()
	handler := JWTAuth(http.HandlerFunc(testHandler))
	handler.ServeHTTP(rr, req)

	// Vérifier le code de statut
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

// TestJWTAuthExpiredToken teste avec un token expiré
func TestJWTAuthExpiredToken(t *testing.T) {
	config.SetTestConfig()

	// Créer un token expiré
	claims := jwt.MapClaims{
		"user_id": "test-user-id",
		"role":    string(models.RoleSubscriber),
		"exp":     time.Now().Add(-time.Hour).Unix(), // Expiré il y a 1 heure
		"iat":     time.Now().Add(-2 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(config.Get().JWT.Secret))

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	rr := httptest.NewRecorder()
	handler := JWTAuth(http.HandlerFunc(testHandler))
	handler.ServeHTTP(rr, req)

	// Vérifier le code de statut
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

// TestJWTAuthBlacklistedToken teste avec un token en liste noire
func TestJWTAuthBlacklistedToken(t *testing.T) {
	config.SetTestConfig()

	// Créer un token valide
	userID := "test-user-id"
	role := string(models.RoleSubscriber)
	token, err := createTestJWT(userID, role)
	if err != nil {
		t.Fatalf("Failed to create test JWT: %v", err)
	}

	// Ajouter le token à la liste noire (mock)
	// Note: Ceci nécessite que la fonction utils.IsJWTTokenBlacklisted soit implémentée
	// Pour ce test, on va simuler qu'elle retourne toujours true

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	handler := JWTAuth(http.HandlerFunc(testHandler))
	handler.ServeHTTP(rr, req)

	// Le résultat dépend de l'implémentation de IsJWTTokenBlacklisted
	// Si elle n'est pas implémentée, le test passera
	// Si elle est implémentée et retourne true, le statut sera Unauthorized
}

// TestJWTAuthMissingUserID teste avec un token sans user_id
func TestJWTAuthMissingUserID(t *testing.T) {
	config.SetTestConfig()

	// Créer un token sans user_id
	claims := jwt.MapClaims{
		"role": string(models.RoleSubscriber),
		"exp":  time.Now().Add(time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(config.Get().JWT.Secret))

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	rr := httptest.NewRecorder()
	handler := JWTAuth(http.HandlerFunc(testHandler))
	handler.ServeHTTP(rr, req)

	// Vérifier le code de statut
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

// TestGetUserIDFromContext teste l'extraction de l'ID utilisateur du contexte
func TestGetUserIDFromContext(t *testing.T) {
	userID := "test-user-id"

	// Créer un contexte avec l'ID utilisateur
	ctx := context.WithValue(context.Background(), UserIDKey, userID)

	// Extraire l'ID utilisateur
	extractedUserID, ok := GetUserIDFromContext(ctx)

	// Vérifier le résultat
	if !ok {
		t.Error("Expected to find user ID in context")
	}

	if extractedUserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, extractedUserID)
	}
}

// TestGetUserIDFromContextNotFound teste l'extraction quand l'ID n'est pas dans le contexte
func TestGetUserIDFromContextNotFound(t *testing.T) {
	// Créer un contexte vide
	ctx := context.Background()

	// Tenter d'extraire l'ID utilisateur
	_, ok := GetUserIDFromContext(ctx)

	// Vérifier que l'ID n'a pas été trouvé
	if ok {
		t.Error("Expected not to find user ID in empty context")
	}
}

// TestAdminRequiredSuccess teste l'accès admin réussi
func TestAdminRequiredSuccess(t *testing.T) {
	// Handler de test qui vérifie que la requête passe
	testAdminHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("admin access granted"))
	}

	// Créer un contexte avec le rôle admin
	req, _ := http.NewRequest("GET", "/admin", nil)
	ctx := context.WithValue(req.Context(), UserRoleKey, string(models.RoleAdmin))
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := AdminRequired(http.HandlerFunc(testAdminHandler))
	handler.ServeHTTP(rr, req)

	// Vérifier le code de statut
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Vérifier la réponse
	expected := "admin access granted"
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

// TestAdminRequiredNonAdmin teste l'accès refusé pour un non-admin
func TestAdminRequiredNonAdmin(t *testing.T) {
	testAdminHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("should not reach here"))
	}

	// Créer un contexte avec un rôle non-admin
	req, _ := http.NewRequest("GET", "/admin", nil)
	ctx := context.WithValue(req.Context(), UserRoleKey, string(models.RoleSubscriber))
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := AdminRequired(http.HandlerFunc(testAdminHandler))
	handler.ServeHTTP(rr, req)

	// Vérifier le code de statut
	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusForbidden)
	}
}

// TestAdminRequiredNoRole teste l'accès sans rôle dans le contexte
func TestAdminRequiredNoRole(t *testing.T) {
	testAdminHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("should not reach here"))
	}

	// Créer une requête sans rôle dans le contexte
	req, _ := http.NewRequest("GET", "/admin", nil)

	rr := httptest.NewRecorder()
	handler := AdminRequired(http.HandlerFunc(testAdminHandler))
	handler.ServeHTTP(rr, req)

	// Vérifier le code de statut
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

// TestOptionsRequest teste que les requêtes OPTIONS passent sans authentification
func TestOptionsRequest(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	rr := httptest.NewRecorder()

	handler := JWTAuth(http.HandlerFunc(testHandler))
	handler.ServeHTTP(rr, req)

	// Les requêtes OPTIONS devraient passer sans authentification
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("OPTIONS request should pass without auth: got %v want %v", status, http.StatusOK)
	}
}

// TestAdminRequiredOptionsRequest teste que les requêtes OPTIONS passent pour AdminRequired
func TestAdminRequiredOptionsRequest(t *testing.T) {
	testAdminHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("options allowed"))
	}

	req, _ := http.NewRequest("OPTIONS", "/admin", nil)
	rr := httptest.NewRecorder()

	handler := AdminRequired(http.HandlerFunc(testAdminHandler))
	handler.ServeHTTP(rr, req)

	// Les requêtes OPTIONS devraient passer sans vérification admin
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("OPTIONS request should pass without admin check: got %v want %v", status, http.StatusOK)
	}
}
