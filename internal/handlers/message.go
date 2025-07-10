package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/internal/dto"
	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/Kaowarstail/Only-Flick-Go/internal/services"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type MessageHandler struct {
	messageService *services.MessageService
}

func NewMessageHandler(db *gorm.DB) *MessageHandler {
	return &MessageHandler{
		messageService: services.NewMessageService(db),
	}
}

// SendMessageClassic - POST /api/v1/messages
func (h *MessageHandler) SendMessageClassic(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	var req dto.SendMessageClassicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request data")
		return
	}
	defer r.Body.Close()

	message, err := h.messageService.SendMessageClassic(req, userID)
	if err != nil {
		if err.Error() == "access denied" {
			respondWithError(w, http.StatusForbidden, "Access denied to this conversation")
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Failed to send message: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, message)
}

// GetMessage - GET /api/v1/messages/{id}
func (h *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	vars := mux.Vars(r)
	messageID := vars["id"]

	if messageID == "" {
		respondWithError(w, http.StatusBadRequest, "Message ID is required")
		return
	}

	// Pour l'instant, on retourne une réponse simple
	// Dans une vraie implémentation, on récupérerait le message depuis la DB
	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Get message endpoint - to be implemented",
		"user_id": userID,
		"message_id": messageID,
	})
}

// MarkAsRead - PUT /api/v1/messages/read
func (h *MessageHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	var req struct {
		ConversationID string `json:"conversation_id" binding:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request data")
		return
	}
	defer r.Body.Close()

	err := h.messageService.MarkMessageAsRead(req.ConversationID, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to mark messages as read: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Messages marked as read successfully",
	})
}
