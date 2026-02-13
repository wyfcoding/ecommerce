package domain

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type AppStatus string

const (
	StatusActive   AppStatus = "ACTIVE"
	StatusDisabled AppStatus = "DISABLED"
	StatusPending  AppStatus = "PENDING"
)

type OpenApiApp struct {
	gorm.Model
	AppID        string    `gorm:"column:app_id;type:varchar(32);unique_index;not null"`
	UserID       string    `gorm:"column:user_id;type:varchar(32);index;not null"`
	AppName      string    `gorm:"column:app_name;type:varchar(100);not null"`
	Description  string    `gorm:"column:description;type:text"`
	APIKey       string    `gorm:"column:api_key;type:varchar(64);unique_index;not null"`
	APISecret    string    `gorm:"column:api_secret;type:varchar(128);not null"`
	Status       AppStatus `gorm:"column:status;type:varchar(20);not null;default:'ACTIVE'"`
	Scopes       string    `gorm:"column:scopes;type:text"`
	CallbackURL  string    `gorm:"column:callback_url;type:varchar(512)"`
	IPWhitelist  string    `gorm:"column:ip_whitelist;type:text"`
	RateLimit    int       `gorm:"column:rate_limit;type:int;not null;default:1000"`
	DailyQuota   int64     `gorm:"column:daily_quota;type:bigint;not null;default:100000"`
	UsedQuota    int64     `gorm:"column:used_quota;type:bigint;not null;default:0"`
	ExpiresAt    *time.Time `gorm:"column:expires_at"`
	LastUsedAt   *time.Time `gorm:"column:last_used_at"`
	
	domainEvents []DomainEvent `gorm:"-" json:"-"`
}

func (OpenApiApp) TableName() string { return "openapi_apps" }

var (
	ErrAppNotFound      = errors.New("app not found")
	ErrAppDisabled      = errors.New("app is disabled")
	ErrAppExpired       = errors.New("app has expired")
	ErrQuotaExceeded    = errors.New("daily quota exceeded")
	ErrScopeNotGranted  = errors.New("scope not granted")
)

func NewOpenApiApp(userID, name, description string, scopes []string) *OpenApiApp {
	appID := fmt.Sprintf("app_%d", time.Now().UnixNano())
	apiKey := generateRandomKey(16)
	apiSecret := generateRandomKey(32)
	
	app := &OpenApiApp{
		AppID:       appID,
		UserID:      userID,
		AppName:     name,
		Description: description,
		APIKey:      apiKey,
		APISecret:   apiSecret,
		Status:      StatusActive,
		Scopes:      strings.Join(scopes, ","),
		RateLimit:   1000,
		DailyQuota:  100000,
		domainEvents: make([]DomainEvent, 0),
	}
	
	app.AddDomainEvent(&AppCreatedEvent{
		AppID:     appID,
		UserID:    userID,
		AppName:   name,
		Scopes:    scopes,
		Timestamp: time.Now(),
	})
	
	return app
}

func (a *OpenApiApp) Enable() error {
	if a.Status == StatusActive {
		return errors.New("app is already active")
	}
	a.Status = StatusActive
	
	a.AddDomainEvent(&AppEnabledEvent{
		AppID:     a.AppID,
		Timestamp: time.Now(),
	})
	
	return nil
}

func (a *OpenApiApp) Disable(reason string) error {
	if a.Status == StatusDisabled {
		return errors.New("app is already disabled")
	}
	a.Status = StatusDisabled
	
	a.AddDomainEvent(&AppDisabledEvent{
		AppID:     a.AppID,
		Reason:    reason,
		Timestamp: time.Now(),
	})
	
	return nil
}

func (a *OpenApiApp) RegenerateSecret() string {
	newSecret := generateRandomKey(32)
	a.APISecret = newSecret
	
	a.AddDomainEvent(&AppSecretRegeneratedEvent{
		AppID:     a.AppID,
		Timestamp: time.Now(),
	})
	
	return newSecret
}

func (a *OpenApiApp) UpdateScopes(scopes []string) {
	a.Scopes = strings.Join(scopes, ",")
	
	a.AddDomainEvent(&AppScopesUpdatedEvent{
		AppID:     a.AppID,
		Scopes:    scopes,
		Timestamp: time.Now(),
	})
}

func (a *OpenApiApp) UpdateRateLimit(limit int) {
	a.RateLimit = limit
	
	a.AddDomainEvent(&AppRateLimitUpdatedEvent{
		AppID:     a.AppID,
		RateLimit: limit,
		Timestamp: time.Now(),
	})
}

