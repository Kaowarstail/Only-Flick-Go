package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/services"
	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
	"github.com/Kaowarstail/Only-Flick-Go/models"
	"github.com/gorilla/mux"
)

// ConversationHandler gère les requêtes liées aux conversations
type ConversationHandler struct {
	conversationService *services.ConversationService
}

// NewConversationHandler crée un nouveau handler pour les conversations
func NewConversationHandler() *ConversationHandler {
	return &ConversationHandler{
		conversationService: services.NewConversationService(database.GetDB()),
	}
}

// GetUserConversations récupère toutes les conversations d'un utilisateur
func (h *ConversationHandler) GetUserConversations(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	conversations, err := h.conversationService.GetUserConversations(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des conversations")
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    conversations,
		Message: "Conversations récupérées avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetConversation récupère une conversation spécifique
func (h *ConversationHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	conversationID := vars["id"]

	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Vérifier si l'utilisateur est participant de la conversation
	isParticipant, err := h.conversationService.IsParticipant(conversationID, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la vérification des permissions")
		return
	}

	if !isParticipant {
		respondWithError(w, http.StatusForbidden, "Accès non autorisé à cette conversation")
		return
	}

	// Récupérer la conversation
	var conversation models.Conversation
	if err := database.GetDB().First(&conversation, "id = ?", conversationID).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Conversation non trouvée")
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    conversation,
		Message: "Conversation récupérée avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// CreateConversation crée une nouvelle conversation
func (h *ConversationHandler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ParticipantID string `json:"participant_id"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Données invalides")
		return
	}
	defer r.Body.Close()

	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	if request.ParticipantID == "" {
		respondWithError(w, http.StatusBadRequest, "ID du participant requis")
		return
	}

	// Créer ou récupérer la conversation existante
	conversation, err := h.conversationService.CreateOrGetConversation(userID, request.ParticipantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la création de la conversation")
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    conversation,
		Message: "Conversation créée avec succès",
	}

	respondWithJSON(w, http.StatusCreated, response)
}

// GetConversationMessages récupère les messages d'une conversation
func (h *ConversationHandler) GetConversationMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	conversationID := vars["id"]

	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Vérifier si l'utilisateur est participant de la conversation
	isParticipant, err := h.conversationService.IsParticipant(conversationID, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la vérification des permissions")
		return
	}

	if !isParticipant {
		respondWithError(w, http.StatusForbidden, "Accès non autorisé à cette conversation")
		return
	}

	// Pagination
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	messages, err := h.conversationService.GetConversationMessages(conversationID, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des messages")
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    messages,
		Message: "Messages récupérés avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// MarkConversationAsRead marque une conversation comme lue
func (h *ConversationHandler) MarkConversationAsRead(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	conversationID := vars["id"]

	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Vérifier si l'utilisateur est participant de la conversation
	isParticipant, err := h.conversationService.IsParticipant(conversationID, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la vérification des permissions")
		return
	}

	if !isParticipant {
		respondWithError(w, http.StatusForbidden, "Accès non autorisé à cette conversation")
		return
	}

	if err := h.conversationService.MarkConversationAsRead(conversationID, userID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors du marquage comme lu")
		return
	}

	response := models.APIResponse{
		Success: true,
		Message: "Conversation marquée comme lue",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetUnreadCount récupère le nombre de messages non lus
func (h *ConversationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	count, err := h.conversationService.GetUnreadCount(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors du calcul des messages non lus")
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    map[string]int{"unread_count": count},
		Message: "Nombre de messages non lus récupéré avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// getUserIDFromContext récupère l'ID utilisateur depuis le contexte
func getUserIDFromContext(ctx interface{}) string {
	userID, err := utils.GetUserIDFromContext(ctx.(context.Context))
	if err != nil {
		return ""
	}
	return userID
}
