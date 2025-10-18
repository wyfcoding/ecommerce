package model

// Product represents a product in the business logic layer.
type Product struct {
	ID          string
	Name        string
	Description string
	Price       float64
	ImageURL    string
}

// ProductRelationship represents a relationship between two products in Neo4j.
type ProductRelationship struct {
	ProductID1 string
	ProductID2 string
	Type       string
	Weight     float64
}
