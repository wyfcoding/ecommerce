package data

import (
	"context"
	"ecommerce/internal/flashsale/biz"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"gorm.io/gorm"

	orderv1 "ecommerce/api/order/v1"
)

// ProviderSet is data providers.
// Using Google's 'wire' tool for dependency injection is a common practice in Go projects.
var ProviderSet = wire.NewSet(NewData, NewFlashSaleRepo, NewDistributedLocker, NewOrderServiceClient)

// Data struct holds the database and Redis clients.
type Data struct {
	db          *gorm.DB
	redisClient *redis.Client
}

// NewData creates a new Data struct with database and Redis connections.
func NewData(db *gorm.DB, redisAddr, redisPassword string) (*Data, func(), error) {
	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0, // use default DB
		PoolSize: 100,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	log.Println("Successfully connected to Redis.")

	// This cleanup function will be called when the service shuts down.
	cleanup := func() {
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("failed to get underlying sql.DB: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
		log.Println("database connection closed")

		if err := rdb.Close(); err != nil {
			log.Printf("failed to close Redis client: %v", err)
		}
		log.Println("Redis client closed")
	}

	d := &Data{
		db:          db,
		redisClient: rdb,
	}
	return d, cleanup, nil
}

// NewFlashSaleRepo is a provider function that creates a new flash sale repository.
// It depends on the Data struct (which has the db connection).
func NewFlashSaleRepo(data *Data) biz.FlashSaleRepo {
	// The actual implementation is in flashsale.go
	return &flashSaleRepo{
		data: data,
	}
}

// NewDistributedLocker creates a new Redis-based distributed locker.
func NewDistributedLocker(data *Data) biz.DistributedLocker {
	return &redisLocker{
		redisClient: data.redisClient,
	}
}

// orderServiceClient implements biz.OrderServiceClient using gRPC.
type orderServiceClient struct {
	client orderv1.OrderClient
}

// NewOrderServiceClient creates a new OrderServiceClient.
func NewOrderServiceClient(client orderv1.OrderClient) biz.OrderServiceClient {
	return &orderServiceClient{
		client: client,
	}
}

// CreateOrderForFlashSale calls the Order Service to create an order for a flash sale.
func (osc *orderServiceClient) CreateOrderForFlashSale(ctx context.Context, userID, productID string, quantity int32, price float64) (string, error) {
	resp, err := osc.client.CreateOrderForFlashSale(ctx, &orderv1.CreateOrderForFlashSaleRequest{
		UserId:    userID,
		ProductId: productID,
		Quantity:  quantity,
		Price:     price,
	})
	if err != nil {
		return "", err
	}
	return resp.OrderId, nil
}

// CompensateCreateOrder calls the Order Service to compensate a previously created order.
func (osc *orderServiceClient) CompensateCreateOrder(ctx context.Context, orderID string) error {
	_, err := osc.client.CompensateCreateOrder(ctx, &orderv1.CompensateCreateOrderRequest{
		OrderId: orderID,
	})
	return err
}
