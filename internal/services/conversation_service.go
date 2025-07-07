package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/models"
	"github.com/google/uuid"
)

type ConversationService struct {
	db *sql.DB
}

func NewConversationService(db *sql.DB) *ConversationService {
	return &ConversationService{db: db}
}

// GetUserConversations retrieves all conversations for a user with pagination
func (s *ConversationService) GetUserConversations(userID string, limit, offset int) ([]models.Conversation, error) {
	query := `
		SELECT DISTINCT
			c.id, c.participant_1_id, c.participant_2_id, c.created_at, c.updated_at, c.is_active,
			-- Get other participant info
			CASE 
				WHEN c.participant_1_id = $1 THEN u2.id
				ELSE u1.id
			END as other_user_id,
			CASE 
				WHEN c.participant_1_id = $1 THEN u2.username
				ELSE u1.username
			END as other_username,
			CASE 
				WHEN c.participant_1_id = $1 THEN u2.avatar_url
				ELSE u1.avatar_url
			END as other_avatar_url,
			-- Get last message
			lm.id as last_message_id,
			lm.content as last_message_content,
			lm.message_type as last_message_type,
			lm.created_at as last_message_created_at,
			lm.is_paid as last_message_is_paid,
			-- Count unread messages
			(SELECT COUNT(*) FROM messages m 
			 WHERE m.conversation_id = c.id 
			 AND m.sender_id != $1 
			 AND m.read_at IS NULL) as unread_count
		FROM conversations c
		JOIN users u1 ON c.participant_1_id = u1.id
		JOIN users u2 ON c.participant_2_id = u2.id
		LEFT JOIN LATERAL (
			SELECT id, content, message_type, created_at, is_paid
			FROM messages 
			WHERE conversation_id = c.id 
			ORDER BY created_at DESC 
			LIMIT 1
		) lm ON true
		WHERE (c.participant_1_id = $1 OR c.participant_2_id = $1) 
		AND c.is_active = true
		ORDER BY COALESCE(lm.created_at, c.updated_at) DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query conversations: %w", err)
	}
	defer rows.Close()

	var conversations []models.Conversation
	for rows.Next() {
		var conv models.Conversation
		var otherUser models.User
		var lastMsg models.Message
		var lastMsgID, lastMsgContent, lastMsgType sql.NullString
		var lastMsgCreatedAt sql.NullTime
		var lastMsgIsPaid sql.NullBool

		err := rows.Scan(
			&conv.ID, &conv.Participant1ID, &conv.Participant2ID,
			&conv.CreatedAt, &conv.UpdatedAt, &conv.IsActive,
			&otherUser.ID, &otherUser.Username, &otherUser.AvatarURL,
			&lastMsgID, &lastMsgContent, &lastMsgType, &lastMsgCreatedAt, &lastMsgIsPaid,
			&conv.UnreadCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}

		// Add other participant to participants
		conv.Participants = []models.User{otherUser}

		// Add last message if exists
		if lastMsgID.Valid {
			lastMsg.ID = lastMsgID.String
			lastMsg.Content = &lastMsgContent.String
			lastMsg.MessageType = lastMsgType.String
			lastMsg.CreatedAt = lastMsgCreatedAt.Time
			lastMsg.IsPaid = lastMsgIsPaid.Bool
			conv.LastMessage = &lastMsg
		}

		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// CreateOrGetConversation creates a new conversation or returns existing one
func (s *ConversationService) CreateOrGetConversation(user1ID, user2ID string) (*models.Conversation, error) {
	if user1ID == user2ID {
		return nil, fmt.Errorf("cannot create conversation with yourself")
	}

	// Ensure consistent ordering for unique constraint
	var participant1, participant2 string
	if user1ID < user2ID {
		participant1, participant2 = user1ID, user2ID
	} else {
		participant1, participant2 = user2ID, user1ID
	}

	// First try to find existing conversation
	var conv models.Conversation
	query := `
		SELECT id, participant_1_id, participant_2_id, created_at, updated_at, is_active
		FROM conversations 
		WHERE participant_1_id = $1 AND participant_2_id = $2
	`
	
	err := s.db.QueryRow(query, participant1, participant2).Scan(
		&conv.ID, &conv.Participant1ID, &conv.Participant2ID,
		&conv.CreatedAt, &conv.UpdatedAt, &conv.IsActive,
	)

	if err == nil {
		// Conversation exists, reactivate if needed
		if !conv.IsActive {
			_, err = s.db.Exec(
				"UPDATE conversations SET is_active = true, updated_at = NOW() WHERE id = $1",
				conv.ID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to reactivate conversation: %w", err)
			}
			conv.IsActive = true
		}
		return &conv, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query existing conversation: %w", err)
	}

	// Create new conversation
	newID := uuid.New().String()
	insertQuery := `
		INSERT INTO conversations (id, participant_1_id, participant_2_id, created_at, updated_at, is_active)
		VALUES ($1, $2, $3, NOW(), NOW(), true)
		RETURNING created_at, updated_at
	`

	err = s.db.QueryRow(insertQuery, newID, participant1, participant2).Scan(
		&conv.CreatedAt, &conv.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	conv.ID = newID
	conv.Participant1ID = participant1
	conv.Participant2ID = participant2
	conv.IsActive = true

	return &conv, nil
}

// GetConversationMessages retrieves messages from a conversation with pagination
func (s *ConversationService) GetConversationMessages(conversationID string, userID string, limit, offset int) ([]models.Message, error) {
	// First verify user is participant
	isParticipant, err := s.IsParticipant(conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, fmt.Errorf("user is not a participant in this conversation")
	}

	query := `
		SELECT 
			m.id, m.conversation_id, m.sender_id, m.content, m.media_url, m.media_type,
			m.thumbnail_url, m.is_paid, m.price, m.is_unlocked, m.unlocked_at, m.unlocked_by,
			m.message_type, m.status, m.created_at, m.read_at,
			u.id as sender_user_id, u.username as sender_username, u.avatar_url as sender_avatar
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.conversation_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(query, conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var sender models.User

		err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Content, &msg.MediaURL,
			&msg.MediaType, &msg.ThumbnailURL, &msg.IsPaid, &msg.Price, &msg.IsUnlocked,
			&msg.UnlockedAt, &msg.UnlockedBy, &msg.MessageType, &msg.Status,
			&msg.CreatedAt, &msg.ReadAt,
			&sender.ID, &sender.Username, &sender.AvatarURL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		msg.Sender = &sender

		// If message is paid and not unlocked by current user, hide content
		if msg.IsPaid && !msg.IsUnlocked && msg.SenderID != userID {
			msg.Content = nil
			msg.MediaURL = nil
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// MarkConversationAsRead marks all unread messages in conversation as read
func (s *ConversationService) MarkConversationAsRead(conversationID, userID string) error {
	// Verify user is participant
	isParticipant, err := s.IsParticipant(conversationID, userID)
	if err != nil {
		return err
	}
	if !isParticipant {
		return fmt.Errorf("user is not a participant in this conversation")
	}

	query := `
		UPDATE messages 
		SET read_at = NOW() 
		WHERE conversation_id = $1 
		AND sender_id != $2 
		AND read_at IS NULL
	`

	_, err = s.db.Exec(query, conversationID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark messages as read: %w", err)
	}

	return nil
}

// GetUnreadCount returns total unread message count for user
func (s *ConversationService) GetUnreadCount(userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM messages m
		JOIN conversations c ON m.conversation_id = c.id
		WHERE (c.participant_1_id = $1 OR c.participant_2_id = $1)
		AND m.sender_id != $1
		AND m.read_at IS NULL
		AND c.is_active = true
	`

	var count int
	err := s.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// IsParticipant checks if user is participant in conversation
func (s *ConversationService) IsParticipant(conversationID, userID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM conversations 
			WHERE id = $1 
			AND (participant_1_id = $2 OR participant_2_id = $2)
			AND is_active = true
		)
	`

	var exists bool
	err := s.db.QueryRow(query, conversationID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check participation: %w", err)
	}

	return exists, nil
}

// UpdateConversationActivity updates the conversation's updated_at timestamp
func (s *ConversationService) UpdateConversationActivity(conversationID string) error {
	query := "UPDATE conversations SET updated_at = NOW() WHERE id = $1"
	_, err := s.db.Exec(query, conversationID)
	if err != nil {
		log.Printf("Failed to update conversation activity: %v", err)
	}
	return err
}
