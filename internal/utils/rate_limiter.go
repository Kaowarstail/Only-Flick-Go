package utils

import (
	"sync"
	"time"
)

// RateLimiter gère la limitation du taux de messages
type RateLimiter struct {
	clients map[string]*ClientRate
	mu      sync.RWMutex
}

// ClientRate tracking des messages par client
type ClientRate struct {
	messages []time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter crée un nouveau rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*ClientRate),
	}
}

// AllowMessage vérifie si un message peut être envoyé
func (rl *RateLimiter) AllowMessage(userID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	clientRate, exists := rl.clients[userID]
	if !exists {
		clientRate = &ClientRate{
			messages: []time.Time{},
			limit:    100, // 100 messages per minute
			window:   time.Minute,
		}
		rl.clients[userID] = clientRate
	}

	// Nettoyer les anciens messages
	cutoff := now.Add(-clientRate.window)
	validMessages := []time.Time{}
	for _, msgTime := range clientRate.messages {
		if msgTime.After(cutoff) {
			validMessages = append(validMessages, msgTime)
		}
	}
	clientRate.messages = validMessages

	// Vérifier limite
	if len(clientRate.messages) >= clientRate.limit {
		return false
	}

	// Ajouter nouveau message
	clientRate.messages = append(clientRate.messages, now)
	return true
}

// GetMessageCount retourne le nombre de messages dans la fenêtre
func (rl *RateLimiter) GetMessageCount(userID string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	clientRate, exists := rl.clients[userID]
	if !exists {
		return 0
	}

	now := time.Now()
	cutoff := now.Add(-clientRate.window)
	count := 0
	for _, msgTime := range clientRate.messages {
		if msgTime.After(cutoff) {
			count++
		}
	}

	return count
}

// CleanupOldClients supprime les clients inactifs
func (rl *RateLimiter) CleanupOldClients() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-time.Hour) // Nettoyer après 1 heure

	for userID, clientRate := range rl.clients {
		if len(clientRate.messages) == 0 {
			delete(rl.clients, userID)
			continue
		}

		// Vérifier si le dernier message est trop ancien
		lastMessage := clientRate.messages[len(clientRate.messages)-1]
		if lastMessage.Before(cutoff) {
			delete(rl.clients, userID)
		}
	}
}

// SetLimit définit une limite personnalisée pour un utilisateur
func (rl *RateLimiter) SetLimit(userID string, limit int, window time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	clientRate, exists := rl.clients[userID]
	if !exists {
		clientRate = &ClientRate{
			messages: []time.Time{},
		}
		rl.clients[userID] = clientRate
	}

	clientRate.limit = limit
	clientRate.window = window
}

// GetStats retourne les statistiques du rate limiter
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	totalClients := len(rl.clients)
	totalMessages := 0
	
	for _, clientRate := range rl.clients {
		totalMessages += len(clientRate.messages)
	}

	return map[string]interface{}{
		"total_clients":  totalClients,
		"total_messages": totalMessages,
		"timestamp":      time.Now(),
	}
}
