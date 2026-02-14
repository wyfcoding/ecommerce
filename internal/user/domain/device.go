package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

var (
	ErrDeviceNotFound     = errors.New("device not found")
	ErrDeviceAlreadyBound = errors.New("device already bound to another user")
	ErrTooManyDevices     = errors.New("too many devices for this user")
	ErrDeviceNotTrusted   = errors.New("device is not trusted")
	ErrDeviceBanned       = errors.New("device is banned")
)

type DeviceStatus int8

const (
	DeviceStatusActive   DeviceStatus = 1
	DeviceStatusInactive DeviceStatus = 2
	DeviceStatusBanned   DeviceStatus = 3
	DeviceStatusLost     DeviceStatus = 4
)

func (s DeviceStatus) String() string {
	switch s {
	case DeviceStatusActive:
		return "active"
	case DeviceStatusInactive:
		return "inactive"
	case DeviceStatusBanned:
		return "banned"
	case DeviceStatusLost:
		return "lost"
	default:
		return "unknown"
	}
}

type DeviceType string

const (
	DeviceTypeIOS     DeviceType = "ios"
	DeviceTypeAndroid DeviceType = "android"
	DeviceTypeWeb     DeviceType = "web"
	DeviceTypeDesktop DeviceType = "desktop"
	DeviceTypeMini    DeviceType = "mini"
)

type DeviceTrustLevel int8

const (
	DeviceTrustLevelNone    DeviceTrustLevel = 0
	DeviceTrustLevelLow     DeviceTrustLevel = 1
	DeviceTrustLevelMedium  DeviceTrustLevel = 2
	DeviceTrustLevelHigh    DeviceTrustLevel = 3
	DeviceTrustLevelTrusted DeviceTrustLevel = 4
)

type UserDevice struct {
	ID            uint             `json:"id"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	UserID        uint64           `json:"user_id"`
	DeviceID      string           `json:"device_id"`
	DeviceName    string           `json:"device_name"`
	DeviceType    DeviceType       `json:"device_type"`
	OS            string           `json:"os"`
	OSVersion     string           `json:"os_version"`
	AppVersion    string           `json:"app_version"`
	Brand         string           `json:"brand"`
	Model         string           `json:"model"`
	ScreenSize    string           `json:"screen_size"`
	NetworkType   string           `json:"network_type"`
	Carrier       string           `json:"carrier"`
	IMEI          string           `json:"imei"`
	IDFA          string           `json:"idfa"`
	IDFV          string           `json:"idfv"`
	AndroidID     string           `json:"android_id"`
	PushToken     string           `json:"push_token"`
	Status        DeviceStatus     `json:"status"`
	TrustLevel    DeviceTrustLevel `json:"trust_level"`
	LastLoginIP   string           `json:"last_login_ip"`
	LastLoginAt   *time.Time       `json:"last_login_at"`
	LastActiveAt  *time.Time       `json:"last_active_at"`
	LoginCount    int64            `json:"login_count"`
	IsActive      bool             `json:"is_active"`
	IsPrimary     bool             `json:"is_primary"`
	IsPushEnabled bool             `json:"is_push_enabled"`
	Fingerprint   string           `json:"fingerprint"`
	Location      string           `json:"location"`
	Timezone      string           `json:"timezone"`
	Language      string           `json:"language"`
	FirstLoginAt  *time.Time       `json:"first_login_at"`
	VerifiedAt    *time.Time       `json:"verified_at"`
	BannedAt      *time.Time       `json:"banned_at"`
	BannedReason  string           `json:"banned_reason"`
}

type DeviceLoginRecord struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UserID      uint64    `json:"user_id"`
	DeviceID    string    `json:"device_id"`
	LoginIP     string    `json:"login_ip"`
	LoginAt     time.Time `json:"login_at"`
	LoginType   string    `json:"login_type"`
	LoginResult string    `json:"login_result"`
	FailReason  string    `json:"fail_reason"`
	Location    string    `json:"location"`
	Country     string    `json:"country"`
	Province    string    `json:"province"`
	City        string    `json:"city"`
	ISP         string    `json:"isp"`
	UserAgent   string    `json:"user_agent"`
	RiskLevel   int8      `json:"risk_level"`
	RiskTags    string    `json:"risk_tags"`
}

type DeviceVerification struct {
	ID          uint       `json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UserID      uint64     `json:"user_id"`
	DeviceID    string     `json:"device_id"`
	VerifyType  string     `json:"verify_type"`
	VerifyCode  string     `json:"verify_code"`
	ExpiresAt   time.Time  `json:"expires_at"`
	VerifiedAt  *time.Time `json:"verified_at"`
	Attempts    int        `json:"attempts"`
	MaxAttempts int        `json:"max_attempts"`
}

