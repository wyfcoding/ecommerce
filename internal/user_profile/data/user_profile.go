package data

import (
	"context"
	"encoding/json"
	"ecommerce/internal/user_profile/biz"
	"ecommerce/internal/user_profile/data/model"
	"fmt"
	"time"

	"github.com/tsuna/go-hbase"
	"github.com/tsuna/go-hbase/hrpc"
)

const (
	userProfileTable = "user_profiles"
	userBehaviorTable = "user_behaviors"
	cfProfile        = "profile"
	cfBehavior       = "behavior"
)

type userProfileRepo struct {
	data *Data
	hbaseClient hbase.Client
}

// NewUserProfileRepo creates a new UserProfileRepo.
func NewUserProfileRepo(data *Data, hbaseClient hbase.Client) biz.UserProfileRepo {
	return &userProfileRepo{data: data, hbaseClient: hbaseClient}
}

// GetUserProfile retrieves a user profile from HBase.
func (r *userProfileRepo) GetUserProfile(ctx context.Context, userID string) (*biz.UserProfile, error) {
	get, err := hrpc.NewGetStr(ctx, userProfileTable, userID,
		hrpc.Families(map[string][]string{cfProfile: nil})) // Get all columns in 'profile' family
	if err != nil {
		return nil, fmt.Errorf("failed to create HBase Get request: %w", err)
	}

	resp, err := r.hbaseClient.Get(get)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile from HBase: %w", err)
	}
	if len(resp.Cells) == 0 {
		return nil, nil // Profile not found
	}

	profile := &biz.UserProfile{UserID: userID}
	for _, cell := range resp.Cells {
		col := string(cell.Qualifier)
		val := string(cell.Value)

		switch col {
		case "gender":
			profile.Gender = val
		case "age_group":
			profile.AgeGroup = val
		case "city":
			profile.City = val
		case "interests":
			json.Unmarshal([]byte(val), &profile.Interests)
		case "recent_categories":
			json.Unmarshal([]byte(val), &profile.RecentCategories)
		case "recent_brands":
			json.Unmarshal([]byte(val), &profile.RecentBrands)
		case "total_spent":
			fmt.Sscanf(val, "%d", &profile.TotalSpent)
		case "order_count":
			fmt.Sscanf(val, "%d", &profile.OrderCount)
		case "last_active_time":
			t, _ := time.Parse(time.RFC3339, val)
			profile.LastActiveTime = t
		case "custom_tags":
			json.Unmarshal([]byte(val), &profile.CustomTags)
		case "updated_at":
			t, _ := time.Parse(time.RFC3339, val)
			profile.UpdatedAt = t
		}
	}
	return profile, nil
}

// UpdateUserProfile updates a user profile in HBase.
func (r *userProfileRepo) UpdateUserProfile(ctx context.Context, profile *biz.UserProfile) (*biz.UserProfile, error) {
	puts := []*hrpc.MutateCell{}
	if profile.Gender != "" {
		puts = append(puts, hrpc.NewPutCell(cfProfile, "gender", []byte(profile.Gender)))
	}
	if profile.AgeGroup != "" {
		puts = append(puts, hrpc.NewPutCell(cfProfile, "age_group", []byte(profile.AgeGroup)))
	}
	if profile.City != "" {
		puts = append(puts, hrpc.NewPutCell(cfProfile, "city", []byte(profile.City)))
	}
	if len(profile.Interests) > 0 {
		b, _ := json.Marshal(profile.Interests)
		puts = append(puts, hrpc.NewPutCell(cfProfile, "interests", b))
	}
	if len(profile.RecentCategories) > 0 {
		b, _ := json.Marshal(profile.RecentCategories)
		puts = append(puts, hrpc.NewPutCell(cfProfile, "recent_categories", b))
	}
	if len(profile.RecentBrands) > 0 {
		b, _ := json.Marshal(profile.RecentBrands)
		puts = append(puts, hrpc.NewPutCell(cfProfile, "recent_brands", b))
	}
	if profile.TotalSpent > 0 {
		puts = append(puts, hrpc.NewPutCell(cfProfile, "total_spent", []byte(fmt.Sprintf("%d", profile.TotalSpent))))
	}
	if profile.OrderCount > 0 {
		puts = append(puts, hrpc.NewPutCell(cfProfile, "order_count", []byte(fmt.Sprintf("%d", profile.OrderCount))))
	}
	if !profile.LastActiveTime.IsZero() {
		puts = append(puts, hrpc.NewPutCell(cfProfile, "last_active_time", []byte(profile.LastActiveTime.Format(time.RFC3339))))
	}
	if len(profile.CustomTags) > 0 {
		b, _ := json.Marshal(profile.CustomTags)
		puts = append(puts, hrpc.NewPutCell(cfProfile, "custom_tags", b))
	}
	puts = append(puts, hrpc.NewPutCell(cfProfile, "updated_at", []byte(time.Now().Format(time.RFC3339))))

	put, err := hrpc.NewPutStr(ctx, userProfileTable, profile.UserID, puts)
	if err != nil {
		return nil, fmt.Errorf("failed to create HBase Put request: %w", err)
	}

	_, err = r.hbaseClient.Put(put)
	if err != nil {
		return nil, fmt.Errorf("failed to update user profile in HBase: %w", err)
	}
	return profile, nil
}

// RecordUserBehavior records a user behavior event to HBase.
func (r *userProfileRepo) RecordUserBehavior(ctx context.Context, event *biz.UserBehaviorEvent) error {
	// Row key for behavior events can be composite: UserID_Timestamp_BehaviorType
	rowKey := fmt.Sprintf("%s_%d_%s", event.UserID, event.EventTime.UnixNano(), event.BehaviorType)

	puts := []*hrpc.MutateCell{
		hrpc.NewPutCell(cfBehavior, "behavior_type", []byte(event.BehaviorType)),
		hrpc.NewPutCell(cfBehavior, "item_id", []byte(event.ItemID)),
		hrpc.NewPutCell(cfBehavior, "event_time", []byte(event.EventTime.Format(time.RFC3339))),
	}
	if len(event.Properties) > 0 {
		b, _ := json.Marshal(event.Properties)
		puts = append(puts, hrpc.NewPutCell(cfBehavior, "properties", b))
	}

	put, err := hrpc.NewPutStr(ctx, userBehaviorTable, rowKey, puts)
	if err != nil {
		return nil, fmt.Errorf("failed to create HBase Put request for behavior: %w", err)
	}

	_, err = r.hbaseClient.Put(put)
	if err != nil {
		return nil, fmt.Errorf("failed to record user behavior in HBase: %w", err)
	}
	return nil
}
