package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/config"
	"github.com/Kaowarstail/Only-Flick-Go/internal/services"
	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
	"github.com/Kaowarstail/Only-Flick-Go/models"
	"github.com/gorilla/mux"
	"github.com/stripe/stripe-go/v76"
	"gorm.io/gorm"
)

// PaymentHandler gère les endpoints de paiement
type PaymentHandler struct {
	db            *gorm.DB
	stripeService *services.StripeService
}

// NewPaymentHandler crée une nouvelle instance du handler de paiement
func NewPaymentHandler(db *gorm.DB, stripeService *services.StripeService) *PaymentHandler {
	return &PaymentHandler{
		db:            db,
		stripeService: stripeService,
	}
}

// PaymentMethodRequest représente une demande d'ajout de méthode de paiement
type PaymentMethodRequest struct {
	PaymentMethodID string `json:"payment_method_id" validate:"required"`
	IsDefault       bool   `json:"is_default"`
}

// PayoutRequest représente une demande de versement
type PayoutRequest struct {
	Amount float64 `json:"amount" validate:"required,min=1"`
}

// GetPaymentMethods récupère les méthodes de paiement de l'utilisateur
func (h *PaymentHandler) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var paymentMethods []models.PaymentMethod
	if err := h.db.Where("user_id = ?", userID).Find(&paymentMethods).Error; err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors de la récupération des méthodes de paiement", err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"payment_methods": paymentMethods,
	})
}

// AddPaymentMethod ajoute une nouvelle méthode de paiement
func (h *PaymentHandler) AddPaymentMethod(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req PaymentMethodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Données JSON invalides", err)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Données de validation invalides", err)
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Utilisateur non trouvé", err)
		return
	}

	// Récupérer les détails de la méthode de paiement depuis Stripe
	stripePM, err := h.stripeService.GetPaymentMethod(req.PaymentMethodID)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Méthode de paiement invalide", err)
		return
	}

	// Si c'est la méthode par défaut, désactiver les autres
	if req.IsDefault {
		h.db.Model(&models.PaymentMethod{}).Where("user_id = ?", userID).Update("is_default", false)
	}

	// Créer la méthode de paiement en base
	paymentMethod := &models.PaymentMethod{
		UserID:         userID,
		StripeMethodID: req.PaymentMethodID,
		Type:           string(stripePM.Type),
		IsDefault:      req.IsDefault,
	}

	if stripePM.Card != nil {
		paymentMethod.Last4 = stripePM.Card.Last4
		paymentMethod.Brand = string(stripePM.Card.Brand)
		paymentMethod.ExpiryMonth = int(stripePM.Card.ExpMonth)
		paymentMethod.ExpiryYear = int(stripePM.Card.ExpYear)
	}

	if err := h.db.Create(paymentMethod).Error; err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors de l'enregistrement de la méthode de paiement", err)
		return
	}

	utils.SendJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"payment_method": paymentMethod,
		"message":        "Méthode de paiement ajoutée avec succès",
	})
}

// DeletePaymentMethod supprime une méthode de paiement
func (h *PaymentHandler) DeletePaymentMethod(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	vars := mux.Vars(r)
	paymentMethodID := vars["id"]

	// Récupérer la méthode de paiement
	var paymentMethod models.PaymentMethod
	if err := h.db.Where("id = ? AND user_id = ?", paymentMethodID, userID).First(&paymentMethod).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Méthode de paiement non trouvée", err)
		return
	}

	// Détacher la méthode de paiement de Stripe
	if _, err := h.stripeService.DetachPaymentMethod(paymentMethod.StripeMethodID); err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors de la suppression sur Stripe", err)
		return
	}

	// Supprimer de la base de données
	if err := h.db.Delete(&paymentMethod).Error; err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors de la suppression de la méthode de paiement", err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Méthode de paiement supprimée avec succès",
	})
}

