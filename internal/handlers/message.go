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

type MessageHandler struct {
	messageService *services.MessageService
}

func NewMessageHandler(db *gorm.DB) *MessageHandler {
	return &MessageHandler{
		messageService: services.NewMessageService(db),
	}
}

// SendMessage - POST /api/v1/messages
func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	var req dto.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request data")
		return
	}
	defer r.Body.Close()

	message, err := h.messageService.SendMessage(req, userID)
	if err != nil {
		if err.Error() == "access denied" {
			respondWithError(w, http.StatusForbidden, "Access denied to this conversation")
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Failed to send message: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    message,
		"message": "Message sent successfully",
	}

	respondWithJSON(w, http.StatusCreated, response)
}

// SendPaidMessage - POST /api/v1/messages/paid
func (h *MessageHandler) SendPaidMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	var req dto.SendPaidMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request data")
		return
	}
	defer r.Body.Close()

	message, err := h.messageService.SendPaidMessage(req, userID)
	if err != nil {
		if err.Error() == "access denied" {
			respondWithError(w, http.StatusForbidden, "Access denied to this conversation")
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Failed to send paid message: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    message,
		"message": "Paid message sent successfully",
	}

	respondWithJSON(w, http.StatusCreated, response)
}

// UnlockPaidMessage - POST /api/v1/messages/{id}/unlock
func (h *MessageHandler) UnlockPaidMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	vars := mux.Vars(r)
	messageID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	err = h.messageService.UnlockPaidMessage(uint(messageID), userID)
	if err != nil {
		switch err.Error() {
		case "access denied":
			respondWithError(w, http.StatusForbidden, "Access denied")
		case "message is not paid":
			respondWithError(w, http.StatusBadRequest, "Message is not a paid message")
		case "message already unlocked":
			respondWithError(w, http.StatusBadRequest, "Message is already unlocked")
		case "cannot unlock own message":
			respondWithError(w, http.StatusBadRequest, "Cannot unlock your own message")
		default:
			respondWithError(w, http.StatusInternalServerError, "Failed to unlock message")
		}
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Message unlocked successfully",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetMessage - GET /api/v1/messages/{id}
func (h *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	vars := mux.Vars(r)
	messageID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	message, err := h.messageService.GetMessageWithAccess(uint(messageID), userID)
	if err != nil {
		if err.Error() == "access denied" {
			respondWithError(w, http.StatusForbidden, "Access denied to this message")
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Failed to fetch message")
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    message,
	}

	respondWithJSON(w, http.StatusOK, response)
}
