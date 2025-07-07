package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/internal/services"
	"github.com/Kaowarstail/Only-Flick-Go/internal/utils"
	"github.com/Kaowarstail/Only-Flick-Go/models"
	websocketPkg "github.com/Kaowarstail/Only-Flick-Go/internal/websocket"
	"github.com/gorilla/mux"
)

// MessageHandler gère les requêtes liées aux messages
type MessageHandler struct {
	messageService *services.MessageService
	hub           *websocketPkg.Hub
}

// NewMessageHandler crée un nouveau handler pour les messages
func NewMessageHandler(hub *websocketPkg.Hub) *MessageHandler {
	return &MessageHandler{
		messageService: services.NewMessageService(database.GetDB()),
		hub:           hub,
	}
}

// getUserIDFromContext récupère l'ID utilisateur depuis le contexte
func getUserIDFromContext(ctx context.Context) string {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return ""
	}
	return userID
}

// SendMessage envoie un message dans une conversation
func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ConversationID string                 `json:"conversation_id"`
		Content        string                 `json:"content"`
		MessageType    models.MessageType     `json:"message_type"`
		MediaURL       string                 `json:"media_url,omitempty"`
		MediaType      models.MediaType       `json:"media_type,omitempty"`
		IsPaid         bool                   `json:"is_paid,omitempty"`
		Price          float64                `json:"price,omitempty"`
		PreviewText    string                 `json:"preview_text,omitempty"`
		Metadata       map[string]interface{} `json:"metadata,omitempty"`
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

	// Validation des données
	if request.ConversationID == "" {
		respondWithError(w, http.StatusBadRequest, "ID de conversation requis")
		return
	}

	if request.Content == "" && request.MediaURL == "" {
		respondWithError(w, http.StatusBadRequest, "Contenu ou média requis")
		return
	}

	// Créer le message
	message := &models.Message{
		ConversationID: request.ConversationID,
		SenderID:       userID,
		Content:        request.Content,
		MessageType:    request.MessageType,
		MediaURL:       request.MediaURL,
		MediaType:      request.MediaType,
		IsPaid:         request.IsPaid,
		Price:          request.Price,
		PreviewText:    request.PreviewText,
		Metadata:       request.Metadata,
	}

	var err error
	if request.IsPaid {
		message, err = h.messageService.SendPaidMessage(message)
	} else {
		message, err = h.messageService.SendMessage(message)
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de l'envoi du message")
		return
	}

	// Récupérer les informations complètes pour le WebSocket
	conversation, err := h.messageService.GetConversationByID(request.ConversationID)
	if err != nil {
		// Log l'erreur mais ne pas faire échouer la requête
		// Le message a été envoyé avec succès
		// TODO: Log properly
	}

	sender, err := h.messageService.GetUserByID(userID)
	if err != nil {
		// Log l'erreur mais ne pas faire échouer la requête
		// TODO: Log properly
	}

	// Broadcaster le message via WebSocket si les informations sont disponibles
	if h.hub != nil && conversation != nil && sender != nil {
		h.hub.BroadcastMessageSent(message, conversation, sender)
	}

	response := models.APIResponse{
		Success: true,
		Data:    message,
		Message: "Message envoyé avec succès",
	}

	respondWithJSON(w, http.StatusCreated, response)
}

// UnlockPaidMessage débloque un message payant
func (h *MessageHandler) UnlockPaidMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	messageID := vars["id"]

	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	message, err := h.messageService.UnlockPaidMessage(messageID, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors du déblocage du message")
		return
	}

	// Récupérer les informations de transaction pour le WebSocket
	transaction, err := h.messageService.GetLatestTransactionForMessage(messageID)
	if err != nil {
		// Log l'erreur mais ne pas faire échouer la requête
		// TODO: Log properly
	}

	// Broadcaster le déblocage via WebSocket
	if h.hub != nil {
		h.hub.BroadcastPaidMessageUnlocked(message, userID, transaction)
	}

	response := models.APIResponse{
		Success: true,
		Data:    message,
		Message: "Message débloqué avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetPaidMessagePreview récupère l'aperçu d'un message payant
func (h *MessageHandler) GetPaidMessagePreview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	messageID := vars["id"]

	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	preview, err := h.messageService.GetPaidMessagePreview(messageID, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération de l'aperçu")
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    preview,
		Message: "Aperçu récupéré avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// MarkMessageAsRead marque un message comme lu
func (h *MessageHandler) MarkMessageAsRead(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	messageID := vars["id"]

	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	if err := h.messageService.MarkMessageAsRead(messageID, userID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors du marquage comme lu")
		return
	}

	response := models.APIResponse{
		Success: true,
		Message: "Message marqué comme lu",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// DeleteMessage supprime un message
func (h *MessageHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	messageID := vars["id"]

	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	if err := h.messageService.DeleteMessage(messageID, userID); err != nil {
		if err.Error() == "message not found or not authorized" {
			respondWithError(w, http.StatusNotFound, "Message non trouvé ou non autorisé")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Erreur lors de la suppression du message")
		}
		return
	}

	response := models.APIResponse{
		Success: true,
		Message: "Message supprimé avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetMessageTransactions récupère les transactions d'un message payant
func (h *MessageHandler) GetMessageTransactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	messageID := vars["id"]

	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Vérifier que l'utilisateur est l'expéditeur du message
	var message models.Message
	if err := database.GetDB().First(&message, "id = ? AND sender_id = ?", messageID, userID).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Message non trouvé ou non autorisé")
		return
	}

	var transactions []models.PaidMessageTransaction
	if err := database.GetDB().Where("message_id = ?", messageID).Find(&transactions).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Erreur lors de la récupération des transactions")
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    transactions,
		Message: "Transactions récupérées avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetUserMessagesStats récupère les statistiques de messages d'un utilisateur
func (h *MessageHandler) GetUserMessagesStats(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Récupérer les statistiques depuis la base de données
	var stats struct {
		TotalMessages     int64   `json:"total_messages"`
		PaidMessages      int64   `json:"paid_messages"`
		UnlockedMessages  int64   `json:"unlocked_messages"`
		TotalEarnings     float64 `json:"total_earnings"`
		TotalSpent        float64 `json:"total_spent"`
		ConversationCount int64   `json:"conversation_count"`
	}

	db := database.GetDB()

	// Messages envoyés
	db.Model(&models.Message{}).Where("sender_id = ?", userID).Count(&stats.TotalMessages)
	db.Model(&models.Message{}).Where("sender_id = ? AND is_paid = true", userID).Count(&stats.PaidMessages)

	// Messages débloqués (achetés)
	db.Model(&models.PaidMessageTransaction{}).Where("buyer_id = ?", userID).Count(&stats.UnlockedMessages)

	// Gains totaux
	db.Model(&models.PaidMessageTransaction{}).Select("COALESCE(SUM(creator_amount), 0)").Where("creator_id = ?", userID).Scan(&stats.TotalEarnings)

	// Dépenses totales
	db.Model(&models.PaidMessageTransaction{}).Select("COALESCE(SUM(amount), 0)").Where("buyer_id = ?", userID).Scan(&stats.TotalSpent)

	// Nombre de conversations
	db.Model(&models.Conversation{}).Where("user1_id = ? OR user2_id = ?", userID, userID).Count(&stats.ConversationCount)

	response := models.APIResponse{
		Success: true,
		Data:    stats,
		Message: "Statistiques récupérées avec succès",
	}

	respondWithJSON(w, http.StatusOK, response)
}