// GetTransactions récupère l'historique des transactions de l'utilisateur
func (h *PaymentHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	// Paramètres de pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var transactions []models.Transaction
	var total int64

	// Compter le total
	h.db.Model(&models.Transaction{}).Where("user_id = ?", userID).Count(&total)

	// Récupérer les transactions avec pagination
	offset := (page - 1) * limit
	if err := h.db.Where("user_id = ?", userID).
		Preload("Subscription").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&transactions).Error; err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors de la récupération des transactions", err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"transactions": transactions,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetTransaction récupère les détails d'une transaction spécifique
func (h *PaymentHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	vars := mux.Vars(r)
	transactionID := vars["id"]

	var transaction models.Transaction
	if err := h.db.Where("id = ? AND user_id = ?", transactionID, userID).
		Preload("Subscription").
		First(&transaction).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Transaction non trouvée", err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"transaction": transaction,
	})
}

// RequestPayout demande un versement pour un créateur
func (h *PaymentHandler) RequestPayout(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req PayoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Données JSON invalides", err)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Données de validation invalides", err)
		return
	}

	// Vérifier que l'utilisateur est un créateur
	var user models.User
	if err := h.db.Where("id = ? AND role = ?", userID, models.RoleCreator).First(&user).Error; err != nil {
		utils.SendError(w, http.StatusForbidden, "Seuls les créateurs peuvent demander des versements", err)
		return
	}

	// Vérifier que le créateur a suffisamment de fonds disponibles
	// (Ici vous devriez implémenter la logique pour calculer le solde disponible)

	// Créer la demande de versement
	payout := &models.Payout{
		CreatorID: userID,
		Amount:    req.Amount,
		Currency:  "EUR",
		Status:    "pending",
		Reference: "PAYOUT-" + userID + "-" + strconv.FormatInt(utils.GenerateTimestamp(), 10),
	}

	if err := h.db.Create(payout).Error; err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors de la création de la demande de versement", err)
		return
	}

	utils.SendJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"payout":  payout,
		"message": "Demande de versement créée avec succès",
	})
}

// GetPayouts récupère l'historique des versements d'un créateur
func (h *PaymentHandler) GetPayouts(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	// Vérifier que l'utilisateur est un créateur
	var user models.User
	if err := h.db.Where("id = ? AND role = ?", userID, models.RoleCreator).First(&user).Error; err != nil {
		utils.SendError(w, http.StatusForbidden, "Seuls les créateurs peuvent voir leurs versements", err)
		return
	}

	// Paramètres de pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var payouts []models.Payout
	var total int64

	// Compter le total
	h.db.Model(&models.Payout{}).Where("creator_id = ?", userID).Count(&total)

	// Récupérer les versements avec pagination
	offset := (page - 1) * limit
	if err := h.db.Where("creator_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&payouts).Error; err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors de la récupération des versements", err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"payouts": payouts,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetPayout récupère les détails d'un versement spécifique
func (h *PaymentHandler) GetPayout(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	vars := mux.Vars(r)
	payoutID := vars["id"]

	var payout models.Payout
	if err := h.db.Where("id = ? AND creator_id = ?", payoutID, userID).First(&payout).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Versement non trouvé", err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"payout": payout,
	})
}

