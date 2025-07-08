package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
	"github.com/Kaowarstail/Only-Flick-Go/models"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// SubscriptionHandler gère les endpoints des abonnements et plans d'abonnement
type SubscriptionHandler struct {
	db *gorm.DB
}

// NewSubscriptionHandler crée une nouvelle instance du handler d'abonnement
func NewSubscriptionHandler(db *gorm.DB) *SubscriptionHandler {
	return &SubscriptionHandler{
		db: db,
	}
}

// CreateSubscriptionPlanRequest représente une demande de création de plan d'abonnement
type CreateSubscriptionPlanRequest struct {
	Name        string  `json:"name" validate:"required,min=3,max=100"`
	Description string  `json:"description" validate:"max=500"`
	Price       float64 `json:"price" validate:"required,min=1"`
	Duration    int     `json:"duration" validate:"required,min=1,max=365"` // En jours
}

// UpdateSubscriptionPlanRequest représente une demande de mise à jour de plan d'abonnement
type UpdateSubscriptionPlanRequest struct {
	Name        *string  `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=500"`
	Price       *float64 `json:"price,omitempty" validate:"omitempty,min=1"`
	Duration    *int     `json:"duration,omitempty" validate:"omitempty,min=1,max=365"`
	IsActive    *bool    `json:"is_active,omitempty"`
}

// CreateSubscriptionPlan crée un nouveau plan d'abonnement (Créateur uniquement)
func (h *SubscriptionHandler) CreateSubscriptionPlan(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	// Vérifier que l'utilisateur est un créateur
	var user models.User
	if err := h.db.Where("id = ? AND role = ?", userID, models.RoleCreator).First(&user).Error; err != nil {
		utils.SendError(w, http.StatusForbidden, "Seuls les créateurs peuvent créer des plans d'abonnement", err)
		return
	}

	var req CreateSubscriptionPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Données JSON invalides", err)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Données de validation invalides", err)
		return
	}

	// Créer le plan d'abonnement
	plan := &models.SubscriptionPlan{
		CreatorID:   userID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Duration:    req.Duration,
		IsActive:    true,
	}

	if err := h.db.Create(plan).Error; err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors de la création du plan d'abonnement", err)
		return
	}

	// Charger le créateur pour la réponse
	h.db.Preload("Creator").First(plan, plan.ID)

	utils.SendJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"subscription_plan": plan,
		"message":           "Plan d'abonnement créé avec succès",
	})
}

// GetSubscriptionPlans récupère tous les plans d'abonnement actifs (Public)
func (h *SubscriptionHandler) GetSubscriptionPlans(w http.ResponseWriter, r *http.Request) {
	// Paramètres de pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Paramètres de filtrage
	creatorID := r.URL.Query().Get("creator_id")
	minPrice := r.URL.Query().Get("min_price")
	maxPrice := r.URL.Query().Get("max_price")

	var plans []models.SubscriptionPlan
	var total int64

	query := h.db.Model(&models.SubscriptionPlan{}).Where("is_active = ?", true)

	// Filtres
	if creatorID != "" {
		query = query.Where("creator_id = ?", creatorID)
	}
	if minPrice != "" {
		if price, err := strconv.ParseFloat(minPrice, 64); err == nil {
			query = query.Where("price >= ?", price)
		}
	}
	if maxPrice != "" {
		if price, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			query = query.Where("price <= ?", price)
		}
	}

	// Compter le total
	query.Count(&total)

	// Récupérer les plans avec pagination
	offset := (page - 1) * limit
	if err := query.Preload("Creator").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&plans).Error; err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors de la récupération des plans d'abonnement", err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"subscription_plans": plans,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetSubscriptionPlan récupère un plan d'abonnement spécifique (Public)
func (h *SubscriptionHandler) GetSubscriptionPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["id"]

	var plan models.SubscriptionPlan
	if err := h.db.Where("id = ? AND is_active = ?", planID, true).
		Preload("Creator").
		First(&plan).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendError(w, http.StatusNotFound, "Plan d'abonnement non trouvé", err)
		} else {
			utils.SendError(w, http.StatusInternalServerError, "Erreur lors de la récupération du plan", err)
		}
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"subscription_plan": plan,
	})
}

