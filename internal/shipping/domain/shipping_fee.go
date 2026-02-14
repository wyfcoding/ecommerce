package domain

import (
	"errors"
	"slices"
	"time"
)

var (
	ErrShippingTemplateNotFound = errors.New("shipping template not found")
	ErrInvalidShippingRule      = errors.New("invalid shipping rule")
	ErrUnsupportedRegion        = errors.New("unsupported region")
	ErrWeightExceeded           = errors.New("weight exceeded")
	ErrDimensionExceeded        = errors.New("dimension exceeded")
)

type ShippingFeeType string

const (
	ShippingFeeTypeByWeight   ShippingFeeType = "BY_WEIGHT"
	ShippingFeeTypeByQuantity ShippingFeeType = "BY_QUANTITY"
	ShippingFeeTypeByVolume   ShippingFeeType = "BY_VOLUME"
	ShippingFeeTypeFixed      ShippingFeeType = "FIXED"
	ShippingFeeTypePiece      ShippingFeeType = "PIECE"
)

type FreeShippingType string

const (
	FreeShippingTypeNone       FreeShippingType = "NONE"
	FreeShippingTypeByAmount   FreeShippingType = "BY_AMOUNT"
	FreeShippingTypeByQuantity FreeShippingType = "BY_QUANTITY"
	FreeShippingTypeByWeight   FreeShippingType = "BY_WEIGHT"
	FreeShippingTypeUnlimited  FreeShippingType = "UNLIMITED"
)

type ShippingTemplate struct {
	ID                uint                `json:"id"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
	MerchantID        uint64              `json:"merchant_id"`
	Name              string              `json:"name"`
	Description       string              `json:"description"`
	FeeType           ShippingFeeType     `json:"fee_type"`
	IsDefault         bool                `json:"is_default"`
	Enabled           bool                `json:"enabled"`
	Rules             []*ShippingRule     `json:"rules"`
	FreeShippingRules []*FreeShippingRule `json:"free_shipping_rules"`
	PickupEnabled     bool                `json:"pickup_enabled"`
	PickupFee         int64               `json:"pickup_fee"`
	CODFee            int64               `json:"cod_fee"`
	InsuranceRate     float64             `json:"insurance_rate"`
	MinInsuranceFee   int64               `json:"min_insurance_fee"`
	MaxInsuranceFee   int64               `json:"max_insurance_fee"`
	ProcessingTime    int                 `json:"processing_time"`
	DeliveryDays      int                 `json:"delivery_days"`
}

type ShippingRule struct {
	ID             uint      `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	TemplateID     uint      `json:"template_id"`
	RegionType     string    `json:"region_type"`
	RegionCodes    []string  `json:"region_codes"`
	RegionNames    []string  `json:"region_names"`
	FirstUnit      int64     `json:"first_unit"`
	FirstFee       int64     `json:"first_fee"`
	AdditionalUnit int64     `json:"additional_unit"`
	AdditionalFee  int64     `json:"additional_fee"`
	MaxWeight      int64     `json:"max_weight"`
	MaxDimension   string    `json:"max_dimension"`
	Enabled        bool      `json:"enabled"`
	Priority       int       `json:"priority"`
}

type FreeShippingRule struct {
	ID          uint             `json:"id"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	TemplateID  uint             `json:"template_id"`
	Type        FreeShippingType `json:"type"`
	RegionCodes []string         `json:"region_codes"`
	RegionNames []string         `json:"region_names"`
	Threshold   int64            `json:"threshold"`
	MaxFee      int64            `json:"max_fee"`
	UserGroups  []string         `json:"user_groups"`
	StartTime   *time.Time       `json:"start_time"`
	EndTime     *time.Time       `json:"end_time"`
	Enabled     bool             `json:"enabled"`
	Priority    int              `json:"priority"`
}

type ShippingFeeCalculation struct {
	ID                  uint      `json:"id"`
	CreatedAt           time.Time `json:"created_at"`
	OrderID             uint64    `json:"order_id"`
	MerchantID          uint64    `json:"merchant_id"`
	TemplateID          uint      `json:"template_id"`
	DestinationCode     string    `json:"destination_code"`
	Weight              int64     `json:"weight"`
	Volume              int64     `json:"volume"`
	Quantity            int32     `json:"quantity"`
	Subtotal            int64     `json:"subtotal"`
	BaseFee             int64     `json:"base_fee"`
	AdditionalFee       int64     `json:"additional_fee"`
	DiscountFee         int64     `json:"discount_fee"`
	InsuranceFee        int64     `json:"insurance_fee"`
	CODFee              int64     `json:"cod_fee"`
	TotalFee            int64     `json:"total_fee"`
	FreeShippingApplied bool      `json:"free_shipping_applied"`
	FreeShippingRuleID  uint      `json:"free_shipping_rule_id"`
}

type ShippingProvider struct {
	ID               uint      `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Code             string    `json:"code"`
	Name             string    `json:"name"`
	Logo             string    `json:"logo"`
	Description      string    `json:"description"`
	Enabled          bool      `json:"enabled"`
	Priority         int       `json:"priority"`
	APIEndpoint      string    `json:"api_endpoint"`
	APIKey           string    `json:"api_key"`
	SecretKey        string    `json:"secret_key"`
	SupportCOD       bool      `json:"support_cod"`
	SupportInsurance bool      `json:"support_insurance"`
	SupportPickup    bool      `json:"support_pickup"`
	TrackingURL      string    `json:"tracking_url"`
	Coverage         []string  `json:"coverage"`
	RateTable        string    `json:"rate_table"`
}

