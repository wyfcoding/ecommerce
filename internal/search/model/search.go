package model

// Product represents a product in the business logic layer.
type Product struct {
	ID          string
	Name        string
	Description string
	Price       float64
	ImageURL    string
}
