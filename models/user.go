package models

import (
	"time"

	"gorm.io/gorm"
)

// UserRole définit le rôle de l'utilisateur
type UserRole string

const (
	RoleAdmin      UserRole = "admin"
	RoleCreator    UserRole = "creator"
	RoleSubscriber UserRole = "subscriber"
)

// User représente un utilisateur du système
type User struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	Username       string     `json:"username" gorm:"uniqueIndex;not null"`
	Email          string     `json:"email" gorm:"uniqueIndex;not null"`
	Password       string     `json:"-" gorm:"not null"` // Le "-" signifie que ce champ ne sera pas inclus dans la sérialisation JSON
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	Role           UserRole   `json:"role" gorm:"type:varchar(20);default:'subscriber'"`
	Biography      string     `json:"biography"`
	ProfilePicture string     `json:"profile_picture"`
	IsActive       bool       `json:"is_active" gorm:"default:true"`
	IsBanned       bool       `json:"is_banned" gorm:"default:false"`
	BanReason      string     `json:"ban_reason"`
	LastLogin      *time.Time `json:"last_login"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	Contents          []Content          `json:"-" gorm:"foreignKey:CreatorID"`
	SubscriptionPlans []SubscriptionPlan `json:"-" gorm:"foreignKey:CreatorID"`
	Subscriptions     []Subscription     `json:"-" gorm:"foreignKey:SubscriberID"`
	Comments          []Comment          `json:"-" gorm:"foreignKey:UserID"`
	Likes             []Like             `json:"-" gorm:"foreignKey:UserID"`
	Notifications     []Notification     `json:"-" gorm:"foreignKey:UserID"`
	SentMessages      []Message          `json:"-" gorm:"foreignKey:SenderID"`
	ReceivedMessages  []Message          `json:"-" gorm:"foreignKey:RecipientID"`
}

// UserResponse est utilisé pour retourner les données utilisateur sans le mot de passe
type UserResponse struct {
	ID             uint      `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Role           UserRole  `json:"role"`
	Biography      string    `json:"biography,omitempty"`
	ProfilePicture string    `json:"profile_picture,omitempty"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreatorProfile contient des informations supplémentaires pour les créateurs
type CreatorProfile struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	UserID           uint      `json:"user_id" gorm:"uniqueIndex"`
	User             User      `json:"-" gorm:"foreignKey:UserID"`
	BannerImage      string    `json:"banner_image"`
	WebsiteURL       string    `json:"website_url"`
	SocialLinks      string    `json:"social_links"`
	PaymentInfo      string    `json:"payment_info" gorm:"-"` // Non stocké en base, géré par service de paiement
	TotalSubscribers int       `json:"total_subscribers" gorm:"-"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Content représente un contenu publié par un créateur
type Content struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	CreatorID    uint           `json:"creator_id"`
	Creator      User           `json:"-" gorm:"foreignKey:CreatorID"`
	Title        string         `json:"title" gorm:"not null"`
	Description  string         `json:"description"`
	Type         string         `json:"type" gorm:"not null"` // image, video, text, etc.
	MediaURL     string         `json:"media_url"`
	ThumbnailURL string         `json:"thumbnail_url"`
	IsPremium    bool           `json:"is_premium" gorm:"default:false"`
	IsPublished  bool           `json:"is_published" gorm:"default:true"`
	ViewCount    int            `json:"view_count" gorm:"default:0"`
	IsFlagged    bool           `json:"is_flagged" gorm:"default:false"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Comments []Comment `json:"-" gorm:"foreignKey:ContentID"`
	Likes    []Like    `json:"-" gorm:"foreignKey:ContentID"`
	Reports  []Report  `json:"-" gorm:"foreignKey:ContentID"`
}

// SubscriptionPlan définit un plan d'abonnement créé par un créateur
type SubscriptionPlan struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	CreatorID   uint      `json:"creator_id"`
	Creator     User      `json:"-" gorm:"foreignKey:CreatorID"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Price       float64   `json:"price" gorm:"not null"`
	Duration    int       `json:"duration" gorm:"not null"` // En jours
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	Subscriptions []Subscription `json:"-" gorm:"foreignKey:PlanID"`
}

// Subscription représente l'abonnement d'un utilisateur à un créateur
type Subscription struct {
	ID            uint             `json:"id" gorm:"primaryKey"`
	SubscriberID  uint             `json:"subscriber_id"`
	Subscriber    User             `json:"-" gorm:"foreignKey:SubscriberID"`
	CreatorID     uint             `json:"creator_id"`
	Creator       User             `json:"-" gorm:"foreignKey:CreatorID"`
	PlanID        uint             `json:"plan_id"`
	Plan          SubscriptionPlan `json:"-" gorm:"foreignKey:PlanID"`
	StartDate     time.Time        `json:"start_date" gorm:"not null"`
	EndDate       time.Time        `json:"end_date" gorm:"not null"`
	IsActive      bool             `json:"is_active" gorm:"default:true"`
	AutoRenew     bool             `json:"auto_renew" gorm:"default:true"`
	PaymentStatus string           `json:"payment_status" gorm:"default:'paid'"` // paid, pending, failed
	TransactionID string           `json:"transaction_id"`
	CreatedAt     time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
}

