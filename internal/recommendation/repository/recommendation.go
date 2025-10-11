package data

import (
	"context"
	"ecommerce/internal/recommendation/biz"
	"ecommerce/internal/recommendation/data/model" // Added for model.ProductRelationship
	"fmt"                                          // Added for fmt.Errorf

	"github.com/neo4j/neo4j-go-driver/v5/neo4j" // Added
)

type recommendationRepo struct {
	data   *Data
	driver neo4j.Driver // Added Neo4j driver
}

// NewRecommendationRepo creates a new RecommendationRepo.
func NewRecommendationRepo(data *Data, driver neo4j.Driver) biz.RecommendationRepo {
	return &recommendationRepo{data: data, driver: driver}
}

// GetRecommendedProducts returns dummy recommended products.
func (r *recommendationRepo) GetRecommendedProducts(ctx context.Context, userID string, count int32) ([]*biz.Product, error) {
	// In a real application, this would involve:
	// 1. Calling a recommendation engine (e.g., based on user history, collaborative filtering, content-based filtering).
	// 2. Querying a database for popular products.
	// 3. Interacting with another service.

	// For now, return dummy data.
	products := []*biz.Product{
		{
			ID:          "rec_prod_1",
			Name:        "Recommended Product A",
			Description: "This is a highly recommended product A.",
			Price:       99.99,
			ImageURL:    "http://example.com/rec_prod_a.jpg",
		},
		{
			ID:          "rec_prod_2",
			Name:        "Recommended Product B",
			Description: "This is a highly recommended product B.",
			Price:       199.99,
			ImageURL:    "http://example.com/rec_prod_b.jpg",
		},
	}

	if int32(len(products)) > count {
		products = products[:count]
	}

	return products, nil
}

// SaveProductRelationshipToNeo4j saves a product relationship to Neo4j.
func (r *recommendationRepo) SaveProductRelationshipToNeo4j(ctx context.Context, rel *model.ProductRelationship) error {
	session := r.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	query := `
		MERGE (p1:Product {id: $productID1})
		MERGE (p2:Product {id: $productID2})
		MERGE (p1)-[r:` + rel.Type + `]->(p2)
		ON CREATE SET r.weight = $weight
		ON MATCH SET r.weight = r.weight + $weight
	`

	_, err := session.WriteTransaction(func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(query, map[string]interface{}{
			"productID1": rel.ProductID1,
			"productID2": rel.ProductID2,
			"weight":     rel.Weight,
		})
		return nil, err
	})

	if err != nil {
		return fmt.Errorf("failed to save product relationship to Neo4j: %w", err)
	}
	return nil
}

// GetRelatedProductsFromNeo4j gets related products from Neo4j.
func (r *recommendationRepo) GetRelatedProductsFromNeo4j(ctx context.Context, productID string, count int32) ([]*biz.Product, error) {
	session := r.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close()

	query := `
		MATCH (p1:Product {id: $productID})-[:BOUGHT_TOGETHER|VIEWED_TOGETHER]->(p2:Product)
		RETURN p2.id AS id, p2.name AS name, p2.description AS description, p2.price AS price, p2.image_url AS image_url
		LIMIT $count
	`

	result, err := session.ReadTransaction(func(tx neo4j.ManagedTransaction) (interface{}, error) {
		records, err := tx.Run(query, map[string]interface{}{
			"productID": productID,
			"count":     count,
		})
		if err != nil {
			return nil, err
		}

		var products []*biz.Product
		for records.Next() {
			record := records.Record()
			products = append(products, &biz.Product{
				ID:          record.Values[0].(string),
				Name:        record.Values[1].(string),
				Description: record.Values[2].(string),
				Price:       record.Values[3].(float64),
				ImageURL:    record.Values[4].(string),
			})
		}
		return products, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get related products from Neo4j: %w", err)
	}

	return result.([]*biz.Product), nil
}
