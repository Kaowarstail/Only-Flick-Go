package services

import (
	"context"
	"log"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/models"
	"gorm.io/gorm"
)

// MetricsUpdater gère la mise à jour périodique des métriques
type MetricsUpdater struct {
	updateInterval time.Duration
	stopChan       chan struct{}
}

// NewMetricsUpdater crée un nouveau updater de métriques
func NewMetricsUpdater(interval time.Duration) *MetricsUpdater {
	return &MetricsUpdater{
		updateInterval: interval,
		stopChan:       make(chan struct{}),
	}
}

// Start démarre la mise à jour périodique des métriques
func (mu *MetricsUpdater) Start(ctx context.Context) {
	ticker := time.NewTicker(mu.updateInterval)
	defer ticker.Stop()

	log.Printf("📊 Démarrage de la mise à jour des métriques toutes les %v", mu.updateInterval)

	// Mise à jour initiale
	mu.updateMetrics()

	for {
		select {
		case <-ctx.Done():
			log.Println("📊 Arrêt de la mise à jour des métriques")
			return
		case <-mu.stopChan:
			log.Println("📊 Arrêt de la mise à jour des métriques")
			return
		case <-ticker.C:
			mu.updateMetrics()
		}
	}
}

// Stop arrête la mise à jour des métriques
func (mu *MetricsUpdater) Stop() {
	close(mu.stopChan)
}

// updateMetrics met à jour toutes les métriques depuis la base de données
func (mu *MetricsUpdater) updateMetrics() {
	db := database.GetDB()
	if db == nil {
		log.Println("⚠️  Base de données non disponible pour la mise à jour des métriques")
		return
	}

	// Mettre à jour les métriques utilisateurs
	mu.updateUserMetrics(db)
	
	// Mettre à jour les métriques de contenu
	mu.updateContentMetrics(db)
	
	// Mettre à jour les métriques d'abonnement
	mu.updateSubscriptionMetrics(db)
	
	// Mettre à jour les métriques de base de données
	mu.updateDatabaseMetrics(db)
}

// updateUserMetrics met à jour les métriques liées aux utilisateurs
func (mu *MetricsUpdater) updateUserMetrics(db *gorm.DB) {
	// Utilisateurs actifs (connectés dans les dernières 24h)
	var activeUsers int64
	// Cette requête devrait être adaptée selon votre modèle de données
	// Ici on suppose qu'il y a un champ last_active_at
	db.Model(&models.User{}).
		Where("last_active_at > ?", time.Now().Add(-24*time.Hour)).
		Count(&activeUsers)
	
	UpdateActiveUsers(float64(activeUsers))
}

// updateContentMetrics met à jour les métriques de contenu
func (mu *MetricsUpdater) updateContentMetrics(db *gorm.DB) {
	// Vous pouvez ajouter ici d'autres métriques de contenu
	// Par exemple, le nombre de vues par type de contenu
	
	var imageContents, videoContents int64
	db.Model(&models.Content{}).Where("type = ?", "image").Count(&imageContents)
	db.Model(&models.Content{}).Where("type = ?", "video").Count(&videoContents)
	
	// Ces métriques pourraient être ajoutées aux métriques existantes
	log.Printf("📊 Contenus images: %d, vidéos: %d", imageContents, videoContents)
}

// updateSubscriptionMetrics met à jour les métriques d'abonnement
func (mu *MetricsUpdater) updateSubscriptionMetrics(db *gorm.DB) {
	// Abonnements actifs
	var activeSubscriptions int64
	db.Model(&models.Subscription{}).
		Where("is_active = ? AND end_date > ?", true, time.Now()).
		Count(&activeSubscriptions)
	
	UpdateActiveSubscriptions(float64(activeSubscriptions))
}

// updateDatabaseMetrics met à jour les métriques de base de données
func (mu *MetricsUpdater) updateDatabaseMetrics(db *gorm.DB) {
	// Obtenir les statistiques de connexion de la base de données
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("⚠️  Erreur lors de l'obtention des statistiques DB: %v", err)
		return
	}
	
	stats := sqlDB.Stats()
	UpdateDatabaseConnections(float64(stats.InUse))
	
	// Log des statistiques pour debug
	log.Printf("📊 Connexions DB - In Use: %d, Open: %d, Idle: %d", 
		stats.InUse, stats.OpenConnections, stats.Idle)
}

// Fonction d'initialisation pour démarrer l'updater
func InitMetricsUpdater() *MetricsUpdater {
	// Mise à jour toutes les 30 secondes
	updater := NewMetricsUpdater(30 * time.Second)
	
	// Démarrer dans une goroutine
	go func() {
		ctx := context.Background()
		updater.Start(ctx)
	}()
	
	return updater
}
