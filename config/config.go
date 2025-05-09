package config

import (
	"encoding/json"
	"os"
	"sync"
)

// Configuration contient tous les paramètres de configuration de l'application
type Configuration struct {
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
		Expiration int    `json:"expiration"` // En heures
	} `json:"jwt"`
}

var (
	config *Configuration
	once   sync.Once
)

// Load charge la configuration à partir d'un fichier
func Load(configFile string) (*Configuration, error) {
	once.Do(func() {
		config = &Configuration{}

		// On charge les valeurs par défaut
		config.Server.Port = "8080"
		config.Server.Timeout = 15
		config.Database.Host = "localhost"
		config.Database.Port = "5432"
		config.Database.SSLMode = "disable"
		config.JWT.Expiration = 24

		// Si un fichier de config est spécifié, on le charge
		if configFile != "" {
			file, err := os.Open(configFile)
			if err != nil {
				return
			}
			defer file.Close()

			decoder := json.NewDecoder(file)
			if err := decoder.Decode(config); err != nil {
				return
			}
		}

		// On écrase avec les variables d'environnement si elles existent
		if port := os.Getenv("PORT"); port != "" {
			config.Server.Port = port
		}
		// Autres variables d'environnement...
	})

	return config, nil
}

// Get retourne la configuration
func Get() *Configuration {
	if config == nil {
		_, _ = Load("")
	}
	return config
}