// GetEarnings récupère les revenus d'un créateur
func (h *PaymentHandler) GetEarnings(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	// Vérifier que l'utilisateur est un créateur
	var user models.User
	if err := h.db.Where("id = ? AND role = ?", userID, models.RoleCreator).First(&user).Error; err != nil {
		utils.SendError(w, http.StatusForbidden, "Seuls les créateurs peuvent voir leurs revenus", err)
		return
	}

	// Calculer les revenus totaux
	var totalEarnings float64
	h.db.Table("transactions").
		Joins("JOIN subscriptions ON transactions.subscription_id = subscriptions.id").
		Where("subscriptions.creator_id = ? AND transactions.status = ?", userID, "success").
		Select("COALESCE(SUM(transactions.amount), 0)").
		Scan(&totalEarnings)

	// Calculer les revenus du mois en cours
	var monthlyEarnings float64
	h.db.Table("transactions").
		Joins("JOIN subscriptions ON transactions.subscription_id = subscriptions.id").
		Where("subscriptions.creator_id = ? AND transactions.status = ? AND DATE_PART('month', transactions.created_at) = DATE_PART('month', CURRENT_DATE) AND DATE_PART('year', transactions.created_at) = DATE_PART('year', CURRENT_DATE)", userID, "success").
		Select("COALESCE(SUM(transactions.amount), 0)").
		Scan(&monthlyEarnings)

	// Calculer les versements effectués
	var totalPayouts float64
	h.db.Model(&models.Payout{}).
		Where("creator_id = ? AND status = ?", userID, "processed").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalPayouts)

	// Solde disponible
	availableBalance := totalEarnings - totalPayouts

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"total_earnings":    totalEarnings,
		"monthly_earnings":  monthlyEarnings,
		"total_payouts":     totalPayouts,
		"available_balance": availableBalance,
	})
}

// GetCreatorEarnings récupère les revenus d'un créateur spécifique (pour l'endpoint /creators/{id}/earnings)
func (h *PaymentHandler) GetCreatorEarnings(w http.ResponseWriter, r *http.Request) {
	currentUserID := r.Context().Value("user_id").(string)
	vars := mux.Vars(r)
	creatorID := vars["id"]

	// Vérifier que l'utilisateur demande ses propres revenus ou est admin
	var user models.User
	if err := h.db.Where("id = ?", currentUserID).First(&user).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Utilisateur non trouvé", err)
		return
	}

	// Seul le créateur lui-même ou un admin peut voir ses revenus
	if creatorID != currentUserID && user.Role != models.RoleAdmin {
		utils.SendError(w, http.StatusForbidden, "Accès refusé", nil)
		return
	}

	// Vérifier que le créateur existe et a le bon rôle
	var creator models.User
	if err := h.db.Where("id = ? AND role = ?", creatorID, models.RoleCreator).First(&creator).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Créateur non trouvé", err)
		return
	}

	// Calculer les revenus totaux
	var totalEarnings float64
	h.db.Table("transactions").
		Joins("JOIN subscriptions ON transactions.subscription_id = subscriptions.id").
		Where("subscriptions.creator_id = ? AND transactions.status = ?", creatorID, "success").
		Select("COALESCE(SUM(transactions.amount), 0)").
		Scan(&totalEarnings)

	// Calculer les revenus du mois en cours
	var monthlyEarnings float64
	h.db.Table("transactions").
		Joins("JOIN subscriptions ON transactions.subscription_id = subscriptions.id").
		Where("subscriptions.creator_id = ? AND transactions.status = ? AND DATE_PART('month', transactions.created_at) = DATE_PART('month', CURRENT_DATE) AND DATE_PART('year', transactions.created_at) = DATE_PART('year', CURRENT_DATE)", creatorID, "success").
		Select("COALESCE(SUM(transactions.amount), 0)").
		Scan(&monthlyEarnings)

	// Calculer les versements effectués
	var totalPayouts float64
	h.db.Model(&models.Payout{}).
		Where("creator_id = ? AND status = ?", creatorID, "processed").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalPayouts)

	// Solde disponible
	availableBalance := totalEarnings - totalPayouts

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"creator_id":        creatorID,
		"total_earnings":    totalEarnings,
		"monthly_earnings":  monthlyEarnings,
		"total_payouts":     totalPayouts,
		"available_balance": availableBalance,
	})
}

