package models

import (
	"time"
	"database/sql/driver"
	"encoding/json"
)

// Conversation represents a chat conversation between two users
type Conversation struct {
	ID            string    `json:"id" db:"id"`
	Participant1ID string   `json:"participant_1_id" db:"participant_1_id"`
	Participant2ID string   `json:"participant_2_id" db:"participant_2_id"`
	Participants   []User   `json:"participants"`
	LastMessage    *Message `json:"last_message"`
	UnreadCount    int      `json:"unread_count"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	IsActive       bool     `json:"is_active" db:"is_active"`
}

// Message represents a message in a conversation
type Message struct {
	ID             string     `json:"id" db:"id"`
	ConversationID string     `json:"conversation_id" db:"conversation_id"`
	SenderID       string     `json:"sender_id" db:"sender_id"`
	Sender         *User      `json:"sender,omitempty"`
	Content        *string    `json:"content" db:"content"`
	MediaURL       *string    `json:"media_url" db:"media_url"`
	MediaType      *string    `json:"media_type" db:"media_type"`
	ThumbnailURL   *string    `json:"thumbnail_url" db:"thumbnail_url"`
	
	// Messages payants
	IsPaid       bool       `json:"is_paid" db:"is_paid"`
	Price        *float64   `json:"price" db:"price"`
	IsUnlocked   bool       `json:"is_unlocked" db:"is_unlocked"`
	UnlockedAt   *time.Time `json:"unlocked_at" db:"unlocked_at"`
	UnlockedBy   *string    `json:"unlocked_by" db:"unlocked_by"`
	
	MessageType  string     `json:"message_type" db:"message_type"`
	Status       string     `json:"status" db:"status"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	ReadAt       *time.Time `json:"read_at" db:"read_at"`
}

// SocialLinks represents user's social media links
type SocialLinks struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Instagram *string    `json:"instagram" db:"instagram"`
	Twitter   *string    `json:"twitter" db:"twitter"`
	TikTok    *string    `json:"tiktok" db:"tiktok"`
	YouTube   *string    `json:"youtube" db:"youtube"`
	Website   *string    `json:"website" db:"website"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// UserStats represents user statistics
type UserStats struct {
	UserID             string    `json:"user_id" db:"user_id"`
	FollowersCount     int       `json:"followers_count" db:"followers_count"`
	FollowingCount     int       `json:"following_count" db:"following_count"`
	PostsCount         int       `json:"posts_count" db:"posts_count"`
	TotalMessagesSent  int       `json:"total_messages_sent" db:"total_messages_sent"`
	LastActiveAt       time.Time `json:"last_active_at" db:"last_active_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// PaidMessageTransaction represents a paid message transaction
type PaidMessageTransaction struct {
	ID            string     `json:"id" db:"id"`
	MessageID     string     `json:"message_id" db:"message_id"`
	BuyerID       string     `json:"buyer_id" db:"buyer_id"`
	SellerID      string     `json:"seller_id" db:"seller_id"`
	Amount        float64    `json:"amount" db:"amount"`
	PlatformFee   float64    `json:"platform_fee" db:"platform_fee"`
	SellerEarnings float64   `json:"seller_earnings" db:"seller_earnings"`
	Status        string     `json:"status" db:"status"`
	PaymentMethod *string    `json:"payment_method" db:"payment_method"`
	TransactionID *string    `json:"transaction_id" db:"transaction_id"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	CompletedAt   *time.Time `json:"completed_at" db:"completed_at"`
}

// Request/Response models

// CreateConversationRequest represents request to create conversation
type CreateConversationRequest struct {
	OtherUserID string `json:"other_user_id" binding:"required"`
}

// SendMessageRequest represents request to send a message
type SendMessageRequest struct {
	ConversationID string  `json:"conversation_id" binding:"required"`
	Content        *string `json:"content"`
	MediaURL       *string `json:"media_url"`
	MessageType    string  `json:"message_type" binding:"required,oneof=text image video"`
}

// SendPaidMessageRequest represents request to send a paid message
type SendPaidMessageRequest struct {
	ConversationID string  `json:"conversation_id" binding:"required"`
	Content        *string `json:"content"`
	MediaURL       *string `json:"media_url"`
	Price          float64 `json:"price" binding:"required,min=0.99,max=500"`
	MessageType    string  `json:"message_type" binding:"required,oneof=paid_text paid_media"`
}

// UnlockPaidMessageRequest represents request to unlock a paid message
type UnlockPaidMessageRequest struct {
	PaymentMethod string `json:"payment_method" binding:"required"`
}

// UpdateUserProfileRequest represents request to update user profile
type UpdateUserProfileRequest struct {
	Bio       *string `json:"bio"`
	Username  *string `json:"username" binding:"omitempty,min=3,max=30"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	IsPrivate *bool   `json:"is_private"`
}

