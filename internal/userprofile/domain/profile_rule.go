package domain

import (
	"context"
	"time"
)

type ProfileRule struct {
	ID           uint64        `json:"id"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	RuleNo       string        `json:"rule_no"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	RuleType     ProfileRuleType `json:"rule_type"`
	Category     TagCategory   `json:"category"`
	Conditions   []*RuleCondition `json:"conditions"`
	Actions      []*RuleAction `json:"actions"`
	Priority     int           `json:"priority"`
	Enabled      bool          `json:"enabled"`
	StartTime    *time.Time    `json:"start_time"`
	EndTime      *time.Time    `json:"end_time"`
	Version      int           `json:"version"`
	CreatedBy    uint64        `json:"created_by"`
	UpdatedBy    uint64        `json:"updated_by"`
}

type ProfileRuleType int8

const (
	RuleTypeTagging      ProfileRuleType = 1
	RuleTypeScoring      ProfileRuleType = 2
	RuleTypeSegmentation ProfileRuleType = 3
	RuleTypeAlert        ProfileRuleType = 4
)

func (t ProfileRuleType) String() string {
	switch t {
	case RuleTypeTagging:
		return "TAGGING"
	case RuleTypeScoring:
		return "SCORING"
	case RuleTypeSegmentation:
		return "SEGMENTATION"
	case RuleTypeAlert:
		return "ALERT"
	default:
		return "UNKNOWN"
	}
}

type RuleCondition struct {
	ID         uint64    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	RuleID     uint64    `json:"rule_id"`
	Field      string    `json:"field"`
	Operator   string    `json:"operator"`
	Value      string    `json:"value"`
	LogicType  string    `json:"logic_type"`
	Score      int       `json:"score"`
	Weight     float64   `json:"weight"`
}

type RuleAction struct {
	ID          uint64    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	RuleID      uint64    `json:"rule_id"`
	ActionType  string    `json:"action_type"`
	TargetField string    `json:"target_field"`
	Value       string    `json:"value"`
	Expression  string    `json:"expression"`
}

type ProfileSegment struct {
	ID           uint64          `json:"id"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	SegmentNo    string          `json:"segment_no"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	SegmentType  SegmentType     `json:"segment_type"`
	Criteria     []*SegmentCriteria `json:"criteria"`
	UserCount    int64           `json:"user_count"`
	LastCalculatedAt *time.Time  `json:"last_calculated_at"`
	Enabled      bool            `json:"enabled"`
	Priority     int             `json:"priority"`
}

type SegmentType int8

const (
	SegmentTypeStatic  SegmentType = 1
	SegmentTypeDynamic SegmentType = 2
)

func (t SegmentType) String() string {
	switch t {
	case SegmentTypeStatic:
		return "STATIC"
	case SegmentTypeDynamic:
		return "DYNAMIC"
	default:
		return "UNKNOWN"
	}
}

type SegmentCriteria struct {
	ID         uint64    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	SegmentID  uint64    `json:"segment_id"`
	Field      string    `json:"field"`
	Operator   string    `json:"operator"`
	Value      string    `json:"value"`
	LogicType  string    `json:"logic_type"`
}

func NewProfileRule(ruleNo, name string, ruleType ProfileRuleType) *ProfileRule {
	return &ProfileRule{
		RuleNo:     ruleNo,
		Name:       name,
		RuleType:   ruleType,
		Conditions: make([]*RuleCondition, 0),
		Actions:    make([]*RuleAction, 0),
		Enabled:    true,
		Version:    1,
	}
}

func (r *ProfileRule) AddCondition(field, operator, value, logicType string, score int, weight float64) *RuleCondition {
	condition := &RuleCondition{
		RuleID:    r.ID,
		Field:     field,
		Operator:  operator,
		Value:     value,
		LogicType: logicType,
		Score:     score,
		Weight:    weight,
	}
	r.Conditions = append(r.Conditions, condition)
	return condition
}

func (r *ProfileRule) AddAction(actionType, targetField, value, expression string) *RuleAction {
	action := &RuleAction{
		RuleID:      r.ID,
		ActionType:  actionType,
		TargetField: targetField,
		Value:       value,
		Expression:  expression,
	}
	r.Actions = append(r.Actions, action)
	return action
}

func (r *ProfileRule) Enable() {
	r.Enabled = true
}

func (r *ProfileRule) Disable() {
	r.Enabled = false
}

func (r *ProfileRule) IsEffective() bool {
	now := time.Now()

	if r.StartTime != nil && now.Before(*r.StartTime) {
		return false
	}

	if r.EndTime != nil && now.After(*r.EndTime) {
		return false
	}

	return r.Enabled
}

func NewProfileSegment(segmentNo, name string, segmentType SegmentType) *ProfileSegment {
	return &ProfileSegment{
		SegmentNo:   segmentNo,
		Name:        name,
		SegmentType: segmentType,
		Criteria:    make([]*SegmentCriteria, 0),
		Enabled:     true,
	}
}

func (s *ProfileSegment) AddCriteria(field, operator, value, logicType string) *SegmentCriteria {
	criteria := &SegmentCriteria{
		SegmentID: s.ID,
		Field:     field,
		Operator:  operator,
		Value:     value,
		LogicType: logicType,
	}
	s.Criteria = append(s.Criteria, criteria)
	return criteria
}

func (s *ProfileSegment) UpdateUserCount(count int64) {
	s.UserCount = count
	now := time.Now()
	s.LastCalculatedAt = &now
}

func (s *ProfileSegment) Enable() {
	s.Enabled = true
}

func (s *ProfileSegment) Disable() {
	s.Enabled = false
}

type ProfileRuleRepository interface {
	Save(ctx context.Context, rule *ProfileRule) error
	FindByID(ctx context.Context, id uint64) (*ProfileRule, error)
	FindByRuleNo(ctx context.Context, ruleNo string) (*ProfileRule, error)
	FindByType(ctx context.Context, ruleType ProfileRuleType) ([]*ProfileRule, error)
	FindEnabled(ctx context.Context) ([]*ProfileRule, error)
	FindEffective(ctx context.Context) ([]*ProfileRule, error)
	Update(ctx context.Context, rule *ProfileRule) error
	Delete(ctx context.Context, id uint64) error
}

type ProfileSegmentRepository interface {
	Save(ctx context.Context, segment *ProfileSegment) error
	FindByID(ctx context.Context, id uint64) (*ProfileSegment, error)
	FindBySegmentNo(ctx context.Context, segmentNo string) (*ProfileSegment, error)
	FindEnabled(ctx context.Context) ([]*ProfileSegment, error)
	FindDynamic(ctx context.Context) ([]*ProfileSegment, error)
	Update(ctx context.Context, segment *ProfileSegment) error
	Delete(ctx context.Context, id uint64) error
}