// CreateSetupIntent crée un SetupIntent Stripe pour configurer une méthode de paiement
func (h *PaymentHandler) CreateSetupIntent(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	// Récupérer l'utilisateur
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Utilisateur non trouvé", err)
		return
	}

	// Créer ou récupérer le client Stripe
	var stripeCustomerID string
	if user.Email != "" {
		// Vérifier si l'utilisateur a déjà un client Stripe (vous pourriez stocker cela dans la base)
		// Pour l'instant, créons toujours un nouveau client
		customer, err := h.stripeService.CreateCustomer(user.Email, user.FirstName+" "+user.LastName)
		if err != nil {
			utils.SendError(w, http.StatusInternalServerError, "Erreur lors de la création du client Stripe", err)
			return
		}
		stripeCustomerID = customer.ID
	}

	// Créer le SetupIntent
	setupIntent, err := h.stripeService.CreateSetupIntent(stripeCustomerID)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors de la création du SetupIntent", err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"client_secret":   setupIntent.ClientSecret,
		"setup_intent_id": setupIntent.ID,
	})
}

// SubscriptionPaymentRequest représente une demande de paiement d'abonnement
type SubscriptionPaymentRequest struct {
	SubscriptionPlanID uint   `json:"subscription_plan_id" validate:"required"`
	PaymentMethodID    string `json:"payment_method_id" validate:"required"`
}

// ProcessSubscriptionPayment traite un paiement d'abonnement
func (h *PaymentHandler) ProcessSubscriptionPayment(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req SubscriptionPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Données JSON invalides", err)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Données de validation invalides", err)
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Utilisateur non trouvé", err)
		return
	}

	// Récupérer le plan d'abonnement
	var plan models.SubscriptionPlan
	if err := h.db.Where("id = ? AND is_active = ?", req.SubscriptionPlanID, true).First(&plan).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Plan d'abonnement non trouvé", err)
		return
	}

	// Vérifier que l'utilisateur n'est pas déjà abonné à ce créateur
	var existingSubscription models.Subscription
	if err := h.db.Where("subscriber_id = ? AND creator_id = ? AND is_active = ?", userID, plan.CreatorID, true).First(&existingSubscription).Error; err == nil {
		utils.SendError(w, http.StatusConflict, "Vous êtes déjà abonné à ce créateur", nil)
		return
	}

	// Récupérer la méthode de paiement
	var paymentMethod models.PaymentMethod
	if err := h.db.Where("id = ? AND user_id = ?", req.PaymentMethodID, userID).First(&paymentMethod).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Méthode de paiement non trouvée", err)
		return
	}

	// Créer le PaymentIntent Stripe
	amount := h.stripeService.FormatAmountForStripe(plan.Price)
	metadata := map[string]string{
		"user_id":    userID,
		"plan_id":    strconv.Itoa(int(plan.ID)),
		"creator_id": plan.CreatorID,
	}

	// Pour traiter le paiement, nous aurions besoin du customer ID Stripe
	// Pour l'instant, créons un client temporaire
	customer, err := h.stripeService.CreateCustomer(user.Email, user.FirstName+" "+user.LastName)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors de la création du client", err)
		return
	}

	paymentIntent, err := h.stripeService.CreatePaymentIntent(
		amount,
		"eur",
		customer.ID,
		paymentMethod.StripeMethodID,
		metadata,
	)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors du traitement du paiement", err)
		return
	}

	// Si le paiement est réussi, créer l'abonnement
	if paymentIntent.Status == "succeeded" {
		// Calculer les dates d'abonnement
		startDate := time.Now()
		endDate := startDate.AddDate(0, 0, plan.Duration)

		subscription := &models.Subscription{
			SubscriberID:  userID,
			CreatorID:     plan.CreatorID,
			PlanID:        plan.ID,
			StartDate:     startDate,
			EndDate:       endDate,
			IsActive:      true,
			AutoRenew:     true,
			PaymentStatus: "paid",
			TransactionID: paymentIntent.ID,
		}

		if err := h.db.Create(subscription).Error; err != nil {
			utils.SendError(w, http.StatusInternalServerError, "Erreur lors de la création de l'abonnement", err)
			return
		}

		// Créer la transaction
		transaction := &models.Transaction{
			UserID:         userID,
			SubscriptionID: &subscription.ID,
			Amount:         plan.Price,
			Currency:       "EUR",
			Status:         "success",
			PaymentMethod:  paymentMethod.Type,
			PaymentID:      paymentIntent.ID,
			Description:    "Abonnement - " + plan.Name,
		}

		if err := h.db.Create(transaction).Error; err != nil {
			utils.SendError(w, http.StatusInternalServerError, "Erreur lors de l'enregistrement de la transaction", err)
			return
		}

		utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"subscription": subscription,
			"transaction":  transaction,
			"message":      "Abonnement créé avec succès",
		})
	} else {
		utils.SendError(w, http.StatusPaymentRequired, "Le paiement a échoué", nil)
	}
}

