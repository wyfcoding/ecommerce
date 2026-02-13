package domain

import (
	"time"
)

const (
	EventTypeProfileCreated     = "user_profile.created"
	EventTypeProfileUpdated     = "user_profile.updated"
	EventTypeProfileRecalculated = "user_profile.recalculated"
	EventTypeProfileExpired     = "user_profile.expired"
	EventTypeProfileArchived    = "user_profile.archived"

	EventTypeTagAdded    = "user_tag.added"
	EventTypeTagUpdated  = "user_tag.updated"
	EventTypeTagRemoved  = "user_tag.removed"
	EventTypeTagExpired  = "user_tag.expired"

	EventTypeBehaviorRecorded = "behavior.recorded"
	EventTypePatternDetected  = "pattern.detected"

	EventTypePreferenceUpdated = "preference.updated"

	EventTypeConsumptionUpdated = "consumption.updated"
	EventTypeSegmentChanged     = "segment.changed"
	EventTypeRiskLevelChanged   = "risk_level.changed"

	EventTypeRuleTriggered = "rule.triggered"
)

type ProfileEvent struct {
	ID            uint64    `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	EventType     string    `json:"event_type"`
	AggregateID   uint64    `json:"aggregate_id"`
	AggregateType string    `json:"aggregate_type"`
	UserID        uint64    `json:"user_id"`
	Payload       map[string]any `json:"payload"`
	Metadata      map[string]any `json:"metadata"`
	Source        string    `json:"source"`
	Version       int       `json:"version"`
}

func NewProfileEvent(eventType string, aggregateID uint64, aggregateType string, userID uint64) *ProfileEvent {
	return &ProfileEvent{
		CreatedAt:     time.Now(),
		EventType:     eventType,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		UserID:        userID,
		Payload:       make(map[string]any),
		Metadata:      make(map[string]any),
		Version:       1,
	}
}

func (e *ProfileEvent) SetPayload(key string, value any) {
	e.Payload[key] = value
}

func (e *ProfileEvent) SetMetadata(key string, value any) {
	e.Metadata[key] = value
}

func (e *ProfileEvent) SetSource(source string) {
	e.Source = source
}

type ProfileCreatedEvent struct {
	*ProfileEvent
	UserID    uint64 `json:"user_id"`
	Status    string `json:"status"`
	CreatedBy string `json:"created_by"`
}

func NewProfileCreatedEvent(profileID, userID uint64) *ProfileCreatedEvent {
	event := NewProfileEvent(EventTypeProfileCreated, profileID, "user_profile", userID)
	return &ProfileCreatedEvent{
		ProfileEvent: event,
		UserID:       userID,
		Status:       "ACTIVE",
	}
}

type ProfileUpdatedEvent struct {
	*ProfileEvent
	UserID        uint64 `json:"user_id"`
	UpdatedFields []string `json:"updated_fields"`
	OldValues     map[string]any `json:"old_values"`
	NewValues     map[string]any `json:"new_values"`
}

func NewProfileUpdatedEvent(profileID, userID uint64, fields []string) *ProfileUpdatedEvent {
	event := NewProfileEvent(EventTypeProfileUpdated, profileID, "user_profile", userID)
	return &ProfileUpdatedEvent{
		ProfileEvent:  event,
		UserID:        userID,
		UpdatedFields: fields,
		OldValues:     make(map[string]any),
		NewValues:     make(map[string]any),
	}
}

type TagAddedEvent struct {
	*ProfileEvent
	UserID    uint64   `json:"user_id"`
	TagKey    string   `json:"tag_key"`
	TagValue  string   `json:"tag_value"`
	Category  string   `json:"category"`
	Source    string   `json:"source"`
}

func NewTagAddedEvent(profileID, userID uint64, tagKey, tagValue string, category TagCategory) *TagAddedEvent {
	event := NewProfileEvent(EventTypeTagAdded, profileID, "user_tag", userID)
	return &TagAddedEvent{
		ProfileEvent: event,
		UserID:       userID,
		TagKey:       tagKey,
		TagValue:     tagValue,
		Category:     category.String(),
	}
}

type TagRemovedEvent struct {
	*ProfileEvent
	UserID   uint64 `json:"user_id"`
	TagKey   string `json:"tag_key"`
	TagValue string `json:"tag_value"`
}

func NewTagRemovedEvent(profileID, userID uint64, tagKey, tagValue string) *TagRemovedEvent {
	event := NewProfileEvent(EventTypeTagRemoved, profileID, "user_tag", userID)
	return &TagRemovedEvent{
		ProfileEvent: event,
		UserID:       userID,
		TagKey:       tagKey,
		TagValue:     tagValue,
	}
}

type BehaviorRecordedEvent struct {
	*ProfileEvent
	UserID       uint64 `json:"user_id"`
	BehaviorType string `json:"behavior_type"`
	TargetType   string `json:"target_type"`
	TargetID     uint64 `json:"target_id"`
	Value        string `json:"value"`
}

func NewBehaviorRecordedEvent(profileID, userID uint64, behaviorType, targetType string, targetID uint64) *BehaviorRecordedEvent {
	event := NewProfileEvent(EventTypeBehaviorRecorded, profileID, "behavior", userID)
	return &BehaviorRecordedEvent{
		ProfileEvent: event,
		UserID:       userID,
		BehaviorType: behaviorType,
		TargetType:   targetType,
		TargetID:     targetID,
	}
}

type SegmentChangedEvent struct {
	*ProfileEvent
	UserID     uint64 `json:"user_id"`
	OldSegment string `json:"old_segment"`
	NewSegment string `json:"new_segment"`
	Reason     string `json:"reason"`
}

func NewSegmentChangedEvent(profileID, userID uint64, oldSegment, newSegment, reason string) *SegmentChangedEvent {
	event := NewProfileEvent(EventTypeSegmentChanged, profileID, "user_profile", userID)
	return &SegmentChangedEvent{
		ProfileEvent: event,
		UserID:       userID,
		OldSegment:   oldSegment,
		NewSegment:   newSegment,
		Reason:       reason,
	}
}

type RiskLevelChangedEvent struct {
	*ProfileEvent
	UserID     uint64 `json:"user_id"`
	OldLevel   int    `json:"old_level"`
	NewLevel   int    `json:"new_level"`
	OldScore   float64 `json:"old_score"`
	NewScore   float64 `json:"new_score"`
	Reason     string `json:"reason"`
}

func NewRiskLevelChangedEvent(profileID, userID uint64, oldLevel, newLevel int, oldScore, newScore float64, reason string) *RiskLevelChangedEvent {
	event := NewProfileEvent(EventTypeRiskLevelChanged, profileID, "user_profile", userID)
	return &RiskLevelChangedEvent{
		ProfileEvent: event,
		UserID:       userID,
		OldLevel:     oldLevel,
		NewLevel:     newLevel,
		OldScore:     oldScore,
		NewScore:     newScore,
		Reason:       reason,
	}
}

type RuleTriggeredEvent struct {
	*ProfileEvent
	UserID   uint64 `json:"user_id"`
	RuleID   uint64 `json:"rule_id"`
	RuleName string `json:"rule_name"`
	RuleType string `json:"rule_type"`
	Action   string `json:"action"`
	Result   string `json:"result"`
}

func NewRuleTriggeredEvent(profileID, userID, ruleID uint64, ruleName string, ruleType ProfileRuleType, action, result string) *RuleTriggeredEvent {
	event := NewProfileEvent(EventTypeRuleTriggered, profileID, "rule", userID)
	return &RuleTriggeredEvent{
		ProfileEvent: event,
		UserID:       userID,
		RuleID:       ruleID,
		RuleName:     ruleName,
		RuleType:     ruleType.String(),
		Action:       action,
		Result:       result,
	}
}
