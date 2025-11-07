package repository

import (
	"time"
)

// Product represents a product document in the recommendation system.
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
	// Add other fields relevant for recommendation, e.g., category, brand, tags, score
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProductRelationship represents a relationship between two products in Neo4j.
type ProductRelationship struct {
	ProductID1 string  `json:"product_id_1"`
	ProductID2 string  `json:"product_id_2"`
	Type       string  `json:"type"` // e.g., "BOUGHT_TOGETHER", "VIEWED_TOGETHER"
	Weight     float64 `json:"weight"`
}