// WebhookEvent représente un événement webhook traité
type WebhookEvent struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		Object map[string]interface{} `json:"object"`
	} `json:"data"`
}

// HandleStripeWebhook gère les webhooks Stripe de manière plus robuste
func (h *PaymentHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()

	// Lire le body de la requête
	body := make([]byte, r.ContentLength)
	if _, err := r.Body.Read(body); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Erreur lors de la lecture du body", err)
		return
	}

	// Vérifier la signature du webhook
	signature := r.Header.Get("Stripe-Signature")
	event, err := h.stripeService.ConstructWebhookEvent(body, signature, cfg.Stripe.WebhookSecret)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Signature invalide", err)
		return
	}

	// Traiter les différents types d'événements
	switch event.Type {
	case "payment_intent.succeeded":
		h.handlePaymentIntentSucceeded(event)
	case "payment_intent.payment_failed":
		h.handlePaymentIntentFailed(event)
	case "customer.subscription.created":
		h.handleSubscriptionCreated(event)
	case "customer.subscription.updated":
		h.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		h.handleSubscriptionDeleted(event)
	case "invoice.payment_succeeded":
		h.handleInvoicePaymentSucceeded(event)
	case "invoice.payment_failed":
		h.handleInvoicePaymentFailed(event)
	case "customer.subscription.trial_will_end":
		h.handleTrialWillEnd(event)
	default:
		// Événement non géré, mais on confirme la réception
		fmt.Printf("Événement webhook non géré: %s\n", event.Type)
	}

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":    "Webhook traité avec succès",
		"event_type": event.Type,
	})
}

// handlePaymentIntentSucceeded traite les paiements réussis
func (h *PaymentHandler) handlePaymentIntentSucceeded(event stripe.Event) {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		fmt.Printf("Erreur lors du parsing du PaymentIntent: %v\n", err)
		return
	}

	// Mettre à jour la transaction en base
	if err := h.db.Model(&models.Transaction{}).
		Where("payment_id = ?", paymentIntent.ID).
		Update("status", "success").Error; err != nil {
		fmt.Printf("Erreur lors de la mise à jour de la transaction: %v\n", err)
	}

	// Activer l'abonnement si nécessaire
	if subscriptionID, exists := paymentIntent.Metadata["subscription_id"]; exists {
		if err := h.db.Model(&models.Subscription{}).
			Where("id = ?", subscriptionID).
			Updates(map[string]interface{}{
				"is_active":      true,
				"payment_status": "paid",
			}).Error; err != nil {
			fmt.Printf("Erreur lors de l'activation de l'abonnement: %v\n", err)
		}
	}
}

