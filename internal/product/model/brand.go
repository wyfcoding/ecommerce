package model

import "time"

// Brand is a Brand model.
type Brand struct {
	ID          uint64
	Name        string
	Logo        *string
	Description *string
	Website     *string
	SortOrder   *uint32
	IsVisible   *bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
