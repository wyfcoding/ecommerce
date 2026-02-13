package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrSpecTemplateNotFound    = errors.New("spec template not found")
	ErrSpecTemplateInUse       = errors.New("spec template is in use")
	ErrSpecOptionNotFound      = errors.New("spec option not found")
	ErrDuplicateSpecName       = errors.New("duplicate spec name")
	ErrInvalidSpecValue        = errors.New("invalid spec value")
)

type SpecTemplateStatus int8

const (
	SpecTemplateStatusEnabled  SpecTemplateStatus = 1
	SpecTemplateStatusDisabled SpecTemplateStatus = 0
)

func (s SpecTemplateStatus) String() string {
	switch s {
	case SpecTemplateStatusEnabled:
		return "ENABLED"
	case SpecTemplateStatusDisabled:
		return "DISABLED"
	default:
		return "UNKNOWN"
	}
}

type SpecTemplate struct {
	ID          uint               `json:"id"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	TemplateNo  string             `json:"template_no"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	CategoryID  uint64             `json:"category_id"`
	Status      SpecTemplateStatus `json:"status"`
	Sort        int                `json:"sort"`
	Specs       []*SpecDefinition  `json:"specs"`
	UseCount    int                `json:"use_count"`
	CreatorID   uint64             `json:"creator_id"`
	CreatorName string             `json:"creator_name"`
}

type SpecDefinition struct {
	ID           uint               `json:"id"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	TemplateID   uint               `json:"template_id"`
	Name         string             `json:"name"`
	DisplayName  string             `json:"display_name"`
	SpecType     SpecType           `json:"spec_type"`
	Required     bool               `json:"required"`
	Sort         int                `json:"sort"`
	Searchable   bool               `json:"searchable"`
	Filterable   bool               `json:"filterable"`
	Options      []*SpecOption      `json:"options"`
	DefaultValue string             `json:"default_value"`
	Unit         string             `json:"unit"`
	MinValue     float64            `json:"min_value"`
	MaxValue     float64            `json:"max_value"`
}

type SpecType int8

const (
	SpecTypeText    SpecType = 1
	SpecTypeNumber  SpecType = 2
	SpecTypeSelect  SpecType = 3
	SpecTypeMultiSelect SpecType = 4
	SpecTypeColor   SpecType = 5
	SpecTypeImage   SpecType = 6
	SpecTypeDate    SpecType = 7
)

func (t SpecType) String() string {
	switch t {
	case SpecTypeText:
		return "TEXT"
	case SpecTypeNumber:
		return "NUMBER"
	case SpecTypeSelect:
		return "SELECT"
	case SpecTypeMultiSelect:
		return "MULTI_SELECT"
	case SpecTypeColor:
		return "COLOR"
	case SpecTypeImage:
		return "IMAGE"
	case SpecTypeDate:
		return "DATE"
	default:
		return "UNKNOWN"
	}
}

type SpecOption struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	SpecID      uint      `json:"spec_id"`
	Value       string    `json:"value"`
	DisplayName string    `json:"display_name"`
	ImageURL    string    `json:"image_url"`
	ColorCode   string    `json:"color_code"`
	Sort        int       `json:"sort"`
	Enabled     bool      `json:"enabled"`
}

type SpecValue struct {
	SpecID    uint64 `json:"spec_id"`
	SpecName  string `json:"spec_name"`
	Value     string `json:"value"`
	ValueID   uint64 `json:"value_id"`
	Unit      string `json:"unit"`
	ImageURL  string `json:"image_url"`
}

type SKUCombination struct {
	ID           uint               `json:"id"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	ProductID    uint64             `json:"product_id"`
	SkuCode      string             `json:"sku_code"`
	SpecValues   []*SpecValue       `json:"spec_values"`
	Price        int64              `json:"price"`
	Stock        int32              `json:"stock"`
	ImageURL     string             `json:"image_url"`
	Enabled      bool               `json:"enabled"`
}

func NewSpecTemplate(templateNo, name, description string, categoryID uint64, creatorID uint64, creatorName string) *SpecTemplate {
	return &SpecTemplate{
		TemplateNo:  templateNo,
		Name:        name,
		Description: description,
		CategoryID:  categoryID,
		Status:      SpecTemplateStatusEnabled,
		Sort:        0,
		Specs:       make([]*SpecDefinition, 0),
		UseCount:    0,
		CreatorID:   creatorID,
		CreatorName: creatorName,
	}
}

func (t *SpecTemplate) AddSpec(name, displayName string, specType SpecType, required bool, sort int) (*SpecDefinition, error) {
	for _, spec := range t.Specs {
		if spec.Name == name {
			return nil, ErrDuplicateSpecName
		}
	}

	spec := &SpecDefinition{
		TemplateID:  t.ID,
		Name:        name,
		DisplayName: displayName,
		SpecType:    specType,
		Required:    required,
		Sort:        sort,
		Searchable:  true,
		Filterable:  true,
		Options:     make([]*SpecOption, 0),
	}

	t.Specs = append(t.Specs, spec)
	return spec, nil
}

func (t *SpecTemplate) RemoveSpec(specID uint) error {
	for i, spec := range t.Specs {
		if spec.ID == specID {
			t.Specs = append(t.Specs[:i], t.Specs[i+1:]...)
			return nil
		}
	}
	return ErrSpecOptionNotFound
}