type ShippingQuote struct {
	ProviderID        uint   `json:"provider_id"`
	ProviderCode      string `json:"provider_code"`
	ProviderName      string `json:"provider_name"`
	ServiceType       string `json:"service_type"`
	ServiceName       string `json:"service_name"`
	EstimatedDays     int    `json:"estimated_days"`
	BaseFee           int64  `json:"base_fee"`
	TotalFee          int64  `json:"total_fee"`
	InsuranceFee      int64  `json:"insurance_fee"`
	CODFee            int64  `json:"cod_fee"`
	Discount          int64  `json:"discount"`
	IsRecommended     bool   `json:"is_recommended"`
	IsAvailable       bool   `json:"is_available"`
	UnavailableReason string `json:"unavailable_reason"`
}

type ShippingFeeRequest struct {
	MerchantID      uint64   `json:"merchant_id"`
	TemplateID      uint     `json:"template_id"`
	DestinationCode string   `json:"destination_code"`
	Weight          int64    `json:"weight"`
	Volume          int64    `json:"volume"`
	Length          int64    `json:"length"`
	Width           int64    `json:"width"`
	Height          int64    `json:"height"`
	Quantity        int32    `json:"quantity"`
	Subtotal        int64    `json:"subtotal"`
	UserGroup       string   `json:"user_group"`
	IsCOD           bool     `json:"is_cod"`
	NeedInsurance   bool     `json:"need_insurance"`
	ProductIDs      []uint64 `json:"product_ids"`
}

