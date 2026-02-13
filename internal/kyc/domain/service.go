// 变更说明：KYC 领域服务，包含OCR服务、人脸识别服务、风险评估服务等
package domain

import (
	"context"
	"fmt"
	"time"
)

// OCRService 证件OCR识别服务接口
type OCRService interface {
	// RecognizeIDCard 识别身份证
	RecognizeIDCard(ctx context.Context, imageData []byte, side DocumentSide) (*IDDocumentInfo, error)
	
	// RecognizePassport 识别护照
	RecognizePassport(ctx context.Context, imageData []byte) (*IDDocumentInfo, error)
	
	// RecognizeDriversLicense 识别驾照
	RecognizeDriversLicense(ctx context.Context, imageData []byte) (*IDDocumentInfo, error)
	
	// RecognizeBusinessLicense 识别营业执照
	RecognizeBusinessLicense(ctx context.Context, imageData []byte) (*IDDocumentInfo, error)
}

// FaceRecognitionService 人脸识别服务接口
type FaceRecognitionService interface {
	// VerifyFace 验证人脸
	// idImage: 证件照片, faceImage: 自拍照片
	VerifyFace(ctx context.Context, idImageURL, faceImageURL string) (*FaceVerificationResult, error)
	
	// VerifyLiveness 活体检测
	VerifyLiveness(ctx context.Context, livenessData []byte) (bool, error)
	
	// DetectFace 检测人脸
	DetectFace(ctx context.Context, imageData []byte) (*FaceDetectionResult, error)
}

// FaceVerificationResult 人脸验证结果
type FaceVerificationResult struct {
	Passed          bool    `json:"passed"`
	SimilarityScore float64 `json:"similarity_score"`
	FaceRect        *Rect   `json:"face_rect"`
}

// FaceDetectionResult 人脸检测结果
type FaceDetectionResult struct {
	Detected  bool    `json:"detected"`
	FaceCount int     `json:"face_count"`
	FaceRect  *Rect   `json:"face_rect"`
	Quality   float64 `json:"quality"`
}

// Rect 矩形区域
type Rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// RiskAssessmentService 风险评估服务接口
type RiskAssessmentService interface {
	// CalculateRiskScore 计算风险评分
	CalculateRiskScore(ctx context.Context, app *KYCApplication) (*RiskAssessmentResult, error)
	
	// CheckBlacklist 检查黑名单
	CheckBlacklist(ctx context.Context, idNumber string) (bool, string, error)
	
	// CheckSanctionsList 检查制裁名单
	CheckSanctionsList(ctx context.Context, name string) (bool, error)
}

// RiskAssessmentResult 风险评估结果
type RiskAssessmentResult struct {
	Score    int          `json:"score"`     // 0-100, 越高风险越大
	Level    string       `json:"level"`     // LOW, MEDIUM, HIGH, CRITICAL
	Factors  []RiskFactor `json:"factors"`
}

// DocumentStorageService 证件存储服务接口
type DocumentStorageService interface {
	// UploadDocument 上传证件
	UploadDocument(ctx context.Context, userID uint64, documentType IDType, side DocumentSide, data []byte) (string, error)
	
	// GetDocumentURL 获取证件URL
	GetDocumentURL(ctx context.Context, documentID string) (string, error)
	
	// DeleteDocument 删除证件
	DeleteDocument(ctx context.Context, documentID string) error
}

// NotificationService 通知服务接口
type NotificationService interface {
	// NotifyKYCApproved 通知KYC通过
	NotifyKYCApproved(ctx context.Context, userID uint64, level KYCLevel) error
	
	// NotifyKYCRejected 通知KYC拒绝
	NotifyKYCRejected(ctx context.Context, userID uint64, reason string) error
	
	// NotifyKYCExpiring 通知KYC即将过期
	NotifyKYCExpiring(ctx context.Context, userID uint64, daysLeft int) error
}

// KYCLimitService KYC限额服务接口
type KYCLimitService interface {
	// GetLimitsByLevel 根据等级获取限额
	GetLimitsByLevel(level KYCLevel) map[string]string
	
	// CheckLimit 检查是否超限
	CheckLimit(ctx context.Context, userID uint64, limitType string, amount string) (bool, error)
}

// KYCDomainService KYC领域服务
type KYCDomainService struct {
	ocrService         OCRService
	faceService        FaceRecognitionService
	riskService        RiskAssessmentService
	storageService     DocumentStorageService
	notificationService NotificationService
}

// NewKYCDomainService 创建KYC领域服务
func NewKYCDomainService(
	ocrService OCRService,
	faceService FaceRecognitionService,
	riskService RiskAssessmentService,
	storageService DocumentStorageService,
	notificationService NotificationService,
) *KYCDomainService {
	return &KYCDomainService{
		ocrService:         ocrService,
		faceService:        faceService,
		riskService:        riskService,
		storageService:     storageService,
		notificationService: notificationService,
	}
}

