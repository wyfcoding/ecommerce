package domain

import (
	"context"
	"errors"
	"slices"
	"time"
)

var (
	ErrInvalidTagCategory = errors.New("invalid tag category")
	ErrTagExpired         = errors.New("tag has expired")
)

type TagCategory int8

const (
	TagCategoryBasic       TagCategory = 1
	TagCategoryBehavior    TagCategory = 2
	TagCategoryConsumption TagCategory = 3
	TagCategoryInterest    TagCategory = 4
	TagCategorySocial      TagCategory = 5
	TagCategoryRisk        TagCategory = 6
	TagCategoryCustom      TagCategory = 99
)

func (c TagCategory) String() string {
	switch c {
	case TagCategoryBasic:
		return "BASIC"
	case TagCategoryBehavior:
		return "BEHAVIOR"
	case TagCategoryConsumption:
		return "CONSUMPTION"
	case TagCategoryInterest:
		return "INTEREST"
	case TagCategorySocial:
		return "SOCIAL"
	case TagCategoryRisk:
		return "RISK"
	case TagCategoryCustom:
		return "CUSTOM"
	default:
		return "UNKNOWN"
	}
}

type TagSource int8

const (
	TagSourceSystem     TagSource = 1
	TagSourceManual     TagSource = 2
	TagSourceAlgorithm  TagSource = 3
	TagSourceThirdParty TagSource = 4
	TagSourceUserInput  TagSource = 5
)

func (s TagSource) String() string {
	switch s {
	case TagSourceSystem:
		return "SYSTEM"
	case TagSourceManual:
		return "MANUAL"
	case TagSourceAlgorithm:
		return "ALGORITHM"
	case TagSourceThirdParty:
		return "THIRD_PARTY"
	case TagSourceUserInput:
		return "USER_INPUT"
	default:
		return "UNKNOWN"
	}
}

type TagStatus int8

const (
	TagStatusActive  TagStatus = 1
	TagStatusExpired TagStatus = 2
	TagStatusInvalid TagStatus = 3
)

func (s TagStatus) String() string {
	switch s {
	case TagStatusActive:
		return "ACTIVE"
	case TagStatusExpired:
		return "EXPIRED"
	case TagStatusInvalid:
		return "INVALID"
	default:
		return "UNKNOWN"
	}
}

type UserTag struct {
	ID          uint64         `json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	ProfileID   uint64         `json:"profile_id"`
	UserID      uint64         `json:"user_id"`
	TagKey      string         `json:"tag_key"`
	TagValue    string         `json:"tag_value"`
	TagType     string         `json:"tag_type"`
	Category    TagCategory    `json:"category"`
	Source      TagSource      `json:"source"`
	Confidence  float64        `json:"confidence"`
	Weight      float64        `json:"weight"`
	Status      TagStatus      `json:"status"`
	ExpiresAt   *time.Time     `json:"expires_at"`
	ValidFrom   *time.Time     `json:"valid_from"`
	ValidTo     *time.Time     `json:"valid_to"`
	Metadata    map[string]any `json:"metadata"`
	ParentTagID uint64         `json:"parent_tag_id"`
	Level       int            `json:"level"`
	Path        string         `json:"path"`
}

type TagDefinition struct {
	ID            uint64           `json:"id"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	TagKey        string           `json:"tag_key"`
	TagName       string           `json:"tag_name"`
	Description   string           `json:"description"`
	Category      TagCategory      `json:"category"`
	TagType       string           `json:"tag_type"`
	DataType      string           `json:"data_type"`
	AllowedValues []string         `json:"allowed_values"`
	DefaultValue  string           `json:"default_value"`
	IsRequired    bool             `json:"is_required"`
	IsMultiple    bool             `json:"is_multiple"`
	IsSystem      bool             `json:"is_system"`
	SortOrder     int              `json:"sort_order"`
	ParentID      uint64           `json:"parent_id"`
	Children      []*TagDefinition `json:"children"`
	Enabled       bool             `json:"enabled"`
}