func (a *OpenApiApp) UpdateQuota(quota int64) {
	a.DailyQuota = quota
	
	a.AddDomainEvent(&AppQuotaUpdatedEvent{
		AppID:     a.AppID,
		DailyQuota: quota,
		Timestamp: time.Now(),
	})
}

func (a *OpenApiApp) RecordUsage() error {
	if a.Status != StatusActive {
		return ErrAppDisabled
	}
	
	if a.ExpiresAt != nil && time.Now().After(*a.ExpiresAt) {
		return ErrAppExpired
	}
	
	if a.UsedQuota >= a.DailyQuota {
		return ErrQuotaExceeded
	}
	
	a.UsedQuota++
	now := time.Now()
	a.LastUsedAt = &now
	
	return nil
}

func (a *OpenApiApp) ResetDailyQuota() {
	a.UsedQuota = 0
}

func (a *OpenApiApp) HasScope(scope string) bool {
	scopes := strings.Split(a.Scopes, ",")
	for _, s := range scopes {
		if strings.TrimSpace(s) == scope {
			return true
		}
	}
	return false
}

func (a *OpenApiApp) IsIPAllowed(ip string) bool {
	if a.IPWhitelist == "" {
		return true
	}
	
	ips := strings.Split(a.IPWhitelist, ",")
	for _, allowedIP := range ips {
		if strings.TrimSpace(allowedIP) == ip {
			return true
		}
	}
	return false
}

func (a *OpenApiApp) AddDomainEvent(event DomainEvent) {
	if a.domainEvents == nil {
		a.domainEvents = make([]DomainEvent, 0)
	}
	a.domainEvents = append(a.domainEvents, event)
}

func (a *OpenApiApp) GetDomainEvents() []DomainEvent {
	return a.domainEvents
}

func (a *OpenApiApp) ClearDomainEvents() {
	a.domainEvents = make([]DomainEvent, 0)
}

func generateRandomKey(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type AppCreatedEvent struct {
	AppID     string    `json:"app_id"`
	UserID    string    `json:"user_id"`
	AppName   string    `json:"app_name"`
	Scopes    []string  `json:"scopes"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *AppCreatedEvent) EventName() string     { return "openapi.app.created" }
func (e *AppCreatedEvent) OccurredAt() time.Time { return e.Timestamp }

type AppEnabledEvent struct {
	AppID     string    `json:"app_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *AppEnabledEvent) EventName() string     { return "openapi.app.enabled" }
func (e *AppEnabledEvent) OccurredAt() time.Time { return e.Timestamp }

type AppDisabledEvent struct {
	AppID     string    `json:"app_id"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *AppDisabledEvent) EventName() string     { return "openapi.app.disabled" }
func (e *AppDisabledEvent) OccurredAt() time.Time { return e.Timestamp }

type AppSecretRegeneratedEvent struct {
	AppID     string    `json:"app_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *AppSecretRegeneratedEvent) EventName() string     { return "openapi.app.secret_regenerated" }
func (e *AppSecretRegeneratedEvent) OccurredAt() time.Time { return e.Timestamp }

type AppScopesUpdatedEvent struct {
	AppID     string    `json:"app_id"`
	Scopes    []string  `json:"scopes"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *AppScopesUpdatedEvent) EventName() string     { return "openapi.app.scopes_updated" }
func (e *AppScopesUpdatedEvent) OccurredAt() time.Time { return e.Timestamp }

type AppRateLimitUpdatedEvent struct {
	AppID     string    `json:"app_id"`
	RateLimit int       `json:"rate_limit"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *AppRateLimitUpdatedEvent) EventName() string     { return "openapi.app.rate_limit_updated" }
func (e *AppRateLimitUpdatedEvent) OccurredAt() time.Time { return e.Timestamp }

type AppQuotaUpdatedEvent struct {
	AppID      string    `json:"app_id"`
	DailyQuota int64     `json:"daily_quota"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e *AppQuotaUpdatedEvent) EventName() string     { return "openapi.app.quota_updated" }
func (e *AppQuotaUpdatedEvent) OccurredAt() time.Time { return e.Timestamp }

type ApiAccessEvent struct {
	AppID     string    `json:"app_id"`
	Path      string    `json:"path"`
	Method    string    `json:"method"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *ApiAccessEvent) EventName() string     { return "openapi.api.access" }
func (e *ApiAccessEvent) OccurredAt() time.Time { return e.Timestamp }
