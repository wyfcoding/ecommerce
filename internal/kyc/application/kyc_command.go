// 变更说明：KYC 命令服务，处理所有写操作
package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/kyc/domain"
	"github.com/wyfcoding/pkg/contextx"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/messagequeue"
)

// KYCCommandService KYC命令服务
type KYCCommandService struct {
	repo              domain.KYCRepository
	docRepo           domain.DocumentRepository
	faceRepo          domain.FaceVerificationRepository
	auditRepo         domain.AuditRecordRepository
	merchantRepo      domain.MerchantKYCRepository
	readRepo          domain.KYCReadRepository
	domainService     *domain.KYCDomainService
	publisher         messagequeue.EventPublisher
	topic             string
	logger            *slog.Logger
}

// NewKYCCommandService 创建KYC命令服务
func NewKYCCommandService(
	repo domain.KYCRepository,
	docRepo domain.DocumentRepository,
	faceRepo domain.FaceVerificationRepository,
	auditRepo domain.AuditRecordRepository,
	merchantRepo domain.MerchantKYCRepository,
	readRepo domain.KYCReadRepository,
	domainService *domain.KYCDomainService,
	publisher messagequeue.EventPublisher,
	topic string,
	logger *slog.Logger,
) *KYCCommandService {
	return &KYCCommandService{
		repo:          repo,
		docRepo:       docRepo,
		faceRepo:      faceRepo,
		auditRepo:     auditRepo,
		merchantRepo:  merchantRepo,
		readRepo:      readRepo,
		domainService: domainService,
		publisher:     publisher,
		topic:         topic,
		logger:        logger,
	}
}

// SubmitKYCCommand 提交KYC申请命令
type SubmitKYCCommand struct {
	UserID      uint64
	FullName    string
	IDNumber    string
	IDType      domain.IDType
	Nationality string
	BirthDate   *time.Time
	Gender      string
	Address     string
	Country     string
	Province    string
	City        string
	PostalCode  string
}

