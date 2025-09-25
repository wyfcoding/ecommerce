package data

import (
	"time"

	"gorm.io/gorm"
)

// ReturnRequest is the database model for a return request.
type ReturnRequest struct {
	gorm.Model
	OrderID   string    `gorm:"type:varchar(100);not null;index"`
	UserID    string    `gorm:"type:varchar(100);not null;index"`
	ProductID string    `gorm:"type:varchar(100);not null;index"`
	Quantity  int32     `gorm:"not null"`
	Reason    string    `gorm:"type:text;not null"`
	Status    string    `gorm:"type:varchar(50);not null"` // e.g., PENDING, APPROVED, REJECTED, RECEIVED, REFUNDED
}

// TableName specifies the table name for the ReturnRequest model.
func (ReturnRequest) TableName() string {
	return "return_requests"
}

// RefundRequest is the database model for a refund request.
type RefundRequest struct {
	gorm.Model
	ReturnRequestID string    `gorm:"type:varchar(100);index"` // Optional: link to a return request
	OrderID         string    `gorm:"type:varchar(100);not null;index"`
	UserID          string    `gorm:"type:varchar(100);not null;index"`
	Amount          float64   `gorm:"not null"`
	Currency        string    `gorm:"type:varchar(10);not null"`
	Status          string    `gorm:"type:varchar(50);not null"` // e.g., PENDING, APPROVED, REJECTED, COMPLETED
}

// TableName specifies the table name for the RefundRequest model.
func (RefundRequest) TableName() string {
	return "refund_requests"
}
