package domain

import (
	"context"
	"time"
)

type KYCApplication struct {
	ApplicationID string    `json:"application_id"`
	UserID        uint64    `json:"user_id"`
	FullName      string    `json:"full_name"`
	IDNumber      string    `json:"id_number"`
	IDType        string    `json:"id_type"`
	Status        string    `json:"status"`
	Reason        string    `json:"reason"`
	CreatedAt     time.Time `json:"created_at"`
	VerifiedAt    time.Time `json:"verified_at"`
}

type KYCRepository interface {
	Save(ctx context.Context, app *KYCApplication) error
	FindByUserID(ctx context.Context, userID uint64) (*KYCApplication, error)
	FindByID(ctx context.Context, appID string) (*KYCApplication, error)
}
