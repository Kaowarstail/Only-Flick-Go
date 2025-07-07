package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/models"
	"github.com/google/uuid"
)

type MessageService struct {
	db                  *sql.DB
	conversationService *ConversationService
}

func NewMessageService(db *sql.DB, conversationService *ConversationService) *MessageService {
	return &MessageService{
		db:                  db,
		conversationService: conversationService,
	}
}

// SendMessage sends a regular message
func (s *MessageService) SendMessage(req models.SendMessageRequest, senderID string) (*models.Message, error) {
	// Validate message content
	if err := s.validateMessageContent(req.Content, req.MediaURL); err != nil {
		return nil, err
	}

	// Verify sender is participant in conversation
	isParticipant, err := s.conversationService.IsParticipant(req.ConversationID, senderID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, fmt.Errorf("sender is not a participant in this conversation")
	}

	// Create message
	messageID := uuid.New().String()
	query := `
		INSERT INTO messages (
			id, conversation_id, sender_id, content, media_url, message_type,
			is_paid, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, false, 'sent', NOW())
		RETURNING created_at
	`

	var message models.Message
	err = s.db.QueryRow(query, messageID, req.ConversationID, senderID, 
		req.Content, req.MediaURL, req.MessageType).Scan(&message.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	message.ID = messageID
	message.ConversationID = req.ConversationID
	message.SenderID = senderID
	message.Content = req.Content
	message.MediaURL = req.MediaURL
	message.MessageType = req.MessageType
	message.IsPaid = false
	message.Status = "sent"

	// Update conversation activity
	s.conversationService.UpdateConversationActivity(req.ConversationID)

	// Update user stats
	s.updateUserMessageStats(senderID)

	return &message, nil
}

// SendPaidMessage sends a paid message
func (s *MessageService) SendPaidMessage(req models.SendPaidMessageRequest, senderID string) (*models.Message, error) {
	// Validate message content and price
	if err := s.validateMessageContent(req.Content, req.MediaURL); err != nil {
		return nil, err
	}
	if err := s.validatePaidMessagePrice(req.Price); err != nil {
		return nil, err
	}

	// Verify sender is participant in conversation
	isParticipant, err := s.conversationService.IsParticipant(req.ConversationID, senderID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, fmt.Errorf("sender is not a participant in this conversation")
	}

	// Create paid message
	messageID := uuid.New().String()
	query := `
		INSERT INTO messages (
			id, conversation_id, sender_id, content, media_url, message_type,
			is_paid, price, is_unlocked, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, true, $7, false, 'sent', NOW())
		RETURNING created_at
	`

	var message models.Message
	err = s.db.QueryRow(query, messageID, req.ConversationID, senderID,
		req.Content, req.MediaURL, req.MessageType, req.Price).Scan(&message.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create paid message: %w", err)
	}

	message.ID = messageID
	message.ConversationID = req.ConversationID
	message.SenderID = senderID
	message.Content = req.Content
	message.MediaURL = req.MediaURL
	message.MessageType = req.MessageType
	message.IsPaid = true
	message.Price = &req.Price
	message.IsUnlocked = false
	message.Status = "sent"

	// Update conversation activity
	s.conversationService.UpdateConversationActivity(req.ConversationID)

	// Update user stats
	s.updateUserMessageStats(senderID)

	return &message, nil
}

// UnlockPaidMessage unlocks a paid message after payment processing
func (s *MessageService) UnlockPaidMessage(messageID, buyerID string, paymentMethod string) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Get message details
	var message models.Message
	var sellerID string
	query := `
		SELECT m.id, m.sender_id, m.price, m.is_paid, m.is_unlocked, m.conversation_id
		FROM messages m
		WHERE m.id = $1 AND m.is_paid = true
	`
	
	err = tx.QueryRow(query, messageID).Scan(
		&message.ID, &sellerID, &message.Price, &message.IsPaid, &message.IsUnlocked, &message.ConversationID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("paid message not found")
		}
		return fmt.Errorf("failed to get message: %w", err)
	}

	// Check if already unlocked
	if message.IsUnlocked {
		return fmt.Errorf("message already unlocked")
	}

	// Verify buyer is participant in conversation
	isParticipant, err := s.conversationService.IsParticipant(message.ConversationID, buyerID)
	if err != nil {
		return err
	}
	if !isParticipant {
		return fmt.Errorf("buyer is not a participant in this conversation")
	}

	// Check buyer is not the sender
	if buyerID == sellerID {
		return fmt.Errorf("cannot unlock your own message")
	}

	// Calculate fees (20% platform fee)
	amount := *message.Price
	platformFee := amount * 0.20
	sellerEarnings := amount - platformFee

	// Create transaction record
	transactionID := uuid.New().String()
	transactionQuery := `
		INSERT INTO paid_message_transactions (
			id, message_id, buyer_id, seller_id, amount, platform_fee, seller_earnings,
			status, payment_method, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', $8, NOW())
	`
	
	_, err = tx.Exec(transactionQuery, transactionID, messageID, buyerID, sellerID,
		amount, platformFee, sellerEarnings, paymentMethod)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// TODO: Here you would integrate with payment processor (Stripe, etc.)
	// For now, we'll mark as completed immediately
	
	// Update transaction status
	_, err = tx.Exec(
		"UPDATE paid_message_transactions SET status = 'completed', completed_at = NOW() WHERE id = $1",
		transactionID,
	)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	// Unlock message
	unlockQuery := `
		UPDATE messages 
		SET is_unlocked = true, unlocked_at = NOW(), unlocked_by = $1
		WHERE id = $2
	`
	
	_, err = tx.Exec(unlockQuery, buyerID, messageID)
	if err != nil {
		return fmt.Errorf("failed to unlock message: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetPaidMessagePreview returns preview of paid message for non-buyers
func (s *MessageService) GetPaidMessagePreview(messageID, userID string) (string, error) {
	var senderID string
	var content sql.NullString
	var price sql.NullFloat64
	var isUnlocked bool
	var conversationID string

	query := `
		SELECT sender_id, content, price, is_unlocked, conversation_id
		FROM messages 
		WHERE id = $1 AND is_paid = true
	`
	
	err := s.db.QueryRow(query, messageID).Scan(&senderID, &content, &price, &isUnlocked, &conversationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("paid message not found")
		}
		return "", fmt.Errorf("failed to get message: %w", err)
	}

	// Verify user is participant
	isParticipant, err := s.conversationService.IsParticipant(conversationID, userID)
	if err != nil {
		return "", err
	}
	if !isParticipant {
		return "", fmt.Errorf("user is not a participant in this conversation")
	}

	// If sender or message is unlocked, return full content
	if userID == senderID || isUnlocked {
		if content.Valid {
			return content.String, nil
		}
		return "", nil
	}

	// Return preview for buyers
	if content.Valid && len(content.String) > 50 {
		return content.String[:50] + "... [Unlock for $" + fmt.Sprintf("%.2f", price.Float64) + "]", nil
	}
	
	return "[Paid message - $" + fmt.Sprintf("%.2f", price.Float64) + "]", nil
}

// MarkMessageAsRead marks a message as read
func (s *MessageService) MarkMessageAsRead(messageID, userID string) error {
	// First get message details to verify permissions
	var senderID, conversationID string
	query := `
		SELECT sender_id, conversation_id 
		FROM messages 
		WHERE id = $1
	`
	
	err := s.db.QueryRow(query, messageID).Scan(&senderID, &conversationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("message not found")
		}
		return fmt.Errorf("failed to get message: %w", err)
	}

	// Verify user is participant and not the sender
	isParticipant, err := s.conversationService.IsParticipant(conversationID, userID)
	if err != nil {
		return err
	}
	if !isParticipant {
		return fmt.Errorf("user is not a participant in this conversation")
	}
	if userID == senderID {
		return fmt.Errorf("cannot mark own message as read")
	}

	// Mark as read
	updateQuery := `
		UPDATE messages 
		SET read_at = NOW() 
		WHERE id = $1 AND read_at IS NULL
	`
	
	_, err = s.db.Exec(updateQuery, messageID)
	if err != nil {
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	return nil
}

// DeleteMessage soft deletes a message (sender only)
func (s *MessageService) DeleteMessage(messageID, userID string) error {
	// Verify user is the sender
	var senderID string
	query := `
		SELECT sender_id 
		FROM messages 
		WHERE id = $1
	`
	
	err := s.db.QueryRow(query, messageID).Scan(&senderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("message not found")
		}
		return fmt.Errorf("failed to get message: %w", err)
	}

	if userID != senderID {
		return fmt.Errorf("can only delete your own messages")
	}

	// Soft delete by updating status
	updateQuery := `
		UPDATE messages 
		SET status = 'deleted', content = NULL, media_url = NULL
		WHERE id = $1
	`
	
	_, err = s.db.Exec(updateQuery, messageID)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// Helper functions

func (s *MessageService) validateMessageContent(content *string, mediaURL *string) error {
	if content == nil && mediaURL == nil {
		return fmt.Errorf("message must have content or media")
	}
	if content != nil && len(*content) > 5000 {
		return fmt.Errorf("message content too long (max 5000 characters)")
	}
	return nil
}

func (s *MessageService) validatePaidMessagePrice(price float64) error {
	if price < 0.99 {
		return fmt.Errorf("minimum price is $0.99")
	}
	if price > 500.00 {
		return fmt.Errorf("maximum price is $500.00")
	}
	return nil
}

func (s *MessageService) updateUserMessageStats(userID string) {
	// Update user stats (fire and forget)
	go func() {
		query := `
			INSERT INTO user_stats (user_id, total_messages_sent, last_active_at, updated_at)
			VALUES ($1, 1, NOW(), NOW())
			ON CONFLICT (user_id) DO UPDATE SET
				total_messages_sent = user_stats.total_messages_sent + 1,
				last_active_at = NOW(),
				updated_at = NOW()
		`
		s.db.Exec(query, userID)
	}()
}
