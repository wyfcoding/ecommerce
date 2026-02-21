// 变更说明：
// 新增用户中心统一领域模型，整合 identity/kyc/userprofile 的核心能力。
// 1. identity 服务为空壳，其身份认证能力直接在 user 服务中实现。
// 2. kyc 实名认证作为 user 的子流程，通过 KYCInfo 值对象嵌入。
// 3. userprofile 的用户画像通过 UserProfile 值对象关联。
// 这样避免了用户注册/登录/认证/画像分散在 4 个服务中导致的跨服务调用开销。
package domain

import (
	"errors"
	"time"
)

// 用户中心扩展错误定义。
var (
	ErrKYCAlreadyVerified = errors.New("user KYC already verified")
	ErrKYCPending         = errors.New("user KYC verification is pending")
	ErrKYCRejected        = errors.New("user KYC verification was rejected")
	ErrMFARequired        = errors.New("multi-factor authentication required")
	ErrMFACodeInvalid     = errors.New("MFA code is invalid")
	ErrOAuthProviderNotSupported = errors.New("OAuth provider not supported")
	ErrAccountLocked      = errors.New("account is locked due to too many failed attempts")
	ErrAccountDisabled    = errors.New("account has been disabled")
)

// KYCStatus 实名认证状态。
type KYCStatus string

const (
	KYCStatusNone     KYCStatus = "NONE"     // 未提交。
	KYCStatusPending  KYCStatus = "PENDING"  // 审核中。
	KYCStatusApproved KYCStatus = "APPROVED" // 已通过。
	KYCStatusRejected KYCStatus = "REJECTED" // 已拒绝。
	KYCStatusExpired  KYCStatus = "EXPIRED"  // 已过期。
)

// KYCLevel 实名认证等级。
type KYCLevel int8

const (
	KYCLevelNone     KYCLevel = 0 // 未认证。
	KYCLevelBasic    KYCLevel = 1 // 基础认证（手机号+姓名）。
	KYCLevelAdvanced KYCLevel = 2 // 高级认证（身份证+人脸识别）。
	KYCLevelPro      KYCLevel = 3 // 专业认证（视频认证+地址证明）。
)

// KYCInfo 实名认证信息值对象。
// 从 kyc 服务合并而来，作为 User 聚合根的子对象。
type KYCInfo struct {
	Status       KYCStatus  `json:"status"`
	Level        KYCLevel   `json:"level"`
	RealName     string     `json:"real_name"`      // 真实姓名。
	IDType       string     `json:"id_type"`         // 证件类型（ID_CARD/PASSPORT/DRIVER_LICENSE）。
	IDNumber     string     `json:"id_number"`       // 证件号码（加密存储）。
	IDFrontImage string     `json:"id_front_image"`  // 证件正面照片 URL。
	IDBackImage  string     `json:"id_back_image"`   // 证件背面照片 URL。
	FaceImage    string     `json:"face_image"`      // 人脸照片 URL。
	SubmittedAt  *time.Time `json:"submitted_at"`    // 提交时间。
	VerifiedAt   *time.Time `json:"verified_at"`     // 审核通过时间。
	RejectReason string     `json:"reject_reason"`   // 拒绝原因。
	ExpiresAt    *time.Time `json:"expires_at"`      // 认证过期时间。
}

// MFAType 多因素认证类型。
type MFAType string

const (
	MFATypeTOTP  MFAType = "TOTP"  // 基于时间的一次性密码。
	MFATypeSMS   MFAType = "SMS"   // 短信验证码。
	MFATypeEmail MFAType = "EMAIL" // 邮箱验证码。
)

// MFAConfig 多因素认证配置。
type MFAConfig struct {
	Enabled   bool    `json:"enabled"`
	Type      MFAType `json:"type"`
	Secret    string  `json:"secret"`     // TOTP 密钥（加密存储）。
	BackupCodes []string `json:"backup_codes"` // 备用恢复码。
	VerifiedAt *time.Time `json:"verified_at"`
}

// LoginAttempt 登录尝试记录。
type LoginAttempt struct {
	UserID    uint64    `json:"user_id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Success   bool      `json:"success"`
	FailReason string   `json:"fail_reason"`
	Timestamp time.Time `json:"timestamp"`
}

// SecurityConfig 安全配置。
type SecurityConfig struct {
	MaxLoginAttempts int32         `json:"max_login_attempts"` // 最大登录失败次数。
	LockDuration     time.Duration `json:"lock_duration"`      // 锁定时长。
	PasswordMinLen   int32         `json:"password_min_len"`   // 密码最小长度。
	RequireUppercase bool          `json:"require_uppercase"`  // 是否要求大写字母。
	RequireNumber    bool          `json:"require_number"`     // 是否要求数字。
	RequireSpecial   bool          `json:"require_special"`    // 是否要求特殊字符。
	SessionTimeout   time.Duration `json:"session_timeout"`    // 会话超时时间。
}

// SubmitKYC 提交实名认证。
func (u *User) SubmitKYC(realName, idType, idNumber, frontImage, backImage, faceImage string) error {
	if u.KYC != nil && u.KYC.Status == KYCStatusApproved {
		return ErrKYCAlreadyVerified
	}
	if u.KYC != nil && u.KYC.Status == KYCStatusPending {
		return ErrKYCPending
	}

	now := time.Now()
	u.KYC = &KYCInfo{
		Status:       KYCStatusPending,
		RealName:     realName,
		IDType:       idType,
		IDNumber:     idNumber,
		IDFrontImage: frontImage,
		IDBackImage:  backImage,
		FaceImage:    faceImage,
		SubmittedAt:  &now,
	}
	return nil
}

// ApproveKYC 审核通过实名认证。
func (u *User) ApproveKYC(level KYCLevel) error {
	if u.KYC == nil || u.KYC.Status != KYCStatusPending {
		return ErrKYCRejected
	}
	now := time.Now()
	u.KYC.Status = KYCStatusApproved
	u.KYC.Level = level
	u.KYC.VerifiedAt = &now
	// 认证有效期 1 年。
	expires := now.AddDate(1, 0, 0)
	u.KYC.ExpiresAt = &expires
	return nil
}

// RejectKYC 拒绝实名认证。
func (u *User) RejectKYC(reason string) error {
	if u.KYC == nil || u.KYC.Status != KYCStatusPending {
		return errors.New("no pending KYC to reject")
	}
	u.KYC.Status = KYCStatusRejected
	u.KYC.RejectReason = reason
	return nil
}

// EnableMFA 启用多因素认证。
func (u *User) EnableMFA(mfaType MFAType, secret string) {
	u.MFA = &MFAConfig{
		Enabled: true,
		Type:    mfaType,
		Secret:  secret,
	}
}

// DisableMFA 禁用多因素认证。
func (u *User) DisableMFA() {
	if u.MFA != nil {
		u.MFA.Enabled = false
	}
}

// IsKYCVerified 判断用户是否已通过实名认证。
func (u *User) IsKYCVerified() bool {
	if u.KYC == nil {
		return false
	}
	if u.KYC.Status != KYCStatusApproved {
		return false
	}
	if u.KYC.ExpiresAt != nil && time.Now().After(*u.KYC.ExpiresAt) {
		return false
	}
	return true
}
