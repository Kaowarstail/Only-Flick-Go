package services

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Métriques HTTP
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Métriques utilisateurs
	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Total number of active users",
		},
	)

	UsersRegistered = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "users_registered_total",
			Help: "Total number of registered users",
		},
		[]string{"role"},
	)

	// Métriques de contenu
	ContentCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "content_created_total",
			Help: "Total number of content items created",
		},
		[]string{"type", "creator_id"},
	)

	ContentViews = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "content_views_total",
			Help: "Total number of content views",
		},
		[]string{"content_id", "content_type"},
	)

	// Métriques de upload
	CloudinaryUploads = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cloudinary_upload_attempts_total",
			Help: "Total number of Cloudinary upload attempts",
		},
		[]string{"status"},
	)

	CloudinaryUploadSuccess = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cloudinary_upload_success_total",
			Help: "Total number of successful Cloudinary uploads",
		},
	)

	CloudinaryUploadDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "cloudinary_upload_duration_seconds",
			Help:    "Duration of Cloudinary uploads in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		},
	)

	// Métriques de base de données
	DatabaseConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "database_connections_active",
			Help: "Number of active database connections",
		},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
		},
		[]string{"operation", "table"},
	)

	// Métriques de paiement
	PaymentAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_attempts_total",
			Help: "Total number of payment attempts",
		},
		[]string{"status", "method"},
	)

	PaymentAmount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_amount_euros",
			Help:    "Payment amounts in euros",
			Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
		},
		[]string{"currency"},
	)

	// Métriques d'abonnement
	SubscriptionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "subscriptions_active_total",
			Help: "Total number of active subscriptions",
		},
	)

	SubscriptionsCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "subscriptions_created_total",
			Help: "Total number of subscriptions created",
		},
		[]string{"plan_type"},
	)

	// Métriques de commentaires et likes
	CommentsCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "comments_created_total",
			Help: "Total number of comments created",
		},
	)

	LikesCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "likes_created_total",
			Help: "Total number of likes created",
		},
	)

	// Métriques d'erreur
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "severity"},
	)
)

// RecordHTTPMetrics enregistre les métriques HTTP
func RecordHTTPMetrics(method, endpoint, status string, duration float64) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// RecordUserRegistration enregistre une inscription utilisateur
func RecordUserRegistration(role string) {
	UsersRegistered.WithLabelValues(role).Inc()
}

// RecordContentCreation enregistre la création de contenu
func RecordContentCreation(contentType, creatorID string) {
	ContentCreated.WithLabelValues(contentType, creatorID).Inc()
}

// RecordContentView enregistre une vue de contenu
func RecordContentView(contentID, contentType string) {
	ContentViews.WithLabelValues(contentID, contentType).Inc()
}

// RecordCloudinaryUpload enregistre un upload Cloudinary
func RecordCloudinaryUpload(status string, duration float64) {
	CloudinaryUploads.WithLabelValues(status).Inc()
	if status == "success" {
		CloudinaryUploadSuccess.Inc()
	}
	CloudinaryUploadDuration.Observe(duration)
}

// RecordDatabaseQuery enregistre une requête de base de données
func RecordDatabaseQuery(operation, table string, duration float64) {
	DatabaseQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

// RecordPaymentAttempt enregistre une tentative de paiement
func RecordPaymentAttempt(status, method string, amount float64) {
	PaymentAttempts.WithLabelValues(status, method).Inc()
	PaymentAmount.WithLabelValues("EUR").Observe(amount)
}

// RecordSubscription enregistre une nouvelle souscription
func RecordSubscription(planType string) {
	SubscriptionsCreated.WithLabelValues(planType).Inc()
}

// RecordError enregistre une erreur
func RecordError(errorType, severity string) {
	ErrorsTotal.WithLabelValues(errorType, severity).Inc()
}

// UpdateActiveUsers met à jour le nombre d'utilisateurs actifs
func UpdateActiveUsers(count float64) {
	ActiveUsers.Set(count)
}

// UpdateActiveSubscriptions met à jour le nombre d'abonnements actifs
func UpdateActiveSubscriptions(count float64) {
	SubscriptionsActive.Set(count)
}

// UpdateDatabaseConnections met à jour le nombre de connexions DB actives
func UpdateDatabaseConnections(count float64) {
	DatabaseConnections.Set(count)
}
