package utils

import (
	"sync"
	"time"
)

// WebSocketMetrics collecte les métriques WebSocket
type WebSocketMetrics struct {
	ActiveConnections int64
	TotalMessages     int64
	MessagesSent      int64
	MessagesReceived  int64
	ConnectionsOpened int64
	ConnectionsClosed int64
	ErrorsCount       int64
	TypingEvents      int64
	UserStatusEvents  int64
	
	// Rate limiting stats
	RateLimitedMessages int64
	
	// Performance metrics
	AverageMessageSize  float64
	PeakConnections     int64
	
	// Timestamps
	StartTime   time.Time
	LastMessage time.Time
	LastError   time.Time
	
	mu sync.RWMutex
}

// NewWebSocketMetrics crée de nouvelles métriques
func NewWebSocketMetrics() *WebSocketMetrics {
	return &WebSocketMetrics{
		StartTime: time.Now(),
	}
}

// IncrementConnections incrémente les connexions
func (m *WebSocketMetrics) IncrementConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveConnections++
	m.ConnectionsOpened++
	
	if m.ActiveConnections > m.PeakConnections {
		m.PeakConnections = m.ActiveConnections
	}
}

// DecrementConnections décrémente les connexions
func (m *WebSocketMetrics) DecrementConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ActiveConnections > 0 {
		m.ActiveConnections--
	}
	m.ConnectionsClosed++
}

// IncrementMessagesSent incrémente les messages envoyés
func (m *WebSocketMetrics) IncrementMessagesSent() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesSent++
	m.TotalMessages++
	m.LastMessage = time.Now()
}

// IncrementMessagesReceived incrémente les messages reçus
func (m *WebSocketMetrics) IncrementMessagesReceived() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesReceived++
	m.TotalMessages++
	m.LastMessage = time.Now()
}

// IncrementErrors incrémente les erreurs
func (m *WebSocketMetrics) IncrementErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorsCount++
	m.LastError = time.Now()
}

// IncrementTypingEvents incrémente les événements de frappe
func (m *WebSocketMetrics) IncrementTypingEvents() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TypingEvents++
}

// IncrementUserStatusEvents incrémente les événements de statut
func (m *WebSocketMetrics) IncrementUserStatusEvents() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UserStatusEvents++
}

// IncrementRateLimited incrémente les messages limités
func (m *WebSocketMetrics) IncrementRateLimited() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RateLimitedMessages++
}

// UpdateMessageSize met à jour la taille moyenne des messages
func (m *WebSocketMetrics) UpdateMessageSize(size int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Calcul de la moyenne mobile
	if m.AverageMessageSize == 0 {
		m.AverageMessageSize = float64(size)
	} else {
		m.AverageMessageSize = (m.AverageMessageSize*0.9) + (float64(size)*0.1)
	}
}

// GetMetrics retourne toutes les métriques
func (m *WebSocketMetrics) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	uptime := time.Since(m.StartTime)
	
	return map[string]interface{}{
		"active_connections":     m.ActiveConnections,
		"total_messages":         m.TotalMessages,
		"messages_sent":          m.MessagesSent,
		"messages_received":      m.MessagesReceived,
		"connections_opened":     m.ConnectionsOpened,
		"connections_closed":     m.ConnectionsClosed,
		"errors_count":           m.ErrorsCount,
		"typing_events":          m.TypingEvents,
		"user_status_events":     m.UserStatusEvents,
		"rate_limited_messages":  m.RateLimitedMessages,
		"average_message_size":   m.AverageMessageSize,
		"peak_connections":       m.PeakConnections,
		"uptime_seconds":         uptime.Seconds(),
		"start_time":             m.StartTime,
		"last_message":           m.LastMessage,
		"last_error":             m.LastError,
		"messages_per_second":    m.calculateMessagesPerSecond(uptime),
		"error_rate":             m.calculateErrorRate(),
	}
}

// GetSimpleMetrics retourne les métriques basiques
func (m *WebSocketMetrics) GetSimpleMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return map[string]interface{}{
		"active_connections": m.ActiveConnections,
		"total_messages":     m.TotalMessages,
		"errors_count":       m.ErrorsCount,
		"uptime_seconds":     time.Since(m.StartTime).Seconds(),
	}
}

// calculateMessagesPerSecond calcule les messages par seconde
func (m *WebSocketMetrics) calculateMessagesPerSecond(uptime time.Duration) float64 {
	if uptime.Seconds() == 0 {
		return 0
	}
	return float64(m.TotalMessages) / uptime.Seconds()
}

// calculateErrorRate calcule le taux d'erreur
func (m *WebSocketMetrics) calculateErrorRate() float64 {
	if m.TotalMessages == 0 {
		return 0
	}
	return float64(m.ErrorsCount) / float64(m.TotalMessages) * 100
}

// Reset remet à zéro les métriques
func (m *WebSocketMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.TotalMessages = 0
	m.MessagesSent = 0
	m.MessagesReceived = 0
	m.ConnectionsOpened = 0
	m.ConnectionsClosed = 0
	m.ErrorsCount = 0
	m.TypingEvents = 0
	m.UserStatusEvents = 0
	m.RateLimitedMessages = 0
	m.AverageMessageSize = 0
	m.PeakConnections = m.ActiveConnections
	m.StartTime = time.Now()
	m.LastMessage = time.Time{}
	m.LastError = time.Time{}
}
