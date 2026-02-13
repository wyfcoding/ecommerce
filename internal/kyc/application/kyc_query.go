// 变更说明：KYC 查询服务，处理所有读操作
package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/kyc/domain"
)

// KYCQueryService KYC查询服务
type KYCQueryService struct {
	repo         domain.KYCRepository
	docRepo      domain.DocumentRepository
	faceRepo     domain.FaceVerificationRepository
	auditRepo    domain.AuditRecordRepository
	merchantRepo domain.MerchantKYCRepository
	readRepo     domain.KYCReadRepository
	logger       *slog.Logger
}

// NewKYCQueryService 创建KYC查询服务
func NewKYCQueryService(
	repo domain.KYCRepository,
	docRepo domain.DocumentRepository,
	faceRepo domain.FaceVerificationRepository,
	auditRepo domain.AuditRecordRepository,
	merchantRepo domain.MerchantKYCRepository,
	readRepo domain.KYCReadRepository,
	logger *slog.Logger,
) *KYCQueryService {
	return &KYCQueryService{
		repo:         repo,
		docRepo:      docRepo,
		faceRepo:     faceRepo,
		auditRepo:    auditRepo,
		merchantRepo: merchantRepo,
		readRepo:     readRepo,
		logger:       logger,
	}
}

// KYCApplicationDTO KYC申请DTO
type KYCApplicationDTO struct {
	ID               uint64               `json:"id"`
	ApplicationID    string               `json:"application_id"`
	UserID           uint64               `json:"user_id"`
	UserType         string               `json:"user_type"`
	MerchantID       uint64               `json:"merchant_id"`
	FullName         string               `json:"full_name"`
	IDNumber         string               `json:"id_number"`
	IDType           string               `json:"id_type"`
	Nationality      string               `json:"nationality"`
	BirthDate        *time.Time           `json:"birth_date"`
	Gender           string               `json:"gender"`
	Address          string               `json:"address"`
	Country          string               `json:"country"`
	Province         string               `json:"province"`
	City             string               `json:"city"`
	PostalCode       string               `json:"postal_code"`
	Status           string               `json:"status"`
	Level            string               `json:"level"`
	RiskScore        int                  `json:"risk_score"`
	Documents        []*DocumentDTO       `json:"documents"`
	FaceVerification *FaceVerificationDTO `json:"face_verification"`
	AuditRecords     []*AuditRecordDTO    `json:"audit_records"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
	VerifiedAt       *time.Time           `json:"verified_at"`
	ExpiresAt        *time.Time           `json:"expires_at"`
	IsExpired        bool                 `json:"is_expired"`
}

// DocumentDTO 证件DTO
type DocumentDTO struct {
	DocumentID   string             `json:"document_id"`
	DocumentType string             `json:"document_type"`
	Side         string             `json:"side"`
	DocumentURL  string             `json:"document_url"`
	OCRInfo      *IDDocumentInfoDTO `json:"ocr_info"`
	Verified     bool               `json:"verified"`
	UploadedAt   time.Time          `json:"uploaded_at"`
}

// IDDocumentInfoDTO OCR信息DTO
type IDDocumentInfoDTO struct {
	Name             string     `json:"name"`
	IDNumber         string     `json:"id_number"`
	Gender           string     `json:"gender"`
	Nationality      string     `json:"nationality"`
	BirthDate        *time.Time `json:"birth_date"`
	Address          string     `json:"address"`
	IssuingAuthority string     `json:"issuing_authority"`
	IssueDate        *time.Time `json:"issue_date"`
	ExpiryDate       *time.Time `json:"expiry_date"`
	ConfidenceScore  float64    `json:"confidence_score"`
}

// FaceVerificationDTO 人脸验证DTO
type FaceVerificationDTO struct {
	VerificationID  string    `json:"verification_id"`
	Passed          bool      `json:"passed"`
	SimilarityScore float64   `json:"similarity_score"`
	LivenessPassed  bool      `json:"liveness_passed"`
	FaceImageURL    string    `json:"face_image_url"`
	VerifiedAt      time.Time `json:"verified_at"`
}

// AuditRecordDTO 审核记录DTO
type AuditRecordDTO struct {
	RecordID    string    `json:"record_id"`
	AuditorID   uint64    `json:"auditor_id"`
	AuditorName string    `json:"auditor_name"`
	Action      string    `json:"action"`
	Reason      string    `json:"reason"`
	Comment     string    `json:"comment"`
	CreatedAt   time.Time `json:"created_at"`
}

// KYCStatusDTO KYC状态DTO
type KYCStatusDTO struct {
	UserID           uint64     `json:"user_id"`
	Status           string     `json:"status"`
	Level            string     `json:"level"`
	Reason           string     `json:"reason"`
	VerifiedAt       *time.Time `json:"verified_at"`
	ExpiresAt        *time.Time `json:"expires_at"`
	IsExpired        bool       `json:"is_expired"`
	MissingDocuments []string   `json:"missing_documents"`
}

// MerchantKYCDTO 商家KYC DTO
type MerchantKYCDTO struct {
	ApplicationID           string         `json:"application_id"`
	CompanyName             string         `json:"company_name"`
	RegistrationNumber      string         `json:"registration_number"`
	TaxID                   string         `json:"tax_id"`
	LegalRepresentativeName string         `json:"legal_representative_name"`
	LegalRepresentativeID   string         `json:"legal_representative_id"`
	BusinessAddress         string         `json:"business_address"`
	Country                 string         `json:"country"`
	Province                string         `json:"province"`
	City                    string         `json:"city"`
	PostalCode              string         `json:"postal_code"`
	ContactPhone            string         `json:"contact_phone"`
	ContactEmail            string         `json:"contact_email"`
	BankName                string         `json:"bank_name"`
	BankAccount             string         `json:"bank_account"`
	BankAccountName         string         `json:"bank_account_name"`
	BusinessDocuments       []*DocumentDTO `json:"business_documents"`
}

// PendingApplicationDTO 待审核申请DTO
type PendingApplicationDTO struct {
	ApplicationID string    `json:"application_id"`
	UserID        uint64    `json:"user_id"`
	FullName      string    `json:"full_name"`
	IDType        string    `json:"id_type"`
	Level         string    `json:"level"`
	RiskScore     int       `json:"risk_score"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// GetKYCApplication 获取KYC申请详情
func (s *KYCQueryService) GetKYCApplication(ctx context.Context, applicationID string) (*KYCApplicationDTO, error) {
	app, err := s.repo.FindByApplicationID(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	return s.toDTO(ctx, app), nil
}

// GetKYCApplicationByUserID 根据用户ID获取KYC申请
func (s *KYCQueryService) GetKYCApplicationByUserID(ctx context.Context, userID uint64) (*KYCApplicationDTO, error) {
	// 先从缓存读取
	if s.readRepo != nil {
		app, err := s.readRepo.FindByUserID(ctx, userID)
		if err == nil && app != nil {
			return s.toDTO(ctx, app), nil
		}
	}

	app, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	return s.toDTO(ctx, app), nil
}

// GetKYCStatus 获取KYC状态
func (s *KYCQueryService) GetKYCStatus(ctx context.Context, userID uint64) (*KYCStatusDTO, error) {
	app, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return &KYCStatusDTO{
			UserID: userID,
			Status: domain.KYCStatusUnspecified.String(),
			Level:  domain.KYCLevelNone.String(),
		}, nil
	}

	dto := &KYCStatusDTO{
		UserID:     userID,
		Status:     app.Status.String(),
		Level:      app.Level.String(),
		VerifiedAt: app.VerifiedAt,
		ExpiresAt:  app.ExpiresAt,
		IsExpired:  app.IsExpired(),
	}

	// 获取最新审核记录
	if len(app.AuditRecords) > 0 {
		latest := app.AuditRecords[len(app.AuditRecords)-1]
		dto.Reason = latest.Reason
	}

	// 获取缺失的证件
	dto.MissingDocuments = s.getMissingDocuments(app)

	return dto, nil
}

// ListPendingApplications 获取待审核列表
func (s *KYCQueryService) ListPendingApplications(ctx context.Context, page, pageSize int, levelFilter domain.KYCLevel, countryFilter string) ([]*PendingApplicationDTO, int64, error) {
	apps, total, err := s.repo.FindPendingApplications(ctx, page, pageSize, levelFilter, countryFilter)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*PendingApplicationDTO, len(apps))
	for i, app := range apps {
		dtos[i] = &PendingApplicationDTO{
			ApplicationID: app.ApplicationID,
			UserID:        app.UserID,
			FullName:      app.FullName,
			IDType:        app.IDType.String(),
			Level:         app.Level.String(),
			RiskScore:     app.RiskScore,
			Status:        app.Status.String(),
			CreatedAt:     app.CreatedAt,
		}
	}

	return dtos, total, nil
}

