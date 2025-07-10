package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter structure pour gérer les limiteurs par utilisateur
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter crée un nouveau limiteur de débit
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// GetLimiter récupère ou crée un limiteur pour un utilisateur
func (rl *RateLimiter) GetLimiter(userID string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[userID]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[userID] = limiter
	}

	return limiter
}

// MessageRateLimit middleware pour limiter l'envoi de messages
func MessageRateLimit() func(http.Handler) http.Handler {
	// 10 messages par minute par utilisateur
	rateLimiter := NewRateLimiter(rate.Every(6*time.Second), 10)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Récupérer l'ID utilisateur du contexte
			userID, ok := GetUserIDFromContext(r.Context())
			if !ok {
				http.Error(w, "User ID not found", http.StatusUnauthorized)
				return
			}

			limiter := rateLimiter.GetLimiter(userID)
			if !limiter.Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{
					"success": false,
					"error": {
						"code": "RATE_LIMIT_EXCEEDED",
						"message": "Too many messages. Please slow down."
					}
				}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// PaidMessageRateLimit middleware plus restrictif pour les messages payants
func PaidMessageRateLimit() func(http.Handler) http.Handler {
	// 5 messages payants par minute par utilisateur
	rateLimiter := NewRateLimiter(rate.Every(12*time.Second), 5)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := GetUserIDFromContext(r.Context())
			if !ok {
				http.Error(w, "User ID not found", http.StatusUnauthorized)
				return
			}

			limiter := rateLimiter.GetLimiter(userID)
			if !limiter.Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{
					"success": false,
					"error": {
						"code": "PAID_MESSAGE_RATE_LIMIT_EXCEEDED",
						"message": "Too many paid messages. Please wait before sending another."
					}
				}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