// handlePaymentIntentFailed traite les paiements échoués
func (h *PaymentHandler) handlePaymentIntentFailed(event stripe.Event) {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		fmt.Printf("Erreur lors du parsing du PaymentIntent: %v\n", err)
		return
	}

	// Mettre à jour la transaction en base
	if err := h.db.Model(&models.Transaction{}).
		Where("payment_id = ?", paymentIntent.ID).
		Update("status", "failed").Error; err != nil {
		fmt.Printf("Erreur lors de la mise à jour de la transaction: %v\n", err)
	}

	// Désactiver l'abonnement si nécessaire
	if subscriptionID, exists := paymentIntent.Metadata["subscription_id"]; exists {
		if err := h.db.Model(&models.Subscription{}).
			Where("id = ?", subscriptionID).
			Updates(map[string]interface{}{
				"is_active":      false,
				"payment_status": "failed",
			}).Error; err != nil {
			fmt.Printf("Erreur lors de la désactivation de l'abonnement: %v\n", err)
		}
	}
}

// handleSubscriptionCreated traite la création d'abonnements
func (h *PaymentHandler) handleSubscriptionCreated(event stripe.Event) {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		fmt.Printf("Erreur lors du parsing de l'abonnement: %v\n", err)
		return
	}

	// Logique pour synchroniser l'abonnement avec la base de données
	fmt.Printf("Abonnement créé: %s\n", subscription.ID)
}

// handleSubscriptionUpdated traite les mises à jour d'abonnements
func (h *PaymentHandler) handleSubscriptionUpdated(event stripe.Event) {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		fmt.Printf("Erreur lors du parsing de l'abonnement: %v\n", err)
		return
	}

	// Logique pour mettre à jour l'abonnement en base
	fmt.Printf("Abonnement mis à jour: %s\n", subscription.ID)
}

// handleSubscriptionDeleted traite la suppression d'abonnements
func (h *PaymentHandler) handleSubscriptionDeleted(event stripe.Event) {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		fmt.Printf("Erreur lors du parsing de l'abonnement: %v\n", err)
		return
	}

	// Désactiver l'abonnement en base
	if userID, exists := subscription.Metadata["user_id"]; exists {
		if err := h.db.Model(&models.Subscription{}).
			Where("transaction_id = ? AND subscriber_id = ?", subscription.ID, userID).
			Update("is_active", false).Error; err != nil {
			fmt.Printf("Erreur lors de la désactivation de l'abonnement: %v\n", err)
		}
	}
}

// handleInvoicePaymentSucceeded traite les paiements de factures réussis
func (h *PaymentHandler) handleInvoicePaymentSucceeded(event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		fmt.Printf("Erreur lors du parsing de la facture: %v\n", err)
		return
	}

	// Enregistrer la transaction de renouvellement
	fmt.Printf("Paiement de facture réussi: %s\n", invoice.ID)
}

// handleInvoicePaymentFailed traite les paiements de factures échoués
func (h *PaymentHandler) handleInvoicePaymentFailed(event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		fmt.Printf("Erreur lors du parsing de la facture: %v\n", err)
		return
	}

	// Gérer l'échec du paiement (notification, suspension, etc.)
	fmt.Printf("Paiement de facture échoué: %s\n", invoice.ID)
}

// handleTrialWillEnd traite les fins de période d'essai
func (h *PaymentHandler) handleTrialWillEnd(event stripe.Event) {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		fmt.Printf("Erreur lors du parsing de l'abonnement: %v\n", err)
		return
	}

	// Envoyer une notification à l'utilisateur
	fmt.Printf("Fin de période d'essai approchant pour: %s\n", subscription.ID)
}

// Utilitaires pour la gestion des paiements

// ValidatePaymentAmount valide qu'un montant est correct pour un paiement
func (h *PaymentHandler) ValidatePaymentAmount(amount float64) error {
	if amount <= 0 {
		return errors.New("le montant doit être positif")
	}

	if amount < 1.00 {
		return errors.New("le montant minimum est de 1.00€")
	}

	if amount > 10000.00 {
		return errors.New("le montant maximum est de 10,000.00€")
	}

	return nil
}

