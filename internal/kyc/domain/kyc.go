// 变更说明：KYC 领域模型定义，包含聚合根、实体、值对象和领域服务
package domain

import (
	"fmt"
	"time"
)

// KYCApplication KYC申请聚合根
// 包含用户实名认证的完整信息，包括个人信息、证件、人脸验证、审核记录等
type KYCApplication struct {
	ID              uint64             `json:"id"`
	ApplicationID   string             `json:"application_id"`
	UserID          uint64             `json:"user_id"`
	UserType        UserType           `json:"user_type"`          // 用户类型：个人/商家
	MerchantID      uint64             `json:"merchant_id"`        // 商家ID（商家认证时使用）
	FullName        string             `json:"full_name"`          // 真实姓名
	IDNumber        string             `json:"id_number"`          // 证件号码
	IDType          IDType             `json:"id_type"`            // 证件类型
	Nationality     string             `json:"nationality"`        // 国籍
	BirthDate       *time.Time         `json:"birth_date"`         // 出生日期
	Gender          string             `json:"gender"`             // 性别 M/F
	Address         string             `json:"address"`            // 详细地址
	Country         string             `json:"country"`            // 国家代码
	Province        string             `json:"province"`           // 省份
	City            string             `json:"city"`               // 城市
	PostalCode      string             `json:"postal_code"`        // 邮编
	Status          KYCStatus          `json:"status"`             // 状态
	Level           KYCLevel           `json:"level"`              // 认证等级
	RiskScore       int                `json:"risk_score"`         // 风险评分 0-100
	Documents       []*Document        `json:"documents"`          // 证件文档列表
	FaceVerification *FaceVerification `json:"face_verification"`  // 人脸验证
	AuditRecords    []*AuditRecord     `json:"audit_records"`      // 审核记录
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	VerifiedAt      *time.Time         `json:"verified_at"`        // 认证通过时间
	ExpiresAt       *time.Time         `json:"expires_at"`         // 过期时间
	Version         int                `json:"version"`            // 版本号（乐观锁）
}

// UserType 用户类型
type UserType int

const (
	UserTypeIndividual UserType = iota // 个人用户
	UserTypeMerchant                   // 商家用户
)

// IDType 证件类型
type IDType int

const (
	IDTypeUnspecified IDType = iota
	IDTypeIDCard         // 身份证
	IDTypePassport       // 护照
	IDTypeDriversLicense // 驾照
	IDTypeResidencePermit // 居住证
	IDTypeHKMacauPass    // 港澳通行证
	IDTypeTaiwanPass     // 台湾通行证
	IDTypeBusinessLicense // 营业执照
)

// KYCStatus KYC状态
type KYCStatus int

const (
	KYCStatusUnspecified KYCStatus = iota
	KYCStatusDraft        // 草稿
	KYCStatusPending      // 待审核
	KYCStatusUnderReview  // 审核中
	KYCStatusApproved     // 已通过
	KYCStatusRejected     // 已拒绝
	KYCStatusExpired      // 已过期
	KYCStatusCancelled    // 已取消
	KYCStatusSuspended    // 已暂停
)

// KYCLevel KYC等级
type KYCLevel int

const (
	KYCLevelUnspecified KYCLevel = iota
	KYCLevelNone        // 未认证
	KYCLevelBasic       // 基础认证（手机+姓名）
	KYCLevelIntermediate // 中级认证（证件+人脸）
	KYCLevelAdvanced    // 高级认证（地址证明+收入证明）
	KYCLevelBusiness    // 企业认证
)

// Document 证件文档实体
type Document struct {
	ID           uint64          `json:"id"`
	DocumentID   string          `json:"document_id"`
	ApplicationID string         `json:"application_id"`
	DocumentType IDType          `json:"document_type"`
	Side         DocumentSide    `json:"side"`          // 正反面
	DocumentURL  string          `json:"document_url"`
	OCRInfo      *IDDocumentInfo `json:"ocr_info"`      // OCR识别信息
	Verified     bool            `json:"verified"`
	UploadedAt   time.Time       `json:"uploaded_at"`
}

// DocumentSide 证件正反面
type DocumentSide int

const (
	DocumentSideUnspecified DocumentSide = iota
	DocumentSideFront    // 正面
	DocumentSideBack     // 反面
	DocumentSideSelfie   // 自拍
)

