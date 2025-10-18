package model

import "time"

// Review represents a review entity in the business layer.
type Review struct {
	ID        uint
	ProductID string
	UserID    string
	Rating    int32 // 1-5 stars
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}