// Comment représente un commentaire laissé par un utilisateur sur un contenu
type Comment struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ContentID uint           `json:"content_id"`
	Content   Content        `json:"-" gorm:"foreignKey:ContentID"`
	UserID    uint           `json:"user_id"`
	User      User           `json:"-" gorm:"foreignKey:UserID"`
	Text      string         `json:"text" gorm:"not null"`
	IsHidden  bool           `json:"is_hidden" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Like représente un "j'aime" d'un utilisateur sur un contenu
type Like struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ContentID uint      `json:"content_id"`
	Content   Content   `json:"-" gorm:"foreignKey:ContentID"`
	UserID    uint      `json:"user_id"`
	User      User      `json:"-" gorm:"foreignKey:UserID"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// Report représente un signalement fait par un utilisateur sur un contenu
type Report struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	ContentID     uint       `json:"content_id"`
	Content       Content    `json:"-" gorm:"foreignKey:ContentID"`
	ReporterID    uint       `json:"reporter_id"`
	Reporter      User       `json:"-" gorm:"foreignKey:ReporterID"`
	Reason        string     `json:"reason" gorm:"not null"`
	Status        string     `json:"status" gorm:"default:'pending'"` // pending, reviewed, dismissed
	ReviewedBy    *uint      `json:"reviewed_by"`
	ReviewerAdmin *User      `json:"-" gorm:"foreignKey:ReviewedBy"`
	ReviewedAt    *time.Time `json:"reviewed_at"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// Message représente un message privé entre deux utilisateurs
type Message struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	SenderID    uint       `json:"sender_id"`
	Sender      User       `json:"-" gorm:"foreignKey:SenderID"`
	RecipientID uint       `json:"recipient_id"`
	Recipient   User       `json:"-" gorm:"foreignKey:RecipientID"`
	Content     string     `json:"content" gorm:"not null"`
	IsRead      bool       `json:"is_read" gorm:"default:false"`
	ReadAt      *time.Time `json:"read_at"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// Notification représente une notification envoyée à un utilisateur
type Notification struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	UserID    uint       `json:"user_id"`
	User      User       `json:"-" gorm:"foreignKey:UserID"`
	Type      string     `json:"type" gorm:"not null"` // new_subscriber, new_comment, etc.
	Message   string     `json:"message" gorm:"not null"`
	IsRead    bool       `json:"is_read" gorm:"default:false"`
	ReadAt    *time.Time `json:"read_at"`
	RelatedID uint       `json:"related_id"` // ID de l'entité liée (commentaire, abonnement, etc.)
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// Transaction représente une transaction financière
type Transaction struct {
	ID             uint          `json:"id" gorm:"primaryKey"`
	UserID         uint          `json:"user_id"`
	User           User          `json:"-" gorm:"foreignKey:UserID"`
	SubscriptionID *uint         `json:"subscription_id"`
	Subscription   *Subscription `json:"-" gorm:"foreignKey:SubscriptionID"`
	Amount         float64       `json:"amount" gorm:"not null"`
	Currency       string        `json:"currency" gorm:"default:'EUR'"`
	Status         string        `json:"status" gorm:"not null"` // success, pending, failed
	PaymentMethod  string        `json:"payment_method"`
	PaymentID      string        `json:"payment_id"` // ID externe du système de paiement
	Description    string        `json:"description"`
	CreatedAt      time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
}

// Payout représente un versement à un créateur
type Payout struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	CreatorID     uint       `json:"creator_id"`
	Creator       User       `json:"-" gorm:"foreignKey:CreatorID"`
	Amount        float64    `json:"amount" gorm:"not null"`
	Currency      string     `json:"currency" gorm:"default:'EUR'"`
	Status        string     `json:"status" gorm:"not null"` // pending, processed, failed
	PaymentMethod string     `json:"payment_method"`
	Reference     string     `json:"reference"`
	ProcessedAt   *time.Time `json:"processed_at"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// ToResponse convertit un User en UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:             u.ID,
		Username:       u.Username,
		Email:          u.Email,
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		Role:           u.Role,
		Biography:      u.Biography,
		ProfilePicture: u.ProfilePicture,
		IsActive:       u.IsActive,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}