// UpdateSocialLinksRequest represents request to update social links
type UpdateSocialLinksRequest struct {
	Instagram *string `json:"instagram"`
	Twitter   *string `json:"twitter"`
	TikTok    *string `json:"tiktok"`
	YouTube   *string `json:"youtube"`
	Website   *string `json:"website"`
}

// UpdateCreatorProfileRequest represents request to update creator profile
type UpdateCreatorProfileRequest struct {
	DisplayName            *string  `json:"display_name"`
	Bio                    *string  `json:"bio"`
	SubscriptionPrice      *float64 `json:"subscription_price" binding:"omitempty,min=4.99,max=99.99"`
	Category               *string  `json:"category"`
	AcceptsCustomRequests  *bool    `json:"accepts_custom_requests"`
}

// CreatorEarnings represents creator earnings data
type CreatorEarnings struct {
	CreatorID             string           `json:"creator_id"`
	TotalEarnings         float64          `json:"total_earnings"`
	CurrentMonthEarnings  float64          `json:"current_month_earnings"`
	LastMonthEarnings     float64          `json:"last_month_earnings"`
	SubscriptionEarnings  float64          `json:"subscription_earnings"`
	PaidMessageEarnings   float64          `json:"paid_message_earnings"`
	MonthlyBreakdown      []MonthlyEarning `json:"monthly_breakdown"`
	LastUpdated           time.Time        `json:"last_updated"`
}

// MonthlyEarning represents earnings for a specific month
type MonthlyEarning struct {
	Year                 int     `json:"year"`
	Month                int     `json:"month"`
	MonthName            string  `json:"month_name"`
	SubscriptionEarnings float64 `json:"subscription_earnings"`
	PaidMessageEarnings  float64 `json:"paid_message_earnings"`
	TotalEarnings        float64 `json:"total_earnings"`
}

// MediaUploadResponse represents response from media upload
type MediaUploadResponse struct {
	MediaURL     string  `json:"media_url"`
	ThumbnailURL *string `json:"thumbnail_url,omitempty"`
	FileType     string  `json:"file_type"`
	FileSize     int64   `json:"file_size"`
}

// APIResponse represents standardized API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// APIError represents API error details
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Meta represents pagination and metadata
type Meta struct {
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// MediaFile représente un fichier média uploadé
type MediaFile struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	UserID     string    `json:"user_id" gorm:"not null;index"`
	FileName   string    `json:"file_name" gorm:"not null"`
	FilePath   string    `json:"file_path" gorm:"not null"`
	FileURL    string    `json:"file_url" gorm:"not null"`
	FileSize   int64     `json:"file_size" gorm:"not null"`
	MediaType  MediaType `json:"media_type" gorm:"not null"`
	MimeType   string    `json:"mime_type" gorm:"not null"`
	UploadedAt time.Time `json:"uploaded_at" gorm:"autoCreateTime"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName spécifie le nom de la table pour MediaFile
func (MediaFile) TableName() string {
	return "media_files"
}

// Helper methods for JSON handling
func (sl *SocialLinks) Value() (driver.Value, error) {
	return json.Marshal(sl)
}

func (sl *SocialLinks) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, sl)
}
