package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/models"
	"github.com/dgrijalva/jwt-go"
)

// JWTMiddleware vérifie la validité du token JWT
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Token d'authentification requis")
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Vérifier la méthode de signature
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			// Retourner la clé secrète (à remplacer par votre clé)
			return []byte("your-secret-key"), nil
		})

		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Token invalide")
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID := claims["user_id"].(string)
			userRole := claims["role"].(string)

			// Ajouter les informations au contexte
			ctx := context.WithValue(r.Context(), "user_id", userID)
			ctx = context.WithValue(ctx, "user_role", userRole)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			respondWithError(w, http.StatusUnauthorized, "Token invalide")
			return
		}
	})
}

// ConversationParticipantMiddleware vérifie que l'utilisateur est participant à la conversation
func ConversationParticipantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(string)
		if userID == "" {
			respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
			return
		}

		// Extraire l'ID de la conversation depuis l'URL
		conversationID := extractConversationIDFromURL(r.URL.Path)
		if conversationID == "" {
			respondWithError(w, http.StatusBadRequest, "ID de conversation requis")
			return
		}

		// Vérifier si l'utilisateur est participant 
		var conversation models.Conversation
		if err := database.GetDB().First(&conversation, "id = ? AND (user1_id = ? OR user2_id = ?)", 
			conversationID, userID, userID).Error; err != nil {
			respondWithError(w, http.StatusForbidden, "Accès non autorisé à cette conversation")
			return
		}

		// Ajouter l'ID de la conversation au contexte
		ctx := context.WithValue(r.Context(), "conversation_id", conversationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CreatorOnlyMiddleware vérifie que l'utilisateur est un créateur
func CreatorOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Context().Value("user_role").(string)
		if userRole != string(models.RoleCreator) && userRole != string(models.RoleAdmin) {
			respondWithError(w, http.StatusForbidden, "Accès réservé aux créateurs")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AdminOnlyMiddleware vérifie que l'utilisateur est un administrateur
func AdminOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Context().Value("user_role").(string)
		if userRole != string(models.RoleAdmin) {
			respondWithError(w, http.StatusForbidden, "Accès réservé aux administrateurs")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// MediaAccessMiddleware vérifie l'accès aux fichiers média
func MediaAccessMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(string)
		if userID == "" {
			respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
			return
		}

		// Extraire l'ID du fichier depuis l'URL
		fileID := extractFileIDFromURL(r.URL.Path)
		if fileID == "" {
			respondWithError(w, http.StatusBadRequest, "ID de fichier requis")
			return
		}

		// Vérifier l'accès au fichier
		var mediaFile models.MediaFile
		if err := database.GetDB().First(&mediaFile, "id = ?", fileID).Error; err != nil {
			respondWithError(w, http.StatusNotFound, "Fichier non trouvé")
			return
		}

		// Vérifier si l'utilisateur est le propriétaire du fichier
		if mediaFile.UserID == userID {
			next.ServeHTTP(w, r)
			return
		}

		// Vérifier si l'utilisateur a accès via une conversation
		var message models.Message
		if err := database.GetDB().First(&message, "media_url LIKE ?", "%"+fileID+"%").Error; err != nil {
			respondWithError(w, http.StatusForbidden, "Accès non autorisé")
			return
		}

		// Vérifier si l'utilisateur est participant à la conversation
		var conversation models.Conversation
		if err := database.GetDB().First(&conversation, "id = ? AND (user1_id = ? OR user2_id = ?)", 
			message.ConversationID, userID, userID).Error; err != nil {
			respondWithError(w, http.StatusForbidden, "Accès non autorisé")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware limite le nombre de requêtes par utilisateur
func RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value("user_id").(string)
			if userID == "" {
				respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
				return
			}

			// Implémentation simple du rate limiting
			// Dans un vrai projet, on utiliserait Redis ou une solution similaire
			key := "rate_limit:" + userID
			
			// Pour cet exemple, on passe directement à la suite
			// Dans un vrai projet, on vérifierait le nombre de requêtes
			next.ServeHTTP(w, r)
		})
	}
}

// Helper functions

// extractConversationIDFromURL extrait l'ID de conversation depuis l'URL
func extractConversationIDFromURL(path string) string {
	// Implémentation simple - dans un vrai projet, on utiliserait gorilla/mux
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "conversations" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// extractFileIDFromURL extrait l'ID de fichier depuis l'URL
func extractFileIDFromURL(path string) string {
	// Implémentation simple - dans un vrai projet, on utiliserait gorilla/mux
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "media" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// respondWithError envoie une réponse d'erreur
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(`{"error": "` + message + `"}`))
}
