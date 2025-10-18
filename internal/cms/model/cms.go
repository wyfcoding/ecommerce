package model

import (
	"time"
)

// ContentPage represents a content page in the business layer.
type ContentPage struct {
	ID          uint
	Title       string
	Slug        string // URL-friendly identifier
	ContentHTML string
	Status      string // e.g., DRAFT, PUBLISHED, ARCHIVED
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ContentBlock represents a reusable content block in the business layer.
type ContentBlock struct {
	ID          uint
	Name        string // Internal name for the block
	ContentHTML string
	Type        string // e.g., HTML, MARKDOWN, JSON
	CreatedAt   time.Time
	UpdatedAt   time.Time
}