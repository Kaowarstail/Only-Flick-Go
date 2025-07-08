package routes

import (
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/config"
	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/handlers"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/Kaowarstail/Only-Flick-Go/internal/services"
	"github.com/gorilla/mux"
)

// RegisterPaymentRoutes enregistre toutes les routes de paiement selon les spécifications API
func RegisterPaymentRoutes(router *mux.Router) {
	// Récupération de la configuration
	cfg := config.Get()

	// Création du service Stripe
	stripeService := services.NewStripeService(cfg.Stripe.SecretKey)

	// Création du handler de paiement
	paymentHandler := handlers.NewPaymentHandler(database.DB, stripeService)

	// Sous-routeur pour l'API v1
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// ===== ROUTES DES MÉTHODES DE PAIEMENT =====
	// GET /api/v1/payments/methods - Méthodes de paiement de l'utilisateur
	apiRouter.Handle("/payments/methods",
		middleware.JWTAuth(http.HandlerFunc(paymentHandler.GetPaymentMethods))).Methods("GET")

	// POST /api/v1/payments/methods - Ajout d'une méthode de paiement
	apiRouter.Handle("/payments/methods",
		middleware.JWTAuth(http.HandlerFunc(paymentHandler.AddPaymentMethod))).Methods("POST")

	// DELETE /api/v1/payments/methods/{id} - Suppression d'une méthode de paiement
	apiRouter.Handle("/payments/methods/{id}",
		middleware.JWTAuth(http.HandlerFunc(paymentHandler.DeletePaymentMethod))).Methods("DELETE")

	// ===== ROUTES DES TRANSACTIONS =====
	// GET /api/v1/transactions - Historique des transactions
	apiRouter.Handle("/transactions",
		middleware.JWTAuth(http.HandlerFunc(paymentHandler.GetTransactions))).Methods("GET")

	// GET /api/v1/transactions/{id} - Détails d'une transaction
	apiRouter.Handle("/transactions/{id}",
		middleware.JWTAuth(http.HandlerFunc(paymentHandler.GetTransaction))).Methods("GET")

	// ===== ROUTES DES REVENUS CRÉATEURS =====
	// GET /api/v1/creators/{id}/earnings - Revenus d'un créateur
	apiRouter.Handle("/creators/{id}/earnings",
		middleware.JWTAuth(http.HandlerFunc(paymentHandler.GetCreatorEarnings))).Methods("GET")

	// ===== ROUTES DES VERSEMENTS =====
	// POST /api/v1/payouts/request - Demande de versement
	apiRouter.Handle("/payouts/request",
		middleware.JWTAuth(http.HandlerFunc(paymentHandler.RequestPayout))).Methods("POST")

	// GET /api/v1/payouts - Historique des versements
	apiRouter.Handle("/payouts",
		middleware.JWTAuth(http.HandlerFunc(paymentHandler.GetPayouts))).Methods("GET")

	// GET /api/v1/payouts/{id} - Détails d'un versement
	apiRouter.Handle("/payouts/{id}",
		middleware.JWTAuth(http.HandlerFunc(paymentHandler.GetPayout))).Methods("GET")

	// ===== ROUTES SPÉCIFIQUES STRIPE =====
	// POST /api/v1/payments/setup-intent - Créer un SetupIntent pour configurer une méthode de paiement
	apiRouter.Handle("/payments/setup-intent",
		middleware.JWTAuth(http.HandlerFunc(paymentHandler.CreateSetupIntent))).Methods("POST")

	// POST /api/v1/payments/process-subscription - Traiter un paiement d'abonnement
	apiRouter.Handle("/payments/process-subscription",
		middleware.JWTAuth(http.HandlerFunc(paymentHandler.ProcessSubscriptionPayment))).Methods("POST")

	// POST /api/v1/payments/webhooks/stripe - Webhook Stripe pour les notifications
	apiRouter.Handle("/payments/webhooks/stripe",
		http.HandlerFunc(paymentHandler.HandleStripeWebhook)).Methods("POST")
}
