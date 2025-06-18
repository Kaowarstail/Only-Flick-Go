package models

import (
	"time"

	"gorm.io/gorm"
)

// PasswordResetToken représente un token de réinitialisation de mot de passe
type PasswordResetToken struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    string         `json:"user_id" gorm:"not null"`
	User      User           `json:"-" gorm:"foreignKey:UserID"`
	Token     string         `json:"-" gorm:"uniqueIndex;not null"` // Hash du token
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	IsUsed    bool           `json:"is_used" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// EmailVerificationToken représente un token de vérification d'email
type EmailVerificationToken struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    string         `json:"user_id" gorm:"not null"`
	User      User           `json:"-" gorm:"foreignKey:UserID"`
	Token     string         `json:"-" gorm:"uniqueIndex;not null"` // Hash du token
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	IsUsed    bool           `json:"is_used" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BlacklistedToken représente un token JWT mis en liste noire
type BlacklistedToken struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Token     string         `json:"-" gorm:"uniqueIndex;not null"` // Hash du token JWT
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// IsExpired vérifie si le token de réinitialisation est expiré
func (p *PasswordResetToken) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

// IsExpired vérifie si le token de vérification d'email est expiré
func (e *EmailVerificationToken) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}
