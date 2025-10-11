package biz

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AggregatedMetric represents a real-time aggregated metric in the business logic layer.
type AggregatedMetric struct {
	MetricName string
	Value      float64
	Timestamp  time.Time
	Labels     map[string]string
}

// RealtimeAnalyticsRepo defines the interface for real-time analytics data access.
type RealtimeAnalyticsRepo interface {
	SaveMetric(ctx context.Context, metric *AggregatedMetric) error
	GetMetric(ctx context.Context, metricName string) (*AggregatedMetric, error)
}

// RealtimeAnalyticsUsecase is the business logic for real-time analytics.
type RealtimeAnalyticsUsecase struct {
	repo              RealtimeAnalyticsRepo
	kafkaReader       *kafka.Reader // For consuming Kafka messages
	logger            *zap.Logger
	userProfileClient UserProfileClient // Added User Profile client
}

// UserProfileClient defines the interface to interact with the User Profile Service.
type UserProfileClient interface {
	UpdateUserProfile(ctx context.Context, profile *UserProfile) (*UserProfile, error)
	// GetUserProfile(ctx context.Context, userID string) (*UserProfile, error) // If needed
}

// UserProfile represents a simplified user profile for updates.
type UserProfile struct {
	UserID           string
	LastActiveTime   time.Time
	RecentCategories []string
	RecentBrands     []string
	// Add other fields that can be updated by real-time behavior
}

// NewRealtimeAnalyticsUsecase creates a new RealtimeAnalyticsUsecase.
func NewRealtimeAnalyticsUsecase(repo RealtimeAnalyticsRepo, kafkaReader *kafka.Reader, logger *zap.Logger, userProfileClient UserProfileClient) *RealtimeAnalyticsUsecase {
	uc := &RealtimeAnalyticsUsecase{
		repo:              repo,
		kafkaReader:       kafkaReader,
		logger:            logger,
		userProfileClient: userProfileClient,
	}
	go uc.startKafkaConsumer() // Start consuming Kafka messages in a goroutine
	return uc
}

// startKafkaConsumer starts consuming messages from Kafka and processes them.
func (uc *RealtimeAnalyticsUsecase) startKafkaConsumer() {
	uc.logger.Info("Starting Kafka consumer for real-time analytics...")
	for {
		m, err := uc.kafkaReader.ReadMessage(context.Background())
		if err != nil {
			uc.logger.Error("Error reading Kafka message", zap.Error(err))
			continue
		}

		uc.logger.Debug("Received Kafka message", zap.String("topic", m.Topic), zap.Int64("offset", m.Offset))

		// Simulate processing the message and updating metrics
		// In a real Flink/Storm-like system, this would be complex stream processing.
		var event map[string]interface{}
		if err := json.Unmarshal(m.Value, &event); err != nil {
			uc.logger.Error("Error unmarshaling Kafka message", zap.Error(err))
			continue
		}

		if eventType, ok := event["event_type"].(string); ok {
			if eventType == "user_behavior" {
				// Process user behavior events
				userID, _ := event["user_id"].(string)
				behaviorType, _ := event["behavior_type"].(string)
				itemID, _ := event["item_id"].(string)
				// properties, _ := event["properties"].(map[string]interface{}) // Properties might be complex

				// Simulate processing for real-time user activity
				if behaviorType == "VIEW" {
					// Increment page views per minute (simplified)
					pageViewsPerMinuteMetric, err := uc.repo.GetMetric(context.Background(), "page_views_per_minute")
					if err != nil {
						uc.logger.Error("Error getting page_views_per_minute metric", zap.Error(err))
					}
					currentViews := 0.0
					if pageViewsPerMinuteMetric != nil {
						currentViews = pageViewsPerMinuteMetric.Value
					}
					uc.repo.SaveMetric(context.Background(), &AggregatedMetric{
						MetricName: "page_views_per_minute",
						Value:      currentViews + 1,
						Timestamp:  time.Now(),
						Labels:     map[string]string{"item_id": itemID},
					})

					// Update user profile in HBase (simulated)
					if uc.userProfileClient != nil {
						profile := &UserProfile{
							UserID:         userID,
							LastActiveTime: time.Now(),
						}
						// Add recent categories/brands based on itemID (simplified)
						if itemID != "" {
							profile.RecentCategories = []string{"category_from_" + itemID}
							profile.RecentBrands = []string{"brand_from_" + itemID}
						}
						_, err := uc.userProfileClient.UpdateUserProfile(context.Background(), profile)
						if err != nil {
							uc.logger.Error("Error updating user profile in HBase", zap.Error(err))
						}
					}
				}
				uc.logger.Info("Processed user behavior event", zap.String("user_id", userID), zap.String("behavior_type", behaviorType))
			}
		}
	}
}

