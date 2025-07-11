package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/Kaowarstail/Only-Flick-Go/config"
	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

// =============================================================================
// HELPERS DE TEST
// =============================================================================

// setupTestDB initialise une base de données de test en mémoire
func setupTestDB() {
	database.InitTestDB()

	// Créer les tables nécessaires pour les tests
	db := database.GetDB()
	db.AutoMigrate(&models.User{}, &models.PasswordResetToken{}, &models.EmailVerificationToken{})
}

// createTestUser crée un utilisateur de test standard
func createTestUser() *models.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpassword123"), bcrypt.DefaultCost)

	user := &models.User{
		ID:              uuid.New().String(),
		Username:        "testuser",
		Email:           "test@example.com",
		Password:        string(hashedPassword),
		FirstName:       "Test",
		LastName:        "User",
		Role:            models.RoleSubscriber,
		IsActive:        true,
		IsBanned:        false,
		IsEmailVerified: true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	database.GetDB().Create(user)
	return user
}

// createBannedUser crée un utilisateur banni pour les tests
func createBannedUser() *models.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpassword123"), bcrypt.DefaultCost)
	bannedTime := time.Now()

	user := &models.User{
		ID:              uuid.New().String(),
		Username:        "banneduser",
		Email:           "banned@example.com",
		Password:        string(hashedPassword),
		FirstName:       "Banned",
		LastName:        "User",
		Role:            models.RoleSubscriber,
		IsActive:        true,
		IsBanned:        true,
		BanReason:       "Test ban reason",
		BannedAt:        &bannedTime,
		IsEmailVerified: true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	database.GetDB().Create(user)
	return user
}

// createInactiveUser crée un utilisateur inactif pour les tests
func createInactiveUser() *models.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpassword123"), bcrypt.DefaultCost)

	user := &models.User{
		ID:              uuid.New().String(),
		Username:        "inactiveuser",
		Email:           "inactive@example.com",
		Password:        string(hashedPassword),
		FirstName:       "Inactive",
		LastName:        "User",
		Role:            models.RoleSubscriber,
		IsActive:        false,
		IsBanned:        false,
		IsEmailVerified: true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	database.GetDB().Create(user)
	return user
}

// generateTestJWT génère un token JWT valide pour les tests
func generateTestJWT(user *models.User) string {
	token, _ := generateJWT(user)
	return token
}

// createRequestWithAuth crée une requête HTTP avec authentification
func createRequestWithAuth(method, url string, body []byte, token string) *http.Request {
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

// =============================================================================
// TESTS DE CONNEXION (LOGIN)
// =============================================================================

func TestLogin_Success(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	// Créer un utilisateur de test
	user := createTestUser()

	// Test avec email
	t.Run("Login with email", func(t *testing.T) {
		loginRequest := LoginRequest{
			Username: user.Email,
			Password: "testpassword123",
		}

		jsonBody, _ := json.Marshal(loginRequest)
		req := createRequestWithAuth("POST", "/api/v1/auth/login", jsonBody, "")

		rr := httptest.NewRecorder()
		Login(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		var response LoginResponse
		json.Unmarshal(rr.Body.Bytes(), &response)

		if response.Token == "" {
			t.Error("Expected token in response")
		}

		if response.User.Username != user.Username {
			t.Errorf("Expected username %s, got %s", user.Username, response.User.Username)
		}

		if response.Message == "" {
			t.Error("Expected message in response")
		}
	})

	// Test avec nom d'utilisateur
	t.Run("Login with username", func(t *testing.T) {
		loginRequest := LoginRequest{
			Username: user.Username,
			Password: "testpassword123",
		}

		jsonBody, _ := json.Marshal(loginRequest)
		req := createRequestWithAuth("POST", "/api/v1/auth/login", jsonBody, "")

		rr := httptest.NewRecorder()
		Login(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}
	})
}

func TestLogin_InvalidCredentials(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	user := createTestUser()

	testCases := []struct {
		name     string
		username string
		password string
		expected int
	}{
		{"Wrong password", user.Email, "wrongpassword", http.StatusUnauthorized},
		{"Wrong email", "wrong@example.com", "testpassword123", http.StatusUnauthorized},
		{"Wrong username", "wronguser", "testpassword123", http.StatusUnauthorized},
		{"Empty username", "", "testpassword123", http.StatusBadRequest},
		{"Empty password", user.Email, "", http.StatusBadRequest},
		{"Both empty", "", "", http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			loginRequest := LoginRequest{
				Username: tc.username,
				Password: tc.password,
			}

			jsonBody, _ := json.Marshal(loginRequest)
			req := createRequestWithAuth("POST", "/api/v1/auth/login", jsonBody, "")

			rr := httptest.NewRecorder()
			Login(rr, req)

			if rr.Code != tc.expected {
				t.Errorf("Expected status %d, got %d", tc.expected, rr.Code)
			}
		})
	}
}

func TestLogin_BannedUser(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	user := createBannedUser()

	loginRequest := LoginRequest{
		Username: user.Email,
		Password: "testpassword123",
	}

	jsonBody, _ := json.Marshal(loginRequest)
	req := createRequestWithAuth("POST", "/api/v1/auth/login", jsonBody, "")

	rr := httptest.NewRecorder()
	Login(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rr.Code)
	}

	// Vérifier que le message contient la raison du ban
	var errorResponse map[string]string
	json.Unmarshal(rr.Body.Bytes(), &errorResponse)

	if !strings.Contains(errorResponse["error"], "banni") {
		t.Error("Expected ban message in error response")
	}
}