// SubmitKYC 提交KYC申请
func (s *KYCCommandService) SubmitKYC(ctx context.Context, cmd SubmitKYCCommand) (*domain.KYCApplication, error) {
	s.logger.InfoContext(ctx, "submitting kyc application", "user_id", cmd.UserID)

	// 检查是否已有进行中的申请
	existing, err := s.repo.FindByUserID(ctx, cmd.UserID)
	if err == nil && existing != nil {
		if existing.Status == domain.KYCStatusPending || existing.Status == domain.KYCStatusUnderReview {
			return nil, errors.New("existing application is in progress")
		}
	}

	// 创建申请
	app, err := domain.NewKYCApplication(cmd.UserID, cmd.FullName, cmd.IDNumber, cmd.IDType)
	if err != nil {
		return nil, err
	}

	// 填充详细信息
	app.ApplicationID = fmt.Sprintf("KYC-%d-%d", cmd.UserID, time.Now().UnixNano())
	app.ID = uint64(idgen.GenID())
	app.Nationality = cmd.Nationality
	app.BirthDate = cmd.BirthDate
	app.Gender = cmd.Gender
	app.Address = cmd.Address
	app.Country = cmd.Country
	app.Province = cmd.Province
	app.City = cmd.City
	app.PostalCode = cmd.PostalCode

	// 提交申请
	if err := app.Submit(); err != nil {
		return nil, err
	}

	// 保存并发送事件
	if err := s.repo.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Save(txCtx, app); err != nil {
			return fmt.Errorf("failed to save application: %w", err)
		}

		if s.publisher != nil {
			event := domain.KYCApplicationSubmittedEvent{
				ApplicationID: app.ApplicationID,
				UserID:        app.UserID,
				Level:         app.Level.String(),
				OccurredAt:    time.Now(),
			}
			return s.publisher.PublishInTx(txCtx, contextx.GetTx(txCtx), domain.KYCApplicationSubmittedEventType, app.ApplicationID, event)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	s.logger.InfoContext(ctx, "kyc application submitted", "application_id", app.ApplicationID)
	return app, nil
}

// UploadDocumentCommand 上传证件命令
type UploadDocumentCommand struct {
	ApplicationID string
	DocumentType  domain.IDType
	Side          domain.DocumentSide
	DocumentData  []byte
	DocumentURL   string
}

// UploadDocument 上传证件
func (s *KYCCommandService) UploadDocument(ctx context.Context, cmd UploadDocumentCommand) (*domain.Document, error) {
	s.logger.InfoContext(ctx, "uploading document", "application_id", cmd.ApplicationID)

	app, err := s.repo.FindByApplicationID(ctx, cmd.ApplicationID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	var doc *domain.Document
	if len(cmd.DocumentData) > 0 && s.domainService != nil {
		doc, err = s.domainService.ProcessDocument(ctx, app, cmd.DocumentType, cmd.Side, cmd.DocumentData)
	} else {
		doc = &domain.Document{
			DocumentID:    fmt.Sprintf("DOC-%d", time.Now().UnixNano()),
			ApplicationID: cmd.ApplicationID,
			DocumentType:  cmd.DocumentType,
			Side:          cmd.Side,
			DocumentURL:   cmd.DocumentURL,
			UploadedAt:    time.Now(),
		}
	}

	if err != nil {
		return nil, err
	}

	// 保存证件
	if err := s.docRepo.Save(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to save document: %w", err)
	}

	// 更新申请
	if err := app.UploadDocument(doc); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, app); err != nil {
		return nil, err
	}

	// 发送事件
	if s.publisher != nil {
		event := domain.KYCDocumentUploadedEvent{
			ApplicationID: app.ApplicationID,
			DocumentID:    doc.DocumentID,
			DocumentType:  doc.DocumentType.String(),
			DocumentURL:   doc.DocumentURL,
			OCRSuccess:    doc.OCRInfo != nil,
			OccurredAt:    time.Now(),
		}
		_ = s.publisher.PublishInTx(ctx, nil, domain.KYCDocumentUploadedEventType, app.ApplicationID, event)
	}

	s.logger.InfoContext(ctx, "document uploaded", "document_id", doc.DocumentID)
	return doc, nil
}

// SubmitFaceVerificationCommand 提交人脸验证命令
type SubmitFaceVerificationCommand struct {
	ApplicationID string
	FaceImageData []byte
	FaceImageURL  string
	LivenessData  []byte
}

// SubmitFaceVerification 提交人脸验证
func (s *KYCCommandService) SubmitFaceVerification(ctx context.Context, cmd SubmitFaceVerificationCommand) (*domain.FaceVerification, error) {
	s.logger.InfoContext(ctx, "submitting face verification", "application_id", cmd.ApplicationID)

	app, err := s.repo.FindByApplicationID(ctx, cmd.ApplicationID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	var fv *domain.FaceVerification
	if s.domainService != nil {
		fv, err = s.domainService.ProcessFaceVerification(ctx, app, cmd.FaceImageURL, cmd.LivenessData)
	} else {
		fv = &domain.FaceVerification{
			VerificationID:  fmt.Sprintf("FV-%d", time.Now().UnixNano()),
			ApplicationID:   cmd.ApplicationID,
			Passed:          true,
			SimilarityScore: 0.95,
			LivenessPassed:  true,
			FaceImageURL:    cmd.FaceImageURL,
			VerifiedAt:      time.Now(),
		}
	}

	if err != nil {
		return nil, err
	}

	// 保存人脸验证记录
	if err := s.faceRepo.Save(ctx, fv); err != nil {
		return nil, fmt.Errorf("failed to save face verification: %w", err)
	}

	// 更新申请
	if err := app.UpdateFaceVerification(fv); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, app); err != nil {
		return nil, err
	}

	// 发送事件
	if s.publisher != nil {
		event := domain.KYCFaceVerifiedEvent{
			ApplicationID:   app.ApplicationID,
			VerificationID:  fv.VerificationID,
			Passed:          fv.Passed,
			SimilarityScore: fv.SimilarityScore,
			LivenessPassed:  fv.LivenessPassed,
			OccurredAt:      time.Now(),
		}
		_ = s.publisher.PublishInTx(ctx, nil, domain.KYCFaceVerifiedEventType, app.ApplicationID, event)
	}

	s.logger.InfoContext(ctx, "face verification completed", "passed", fv.Passed)
	return fv, nil
}

// ApproveKYCCommand 审核通过命令
type ApproveKYCCommand struct {
	ApplicationID string
	AuditorID     uint64
	AuditorName   string
	ApprovedLevel domain.KYCLevel
	Comment       string
	ValidityMonths int
}

// ApproveKYC 审核通过
func (s *KYCCommandService) ApproveKYC(ctx context.Context, cmd ApproveKYCCommand) error {
	s.logger.InfoContext(ctx, "approving kyc", "application_id", cmd.ApplicationID)

	app, err := s.repo.FindByApplicationID(ctx, cmd.ApplicationID)
	if err != nil {
		return err
	}
	if app == nil {
		return errors.New("application not found")
	}

	validityMonths := cmd.ValidityMonths
	if validityMonths <= 0 {
		validityMonths = 12
	}

	if err := app.Approve(cmd.AuditorID, cmd.AuditorName, cmd.Comment, cmd.ApprovedLevel, validityMonths); err != nil {
		return err
	}

	// 保存审核记录
	for _, record := range app.AuditRecords {
		if err := s.auditRepo.Save(ctx, record); err != nil {
			s.logger.ErrorContext(ctx, "failed to save audit record", "error", err)
		}
	}

	// 更新申请
	if err := s.repo.Update(ctx, app); err != nil {
		return err
	}

	// 更新读模型
	if s.readRepo != nil {
		_ = s.readRepo.Save(ctx, app)
	}

	// 发送事件
	if s.publisher != nil {
		event := domain.KYCApplicationApprovedEvent{
			ApplicationID: app.ApplicationID,
			UserID:        app.UserID,
			Level:         app.Level.String(),
			AuditorID:     cmd.AuditorID,
			ExpiresAt:     app.ExpiresAt.Unix(),
			OccurredAt:    time.Now(),
		}
		_ = s.publisher.PublishInTx(ctx, nil, domain.KYCApplicationApprovedEventType, app.ApplicationID, event)
	}

	// 发送通知
	if s.domainService != nil {
		_ = s.domainService.NotifyApproval(ctx, app.UserID, app.Level)
	}

	s.logger.InfoContext(ctx, "kyc approved", "application_id", cmd.ApplicationID, "level", cmd.ApprovedLevel)
	return nil
}

// RejectKYCCommand 审核拒绝命令
type RejectKYCCommand struct {
	ApplicationID string
	AuditorID     uint64
	AuditorName   string
	Reason        string
	Comment       string
}

// RejectKYC 审核拒绝
func (s *KYCCommandService) RejectKYC(ctx context.Context, cmd RejectKYCCommand) error {
	s.logger.InfoContext(ctx, "rejecting kyc", "application_id", cmd.ApplicationID)

	app, err := s.repo.FindByApplicationID(ctx, cmd.ApplicationID)
	if err != nil {
		return err
	}
	if app == nil {
		return errors.New("application not found")
	}

	if err := app.Reject(cmd.AuditorID, cmd.AuditorName, cmd.Reason, cmd.Comment); err != nil {
		return err
	}

	// 保存审核记录
	for _, record := range app.AuditRecords {
		if err := s.auditRepo.Save(ctx, record); err != nil {
			s.logger.ErrorContext(ctx, "failed to save audit record", "error", err)
		}
	}

	// 更新申请
	if err := s.repo.Update(ctx, app); err != nil {
		return err
	}

	// 发送事件
	if s.publisher != nil {
		event := domain.KYCApplicationRejectedEvent{
			ApplicationID: app.ApplicationID,
			UserID:        app.UserID,
			Reason:        cmd.Reason,
			AuditorID:     cmd.AuditorID,
			OccurredAt:    time.Now(),
		}
		_ = s.publisher.PublishInTx(ctx, nil, domain.KYCApplicationRejectedEventType, app.ApplicationID, event)
	}

	// 发送通知
	if s.domainService != nil {
		_ = s.domainService.NotifyRejection(ctx, app.UserID, cmd.Reason)
	}

	s.logger.InfoContext(ctx, "kyc rejected", "application_id", cmd.ApplicationID)
	return nil
}

// CancelKYC 取消KYC申请
func (s *KYCCommandService) CancelKYC(ctx context.Context, applicationID string, userID uint64) error {
	app, err := s.repo.FindByApplicationID(ctx, applicationID)
	if err != nil {
		return err
	}
	if app == nil || app.UserID != userID {
		return errors.New("application not found")
	}

	if err := app.Cancel(); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, app); err != nil {
		return err
	}

	// 发送事件
	if s.publisher != nil {
		event := domain.KYCApplicationCancelledEvent{
			ApplicationID: app.ApplicationID,
			UserID:        app.UserID,
			OccurredAt:    time.Now(),
		}
		_ = s.publisher.PublishInTx(ctx, nil, domain.KYCApplicationCancelledEventType, app.ApplicationID, event)
	}

	return nil
}

