package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Kaowarstail/Only-Flick-Go/internal/dto"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/Kaowarstail/Only-Flick-Go/internal/services"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type ConversationHandler struct {
	conversationService *services.ConversationService
	messageService      *services.MessageService
}

func NewConversationHandler(db *gorm.DB) *ConversationHandler {
	return &ConversationHandler{
		conversationService: services.NewConversationService(db),
		messageService:      services.NewMessageService(db),
	}
}

// GetUserConversations - GET /api/v1/conversations
func (h *ConversationHandler) GetUserConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	page := 1
	limit := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	conversations, total, err := h.conversationService.GetUserConversations(userID, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch conversations")
		return
	}

	// Calculer unread total
	unreadTotal := 0
	for _, conv := range conversations {
		unreadTotal += conv.UnreadCount
	}

	response := map[string]interface{}{
		"success": true,
		"data": dto.ConversationsResponse{
			Conversations: conversations,
			Total:         total,
			UnreadTotal:   unreadTotal,
		},
		"meta": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	}

	respondWithJSON(w, http.StatusOK, response)
}

// CreateConversation - POST /api/v1/conversations
func (h *ConversationHandler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	var req dto.CreateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request data")
		return
	}
	defer r.Body.Close()

	// Vérifier qu'on ne crée pas conversation avec soi-même
	if req.OtherUserID == userID {
		respondWithError(w, http.StatusBadRequest, "Cannot create conversation with yourself")
		return
	}

	conversation, err := h.conversationService.CreateOrGetConversation(userID, req.OtherUserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create conversation")
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    conversation,
		"message": "Conversation created successfully",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetConversationMessages - GET /api/v1/conversations/{id}/messages
func (h *ConversationHandler) GetConversationMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	vars := mux.Vars(r)
	conversationID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	page := 1
	limit := 50

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	messages, err := h.messageService.GetMessages(uint(conversationID), userID, page, limit)
	if err != nil {
		if err.Error() == "access denied" {
			respondWithError(w, http.StatusForbidden, "Access denied to this conversation")
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Failed to fetch messages")
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    messages,
		"meta": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       messages.Total,
			"total_pages": (messages.Total + int64(limit) - 1) / int64(limit),
			"has_more":    messages.HasMore,
		},
	}

	respondWithJSON(w, http.StatusOK, response)
}

// MarkConversationAsRead - PUT /api/v1/conversations/{id}/read
func (h *ConversationHandler) MarkConversationAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	vars := mux.Vars(r)
	conversationID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	err = h.messageService.MarkAsRead(uint(conversationID), userID)
	if err != nil {
		if err.Error() == "access denied" {
			respondWithError(w, http.StatusForbidden, "Access denied to this conversation")
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Failed to mark messages as read")
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Messages marked as read",
	}

	respondWithJSON(w, http.StatusOK, response)
}
