package model

import "time"

// Ticket represents a customer service ticket in the business logic layer.
type Ticket struct {
	ID          uint
	TicketID    string
	UserID      uint64
	Subject     string
	Description string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Messages    []*TicketMessage // Populated when getting ticket details
}

// TicketMessage represents a message within a customer service ticket in the business logic layer.
type TicketMessage struct {
	ID         uint
	MessageID  string
	TicketID   string
	SenderID   uint64
	SenderType string
	Content    string
	CreatedAt  time.Time
}