func NewShippingTemplate(merchantID uint64, name string, feeType ShippingFeeType) *ShippingTemplate {
	return &ShippingTemplate{
		MerchantID:        merchantID,
		Name:              name,
		FeeType:           feeType,
		IsDefault:         false,
		Enabled:           true,
		Rules:             make([]*ShippingRule, 0),
		FreeShippingRules: make([]*FreeShippingRule, 0),
		PickupEnabled:     false,
		InsuranceRate:     0.005,
		ProcessingTime:    24,
		DeliveryDays:      3,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

func (t *ShippingTemplate) AddRule(rule *ShippingRule) {
	rule.TemplateID = t.ID
	t.Rules = append(t.Rules, rule)
	t.UpdatedAt = time.Now()
}

func (t *ShippingTemplate) RemoveRule(ruleID uint) {
	for i, r := range t.Rules {
		if r.ID == ruleID {
			t.Rules = append(t.Rules[:i], t.Rules[i+1:]...)
			break
		}
	}
	t.UpdatedAt = time.Now()
}

func (t *ShippingTemplate) AddFreeShippingRule(rule *FreeShippingRule) {
	rule.TemplateID = t.ID
	t.FreeShippingRules = append(t.FreeShippingRules, rule)
	t.UpdatedAt = time.Now()
}

func (t *ShippingTemplate) RemoveFreeShippingRule(ruleID uint) {
	for i, r := range t.FreeShippingRules {
		if r.ID == ruleID {
			t.FreeShippingRules = append(t.FreeShippingRules[:i], t.FreeShippingRules[i+1:]...)
			break
		}
	}
	t.UpdatedAt = time.Now()
}

func (t *ShippingTemplate) SetDefault(isDefault bool) {
	t.IsDefault = isDefault
	t.UpdatedAt = time.Now()
}

func (t *ShippingTemplate) Enable() {
	t.Enabled = true
	t.UpdatedAt = time.Now()
}

func (t *ShippingTemplate) Disable() {
	t.Enabled = false
	t.UpdatedAt = time.Now()
}

func (t *ShippingTemplate) CalculateFee(req *ShippingFeeRequest) (*ShippingFeeCalculation, error) {
	calc := &ShippingFeeCalculation{
		MerchantID:      req.MerchantID,
		TemplateID:      t.ID,
		DestinationCode: req.DestinationCode,
		Weight:          req.Weight,
		Volume:          req.Volume,
		Quantity:        req.Quantity,
		Subtotal:        req.Subtotal,
		CreatedAt:       time.Now(),
	}

	rule := t.findApplicableRule(req.DestinationCode)
	if rule == nil {
		return nil, ErrUnsupportedRegion
	}

	switch t.FeeType {
	case ShippingFeeTypeByWeight:
		calc.BaseFee = rule.FirstFee
		if req.Weight > rule.FirstUnit {
			additionalWeight := req.Weight - rule.FirstUnit
			additionalUnits := additionalWeight / rule.AdditionalUnit
			if additionalWeight%rule.AdditionalUnit > 0 {
				additionalUnits++
			}
			calc.AdditionalFee = additionalUnits * rule.AdditionalFee
		}
	case ShippingFeeTypeByQuantity:
		calc.BaseFee = rule.FirstFee
		if req.Quantity > int32(rule.FirstUnit) {
			additionalQuantity := int64(req.Quantity) - rule.FirstUnit
			additionalUnits := additionalQuantity / rule.AdditionalUnit
			if additionalQuantity%rule.AdditionalUnit > 0 {
				additionalUnits++
			}
			calc.AdditionalFee = additionalUnits * rule.AdditionalFee
		}
	case ShippingFeeTypeFixed:
		calc.BaseFee = rule.FirstFee
	case ShippingFeeTypeByVolume:
		calc.BaseFee = rule.FirstFee
		if req.Volume > rule.FirstUnit {
			additionalVolume := req.Volume - rule.FirstUnit
			additionalUnits := additionalVolume / rule.AdditionalUnit
			if additionalVolume%rule.AdditionalUnit > 0 {
				additionalUnits++
			}
			calc.AdditionalFee = additionalUnits * rule.AdditionalFee
		}
	}

	freeShippingRule := t.findApplicableFreeShippingRule(req)
	if freeShippingRule != nil {
		calc.FreeShippingApplied = true
		calc.FreeShippingRuleID = freeShippingRule.ID
		if freeShippingRule.MaxFee > 0 {
			calc.DiscountFee = min(calc.BaseFee+calc.AdditionalFee, freeShippingRule.MaxFee)
		} else {
			calc.DiscountFee = calc.BaseFee + calc.AdditionalFee
		}
	}

	if req.NeedInsurance && t.InsuranceRate > 0 {
		insuranceFee := int64(float64(req.Subtotal) * t.InsuranceRate)
		if t.MinInsuranceFee > 0 && insuranceFee < t.MinInsuranceFee {
			insuranceFee = t.MinInsuranceFee
		}
		if t.MaxInsuranceFee > 0 && insuranceFee > t.MaxInsuranceFee {
			insuranceFee = t.MaxInsuranceFee
		}
		calc.InsuranceFee = insuranceFee
	}

	if req.IsCOD {
		calc.CODFee = t.CODFee
	}

	calc.TotalFee = max(calc.BaseFee+calc.AdditionalFee-calc.DiscountFee+calc.InsuranceFee+calc.CODFee, 0)

	return calc, nil
}

func (t *ShippingTemplate) findApplicableRule(destinationCode string) *ShippingRule {
	for _, rule := range t.Rules {
		if !rule.Enabled {
			continue
		}
		for _, code := range rule.RegionCodes {
			if code == destinationCode || isRegionMatch(code, destinationCode) {
				return rule
			}
		}
	}
	return nil
}

func (t *ShippingTemplate) findApplicableFreeShippingRule(req *ShippingFeeRequest) *FreeShippingRule {
	now := time.Now()
	for _, rule := range t.FreeShippingRules {
		if !rule.Enabled {
			continue
		}
		if rule.StartTime != nil && now.Before(*rule.StartTime) {
			continue
		}
		if rule.EndTime != nil && now.After(*rule.EndTime) {
			continue
		}
		if len(rule.RegionCodes) > 0 {
			matched := false
			for _, code := range rule.RegionCodes {
				if code == req.DestinationCode || isRegionMatch(code, req.DestinationCode) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		if len(rule.UserGroups) > 0 && req.UserGroup != "" {
			matched := slices.Contains(rule.UserGroups, req.UserGroup)
			if !matched {
				continue
			}
		}
		switch rule.Type {
		case FreeShippingTypeByAmount:
			if req.Subtotal >= rule.Threshold {
				return rule
			}
		case FreeShippingTypeByQuantity:
			if int64(req.Quantity) >= rule.Threshold {
				return rule
			}
		case FreeShippingTypeByWeight:
			if req.Weight >= rule.Threshold {
				return rule
			}
		case FreeShippingTypeUnlimited:
			return rule
		}
	}
	return nil
}

func NewShippingRule(templateID uint, regionType string, regionCodes, regionNames []string) *ShippingRule {
	return &ShippingRule{
		TemplateID:     templateID,
		RegionType:     regionType,
		RegionCodes:    regionCodes,
		RegionNames:    regionNames,
		FirstUnit:      1000,
		FirstFee:       1000,
		AdditionalUnit: 500,
		AdditionalFee:  500,
		Enabled:        true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func (r *ShippingRule) SetFee(firstUnit, firstFee, additionalUnit, additionalFee int64) {
	r.FirstUnit = firstUnit
	r.FirstFee = firstFee
	r.AdditionalUnit = additionalUnit
	r.AdditionalFee = additionalFee
	r.UpdatedAt = time.Now()
}

func (r *ShippingRule) SetMaxWeight(maxWeight int64) {
	r.MaxWeight = maxWeight
	r.UpdatedAt = time.Now()
}

func (r *ShippingRule) Enable() {
	r.Enabled = true
	r.UpdatedAt = time.Now()
}

func (r *ShippingRule) Disable() {
	r.Enabled = false
	r.UpdatedAt = time.Now()
}

func NewFreeShippingRule(templateID uint, freeShippingType FreeShippingType, threshold int64) *FreeShippingRule {
	return &FreeShippingRule{
		TemplateID:  templateID,
		Type:        freeShippingType,
		Threshold:   threshold,
		Enabled:     true,
		RegionCodes: make([]string, 0),
		RegionNames: make([]string, 0),
		UserGroups:  make([]string, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (r *FreeShippingRule) SetRegions(codes, names []string) {
	r.RegionCodes = codes
	r.RegionNames = names
	r.UpdatedAt = time.Now()
}

func (r *FreeShippingRule) SetUserGroups(groups []string) {
	r.UserGroups = groups
	r.UpdatedAt = time.Now()
}

func (r *FreeShippingRule) SetTimeRange(start, end time.Time) {
	r.StartTime = &start
	r.EndTime = &end
	r.UpdatedAt = time.Now()
}

func (r *FreeShippingRule) Enable() {
	r.Enabled = true
	r.UpdatedAt = time.Now()
}

func (r *FreeShippingRule) Disable() {
	r.Enabled = false
	r.UpdatedAt = time.Now()
}

func NewShippingProvider(code, name string) *ShippingProvider {
	return &ShippingProvider{
		Code:             code,
		Name:             name,
		Enabled:          true,
		Coverage:         make([]string, 0),
		SupportCOD:       false,
		SupportInsurance: false,
		SupportPickup:    false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func (p *ShippingProvider) SetAPI(endpoint, apiKey, secretKey string) {
	p.APIEndpoint = endpoint
	p.APIKey = apiKey
	p.SecretKey = secretKey
	p.UpdatedAt = time.Now()
}

func (p *ShippingProvider) SetCoverage(regions []string) {
	p.Coverage = regions
	p.UpdatedAt = time.Now()
}

func (p *ShippingProvider) Enable() {
	p.Enabled = true
	p.UpdatedAt = time.Now()
}

func (p *ShippingProvider) Disable() {
	p.Enabled = false
	p.UpdatedAt = time.Now()
}

func isRegionMatch(pattern, code string) bool {
	if len(pattern) > len(code) {
		return false
	}
	return code[:len(pattern)] == pattern
}

type ShippingTemplateRepository interface {
	FindByID(ctx any, id uint) (*ShippingTemplate, error)
	FindByMerchantID(ctx any, merchantID uint64) ([]*ShippingTemplate, error)
	FindDefaultByMerchantID(ctx any, merchantID uint64) (*ShippingTemplate, error)
	Save(ctx any, template *ShippingTemplate) error
	Update(ctx any, template *ShippingTemplate) error
	Delete(ctx any, id uint) error

	SaveRule(ctx any, rule *ShippingRule) error
	DeleteRule(ctx any, ruleID uint) error

	SaveFreeShippingRule(ctx any, rule *FreeShippingRule) error
	DeleteFreeShippingRule(ctx any, ruleID uint) error

	SaveProvider(ctx any, provider *ShippingProvider) error
	FindProviderByID(ctx any, id uint) (*ShippingProvider, error)
	FindProviderByCode(ctx any, code string) (*ShippingProvider, error)
	FindEnabledProviders(ctx any) ([]*ShippingProvider, error)
}

type ShippingFeeService interface {
	CalculateFee(ctx any, req *ShippingFeeRequest) (*ShippingFeeCalculation, error)
	GetQuotes(ctx any, req *ShippingFeeRequest) ([]*ShippingQuote, error)
	GetTemplate(ctx any, templateID uint) (*ShippingTemplate, error)
	CreateTemplate(ctx any, merchantID uint64, name string, feeType ShippingFeeType) (*ShippingTemplate, error)
	UpdateTemplate(ctx any, templateID uint, updates map[string]any) error
	DeleteTemplate(ctx any, templateID uint) error
	AddShippingRule(ctx any, templateID uint, rule *ShippingRule) error
	AddFreeShippingRule(ctx any, templateID uint, rule *FreeShippingRule) error
}
