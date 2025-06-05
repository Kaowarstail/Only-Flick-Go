package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	"github.com/Kaowarstail/Only-Flick-Go/config"
	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

// LoginRequest représente les données de connexion
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse représente la réponse après une connexion réussie
type LoginResponse struct {
	Token   string              `json:"token"`
	User    models.UserResponse `json:"user"`
	Message string              `json:"message"`
}

// Login gère l'authentification des utilisateurs
func Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest LoginRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&loginRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données d'authentification invalides")
		return
	}
	defer r.Body.Close()

	// Recherche de l'utilisateur dans la base de données
	var user models.User
	result := database.GetDB().Where("username = ?", loginRequest.Username).First(&user)
	if result.Error != nil {
		respondWithError(w, http.StatusUnauthorized, "Nom d'utilisateur ou mot de passe incorrect")
		return
	}

	// Vérification du mot de passe
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Nom d'utilisateur ou mot de passe incorrect")
		return
	}

	// Génération du token JWT
	token, err := generateJWT(&user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la génération du token")
		return
	}

	// Réponse avec le token JWT
	response := LoginResponse{
		Token:   token,
		User:    user.ToResponse(),
		Message: "Authentification réussie",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// Logout gère la déconnexion des utilisateurs
// Note: Dans une implémentation avec JWT, la déconnexion côté serveur est souvent gérée via une liste noire de tokens
func Logout(w http.ResponseWriter, r *http.Request) {
	// Dans une implémentation simple, on peut juste demander au client de supprimer le token
	// Pour une implémentation plus sécurisée, on pourrait ajouter le token à une liste noire

	// Effacer le cookie côté client (si on utilise des cookies)
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Déconnecté avec succès"})
}

// RegisterRequest représente les données d'inscription
type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// Register gère l'inscription des nouveaux utilisateurs
func Register(w http.ResponseWriter, r *http.Request) {
	var registerRequest RegisterRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&registerRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données d'inscription invalides")
		return
	}
	defer r.Body.Close()

	// Validation des entrées
	if registerRequest.Username == "" || registerRequest.Email == "" || registerRequest.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Nom d'utilisateur, email et mot de passe sont requis")
		return
	}

	// Vérifier si l'utilisateur ou l'email existe déjà
	var existingUser models.User
	result := database.GetDB().Where("username = ? OR email = ?", registerRequest.Username, registerRequest.Email).First(&existingUser)
	if result.Error == nil {
		respondWithError(w, http.StatusConflict, "Nom d'utilisateur ou email déjà utilisé")
		return
	}

	// Hasher le mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la création du compte")
		return
	}

	// Créer le nouvel utilisateur
	newUser := models.User{
		Username:  registerRequest.Username,
		Email:     registerRequest.Email,
		Password:  string(hashedPassword),
		FirstName: registerRequest.FirstName,
		LastName:  registerRequest.LastName,
		Role:      models.RoleSubscriber,
		IsActive:  true,
	}

	// Sauvegarder l'utilisateur dans la base de données
	result = database.GetDB().Create(&newUser)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la création du compte: "+result.Error.Error())
		return
	}

	// Générer un token JWT pour le nouvel utilisateur
	tokenString, err := generateJWT(&newUser)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la génération du token")
		return
	}

	// Préparer la réponse
	response := LoginResponse{
		Token:   tokenString,
		User:    newUser.ToResponse(),
		Message: "Compte créé avec succès",
	}

	respondWithJSON(w, http.StatusCreated, response)
}

// GenerateJWT génère un token JWT pour un utilisateur
func generateJWT(user *models.User) (string, error) {
	cfg := config.Get()

	// Création des claims (revendications)
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * time.Duration(cfg.JWT.Expiration)).Unix(),
	}

	// Création du token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Signature du token avec la clé secrète
	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// RefreshToken renouvelle le token JWT d'un utilisateur
func RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Extraire le token existant de l'en-tête
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "Token d'authentification manquant")
		return
	}

	// Format attendu: "Bearer {token}"
	tokenString := authHeader[7:]

	// Parser le token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Get().JWT.Secret), nil
	})

	if err != nil || !token.Valid {
		respondWithError(w, http.StatusUnauthorized, "Token invalide")
		return
	}

	// Extraire les claims du token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'extraction des claims")
		return
	}

	// Récupérer l'ID de l'utilisateur
	userID, ok := claims["user_id"].(float64)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "ID utilisateur invalide dans le token")
		return
	}

	// Récupérer l'utilisateur de la base de données
	var user models.User
	result := database.GetDB().First(&user, uint(userID))
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Générer un nouveau token
	newToken, err := generateJWT(&user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la génération du nouveau token")
		return
	}

	// Renvoyer le nouveau token
	respondWithJSON(w, http.StatusOK, map[string]string{
		"token":   newToken,
		"message": "Token renouvelé avec succès",
	})
}

// GetCurrentUser récupère les informations de l'utilisateur actuellement connecté
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Extraire l'ID utilisateur du contexte (défini par le middleware JWTAuth)
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Récupérer l'utilisateur de la base de données
	var user models.User
	result := database.GetDB().First(&user, userID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Renvoyer les informations de l'utilisateur
	respondWithJSON(w, http.StatusOK, user.ToResponse())
}
