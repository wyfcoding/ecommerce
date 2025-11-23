package utils

import "math"

// Pagination represents pagination information.
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// NewPagination creates a new Pagination struct.
func NewPagination(page, pageSize int, totalItems int64) *Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))

	return &Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

// Offset returns the offset for database queries.
func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// HasNext checks if there is a next page.
func (p *Pagination) HasNext() bool {
	return p.Page < p.TotalPages
}

// HasPrev checks if there is a previous page.
func (p *Pagination) HasPrev() bool {
	return p.Page > 1
}
