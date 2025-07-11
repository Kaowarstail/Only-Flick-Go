package config

import "sync"

// Configuration de test pour les tests unitaires
var testConfig = &Configuration{
	Server: struct {
		Port    string
		Timeout int
	}{
		Port:    "8080",
		Timeout: 30,
	},
	Database: struct {
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
		SSLMode  string
	}{
		Host:     "localhost",
		Port:     "5432",
		User:     "test",
		Password: "test",
		DBName:   "test_db",
		SSLMode:  "disable",
	},
	JWT: struct {
		Secret     string
		Expiration int
	}{
		Secret:     "test-secret-key-for-jwt-tokens",
		Expiration: 3600, // 1 heure
	},
	CORS: struct {
		AllowedOrigins   []string
		AllowCredentials bool
		MaxAge           int
	}{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           3600,
	},
	Cloudinary: struct {
		CloudName string
		APIKey    string
		APISecret string
	}{
		CloudName: "test",
		APIKey:    "test",
		APISecret: "test",
	},
	Stripe: struct {
		SecretKey      string
		PublishableKey string
		WebhookSecret  string
	}{
		SecretKey:      "test",
		PublishableKey: "test",
		WebhookSecret:  "test",
	},
}

// SetTestConfig configure la configuration pour les tests
func SetTestConfig() {
	config = testConfig
}

// ResetConfig remet la configuration à nil (pour forcer le rechargement)
func ResetConfig() {
	once = sync.Once{}
	config = nil
}
