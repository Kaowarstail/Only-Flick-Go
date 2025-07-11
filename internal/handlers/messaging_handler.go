package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/Kaowarstail/Only-Flick-Go/internal/services"
	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
	"github.com/Kaowarstail/Only-Flick-Go/models"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// MessagingHandler gère les endpoints de messagerie
type MessagingHandler struct {
	messagingManager *services.MessagingServiceManager
}

// NewMessagingHandler crée une nouvelle instance
func NewMessagingHandler(db *gorm.DB) *MessagingHandler {
	return &MessagingHandler{
		messagingManager: services.NewMessagingServiceManager(db),
	}
}

// ========== Conversation Endpoints ==========

// GetUserConversations récupère les conversations d'un utilisateur
// GET /api/conversations?page=1&limit=20
func (h *MessagingHandler) GetUserConversations(w http.ResponseWriter, r *http.Request) {
	// Extraire l'ID utilisateur du contexte (défini par le middleware JWTAuth)
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
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

	// Appeler le service
	conversationsResponse, err := h.messagingManager.GetMessagingService().GetUserConversations(userID, page, limit)
	if err != nil {
		if serviceErr, ok := services.IsServiceError(err); ok {
			respondWithError(w, serviceErr.HTTPStatus, serviceErr.Message)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des conversations")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, conversationsResponse)
}

// CreateOrGetConversation crée ou récupère une conversation directe
// POST /api/conversations
func (h *MessagingHandler) CreateOrGetConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	var request models.CreateConversationRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}
	defer r.Body.Close()

	// Validation
	if err := request.Validate(); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Nettoyer les entrées
	request.OtherUserID = utils.SanitizeInput(request.OtherUserID)

	// Appeler le service
	conversation, err := h.messagingManager.GetMessagingService().GetOrCreateDirectConversation(userID, request.OtherUserID)
	if err != nil {
		if serviceErr, ok := services.IsServiceError(err); ok {
			respondWithError(w, serviceErr.HTTPStatus, serviceErr.Message)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Erreur lors de la création de la conversation")
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"conversation": conversation,
		"message":      "Conversation créée avec succès",
	})
}

// GetConversationMessages récupère les messages d'une conversation
// GET /api/conversations/{conversationId}/messages?page=1&limit=50
func (h *MessagingHandler) GetConversationMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Extraire l'ID de conversation
	vars := mux.Vars(r)
	conversationID := vars["conversationId"]

	// Paramètres de pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// Appeler le service
	messagesResponse, err := h.messagingManager.GetMessagingService().GetConversationMessages(conversationID, userID, page, limit)
	if err != nil {
		if serviceErr, ok := services.IsServiceError(err); ok {
			respondWithError(w, serviceErr.HTTPStatus, serviceErr.Message)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des messages")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, messagesResponse)
}

// MarkConversationAsRead marque une conversation comme lue
// PUT /api/conversations/{conversationId}/read
func (h *MessagingHandler) MarkConversationAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Extraire l'ID de conversation
	vars := mux.Vars(r)
	conversationID := vars["conversationId"]

	// Appeler le service
	err := h.messagingManager.GetMessagingService().MarkConversationAsRead(conversationID, userID)
	if err != nil {
		if serviceErr, ok := services.IsServiceError(err); ok {
			respondWithError(w, serviceErr.HTTPStatus, serviceErr.Message)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Erreur lors du marquage comme lu")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Conversation marquée comme lue",
	})
}

// ========== Message Endpoints ==========

// SendMessage envoie un message
// POST /api/messages
func (h *MessagingHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	var request services.SendMessageRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}
	defer r.Body.Close()

	// Validation
	if err := request.Validate(); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Nettoyer le contenu si présent
	if request.Content != nil {
		cleaned := utils.CleanMessageContent(*request.Content)
		request.Content = &cleaned
	}

	// Appeler le service
	message, err := h.messagingManager.GetMessagingService().SendMessage(&request, userID)
	if err != nil {
		if serviceErr, ok := services.IsServiceError(err); ok {
			respondWithError(w, serviceErr.HTTPStatus, serviceErr.Message)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'envoi du message")
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message": message,
		"success": "Message envoyé avec succès",
	})
}