// IDDocumentInfo OCR识别的证件信息
type IDDocumentInfo struct {
	Name             string     `json:"name"`              // 姓名
	IDNumber         string     `json:"id_number"`         // 证件号
	Gender           string     `json:"gender"`            // 性别
	Nationality      string     `json:"nationality"`       // 民族/国籍
	BirthDate        *time.Time `json:"birth_date"`        // 出生日期
	Address          string     `json:"address"`           // 地址
	IssuingAuthority string     `json:"issuing_authority"` // 签发机关
	IssueDate        *time.Time `json:"issue_date"`        // 签发日期
	ExpiryDate       *time.Time `json:"expiry_date"`       // 有效期
	ConfidenceScore  float64    `json:"confidence_score"`  // 置信度
}

// FaceVerification 人脸验证实体
type FaceVerification struct {
	ID              uint64    `json:"id"`
	VerificationID  string    `json:"verification_id"`
	ApplicationID   string    `json:"application_id"`
	Passed          bool      `json:"passed"`
	SimilarityScore float64   `json:"similarity_score"` // 相似度分数
	LivenessPassed  bool      `json:"liveness_passed"`  // 活体检测是否通过
	FaceImageURL    string    `json:"face_image_url"`
	VerifiedAt      time.Time `json:"verified_at"`
}

// AuditRecord 审核记录实体
type AuditRecord struct {
	ID          uint64      `json:"id"`
	RecordID    string      `json:"record_id"`
	ApplicationID string    `json:"application_id"`
	AuditorID   uint64      `json:"auditor_id"`
	AuditorName string      `json:"auditor_name"`
	Action      AuditAction `json:"action"`
	Reason      string      `json:"reason"`
	Comment     string      `json:"comment"`
	CreatedAt   time.Time   `json:"created_at"`
}

// AuditAction 审核动作
type AuditAction int

const (
	AuditActionUnspecified AuditAction = iota
	AuditActionApprove      // 通过
	AuditActionReject       // 拒绝
	AuditActionRequestInfo  // 要求补充信息
	AuditActionEscalate     // 上报
)

// MerchantKYCInfo 商家KYC信息
type MerchantKYCInfo struct {
	ID                      uint64   `json:"id"`
	ApplicationID           string   `json:"application_id"`
	CompanyName             string   `json:"company_name"`              // 公司名称
	RegistrationNumber      string   `json:"registration_number"`       // 注册号
	TaxID                   string   `json:"tax_id"`                    // 税号
	LegalRepresentativeName string   `json:"legal_representative_name"` // 法人姓名
	LegalRepresentativeID   string   `json:"legal_representative_id"`   // 法人身份证
	BusinessAddress         string   `json:"business_address"`          // 经营地址
	Country                 string   `json:"country"`
	Province                string   `json:"province"`
	City                    string   `json:"city"`
	PostalCode              string   `json:"postal_code"`
	ContactPhone            string   `json:"contact_phone"`
	ContactEmail            string   `json:"contact_email"`
	BankName                string   `json:"bank_name"`                 // 开户银行
	BankAccount             string   `json:"bank_account"`              // 银行账号
	BankAccountName         string   `json:"bank_account_name"`         // 账户名
	BusinessDocuments       []*Document `json:"business_documents"`
}

// KYCLimit KYC限额值对象
type KYCLimit struct {
	LimitType  string `json:"limit_type"`  // daily_withdraw, monthly_transaction
	LimitValue string `json:"limit_value"` // 限额值
	UsedValue  string `json:"used_value"`  // 已用值
	Currency   string `json:"currency"`    // 币种
}

// RiskFactor 风险因素值对象
type RiskFactor struct {
	FactorType  string  `json:"factor_type"`  // ID_AGE, ADDRESS_MATCH, FACE_SIMILARITY
	Description string  `json:"description"`
	Score       int     `json:"score"`
	Weight      float64 `json:"weight"`
}

// NewKYCApplication 创建新的KYC申请（工厂方法）
func NewKYCApplication(userID uint64, fullName, idNumber string, idType IDType) (*KYCApplication, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is required")
	}
	if fullName == "" {
		return nil, fmt.Errorf("full name is required")
	}
	if idNumber == "" {
		return nil, fmt.Errorf("id number is required")
	}

	now := time.Now()
	return &KYCApplication{
		UserID:        userID,
		UserType:      UserTypeIndividual,
		FullName:      fullName,
		IDNumber:      idNumber,
		IDType:        idType,
		Status:        KYCStatusDraft,
		Level:         KYCLevelNone,
		Documents:     []*Document{},
		AuditRecords:  []*AuditRecord{},
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       1,
	}, nil
}

