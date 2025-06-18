package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	"github.com/Kaowarstail/Only-Flick-Go/config"
	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
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

	// Nettoyer les entrées
	loginRequest.Username = utils.SanitizeInput(loginRequest.Username)
	loginRequest.Password = utils.SanitizeInput(loginRequest.Password)

	// Validation des entrées
	if loginRequest.Username == "" || loginRequest.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Nom d'utilisateur et mot de passe requis")
		return
	}

	// Recherche de l'utilisateur dans la base de données
	var user models.User
	result := database.GetDB().Where("username = ? OR email = ?", loginRequest.Username, loginRequest.Username).First(&user)
	if result.Error != nil {
		respondWithError(w, http.StatusUnauthorized, "Nom d'utilisateur ou mot de passe incorrect")
		return
	}

	// Vérifier si le compte est actif
	if !user.IsActive {
		respondWithError(w, http.StatusForbidden, "Compte désactivé")
		return
	}

	// Vérifier si le compte est banni
	if user.IsBanned {
		message := "Compte banni"
		if user.BanReason != "" {
			message += ": " + user.BanReason
		}
		respondWithError(w, http.StatusForbidden, message)
		return
	}

	// Vérification du mot de passe
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Nom d'utilisateur ou mot de passe incorrect")
		return
	}

	// Mettre à jour la date de dernière connexion
	now := time.Now()
	user.LastLogin = &now
	database.GetDB().Save(&user)

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
func Logout(w http.ResponseWriter, r *http.Request) {
	// Récupérer le token depuis l'en-tête Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenString := authHeader[7:]

		// Parser le token pour obtenir la date d'expiration
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.Get().JWT.Secret), nil
		})

		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if exp, ok := claims["exp"].(float64); ok {
					expiresAt := time.Unix(int64(exp), 0)
					// Ajouter le token à la liste noire
					utils.BlacklistJWTToken(tokenString, expiresAt)
				}
			}
		}
	}

	// Effacer le cookie côté client (si on utilise des cookies)
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
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

	// Nettoyer les entrées
	registerRequest.Username = utils.SanitizeInput(registerRequest.Username)
	registerRequest.Email = utils.SanitizeInput(registerRequest.Email)
	registerRequest.FirstName = utils.SanitizeInput(registerRequest.FirstName)
	registerRequest.LastName = utils.SanitizeInput(registerRequest.LastName)

	// Validation des entrées obligatoires
	if registerRequest.Username == "" || registerRequest.Email == "" || registerRequest.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Nom d'utilisateur, email et mot de passe sont requis")
		return
	}

	// Validation du nom d'utilisateur
	if valid, message := utils.ValidateUsername(registerRequest.Username); !valid {
		respondWithError(w, http.StatusBadRequest, message)
		return
	}

	// Validation de l'email
	if !utils.ValidateEmail(registerRequest.Email) {
		respondWithError(w, http.StatusBadRequest, "Format d'email invalide")
		return
	}

	// Validation du mot de passe
	if valid, message := utils.ValidatePassword(registerRequest.Password); !valid {
		respondWithError(w, http.StatusBadRequest, message)
		return
	}

	// Validation des noms (optionnels)
	if valid, message := utils.ValidateName(registerRequest.FirstName); !valid {
		respondWithError(w, http.StatusBadRequest, "Prénom invalide: "+message)
		return
	}
	if valid, message := utils.ValidateName(registerRequest.LastName); !valid {
		respondWithError(w, http.StatusBadRequest, "Nom de famille invalide: "+message)
		return
	}

	// Vérifier si l'utilisateur ou l'email existe déjà
	var existingUser models.User
	result := database.GetDB().Where("username = ? OR email = ?", registerRequest.Username, registerRequest.Email).First(&existingUser)
	if result.Error == nil {
		if existingUser.Username == registerRequest.Username {
			respondWithError(w, http.StatusConflict, "Ce nom d'utilisateur est déjà utilisé")
		} else {
			respondWithError(w, http.StatusConflict, "Cette adresse email est déjà utilisée")
		}
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
		ID:              uuid.New().String(),
		Username:        registerRequest.Username,
		Email:           registerRequest.Email,
		Password:        string(hashedPassword),
		FirstName:       registerRequest.FirstName,
		LastName:        registerRequest.LastName,
		Role:            models.RoleSubscriber,
		IsActive:        true,
		IsEmailVerified: false, // L'email n'est pas encore vérifié
	}

	// Sauvegarder l'utilisateur dans la base de données
	result = database.GetDB().Create(&newUser)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la création du compte: "+result.Error.Error())
		return
	}

	// Créer un token de vérification d'email
	verificationToken, err := utils.CreateEmailVerificationToken(newUser.ID)
	if err != nil {
		// Log l'erreur mais ne pas empêcher l'inscription
		// TODO: implémenter un système de logging
	}

	// Générer un token JWT pour le nouvel utilisateur
	tokenString, err := generateJWT(&newUser)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la génération du token")
		return
	}

	// Préparer la réponse
	response := map[string]interface{}{
		"token":   tokenString,
		"user":    newUser.ToResponse(),
		"message": "Compte créé avec succès",
	}

	// Ajouter le token de vérification si disponible (pour les tests ou développement)
	if verificationToken != "" {
		response["verification_token"] = verificationToken
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
		"role":     string(user.Role),
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
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		respondWithError(w, http.StatusUnauthorized, "Token d'authentification manquant")
		return
	}

	// Format attendu: "Bearer {token}"
	tokenString := authHeader[7:]

	// Vérifier si le token est en liste noire
	if utils.IsJWTTokenBlacklisted(tokenString) {
		respondWithError(w, http.StatusUnauthorized, "Token révoqué")
		return
	}

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
	userID, ok := claims["user_id"].(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "ID utilisateur invalide dans le token")
		return
	}

	// Récupérer l'utilisateur de la base de données
	var user models.User
	result := database.GetDB().First(&user, "id = ?", userID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Vérifier si le compte est toujours actif
	if !user.IsActive {
		respondWithError(w, http.StatusForbidden, "Compte désactivé")
		return
	}

	if user.IsBanned {
		respondWithError(w, http.StatusForbidden, "Compte banni")
		return
	}

	// Ajouter l'ancien token à la liste noire
	if exp, ok := claims["exp"].(float64); ok {
		expiresAt := time.Unix(int64(exp), 0)
		utils.BlacklistJWTToken(tokenString, expiresAt)
	}

	// Générer un nouveau token
	newToken, err := generateJWT(&user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la génération du nouveau token")
		return
	}

	// Renvoyer le nouveau token
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"token":   newToken,
		"user":    user.ToResponse(),
		"message": "Token renouvelé avec succès",
	})
}