// ========== Dashboard & Stats ==========

// GetMessagingDashboard récupère le tableau de bord de messagerie
// GET /api/messaging/dashboard?page=1&limit=20
func (h *MessagingHandler) GetMessagingDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
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

	// Appeler le service
	dashboard, err := h.messagingManager.GetMessagingService().GetUserMessagingDashboard(userID, page, limit)
	if err != nil {
		if serviceErr, ok := services.IsServiceError(err); ok {
			respondWithError(w, serviceErr.HTTPStatus, serviceErr.Message)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération du tableau de bord")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, dashboard)
}

// GetMessagingStats récupère les statistiques de messagerie
// GET /api/messaging/stats
func (h *MessagingHandler) GetMessagingStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Appeler le service
	stats, err := h.messagingManager.GetMessagingService().GetMessagingStats(userID)
	if err != nil {
		if serviceErr, ok := services.IsServiceError(err); ok {
			respondWithError(w, serviceErr.HTTPStatus, serviceErr.Message)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des statistiques")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, stats)
}

// ========== Search ==========

// SearchMessaging recherche dans conversations et messages
// GET /api/messaging/search?q=query&type=all&limit=20
func (h *MessagingHandler) SearchMessaging(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Paramètres de recherche
	query := r.URL.Query().Get("q")
	searchType := r.URL.Query().Get("type")
	if searchType == "" {
		searchType = "all"
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Validation de la requête
	if len(query) < 2 {
		respondWithError(w, http.StatusBadRequest, "La requête de recherche doit contenir au moins 2 caractères")
		return
	}

	// Nettoyer la requête
	query = utils.NormalizeSearchQuery(query)

	// Appeler le service
	results, err := h.messagingManager.GetMessagingService().SearchInMessaging(userID, query, searchType, limit)
	if err != nil {
		if serviceErr, ok := services.IsServiceError(err); ok {
			respondWithError(w, serviceErr.HTTPStatus, serviceErr.Message)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Erreur lors de la recherche")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, results)
}

// ========== Advanced Operations ==========

// StartConversationAndSendMessage crée une conversation et envoie le premier message
// POST /api/messaging/start
func (h *MessagingHandler) StartConversationAndSendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	var request struct {
		OtherUserID string                      `json:"other_user_id"`
		Message     services.SendMessageRequest `json:"message"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}
	defer r.Body.Close()

	// Validation
	if request.OtherUserID == "" {
		respondWithError(w, http.StatusBadRequest, "other_user_id requis")
		return
	}

	if err := request.Message.Validate(); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Nettoyer les entrées
	request.OtherUserID = utils.SanitizeInput(request.OtherUserID)
	if request.Message.Content != nil {
		cleaned := utils.CleanMessageContent(*request.Message.Content)
		request.Message.Content = &cleaned
	}

	// Appeler le service
	result, err := h.messagingManager.GetMessagingService().StartConversationAndSendMessage(
		userID, request.OtherUserID, &request.Message)
	if err != nil {
		if serviceErr, ok := services.IsServiceError(err); ok {
			respondWithError(w, serviceErr.HTTPStatus, serviceErr.Message)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Erreur lors de la création de la conversation")
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"conversation": result.Conversation,
		"message":      result.Message,
		"success":      "Conversation créée et message envoyé avec succès",
	})
}

// MarkAllConversationsAsRead marque toutes les conversations comme lues
// PUT /api/conversations/read-all
func (h *MessagingHandler) MarkAllConversationsAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Impossible d'extraire l'ID utilisateur")
		return
	}

	// Appeler le service
	err := h.messagingManager.GetMessagingService().MarkAllConversationsAsRead(userID)
	if err != nil {
		if serviceErr, ok := services.IsServiceError(err); ok {
			respondWithError(w, serviceErr.HTTPStatus, serviceErr.Message)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Erreur lors du marquage des conversations")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Toutes les conversations ont été marquées comme lues",
	})
}