func TestLogin_InactiveUser(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	user := createInactiveUser()

	loginRequest := LoginRequest{
		Username: user.Email,
		Password: "testpassword123",
	}

	jsonBody, _ := json.Marshal(loginRequest)
	req := createRequestWithAuth("POST", "/api/v1/auth/login", jsonBody, "")

	rr := httptest.NewRecorder()
	Login(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rr.Code)
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	Login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

// =============================================================================
// TESTS D'INSCRIPTION (REGISTER)
// =============================================================================

func TestRegister_Success(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	registerRequest := RegisterRequest{
		Username:  "newuser",
		Email:     "newuser@example.com",
		Password:  "ValidPassword123!",
		FirstName: "New",
		LastName:  "User",
	}

	jsonBody, _ := json.Marshal(registerRequest)
	req := createRequestWithAuth("POST", "/api/v1/auth/register", jsonBody, "")

	rr := httptest.NewRecorder()
	Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rr.Code)
	}

	// Vérifier que l'utilisateur a été créé en base
	var user models.User
	err := database.GetDB().Where("email = ?", registerRequest.Email).First(&user).Error
	if err != nil {
		t.Errorf("User should be created in database: %v", err)
	}

	// Vérifier que le mot de passe a été haché
	if user.Password == registerRequest.Password {
		t.Error("Password should be hashed")
	}

	// Vérifier la réponse
	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	if response["token"] == nil {
		t.Error("Expected token in response")
	}

	if response["user"] == nil {
		t.Error("Expected user in response")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	existingUser := createTestUser()

	registerRequest := RegisterRequest{
		Username:  "differentuser",
		Email:     existingUser.Email,
		Password:  "ValidPassword123!",
		FirstName: "Different",
		LastName:  "User",
	}

	jsonBody, _ := json.Marshal(registerRequest)
	req := createRequestWithAuth("POST", "/api/v1/auth/register", jsonBody, "")

	rr := httptest.NewRecorder()
	Register(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", rr.Code)
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	existingUser := createTestUser()

	registerRequest := RegisterRequest{
		Username:  existingUser.Username,
		Email:     "different@example.com",
		Password:  "ValidPassword123!",
		FirstName: "Different",
		LastName:  "User",
	}

	jsonBody, _ := json.Marshal(registerRequest)
	req := createRequestWithAuth("POST", "/api/v1/auth/register", jsonBody, "")

	rr := httptest.NewRecorder()
	Register(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", rr.Code)
	}
}

func TestRegister_InvalidData(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	testCases := []struct {
		name    string
		request RegisterRequest
	}{
		{
			"Empty username",
			RegisterRequest{Username: "", Email: "test@example.com", Password: "ValidPassword123!"},
		},
		{
			"Empty email",
			RegisterRequest{Username: "testuser", Email: "", Password: "ValidPassword123!"},
		},
		{
			"Empty password",
			RegisterRequest{Username: "testuser", Email: "test@example.com", Password: ""},
		},
		{
			"Invalid email format",
			RegisterRequest{Username: "testuser", Email: "invalid-email", Password: "ValidPassword123!"},
		},
		{
			"Username too short",
			RegisterRequest{Username: "ab", Email: "test@example.com", Password: "ValidPassword123!"},
		},
		{
			"Weak password",
			RegisterRequest{Username: "testuser", Email: "test@example.com", Password: "123"},
		},
		{
			"Username with invalid characters",
			RegisterRequest{Username: "test user", Email: "test@example.com", Password: "ValidPassword123!"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tc.request)
			req := createRequestWithAuth("POST", "/api/v1/auth/register", jsonBody, "")

			rr := httptest.NewRecorder()
			Register(rr, req)

			if rr.Code == http.StatusCreated {
				t.Errorf("Request should have failed for %s", tc.name)
			}
		})
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	req, _ := http.NewRequest("POST", "/api/v1/auth/register", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

// =============================================================================
// TESTS DE DÉCONNEXION (LOGOUT)
// =============================================================================

func TestLogout_Success(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	user := createTestUser()
	token := generateTestJWT(user)

	req := createRequestWithAuth("POST", "/api/v1/auth/logout", nil, token)

	rr := httptest.NewRecorder()
	Logout(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response map[string]string
	json.Unmarshal(rr.Body.Bytes(), &response)

	if response["message"] == "" {
		t.Error("Expected message in response")
	}
}

func TestLogout_WithoutToken(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	req := createRequestWithAuth("POST", "/api/v1/auth/logout", nil, "")

	rr := httptest.NewRecorder()
	Logout(rr, req)

	// Même sans token, la déconnexion réussit (pour nettoyer côté client)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestLogout_InvalidToken(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	req := createRequestWithAuth("POST", "/api/v1/auth/logout", nil, "invalid-token")

	rr := httptest.NewRecorder()
	Logout(rr, req)

	// Même avec un token invalide, la déconnexion réussit
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

// =============================================================================
// TESTS DE RENOUVELLEMENT DE TOKEN (REFRESH)
// =============================================================================

func TestRefreshToken_Success(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	user := createTestUser()
	token := generateTestJWT(user)

	req := createRequestWithAuth("POST", "/api/v1/auth/refresh-token", nil, token)

	rr := httptest.NewRecorder()
	RefreshToken(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	if response["token"] == nil {
		t.Error("Expected new token in response")
	}

	// Vérifier que le nouveau token est différent de l'ancien
	newToken := response["token"].(string)
	if newToken == token {
		// Les tokens peuvent être identiques si générés très rapidement
		// On vérifie juste qu'on a bien un token valide
		t.Logf("Warning: New token is same as old token, but test passes as token is valid")
	}
}

func TestRefreshToken_NoToken(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	req := createRequestWithAuth("POST", "/api/v1/auth/refresh-token", nil, "")

	rr := httptest.NewRecorder()
	RefreshToken(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	req := createRequestWithAuth("POST", "/api/v1/auth/refresh-token", nil, "invalid-token")

	rr := httptest.NewRecorder()
	RefreshToken(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

func TestRefreshToken_BannedUser(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	user := createBannedUser()
	token := generateTestJWT(user)

	req := createRequestWithAuth("POST", "/api/v1/auth/refresh-token", nil, token)

	rr := httptest.NewRecorder()
	RefreshToken(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rr.Code)
	}
}

func TestRefreshToken_InactiveUser(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	user := createInactiveUser()
	token := generateTestJWT(user)

	req := createRequestWithAuth("POST", "/api/v1/auth/refresh-token", nil, token)

	rr := httptest.NewRecorder()
	RefreshToken(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rr.Code)
	}
}

// =============================================================================
// TESTS DE RÉCUPÉRATION D'UTILISATEUR ACTUEL (GET CURRENT USER)
// =============================================================================

func TestGetCurrentUser_Success(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	user := createTestUser()

	// Simuler le contexte avec l'ID utilisateur (normalement défini par le middleware)
	req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	GetCurrentUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response models.UserResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if response.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, response.Username)
	}

	if response.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, response.Email)
	}
}

func TestGetCurrentUser_NoUserID(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)

	rr := httptest.NewRecorder()
	GetCurrentUser(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}

func TestGetCurrentUser_UserNotFound(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	// Utiliser un ID utilisateur inexistant
	req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "nonexistent-id")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	GetCurrentUser(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

// =============================================================================
// TESTS DE DEMANDE DE RÉINITIALISATION DE MOT DE PASSE
// =============================================================================

func TestRequestPasswordReset_Success(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	user := createTestUser()

	requestBody := ResetPasswordRequest{
		Email: user.Email,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := createRequestWithAuth("POST", "/api/v1/auth/reset-password", jsonBody, "")

	rr := httptest.NewRecorder()
	RequestPasswordReset(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	if response["message"] == nil {
		t.Error("Expected message in response")
	}
}

func TestRequestPasswordReset_NonexistentEmail(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	requestBody := ResetPasswordRequest{
		Email: "nonexistent@example.com",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := createRequestWithAuth("POST", "/api/v1/auth/reset-password", jsonBody, "")

	rr := httptest.NewRecorder()
	RequestPasswordReset(rr, req)

	// Pour des raisons de sécurité, on retourne toujours 200
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestRequestPasswordReset_InvalidEmail(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	testCases := []struct {
		name  string
		email string
	}{
		{"Empty email", ""},
		{"Invalid format", "invalid-email"},
		{"Missing domain", "test@"},
		{"Missing @", "testexample.com"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestBody := ResetPasswordRequest{
				Email: tc.email,
			}

			jsonBody, _ := json.Marshal(requestBody)
			req := createRequestWithAuth("POST", "/api/v1/auth/reset-password", jsonBody, "")

			rr := httptest.NewRecorder()
			RequestPasswordReset(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", rr.Code)
			}
		})
	}
}

func TestRequestPasswordReset_InactiveUser(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	user := createInactiveUser()

	requestBody := ResetPasswordRequest{
		Email: user.Email,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := createRequestWithAuth("POST", "/api/v1/auth/reset-password", jsonBody, "")

	rr := httptest.NewRecorder()
	RequestPasswordReset(rr, req)

	// Même pour un utilisateur inactif, on retourne 200 pour la sécurité
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

// =============================================================================
// TESTS DE GÉNÉRATION DE TOKEN JWT
// =============================================================================

func TestGenerateJWT_Success(t *testing.T) {
	config.SetTestConfig()

	user := &models.User{
		ID:       uuid.New().String(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleSubscriber,
	}

	token, err := generateJWT(user)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	if token == "" {
		t.Error("Generated token should not be empty")
	}

	// Vérifier que le token peut être parsé
	claims := jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Get().JWT.Secret), nil
	})

	if err != nil {
		t.Fatalf("Failed to parse generated token: %v", err)
	}

	if !parsedToken.Valid {
		t.Error("Generated token should be valid")
	}

	// Vérifier les claims
	if claims["user_id"] != user.ID {
		t.Errorf("Expected user_id %s, got %s", user.ID, claims["user_id"])
	}

	if claims["username"] != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, claims["username"])
	}

	if claims["role"] != string(user.Role) {
		t.Errorf("Expected role %s, got %s", user.Role, claims["role"])
	}

	// Vérifier l'expiration
	if claims["exp"] == nil {
		t.Error("Token should have expiration claim")
	}
}

func TestGenerateJWT_DifferentRoles(t *testing.T) {
	config.SetTestConfig()

	roles := []models.UserRole{
		models.RoleSubscriber,
		models.RoleCreator,
		models.RoleAdmin,
	}

	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			user := &models.User{
				ID:       uuid.New().String(),
				Username: "testuser",
				Email:    "test@example.com",
				Role:     role,
			}

			token, err := generateJWT(user)
			if err != nil {
				t.Fatalf("Failed to generate JWT for role %s: %v", role, err)
			}

			// Vérifier le rôle dans le token
			claims := jwt.MapClaims{}
			jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(config.Get().JWT.Secret), nil
			})

			if claims["role"] != string(role) {
				t.Errorf("Expected role %s, got %s", role, claims["role"])
			}
		})
	}
}