// CheckUserPaymentMethodOwnership vérifie qu'une méthode de paiement appartient à l'utilisateur
func (h *PaymentHandler) CheckUserPaymentMethodOwnership(userID, paymentMethodID string) (*models.PaymentMethod, error) {
	var paymentMethod models.PaymentMethod
	if err := h.db.Where("id = ? AND user_id = ?", paymentMethodID, userID).First(&paymentMethod).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("méthode de paiement non trouvée ou non autorisée")
		}
		return nil, err
	}
	return &paymentMethod, nil
}

// CheckCreatorEligibility vérifie qu'un utilisateur peut recevoir des paiements
func (h *PaymentHandler) CheckCreatorEligibility(userID string) (*models.User, error) {
	var user models.User
	if err := h.db.Where("id = ? AND role = ? AND is_active = ? AND is_banned = ?", userID, models.RoleCreator, true, false).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("créateur non trouvé ou non éligible")
		}
		return nil, err
	}
	return &user, nil
}

// CalculateCreatorAvailableBalance calcule le solde disponible d'un créateur
func (h *PaymentHandler) CalculateCreatorAvailableBalance(creatorID string) (float64, error) {
	// Calculer les revenus totaux
	var totalEarnings float64
	if err := h.db.Table("transactions").
		Joins("JOIN subscriptions ON transactions.subscription_id = subscriptions.id").
		Where("subscriptions.creator_id = ? AND transactions.status = ?", creatorID, "success").
		Select("COALESCE(SUM(transactions.amount), 0)").
		Scan(&totalEarnings).Error; err != nil {
		return 0, err
	}

	// Calculer les versements effectués
	var totalPayouts float64
	if err := h.db.Model(&models.Payout{}).
		Where("creator_id = ? AND status IN (?)", creatorID, []string{"processed", "pending"}).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalPayouts).Error; err != nil {
		return 0, err
	}

	return totalEarnings - totalPayouts, nil
}

// CreatePaymentTransaction crée une transaction de paiement en base
func (h *PaymentHandler) CreatePaymentTransaction(userID string, subscriptionID *uint, amount float64, paymentID, description string, status string) (*models.Transaction, error) {
	transaction := &models.Transaction{
		UserID:         userID,
		SubscriptionID: subscriptionID,
		Amount:         amount,
		Currency:       "EUR",
		Status:         status,
		PaymentMethod:  "card", // Par défaut, pourrait être paramétré
		PaymentID:      paymentID,
		Description:    description,
	}

	if err := h.db.Create(transaction).Error; err != nil {
		return nil, err
	}

	return transaction, nil
}

// SendPaymentNotification envoie une notification liée aux paiements
func (h *PaymentHandler) SendPaymentNotification(userID, notificationType, message string, relatedID uint) error {
	notification := &models.Notification{
		UserID:    userID,
		Type:      notificationType,
		Message:   message,
		RelatedID: relatedID,
		IsRead:    false,
	}

	return h.db.Create(notification).Error
}

// FormatCurrency formate un montant en devise
func (h *PaymentHandler) FormatCurrency(amount float64, currency string) string {
	switch currency {
	case "EUR":
		return fmt.Sprintf("%.2f€", amount)
	case "USD":
		return fmt.Sprintf("$%.2f", amount)
	default:
		return fmt.Sprintf("%.2f %s", amount, currency)
	}
}

// GetPaymentMethodSummary retourne un résumé sécurisé d'une méthode de paiement
type PaymentMethodSummary struct {
	ID          uint   `json:"id"`
	Type        string `json:"type"`
	Last4       string `json:"last4"`
	Brand       string `json:"brand"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
	IsDefault   bool   `json:"is_default"`
}

func (h *PaymentHandler) GetPaymentMethodSummary(pm *models.PaymentMethod) PaymentMethodSummary {
	return PaymentMethodSummary{
		ID:          pm.ID,
		Type:        pm.Type,
		Last4:       pm.Last4,
		Brand:       pm.Brand,
		ExpiryMonth: pm.ExpiryMonth,
		ExpiryYear:  pm.ExpiryYear,
		IsDefault:   pm.IsDefault,
	}
}