func NewUserDevice(userID uint64, deviceID string, deviceType DeviceType) *UserDevice {
	now := time.Now()
	return &UserDevice{
		UserID:        userID,
		DeviceID:      deviceID,
		DeviceType:    deviceType,
		Status:        DeviceStatusActive,
		TrustLevel:    DeviceTrustLevelNone,
		IsActive:      true,
		IsPushEnabled: true,
		LoginCount:    0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (d *UserDevice) UpdateDeviceInfo(name, os, osVersion, appVersion, brand, model string) {
	if name != "" {
		d.DeviceName = name
	}
	if os != "" {
		d.OS = os
	}
	if osVersion != "" {
		d.OSVersion = osVersion
	}
	if appVersion != "" {
		d.AppVersion = appVersion
	}
	if brand != "" {
		d.Brand = brand
	}
	if model != "" {
		d.Model = model
	}
	d.UpdatedAt = time.Now()
}

func (d *UserDevice) RecordLogin(ip, location string) {
	now := time.Now()
	d.LastLoginIP = ip
	d.LastLoginAt = &now
	d.LastActiveAt = &now
	d.LoginCount++
	if d.FirstLoginAt == nil {
		d.FirstLoginAt = &now
	}
	d.IsActive = true
	d.UpdatedAt = now
}

func (d *UserDevice) RecordActivity() {
	now := time.Now()
	d.LastActiveAt = &now
	d.UpdatedAt = now
}

func (d *UserDevice) SetPrimary(isPrimary bool) {
	d.IsPrimary = isPrimary
	d.UpdatedAt = time.Now()
}

func (d *UserDevice) SetPushEnabled(enabled bool) {
	d.IsPushEnabled = enabled
	d.UpdatedAt = time.Now()
}

func (d *UserDevice) UpdatePushToken(token string) {
	d.PushToken = token
	d.UpdatedAt = time.Now()
}

func (d *UserDevice) IncreaseTrustLevel() {
	if d.TrustLevel < DeviceTrustLevelTrusted {
		d.TrustLevel++
		d.UpdatedAt = time.Now()
	}
}

func (d *UserDevice) DecreaseTrustLevel() {
	if d.TrustLevel > DeviceTrustLevelNone {
		d.TrustLevel--
		d.UpdatedAt = time.Now()
	}
}

func (d *UserDevice) SetTrustLevel(level DeviceTrustLevel) {
	d.TrustLevel = level
	d.UpdatedAt = time.Now()
}

func (d *UserDevice) Ban(reason string) {
	d.Status = DeviceStatusBanned
	now := time.Now()
	d.BannedAt = &now
	d.BannedReason = reason
	d.IsActive = false
	d.UpdatedAt = now
}

func (d *UserDevice) Unban() {
	d.Status = DeviceStatusActive
	d.BannedAt = nil
	d.BannedReason = ""
	d.IsActive = true
	d.UpdatedAt = time.Now()
}

func (d *UserDevice) MarkAsLost() {
	d.Status = DeviceStatusLost
	d.IsActive = false
	d.UpdatedAt = time.Now()
}

func (d *UserDevice) Deactivate() {
	d.Status = DeviceStatusInactive
	d.IsActive = false
	d.UpdatedAt = time.Now()
}

func (d *UserDevice) Activate() {
	d.Status = DeviceStatusActive
	d.IsActive = true
	d.UpdatedAt = time.Now()
}

func (d *UserDevice) Verify() {
	now := time.Now()
	d.VerifiedAt = &now
	d.TrustLevel = DeviceTrustLevelMedium
	d.UpdatedAt = now
}

func (d *UserDevice) IsTrusted() bool {
	return d.TrustLevel >= DeviceTrustLevelHigh && d.Status == DeviceStatusActive
}

func (d *UserDevice) CanLogin() bool {
	return d.Status != DeviceStatusBanned && d.IsActive
}

func (d *UserDevice) GenerateFingerprint() string {
	data := string(d.DeviceType) + d.OS + d.OSVersion + d.Brand + d.Model + d.IMEI + d.AndroidID + d.IDFA + d.IDFV
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func NewDeviceLoginRecord(userID uint64, deviceID, ip string, loginType, result string) *DeviceLoginRecord {
	return &DeviceLoginRecord{
		UserID:      userID,
		DeviceID:    deviceID,
		LoginIP:     ip,
		LoginAt:     time.Now(),
		LoginType:   loginType,
		LoginResult: result,
	}
}

func (r *DeviceLoginRecord) SetLocation(country, province, city, isp string) {
	r.Country = country
	r.Province = province
	r.City = city
	r.ISP = isp
	r.Location = country + " " + province + " " + city
}

func (r *DeviceLoginRecord) SetRisk(level int8, tags string) {
	r.RiskLevel = level
	r.RiskTags = tags
}

func NewDeviceVerification(userID uint64, deviceID, verifyType string, maxAttempts int) *DeviceVerification {
	return &DeviceVerification{
		UserID:      userID,
		DeviceID:    deviceID,
		VerifyType:  verifyType,
		ExpiresAt:   time.Now().Add(15 * time.Minute),
		MaxAttempts: maxAttempts,
		Attempts:    0,
		CreatedAt:   time.Now(),
	}
}

func (v *DeviceVerification) SetCode(code string) {
	v.VerifyCode = code
}

func (v *DeviceVerification) Verify(code string) bool {
	if v.IsExpired() || v.Attempts >= v.MaxAttempts {
		return false
	}
	v.Attempts++
	if v.VerifyCode == code {
		now := time.Now()
		v.VerifiedAt = &now
		return true
	}
	return false
}

func (v *DeviceVerification) IsExpired() bool {
	return time.Now().After(v.ExpiresAt)
}

func (v *DeviceVerification) IsVerified() bool {
	return v.VerifiedAt != nil
}

func (v *DeviceVerification) RemainingAttempts() int {
	return v.MaxAttempts - v.Attempts
}

type DeviceRepository interface {
	FindByID(ctx any, id uint) (*UserDevice, error)
	FindByDeviceID(ctx any, deviceID string) (*UserDevice, error)
	FindByUserID(ctx any, userID uint64) ([]*UserDevice, error)
	FindActiveByUserID(ctx any, userID uint64) ([]*UserDevice, error)
	FindPrimaryByUserID(ctx any, userID uint64) (*UserDevice, error)
	Save(ctx any, device *UserDevice) error
	Update(ctx any, device *UserDevice) error
	Delete(ctx any, id uint) error
	CountByUserID(ctx any, userID uint64) (int64, error)

	SaveLoginRecord(ctx any, record *DeviceLoginRecord) error
	FindLoginRecords(ctx any, userID uint64, limit, offset int) ([]*DeviceLoginRecord, error)
	FindRecentLogins(ctx any, userID uint64, duration time.Duration) ([]*DeviceLoginRecord, error)

	SaveVerification(ctx any, verification *DeviceVerification) error
	FindVerification(ctx any, userID uint64, deviceID string) (*DeviceVerification, error)
}

type DeviceSecurityService interface {
	CheckDeviceRisk(device *UserDevice, ip string) (int8, []string, error)
	IsNewDevice(userID uint64, deviceID string) bool
	ShouldVerifyDevice(device *UserDevice) bool
	SendVerificationCode(userID uint64, deviceID string, verifyType string) error
}