// ProcessDocument 处理证件上传
func (s *KYCDomainService) ProcessDocument(ctx context.Context, app *KYCApplication, docType IDType, side DocumentSide, imageData []byte) (*Document, error) {
	// 1. 上传证件到存储
	docURL, err := s.storageService.UploadDocument(ctx, app.UserID, docType, side, imageData)
	if err != nil {
		return nil, err
	}

	// 2. OCR识别
	var ocrInfo *IDDocumentInfo
	switch docType {
	case IDTypeIDCard:
		ocrInfo, err = s.ocrService.RecognizeIDCard(ctx, imageData, side)
	case IDTypePassport:
		ocrInfo, err = s.ocrService.RecognizePassport(ctx, imageData)
	case IDTypeDriversLicense:
		ocrInfo, err = s.ocrService.RecognizeDriversLicense(ctx, imageData)
	case IDTypeBusinessLicense:
		ocrInfo, err = s.ocrService.RecognizeBusinessLicense(ctx, imageData)
	}

	doc := &Document{
		DocumentID:   generateDocumentID(),
		ApplicationID: app.ApplicationID,
		DocumentType: docType,
		Side:         side,
		DocumentURL:  docURL,
		OCRInfo:      ocrInfo,
		UploadedAt:   time.Now(),
	}

	if err != nil {
		doc.Verified = false
	} else {
		doc.Verified = true
	}

	return doc, nil
}

// ProcessFaceVerification 处理人脸验证
func (s *KYCDomainService) ProcessFaceVerification(ctx context.Context, app *KYCApplication, faceImageURL string, livenessData []byte) (*FaceVerification, error) {
	// 1. 获取证件照片
	var idImageURL string
	for _, doc := range app.Documents {
		if doc.Side == DocumentSideFront {
			idImageURL = doc.DocumentURL
			break
		}
	}

	if idImageURL == "" {
		return nil, ErrNoIDDocument
	}

	// 2. 人脸比对
	result, err := s.faceService.VerifyFace(ctx, idImageURL, faceImageURL)
	if err != nil {
		return nil, err
	}

	// 3. 活体检测
	livenessPassed := true
	if len(livenessData) > 0 {
		livenessPassed, err = s.faceService.VerifyLiveness(ctx, livenessData)
		if err != nil {
			livenessPassed = false
		}
	}

	fv := &FaceVerification{
		VerificationID:  generateVerificationID(),
		ApplicationID:   app.ApplicationID,
		Passed:          result.Passed && livenessPassed,
		SimilarityScore: result.SimilarityScore,
		LivenessPassed:  livenessPassed,
		FaceImageURL:    faceImageURL,
		VerifiedAt:      time.Now(),
	}

	return fv, nil
}

// AssessRisk 风险评估
func (s *KYCDomainService) AssessRisk(ctx context.Context, app *KYCApplication) (*RiskAssessmentResult, error) {
	// 1. 计算风险评分
	result, err := s.riskService.CalculateRiskScore(ctx, app)
	if err != nil {
		return nil, err
	}

	// 2. 检查黑名单
	isBlacklisted, reason, err := s.riskService.CheckBlacklist(ctx, app.IDNumber)
	if err == nil && isBlacklisted {
		result.Score = 100
		result.Level = "CRITICAL"
		result.Factors = append(result.Factors, RiskFactor{
			FactorType:  "BLACKLIST",
			Description: reason,
			Score:       100,
			Weight:      1.0,
		})
	}

	// 3. 检查制裁名单
	isSanctioned, err := s.riskService.CheckSanctionsList(ctx, app.FullName)
	if err == nil && isSanctioned {
		result.Score = 100
		result.Level = "CRITICAL"
		result.Factors = append(result.Factors, RiskFactor{
			FactorType:  "SANCTIONS_LIST",
			Description: "Name found in sanctions list",
			Score:       100,
			Weight:      1.0,
		})
	}

	return result, nil
}

// NotifyApproval 通知审核通过
func (s *KYCDomainService) NotifyApproval(ctx context.Context, userID uint64, level KYCLevel) error {
	if s.notificationService != nil {
		return s.notificationService.NotifyKYCApproved(ctx, userID, level)
	}
	return nil
}

// NotifyRejection 通知审核拒绝
func (s *KYCDomainService) NotifyRejection(ctx context.Context, userID uint64, reason string) error {
	if s.notificationService != nil {
		return s.notificationService.NotifyKYCRejected(ctx, userID, reason)
	}
	return nil
}

// 错误定义
var (
	ErrNoIDDocument = &KYCError{Code: "NO_ID_DOCUMENT", Message: "no ID document found"}
)

// KYCError KYC错误
type KYCError struct {
	Code    string
	Message string
}

func (e *KYCError) Error() string {
	return e.Message
}

// 辅助函数
func generateDocumentID() string {
	return fmt.Sprintf("DOC-%d", time.Now().UnixNano())
}

func generateVerificationID() string {
	return fmt.Sprintf("FV-%d", time.Now().UnixNano())
}
