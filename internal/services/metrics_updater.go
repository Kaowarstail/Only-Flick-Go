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
	db.Model(&models.User{}).
		Where("last_active_at > ? OR last_login > ?",
			time.Now().Add(-24*time.Hour),
			time.Now().Add(-24*time.Hour)).
		Count(&activeUsers)

	UpdateActiveUsers(float64(activeUsers))

	// Total d'utilisateurs par rôle
	var totalUsers, creators, subscribers int64
	db.Model(&models.User{}).Count(&totalUsers)
	db.Model(&models.User{}).Where("role = ?", models.RoleCreator).Count(&creators)
	db.Model(&models.User{}).Where("role = ?", models.RoleSubscriber).Count(&subscribers)

	// Mettre à jour les métriques par rôle
	UpdateUsersByRole(string(models.RoleCreator), float64(creators))
	UpdateUsersByRole(string(models.RoleSubscriber), float64(subscribers))
	UpdateUsersByRole("total", float64(totalUsers))

	log.Printf("📊 Utilisateurs - Actifs: %d, Total: %d, Créateurs: %d, Abonnés: %d",
		activeUsers, totalUsers, creators, subscribers)
}

// updateContentMetrics met à jour les métriques de contenu
func (mu *MetricsUpdater) updateContentMetrics(db *gorm.DB) {
	// Total de contenus par type
	var totalContents, imageContents, videoContents int64
	db.Model(&models.Content{}).Count(&totalContents)
	db.Model(&models.Content{}).Where("type = ?", "image").Count(&imageContents)
	db.Model(&models.Content{}).Where("type = ?", "video").Count(&videoContents)

	// Mettre à jour les métriques par type
	UpdateContentByType("image", float64(imageContents))
	UpdateContentByType("video", float64(videoContents))
	UpdateContentByType("total", float64(totalContents))

	// Contenus premium vs gratuits
	var premiumContents, freeContents int64
	db.Model(&models.Content{}).Where("is_premium = ?", true).Count(&premiumContents)
	db.Model(&models.Content{}).Where("is_premium = ?", false).Count(&freeContents)

	UpdateContentByType("premium", float64(premiumContents))
	UpdateContentByType("free", float64(freeContents))

	// Total des vues
	var totalViews int64
	db.Model(&models.Content{}).Select("COALESCE(SUM(view_count), 0)").Scan(&totalViews)
	UpdateTotalContentViews(float64(totalViews))

	// Total des commentaires et likes
	var totalComments, totalLikes int64
	db.Model(&models.Comment{}).Count(&totalComments)
	db.Model(&models.Like{}).Count(&totalLikes)

	log.Printf("📊 Contenus - Total: %d (Images: %d, Vidéos: %d), Premium: %d, Vues: %d, Commentaires: %d, Likes: %d",
		totalContents, imageContents, videoContents, premiumContents, totalViews, totalComments, totalLikes)
}

// updateSubscriptionMetrics met à jour les métriques d'abonnement
func (mu *MetricsUpdater) updateSubscriptionMetrics(db *gorm.DB) {
	// Abonnements actifs
	var activeSubscriptions int64
	db.Model(&models.Subscription{}).
		Where("is_active = ? AND end_date > ?", true, time.Now()).
		Count(&activeSubscriptions)

	UpdateActiveSubscriptions(float64(activeSubscriptions))

	// Total abonnements créés aujourd'hui
	var todaySubscriptions int64
	today := time.Now().Truncate(24 * time.Hour)
	db.Model(&models.Subscription{}).
		Where("created_at >= ?", today).
		Count(&todaySubscriptions)

	// Revenus totaux (approximation basée sur les abonnements actifs)
	var totalRevenue float64
	db.Model(&models.Subscription{}).
		Joins("JOIN subscription_plans ON subscriptions.plan_id = subscription_plans.id").
		Where("subscriptions.is_active = ? AND subscriptions.end_date > ?", true, time.Now()).
		Select("COALESCE(SUM(subscription_plans.price), 0)").
		Scan(&totalRevenue)

	UpdateTotalRevenue(totalRevenue)

	log.Printf("📊 Abonnements - Actifs: %d, Nouveaux aujourd'hui: %d, Revenus: %.2f€",
		activeSubscriptions, todaySubscriptions, totalRevenue)
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
