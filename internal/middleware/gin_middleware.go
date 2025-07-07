package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
)

// CORSMiddleware configure les en-têtes CORS pour Gin
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// LoggerMiddleware middleware de journalisation pour Gin
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Traiter la requête
		c.Next()
		
		// Calculer la latence
		latency := time.Since(start)
		
		// Obtenir le statut de la réponse
		status := c.Writer.Status()
		
		// Logger les informations
		log.Printf("[%s] %s %s %d %v",
			c.Request.Method,
			c.Request.RequestURI,
			c.ClientIP(),
			status,
			latency,
		)
	}
}

// JWTMiddleware middleware d'authentification JWT pour Gin
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header required",
			})
			c.Abort()
			return
		}

		// Extraire le token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Valider le token
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid token",
			})
			c.Abort()
			return
		}

		// Ajouter les claims au contexte
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// AdminOnlyMiddleware vérifie que l'utilisateur est admin
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CreatorOnlyMiddleware vérifie que l'utilisateur est créateur
func CreatorOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != "creator" && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Creator access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware limite le nombre de requêtes par utilisateur
func RateLimitMiddleware(limit int) gin.HandlerFunc {
	// Utiliser le rate limiter existant
	rateLimiter := utils.NewRateLimiter()
	
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			userID = c.ClientIP() // Utiliser l'IP si pas d'utilisateur
		}

		if !rateLimiter.AllowMessage(userID) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
