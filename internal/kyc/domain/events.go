// 变更说明：KYC 领域事件定义
package domain

import "time"

const (
	KYCApplicationCreatedEventType   = "kyc.application.created"
	KYCApplicationSubmittedEventType = "kyc.application.submitted"
	KYCApplicationApprovedEventType  = "kyc.application.approved"
	KYCApplicationRejectedEventType  = "kyc.application.rejected"
	KYCApplicationCancelledEventType = "kyc.application.cancelled"
	KYCDocumentUploadedEventType     = "kyc.document.uploaded"
	KYCFaceVerifiedEventType         = "kyc.face.verified"
	KYCRiskScoreUpdatedEventType     = "kyc.risk_score.updated"
	KYCLevelUpgradedEventType        = "kyc.level.upgraded"
	KYCExpiredEventType              = "kyc.expired"
)

// KYCApplicationCreatedEvent KYC申请创建事件
type KYCApplicationCreatedEvent struct {
	ApplicationID string    `json:"application_id"`
	UserID        uint64    `json:"user_id"`
	UserType      string    `json:"user_type"`
	FullName      string    `json:"full_name"`
	IDType        string    `json:"id_type"`
	Level         string    `json:"level"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// KYCApplicationSubmittedEvent KYC申请提交事件
type KYCApplicationSubmittedEvent struct {
	ApplicationID string    `json:"application_id"`
	UserID        uint64    `json:"user_id"`
	Level         string    `json:"level"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// KYCApplicationApprovedEvent KYC申请通过事件
type KYCApplicationApprovedEvent struct {
	ApplicationID string    `json:"application_id"`
	UserID        uint64    `json:"user_id"`
	Level         string    `json:"level"`
	AuditorID     uint64    `json:"auditor_id"`
	ExpiresAt     int64     `json:"expires_at"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// KYCApplicationRejectedEvent KYC申请拒绝事件
type KYCApplicationRejectedEvent struct {
	ApplicationID string    `json:"application_id"`
	UserID        uint64    `json:"user_id"`
	Reason        string    `json:"reason"`
	AuditorID     uint64    `json:"auditor_id"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// KYCApplicationCancelledEvent KYC申请取消事件
type KYCApplicationCancelledEvent struct {
	ApplicationID string    `json:"application_id"`
	UserID        uint64    `json:"user_id"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// KYCDocumentUploadedEvent 证件上传事件
type KYCDocumentUploadedEvent struct {
	ApplicationID string    `json:"application_id"`
	DocumentID    string    `json:"document_id"`
	DocumentType  string    `json:"document_type"`
	DocumentURL   string    `json:"document_url"`
	OCRSuccess    bool      `json:"ocr_success"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// KYCFaceVerifiedEvent 人脸验证事件
type KYCFaceVerifiedEvent struct {
	ApplicationID   string    `json:"application_id"`
	VerificationID  string    `json:"verification_id"`
	Passed          bool      `json:"passed"`
	SimilarityScore float64   `json:"similarity_score"`
	LivenessPassed  bool      `json:"liveness_passed"`
	OccurredAt      time.Time `json:"occurred_at"`
}

// KYCRiskScoreUpdatedEvent 风险评分更新事件
type KYCRiskScoreUpdatedEvent struct {
	ApplicationID string       `json:"application_id"`
	UserID        uint64       `json:"user_id"`
	RiskScore     int          `json:"risk_score"`
	RiskFactors   []RiskFactor `json:"risk_factors"`
	OccurredAt    time.Time    `json:"occurred_at"`
}

// KYCLevelUpgradedEvent KYC等级升级事件
type KYCLevelUpgradedEvent struct {
	ApplicationID string    `json:"application_id"`
	UserID        uint64    `json:"user_id"`
	OldLevel      string    `json:"old_level"`
	NewLevel      string    `json:"new_level"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// KYCExpiredEvent KYC过期事件
type KYCExpiredEvent struct {
	ApplicationID string    `json:"application_id"`
	UserID        uint64    `json:"user_id"`
	ExpiredAt     int64     `json:"expired_at"`
	OccurredAt    time.Time `json:"occurred_at"`
}
