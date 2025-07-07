package config

import (
	"encoding/json"
	"fmt"
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
	Cloudinary struct {
		CloudName string
		APIKey    string
		APISecret string
	}
}

var (
	config *Configuration
	once   sync.Once
)

// Load charge la configuration depuis le fichier config.json
func Load() (*Configuration, error) {
	var err error
	once.Do(func() {
		config = &Configuration{}

		// Lire le fichier config.json
		file, fileErr := os.Open("config.json")
		if fileErr != nil {
			// Fallback aux variables d'environnement si config.json n'existe pas
			loadFromEnv()
			return
		}
		defer file.Close()

		// Décoder le JSON
		decoder := json.NewDecoder(file)
		var jsonConfig struct {
			Server struct {
				Port    string `json:"port"`
				Timeout int    `json:"timeout"`
			} `json:"server"`
			Database struct {
				Host     string `json:"host"`
				Port     string `json:"port"`
				User     string `json:"user"`
				Password string `json:"password"`
				DBName   string `json:"dbname"`
				SSLMode  string `json:"sslmode"`
			} `json:"database"`
			JWT struct {
				Secret     string `json:"secret"`
				Expiration int    `json:"expiration"`
			} `json:"jwt"`
			Cloudinary struct {
				CloudName string `json:"cloud_name"`
				APIKey    string `json:"api_key"`
				APISecret string `json:"api_secret"`
			} `json:"cloudinary"`
		}

		if decodeErr := decoder.Decode(&jsonConfig); decodeErr != nil {
			err = fmt.Errorf("erreur lors du décodage de config.json: %w", decodeErr)
			return
		}

		// Mapper les valeurs
		config.Server.Port = jsonConfig.Server.Port
		config.Server.Timeout = jsonConfig.Server.Timeout

		config.Database.Host = jsonConfig.Database.Host
		config.Database.Port = jsonConfig.Database.Port
		config.Database.User = jsonConfig.Database.User
		config.Database.Password = jsonConfig.Database.Password
		config.Database.DBName = jsonConfig.Database.DBName
		config.Database.SSLMode = jsonConfig.Database.SSLMode

		config.JWT.Secret = jsonConfig.JWT.Secret
		config.JWT.Expiration = jsonConfig.JWT.Expiration

		// Cloudinary : priorité aux variables d'environnement, puis config.json
		config.Cloudinary.CloudName = getEnv("CLOUDINARY_CLOUD_NAME", jsonConfig.Cloudinary.CloudName)
		config.Cloudinary.APIKey = getEnv("CLOUDINARY_API_KEY", jsonConfig.Cloudinary.APIKey)
		config.Cloudinary.APISecret = getEnv("CLOUDINARY_API_SECRET", jsonConfig.Cloudinary.APISecret)

		// CORS avec valeurs par défaut
		config.CORS.AllowedOrigins = []string{
			"http://localhost:3000",
			"http://localhost:49219",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:49219",
		}
		config.CORS.AllowCredentials = true
		config.CORS.MaxAge = 86400

		fmt.Printf("🔧 [Config] Cloudinary chargé depuis .env - CloudName: %s, APIKey: %s***\n",
			config.Cloudinary.CloudName,
			config.Cloudinary.APIKey[:min(4, len(config.Cloudinary.APIKey))])
	})

	return config, err
}

// loadFromEnv charge la configuration depuis les variables d'environnement (fallback)
func loadFromEnv() {
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

	// Cloudinary
	config.Cloudinary.CloudName = getEnv("CLOUDINARY_CLOUD_NAME", "your-cloud-name")
	config.Cloudinary.APIKey = getEnv("CLOUDINARY_API_KEY", "your-api-key")
	config.Cloudinary.APISecret = getEnv("CLOUDINARY_API_SECRET", "your-api-secret")
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
