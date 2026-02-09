// 变更说明：新增KYC（实名认证）功能，支撑合规性检查、实名认证流程、证件管理与风险分级。
// 假设：KYC默认有效期为证件有效期。
package domain

import (
	"context"
	"errors"
	"time"
)

// --- KYC 等级 ---

// KYCLevel KYC认证等级
type KYCLevel int

const (
	KYCLevel0 KYCLevel = 0 // 未认证
	KYCLevel1 KYCLevel = 1 // 基础认证（姓名、证件号）
	KYCLevel2 KYCLevel = 2 // 高级认证（人脸识别、证件照片）
	KYCLevel3 KYCLevel = 3 // 视频/线下核验（适用于高额金融业务）
)

// --- KYC 状态 ---

// KYCStatus KYC状态
type KYCStatus int

const (
	KYCStatusNone     KYCStatus = 0 // 未申请
	KYCStatusPending  KYCStatus = 1 // 待提交
	KYCStatusInReview KYCStatus = 2 // 审核中
	KYCStatusApproved KYCStatus = 3 // 已通过
	KYCStatusRejected KYCStatus = 4 // 已拒绝
	KYCStatusExpired  KYCStatus = 5 // 已过期
)

// --- 证件类型 ---

// IDType 证件类型
type IDType string

const (
	IDTypeIDCard   IDType = "IDCARD"   // 身份证
	IDTypePassport IDType = "PASSPORT" // 护照
	IDTypeDriver   IDType = "DRIVER"   // 驾照
)

// --- KYC 聚合根 ---

// KYCRecord KYC记录聚合根
type KYCRecord struct {
	ID             uint64      `json:"id"`
	UserID         uint64      `json:"user_id"`
	Level          KYCLevel    `json:"level"`
	Status         KYCStatus   `json:"status"`
	IDInfo         *IDDocument `json:"id_info"`          // 证件信息
	FaceMatchScore float64     `json:"face_match_score"` // 人脸匹配得分
	RiskScore      int         `json:"risk_score"`       // 风险评估得分
	Notes          string      `json:"notes"`            // 审核备注
	SubmittedAt    *time.Time  `json:"submitted_at"`
	ReviewedAt     *time.Time  `json:"reviewed_at"`
	ReviewedBy     string      `json:"reviewed_by"`
	ExpiredAt      *time.Time  `json:"expired_at"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

// IDDocument 证件文档
type IDDocument struct {
	Type        IDType     `json:"type"`
	Number      string     `json:"number"`       // 证件号
	Name        string     `json:"name"`         // 真实姓名
	FrontImage  string     `json:"front_image"`  // 正面照片
	BackImage   string     `json:"back_image"`   // 反面照片
	SelfieImage string     `json:"selfie_image"` // 手持/自拍照片
	Issuer      string     `json:"issuer"`       // 签发机构
	IssueDate   *time.Time `json:"issue_date"`   // 签发日期
	ExpiryDate  *time.Time `json:"expiry_date"`  // 到期日期
}

// --- 实名认证会话 (KYC Session) ---

// KYCSession KYC认证会话，用于追踪单次认证过程
type KYCSession struct {
	ID          string    `json:"id"`
	UserID      uint64    `json:"user_id"`
	TargetLevel KYCLevel  `json:"target_level"`
	Step        string    `json:"step"` // 当前步骤：DOC_UPLOAD/FACE_VERIFY/MANUAL_REVIEW
	IsCompleted bool      `json:"is_completed"`
	Result      string    `json:"result"` // SUCCESS/FAIL
	StartTime   time.Time `json:"start_time"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// --- 风险概览 (Risk Profile) ---

// RiskProfile 用户风险概览
type RiskProfile struct {
	UserID        uint64    `json:"user_id"`
	RiskLevel     string    `json:"risk_level"` // LOW/MEDIUM/HIGH/CRITICAL
	Score         int       `json:"score"`
	Labels        []string  `json:"labels"` // 风险标签：FRAUD_SUSPECT/PEP(政治敏感)/SANCTIONED
	LastCheckedAt time.Time `json:"last_checked_at"`
}

// NewKYCRecord 创建KYC记录
func NewKYCRecord(userID uint64) *KYCRecord {
	return &KYCRecord{
		UserID: userID,
		Status: KYCStatusPending,
		Level:  KYCLevel0,
	}
}

// Submit 提交申请
func (r *KYCRecord) Submit(level KYCLevel, doc *IDDocument) error {
	if r.Status == KYCStatusInReview {
		return errors.New("kyc is already in review")
	}

	r.Level = level
	r.IDInfo = doc
	r.Status = KYCStatusInReview
	now := time.Now()
	r.SubmittedAt = &now

	// 默认过期日期设为证件到期日
	if doc != nil && doc.ExpiryDate != nil {
		r.ExpiredAt = doc.ExpiryDate
	}

	return nil
}

// Approve 审批通过
func (r *KYCRecord) Approve(reviewer string, notes string) error {
	if r.Status != KYCStatusInReview {
		return errors.New("can only approve records in review status")
	}

	r.Status = KYCStatusApproved
	now := time.Now()
	r.ReviewedAt = &now
	r.ReviewedBy = reviewer
	r.Notes = notes

	return nil
}

// Reject 审批拒绝
func (r *KYCRecord) Reject(reviewer string, reason string) error {
	if r.Status != KYCStatusInReview {
		return errors.New("can only reject records in review status")
	}

	r.Status = KYCStatusRejected
	now := time.Now()
	r.ReviewedAt = &now
	r.ReviewedBy = reviewer
	r.Notes = reason

	return nil
}

// IsExpired 检查是否过期
func (r *KYCRecord) IsExpired() bool {
	if r.Status != KYCStatusApproved {
		return false
	}
	return r.ExpiredAt != nil && time.Now().After(*r.ExpiredAt)
}

// --- KYC 仓储接口 ---

// KYCRepository KYC仓储接口
type KYCRepository interface {
	FindByUserID(ctx context.Context, userID uint64) (*KYCRecord, error)
	Save(ctx context.Context, record *KYCRecord) error

	// KYCSession 处理
	CreateSession(ctx context.Context, session *KYCSession) error
	GetSession(ctx context.Context, sessionID string) (*KYCSession, error)

	// 风险概览
	GetRiskProfile(ctx context.Context, userID uint64) (*RiskProfile, error)
	UpdateRiskProfile(ctx context.Context, profile *RiskProfile) error
}

// --- KYC 服务接口 ---

// KYCService KYC领域服务接口
type KYCService interface {
	// 自动评估风险分值
	EvaluateRisk(ctx context.Context, userID uint64, info *IDDocument) (int, error)
	// 对接人脸识别插件
	VerifyFace(ctx context.Context, userID uint64, selfieImg string, idCardImg string) (float64, error)
}
