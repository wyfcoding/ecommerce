package biz

import (
	"context"
	"errors"
	"time"
)

var (
	ErrConfigNotFound = errors.New("config entry not found")
)

// ConfigEntry represents a configuration entry in the business logic layer.
type ConfigEntry struct {
	ID          uint
	Key         string
	Value       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ConfigRepo defines the interface for configuration data access.
type ConfigRepo interface {
	GetConfig(ctx context.Context, key string) (*ConfigEntry, error)
	SetConfig(ctx context.Context, entry *ConfigEntry) (*ConfigEntry, error)
	ListConfigs(ctx context.Context) ([]*ConfigEntry, error)
}

// ConfigUsecase is the business logic for configuration management.
type ConfigUsecase struct {
	repo ConfigRepo
}

// NewConfigUsecase creates a new ConfigUsecase.
func NewConfigUsecase(repo ConfigRepo) *ConfigUsecase {
	return &ConfigUsecase{repo: repo}
}

// GetConfig retrieves a configuration entry by its key.
func (uc *ConfigUsecase) GetConfig(ctx context.Context, key string) (*ConfigEntry, error) {
	entry, err := uc.repo.GetConfig(ctx, key)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, ErrConfigNotFound
	}
	return entry, nil
}

// SetConfig creates or updates a configuration entry.
func (uc *ConfigUsecase) SetConfig(ctx context.Context, key, value, description string) (*ConfigEntry, error) {
	entry := &ConfigEntry{
		Key:         key,
		Value:       value,
		Description: description,
	}
	return uc.repo.SetConfig(ctx, entry)
}

// ListConfigs lists all configuration entries.
func (uc *ConfigUsecase) ListConfigs(ctx context.Context) ([]*ConfigEntry, error) {
	return uc.repo.ListConfigs(ctx)
}
