package data

import (
	"context"
	"ecommerce/internal/search/biz"
	"ecommerce/pkg/elasticsearch" // Assuming this package exists
	"fmt"
	"encoding/json"

	"github.com/olivere/elastic/v7" // Assuming elastic.v7 client
)

const (
	productIndex = "products"
)

type searchRepo struct {
	data *Data
	esClient *elastic.Client
}

// NewSearchRepo creates a new SearchRepo.
func NewSearchRepo(data *Data, esClient *elastic.Client) biz.SearchRepo {
	return &searchRepo{
		data: data,
		esClient: esClient,
	}
}

// SearchProducts searches for products in Elasticsearch.
func (r *searchRepo) SearchProducts(ctx context.Context, query string, pageSize, pageToken int32) ([]*biz.Product, int32, error) {
	// Build the search query
	q := elastic.NewMultiMatchQuery(query, "name", "description").
		Type("best_fields").
		TieBreaker(0.3)

	// Execute the search
	searchResult, err := r.esClient.Search().
		Index(productIndex).
		Query(q).
		From(int(pageToken)).
		Size(int(pageSize)).
		Do(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search products in Elasticsearch: %w", err)
	}

	// Process search results
	products := make([]*biz.Product, 0)
	for _, hit := range searchResult.Hits.Hits {
		var p Product
		err := json.Unmarshal(hit.Source, &p) // Assuming Product struct matches ES document
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal product from Elasticsearch: %w", err)
		}
		products = append(products, r.toBizProduct(&p))
	}

	totalHits := int32(searchResult.Hits.TotalHits.Value)
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