// GetCurrentUser récupère les informations de l'utilisateur actuellement connecté
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Extraire l'ID utilisateur du contexte (défini par le middleware JWTAuth)
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Récupérer l'utilisateur de la base de données
	var user models.User
	result := database.GetDB().First(&user, "id = ?", userID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Renvoyer les informations de l'utilisateur
	respondWithJSON(w, http.StatusOK, user.ToResponse())
}

// ResetPasswordRequest représente la demande de réinitialisation de mot de passe
type ResetPasswordRequest struct {
	Email string `json:"email"`
}

// NewPasswordRequest représente la structure pour définir un nouveau mot de passe
type NewPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// RequestPasswordReset gère les demandes de réinitialisation de mot de passe
func RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var resetRequest ResetPasswordRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&resetRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}
	defer r.Body.Close()

	// Nettoyer l'entrée
	resetRequest.Email = utils.SanitizeInput(resetRequest.Email)

	// Validation de l'email
	if resetRequest.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Email requis")
		return
	}

	if !utils.ValidateEmail(resetRequest.Email) {
		respondWithError(w, http.StatusBadRequest, "Format d'email invalide")
		return
	}

	// Vérifier si l'utilisateur existe
	var user models.User
	result := database.GetDB().Where("email = ?", resetRequest.Email).First(&user)
	if result.Error != nil {
		// Pour des raisons de sécurité, on renvoie toujours le même message
		// même si l'email n'existe pas
		respondWithJSON(w, http.StatusOK, map[string]string{
			"message": "Si l'email existe, un lien de réinitialisation a été envoyé",
		})
		return
	}

	// Vérifier si le compte est actif
	if !user.IsActive {
		respondWithJSON(w, http.StatusOK, map[string]string{
			"message": "Si l'email existe, un lien de réinitialisation a été envoyé",
		})
		return
	}

	// Générer un token de réinitialisation
	resetToken, err := utils.CreatePasswordResetToken(user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la génération du token de réinitialisation")
		return
	}

	// TODO: Envoyer l'email avec le token de réinitialisation
	// Pour l'instant, on inclut le token dans la réponse (uniquement pour les tests/développement)
	response := map[string]interface{}{
		"message": "Si l'email existe, un lien de réinitialisation a été envoyé",
	}

	// En développement, on peut inclure le token pour les tests
	if config.Get().Server.Port == "8080" { // Mode développement
		response["reset_token"] = resetToken
	}

	respondWithJSON(w, http.StatusOK, response)
}

