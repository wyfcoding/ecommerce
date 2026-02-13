package domain

import (
	"context"
	"errors"
	"math"
	"time"
)

var (
	ErrTemplateNotFound     = errors.New("shipping template not found")
	ErrInvalidTemplateType  = errors.New("invalid template type")
	ErrInvalidRegionConfig  = errors.New("invalid region configuration")
)

type TemplateType int8

const (
	TemplateTypeWeight   TemplateType = 1
	TemplateTypeVolume   TemplateType = 2
	TemplateTypeQuantity TemplateType = 3
	TemplateTypeFixed    TemplateType = 4
)

type ShippingTemplate struct {
	ID           uint64                `json:"id"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	Name         string                `json:"name"`
	MerchantID   uint64                `json:"merchant_id"`
	Type         TemplateType          `json:"type"`
	IsDefault    bool                  `json:"is_default"`
	Enabled      bool                  `json:"enabled"`
	Description  string                `json:"description"`
	DefaultRule  *ShippingRule         `json:"default_rule"`
	RegionRules  []*RegionShippingRule `json:"region_rules"`
	FreeShipping *FreeShippingRule     `json:"free_shipping"`
}

type ShippingRule struct {
	ID             uint64  `json:"id"`
	FirstUnit      float64 `json:"first_unit"`
	FirstFee       int64   `json:"first_fee"`
	AdditionalUnit float64 `json:"additional_unit"`
	AdditionalFee  int64   `json:"additional_fee"`
}

type RegionShippingRule struct {
	ID             uint64   `json:"id"`
	TemplateID     uint64   `json:"template_id"`
	Provinces      []string `json:"provinces"`
	Cities         []string `json:"cities"`
	Districts      []string `json:"districts"`
	FirstUnit      float64  `json:"first_unit"`
	FirstFee       int64    `json:"first_fee"`
	AdditionalUnit float64  `json:"additional_unit"`
	AdditionalFee  int64    `json:"additional_fee"`
}

func (r *RegionShippingRule) ToShippingRule() *ShippingRule {
	return &ShippingRule{
		ID:             r.ID,
		FirstUnit:      r.FirstUnit,
		FirstFee:       r.FirstFee,
		AdditionalUnit: r.AdditionalUnit,
		AdditionalFee:  r.AdditionalFee,
	}
}

type FreeShippingRule struct {
	ID           uint64   `json:"id"`
	TemplateID   uint64   `json:"template_id"`
	Enabled      bool     `json:"enabled"`
	MinAmount    int64    `json:"min_amount"`
	MinQuantity  int32    `json:"min_quantity"`
	Provinces    []string `json:"provinces"`
	Cities       []string `json:"cities"`
	ExcludeItems []uint64 `json:"exclude_items"`
}

type ShippingFeeRequest struct {
	TemplateID  uint64            `json:"template_id"`
	Weight      float64           `json:"weight"`
	Volume      float64           `json:"volume"`
	Quantity    int32             `json:"quantity"`
	Amount      int64             `json:"amount"`
	Province    string            `json:"province"`
	City        string            `json:"city"`
	District    string            `json:"district"`
	Items       []*ShippingItem   `json:"items"`
}

type ShippingItem struct {
	ProductID uint64  `json:"product_id"`
	SkuID     uint64  `json:"sku_id"`
	Weight    float64 `json:"weight"`
	Volume    float64 `json:"volume"`
	Quantity  int32   `json:"quantity"`
	Price     int64   `json:"price"`
}

type ShippingFeeResult struct {
	TotalFee       int64            `json:"total_fee"`
	BaseFee        int64            `json:"base_fee"`
	WeightFee      int64            `json:"weight_fee"`
	VolumeFee      int64            `json:"volume_fee"`
	QuantityFee    int64            `json:"quantity_fee"`
	IsFreeShipping bool             `json:"is_free_shipping"`
	AppliedRule    *AppliedRuleInfo `json:"applied_rule"`
}

type AppliedRuleInfo struct {
	RuleType   string  `json:"rule_type"`
	FirstUnit  float64 `json:"first_unit"`
	FirstFee   int64   `json:"first_fee"`
	AddUnit    float64 `json:"add_unit"`
	AddFee     int64   `json:"add_fee"`
	TotalUnits float64 `json:"total_units"`
}

func NewShippingTemplate(name string, merchantID uint64, templateType TemplateType) *ShippingTemplate {
	return &ShippingTemplate{
		Name:        name,
		MerchantID:  merchantID,
		Type:        templateType,
		IsDefault:   false,
		Enabled:     true,
		RegionRules: []*RegionShippingRule{},
	}
}

func (t *ShippingTemplate) SetDefaultRule(firstUnit, additionalUnit float64, firstFee, additionalFee int64) {
	t.DefaultRule = &ShippingRule{
		FirstUnit:      firstUnit,
		FirstFee:       firstFee,
		AdditionalUnit: additionalUnit,
		AdditionalFee:  additionalFee,
	}
}

func (t *ShippingTemplate) AddRegionRule(provinces, cities, districts []string, firstUnit, additionalUnit float64, firstFee, additionalFee int64) {
	rule := &RegionShippingRule{
		TemplateID:     t.ID,
		Provinces:      provinces,
		Cities:         cities,
		Districts:      districts,
		FirstUnit:      firstUnit,
		FirstFee:       firstFee,
		AdditionalUnit: additionalUnit,
		AdditionalFee:  additionalFee,
	}
	t.RegionRules = append(t.RegionRules, rule)
}

func (t *ShippingTemplate) SetFreeShipping(enabled bool, minAmount int64, minQuantity int32, provinces, cities []string) {
	t.FreeShipping = &FreeShippingRule{
		TemplateID:  t.ID,
		Enabled:     enabled,
		MinAmount:   minAmount,
		MinQuantity: minQuantity,
		Provinces:   provinces,
		Cities:      cities,
	}
}

func (t *ShippingTemplate) CalculateFee(req *ShippingFeeRequest) (*ShippingFeeResult, error) {
	result := &ShippingFeeResult{}

	if t.FreeShipping != nil && t.FreeShipping.Enabled {
		if t.isFreeShippingApplicable(req) {
			result.IsFreeShipping = true
			result.TotalFee = 0
			return result, nil
		}
	}

	var rule *ShippingRule
	regionRule := t.findApplicableRule(req.Province, req.City, req.District)
	if regionRule != nil {
		rule = regionRule.ToShippingRule()
	} else {
		rule = t.DefaultRule
	}
	if rule == nil {
		return nil, ErrInvalidRegionConfig
	}

	var fee int64
	var totalUnits float64

	switch t.Type {
	case TemplateTypeWeight:
		fee, totalUnits = t.calculateByWeight(rule, req.Weight)
		result.WeightFee = fee
	case TemplateTypeVolume:
		fee, totalUnits = t.calculateByVolume(rule, req.Volume)
		result.VolumeFee = fee
	case TemplateTypeQuantity:
		fee, totalUnits = t.calculateByQuantity(rule, req.Quantity)
		result.QuantityFee = fee
	case TemplateTypeFixed:
		fee = rule.FirstFee
		totalUnits = 1
	default:
		return nil, ErrInvalidTemplateType
	}

	result.BaseFee = fee
	result.TotalFee = fee
	result.AppliedRule = &AppliedRuleInfo{
		RuleType:   t.Type.String(),
		FirstUnit:  rule.FirstUnit,
		FirstFee:   rule.FirstFee,
		AddUnit:    rule.AdditionalUnit,
		AddFee:     rule.AdditionalFee,
		TotalUnits: totalUnits,
	}

	return result, nil
}

func (t *ShippingTemplate) isFreeShippingApplicable(req *ShippingFeeRequest) bool {
	if !t.FreeShipping.Enabled {
		return false
	}

	if len(t.FreeShipping.Provinces) > 0 {
		found := false
		for _, p := range t.FreeShipping.Provinces {
			if p == req.Province || p == "*" {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(t.FreeShipping.Cities) > 0 {
		found := false
		for _, c := range t.FreeShipping.Cities {
			if c == req.City || c == "*" {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if t.FreeShipping.MinAmount > 0 && req.Amount >= t.FreeShipping.MinAmount {
		return true
	}
	if t.FreeShipping.MinQuantity > 0 && req.Quantity >= t.FreeShipping.MinQuantity {
		return true
	}

	return false
}

func (t *ShippingTemplate) findApplicableRule(province, city, district string) *RegionShippingRule {
	for _, rule := range t.RegionRules {
		if t.matchRegion(rule, province, city, district) {
			return rule
		}
	}
	return nil
}

func (t *ShippingTemplate) matchRegion(rule *RegionShippingRule, province, city, district string) bool {
	for _, p := range rule.Provinces {
		if p == province || p == "*" {
			if len(rule.Cities) == 0 {
				return true
			}
			for _, c := range rule.Cities {
				if c == city || c == "*" {
					if len(rule.Districts) == 0 {
						return true
					}
					for _, d := range rule.Districts {
						if d == district || d == "*" {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func (t *ShippingTemplate) calculateByWeight(rule *ShippingRule, weight float64) (int64, float64) {
	if weight <= rule.FirstUnit {
		return rule.FirstFee, 1
	}
	additionalWeight := weight - rule.FirstUnit
	additionalUnits := math.Ceil(additionalWeight / rule.AdditionalUnit)
	totalFee := rule.FirstFee + int64(additionalUnits)*rule.AdditionalFee
	return totalFee, 1 + additionalUnits
}

func (t *ShippingTemplate) calculateByVolume(rule *ShippingRule, volume float64) (int64, float64) {
	if volume <= rule.FirstUnit {
		return rule.FirstFee, 1
	}
	additionalVolume := volume - rule.FirstUnit
	additionalUnits := math.Ceil(additionalVolume / rule.AdditionalUnit)
	totalFee := rule.FirstFee + int64(additionalUnits)*rule.AdditionalFee
	return totalFee, 1 + additionalUnits
}

func (t *ShippingTemplate) calculateByQuantity(rule *ShippingRule, quantity int32) (int64, float64) {
	if quantity <= int32(rule.FirstUnit) {
		return rule.FirstFee, 1
	}
	additionalQuantity := float64(quantity) - rule.FirstUnit
	additionalUnits := math.Ceil(additionalQuantity / rule.AdditionalUnit)
	totalFee := rule.FirstFee + int64(additionalUnits)*rule.AdditionalFee
	return totalFee, 1 + additionalUnits
}

func (t TemplateType) String() string {
	switch t {
	case TemplateTypeWeight:
		return "WEIGHT"
	case TemplateTypeVolume:
		return "VOLUME"
	case TemplateTypeQuantity:
		return "QUANTITY"
	case TemplateTypeFixed:
		return "FIXED"
	default:
		return "UNKNOWN"
	}
}

type ShippingTemplateRepository interface {
	Save(ctx context.Context, template *ShippingTemplate) error
	FindByID(ctx context.Context, id uint64) (*ShippingTemplate, error)
	FindByMerchantID(ctx context.Context, merchantID uint64, limit, offset int) ([]*ShippingTemplate, error)
	FindDefaultByMerchantID(ctx context.Context, merchantID uint64) (*ShippingTemplate, error)
	Update(ctx context.Context, template *ShippingTemplate) error
	Delete(ctx context.Context, id uint64) error
}

type ShippingFeeCalculator interface {
	Calculate(ctx context.Context, req *ShippingFeeRequest) (*ShippingFeeResult, error)
	CalculateBatch(ctx context.Context, requests []*ShippingFeeRequest) ([]*ShippingFeeResult, error)
}
