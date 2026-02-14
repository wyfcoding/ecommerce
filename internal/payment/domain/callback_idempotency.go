package domain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

var (
	ErrIdempotencyKeyNotFound   = errors.New("idempotency key not found")
	ErrIdempotencyKeyExpired    = errors.New("idempotency key expired")
	ErrDuplicateCallback        = errors.New("duplicate callback detected")
	ErrCallbackProcessing       = errors.New("callback is processing")
	ErrCallbackAlreadyProcessed = errors.New("callback already processed")
	ErrInvalidCallbackSignature = errors.New("invalid callback signature")
)

type CallbackStatus int8

const (
	CallbackStatusPending    CallbackStatus = 0
	CallbackStatusProcessing CallbackStatus = 1
	CallbackStatusSuccess    CallbackStatus = 2
	CallbackStatusFailed     CallbackStatus = 3
	CallbackStatusDuplicate  CallbackStatus = 4
)

func (s CallbackStatus) String() string {
	switch s {
	case CallbackStatusPending:
		return "PENDING"
	case CallbackStatusProcessing:
		return "PROCESSING"
	case CallbackStatusSuccess:
		return "SUCCESS"
	case CallbackStatusFailed:
		return "FAILED"
	case CallbackStatusDuplicate:
		return "DUPLICATE"
	default:
		return "UNKNOWN"
	}
}

type IdempotencyKey struct {
	ID           uint           `json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	Key          string         `json:"key"`
	KeyHash      string         `json:"key_hash"`
	BusinessType string         `json:"business_type"`
	BusinessID   uint64         `json:"business_id"`
	Status       CallbackStatus `json:"status"`
	ResponseData string         `json:"response_data"`
	ExpiresAt    time.Time      `json:"expires_at"`
	ProcessedAt  *time.Time     `json:"processed_at"`
	Attempts     int            `json:"attempts"`
}

type CallbackRecord struct {
	ID              uint           `json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	CallbackNo      string         `json:"callback_no"`
	IdempotencyKey  string         `json:"idempotency_key"`
	Channel         string         `json:"channel"`
	ChannelCallback string         `json:"channel_callback"`
	BusinessType    string         `json:"business_type"`
	BusinessID      uint64         `json:"business_id"`
	Status          CallbackStatus `json:"status"`
	RawData         string         `json:"raw_data"`
	ParsedData      string         `json:"parsed_data"`
	Signature       string         `json:"signature"`
	SignatureValid  bool           `json:"signature_valid"`
	ProcessedAt     *time.Time     `json:"processed_at"`
	ProcessResult   string         `json:"process_result"`
	ProcessError    string         `json:"process_error"`
	ProcessDuration int64          `json:"process_duration"`
	IPAddress       string         `json:"ip_address"`
	UserAgent       string         `json:"user_agent"`
}

type IdempotencyConfig struct {
	KeyTTL            time.Duration
	MaxKeyLength      int
	EnableHashing     bool
	HashAlgorithm     string
	EnableDedupWindow bool
	DedupWindow       time.Duration
}

func DefaultIdempotencyConfig() *IdempotencyConfig {
	return &IdempotencyConfig{
		KeyTTL:            time.Hour * 24 * 7,
		MaxKeyLength:      256,
		EnableHashing:     true,
		HashAlgorithm:     "sha256",
		EnableDedupWindow: true,
		DedupWindow:       time.Minute * 5,
	}
}

func NewIdempotencyKey(key, businessType string, businessID uint64, ttl time.Duration) *IdempotencyKey {
	now := time.Now()
	return &IdempotencyKey{
		Key:          key,
		KeyHash:      hashKey(key),
		BusinessType: businessType,
		BusinessID:   businessID,
		Status:       CallbackStatusPending,
		ExpiresAt:    now.Add(ttl),
		Attempts:     0,
	}
}

func NewCallbackRecord(callbackNo, idempotencyKey, channel, businessType string, businessID uint64, rawData string) *CallbackRecord {
	return &CallbackRecord{
		CallbackNo:     callbackNo,
		IdempotencyKey: idempotencyKey,
		Channel:        channel,
		BusinessType:   businessType,
		BusinessID:     businessID,
		Status:         CallbackStatusPending,
		RawData:        rawData,
		SignatureValid: false,
	}
}

func (k *IdempotencyKey) IsExpired() bool {
	return time.Now().After(k.ExpiresAt)
}

