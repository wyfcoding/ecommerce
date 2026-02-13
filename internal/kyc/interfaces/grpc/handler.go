// 变更说明：KYC gRPC 接口处理器
package grpc

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/go-api/kyc/v1"
	"github.com/wyfcoding/ecommerce/internal/kyc/application"
	"github.com/wyfcoding/ecommerce/internal/kyc/domain"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// KYCHandler KYC gRPC处理器
type KYCHandler struct {
	pb.UnimplementedKYCServiceServer
	cmd   *application.KYCCommandService
	query *application.KYCQueryService
}

// NewKYCHandler 创建KYC处理器
func NewKYCHandler(cmd *application.KYCCommandService, query *application.KYCQueryService) *KYCHandler {
	return &KYCHandler{
		cmd:   cmd,
		query: query,
	}
}

// SubmitKYC 提交KYC申请
func (h *KYCHandler) SubmitKYC(ctx context.Context, req *pb.SubmitKYCRequest) (*pb.SubmitKYCResponse, error) {
	cmd := application.SubmitKYCCommand{
		UserID:      req.UserId,
		FullName:    req.FullName,
		IDNumber:    req.IdNumber,
		IDType:      pbIDTypeToDomain(req.IdType),
		Nationality: req.Nationality,
		Gender:      req.Gender,
		Address:     req.Address,
		Country:     req.Country,
		Province:    req.Province,
		City:        req.City,
		PostalCode:  req.PostalCode,
	}

	if req.BirthDate != nil {
		birthDate := req.BirthDate.AsTime()
		cmd.BirthDate = &birthDate
	}

	app, err := h.cmd.SubmitKYC(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return &pb.SubmitKYCResponse{
		ApplicationId:    app.ApplicationID,
		Status:           pbStatusFromDomain(app.Status),
		CurrentLevel:     pbLevelFromDomain(app.Level),
		RequiredDocuments: getRequiredDocumentTypes(app),
		Message:          "KYC application submitted successfully",
	}, nil
}

// UploadDocument 上传证件
func (h *KYCHandler) UploadDocument(ctx context.Context, req *pb.UploadDocumentRequest) (*pb.UploadDocumentResponse, error) {
	cmd := application.UploadDocumentCommand{
		ApplicationID: req.ApplicationId,
		DocumentType:  pbIDTypeToDomain(req.DocumentType),
		Side:          pbSideToDomain(req.Side),
		DocumentData:  req.DocumentData,
		DocumentURL:   req.DocumentUrl,
	}

	doc, err := h.cmd.UploadDocument(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return &pb.UploadDocumentResponse{
		DocumentId:   doc.DocumentID,
		DocumentUrl:  doc.DocumentURL,
		OcrSuccess:   doc.OCRInfo != nil,
		OcrInfo:      toPbDocumentInfoFromDomain(doc.OCRInfo),
	}, nil
}

// SubmitFaceVerification 提交人脸验证
func (h *KYCHandler) SubmitFaceVerification(ctx context.Context, req *pb.SubmitFaceVerificationRequest) (*pb.SubmitFaceVerificationResponse, error) {
	cmd := application.SubmitFaceVerificationCommand{
		ApplicationID: req.ApplicationId,
		FaceImageData: req.FaceImage,
		FaceImageURL:  req.FaceImageUrl,
		LivenessData:  []byte(req.LivenessData),
	}

	fv, err := h.cmd.SubmitFaceVerification(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return &pb.SubmitFaceVerificationResponse{
		VerificationId:  fv.VerificationID,
		Passed:          fv.Passed,
		SimilarityScore: float32(fv.SimilarityScore),
		LivenessPassed:  fv.LivenessPassed,
		Message:         "Face verification completed",
	}, nil
}

// GetKYCApplication 获取KYC申请详情
func (h *KYCHandler) GetKYCApplication(ctx context.Context, req *pb.GetKYCApplicationRequest) (*pb.GetKYCApplicationResponse, error) {
	var dto *application.KYCApplicationDTO
	var err error

	if req.ApplicationId != "" {
		dto, err = h.query.GetKYCApplication(ctx, req.ApplicationId)
	} else {
		dto, err = h.query.GetKYCApplicationByUserID(ctx, req.UserId)
	}

	if err != nil {
		return nil, err
	}

	return &pb.GetKYCApplicationResponse{
		Application: toPbApplication(dto),
	}, nil
}

// GetKYCStatus 获取KYC状态
func (h *KYCHandler) GetKYCStatus(ctx context.Context, req *pb.GetKYCStatusRequest) (*pb.GetKYCStatusResponse, error) {
	dto, err := h.query.GetKYCStatus(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	resp := &pb.GetKYCStatusResponse{
		UserId:     dto.UserID,
		Status:     pbStatusFromString(dto.Status),
		Level:      pbLevelFromString(dto.Level),
		Reason:     dto.Reason,
		IsExpired:  dto.IsExpired,
	}

	if dto.VerifiedAt != nil {
		resp.VerifiedAt = timestamppb.New(*dto.VerifiedAt)
	}
	if dto.ExpiresAt != nil {
		resp.ExpiresAt = timestamppb.New(*dto.ExpiresAt)
	}

	return resp, nil
}

// CancelKYCApplication 取消KYC申请
func (h *KYCHandler) CancelKYCApplication(ctx context.Context, req *pb.CancelKYCApplicationRequest) (*emptypb.Empty, error) {
	err := h.cmd.CancelKYC(ctx, req.ApplicationId, req.UserId)
	return &emptypb.Empty{}, err
}

// ResubmitKYC 重新提交KYC
func (h *KYCHandler) ResubmitKYC(ctx context.Context, req *pb.ResubmitKYCRequest) (*pb.SubmitKYCResponse, error) {
	cmd := application.SubmitKYCCommand{
		UserID:   req.UserId,
		FullName: req.FullName,
		IDNumber: req.IdNumber,
		IDType:   pbIDTypeToDomain(req.IdType),
	}

	app, err := h.cmd.SubmitKYC(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return &pb.SubmitKYCResponse{
		ApplicationId:    app.ApplicationID,
		Status:           pbStatusFromDomain(app.Status),
		CurrentLevel:     pbLevelFromDomain(app.Level),
		RequiredDocuments: getRequiredDocumentTypes(app),
		Message:          "KYC application resubmitted successfully",
	}, nil
}

// RecognizeIDDocument 证件OCR识别
func (h *KYCHandler) RecognizeIDDocument(ctx context.Context, req *pb.RecognizeIDDocumentRequest) (*pb.RecognizeIDDocumentResponse, error) {
	return &pb.RecognizeIDDocumentResponse{
		Success:       false,
		ErrorMessage:  "OCR service not implemented",
	}, nil
}

// ListPendingApplications 获取待审核列表
func (h *KYCHandler) ListPendingApplications(ctx context.Context, req *pb.ListPendingApplicationsRequest) (*pb.ListPendingApplicationsResponse, error) {
	dtos, total, err := h.query.ListPendingApplications(
		ctx,
		int(req.Page),
		int(req.PageSize),
		pbLevelToDomain(req.LevelFilter),
		req.CountryFilter,
	)
	if err != nil {
		return nil, err
	}

	applications := make([]*pb.KYCApplication, len(dtos))
	for i, dto := range dtos {
		applications[i] = &pb.KYCApplication{
			ApplicationId: dto.ApplicationID,
			UserId:        dto.UserID,
			FullName:      dto.FullName,
			IdType:        pbIDTypeFromString(dto.IDType),
			Level:         pbLevelFromString(dto.Level),
			RiskScore:     int32(dto.RiskScore),
			Status:        pbStatusFromString(dto.Status),
			CreatedAt:     timestamppb.New(dto.CreatedAt),
		}
	}

	return &pb.ListPendingApplicationsResponse{
		Applications: applications,
		Total:        int32(total),
		Page:         req.Page,
		PageSize:     req.PageSize,
	}, nil
}

// ApproveKYC 审核通过
func (h *KYCHandler) ApproveKYC(ctx context.Context, req *pb.ApproveKYCRequest) (*emptypb.Empty, error) {
	cmd := application.ApproveKYCCommand{
		ApplicationID:  req.ApplicationId,
		AuditorID:      req.AuditorId,
		AuditorName:    req.AuditorName,
		ApprovedLevel:  pbLevelToDomain(req.ApprovedLevel),
		Comment:        req.Comment,
		ValidityMonths: int(req.ValidityMonths),
	}

	err := h.cmd.ApproveKYC(ctx, cmd)
	return &emptypb.Empty{}, err
}

// RejectKYC 审核拒绝
func (h *KYCHandler) RejectKYC(ctx context.Context, req *pb.RejectKYCRequest) (*emptypb.Empty, error) {
	cmd := application.RejectKYCCommand{
		ApplicationID: req.ApplicationId,
		AuditorID:     req.AuditorId,
		AuditorName:   req.AuditorName,
		Reason:        req.Reason,
		Comment:       req.Comment,
	}

	err := h.cmd.RejectKYC(ctx, cmd)
	return &emptypb.Empty{}, err
}

// GetAuditHistory 获取审核历史
func (h *KYCHandler) GetAuditHistory(ctx context.Context, req *pb.GetAuditHistoryRequest) (*pb.GetAuditHistoryResponse, error) {
	records, total, err := h.query.GetAuditHistory(ctx, req.ApplicationId, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}

	pbRecords := make([]*pb.AuditRecord, len(records))
	for i, r := range records {
		pbRecords[i] = &pb.AuditRecord{
			RecordId:    r.RecordID,
			AuditorId:   r.AuditorID,
			AuditorName: r.AuditorName,
			Action:      pbActionFromString(r.Action),
			Reason:      r.Reason,
			Comment:     r.Comment,
			CreatedAt:   timestamppb.New(r.CreatedAt),
		}
	}

	return &pb.GetAuditHistoryResponse{
		Records: pbRecords,
		Total:   int32(total),
	}, nil
}

// GetUserKYCLevel 获取用户KYC等级
func (h *KYCHandler) GetUserKYCLevel(ctx context.Context, req *pb.GetUserKYCLevelRequest) (*pb.GetUserKYCLevelResponse, error) {
	dto, err := h.query.GetKYCStatus(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	resp := &pb.GetUserKYCLevelResponse{
		UserId: dto.UserID,
		Level:  pbLevelFromString(dto.Level),
		Status: pbStatusFromString(dto.Status),
	}
	if dto.ExpiresAt != nil {
		resp.ExpiresAt = timestamppb.New(*dto.ExpiresAt)
	}
	return resp, nil
}

// UpgradeKYCLevel 升级KYC等级
func (h *KYCHandler) UpgradeKYCLevel(ctx context.Context, req *pb.UpgradeKYCLevelRequest) (*pb.UpgradeKYCLevelResponse, error) {
	return &pb.UpgradeKYCLevelResponse{
		ApplicationId:    "",
		CurrentLevel:     req.TargetLevel,
		TargetLevel:      req.TargetLevel,
		RequiredDocuments: []string{},
	}, nil
}

// GetRiskScore 获取风险评分
func (h *KYCHandler) GetRiskScore(ctx context.Context, req *pb.GetRiskScoreRequest) (*pb.GetRiskScoreResponse, error) {
	result, err := h.cmd.CalculateRiskScore(ctx, req.ApplicationId)
	if err != nil {
		return nil, err
	}

	factors := make([]*pb.RiskFactor, len(result.Factors))
	for i, f := range result.Factors {
		factors[i] = &pb.RiskFactor{
			FactorType:  f.FactorType,
			Description: f.Description,
			Score:       int32(f.Score),
			Weight:      float32(f.Weight),
		}
	}

	return &pb.GetRiskScoreResponse{
		UserId:        req.UserId,
		RiskScore:     int32(result.Score),
		RiskLevel:     result.Level,
		Factors:       factors,
		CalculatedAt:  timestamppb.Now(),
	}, nil
}

// SubmitMerchantKYC 提交商家KYC
func (h *KYCHandler) SubmitMerchantKYC(ctx context.Context, req *pb.SubmitMerchantKYCRequest) (*pb.SubmitKYCResponse, error) {
	cmd := application.SubmitMerchantKYCCommand{
		MerchantID:                req.MerchantId,
		CompanyName:               req.CompanyName,
		RegistrationNumber:        req.RegistrationNumber,
		TaxID:                     req.TaxId,
		LegalRepresentativeName:   req.LegalRepresentativeName,
		LegalRepresentativeID:     req.LegalRepresentativeId,
		BusinessAddress:           req.BusinessAddress,
		Country:                   req.Country,
		Province:                  req.Province,
		City:                      req.City,
		PostalCode:                req.PostalCode,
		ContactPhone:              req.ContactPhone,
		ContactEmail:              req.ContactEmail,
		BankName:                  req.BankName,
		BankAccount:               req.BankAccount,
		BankAccountName:           req.BankAccountName,
	}

	app, err := h.cmd.SubmitMerchantKYC(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return &pb.SubmitKYCResponse{
		ApplicationId: app.ApplicationID,
		Status:        pbStatusFromDomain(app.Status),
		CurrentLevel:  pbLevelFromDomain(app.Level),
		Message:       "Merchant KYC application submitted successfully",
	}, nil
}

// GetMerchantKYC 获取商家KYC详情
func (h *KYCHandler) GetMerchantKYC(ctx context.Context, req *pb.GetMerchantKYCRequest) (*pb.GetMerchantKYCResponse, error) {
	dto, err := h.query.GetMerchantKYC(ctx, req.MerchantId)
	if err != nil {
		return nil, err
	}

	return &pb.GetMerchantKYCResponse{
		MerchantInfo: &pb.MerchantKYCInfo{
			CompanyName:             dto.CompanyName,
			RegistrationNumber:      dto.RegistrationNumber,
			TaxId:                   dto.TaxID,
			LegalRepresentativeName: dto.LegalRepresentativeName,
			LegalRepresentativeId:   dto.LegalRepresentativeID,
			BusinessAddress:         dto.BusinessAddress,
			Country:                 dto.Country,
			Province:                dto.Province,
			City:                    dto.City,
			PostalCode:              dto.PostalCode,
			ContactPhone:            dto.ContactPhone,
			ContactEmail:            dto.ContactEmail,
			BankName:                dto.BankName,
			BankAccount:             dto.BankAccount,
			BankAccountName:         dto.BankAccountName,
		},
	}, nil
}

// 辅助转换函数

func pbIDTypeToDomain(t pb.IDType) domain.IDType {
	switch t {
	case pb.IDType_ID_TYPE_ID_CARD:
		return domain.IDTypeIDCard
	case pb.IDType_ID_TYPE_PASSPORT:
		return domain.IDTypePassport
	case pb.IDType_ID_TYPE_DRIVERS_LICENSE:
		return domain.IDTypeDriversLicense
	case pb.IDType_ID_TYPE_RESIDENCE_PERMIT:
		return domain.IDTypeResidencePermit
	case pb.IDType_ID_TYPE_HK_MACAU_PASS:
		return domain.IDTypeHKMacauPass
	case pb.IDType_ID_TYPE_TAIWAN_PASS:
		return domain.IDTypeTaiwanPass
	case pb.IDType_ID_TYPE_BUSINESS_LICENSE:
		return domain.IDTypeBusinessLicense
	default:
		return domain.IDTypeUnspecified
	}
}

func pbIDTypeFromString(s string) pb.IDType {
	switch s {
	case "ID_CARD":
		return pb.IDType_ID_TYPE_ID_CARD
	case "PASSPORT":
		return pb.IDType_ID_TYPE_PASSPORT
	case "DRIVERS_LICENSE":
		return pb.IDType_ID_TYPE_DRIVERS_LICENSE
	case "RESIDENCE_PERMIT":
		return pb.IDType_ID_TYPE_RESIDENCE_PERMIT
	case "HK_MACAU_PASS":
		return pb.IDType_ID_TYPE_HK_MACAU_PASS
	case "TAIWAN_PASS":
		return pb.IDType_ID_TYPE_TAIWAN_PASS
	case "BUSINESS_LICENSE":
		return pb.IDType_ID_TYPE_BUSINESS_LICENSE
	default:
		return pb.IDType_ID_TYPE_UNSPECIFIED
	}
}

func pbIDTypeFromDomain(t domain.IDType) pb.IDType {
	switch t {
	case domain.IDTypeIDCard:
		return pb.IDType_ID_TYPE_ID_CARD
	case domain.IDTypePassport:
		return pb.IDType_ID_TYPE_PASSPORT
	case domain.IDTypeDriversLicense:
		return pb.IDType_ID_TYPE_DRIVERS_LICENSE
	case domain.IDTypeResidencePermit:
		return pb.IDType_ID_TYPE_RESIDENCE_PERMIT
	case domain.IDTypeHKMacauPass:
		return pb.IDType_ID_TYPE_HK_MACAU_PASS
	case domain.IDTypeTaiwanPass:
		return pb.IDType_ID_TYPE_TAIWAN_PASS
	case domain.IDTypeBusinessLicense:
		return pb.IDType_ID_TYPE_BUSINESS_LICENSE
	default:
		return pb.IDType_ID_TYPE_UNSPECIFIED
	}
}

func pbSideToDomain(s pb.DocumentSide) domain.DocumentSide {
	switch s {
	case pb.DocumentSide_DOCUMENT_SIDE_FRONT:
		return domain.DocumentSideFront
	case pb.DocumentSide_DOCUMENT_SIDE_BACK:
		return domain.DocumentSideBack
	case pb.DocumentSide_DOCUMENT_SIDE_SELFIE:
		return domain.DocumentSideSelfie
	default:
		return domain.DocumentSideUnspecified
	}
}

func pbStatusFromDomain(s domain.KYCStatus) pb.KYCStatus {
	switch s {
	case domain.KYCStatusDraft:
		return pb.KYCStatus_KYC_STATUS_DRAFT
	case domain.KYCStatusPending:
		return pb.KYCStatus_KYC_STATUS_PENDING
	case domain.KYCStatusUnderReview:
		return pb.KYCStatus_KYC_STATUS_UNDER_REVIEW
	case domain.KYCStatusApproved:
		return pb.KYCStatus_KYC_STATUS_APPROVED
	case domain.KYCStatusRejected:
		return pb.KYCStatus_KYC_STATUS_REJECTED
	case domain.KYCStatusExpired:
		return pb.KYCStatus_KYC_STATUS_EXPIRED
	case domain.KYCStatusCancelled:
		return pb.KYCStatus_KYC_STATUS_CANCELLED
	case domain.KYCStatusSuspended:
		return pb.KYCStatus_KYC_STATUS_SUSPENDED
	default:
		return pb.KYCStatus_KYC_STATUS_UNSPECIFIED
	}
}

func pbStatusFromString(s string) pb.KYCStatus {
	switch s {
	case "DRAFT":
		return pb.KYCStatus_KYC_STATUS_DRAFT
	case "PENDING":
		return pb.KYCStatus_KYC_STATUS_PENDING
	case "UNDER_REVIEW":
		return pb.KYCStatus_KYC_STATUS_UNDER_REVIEW
	case "APPROVED":
		return pb.KYCStatus_KYC_STATUS_APPROVED
	case "REJECTED":
		return pb.KYCStatus_KYC_STATUS_REJECTED
	case "EXPIRED":
		return pb.KYCStatus_KYC_STATUS_EXPIRED
	case "CANCELLED":
		return pb.KYCStatus_KYC_STATUS_CANCELLED
	case "SUSPENDED":
		return pb.KYCStatus_KYC_STATUS_SUSPENDED
	default:
		return pb.KYCStatus_KYC_STATUS_UNSPECIFIED
	}
}

func pbLevelFromDomain(l domain.KYCLevel) pb.KYCLevel {
	switch l {
	case domain.KYCLevelNone:
		return pb.KYCLevel_KYC_LEVEL_NONE
	case domain.KYCLevelBasic:
		return pb.KYCLevel_KYC_LEVEL_BASIC
	case domain.KYCLevelIntermediate:
		return pb.KYCLevel_KYC_LEVEL_INTERMEDIATE
	case domain.KYCLevelAdvanced:
		return pb.KYCLevel_KYC_LEVEL_ADVANCED
	case domain.KYCLevelBusiness:
		return pb.KYCLevel_KYC_LEVEL_BUSINESS
	default:
		return pb.KYCLevel_KYC_LEVEL_UNSPECIFIED
	}
}

func pbLevelFromString(s string) pb.KYCLevel {
	switch s {
	case "NONE":
		return pb.KYCLevel_KYC_LEVEL_NONE
	case "BASIC":
		return pb.KYCLevel_KYC_LEVEL_BASIC
	case "INTERMEDIATE":
		return pb.KYCLevel_KYC_LEVEL_INTERMEDIATE
	case "ADVANCED":
		return pb.KYCLevel_KYC_LEVEL_ADVANCED
	case "BUSINESS":
		return pb.KYCLevel_KYC_LEVEL_BUSINESS
	default:
		return pb.KYCLevel_KYC_LEVEL_UNSPECIFIED
	}
}

func pbLevelToDomain(l pb.KYCLevel) domain.KYCLevel {
	switch l {
	case pb.KYCLevel_KYC_LEVEL_NONE:
		return domain.KYCLevelNone
	case pb.KYCLevel_KYC_LEVEL_BASIC:
		return domain.KYCLevelBasic
	case pb.KYCLevel_KYC_LEVEL_INTERMEDIATE:
		return domain.KYCLevelIntermediate
	case pb.KYCLevel_KYC_LEVEL_ADVANCED:
		return domain.KYCLevelAdvanced
	case pb.KYCLevel_KYC_LEVEL_BUSINESS:
		return domain.KYCLevelBusiness
	default:
		return domain.KYCLevelUnspecified
	}
}

func pbActionFromString(a string) pb.AuditAction {
	switch a {
	case "APPROVE":
		return pb.AuditAction_AUDIT_ACTION_APPROVE
	case "REJECT":
		return pb.AuditAction_AUDIT_ACTION_REJECT
	case "REQUEST_INFO":
		return pb.AuditAction_AUDIT_ACTION_REQUEST_INFO
	case "ESCALATE":
		return pb.AuditAction_AUDIT_ACTION_ESCALATE
	default:
		return pb.AuditAction_AUDIT_ACTION_UNSPECIFIED
	}
}

func toPbApplication(dto *application.KYCApplicationDTO) *pb.KYCApplication {
	app := &pb.KYCApplication{
		ApplicationId: dto.ApplicationID,
		UserId:        dto.UserID,
		FullName:      dto.FullName,
		IdNumber:      dto.IDNumber,
		IdType:        pbIDTypeFromString(dto.IDType),
		Nationality:   dto.Nationality,
		Gender:        dto.Gender,
		Address:       dto.Address,
		Country:       dto.Country,
		Province:      dto.Province,
		City:          dto.City,
		PostalCode:    dto.PostalCode,
		Status:        pbStatusFromString(dto.Status),
		Level:         pbLevelFromString(dto.Level),
		RiskScore:     int32(dto.RiskScore),
		CreatedAt:     timestamppb.New(dto.CreatedAt),
		UpdatedAt:     timestamppb.New(dto.UpdatedAt),
	}

	if dto.BirthDate != nil {
		app.BirthDate = timestamppb.New(*dto.BirthDate)
	}
	if dto.VerifiedAt != nil {
		app.VerifiedAt = timestamppb.New(*dto.VerifiedAt)
	}
	if dto.ExpiresAt != nil {
		app.ExpiresAt = timestamppb.New(*dto.ExpiresAt)
	}

	for _, doc := range dto.Documents {
		app.Documents = append(app.Documents, toPbDocument(doc))
	}

	if dto.FaceVerification != nil {
		app.FaceVerification = &pb.FaceVerification{
			VerificationId:  dto.FaceVerification.VerificationID,
			Passed:          dto.FaceVerification.Passed,
			SimilarityScore: float32(dto.FaceVerification.SimilarityScore),
			LivenessPassed:  dto.FaceVerification.LivenessPassed,
			FaceImageUrl:    dto.FaceVerification.FaceImageURL,
			VerifiedAt:      timestamppb.New(dto.FaceVerification.VerifiedAt),
		}
	}

	for _, r := range dto.AuditRecords {
		app.AuditRecords = append(app.AuditRecords, &pb.AuditRecord{
			RecordId:    r.RecordID,
			AuditorId:   r.AuditorID,
			AuditorName: r.AuditorName,
			Action:      pbActionFromString(r.Action),
			Reason:      r.Reason,
			Comment:     r.Comment,
			CreatedAt:   timestamppb.New(r.CreatedAt),
		})
	}

	return app
}

func toPbDocument(dto *application.DocumentDTO) *pb.Document {
	doc := &pb.Document{
		DocumentId:   dto.DocumentID,
		DocumentType: pbIDTypeFromString(dto.DocumentType),
		Side:         toPbSide(dto.Side),
		DocumentUrl:  dto.DocumentURL,
		Verified:     dto.Verified,
		UploadedAt:   timestamppb.New(dto.UploadedAt),
	}

	if dto.OCRInfo != nil {
		doc.OcrInfo = toPbDocumentInfo(dto.OCRInfo)
	}

	return doc
}

func toPbSide(s string) pb.DocumentSide {
	switch s {
	case "FRONT":
		return pb.DocumentSide_DOCUMENT_SIDE_FRONT
	case "BACK":
		return pb.DocumentSide_DOCUMENT_SIDE_BACK
	case "SELFIE":
		return pb.DocumentSide_DOCUMENT_SIDE_SELFIE
	default:
		return pb.DocumentSide_DOCUMENT_SIDE_UNSPECIFIED
	}
}

func toPbDocumentInfo(info *application.IDDocumentInfoDTO) *pb.IDDocumentInfo {
	if info == nil {
		return nil
	}

	pbInfo := &pb.IDDocumentInfo{
		Name:             info.Name,
		IdNumber:         info.IDNumber,
		Gender:           info.Gender,
		Nationality:      info.Nationality,
		Address:          info.Address,
		IssuingAuthority: info.IssuingAuthority,
		ConfidenceScore:  float32(info.ConfidenceScore),
	}

	if info.BirthDate != nil {
		pbInfo.BirthDate = timestamppb.New(*info.BirthDate)
	}
	if info.IssueDate != nil {
		pbInfo.IssueDate = timestamppb.New(*info.IssueDate)
	}
	if info.ExpiryDate != nil {
		pbInfo.ExpiryDate = timestamppb.New(*info.ExpiryDate)
	}

	return pbInfo
}

func toPbDocumentInfoFromDomain(info *domain.IDDocumentInfo) *pb.IDDocumentInfo {
	if info == nil {
		return nil
	}

	pbInfo := &pb.IDDocumentInfo{
		Name:             info.Name,
		IdNumber:         info.IDNumber,
		Gender:           info.Gender,
		Nationality:      info.Nationality,
		Address:          info.Address,
		IssuingAuthority: info.IssuingAuthority,
		ConfidenceScore:  float32(info.ConfidenceScore),
	}

	if info.BirthDate != nil {
		pbInfo.BirthDate = timestamppb.New(*info.BirthDate)
	}
	if info.IssueDate != nil {
		pbInfo.IssueDate = timestamppb.New(*info.IssueDate)
	}
	if info.ExpiryDate != nil {
		pbInfo.ExpiryDate = timestamppb.New(*info.ExpiryDate)
	}

	return pbInfo
}

func getRequiredDocumentTypes(app *domain.KYCApplication) []string {
	types := app.GetRequiredDocuments()
	result := make([]string, len(types))
	for i, t := range types {
		result[i] = t.String()
	}
	return result
}