// ResetPassword gère la réinitialisation effective du mot de passe avec token
func ResetPassword(w http.ResponseWriter, r *http.Request) {
	// Récupérer le token depuis les paramètres de l'URL
	vars := mux.Vars(r)
	token := vars["token"]

	if token == "" {
		respondWithError(w, http.StatusBadRequest, "Token de réinitialisation requis")
		return
	}

	var newPasswordRequest struct {
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newPasswordRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}
	defer r.Body.Close()

	// Nettoyer les entrées
	newPasswordRequest.Password = utils.SanitizeInput(newPasswordRequest.Password)
	newPasswordRequest.ConfirmPassword = utils.SanitizeInput(newPasswordRequest.ConfirmPassword)

	// Validation des données
	if newPasswordRequest.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Nouveau mot de passe requis")
		return
	}

	if newPasswordRequest.Password != newPasswordRequest.ConfirmPassword {
		respondWithError(w, http.StatusBadRequest, "Les mots de passe ne correspondent pas")
		return
	}

	// Validation du nouveau mot de passe
	if valid, message := utils.ValidatePassword(newPasswordRequest.Password); !valid {
		respondWithError(w, http.StatusBadRequest, message)
		return
	}

	// Valider le token de réinitialisation
	resetToken, err := utils.ValidatePasswordResetToken(token)
	if err != nil {
		if err == utils.ErrTokenExpired {
			respondWithError(w, http.StatusBadRequest, "Token de réinitialisation expiré")
		} else {
			respondWithError(w, http.StatusBadRequest, "Token de réinitialisation invalide")
		}
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	result := database.GetDB().First(&user, "id = ?", resetToken.UserID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Hasher le nouveau mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPasswordRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors du hachage du mot de passe")
		return
	}

	// Mettre à jour le mot de passe
	user.Password = string(hashedPassword)
	if err := database.GetDB().Save(&user).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour du mot de passe")
		return
	}

	// Marquer le token comme utilisé
	if err := utils.UsePasswordResetToken(resetToken.ID); err != nil {
		// Log l'erreur mais continuer
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Mot de passe réinitialisé avec succès",
	})
}