func (k *IdempotencyKey) IsProcessed() bool {
	return k.Status == CallbackStatusSuccess || k.Status == CallbackStatusFailed
}

func (k *IdempotencyKey) IsProcessing() bool {
	return k.Status == CallbackStatusProcessing
}

func (k *IdempotencyKey) MarkProcessing() {
	k.Status = CallbackStatusProcessing
	k.Attempts++
}

func (k *IdempotencyKey) MarkSuccess(responseData string) {
	now := time.Now()
	k.Status = CallbackStatusSuccess
	k.ResponseData = responseData
	k.ProcessedAt = &now
}

func (k *IdempotencyKey) MarkFailed() {
	k.Status = CallbackStatusFailed
}

func (r *CallbackRecord) SetParsedData(data string) {
	r.ParsedData = data
}

func (r *CallbackRecord) SetSignature(signature string, valid bool) {
	r.Signature = signature
	r.SignatureValid = valid
}

func (r *CallbackRecord) SetRequestInfo(ip, userAgent string) {
	r.IPAddress = ip
	r.UserAgent = userAgent
}

func (r *CallbackRecord) MarkProcessing() {
	r.Status = CallbackStatusProcessing
}

func (r *CallbackRecord) MarkSuccess(result string, duration int64) {
	now := time.Now()
	r.Status = CallbackStatusSuccess
	r.ProcessedAt = &now
	r.ProcessResult = result
	r.ProcessDuration = duration
}

func (r *CallbackRecord) MarkFailed(err string, duration int64) {
	now := time.Now()
	r.Status = CallbackStatusFailed
	r.ProcessedAt = &now
	r.ProcessError = err
	r.ProcessDuration = duration
}

func (r *CallbackRecord) MarkDuplicate() {
	r.Status = CallbackStatusDuplicate
}

type IdempotencyService struct {
	config     *IdempotencyConfig
	repository IdempotencyRepository
}

func NewIdempotencyService(config *IdempotencyConfig, repo IdempotencyRepository) *IdempotencyService {
	return &IdempotencyService{
		config:     config,
		repository: repo,
	}
}

func (s *IdempotencyService) AcquireKey(ctx context.Context, key, businessType string, businessID uint64) (*IdempotencyKey, error) {
	existingKey, err := s.repository.FindByKey(ctx, key)
	if err == nil && existingKey != nil {
		if existingKey.IsExpired() {
			return nil, ErrIdempotencyKeyExpired
		}

		if existingKey.IsProcessed() {
			return existingKey, ErrCallbackAlreadyProcessed
		}

		if existingKey.IsProcessing() {
			return existingKey, ErrCallbackProcessing
		}

		return existingKey, nil
	}

	newKey := NewIdempotencyKey(key, businessType, businessID, s.config.KeyTTL)

	if err := s.repository.Save(ctx, newKey); err != nil {
		return nil, fmt.Errorf("failed to save idempotency key: %w", err)
	}

	return newKey, nil
}

func (s *IdempotencyService) ReleaseKey(ctx context.Context, key string, status CallbackStatus, responseData string) error {
	idempotencyKey, err := s.repository.FindByKey(ctx, key)
	if err != nil {
		return ErrIdempotencyKeyNotFound
	}

	switch status {
	case CallbackStatusSuccess:
		idempotencyKey.MarkSuccess(responseData)
	case CallbackStatusFailed:
		idempotencyKey.MarkFailed()
	}

	return s.repository.Update(ctx, idempotencyKey)
}

func (s *IdempotencyService) GetCachedResponse(ctx context.Context, key string) (string, error) {
	idempotencyKey, err := s.repository.FindByKey(ctx, key)
	if err != nil {
		return "", ErrIdempotencyKeyNotFound
	}

	if idempotencyKey.IsExpired() {
		return "", ErrIdempotencyKeyExpired
	}

	if !idempotencyKey.IsProcessed() {
		return "", ErrCallbackProcessing
	}

	return idempotencyKey.ResponseData, nil
}

func (s *IdempotencyService) CheckDuplicate(ctx context.Context, key string) (bool, *IdempotencyKey, error) {
	idempotencyKey, err := s.repository.FindByKey(ctx, key)
	if err != nil {
		return false, nil, nil
	}

	if idempotencyKey.IsExpired() {
		return false, nil, ErrIdempotencyKeyExpired
	}

	return idempotencyKey.IsProcessed(), idempotencyKey, nil
}

type CallbackProcessor interface {
	Process(ctx context.Context, callback *CallbackRecord) (string, error)
}

