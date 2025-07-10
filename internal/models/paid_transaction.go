package models

import (
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/models"
	"gorm.io/gorm"
)

type PaidMessageTransaction struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	MessageID      uint            `gorm:"not null;index" json:"message_id"`
	Message        EnhancedMessage `gorm:"foreignKey:MessageID" json:"message"`
	BuyerID        string          `gorm:"not null;index" json:"buyer_id"`
	Buyer          models.User     `gorm:"foreignKey:BuyerID" json:"buyer"`
	SellerID       string          `gorm:"not null;index" json:"seller_id"`
	Seller         models.User     `gorm:"foreignKey:SellerID" json:"seller"`
	Amount         float64         `gorm:"type:decimal(10,2);not null" json:"amount"`
	PlatformFee    float64         `gorm:"type:decimal(10,2);not null" json:"platform_fee"` // 20%
	SellerEarnings float64         `gorm:"type:decimal(10,2);not null" json:"seller_earnings"`
	Status         string          `gorm:"default:pending;index" json:"status"` // 'pending', 'completed', 'failed', 'refunded'
	PaymentMethod  *string         `json:"payment_method"`
	TransactionID  *string         `gorm:"index" json:"transaction_id"`
	CreatedAt      time.Time       `json:"created_at"`
	CompletedAt    *time.Time      `json:"completed_at"`
}

func (PaidMessageTransaction) TableName() string {
	return "paid_message_transactions"
}

// Hook pour calculer commission automatiquement
func (pmt *PaidMessageTransaction) BeforeCreate(tx *gorm.DB) error {
	pmt.PlatformFee = pmt.Amount * 0.20 // 20% commission OnlyFlick
	pmt.SellerEarnings = pmt.Amount - pmt.PlatformFee
	return nil
}
