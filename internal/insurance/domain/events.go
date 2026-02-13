package domain

import "time"

type PolicyCreatedEvent struct {
	PolicyID  string
	OrderID   string
	UserID    string
	Timestamp time.Time
}

func (e *PolicyCreatedEvent) EventName() string { return "insurance.policy.created" }

type ClaimFiledEvent struct {
	ClaimID   string
	PolicyID  string
	UserID    string
	Timestamp time.Time
}

func (e *ClaimFiledEvent) EventName() string { return "insurance.claim.filed" }

type ClaimProcessedEvent struct {
	ClaimID   string
	Status    string
	Timestamp time.Time
}

func (e *ClaimProcessedEvent) EventName() string { return "insurance.claim.processed" }
