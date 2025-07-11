package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Kaowarstail/Only-Flick-Go/internal/routes"
	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MessagingTestServer est un serveur de test pour le système de messagerie
type MessagingTestServer struct {
	DB     *gorm.DB
	Router *mux.Router
}

// NewMessagingTestServer crée un nouveau serveur de test
func NewMessagingTestServer() (*MessagingTestServer, error) {
	// Configuration de la base de données
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "password")
	dbname := getEnv("DB_NAME", "onlyflick_test")
	sslmode := getEnv("DB_SSLMODE", "disable")

	// Connexion à la base de données
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("erreur de connexion à la base de données: %v", err)
	}

	// Configuration du routeur
	router := mux.NewRouter()

	// Configuration des routes de messagerie
	routes.SetupMessagingRoutes(router, db)

	// Route de santé
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "messaging-test-server",
		})
	}).Methods("GET")

	return &MessagingTestServer{
		DB:     db,
		Router: router,
	}, nil
}

// Start démarre le serveur de test
func (s *MessagingTestServer) Start(port string) error {
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Serveur de test messagerie démarré sur le port %s", port)
	log.Printf("📋 Endpoints disponibles:")
	log.Printf("   GET  /health - Vérification de santé")
	log.Printf("   GET  /api/conversations - Liste des conversations")
	log.Printf("   POST /api/conversations - Créer une conversation")
	log.Printf("   POST /api/messages - Envoyer un message")
	log.Printf("   GET  /api/messaging/dashboard - Tableau de bord")
	log.Printf("   GET  /api/messaging/search - Recherche")

	return http.ListenAndServe(":"+port, s.Router)
}

// getEnv récupère une variable d'environnement avec une valeur par défaut
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// main fonction principale pour les tests
func main() {
	// Créer le serveur de test
	server, err := NewMessagingTestServer()
	if err != nil {
		log.Fatalf("❌ Erreur lors de la création du serveur: %v", err)
	}

	// Démarrer le serveur
	port := getEnv("PORT", "8080")
	if err := server.Start(port); err != nil {
		log.Fatalf("❌ Erreur lors du démarrage du serveur: %v", err)
	}
}

// ExampleUsage montre comment utiliser les services directement
func ExampleUsage() {
	// Cette fonction montre comment utiliser les services de messagerie
	// directement sans passer par l'API HTTP

	/*
		// Configuration de la base de données
		db, err := gorm.Open(postgres.Open("your-dsn"), &gorm.Config{})
		if err != nil {
			log.Fatal(err)
		}

		// Créer le gestionnaire de services
		manager := services.NewMessagingServiceManager(db)
		messagingService := manager.GetMessagingService()

		// Exemple 1: Récupérer les conversations d'un utilisateur
		conversations, err := messagingService.GetUserConversations("user-123", 1, 20)
		if err != nil {
			log.Printf("Erreur: %v", err)
			return
		}
		fmt.Printf("Conversations trouvées: %d\n", len(conversations.Conversations))

		// Exemple 2: Créer ou récupérer une conversation directe
		conversation, err := messagingService.GetOrCreateDirectConversation("user-123", "user-456")
		if err != nil {
			log.Printf("Erreur: %v", err)
			return
		}
		fmt.Printf("Conversation créée/récupérée: %s\n", conversation.ID)

		// Exemple 3: Envoyer un message
		request := &services.SendMessageRequest{
			ConversationID: conversation.ID,
			Content:        stringPtr("Bonjour ! Comment allez-vous ?"),
			MessageType:    "text",
		}

		message, err := messagingService.SendMessage(request, "user-123")
		if err != nil {
			log.Printf("Erreur: %v", err)
			return
		}
		fmt.Printf("Message envoyé: %s\n", message.ID)

		// Exemple 4: Récupérer le tableau de bord
		dashboard, err := messagingService.GetUserMessagingDashboard("user-123", 1, 10)
		if err != nil {
			log.Printf("Erreur: %v", err)
			return
		}
		fmt.Printf("Dashboard - Conversations: %d, Messages non lus: %d\n",
			len(dashboard.RecentConversations), dashboard.Stats.TotalUnreadMessages)

		// Exemple 5: Recherche
		searchResults, err := messagingService.SearchInMessaging("user-123", "bonjour", "all", 20)
		if err != nil {
			log.Printf("Erreur: %v", err)
			return
		}
		fmt.Printf("Résultats de recherche: %d conversations, %d messages\n",
			len(searchResults.Conversations), len(searchResults.Messages))
	*/
}

// stringPtr est une fonction utilitaire pour créer un pointeur vers une string
func stringPtr(s string) *string {
	return &s
}
