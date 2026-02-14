package domain

import (
	"context"
	"slices"
	"time"
)

type UserPreferences struct {
	ID                  uint64                   `json:"id"`
	CreatedAt           time.Time                `json:"created_at"`
	UpdatedAt           time.Time                `json:"updated_at"`
	UserID              uint64                   `json:"user_id"`
	ProfileID           uint64                   `json:"profile_id"`
	CategoryPreferences []*CategoryPreference    `json:"category_preferences"`
	BrandPreferences    []*BrandPreference       `json:"brand_preferences"`
	PricePreferences    *PricePreference         `json:"price_preferences"`
	TimePreferences     *TimePreference          `json:"time_preferences"`
	ChannelPreferences  []*ChannelPreference     `json:"channel_preferences"`
	PaymentPreferences  []*PaymentPreference     `json:"payment_preferences"`
	DeliveryPreferences *DeliveryPreference      `json:"delivery_preferences"`
	ContentPreferences  *ContentPreference       `json:"content_preferences"`
	CommunicationPrefs  *CommunicationPreference `json:"communication_prefs"`
	OverallStyle        string                   `json:"overall_style"`
	PreferenceStrength  float64                  `json:"preference_strength"`
	LastUpdatedSource   string                   `json:"last_updated_source"`
}

type CategoryPreference struct {
	ID             uint64     `json:"id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	PreferenceID   uint64     `json:"preference_id"`
	UserID         uint64     `json:"user_id"`
	CategoryID     uint64     `json:"category_id"`
	CategoryName   string     `json:"category_name"`
	CategoryPath   string     `json:"category_path"`
	Score          float64    `json:"score"`
	ViewCount      int64      `json:"view_count"`
	PurchaseCount  int64      `json:"purchase_count"`
	LastViewAt     *time.Time `json:"last_view_at"`
	LastPurchaseAt *time.Time `json:"last_purchase_at"`
	IsFavorite     bool       `json:"is_favorite"`
	Level          int        `json:"level"`
}

type BrandPreference struct {
	ID             uint64     `json:"id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	PreferenceID   uint64     `json:"preference_id"`
	UserID         uint64     `json:"user_id"`
	BrandID        uint64     `json:"brand_id"`
	BrandName      string     `json:"brand_name"`
	Score          float64    `json:"score"`
	ViewCount      int64      `json:"view_count"`
	PurchaseCount  int64      `json:"purchase_count"`
	LastViewAt     *time.Time `json:"last_view_at"`
	LastPurchaseAt *time.Time `json:"last_purchase_at"`
	IsFavorite     bool       `json:"is_favorite"`
	LoyaltyLevel   int        `json:"loyalty_level"`
}

type PricePreference struct {
	ID                 uint64    `json:"id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	PreferenceID       uint64    `json:"preference_id"`
	UserID             uint64    `json:"user_id"`
	MinPrice           int64     `json:"min_price"`
	MaxPrice           int64     `json:"max_price"`
	AvgPurchasePrice   int64     `json:"avg_purchase_price"`
	PriceSensitivity   float64   `json:"price_sensitivity"`
	DiscountPreference float64   `json:"discount_preference"`
	PriceRange         string    `json:"price_range"`
	PremiumRatio       float64   `json:"premium_ratio"`
	BudgetRatio        float64   `json:"budget_ratio"`
}

type TimePreference struct {
	ID             uint64    `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	PreferenceID   uint64    `json:"preference_id"`
	UserID         uint64    `json:"user_id"`
	PreferredHours []int     `json:"preferred_hours"`
	PreferredDays  []int     `json:"preferred_days"`
	PeakHour       int       `json:"peak_hour"`
	PeakDay        int       `json:"peak_day"`
	MorningRatio   float64   `json:"morning_ratio"`
	AfternoonRatio float64   `json:"afternoon_ratio"`
	EveningRatio   float64   `json:"evening_ratio"`
	NightRatio     float64   `json:"night_ratio"`
	WeekendRatio   float64   `json:"weekend_ratio"`
	WeekdayRatio   float64   `json:"weekday_ratio"`
	HolidayActive  bool      `json:"holiday_active"`
}

type ChannelPreference struct {
	ID           uint64     `json:"id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	PreferenceID uint64     `json:"preference_id"`
	UserID       uint64     `json:"user_id"`
	Channel      string     `json:"channel"`
	Score        float64    `json:"score"`
	VisitCount   int64      `json:"visit_count"`
	ConvertCount int64      `json:"convert_count"`
	LastVisitAt  *time.Time `json:"last_visit_at"`
	IsPreferred  bool       `json:"is_preferred"`
}

type PaymentPreference struct {
	ID           uint64     `json:"id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	PreferenceID uint64     `json:"preference_id"`
	UserID       uint64     `json:"user_id"`
	PaymentType  string     `json:"payment_type"`
	Score        float64    `json:"score"`
	UseCount     int64      `json:"use_count"`
	TotalAmount  int64      `json:"total_amount"`
	LastUsedAt   *time.Time `json:"last_used_at"`
	IsPreferred  bool       `json:"is_preferred"`
}

