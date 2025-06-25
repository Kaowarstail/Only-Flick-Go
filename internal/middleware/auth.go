package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"

	"github.com/Kaowarstail/Only-Flick-Go/config"
	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
)

// Clés de contexte pour stocker des informations dans la requête HTTP
type contextKey string

const (
	UserIDKey   contextKey = "userID"
	UserRoleKey contextKey = "userRole"
)

// JWTAuth authentifie les requêtes à l'aide de JWT
func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Permettre les requêtes preflight OPTIONS sans authentification
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}
		
		// Récupération du token depuis l'en-tête Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Token d'authentification manquant", http.StatusUnauthorized)
			return
		}

		// Format attendu: "Bearer {token}"
		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			http.Error(w, "Format de token invalide", http.StatusUnauthorized)
			return
		}

		tokenString := bearerToken[1]

		// Vérifier si le token est en liste noire
		if utils.IsJWTTokenBlacklisted(tokenString) {
			http.Error(w, "Token révoqué", http.StatusUnauthorized)
			return
		}

		// Vérification du token
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.Get().JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Token invalide ou expiré", http.StatusUnauthorized)
			return
		}

		// Extraction des informations utilisateur du token
		userID, ok := claims["user_id"].(string)
		if !ok {
			http.Error(w, "Token invalide: ID utilisateur manquant", http.StatusUnauthorized)
			return
		}

		userRole, _ := claims["role"].(string)

		// Ajout de l'ID utilisateur et du rôle au contexte de la requête
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, UserRoleKey, userRole)

		// Appel du gestionnaire suivant avec le contexte enrichi
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext extrait l'ID utilisateur du contexte
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