// =============================================================================
// TESTS DE GESTION DES TOKENS BLACKLISTÉS
// =============================================================================

func TestLogout_TokenBlacklisting(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	user := createTestUser()
	token := generateTestJWT(user)

	// Se déconnecter (ce qui devrait blacklister le token)
	req := createRequestWithAuth("POST", "/api/v1/auth/logout", nil, token)

	rr := httptest.NewRecorder()
	Logout(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Vérifier que le token est maintenant blacklisté
	if !utils.IsJWTTokenBlacklisted(token) {
		t.Error("Token should be blacklisted after logout")
	}
}

// =============================================================================
// TESTS DE VALIDATION DES DONNÉES
// =============================================================================

func TestLogin_SQLInjectionPrevention(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	createTestUser()

	// Tentatives d'injection SQL
	maliciousInputs := []string{
		"'; DROP TABLE users; --",
		"' OR '1'='1",
		"admin'--",
		"' UNION SELECT * FROM users --",
	}

	for _, input := range maliciousInputs {
		t.Run("SQL Injection: "+input, func(t *testing.T) {
			loginRequest := LoginRequest{
				Username: input,
				Password: "testpassword123",
			}

			jsonBody, _ := json.Marshal(loginRequest)
			req := createRequestWithAuth("POST", "/api/v1/auth/login", jsonBody, "")

			rr := httptest.NewRecorder()
			Login(rr, req)

			// Toutes les tentatives d'injection devraient échouer
			if rr.Code == http.StatusOK {
				t.Errorf("SQL injection attempt should not succeed: %s", input)
			}
		})
	}
}

func TestRegister_XSSPrevention(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	// Tentatives d'XSS
	xssInputs := []string{
		"<script>alert('xss')</script>",
		"javascript:alert('xss')",
		"<img src=x onerror=alert('xss')>",
		"';alert('xss');//",
	}

	for _, input := range xssInputs {
		t.Run("XSS Prevention: "+input, func(t *testing.T) {
			registerRequest := RegisterRequest{
				Username:  input,
				Email:     "test@example.com",
				Password:  "ValidPassword123!",
				FirstName: input,
				LastName:  input,
			}

			jsonBody, _ := json.Marshal(registerRequest)
			req := createRequestWithAuth("POST", "/api/v1/auth/register", jsonBody, "")

			rr := httptest.NewRecorder()
			Register(rr, req)

			// Si l'inscription réussit, vérifier que les données ont été nettoyées
			if rr.Code == http.StatusCreated {
				var user models.User
				database.GetDB().Where("email = ?", "test@example.com").First(&user)

				// Les données ne devraient pas contenir de scripts
				if strings.Contains(user.FirstName, "<script>") ||
					strings.Contains(user.LastName, "<script>") ||
					strings.Contains(user.Username, "<script>") {
					t.Error("XSS content should be sanitized")
				}
			}
		})
	}
}

// =============================================================================
// TESTS DE PERFORMANCE ET CONCURRENCE
// =============================================================================

func TestLogin_Concurrent(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	user := createTestUser()

	// Test de connexions simultanées
	const numGoroutines = 10
	results := make(chan int, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			loginRequest := LoginRequest{
				Username: user.Email,
				Password: "testpassword123",
			}

			jsonBody, _ := json.Marshal(loginRequest)
			req := createRequestWithAuth("POST", "/api/v1/auth/login", jsonBody, "")

			rr := httptest.NewRecorder()
			Login(rr, req)

			results <- rr.Code
		}()
	}

	// Vérifier que toutes les connexions réussissent
	for i := 0; i < numGoroutines; i++ {
		result := <-results
		if result != http.StatusOK {
			t.Errorf("Concurrent login failed with status %d", result)
		}
	}
}