type DeliveryPreference struct {
	ID              uint64    `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PreferenceID    uint64    `json:"preference_id"`
	UserID          uint64    `json:"user_id"`
	PreferredMethod string    `json:"preferred_method"`
	ExpressRatio    float64   `json:"express_ratio"`
	StandardRatio   float64   `json:"standard_ratio"`
	EconomyRatio    float64   `json:"economy_ratio"`
	SameDayRatio    float64   `json:"same_day_ratio"`
	PickupRatio     float64   `json:"pickup_ratio"`
	SpeedPriority   int       `json:"speed_priority"`
	CostPriority    int       `json:"cost_priority"`
	InsuranceRatio  float64   `json:"insurance_ratio"`
}

type ContentPreference struct {
	ID             uint64     `json:"id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	PreferenceID   uint64     `json:"preference_id"`
	UserID         uint64     `json:"user_id"`
	ContentType    string     `json:"content_type"`
	ViewCount      int64      `json:"view_count"`
	LikeCount      int64      `json:"like_count"`
	ShareCount     int64      `json:"share_count"`
	CommentCount   int64      `json:"comment_count"`
	Score          float64    `json:"score"`
	LastInteractAt *time.Time `json:"last_interact_at"`
}

type CommunicationPreference struct {
	ID                uint64    `json:"id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	PreferenceID      uint64    `json:"preference_id"`
	UserID            uint64    `json:"user_id"`
	EmailEnabled      bool      `json:"email_enabled"`
	SmsEnabled        bool      `json:"sms_enabled"`
	PushEnabled       bool      `json:"push_enabled"`
	InAppEnabled      bool      `json:"in_app_enabled"`
	MarketingEnabled  bool      `json:"marketing_enabled"`
	OrderUpdates      bool      `json:"order_updates"`
	PromotionAlerts   bool      `json:"promotion_alerts"`
	PriceDropAlerts   bool      `json:"price_drop_alerts"`
	ReviewReminders   bool      `json:"review_reminders"`
	PreferredLanguage string    `json:"preferred_language"`
	QuietHoursStart   int       `json:"quiet_hours_start"`
	QuietHoursEnd     int       `json:"quiet_hours_end"`
}

func NewUserPreferences(userID, profileID uint64) *UserPreferences {
	return &UserPreferences{
		UserID:              userID,
		ProfileID:           profileID,
		CategoryPreferences: make([]*CategoryPreference, 0),
		BrandPreferences:    make([]*BrandPreference, 0),
		ChannelPreferences:  make([]*ChannelPreference, 0),
		PaymentPreferences:  make([]*PaymentPreference, 0),
	}
}

func (p *UserPreferences) AddCategoryPreference(pref *CategoryPreference) {
	for i, cp := range p.CategoryPreferences {
		if cp.CategoryID == pref.CategoryID {
			p.CategoryPreferences[i] = pref
			return
		}
	}
	p.CategoryPreferences = append(p.CategoryPreferences, pref)
}

func (p *UserPreferences) AddBrandPreference(pref *BrandPreference) {
	for i, bp := range p.BrandPreferences {
		if bp.BrandID == pref.BrandID {
			p.BrandPreferences[i] = pref
			return
		}
	}
	p.BrandPreferences = append(p.BrandPreferences, pref)
}

func (p *UserPreferences) GetTopCategories(limit int) []*CategoryPreference {
	if len(p.CategoryPreferences) <= limit {
		return p.CategoryPreferences
	}

	sorted := make([]*CategoryPreference, len(p.CategoryPreferences))
	copy(sorted, p.CategoryPreferences)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Score > sorted[i].Score {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted[:limit]
}

func (p *UserPreferences) GetTopBrands(limit int) []*BrandPreference {
	if len(p.BrandPreferences) <= limit {
		return p.BrandPreferences
	}

	sorted := make([]*BrandPreference, len(p.BrandPreferences))
	copy(sorted, p.BrandPreferences)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Score > sorted[i].Score {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted[:limit]
}

func (p *UserPreferences) CalculatePreferenceStrength() {
	var totalScore float64
	var count int

	for _, cp := range p.CategoryPreferences {
		totalScore += cp.Score
		count++
	}

	for _, bp := range p.BrandPreferences {
		totalScore += bp.Score
		count++
	}

	if count > 0 {
		p.PreferenceStrength = totalScore / float64(count)
	}
}

func NewCategoryPreference(preferenceID, userID, categoryID uint64, categoryName string) *CategoryPreference {
	return &CategoryPreference{
		PreferenceID: preferenceID,
		UserID:       userID,
		CategoryID:   categoryID,
		CategoryName: categoryName,
	}
}

func (p *CategoryPreference) RecordView() {
	p.ViewCount++
	now := time.Now()
	p.LastViewAt = &now
	p.calculateScore()
}

func (p *CategoryPreference) RecordPurchase() {
	p.PurchaseCount++
	now := time.Now()
	p.LastPurchaseAt = &now
	p.calculateScore()
}

func (p *CategoryPreference) calculateScore() {
	viewScore := float64(p.ViewCount) * 0.3
	purchaseScore := float64(p.PurchaseCount) * 0.7
	p.Score = viewScore + purchaseScore
}

func NewBrandPreference(preferenceID, userID, brandID uint64, brandName string) *BrandPreference {
	return &BrandPreference{
		PreferenceID: preferenceID,
		UserID:       userID,
		BrandID:      brandID,
		BrandName:    brandName,
	}
}

func (p *BrandPreference) RecordView() {
	p.ViewCount++
	now := time.Now()
	p.LastViewAt = &now
	p.calculateScore()
}

func (p *BrandPreference) RecordPurchase() {
	p.PurchaseCount++
	now := time.Now()
	p.LastPurchaseAt = &now
	p.calculateScore()
}

func (p *BrandPreference) calculateScore() {
	viewScore := float64(p.ViewCount) * 0.3
	purchaseScore := float64(p.PurchaseCount) * 0.7
	p.Score = viewScore + purchaseScore
}

func NewPricePreference(preferenceID, userID uint64) *PricePreference {
	return &PricePreference{
		PreferenceID: preferenceID,
		UserID:       userID,
	}
}

func (p *PricePreference) UpdatePriceRange(minPrice, maxPrice, avgPrice int64) {
	p.MinPrice = minPrice
	p.MaxPrice = maxPrice
	p.AvgPurchasePrice = avgPrice

	if avgPrice > 0 {
		if avgPrice > 100000 {
			p.PriceRange = "PREMIUM"
		} else if avgPrice > 50000 {
			p.PriceRange = "HIGH"
		} else if avgPrice > 10000 {
			p.PriceRange = "MEDIUM"
		} else {
			p.PriceRange = "BUDGET"
		}
	}
}

func NewTimePreference(preferenceID, userID uint64) *TimePreference {
	return &TimePreference{
		PreferenceID:   preferenceID,
		UserID:         userID,
		PreferredHours: make([]int, 0),
		PreferredDays:  make([]int, 0),
	}
}

func (t *TimePreference) SetPeakTime(hour, day int) {
	t.PeakHour = hour
	t.PeakDay = day
}

func (t *TimePreference) AddPreferredHour(hour int) {
	if slices.Contains(t.PreferredHours, hour) {
		return
	}
	t.PreferredHours = append(t.PreferredHours, hour)
}

func (t *TimePreference) AddPreferredDay(day int) {
	if slices.Contains(t.PreferredDays, day) {
		return
	}
	t.PreferredDays = append(t.PreferredDays, day)
}

type UserPreferencesRepository interface {
	Save(ctx context.Context, preferences *UserPreferences) error
	FindByID(ctx context.Context, id uint64) (*UserPreferences, error)
	FindByUserID(ctx context.Context, userID uint64) (*UserPreferences, error)
	FindByProfileID(ctx context.Context, profileID uint64) (*UserPreferences, error)
	Update(ctx context.Context, preferences *UserPreferences) error
	Delete(ctx context.Context, userID uint64) error
}

type CategoryPreferenceRepository interface {
	Save(ctx context.Context, pref *CategoryPreference) error
	FindByUserID(ctx context.Context, userID uint64) ([]*CategoryPreference, error)
	FindByCategoryID(ctx context.Context, categoryID uint64, limit int) ([]*CategoryPreference, error)
	Delete(ctx context.Context, id uint64) error
}

type BrandPreferenceRepository interface {
	Save(ctx context.Context, pref *BrandPreference) error
	FindByUserID(ctx context.Context, userID uint64) ([]*BrandPreference, error)
	FindByBrandID(ctx context.Context, brandID uint64, limit int) ([]*BrandPreference, error)
	Delete(ctx context.Context, id uint64) error
}