// NewMerchantKYCApplication 创建商家KYC申请
func NewMerchantKYCApplication(merchantID uint64, companyName, registrationNumber string) (*KYCApplication, error) {
	if merchantID == 0 {
		return nil, fmt.Errorf("merchant id is required")
	}
	if companyName == "" {
		return nil, fmt.Errorf("company name is required")
	}
	if registrationNumber == "" {
		return nil, fmt.Errorf("registration number is required")
	}

	now := time.Now()
	return &KYCApplication{
		UserID:        merchantID,
		MerchantID:    merchantID,
		UserType:      UserTypeMerchant,
		Status:        KYCStatusDraft,
		Level:         KYCLevelNone,
		Documents:     []*Document{},
		AuditRecords:  []*AuditRecord{},
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       1,
	}, nil
}

// Submit 提交申请
func (a *KYCApplication) Submit() error {
	if a.Status != KYCStatusDraft && a.Status != KYCStatusRejected {
		return fmt.Errorf("cannot submit application in current status: %s", a.Status.String())
	}

	// 验证必填字段
	if err := a.validateBasicInfo(); err != nil {
		return err
	}

	a.Status = KYCStatusPending
	a.UpdatedAt = time.Now()
	return nil
}

// UploadDocument 上传证件
func (a *KYCApplication) UploadDocument(doc *Document) error {
	if a.Status == KYCStatusApproved || a.Status == KYCStatusCancelled {
		return fmt.Errorf("cannot upload document in current status")
	}

	doc.ApplicationID = a.ApplicationID
	doc.UploadedAt = time.Now()
	a.Documents = append(a.Documents, doc)
	a.UpdatedAt = time.Now()
	return nil
}

// UpdateFaceVerification 更新人脸验证结果
func (a *KYCApplication) UpdateFaceVerification(fv *FaceVerification) error {
	if a.Status == KYCStatusApproved || a.Status == KYCStatusCancelled {
		return fmt.Errorf("cannot update face verification in current status")
	}

	fv.ApplicationID = a.ApplicationID
	fv.VerifiedAt = time.Now()
	a.FaceVerification = fv
	a.UpdatedAt = time.Now()
	return nil
}

// Approve 审核通过
func (a *KYCApplication) Approve(auditorID uint64, auditorName, comment string, level KYCLevel, validityMonths int) error {
	if a.Status != KYCStatusPending && a.Status != KYCStatusUnderReview {
		return fmt.Errorf("cannot approve application in current status")
	}

	now := time.Now()
	expiryDate := now.AddDate(0, validityMonths, 0)

	a.Status = KYCStatusApproved
	a.Level = level
	a.VerifiedAt = &now
	a.ExpiresAt = &expiryDate
	a.UpdatedAt = now

	// 添加审核记录
	record := &AuditRecord{
		RecordID:    generateRecordID(),
		ApplicationID: a.ApplicationID,
		AuditorID:   auditorID,
		AuditorName: auditorName,
		Action:      AuditActionApprove,
		Comment:     comment,
		CreatedAt:   now,
	}
	a.AuditRecords = append(a.AuditRecords, record)

	return nil
}

// Reject 审核拒绝
func (a *KYCApplication) Reject(auditorID uint64, auditorName, reason, comment string) error {
	if a.Status != KYCStatusPending && a.Status != KYCStatusUnderReview {
		return fmt.Errorf("cannot reject application in current status")
	}

	now := time.Now()
	a.Status = KYCStatusRejected
	a.UpdatedAt = now

	// 添加审核记录
	record := &AuditRecord{
		RecordID:    generateRecordID(),
		ApplicationID: a.ApplicationID,
		AuditorID:   auditorID,
		AuditorName: auditorName,
		Action:      AuditActionReject,
		Reason:      reason,
		Comment:     comment,
		CreatedAt:   now,
	}
	a.AuditRecords = append(a.AuditRecords, record)

	return nil
}

// Cancel 取消申请
func (a *KYCApplication) Cancel() error {
	if a.Status == KYCStatusApproved {
		return fmt.Errorf("cannot cancel approved application")
	}

	a.Status = KYCStatusCancelled
	a.UpdatedAt = time.Now()
	return nil
}

// StartReview 开始审核
func (a *KYCApplication) StartReview(auditorID uint64, auditorName string) error {
	if a.Status != KYCStatusPending {
		return fmt.Errorf("can only start review for pending applications")
	}

	a.Status = KYCStatusUnderReview
	a.UpdatedAt = time.Now()

	// 添加审核记录
	record := &AuditRecord{
		RecordID:    generateRecordID(),
		ApplicationID: a.ApplicationID,
		AuditorID:   auditorID,
		AuditorName: auditorName,
		Action:      AuditActionRequestInfo,
		Comment:     "开始审核",
		CreatedAt:   time.Now(),
	}
	a.AuditRecords = append(a.AuditRecords, record)

	return nil
}

