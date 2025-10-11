package data

import (
	"context"
	"ecommerce/internal/analytics/biz"
	"ecommerce/internal/analytics/data/model"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type analyticsRepo struct {
	data     *Data
	chClient clickhouse.Conn
}

// NewAnalyticsRepo creates a new AnalyticsRepo.
func NewAnalyticsRepo(data *Data, chClient clickhouse.Conn) biz.AnalyticsRepo {
	return &analyticsRepo{data: data, chClient: chClient}
}

// RecordProductView records a product view event to ClickHouse.
func (r *analyticsRepo) RecordProductView(ctx context.Context, event *biz.ProductViewEvent) error {
	// Prepare the event for ClickHouse
	chEvent := model.ProductViewEvent{
		UserID:    event.UserID,
		ProductID: event.ProductID,
		ViewTime:  event.ViewTime,
	}

	// Insert into ClickHouse
	// Assuming a table named 'product_views' exists in ClickHouse
	// CREATE TABLE product_views (user_id UInt64, product_id UInt64, view_time DateTime) ENGINE = MergeTree() ORDER BY view_time;
	err := r.chClient.AsyncInsert(ctx, "INSERT INTO product_views (user_id, product_id, view_time) VALUES (?, ?, ?)", false, chEvent.UserID, chEvent.ProductID, chEvent.ViewTime)
	if err != nil {
		return fmt.Errorf("failed to insert product view event into ClickHouse: %w", err)
	}
	return nil
}
