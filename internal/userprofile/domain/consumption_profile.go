package domain

import (
	"context"
	"time"
)

type ConsumptionProfile struct {
	ID                     uint64              `json:"id"`
	CreatedAt              time.Time           `json:"created_at"`
	UpdatedAt              time.Time           `json:"updated_at"`
	UserID                 uint64              `json:"user_id"`
	ProfileID              uint64              `json:"profile_id"`
	TotalSpent             int64               `json:"total_spent"`
	TotalOrders            int64               `json:"total_orders"`
	AvgOrderValue          int64               `json:"avg_order_value"`
	MaxOrderValue          int64               `json:"max_order_value"`
	MinOrderValue          int64               `json:"min_order_value"`
	MedianOrderValue       int64               `json:"median_order_value"`
	TotalRefunded          int64               `json:"total_refunded"`
	RefundCount            int64               `json:"refund_count"`
	TotalDiscount          int64               `json:"total_discount"`
	DiscountUsageCount     int64               `json:"discount_usage_count"`
	CouponUsageRate        float64             `json:"coupon_usage_rate"`
	AvgDiscountRate        float64             `json:"avg_discount_rate"`
	SpendingLevel          SpendingLevel       `json:"spending_level"`
	ConsumptionCapacity    int                 `json:"consumption_capacity"`
	ConsumptionFrequency   ConsumptionFrequency `json:"consumption_frequency"`
	PurchaseCycle          int                 `json:"purchase_cycle"`
	PredictedMonthlySpend  int64               `json:"predicted_monthly_spend"`
	PredictedYearlySpend   int64               `json:"predicted_yearly_spend"`
	LTV                    int64               `json:"ltv"`
	LTV12Months            int64               `json:"ltv_12_months"`
	FirstPurchaseAt        *time.Time          `json:"first_purchase_at"`
	LastPurchaseAt         *time.Time          `json:"last_purchase_at"`
	MonthlySpending        map[string]int64    `json:"monthly_spending"`
	CategorySpending       map[uint64]int64    `json:"category_spending"`
	BrandSpending          map[uint64]int64    `json:"brand_spending"`
	PaymentMethodSpending  map[string]int64    `json:"payment_method_spending"`
	SpendingTrend          SpendingTrend       `json:"spending_trend"`
	GrowthRate             float64             `json:"growth_rate"`
	ChurnRisk              float64             `json:"churn_risk"`
	ValueSegment           ValueSegment        `json:"value_segment"`
	RFM                    *RFMScore           `json:"rfm"`
	PurchasePatterns       []*PurchasePattern  `json:"purchase_patterns"`
}

type SpendingLevel int8

const (
	SpendingLevelLow      SpendingLevel = 1
	SpendingLevelMedium   SpendingLevel = 2
	SpendingLevelHigh     SpendingLevel = 3
	SpendingLevelPremium  SpendingLevel = 4
	SpendingLevelVIP      SpendingLevel = 5
)

func (l SpendingLevel) String() string {
	switch l {
	case SpendingLevelLow:
		return "LOW"
	case SpendingLevelMedium:
		return "MEDIUM"
	case SpendingLevelHigh:
		return "HIGH"
	case SpendingLevelPremium:
		return "PREMIUM"
	case SpendingLevelVIP:
		return "VIP"
	default:
		return "UNKNOWN"
	}
}

type ConsumptionFrequency int8

const (
	FrequencyInactive   ConsumptionFrequency = 0
	FrequencyRare       ConsumptionFrequency = 1
	FrequencyOccasional ConsumptionFrequency = 2
	FrequencyRegular    ConsumptionFrequency = 3
	FrequencyFrequent   ConsumptionFrequency = 4
	FrequencyHeavy      ConsumptionFrequency = 5
)

func (f ConsumptionFrequency) String() string {
	switch f {
	case FrequencyInactive:
		return "INACTIVE"
	case FrequencyRare:
		return "RARE"
	case FrequencyOccasional:
		return "OCCASIONAL"
	case FrequencyRegular:
		return "REGULAR"
	case FrequencyFrequent:
		return "FREQUENT"
	case FrequencyHeavy:
		return "HEAVY"
	default:
		return "UNKNOWN"
	}
}

type SpendingTrend int8

const (
	TrendDeclining  SpendingTrend = -1
	TrendStable     SpendingTrend = 0
	TrendGrowing    SpendingTrend = 1
	TrendRapidGrowth SpendingTrend = 2
)

func (t SpendingTrend) String() string {
	switch t {
	case TrendDeclining:
		return "DECLINING"
	case TrendStable:
		return "STABLE"
	case TrendGrowing:
		return "GROWING"
	case TrendRapidGrowth:
		return "RAPID_GROWTH"
	default:
		return "UNKNOWN"
	}
}

