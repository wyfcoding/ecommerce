package service

import (
	"context"
	"encoding/json"
	"time"

	"ecommerce/internal/realtime_analytics/client"
	"ecommerce/internal/realtime_analytics/model"
	"ecommerce/internal/realtime_analytics/repository"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// RealtimeAnalyticsService is the business logic for real-time analytics.
type RealtimeAnalyticsService struct {
	repo              repository.RealtimeAnalyticsRepo
	kafkaReader       *kafka.Reader // For consuming Kafka messages
	logger            *zap.Logger
	userProfileClient client.UserProfileClient // Added User Profile client
}

// NewRealtimeAnalyticsService creates a new RealtimeAnalyticsService.
func NewRealtimeAnalyticsService(repo repository.RealtimeAnalyticsRepo, kafkaReader *kafka.Reader, logger *zap.Logger, userProfileClient client.UserProfileClient) *RealtimeAnalyticsService {
	svc := &RealtimeAnalyticsService{
		repo:              repo,
		kafkaReader:       kafkaReader,
		logger:            logger,
		userProfileClient: userProfileClient,
	}
	go svc.startKafkaConsumer() // Start consuming Kafka messages in a goroutine
	return svc
}

// startKafkaConsumer starts consuming messages from Kafka and processes them.
func (s *RealtimeAnalyticsService) startKafkaConsumer() {
	s.logger.Info("Starting Kafka consumer for real-time analytics...")
	for {
		m, err := s.kafkaReader.ReadMessage(context.Background())
		if err != nil {
			s.logger.Error("Error reading Kafka message", zap.Error(err))
			continue
		}

		s.logger.Debug("Received Kafka message", zap.String("topic", m.Topic), zap.Int64("offset", m.Offset))

		// Simulate processing the message and updating metrics
		// In a real Flink/Storm-like system, this would be complex stream processing.
		var event map[string]interface{}
		if err := json.Unmarshal(m.Value, &event); err != nil {
			s.logger.Error("Error unmarshaling Kafka message", zap.Error(err))
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
					pageViewsPerMinuteMetric, err := s.repo.GetMetric(context.Background(), "page_views_per_minute")
					if err != nil {
						s.logger.Error("Error getting page_views_per_minute metric", zap.Error(err))
					}
					currentViews := 0.0
					if pageViewsPerMinuteMetric != nil {
						currentViews = pageViewsPerMinuteMetric.Value
					}
					s.repo.SaveMetric(context.Background(), &model.AggregatedMetric{
						MetricName: "page_views_per_minute",
						Value:      currentViews + 1,
						Timestamp:  time.Now(),
						Labels:     map[string]string{"item_id": itemID},
					})

					// Update user profile in HBase (simulated)
					if s.userProfileClient != nil {
						profile := &model.UserProfile{
							UserID:         userID,
							LastActiveTime: time.Now(),
						}
						// Add recent categories/brands based on itemID (simplified)
						if itemID != "" {
							profile.RecentCategories = []string{"category_from_" + itemID}
							profile.RecentBrands = []string{"brand_from_" + itemID}
						}
						_, err := s.userProfileClient.UpdateUserProfile(context.Background(), profile)
						if err != nil {
							s.logger.Error("Error updating user profile in HBase", zap.Error(err))
						}
					}
				}
				s.logger.Info("Processed user behavior event", zap.String("user_id", userID), zap.String("behavior_type", behaviorType))
			}
		}
	}
}

// GetRealtimeSalesMetrics retrieves real-time sales metrics.
func (s *RealtimeAnalyticsService) GetRealtimeSalesMetrics(ctx context.Context) (*model.RealtimeSalesMetrics, error) {
	gmvMetric, err := s.repo.GetMetric(ctx, "current_gmv")
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

	return &model.RealtimeSalesMetrics{
			CurrentGmv:    uint64(gmv),
			CurrentOrders: orders,
			ActiveUsers:   users,
		},
		nil
}

// GetRealtimeUserActivity retrieves real-time user activity metrics.
func (s *RealtimeAnalyticsService) GetRealtimeUserActivity(ctx context.Context) (*model.RealtimeUserActivity, error) {
	// Simulate metrics
	onlineUsers := uint32(200)
	newUsersLastHour := uint32(10)
	pageViewsPerMinute := map[string]uint32{"home": 50, "product_detail": 30}

	return &model.RealtimeUserActivity{
			OnlineUsers:        onlineUsers,
			NewUsersLastHour:   newUsersLastHour,
			PageViewsPerMinute: pageViewsPerMinute,
		},
		nil
}

// RecordUserBehavior records a user behavior event and processes it for real-time analytics.
func (s *RealtimeAnalyticsService) RecordUserBehavior(ctx context.Context, userID, behaviorType, itemID string, properties map[string]string, eventTime time.Time) error {
	// This method would typically be called by an RPC from the service layer.
	// Here, we simulate processing and updating metrics based on the behavior.

	// Example: Increment active users or page views
	if behaviorType == "VIEW" {
		// Simulate incrementing a page view counter
		pageViewMetric, err := s.repo.GetMetric(ctx, "total_page_views")
		if err != nil {
			s.logger.Error("Error getting total_page_views metric", zap.Error(err))
		}
		currentViews := 0.0
		if pageViewMetric != nil {
			currentViews = pageViewMetric.Value
		}
		s.repo.SaveMetric(ctx, &model.AggregatedMetric{
			MetricName: "total_page_views",
			Value:      currentViews + 1,
			Timestamp:  time.Now(),
		})
	}

	// In a real system, this would also push to a stream for further processing
	// (e.g., to update user profiles in HBase, or for real-time recommendations).
	s.logger.Info("User behavior recorded", zap.String("user_id", userID), zap.String("behavior_type", behaviorType), zap.String("item_id", itemID))
	return nil
}
