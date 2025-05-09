package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	"github.com/Kaowarstail/Only-Flick-Go/config"
	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
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

// Logout déconnecte l'utilisateur
func Logout(w http.ResponseWriter, r *http.Request) {
	// Dans une implémentation réelle, on pourrait ajouter le token à une liste noire
	// ou utiliser des tokens avec une durée de vie très courte
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Déconnexion réussie"})
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
