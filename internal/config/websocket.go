package config

import (
	"time"
)

// WebSocketConfig configuration pour les WebSockets
type WebSocketConfig struct {
	// Buffer sizes
	ReadBufferSize  int `env:"WS_READ_BUFFER_SIZE" envDefault:"1024"`
	WriteBufferSize int `env:"WS_WRITE_BUFFER_SIZE" envDefault:"1024"`
	
	// Timeouts
	HandshakeTimeout time.Duration `env:"WS_HANDSHAKE_TIMEOUT" envDefault:"10s"`
	WriteTimeout     time.Duration `env:"WS_WRITE_TIMEOUT" envDefault:"10s"`
	ReadTimeout      time.Duration `env:"WS_READ_TIMEOUT" envDefault:"60s"`
	PingPeriod       time.Duration `env:"WS_PING_PERIOD" envDefault:"54s"`
	PongWait         time.Duration `env:"WS_PONG_WAIT" envDefault:"60s"`
	
	// Rate limiting
	MessageRateLimit int `env:"WS_MESSAGE_RATE_LIMIT" envDefault:"100"` // messages per minute
	ConnectionLimit  int `env:"WS_CONNECTION_LIMIT" envDefault:"1000"`  // max connections
	
	// Security
	AllowedOrigins []string `env:"WS_ALLOWED_ORIGINS" envDefault:"https://onlyflick.app,http://localhost:3000"`
	
	// Features
	EnableTypingIndicators bool `env:"WS_ENABLE_TYPING" envDefault:"true"`
	EnablePresence         bool `env:"WS_ENABLE_PRESENCE" envDefault:"true"`
	EnableMetrics          bool `env:"WS_ENABLE_METRICS" envDefault:"true"`
	
	// Auto-cleanup
	TypingTimeout      time.Duration `env:"WS_TYPING_TIMEOUT" envDefault:"3s"`
	InactivityTimeout  time.Duration `env:"WS_INACTIVITY_TIMEOUT" envDefault:"30m"`
	CleanupInterval    time.Duration `env:"WS_CLEANUP_INTERVAL" envDefault:"5m"`
}

// GetWebSocketConfig retourne la configuration WebSocket
func GetWebSocketConfig() WebSocketConfig {
	return WebSocketConfig{
		ReadBufferSize:         1024,
		WriteBufferSize:        1024,
		HandshakeTimeout:       10 * time.Second,
		WriteTimeout:           10 * time.Second,
		ReadTimeout:            60 * time.Second,
		PingPeriod:             54 * time.Second,
		PongWait:               60 * time.Second,
		MessageRateLimit:       100,
		ConnectionLimit:        1000,
		AllowedOrigins:         []string{"https://onlyflick.app", "http://localhost:3000"},
		EnableTypingIndicators: true,
		EnablePresence:         true,
		EnableMetrics:          true,
		TypingTimeout:          3 * time.Second,
		InactivityTimeout:      30 * time.Minute,
		CleanupInterval:        5 * time.Minute,
	}
}