// GetAuditHistory 获取审核历史
func (s *KYCQueryService) GetAuditHistory(ctx context.Context, applicationID string, page, pageSize int) ([]*AuditRecordDTO, int64, error) {
	records, total, err := s.auditRepo.FindByApplicationID(ctx, applicationID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*AuditRecordDTO, len(records))
	for i, r := range records {
		dtos[i] = &AuditRecordDTO{
			RecordID:    r.RecordID,
			AuditorID:   r.AuditorID,
			AuditorName: r.AuditorName,
			Action:      r.Action.String(),
			Reason:      r.Reason,
			Comment:     r.Comment,
			CreatedAt:   r.CreatedAt,
		}
	}

	return dtos, total, nil
}

// GetMerchantKYC 获取商家KYC信息
func (s *KYCQueryService) GetMerchantKYC(ctx context.Context, merchantID uint64) (*MerchantKYCDTO, error) {
	info, err := s.merchantRepo.FindByMerchantID(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, errors.New("merchant kyc not found")
	}

	dto := &MerchantKYCDTO{
		ApplicationID:           info.ApplicationID,
		CompanyName:             info.CompanyName,
		RegistrationNumber:      info.RegistrationNumber,
		TaxID:                   info.TaxID,
		LegalRepresentativeName: info.LegalRepresentativeName,
		LegalRepresentativeID:   info.LegalRepresentativeID,
		BusinessAddress:         info.BusinessAddress,
		Country:                 info.Country,
		Province:                info.Province,
		City:                    info.City,
		PostalCode:              info.PostalCode,
		ContactPhone:            info.ContactPhone,
		ContactEmail:            info.ContactEmail,
		BankName:                info.BankName,
		BankAccount:             info.BankAccount,
		BankAccountName:         info.BankAccountName,
	}

	for _, doc := range info.BusinessDocuments {
		dto.BusinessDocuments = append(dto.BusinessDocuments, toDocumentDTO(doc))
	}

	return dto, nil
}

