package data

import (
	"bytes"
	"context"
	"ecommerce/internal/search/biz" // Assuming this package exists
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8" // Changed to official v8 client
)

const (
	productIndex = "products"
)

type searchRepo struct {
	data     *Data
	esClient *elasticsearch.Client // Changed to official v8 client
}

// NewSearchRepo creates a new SearchRepo.
func NewSearchRepo(data *Data, esClient *elasticsearch.Client) biz.SearchRepo {
	return &searchRepo{
		data:     data,
		esClient: esClient,
	}
}

// SearchProducts searches for products in Elasticsearch.
func (r *searchRepo) SearchProducts(ctx context.Context, query string, pageSize, pageToken int32) ([]*biz.Product, int32, error) {
	// Build the search query in JSON format
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":       query,
				"fields":      []string{"name", "description"},
				"type":        "best_fields",
				"tie_breaker": 0.3,
			},
		},
		"from": pageToken,
		"size": pageSize,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, 0, fmt.Errorf("failed to encode search query: %w", err)
	}

	// Execute the search
	res, err := r.esClient.Search(
		r.esClient.Search.WithContext(ctx),
		r.esClient.Search.WithIndex(productIndex),
		r.esClient.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search products in Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, fmt.Errorf("failed to search products in Elasticsearch: %s", res.String())
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, 0, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Process search results
	products := make([]*biz.Product, 0)
	hits, found := r["hits"].(map[string]interface{})["hits"].([]interface{})
	if !found {
		return nil, 0, fmt.Errorf("no hits found in search response")
	}

	for _, hit := range hits {
		source, found := hit.(map[string]interface{})["_source"]
		if !found {
			continue
		}
		var p Product
		// Re-marshal and unmarshal to convert map[string]interface{} to Product struct
		jsonSource, err := json.Marshal(source)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal source: %w", err)
		}
		if err := json.Unmarshal(jsonSource, &p); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal product from Elasticsearch: %w", err)
		}
		products = append(products, r.toBizProduct(&p))
	}

	totalHits := int32(0)
	if total, found := r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64); found {
		totalHits = int32(total)
	}

	nextPageToken := pageToken + pageSize
	if nextPageToken >= totalHits {
		nextPageToken = 0 // No more pages
	}

	return products, totalHits, nil
}

// toBizProduct converts a data.Product to a biz.Product.
func (r *searchRepo) toBizProduct(p *Product) *biz.Product {
	if p == nil {
		return nil
	}
	return &biz.Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		ImageURL:    p.ImageURL,
	}
}