// UpdateRiskScore 更新风险评分
func (a *KYCApplication) UpdateRiskScore(score int, factors []RiskFactor) {
	a.RiskScore = score
	a.UpdatedAt = time.Now()
}

// IsExpired 检查是否过期
func (a *KYCApplication) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

// CanResubmit 检查是否可以重新提交
func (a *KYCApplication) CanResubmit() bool {
	return a.Status == KYCStatusRejected
}

// GetRequiredDocuments 获取所需证件列表
func (a *KYCApplication) GetRequiredDocuments() []IDType {
	switch a.Level {
	case KYCLevelBasic:
		return []IDType{}
	case KYCLevelIntermediate:
		return []IDType{a.IDType}
	case KYCLevelAdvanced:
		return []IDType{a.IDType, IDTypeResidencePermit}
	case KYCLevelBusiness:
		return []IDType{IDTypeBusinessLicense, IDTypeIDCard}
	default:
		return []IDType{a.IDType}
	}
}

// validateBasicInfo 验证基本信息
func (a *KYCApplication) validateBasicInfo() error {
	if a.FullName == "" {
		return fmt.Errorf("full name is required")
	}
	if a.IDNumber == "" {
		return fmt.Errorf("id number is required")
	}
	if a.IDType == IDTypeUnspecified {
		return fmt.Errorf("id type is required")
	}
	return nil
}

// String 状态字符串表示
func (s KYCStatus) String() string {
	switch s {
	case KYCStatusDraft:
		return "DRAFT"
	case KYCStatusPending:
		return "PENDING"
	case KYCStatusUnderReview:
		return "UNDER_REVIEW"
	case KYCStatusApproved:
		return "APPROVED"
	case KYCStatusRejected:
		return "REJECTED"
	case KYCStatusExpired:
		return "EXPIRED"
	case KYCStatusCancelled:
		return "CANCELLED"
	case KYCStatusSuspended:
		return "SUSPENDED"
	default:
		return "UNSPECIFIED"
	}
}

// String 等级字符串表示
func (l KYCLevel) String() string {
	switch l {
	case KYCLevelNone:
		return "NONE"
	case KYCLevelBasic:
		return "BASIC"
	case KYCLevelIntermediate:
		return "INTERMEDIATE"
	case KYCLevelAdvanced:
		return "ADVANCED"
	case KYCLevelBusiness:
		return "BUSINESS"
	default:
		return "UNSPECIFIED"
	}
}

// String 证件类型字符串表示
func (t IDType) String() string {
	switch t {
	case IDTypeIDCard:
		return "ID_CARD"
	case IDTypePassport:
		return "PASSPORT"
	case IDTypeDriversLicense:
		return "DRIVERS_LICENSE"
	case IDTypeResidencePermit:
		return "RESIDENCE_PERMIT"
	case IDTypeHKMacauPass:
		return "HK_MACAU_PASS"
	case IDTypeTaiwanPass:
		return "TAIWAN_PASS"
	case IDTypeBusinessLicense:
		return "BUSINESS_LICENSE"
	default:
		return "UNSPECIFIED"
	}
}

// generateRecordID 生成审核记录ID
func generateRecordID() string {
	return fmt.Sprintf("AUD-%d", time.Now().UnixNano())
}

// String 用户类型字符串表示
func (t UserType) String() string {
	switch t {
	case UserTypeIndividual:
		return "INDIVIDUAL"
	case UserTypeMerchant:
		return "MERCHANT"
	default:
		return "UNSPECIFIED"
	}
}

// String 证件正反面字符串表示
func (s DocumentSide) String() string {
	switch s {
	case DocumentSideFront:
		return "FRONT"
	case DocumentSideBack:
		return "BACK"
	case DocumentSideSelfie:
		return "SELFIE"
	default:
		return "UNSPECIFIED"
	}
}

// String 审核动作字符串表示
func (a AuditAction) String() string {
	switch a {
	case AuditActionApprove:
		return "APPROVE"
	case AuditActionReject:
		return "REJECT"
	case AuditActionRequestInfo:
		return "REQUEST_INFO"
	case AuditActionEscalate:
		return "ESCALATE"
	default:
		return "UNSPECIFIED"
	}
}