type CallbackValidator interface {
	ValidateSignature(ctx context.Context, channel, rawData, signature string) bool
	ValidateTimestamp(ctx context.Context, timestamp int64, maxDelay time.Duration) bool
	ValidateAmount(ctx context.Context, businessID uint64, amount int64) bool
}

type CallbackHandler struct {
	idempotencyService *IdempotencyService
	processor          CallbackProcessor
	validator          CallbackValidator
	repository         CallbackRepository
	config             *IdempotencyConfig
}

func NewCallbackHandler(idempotencyService *IdempotencyService, processor CallbackProcessor, validator CallbackValidator, repo CallbackRepository, config *IdempotencyConfig) *CallbackHandler {
	return &CallbackHandler{
		idempotencyService: idempotencyService,
		processor:          processor,
		validator:          validator,
		repository:         repo,
		config:             config,
	}
}

func (h *CallbackHandler) Handle(ctx context.Context, callback *CallbackRecord) (string, error) {
	isDuplicate, existingKey, err := h.idempotencyService.CheckDuplicate(ctx, callback.IdempotencyKey)
	if err != nil && err != ErrIdempotencyKeyNotFound {
		return "", err
	}

	if isDuplicate && existingKey != nil {
		callback.MarkDuplicate()
		h.repository.Save(ctx, callback)
		return existingKey.ResponseData, ErrDuplicateCallback
	}

	idempotencyKey, err := h.idempotencyService.AcquireKey(ctx, callback.IdempotencyKey, callback.BusinessType, callback.BusinessID)
	if err != nil {
		if err == ErrCallbackAlreadyProcessed {
			callback.MarkDuplicate()
			h.repository.Save(ctx, callback)
			return idempotencyKey.ResponseData, ErrDuplicateCallback
		}
		if err == ErrCallbackProcessing {
			return "", ErrCallbackProcessing
		}
		return "", err
	}

	if !callback.SignatureValid {
		if h.validator != nil {
			callback.SignatureValid = h.validator.ValidateSignature(ctx, callback.Channel, callback.RawData, callback.Signature)
		}
		if !callback.SignatureValid {
			callback.MarkFailed("invalid signature", 0)
			h.repository.Save(ctx, callback)
			return "", ErrInvalidCallbackSignature
		}
	}

	callback.MarkProcessing()
	h.repository.Save(ctx, callback)

	startTime := time.Now()
	result, processErr := h.processor.Process(ctx, callback)
	duration := time.Since(startTime).Milliseconds()

	if processErr != nil {
		callback.MarkFailed(processErr.Error(), duration)
		h.idempotencyService.ReleaseKey(ctx, callback.IdempotencyKey, CallbackStatusFailed, "")
		h.repository.Update(ctx, callback)
		return "", processErr
	}

	callback.MarkSuccess(result, duration)
	h.idempotencyService.ReleaseKey(ctx, callback.IdempotencyKey, CallbackStatusSuccess, result)
	h.repository.Update(ctx, callback)

	return result, nil
}

func (h *CallbackHandler) HandleWithRetry(ctx context.Context, callback *CallbackRecord, maxRetries int) (string, error) {
	var lastErr error
	for i := range maxRetries {
		result, err := h.Handle(ctx, callback)
		if err == nil {
			return result, nil
		}

		if err == ErrDuplicateCallback || err == ErrInvalidCallbackSignature {
			return result, err
		}

		lastErr = err

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Duration(i+1) * time.Second):
		}
	}

	return "", lastErr
}

func hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

type IdempotencyRepository interface {
	Save(ctx context.Context, key *IdempotencyKey) error
	FindByKey(ctx context.Context, key string) (*IdempotencyKey, error)
	FindByID(ctx context.Context, id uint64) (*IdempotencyKey, error)
	Update(ctx context.Context, key *IdempotencyKey) error
	DeleteExpired(ctx context.Context) error
}

type CallbackRepository interface {
	Save(ctx context.Context, record *CallbackRecord) error
	FindByID(ctx context.Context, id uint64) (*CallbackRecord, error)
	FindByCallbackNo(ctx context.Context, callbackNo string) (*CallbackRecord, error)
	FindByIdempotencyKey(ctx context.Context, key string) ([]*CallbackRecord, error)
	FindByBusinessID(ctx context.Context, businessType string, businessID uint64) ([]*CallbackRecord, error)
	Update(ctx context.Context, record *CallbackRecord) error
}
