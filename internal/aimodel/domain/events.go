package domain

import "time"

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

const (
	AIModelCreatedEventType       = "aimodel.model.created"
	AIModelStatusUpdatedEventType = "aimodel.model.status.updated"
)

type AIModelCreatedEvent struct {
	ModelID   uint64      `json:"model_id"`
	ModelNo   string      `json:"model_no"`
	Status    ModelStatus `json:"status"`
	Timestamp time.Time   `json:"timestamp"`
}

func (e *AIModelCreatedEvent) EventName() string     { return AIModelCreatedEventType }
func (e *AIModelCreatedEvent) OccurredAt() time.Time { return e.Timestamp }

type AIModelStatusUpdatedEvent struct {
	ModelID   uint64      `json:"model_id"`
	OldStatus ModelStatus `json:"old_status"`
	NewStatus ModelStatus `json:"new_status"`
	Timestamp time.Time   `json:"timestamp"`
}

func (e *AIModelStatusUpdatedEvent) EventName() string     { return AIModelStatusUpdatedEventType }
func (e *AIModelStatusUpdatedEvent) OccurredAt() time.Time { return e.Timestamp }

type ModelTrainingStartedEvent struct {
	ModelID   uint64    `json:"model_id"`
	ModelNo   string    `json:"model_no"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *ModelTrainingStartedEvent) EventName() string     { return "aimodel.training.started" }
func (e *ModelTrainingStartedEvent) OccurredAt() time.Time { return e.Timestamp }

type ModelTrainingCompletedEvent struct {
	ModelID   uint64    `json:"model_id"`
	ModelNo   string    `json:"model_no"`
	Accuracy  float64   `json:"accuracy"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *ModelTrainingCompletedEvent) EventName() string     { return "aimodel.training.completed" }
func (e *ModelTrainingCompletedEvent) OccurredAt() time.Time { return e.Timestamp }

type ModelDeployedEvent struct {
	ModelID   uint64    `json:"model_id"`
	ModelNo   string    `json:"model_no"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *ModelDeployedEvent) EventName() string     { return "aimodel.deployed" }
func (e *ModelDeployedEvent) OccurredAt() time.Time { return e.Timestamp }

type ModelPredictionEvent struct {
	ModelID    uint64    `json:"model_id"`
	ModelNo    string    `json:"model_no"`
	UserID     uint64    `json:"user_id"`
	Confidence float64   `json:"confidence"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e *ModelPredictionEvent) EventName() string     { return "aimodel.prediction" }
func (e *ModelPredictionEvent) OccurredAt() time.Time { return e.Timestamp }