type ValueSegment int8

const (
	SegmentChurned    ValueSegment = 1
	SegmentAtRisk     ValueSegment = 2
	SegmentNew        ValueSegment = 3
	SegmentPotential  ValueSegment = 4
	SegmentLoyal      ValueSegment = 5
	SegmentChampion   ValueSegment = 6
	SegmentVIP        ValueSegment = 7
)

func (s ValueSegment) String() string {
	switch s {
	case SegmentChurned:
		return "CHURNED"
	case SegmentAtRisk:
		return "AT_RISK"
	case SegmentNew:
		return "NEW"
	case SegmentPotential:
		return "POTENTIAL"
	case SegmentLoyal:
		return "LOYAL"
	case SegmentChampion:
		return "CHAMPION"
	case SegmentVIP:
		return "VIP"
	default:
		return "UNKNOWN"
	}
}

type RFMScore struct {
	ID            uint64    `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ProfileID     uint64    `json:"profile_id"`
	UserID        uint64    `json:"user_id"`
	RecencyScore  int       `json:"recency_score"`
	FrequencyScore int      `json:"frequency_score"`
	MonetaryScore int       `json:"monetary_score"`
	RecencyDays   int       `json:"recency_days"`
	FrequencyCount int      `json:"frequency_count"`
	MonetaryValue int64     `json:"monetary_value"`
	TotalScore    int       `json:"total_score"`
	Segment       string    `json:"segment"`
	LastPurchaseAt *time.Time `json:"last_purchase_at"`
}

type PurchasePattern struct {
	ID              uint64    `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	ProfileID       uint64    `json:"profile_id"`
	UserID          uint64    `json:"user_id"`
	PatternType     string    `json:"pattern_type"`
	PatternName     string    `json:"pattern_name"`
	Description     string    `json:"description"`
	Frequency       int       `json:"frequency"`
	AvgAmount       int64     `json:"avg_amount"`
	Confidence      float64   `json:"confidence"`
	FirstObservedAt *time.Time `json:"first_observed_at"`
	LastObservedAt  *time.Time `json:"last_observed_at"`
	Attributes      map[string]any `json:"attributes"`
}

func NewConsumptionProfile(userID, profileID uint64) *ConsumptionProfile {
	return &ConsumptionProfile{
		UserID:                userID,
		ProfileID:             profileID,
		MonthlySpending:       make(map[string]int64),
		CategorySpending:      make(map[uint64]int64),
		BrandSpending:         make(map[uint64]int64),
		PaymentMethodSpending: make(map[string]int64),
		PurchasePatterns:      make([]*PurchasePattern, 0),
	}
}

func (p *ConsumptionProfile) RecordPurchase(amount int64, categoryID, brandID uint64, paymentMethod string, purchasedAt time.Time) {
	p.TotalOrders++
	p.TotalSpent += amount

	if p.MinOrderValue == 0 || amount < p.MinOrderValue {
		p.MinOrderValue = amount
	}
	if amount > p.MaxOrderValue {
		p.MaxOrderValue = amount
	}

	p.AvgOrderValue = p.TotalSpent / p.TotalOrders

	if p.FirstPurchaseAt == nil {
		p.FirstPurchaseAt = &purchasedAt
	}
	p.LastPurchaseAt = &purchasedAt

	if categoryID > 0 {
		p.CategorySpending[categoryID] += amount
	}
	if brandID > 0 {
		p.BrandSpending[brandID] += amount
	}
	if paymentMethod != "" {
		p.PaymentMethodSpending[paymentMethod] += amount
	}

	monthKey := purchasedAt.Format("2006-01")
	p.MonthlySpending[monthKey] += amount

	p.calculateSpendingLevel()
	p.calculateFrequency()
	p.calculateLTV()
}

func (p *ConsumptionProfile) RecordRefund(amount int64) {
	p.TotalRefunded += amount
	p.RefundCount++
}

func (p *ConsumptionProfile) RecordDiscount(discountAmount int64) {
	p.TotalDiscount += discountAmount
	p.DiscountUsageCount++

	if p.TotalOrders > 0 {
		p.CouponUsageRate = float64(p.DiscountUsageCount) / float64(p.TotalOrders)
	}
	if p.TotalSpent > 0 {
		p.AvgDiscountRate = float64(p.TotalDiscount) / float64(p.TotalSpent)
	}
}

func (p *ConsumptionProfile) calculateSpendingLevel() {
	avgMonthly := p.AvgOrderValue * int64(p.ConsumptionFrequency)

	switch {
	case avgMonthly >= 100000:
		p.SpendingLevel = SpendingLevelVIP
	case avgMonthly >= 50000:
		p.SpendingLevel = SpendingLevelPremium
	case avgMonthly >= 20000:
		p.SpendingLevel = SpendingLevelHigh
	case avgMonthly >= 5000:
		p.SpendingLevel = SpendingLevelMedium
	default:
		p.SpendingLevel = SpendingLevelLow
	}
}

