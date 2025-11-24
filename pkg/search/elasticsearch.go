package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
)

// ESClient wraps elasticsearch.Client.
type ESClient struct {
	client *elasticsearch.Client
}

// NewESClient creates a new Elasticsearch client.
func NewESClient(addresses []string, username, password string) (*ESClient, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}
	return &ESClient{client: client}, nil
}

// Index indexes a document.
func (c *ESClient) Index(ctx context.Context, index, id string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	res, err := c.client.Index(
		index,
		bytes.NewReader(data),
		c.client.Index.WithContext(ctx),
		c.client.Index.WithDocumentID(id),
	)
	if err != nil {
		return fmt.Errorf("index error: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index failed: %s", res.String())
	}
	return nil
}

// Search searches for documents.
func (c *ESClient) Search(ctx context.Context, index string, query map[string]interface{}) (map[string]interface{}, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("encode query error: %w", err)
	}

	res, err := c.client.Search(
		c.client.Search.WithContext(ctx),
		c.client.Search.WithIndex(index),
		c.client.Search.WithBody(&buf),
		c.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("search error: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search failed: %s", res.String())
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return r, nil
}