// toDTO 转换为DTO
func (s *KYCQueryService) toDTO(ctx context.Context, app *domain.KYCApplication) *KYCApplicationDTO {
	dto := &KYCApplicationDTO{
		ID:            app.ID,
		ApplicationID: app.ApplicationID,
		UserID:        app.UserID,
		UserType:      app.UserType.String(),
		MerchantID:    app.MerchantID,
		FullName:      app.FullName,
		IDNumber:      maskIDNumber(app.IDNumber),
		IDType:        app.IDType.String(),
		Nationality:   app.Nationality,
		BirthDate:     app.BirthDate,
		Gender:        app.Gender,
		Address:       app.Address,
		Country:       app.Country,
		Province:      app.Province,
		City:          app.City,
		PostalCode:    app.PostalCode,
		Status:        app.Status.String(),
		Level:         app.Level.String(),
		RiskScore:     app.RiskScore,
		CreatedAt:     app.CreatedAt,
		UpdatedAt:     app.UpdatedAt,
		VerifiedAt:    app.VerifiedAt,
		ExpiresAt:     app.ExpiresAt,
		IsExpired:     app.IsExpired(),
	}

	// 转换证件
	for _, doc := range app.Documents {
		dto.Documents = append(dto.Documents, toDocumentDTO(doc))
	}

	// 转换人脸验证
	if app.FaceVerification != nil {
		dto.FaceVerification = &FaceVerificationDTO{
			VerificationID:  app.FaceVerification.VerificationID,
			Passed:          app.FaceVerification.Passed,
			SimilarityScore: app.FaceVerification.SimilarityScore,
			LivenessPassed:  app.FaceVerification.LivenessPassed,
			FaceImageURL:    app.FaceVerification.FaceImageURL,
			VerifiedAt:      app.FaceVerification.VerifiedAt,
		}
	}

	// 转换审核记录
	for _, r := range app.AuditRecords {
		dto.AuditRecords = append(dto.AuditRecords, &AuditRecordDTO{
			RecordID:    r.RecordID,
			AuditorID:   r.AuditorID,
			AuditorName: r.AuditorName,
			Action:      r.Action.String(),
			Reason:      r.Reason,
			Comment:     r.Comment,
			CreatedAt:   r.CreatedAt,
		})
	}

	return dto
}

// toDocumentDTO 转换证件DTO
func toDocumentDTO(doc *domain.Document) *DocumentDTO {
	dto := &DocumentDTO{
		DocumentID:   doc.DocumentID,
		DocumentType: doc.DocumentType.String(),
		Side:         doc.Side.String(),
		DocumentURL:  doc.DocumentURL,
		Verified:     doc.Verified,
		UploadedAt:   doc.UploadedAt,
	}

	if doc.OCRInfo != nil {
		dto.OCRInfo = &IDDocumentInfoDTO{
			Name:             doc.OCRInfo.Name,
			IDNumber:         maskIDNumber(doc.OCRInfo.IDNumber),
			Gender:           doc.OCRInfo.Gender,
			Nationality:      doc.OCRInfo.Nationality,
			BirthDate:        doc.OCRInfo.BirthDate,
			Address:          doc.OCRInfo.Address,
			IssuingAuthority: doc.OCRInfo.IssuingAuthority,
			IssueDate:        doc.OCRInfo.IssueDate,
			ExpiryDate:       doc.OCRInfo.ExpiryDate,
			ConfidenceScore:  doc.OCRInfo.ConfidenceScore,
		}
	}

	return dto
}

// getMissingDocuments 获取缺失的证件
func (s *KYCQueryService) getMissingDocuments(app *domain.KYCApplication) []string {
	required := app.GetRequiredDocuments()
	existing := make(map[domain.IDType]bool)
	for _, doc := range app.Documents {
		existing[doc.DocumentType] = true
	}

	var missing []string
	for _, t := range required {
		if !existing[t] {
			missing = append(missing, t.String())
		}
	}

	// 检查人脸验证
	if app.FaceVerification == nil {
		missing = append(missing, "FACE_VERIFICATION")
	}

	return missing
}

// maskIDNumber 遮蔽证件号
func maskIDNumber(idNumber string) string {
	if len(idNumber) <= 6 {
		return idNumber
	}
	return idNumber[:3] + "****" + idNumber[len(idNumber)-3:]
}