func (t *SpecTemplate) GetSpec(specID uint) *SpecDefinition {
	for _, spec := range t.Specs {
		if spec.ID == specID {
			return spec
		}
	}
	return nil
}

func (t *SpecTemplate) Enable() {
	t.Status = SpecTemplateStatusEnabled
}

func (t *SpecTemplate) Disable() {
	t.Status = SpecTemplateStatusDisabled
}

func (t *SpecTemplate) IsEnabled() bool {
	return t.Status == SpecTemplateStatusEnabled
}

func (t *SpecTemplate) IncrementUseCount() {
	t.UseCount++
}

func (t *SpecTemplate) DecrementUseCount() {
	if t.UseCount > 0 {
		t.UseCount--
	}
}

func (s *SpecDefinition) AddOption(value, displayName, imageURL, colorCode string, sort int) (*SpecOption, error) {
	for _, opt := range s.Options {
		if opt.Value == value {
			return nil, ErrDuplicateSpecName
		}
	}

	option := &SpecOption{
		SpecID:      s.ID,
		Value:       value,
		DisplayName: displayName,
		ImageURL:    imageURL,
		ColorCode:   colorCode,
		Sort:        sort,
		Enabled:     true,
	}

	s.Options = append(s.Options, option)
	return option, nil
}

func (s *SpecDefinition) RemoveOption(optionID uint) error {
	for i, opt := range s.Options {
		if opt.ID == optionID {
			s.Options = append(s.Options[:i], s.Options[i+1:]...)
			return nil
		}
	}
	return ErrSpecOptionNotFound
}

func (s *SpecDefinition) GetOption(optionID uint) *SpecOption {
	for _, opt := range s.Options {
		if opt.ID == optionID {
			return opt
		}
	}
	return nil
}

func (s *SpecDefinition) ValidateValue(value string) error {
	switch s.SpecType {
	case SpecTypeSelect, SpecTypeMultiSelect:
		found := false
		for _, opt := range s.Options {
			if opt.Value == value && opt.Enabled {
				found = true
				break
			}
		}
		if !found {
			return ErrInvalidSpecValue
		}
	case SpecTypeNumber:
	case SpecTypeText:
	case SpecTypeColor:
	case SpecTypeImage:
	case SpecTypeDate:
	}
	return nil
}

func (o *SpecOption) Enable() {
	o.Enabled = true
}

func (o *SpecOption) Disable() {
	o.Enabled = false
}

func NewSKUCombination(productID uint64, skuCode string, specValues []*SpecValue, price int64, stock int32) *SKUCombination {
	return &SKUCombination{
		ProductID:  productID,
		SkuCode:    skuCode,
		SpecValues: specValues,
		Price:      price,
		Stock:      stock,
		Enabled:    true,
	}
}

func (c *SKUCombination) GetSpecValue(specName string) *SpecValue {
	for _, sv := range c.SpecValues {
		if sv.SpecName == specName {
			return sv
		}
	}
	return nil
}

func (c *SKUCombination) GenerateSkuCode(prefix string) string {
	code := prefix
	for _, sv := range c.SpecValues {
		code += fmt.Sprintf("-%s", sv.Value)
	}
	return code
}

func (c *SKUCombination) UpdatePrice(price int64) {
	c.Price = price
}

func (c *SKUCombination) UpdateStock(stock int32) {
	c.Stock = stock
}

func (c *SKUCombination) Enable() {
	c.Enabled = true
}

func (c *SKUCombination) Disable() {
	c.Enabled = false
}

type SpecTemplateRepository interface {
	Save(ctx context.Context, template *SpecTemplate) error
	FindByID(ctx context.Context, id uint64) (*SpecTemplate, error)
	FindByTemplateNo(ctx context.Context, templateNo string) (*SpecTemplate, error)
	FindByCategoryID(ctx context.Context, categoryID uint64) ([]*SpecTemplate, error)
	FindEnabled(ctx context.Context, limit, offset int) ([]*SpecTemplate, error)
	FindAll(ctx context.Context, limit, offset int) ([]*SpecTemplate, error)
	Update(ctx context.Context, template *SpecTemplate) error
	Delete(ctx context.Context, id uint64) error
}

type SpecDefinitionRepository interface {
	Save(ctx context.Context, spec *SpecDefinition) error
	FindByID(ctx context.Context, id uint64) (*SpecDefinition, error)
	FindByTemplateID(ctx context.Context, templateID uint) ([]*SpecDefinition, error)
	Update(ctx context.Context, spec *SpecDefinition) error
	Delete(ctx context.Context, id uint64) error
}

type SpecOptionRepository interface {
	Save(ctx context.Context, option *SpecOption) error
	FindByID(ctx context.Context, id uint64) (*SpecOption, error)
	FindBySpecID(ctx context.Context, specID uint) ([]*SpecOption, error)
	Update(ctx context.Context, option *SpecOption) error
	Delete(ctx context.Context, id uint64) error
}

type SKUCombinationRepository interface {
	Save(ctx context.Context, combination *SKUCombination) error
	FindByID(ctx context.Context, id uint64) (*SKUCombination, error)
	FindByProductID(ctx context.Context, productID uint64) ([]*SKUCombination, error)
	FindBySkuCode(ctx context.Context, skuCode string) (*SKUCombination, error)
	Update(ctx context.Context, combination *SKUCombination) error
	Delete(ctx context.Context, id uint64) error
	DeleteByProductID(ctx context.Context, productID uint64) error
}