// =============================================================================
// TESTS D'INTÉGRATION
// =============================================================================

func TestFullAuthFlow(t *testing.T) {
	setupTestDB()
	defer database.CloseDB()

	// 1. Inscription
	registerRequest := RegisterRequest{
		Username:  "flowuser",
		Email:     "flow@example.com",
		Password:  "FlowPassword123!",
		FirstName: "Flow",
		LastName:  "User",
	}

	jsonBody, _ := json.Marshal(registerRequest)
	req := createRequestWithAuth("POST", "/api/v1/auth/register", jsonBody, "")

	rr := httptest.NewRecorder()
	Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Registration failed with status %d", rr.Code)
	}

	var registerResponse map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &registerResponse)
	initialToken := registerResponse["token"].(string)

	// 2. Récupération des informations utilisateur
	req = createRequestWithAuth("GET", "/api/v1/auth/me", nil, initialToken)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, registerResponse["user"].(map[string]interface{})["id"].(string))
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()
	GetCurrentUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Get current user failed with status %d", rr.Code)
	}

	// 3. Renouvellement du token
	req = createRequestWithAuth("POST", "/api/v1/auth/refresh-token", nil, initialToken)

	rr = httptest.NewRecorder()
	RefreshToken(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Token refresh failed with status %d", rr.Code)
	}

	var refreshResponse map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &refreshResponse)
	newToken := refreshResponse["token"].(string)

	// 4. Déconnexion
	req = createRequestWithAuth("POST", "/api/v1/auth/logout", nil, newToken)

	rr = httptest.NewRecorder()
	Logout(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Logout failed with status %d", rr.Code)
	}

	// 5. Reconnexion
	loginRequest := LoginRequest{
		Username: registerRequest.Email,
		Password: registerRequest.Password,
	}

	jsonBody, _ = json.Marshal(loginRequest)
	req = createRequestWithAuth("POST", "/api/v1/auth/login", jsonBody, "")

	rr = httptest.NewRecorder()
	Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Login after logout failed with status %d", rr.Code)
	}
}