// UpdateSubscriptionPlan met à jour un plan d'abonnement (Propriétaire/Admin)
func (h *SubscriptionHandler) UpdateSubscriptionPlan(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	vars := mux.Vars(r)
	planID := vars["id"]

	// Récupérer le plan existant
	var plan models.SubscriptionPlan
	if err := h.db.Where("id = ?", planID).First(&plan).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Plan d'abonnement non trouvé", err)
		return
	}

	// Vérifier les permissions (propriétaire ou admin)
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Utilisateur non trouvé", err)
		return
	}

	if plan.CreatorID != userID && user.Role != models.RoleAdmin {
		utils.SendError(w, http.StatusForbidden, "Vous n'avez pas l'autorisation de modifier ce plan", nil)
		return
	}

	var req UpdateSubscriptionPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Données JSON invalides", err)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Données de validation invalides", err)
		return
	}

	// Mettre à jour les champs modifiés
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.Duration != nil {
		updates["duration"] = *req.Duration
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		utils.SendError(w, http.StatusBadRequest, "Aucune donnée à mettre à jour", nil)
		return
	}

	if err := h.db.Model(&plan).Updates(updates).Error; err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors de la mise à jour du plan", err)
		return
	}

	// Recharger le plan mis à jour
	h.db.Preload("Creator").First(&plan, plan.ID)

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"subscription_plan": plan,
		"message":           "Plan d'abonnement mis à jour avec succès",
	})
}

// DeleteSubscriptionPlan supprime un plan d'abonnement (Propriétaire/Admin)
func (h *SubscriptionHandler) DeleteSubscriptionPlan(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	vars := mux.Vars(r)
	planID := vars["id"]

	// Récupérer le plan existant
	var plan models.SubscriptionPlan
	if err := h.db.Where("id = ?", planID).First(&plan).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Plan d'abonnement non trouvé", err)
		return
	}

	// Vérifier les permissions (propriétaire ou admin)
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Utilisateur non trouvé", err)
		return
	}

	if plan.CreatorID != userID && user.Role != models.RoleAdmin {
		utils.SendError(w, http.StatusForbidden, "Vous n'avez pas l'autorisation de supprimer ce plan", nil)
		return
	}

	// Vérifier s'il y a des abonnements actifs
	var activeSubscriptions int64
	h.db.Model(&models.Subscription{}).
		Where("plan_id = ? AND is_active = ?", planID, true).
		Count(&activeSubscriptions)

	if activeSubscriptions > 0 {
		utils.SendError(w, http.StatusConflict, "Impossible de supprimer un plan avec des abonnements actifs", nil)
		return
	}

	// Supprimer le plan (soft delete)
	if err := h.db.Delete(&plan).Error; err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors de la suppression du plan", err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Plan d'abonnement supprimé avec succès",
	})
}

// GetCreatorSubscriptionPlans récupère les plans d'abonnement d'un créateur (Public)
func (h *SubscriptionHandler) GetCreatorSubscriptionPlans(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	creatorID := vars["id"]

	// Vérifier que le créateur existe
	var creator models.User
	if err := h.db.Where("id = ? AND role = ?", creatorID, models.RoleCreator).First(&creator).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Créateur non trouvé", err)
		return
	}

	var plans []models.SubscriptionPlan
	if err := h.db.Where("creator_id = ? AND is_active = ?", creatorID, true).
		Order("price ASC").
		Find(&plans).Error; err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors de la récupération des plans", err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"creator":            creator.ToResponse(),
		"subscription_plans": plans,
	})
}

// CreateSubscriptionRequest représente une demande de création d'abonnement
type CreateSubscriptionRequest struct {
	PlanID          uint   `json:"plan_id" validate:"required"`
	PaymentMethodID string `json:"payment_method_id" validate:"required"`
}

// CreateSubscription crée un nouvel abonnement (Authentifié)
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Données JSON invalides", err)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Données de validation invalides", err)
		return
	}

	// Récupérer le plan d'abonnement
	var plan models.SubscriptionPlan
	if err := h.db.Where("id = ? AND is_active = ?", req.PlanID, true).First(&plan).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Plan d'abonnement non trouvé", err)
		return
	}

	// Vérifier que l'utilisateur n'est pas déjà abonné à ce créateur
	var existingSubscription models.Subscription
	if err := h.db.Where("subscriber_id = ? AND creator_id = ? AND is_active = ?",
		userID, plan.CreatorID, true).First(&existingSubscription).Error; err == nil {
		utils.SendError(w, http.StatusConflict, "Vous êtes déjà abonné à ce créateur", nil)
		return
	}

	// Calculer les dates d'abonnement
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, plan.Duration)

	// Créer l'abonnement (statut pending, sera activé après paiement)
	subscription := &models.Subscription{
		SubscriberID:  userID,
		CreatorID:     plan.CreatorID,
		PlanID:        plan.ID,
		StartDate:     startDate,
		EndDate:       endDate,
		IsActive:      false, // Sera activé après paiement
		AutoRenew:     true,
		PaymentStatus: "pending",
	}

	if err := h.db.Create(subscription).Error; err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors de la création de l'abonnement", err)
		return
	}

	// Charger les relations pour la réponse
	h.db.Preload("Plan").Preload("Creator").Preload("Subscriber").First(subscription, subscription.ID)

	utils.SendJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"subscription": subscription,
		"message":      "Abonnement créé avec succès. Procédez au paiement pour l'activer.",
		"next_step":    "Utilisez l'endpoint de paiement pour finaliser votre abonnement",
	})
}