func (p *ConsumptionProfile) calculateFrequency() {
	if p.LastPurchaseAt == nil {
		p.ConsumptionFrequency = FrequencyInactive
		return
	}

	daysSinceLastPurchase := int(time.Since(*p.LastPurchaseAt).Hours() / 24)

	if daysSinceLastPurchase > 180 {
		p.ConsumptionFrequency = FrequencyInactive
	} else if daysSinceLastPurchase > 90 {
		p.ConsumptionFrequency = FrequencyRare
	} else if p.TotalOrders <= 3 {
		p.ConsumptionFrequency = FrequencyOccasional
	} else if p.TotalOrders <= 10 {
		p.ConsumptionFrequency = FrequencyRegular
	} else if p.TotalOrders <= 30 {
		p.ConsumptionFrequency = FrequencyFrequent
	} else {
		p.ConsumptionFrequency = FrequencyHeavy
	}

	if p.TotalOrders > 1 && p.FirstPurchaseAt != nil {
		days := int(time.Since(*p.FirstPurchaseAt).Hours() / 24)
		if days > 0 {
			p.PurchaseCycle = days / int(p.TotalOrders)
		}
	}
}

func (p *ConsumptionProfile) calculateLTV() {
	if p.FirstPurchaseAt == nil {
		return
	}

	months := time.Since(*p.FirstPurchaseAt).Hours() / 24 / 30
	if months > 0 {
		monthlyAvg := float64(p.TotalSpent) / months
		p.PredictedMonthlySpend = int64(monthlyAvg)
		p.PredictedYearlySpend = int64(monthlyAvg * 12)
		p.LTV = int64(monthlyAvg * 24)
	}

	p.LTV12Months = p.TotalSpent
}

func (p *ConsumptionProfile) CalculateTrend() {
	if len(p.MonthlySpending) < 2 {
		p.SpendingTrend = TrendStable
		return
	}

	var months []string
	for m := range p.MonthlySpending {
		months = append(months, m)
	}

	for i := 0; i < len(months)-1; i++ {
		for j := i + 1; j < len(months); j++ {
			if months[i] > months[j] {
				months[i], months[j] = months[j], months[i]
			}
		}
	}

	recentMonths := months
	if len(recentMonths) > 3 {
		recentMonths = recentMonths[len(recentMonths)-3:]
	}

	var recentTotal, previousTotal int64
	for i, m := range recentMonths {
		if i < len(recentMonths)/2 {
			previousTotal += p.MonthlySpending[m]
		} else {
			recentTotal += p.MonthlySpending[m]
		}
	}

	if previousTotal > 0 {
		p.GrowthRate = float64(recentTotal-previousTotal) / float64(previousTotal)

		if p.GrowthRate > 0.5 {
			p.SpendingTrend = TrendRapidGrowth
		} else if p.GrowthRate > 0.1 {
			p.SpendingTrend = TrendGrowing
		} else if p.GrowthRate < -0.1 {
			p.SpendingTrend = TrendDeclining
		} else {
			p.SpendingTrend = TrendStable
		}
	}
}

func (p *ConsumptionProfile) CalculateChurnRisk() {
	if p.LastPurchaseAt == nil {
		p.ChurnRisk = 1.0
		return
	}

	daysSinceLastPurchase := time.Since(*p.LastPurchaseAt).Hours() / 24

	switch {
	case daysSinceLastPurchase > 180:
		p.ChurnRisk = 0.95
	case daysSinceLastPurchase > 90:
		p.ChurnRisk = 0.75
	case daysSinceLastPurchase > 60:
		p.ChurnRisk = 0.5
	case daysSinceLastPurchase > 30:
		p.ChurnRisk = 0.25
	default:
		p.ChurnRisk = 0.1
	}

	if p.SpendingTrend == TrendDeclining {
		p.ChurnRisk = min(p.ChurnRisk+0.2, 1.0)
	}
}

func (p *ConsumptionProfile) DetermineValueSegment() {
	if p.RFM == nil {
		p.ValueSegment = SegmentNew
		return
	}

	switch {
	case p.RFM.TotalScore >= 9:
		p.ValueSegment = SegmentVIP
	case p.RFM.TotalScore >= 7:
		p.ValueSegment = SegmentChampion
	case p.RFM.TotalScore >= 5:
		p.ValueSegment = SegmentLoyal
	case p.RFM.TotalScore >= 3:
		p.ValueSegment = SegmentPotential
	case p.ChurnRisk > 0.7:
		p.ValueSegment = SegmentAtRisk
	default:
		p.ValueSegment = SegmentNew
	}
}

