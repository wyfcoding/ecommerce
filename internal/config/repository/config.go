package repository

import (
	"context"

	"ecommerce/internal/config/model"
)

// ConfigRepo defines the interface for configuration data access.
type ConfigRepo interface {
	GetConfig(ctx context.Context, key string) (*model.ConfigEntry, error)
	SetConfig(ctx context.Context, entry *model.ConfigEntry) (*model.ConfigEntry, error)
	ListConfigs(ctx context.Context) ([]*model.ConfigEntry, error)
}