package model

import "time"

// Session represents a user session.
type Session struct {
	ID        uint
	UserID    uint64
	Token     string
	ExpiresAt time.Time
}

// User represents the user model from the user service.
type User struct {
	ID       uint64
	Username string
}
