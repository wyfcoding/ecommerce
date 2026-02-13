package domain

import (
	"context"
	"time"
)

type BehaviorFeatures struct {
	ID                   uint64             `json:"id"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
	UserID               uint64             `json:"user_id"`
	ProfileID            uint64             `json:"profile_id"`
	BrowseCount          int64              `json:"browse_count"`
	SearchCount          int64              `json:"search_count"`
	PurchaseCount        int64              `json:"purchase_count"`
	CartCount            int64              `json:"cart_count"`
	WishlistCount        int64              `json:"wishlist_count"`
	ShareCount           int64              `json:"share_count"`
	CommentCount         int64              `json:"comment_count"`
	ReviewCount          int64              `json:"review_count"`
	ReturnCount          int64              `json:"return_count"`
	CancelCount          int64              `json:"cancel_count"`
	AvgBrowseDuration    int64              `json:"avg_browse_duration"`
	TotalBrowseDuration  int64              `json:"total_browse_duration"`
	AvgSessionDuration   int64              `json:"avg_session_duration"`
	SessionCount         int64              `json:"session_count"`
	ActiveDays           int                `json:"active_days"`
	LastActiveAt         *time.Time         `json:"last_active_at"`
	FirstActiveAt        *time.Time         `json:"first_active_at"`
	PeakActiveHour       int                `json:"peak_active_hour"`
	PeakActiveDay        int                `json:"peak_active_day"`
	DeviceTypes          map[string]int     `json:"device_types"`
	Platforms            map[string]int     `json:"platforms"`
	BrowseCategories     map[uint64]int     `json:"browse_categories"`
	SearchKeywords       map[string]int     `json:"search_keywords"`
	ViewedProducts       map[uint64]int     `json:"viewed_products"`
	PurchasedCategories  map[uint64]int     `json:"purchased_categories"`
	PurchasedBrands      map[uint64]int     `json:"purchased_brands"`
	CartAbandonRate      float64            `json:"cart_abandon_rate"`
	ConversionRate       float64            `json:"conversion_rate"`
	ReturnRate           float64            `json:"return_rate"`
	RepeatPurchaseRate   float64            `json:"repeat_purchase_rate"`
	ActivityScore        int                `json:"activity_score"`
	EngagementScore      int                `json:"engagement_score"`
	BehaviorPatterns     []*BehaviorPattern `json:"behavior_patterns"`
	RecentBehaviors      []*RecentBehavior  `json:"recent_behaviors"`
}

type BehaviorPattern struct {
	ID           uint64    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	FeatureID    uint64    `json:"feature_id"`
	UserID       uint64    `json:"user_id"`
	PatternType  string    `json:"pattern_type"`
	PatternName  string    `json:"pattern_name"`
	Description  string    `json:"description"`
	Frequency    int       `json:"frequency"`
	Confidence   float64   `json:"confidence"`
	FirstSeenAt  *time.Time `json:"first_seen_at"`
	LastSeenAt   *time.Time `json:"last_seen_at"`
	Attributes   map[string]any `json:"attributes"`
}

type RecentBehavior struct {
	ID           uint64    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UserID       uint64    `json:"user_id"`
	BehaviorType string    `json:"behavior_type"`
	TargetType   string    `json:"target_type"`
	TargetID     uint64    `json:"target_id"`
	Value        string    `json:"value"`
	Duration     int64     `json:"duration"`
	Source       string    `json:"source"`
	IPAddress    string    `json:"ip_address"`
	DeviceID     string    `json:"device_id"`
	SessionID    string    `json:"session_id"`
}

type BehaviorType string

const (
	BehaviorTypeBrowse    BehaviorType = "BROWSE"
	BehaviorTypeSearch    BehaviorType = "SEARCH"
	BehaviorTypeAddToCart BehaviorType = "ADD_TO_CART"
	BehaviorTypePurchase  BehaviorType = "PURCHASE"
	BehaviorTypeShare     BehaviorType = "SHARE"
	BehaviorTypeComment   BehaviorType = "COMMENT"
	BehaviorTypeReview    BehaviorType = "REVIEW"
	BehaviorTypeWishlist  BehaviorType = "WISHLIST"
	BehaviorTypeReturn    BehaviorType = "RETURN"
	BehaviorTypeCancel    BehaviorType = "CANCEL"
)

func NewBehaviorFeatures(userID, profileID uint64) *BehaviorFeatures {
	return &BehaviorFeatures{
		UserID:              userID,
		ProfileID:           profileID,
		DeviceTypes:         make(map[string]int),
		Platforms:           make(map[string]int),
		BrowseCategories:    make(map[uint64]int),
		SearchKeywords:      make(map[string]int),
		ViewedProducts:      make(map[uint64]int),
		PurchasedCategories: make(map[uint64]int),
		PurchasedBrands:     make(map[uint64]int),
		BehaviorPatterns:    make([]*BehaviorPattern, 0),
		RecentBehaviors:     make([]*RecentBehavior, 0),
	}
}

func (f *BehaviorFeatures) RecordBrowse(productID, categoryID uint64, duration int64) {
	f.BrowseCount++
	f.TotalBrowseDuration += duration
	f.AvgBrowseDuration = f.TotalBrowseDuration / f.BrowseCount

	if productID > 0 {
		f.ViewedProducts[productID]++
	}
	if categoryID > 0 {
		f.BrowseCategories[categoryID]++
	}

	f.updateActivityScore()
}

func (f *BehaviorFeatures) RecordSearch(keyword string) {
	f.SearchCount++
	if keyword != "" {
		f.SearchKeywords[keyword]++
	}
	f.updateActivityScore()
}

func (f *BehaviorFeatures) RecordPurchase(categoryID, brandID uint64) {
	f.PurchaseCount++
	if categoryID > 0 {
		f.PurchasedCategories[categoryID]++
	}
	if brandID > 0 {
		f.PurchasedBrands[brandID]++
	}
	f.calculateRates()
	f.updateActivityScore()
}

func (f *BehaviorFeatures) RecordCart() {
	f.CartCount++
	f.calculateRates()
}

func (f *BehaviorFeatures) RecordReturn() {
	f.ReturnCount++
	f.calculateRates()
}

func (f *BehaviorFeatures) RecordCancel() {
	f.CancelCount++
}

func (f *BehaviorFeatures) RecordShare() {
	f.ShareCount++
	f.updateEngagementScore()
}

func (f *BehaviorFeatures) RecordComment() {
	f.CommentCount++
	f.updateEngagementScore()
}

func (f *BehaviorFeatures) RecordReview() {
	f.ReviewCount++
	f.updateEngagementScore()
}

func (f *BehaviorFeatures) RecordDevice(deviceType string) {
	f.DeviceTypes[deviceType]++
}

func (f *BehaviorFeatures) RecordPlatform(platform string) {
	f.Platforms[platform]++
}

func (f *BehaviorFeatures) RecordSession(duration int64) {
	f.SessionCount++
	f.TotalBrowseDuration += duration
	f.AvgSessionDuration = f.TotalBrowseDuration / f.SessionCount
}

func (f *BehaviorFeatures) SetActiveTime(firstActive, lastActive time.Time) {
	f.FirstActiveAt = &firstActive
	f.LastActiveAt = &lastActive

	days := int(lastActive.Sub(firstActive).Hours() / 24)
	if days > 0 {
		f.ActiveDays = days
	}
}

func (f *BehaviorFeatures) SetPeakTime(hour, day int) {
	f.PeakActiveHour = hour
	f.PeakActiveDay = day
}

func (f *BehaviorFeatures) calculateRates() {
	if f.CartCount > 0 {
		f.ConversionRate = float64(f.PurchaseCount) / float64(f.CartCount)
		f.CartAbandonRate = 1.0 - f.ConversionRate
	}

	if f.PurchaseCount > 0 {
		f.ReturnRate = float64(f.ReturnCount) / float64(f.PurchaseCount)
	}
}

func (f *BehaviorFeatures) updateActivityScore() {
	score := 0

	if f.BrowseCount > 100 {
		score += 30
	} else if f.BrowseCount > 50 {
		score += 20
	} else if f.BrowseCount > 10 {
		score += 10
	}

	if f.SearchCount > 50 {
		score += 20
	} else if f.SearchCount > 20 {
		score += 15
	} else if f.SearchCount > 5 {
		score += 10
	}

	if f.PurchaseCount > 10 {
		score += 30
	} else if f.PurchaseCount > 5 {
		score += 20
	} else if f.PurchaseCount > 0 {
		score += 10
	}

	if f.ActiveDays > 30 {
		score += 20
	} else if f.ActiveDays > 7 {
		score += 10
	}

	f.ActivityScore = score
}

func (f *BehaviorFeatures) updateEngagementScore() {
	score := 0

	if f.ShareCount > 20 {
		score += 30
	} else if f.ShareCount > 10 {
		score += 20
	} else if f.ShareCount > 0 {
		score += 10
	}

	if f.CommentCount > 30 {
		score += 25
	} else if f.CommentCount > 10 {
		score += 15
	} else if f.CommentCount > 0 {
		score += 5
	}

	if f.ReviewCount > 20 {
		score += 25
	} else if f.ReviewCount > 5 {
		score += 15
	} else if f.ReviewCount > 0 {
		score += 5
	}

	if f.WishlistCount > 50 {
		score += 20
	} else if f.WishlistCount > 20 {
		score += 10
	}

	f.EngagementScore = score
}

func (f *BehaviorFeatures) AddBehaviorPattern(pattern *BehaviorPattern) {
	f.BehaviorPatterns = append(f.BehaviorPatterns, pattern)
}

func (f *BehaviorFeatures) AddRecentBehavior(behavior *RecentBehavior) {
	f.RecentBehaviors = append(f.RecentBehaviors, behavior)

	if len(f.RecentBehaviors) > 100 {
		f.RecentBehaviors = f.RecentBehaviors[len(f.RecentBehaviors)-100:]
	}
}

func (f *BehaviorFeatures) GetTopCategories(limit int) []uint64 {
	type catCount struct {
		id    uint64
		count int
	}
	var categories []catCount
	for id, count := range f.BrowseCategories {
		categories = append(categories, catCount{id: id, count: count})
	}

	for i := 0; i < len(categories)-1; i++ {
		for j := i + 1; j < len(categories); j++ {
			if categories[j].count > categories[i].count {
				categories[i], categories[j] = categories[j], categories[i]
			}
		}
	}

	result := make([]uint64, 0, limit)
	for i := 0; i < len(categories) && i < limit; i++ {
		result = append(result, categories[i].id)
	}
	return result
}

func (f *BehaviorFeatures) GetTopKeywords(limit int) []string {
	type kwCount struct {
		keyword string
		count   int
	}
	var keywords []kwCount
	for kw, count := range f.SearchKeywords {
		keywords = append(keywords, kwCount{keyword: kw, count: count})
	}

	for i := 0; i < len(keywords)-1; i++ {
		for j := i + 1; j < len(keywords); j++ {
			if keywords[j].count > keywords[i].count {
				keywords[i], keywords[j] = keywords[j], keywords[i]
			}
		}
	}

	result := make([]string, 0, limit)
	for i := 0; i < len(keywords) && i < limit; i++ {
		result = append(result, keywords[i].keyword)
	}
	return result
}

func NewBehaviorPattern(featureID, userID uint64, patternType, patternName string) *BehaviorPattern {
	return &BehaviorPattern{
		FeatureID:   featureID,
		UserID:      userID,
		PatternType: patternType,
		PatternName: patternName,
		Attributes:  make(map[string]any),
	}
}

func (p *BehaviorPattern) RecordOccurrence() {
	p.Frequency++
	now := time.Now()
	if p.FirstSeenAt == nil {
		p.FirstSeenAt = &now
	}
	p.LastSeenAt = &now
}

func NewRecentBehavior(userID uint64, behaviorType BehaviorType, targetType string, targetID uint64) *RecentBehavior {
	return &RecentBehavior{
		UserID:       userID,
		BehaviorType: string(behaviorType),
		TargetType:   targetType,
		TargetID:     targetID,
	}
}

type BehaviorFeaturesRepository interface {
	Save(ctx context.Context, features *BehaviorFeatures) error
	FindByID(ctx context.Context, id uint64) (*BehaviorFeatures, error)
	FindByUserID(ctx context.Context, userID uint64) (*BehaviorFeatures, error)
	FindByProfileID(ctx context.Context, profileID uint64) (*BehaviorFeatures, error)
	Update(ctx context.Context, features *BehaviorFeatures) error
	Delete(ctx context.Context, userID uint64) error
}

type BehaviorPatternRepository interface {
	Save(ctx context.Context, pattern *BehaviorPattern) error
	FindByID(ctx context.Context, id uint64) (*BehaviorPattern, error)
	FindByFeatureID(ctx context.Context, featureID uint64) ([]*BehaviorPattern, error)
	FindByUserID(ctx context.Context, userID uint64) ([]*BehaviorPattern, error)
	Delete(ctx context.Context, id uint64) error
}

type RecentBehaviorRepository interface {
	Save(ctx context.Context, behavior *RecentBehavior) error
	FindByUserID(ctx context.Context, userID uint64, limit int) ([]*RecentBehavior, error)
	FindByType(ctx context.Context, userID uint64, behaviorType BehaviorType, limit int) ([]*RecentBehavior, error)
	DeleteOldBehaviors(ctx context.Context, before time.Time) error
}