// GetRealtimeSalesMetrics retrieves real-time sales metrics.
func (uc *RealtimeAnalyticsUsecase) GetRealtimeSalesMetrics(ctx context.Context) (*RealtimeSalesMetrics, error) {
	gmvMetric, err := uc.repo.GetMetric(ctx, "current_gmv")
	if err != nil {
		return nil, err
	}
	gmv := 0.0
	if gmvMetric != nil {
		gmv = gmvMetric.Value
	}

	// Simulate other metrics
	orders := uint32(100)
	users := uint32(50)

	return &RealtimeSalesMetrics{
			CurrentGmv:    uint64(gmv),
			CurrentOrders: orders,
			ActiveUsers:   users,
		},
		nil
}

// GetRealtimeUserActivity retrieves real-time user activity metrics.
func (uc *RealtimeAnalyticsUsecase) GetRealtimeUserActivity(ctx context.Context) (*RealtimeUserActivity, error) {
	// Simulate metrics
	onlineUsers := uint32(200)
	newUsersLastHour := uint32(10)
	pageViewsPerMinute := map[string]uint32{"home": 50, "product_detail": 30}

	return &RealtimeUserActivity{
			OnlineUsers:        onlineUsers,
			NewUsersLastHour:   newUsersLastHour,
			PageViewsPerMinute: pageViewsPerMinute,
		},
		nil
}

// RecordUserBehavior records a user behavior event and processes it for real-time analytics.
func (uc *RealtimeAnalyticsUsecase) RecordUserBehavior(ctx context.Context, userID, behaviorType, itemID string, properties map[string]string, eventTime time.Time) error {
	// This method would typically be called by an RPC from the service layer.
	// Here, we simulate processing and updating metrics based on the behavior.

	// Example: Increment active users or page views
	if behaviorType == "VIEW" {
		// Simulate incrementing a page view counter
		pageViewMetric, err := uc.repo.GetMetric(ctx, "total_page_views")
		if err != nil {
			uc.logger.Error("Error getting total_page_views metric", zap.Error(err))
		}
		currentViews := 0.0
		if pageViewMetric != nil {
			currentViews = pageViewMetric.Value
		}
		uc.repo.SaveMetric(ctx, &AggregatedMetric{
			MetricName: "total_page_views",
			Value:      currentViews + 1,
			Timestamp:  time.Now(),
		})
	}

	// In a real system, this would also push to a stream for further processing
	// (e.g., to update user profiles in HBase, or for real-time recommendations).
	uc.logger.Info("User behavior recorded", zap.String("user_id", userID), zap.String("behavior_type", behaviorType), zap.String("item_id", itemID))
	return nil
}

// DTOs for RPC responses (should ideally be in service layer, but for simplicity here)
type RealtimeSalesMetrics struct {
	CurrentGmv    uint64
	CurrentOrders uint32
	ActiveUsers   uint32
	LastUpdated   *timestamppb.Timestamp
}

type RealtimeUserActivity struct {
	OnlineUsers        uint32
	NewUsersLastHour   uint32
	PageViewsPerMinute map[string]uint32
}