// VerifyEmail vérifie l'adresse email d'un utilisateur
func VerifyEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	if token == "" {
		respondWithError(w, http.StatusBadRequest, "Token de vérification requis")
		return
	}

	// Valider le token de vérification
	verificationToken, err := utils.ValidateEmailVerificationToken(token)
	if err != nil {
		if err == utils.ErrTokenExpired {
			respondWithError(w, http.StatusBadRequest, "Token de vérification expiré")
		} else {
			respondWithError(w, http.StatusBadRequest, "Token de vérification invalide")
		}
		return
	}

	// Mettre à jour le statut de vérification de l'utilisateur
	err = database.GetDB().Model(&models.User{}).
		Where("id = ?", verificationToken.UserID).
		Update("is_email_verified", true).Error

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la vérification de l'email")
		return
	}

	// Marquer le token comme utilisé
	if err := utils.UseEmailVerificationToken(verificationToken.ID); err != nil {
		// Log l'erreur mais continuer
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Email vérifié avec succès",
	})
}

// ResendEmailVerification renvoie un email de vérification
func ResendEmailVerification(w http.ResponseWriter, r *http.Request) {
	// Récupérer l'ID utilisateur du contexte
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	result := database.GetDB().First(&user, "id = ?", userID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Vérifier si l'email n'est pas déjà vérifié
	if user.IsEmailVerified {
		respondWithError(w, http.StatusBadRequest, "Email déjà vérifié")
		return
	}

	// Créer un nouveau token de vérification
	verificationToken, err := utils.CreateEmailVerificationToken(user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la génération du token de vérification")
		return
	}

	// TODO: Envoyer l'email de vérification
	// Pour l'instant, on inclut le token dans la réponse (uniquement pour les tests/développement)
	response := map[string]interface{}{
		"message": "Un nouvel email de vérification a été envoyé",
	}

	// En développement, on peut inclure le token pour les tests
	if config.Get().Server.Port == "8080" { // Mode développement
		response["verification_token"] = verificationToken
	}

	respondWithJSON(w, http.StatusOK, response)
}

// ChangePassword permet à un utilisateur connecté de changer son mot de passe
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Récupérer l'ID utilisateur du contexte
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	var changePasswordRequest struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&changePasswordRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}
	defer r.Body.Close()

	// Nettoyer les entrées
	changePasswordRequest.CurrentPassword = utils.SanitizeInput(changePasswordRequest.CurrentPassword)
	changePasswordRequest.NewPassword = utils.SanitizeInput(changePasswordRequest.NewPassword)
	changePasswordRequest.ConfirmPassword = utils.SanitizeInput(changePasswordRequest.ConfirmPassword)

	// Validation des données
	if changePasswordRequest.CurrentPassword == "" || changePasswordRequest.NewPassword == "" {
		respondWithError(w, http.StatusBadRequest, "Mot de passe actuel et nouveau mot de passe requis")
		return
	}

	if changePasswordRequest.NewPassword != changePasswordRequest.ConfirmPassword {
		respondWithError(w, http.StatusBadRequest, "Les nouveaux mots de passe ne correspondent pas")
		return
	}

	// Validation du nouveau mot de passe
	if valid, message := utils.ValidatePassword(changePasswordRequest.NewPassword); !valid {
		respondWithError(w, http.StatusBadRequest, message)
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	result := database.GetDB().First(&user, "id = ?", userID)
	if result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Vérifier le mot de passe actuel
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(changePasswordRequest.CurrentPassword)); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Mot de passe actuel incorrect")
		return
	}

	// Vérifier que le nouveau mot de passe est différent de l'ancien
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(changePasswordRequest.NewPassword)); err == nil {
		respondWithError(w, http.StatusBadRequest, "Le nouveau mot de passe doit être différent de l'ancien")
		return
	}

	// Hasher le nouveau mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(changePasswordRequest.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors du hachage du mot de passe")
		return
	}

	// Mettre à jour le mot de passe
	user.Password = string(hashedPassword)
	if err := database.GetDB().Save(&user).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour du mot de passe")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Mot de passe modifié avec succès",
	})
}