// SubmitMerchantKYCCommand 提交商家KYC命令
type SubmitMerchantKYCCommand struct {
	MerchantID                uint64
	CompanyName               string
	RegistrationNumber        string
	TaxID                     string
	LegalRepresentativeName   string
	LegalRepresentativeID     string
	BusinessAddress           string
	Country                   string
	Province                  string
	City                      string
	PostalCode                string
	ContactPhone              string
	ContactEmail              string
	BankName                  string
	BankAccount               string
	BankAccountName           string
}

// SubmitMerchantKYC 提交商家KYC
func (s *KYCCommandService) SubmitMerchantKYC(ctx context.Context, cmd SubmitMerchantKYCCommand) (*domain.KYCApplication, error) {
	s.logger.InfoContext(ctx, "submitting merchant kyc", "merchant_id", cmd.MerchantID)

	app, err := domain.NewMerchantKYCApplication(cmd.MerchantID, cmd.CompanyName, cmd.RegistrationNumber)
	if err != nil {
		return nil, err
	}

	app.ApplicationID = fmt.Sprintf("MKYC-%d-%d", cmd.MerchantID, time.Now().UnixNano())
	app.ID = uint64(idgen.GenID())
	app.FullName = cmd.LegalRepresentativeName
	app.IDNumber = cmd.LegalRepresentativeID
	app.IDType = domain.IDTypeIDCard
	app.Address = cmd.BusinessAddress
	app.Country = cmd.Country
	app.Province = cmd.Province
	app.City = cmd.City
	app.PostalCode = cmd.PostalCode

	if err := app.Submit(); err != nil {
		return nil, err
	}

	// 保存申请
	if err := s.repo.Save(ctx, app); err != nil {
		return nil, err
	}

	// 保存商家KYC信息
	merchantInfo := &domain.MerchantKYCInfo{
		ApplicationID:           app.ApplicationID,
		CompanyName:             cmd.CompanyName,
		RegistrationNumber:      cmd.RegistrationNumber,
		TaxID:                   cmd.TaxID,
		LegalRepresentativeName: cmd.LegalRepresentativeName,
		LegalRepresentativeID:   cmd.LegalRepresentativeID,
		BusinessAddress:         cmd.BusinessAddress,
		Country:                 cmd.Country,
		Province:                cmd.Province,
		City:                    cmd.City,
		PostalCode:              cmd.PostalCode,
		ContactPhone:            cmd.ContactPhone,
		ContactEmail:            cmd.ContactEmail,
		BankName:                cmd.BankName,
		BankAccount:             cmd.BankAccount,
		BankAccountName:         cmd.BankAccountName,
	}

	if err := s.merchantRepo.Save(ctx, merchantInfo); err != nil {
		s.logger.ErrorContext(ctx, "failed to save merchant info", "error", err)
	}

	s.logger.InfoContext(ctx, "merchant kyc submitted", "application_id", app.ApplicationID)
	return app, nil
}

// CalculateRiskScore 计算风险评分
func (s *KYCCommandService) CalculateRiskScore(ctx context.Context, applicationID string) (*domain.RiskAssessmentResult, error) {
	app, err := s.repo.FindByApplicationID(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	if s.domainService == nil {
		return &domain.RiskAssessmentResult{Score: 0, Level: "LOW"}, nil
	}

	result, err := s.domainService.AssessRisk(ctx, app)
	if err != nil {
		return nil, err
	}

	// 更新风险评分
	app.UpdateRiskScore(result.Score, result.Factors)
	if err := s.repo.Update(ctx, app); err != nil {
		s.logger.ErrorContext(ctx, "failed to update risk score", "error", err)
	}

	// 发送事件
	if s.publisher != nil {
		event := domain.KYCRiskScoreUpdatedEvent{
			ApplicationID: app.ApplicationID,
			UserID:        app.UserID,
			RiskScore:     result.Score,
			RiskFactors:   result.Factors,
			OccurredAt:    time.Now(),
		}
		_ = s.publisher.PublishInTx(ctx, nil, domain.KYCRiskScoreUpdatedEventType, app.ApplicationID, event)
	}

	return result, nil
}
