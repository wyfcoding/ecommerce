package data

import (
	"time"
)

// AIModelRequest represents a record of an AI model request.
type AIModelRequest struct {
	RequestID   string            `json:"request_id"`
	ModelType   string            `json:"model_type"` // e.g., "text_generation", "text_classification", "recommendation_ranking"
	Input       string            `json:"input"`      // JSON string of input data
	Parameters  map[string]string `json:"parameters"`
	RequestedAt time.Time         `json:"requested_at"`
}

// AIModelResponse represents a record of an AI model response.
type AIModelResponse struct {
	ResponseID  string    `json:"response_id"`
	RequestID   string    `json:"request_id"`
	Output      string    `json:"output"` // JSON string of output data
	Status      string    `json:"status"` // e.g., "SUCCESS", "FAILED"
	RespondedAt time.Time `json:"responded_at"`
	Error       string    `json:"error,omitempty"`
}
