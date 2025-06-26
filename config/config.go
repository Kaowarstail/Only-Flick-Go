package config

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

// Configuration contient tous les paramètres de l'application
type Configuration struct {
	Server struct {
		Port    string
		Timeout int
	}
	Database struct {
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
		SSLMode  string
	}
	JWT struct {
		Secret     string
		Expiration int
	}
	CORS struct {
		AllowedOrigins   []string
		AllowCredentials bool
		MaxAge           int
	}
}

var (
	config *Configuration
	once   sync.Once
)

// Load charge la configuration depuis les variables d’environnement
func Load() (*Configuration, error) {
	once.Do(func() {
		config = &Configuration{}

		// Server
		config.Server.Port = getEnv("PORT", "8080")
		config.Server.Timeout = getEnvAsInt("TIMEOUT", 15)

		// Database
		config.Database.Host = getEnv("DB_HOST", "localhost")
		config.Database.Port = getEnv("DB_PORT", "5432")
		config.Database.User = getEnv("DB_USER", "postgres")
		config.Database.Password = getEnv("DB_PASSWORD", "")
		config.Database.DBName = getEnv("DB_NAME", "")
		config.Database.SSLMode = getEnv("DB_SSLMODE", "disable")

		// JWT
		config.JWT.Secret = getEnv("JWT_SECRET", "my-secret")
		config.JWT.Expiration = getEnvAsInt("JWT_EXPIRATION", 24)

		// CORS
		config.CORS.AllowedOrigins = getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{
			"http://localhost:3000",
			"http://localhost:49219",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:49219",
		})
		config.CORS.AllowCredentials = getEnvAsBool("CORS_ALLOW_CREDENTIALS", true)
		config.CORS.MaxAge = getEnvAsInt("CORS_MAX_AGE", 86400)
	})
	return config, nil
}

// Get retourne la configuration
func Get() *Configuration {
	if config == nil {
		_, _ = Load()
	}
	return config
}

// Helpers
func getEnv(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultVal int) int {
	if valStr := os.Getenv(key); valStr != "" {
		if val, err := strconv.Atoi(valStr); err == nil {
			return val
		}
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	if valStr := os.Getenv(key); valStr != "" {
		if val, err := strconv.ParseBool(valStr); err == nil {
			return val
		}
	}
	return defaultVal
}

func getEnvAsSlice(key string, defaultVal []string) []string {
	if valStr := os.Getenv(key); valStr != "" {
		var result []string
		for _, v := range strings.Split(valStr, ",") {
			result = append(result, strings.TrimSpace(v))
		}
		return result
	}
	return defaultVal
}