// GetUserSubscriptions récupère les abonnements de l'utilisateur (Authentifié)
func (h *SubscriptionHandler) GetUserSubscriptions(w http.ResponseWriter, r *http.Request) {
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

	// Filtre de statut
	status := r.URL.Query().Get("status") // active, expired, cancelled

	var subscriptions []models.Subscription
	var total int64

	query := h.db.Model(&models.Subscription{}).Where("subscriber_id = ?", userID)

	// Filtres
	if status == "active" {
		query = query.Where("is_active = ? AND end_date > ?", true, time.Now())
	} else if status == "expired" {
		query = query.Where("end_date <= ?", time.Now())
	} else if status == "cancelled" {
		query = query.Where("is_active = ?", false)
	}

	// Compter le total
	query.Count(&total)

	// Récupérer les abonnements avec pagination
	offset := (page - 1) * limit
	if err := query.Preload("Plan").Preload("Creator").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&subscriptions).Error; err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors de la récupération des abonnements", err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"subscriptions": subscriptions,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetSubscription récupère un abonnement spécifique (Propriétaire/Admin)
func (h *SubscriptionHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	vars := mux.Vars(r)
	subscriptionID := vars["id"]

	var subscription models.Subscription
	if err := h.db.Where("id = ?", subscriptionID).
		Preload("Plan").Preload("Creator").Preload("Subscriber").
		First(&subscription).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Abonnement non trouvé", err)
		return
	}

	// Vérifier les permissions (abonné, créateur ou admin)
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Utilisateur non trouvé", err)
		return
	}

	if subscription.SubscriberID != userID &&
		subscription.CreatorID != userID &&
		user.Role != models.RoleAdmin {
		utils.SendError(w, http.StatusForbidden, "Vous n'avez pas l'autorisation de voir cet abonnement", nil)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"subscription": subscription,
	})
}

// CancelSubscription annule un abonnement (Propriétaire/Admin)
func (h *SubscriptionHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	vars := mux.Vars(r)
	subscriptionID := vars["id"]

	var subscription models.Subscription
	if err := h.db.Where("id = ?", subscriptionID).First(&subscription).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Abonnement non trouvé", err)
		return
	}

	// Vérifier les permissions (abonné ou admin)
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Utilisateur non trouvé", err)
		return
	}

	if subscription.SubscriberID != userID && user.Role != models.RoleAdmin {
		utils.SendError(w, http.StatusForbidden, "Vous n'avez pas l'autorisation d'annuler cet abonnement", nil)
		return
	}

	// Annuler l'abonnement
	if err := h.db.Model(&subscription).Updates(map[string]interface{}{
		"is_active":  false,
		"auto_renew": false,
		"updated_at": time.Now(),
	}).Error; err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors de l'annulation de l'abonnement", err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Abonnement annulé avec succès",
	})
}

// RenewSubscription renouvelle un abonnement (Propriétaire)
func (h *SubscriptionHandler) RenewSubscription(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	vars := mux.Vars(r)
	subscriptionID := vars["id"]

	var subscription models.Subscription
	if err := h.db.Where("id = ?", subscriptionID).
		Preload("Plan").
		First(&subscription).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "Abonnement non trouvé", err)
		return
	}

	// Vérifier les permissions (seul l'abonné peut renouveler)
	if subscription.SubscriberID != userID {
		utils.SendError(w, http.StatusForbidden, "Vous n'avez pas l'autorisation de renouveler cet abonnement", nil)
		return
	}

	// Vérifier que l'abonnement peut être renouvelé
	if subscription.IsActive && subscription.EndDate.After(time.Now()) {
		utils.SendError(w, http.StatusBadRequest, "L'abonnement est encore actif", nil)
		return
	}

	// Calculer les nouvelles dates
	startDate := time.Now()
	if subscription.EndDate.After(time.Now()) {
		// Si l'abonnement n'est pas encore expiré, commencer à la date d'expiration
		startDate = subscription.EndDate
	}
	endDate := startDate.AddDate(0, 0, subscription.Plan.Duration)

	// Renouveler l'abonnement
	if err := h.db.Model(&subscription).Updates(map[string]interface{}{
		"start_date":     startDate,
		"end_date":       endDate,
		"is_active":      true,
		"auto_renew":     true,
		"payment_status": "pending", // Sera mis à jour après paiement
		"updated_at":     time.Now(),
	}).Error; err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Erreur lors du renouvellement de l'abonnement", err)
		return
	}

	// Recharger l'abonnement avec les relations
	h.db.Preload("Plan").Preload("Creator").Preload("Subscriber").First(&subscription, subscription.ID)

	utils.SendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"subscription": subscription,
		"message":      "Abonnement renouvelé avec succès. Procédez au paiement pour l'activer.",
		"next_step":    "Utilisez l'endpoint de paiement pour finaliser le renouvellement",
	})
}