func (p *ConsumptionProfile) AddPurchasePattern(pattern *PurchasePattern) {
	p.PurchasePatterns = append(p.PurchasePatterns, pattern)
}

func NewRFMScore(profileID, userID uint64) *RFMScore {
	return &RFMScore{
		ProfileID: profileID,
		UserID:    userID,
	}
}

func (r *RFMScore) Calculate(recencyDays, frequencyCount int, monetaryValue int64) {
	r.RecencyDays = recencyDays
	r.FrequencyCount = frequencyCount
	r.MonetaryValue = monetaryValue

	switch {
	case recencyDays <= 7:
		r.RecencyScore = 5
	case recencyDays <= 30:
		r.RecencyScore = 4
	case recencyDays <= 90:
		r.RecencyScore = 3
	case recencyDays <= 180:
		r.RecencyScore = 2
	default:
		r.RecencyScore = 1
	}

	switch {
	case frequencyCount >= 20:
		r.FrequencyScore = 5
	case frequencyCount >= 10:
		r.FrequencyScore = 4
	case frequencyCount >= 5:
		r.FrequencyScore = 3
	case frequencyCount >= 2:
		r.FrequencyScore = 2
	default:
		r.FrequencyScore = 1
	}

	switch {
	case monetaryValue >= 100000:
		r.MonetaryScore = 5
	case monetaryValue >= 50000:
		r.MonetaryScore = 4
	case monetaryValue >= 20000:
		r.MonetaryScore = 3
	case monetaryValue >= 5000:
		r.MonetaryScore = 2
	default:
		r.MonetaryScore = 1
	}

	r.TotalScore = r.RecencyScore + r.FrequencyScore + r.MonetaryScore
	r.determineSegment()
}

func (r *RFMScore) determineSegment() {
	switch {
	case r.RecencyScore >= 4 && r.FrequencyScore >= 4 && r.MonetaryScore >= 4:
		r.Segment = "CHAMPION"
	case r.RecencyScore >= 3 && r.FrequencyScore >= 3 && r.MonetaryScore >= 3:
		r.Segment = "LOYAL"
	case r.RecencyScore >= 4 && r.FrequencyScore <= 2:
		r.Segment = "NEW"
	case r.RecencyScore <= 2 && r.FrequencyScore >= 3:
		r.Segment = "AT_RISK"
	case r.RecencyScore <= 2 && r.FrequencyScore <= 2:
		r.Segment = "CHURNED"
	default:
		r.Segment = "POTENTIAL"
	}
}

func NewPurchasePattern(profileID, userID uint64, patternType, patternName string) *PurchasePattern {
	return &PurchasePattern{
		ProfileID:   profileID,
		UserID:      userID,
		PatternType: patternType,
		PatternName: patternName,
		Attributes:  make(map[string]any),
	}
}

func (p *PurchasePattern) RecordOccurrence(amount int64) {
	p.Frequency++
	now := time.Now()
	if p.FirstObservedAt == nil {
		p.FirstObservedAt = &now
	}
	p.LastObservedAt = &now
}

type ConsumptionProfileRepository interface {
	Save(ctx context.Context, profile *ConsumptionProfile) error
	FindByID(ctx context.Context, id uint64) (*ConsumptionProfile, error)
	FindByUserID(ctx context.Context, userID uint64) (*ConsumptionProfile, error)
	FindByProfileID(ctx context.Context, profileID uint64) (*ConsumptionProfile, error)
	FindBySpendingLevel(ctx context.Context, level SpendingLevel, limit int) ([]*ConsumptionProfile, error)
	FindByValueSegment(ctx context.Context, segment ValueSegment, limit int) ([]*ConsumptionProfile, error)
	FindHighChurnRisk(ctx context.Context, threshold float64, limit int) ([]*ConsumptionProfile, error)
	Update(ctx context.Context, profile *ConsumptionProfile) error
	Delete(ctx context.Context, userID uint64) error
}

type RFMScoreRepository interface {
	Save(ctx context.Context, score *RFMScore) error
	FindByID(ctx context.Context, id uint64) (*RFMScore, error)
	FindByUserID(ctx context.Context, userID uint64) (*RFMScore, error)
	FindBySegment(ctx context.Context, segment string, limit int) ([]*RFMScore, error)
	Update(ctx context.Context, score *RFMScore) error
}

type PurchasePatternRepository interface {
	Save(ctx context.Context, pattern *PurchasePattern) error
	FindByID(ctx context.Context, id uint64) (*PurchasePattern, error)
	FindByProfileID(ctx context.Context, profileID uint64) ([]*PurchasePattern, error)
	FindByUserID(ctx context.Context, userID uint64) ([]*PurchasePattern, error)
	Delete(ctx context.Context, id uint64) error
}