type TagGroup struct {
	ID          uint64           `json:"id"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	GroupKey    string           `json:"group_key"`
	GroupName   string           `json:"group_name"`
	Description string           `json:"description"`
	Category    TagCategory      `json:"category"`
	Tags        []*TagDefinition `json:"tags"`
	SortOrder   int              `json:"sort_order"`
	Enabled     bool             `json:"enabled"`
}

func NewUserTag(profileID, userID uint64, tagKey, tagValue string, category TagCategory, source TagSource) *UserTag {
	return &UserTag{
		ProfileID:  profileID,
		UserID:     userID,
		TagKey:     tagKey,
		TagValue:   tagValue,
		Category:   category,
		Source:     source,
		Confidence: 1.0,
		Weight:     1.0,
		Status:     TagStatusActive,
		Metadata:   make(map[string]any),
	}
}

func (t *UserTag) SetConfidence(confidence float64) {
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}
	t.Confidence = confidence
}

func (t *UserTag) SetWeight(weight float64) {
	if weight < 0 {
		weight = 0
	}
	t.Weight = weight
}

func (t *UserTag) SetExpiration(expiresAt time.Time) {
	t.ExpiresAt = &expiresAt
}

func (t *UserTag) SetValidity(validFrom, validTo time.Time) {
	t.ValidFrom = &validFrom
	t.ValidTo = &validTo
}

func (t *UserTag) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*t.ExpiresAt)
}

func (t *UserTag) IsValid() bool {
	now := time.Now()

	if t.ValidFrom != nil && now.Before(*t.ValidFrom) {
		return false
	}

	if t.ValidTo != nil && now.After(*t.ValidTo) {
		return false
	}

	return t.Status == TagStatusActive && !t.IsExpired()
}

func (t *UserTag) Expire() {
	t.Status = TagStatusExpired
	now := time.Now()
	t.ExpiresAt = &now
}

func (t *UserTag) Invalidate() {
	t.Status = TagStatusInvalid
}

func (t *UserTag) Activate() {
	t.Status = TagStatusActive
	t.ExpiresAt = nil
}

func (t *UserTag) SetMetadata(key string, value any) {
	if t.Metadata == nil {
		t.Metadata = make(map[string]any)
	}
	t.Metadata[key] = value
}

func (t *UserTag) GetMetadata(key string) (any, bool) {
	if t.Metadata == nil {
		return nil, false
	}
	val, ok := t.Metadata[key]
	return val, ok
}

func NewTagDefinition(tagKey, tagName string, category TagCategory, tagType, dataType string) *TagDefinition {
	return &TagDefinition{
		TagKey:        tagKey,
		TagName:       tagName,
		Category:      category,
		TagType:       tagType,
		DataType:      dataType,
		AllowedValues: make([]string, 0),
		Children:      make([]*TagDefinition, 0),
		Enabled:       true,
	}
}

func (d *TagDefinition) AddAllowedValue(value string) {
	if slices.Contains(d.AllowedValues, value) {
		return
	}
	d.AllowedValues = append(d.AllowedValues, value)
}

func (d *TagDefinition) RemoveAllowedValue(value string) {
	for i, v := range d.AllowedValues {
		if v == value {
			d.AllowedValues = append(d.AllowedValues[:i], d.AllowedValues[i+1:]...)
			return
		}
	}
}

func (d *TagDefinition) IsValueAllowed(value string) bool {
	if len(d.AllowedValues) == 0 {
		return true
	}
	return slices.Contains(d.AllowedValues, value)
}

func (d *TagDefinition) AddChild(child *TagDefinition) {
	child.ParentID = d.ID
	d.Children = append(d.Children, child)
}

func (d *TagDefinition) Enable() {
	d.Enabled = true
}

func (d *TagDefinition) Disable() {
	d.Enabled = false
}

type UserTagRepository interface {
	Save(ctx context.Context, tag *UserTag) error
	FindByID(ctx context.Context, id uint64) (*UserTag, error)
	FindByProfileID(ctx context.Context, profileID uint64) ([]*UserTag, error)
	FindByUserID(ctx context.Context, userID uint64) ([]*UserTag, error)
	FindByTagKey(ctx context.Context, userID uint64, tagKey string) (*UserTag, error)
	FindByCategory(ctx context.Context, userID uint64, category TagCategory) ([]*UserTag, error)
	FindExpired(ctx context.Context, limit int) ([]*UserTag, error)
	Update(ctx context.Context, tag *UserTag) error
	Delete(ctx context.Context, id uint64) error
	DeleteByUserID(ctx context.Context, userID uint64) error
}

type TagDefinitionRepository interface {
	Save(ctx context.Context, definition *TagDefinition) error
	FindByID(ctx context.Context, id uint64) (*TagDefinition, error)
	FindByTagKey(ctx context.Context, tagKey string) (*TagDefinition, error)
	FindByCategory(ctx context.Context, category TagCategory) ([]*TagDefinition, error)
	FindAll(ctx context.Context) ([]*TagDefinition, error)
	FindEnabled(ctx context.Context) ([]*TagDefinition, error)
	Update(ctx context.Context, definition *TagDefinition) error
	Delete(ctx context.Context, id uint64) error
}

type TagGroupRepository interface {
	Save(ctx context.Context, group *TagGroup) error
	FindByID(ctx context.Context, id uint64) (*TagGroup, error)
	FindByGroupKey(ctx context.Context, groupKey string) (*TagGroup, error)
	FindByCategory(ctx context.Context, category TagCategory) ([]*TagGroup, error)
	FindAll(ctx context.Context) ([]*TagGroup, error)
	Update(ctx context.Context, group *TagGroup) error
	Delete(ctx context.Context, id uint64) error
}
