package service

import (
	"context"
	"errors"
	"time"

	"ecommerce/internal/config/model"
	"ecommerce/internal/config/repository"
)

var (
	ErrConfigNotFound = errors.New("config entry not found")
)

// ConfigService is the business logic for configuration management.
type ConfigService struct {
	repo repository.ConfigRepo
}

// NewConfigService creates a new ConfigService.
func NewConfigService(repo repository.ConfigRepo) *ConfigService {
	return &ConfigService{repo: repo}
}

// GetConfig retrieves a configuration entry by its key.
func (s *ConfigService) GetConfig(ctx context.Context, key string) (*model.ConfigEntry, error) {
	entry, err := s.repo.GetConfig(ctx, key)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, ErrConfigNotFound
	}
	return entry, nil
}

// SetConfig creates or updates a configuration entry.
func (s *ConfigService) SetConfig(ctx context.Context, key, value, description string) (*model.ConfigEntry, error) {
	entry := &model.ConfigEntry{
		Key:         key,
		Value:       value,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	return s.repo.SetConfig(ctx, entry)
}

// ListConfigs lists all configuration entries.
func (s *ConfigService) ListConfigs(ctx context.Context) ([]*model.ConfigEntry, error) {
	return s.repo.ListConfigs(ctx)
}
