package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Handler de test simple
func testHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
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

// TestJWTAuthInvalidTokenFormat teste les formats de token invalides (simple)
func TestJWTAuthInvalidTokenFormat(t *testing.T) {
	testCases := []struct {
		name   string
		header string
	}{
		{"No Bearer prefix", "invalid-token"},
		{"Wrong prefix", "Basic invalid-token"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tc.header)

			rr := httptest.NewRecorder()
			handler := JWTAuth(http.HandlerFunc(testHandler))
			handler.ServeHTTP(rr, req)

			// Vérifier le code de statut
			if status := rr.Code; status != http.StatusUnauthorized {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
			}
		})
	}
}

// TestOptionsRequest teste que les requêtes OPTIONS passent sans authentification
func TestOptionsRequest(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	rr := httptest.NewRecorder()

	handler := JWTAuth(http.HandlerFunc(testHandler))
	handler.ServeHTTP(rr, req)

	// Vérifier le code de statut - OPTIONS doit passer sans token
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("OPTIONS request should pass without token: got %v want %v", status, http.StatusOK)
	}
}

// TestGetUserIDFromContextNotFound teste quand l'ID utilisateur n'est pas dans le contexte
func TestGetUserIDFromContextNotFound(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test", nil)

	userID, ok := GetUserIDFromContext(req.Context())
	if ok || userID != "" {
		t.Errorf("Expected no user ID in context, got %v, %v", userID, ok)
	}
}

// TestAdminRequiredOptionsRequest teste que les requêtes OPTIONS passent sans vérification admin
func TestAdminRequiredOptionsRequest(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", "/admin/test", nil)
	rr := httptest.NewRecorder()

	handler := AdminRequired(http.HandlerFunc(testHandler))
	handler.ServeHTTP(rr, req)

	// Vérifier le code de statut - OPTIONS doit passer
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("OPTIONS request should pass: got %v want %v", status, http.StatusOK)
	}
}
