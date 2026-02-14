// 变更说明：KYC 仓储 MySQL 实现
package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/kyc/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// KYCApplicationModel KYC申请数据库模型
type KYCApplicationModel struct {
	gorm.Model
	ApplicationID string     `gorm:"column:application_id;type:varchar(64);uniqueIndex;not null"`
	UserID        uint64     `gorm:"column:user_id;index;not null"`
	UserType      int        `gorm:"column:user_type;default:0"`
	MerchantID    uint64     `gorm:"column:merchant_id;index"`
	FullName      string     `gorm:"column:full_name;type:varchar(128);not null"`
	IDNumber      string     `gorm:"column:id_number;type:varchar(64);index;not null"`
	IDType        int        `gorm:"column:id_type;not null"`
	Nationality   string     `gorm:"column:nationality;type:varchar(64)"`
	BirthDate     *time.Time `gorm:"column:birth_date"`
	Gender        string     `gorm:"column:gender;type:char(1)"`
	Address       string     `gorm:"column:address;type:varchar(512)"`
	Country       string     `gorm:"column:country;type:varchar(32)"`
	Province      string     `gorm:"column:province;type:varchar(64)"`
	City          string     `gorm:"column:city;type:varchar(64)"`
	PostalCode    string     `gorm:"column:postal_code;type:varchar(16)"`
	Status        int        `gorm:"column:status;index;not null;default:1"`
	Level         int        `gorm:"column:level;not null;default:1"`
	RiskScore     int        `gorm:"column:risk_score;default:0"`
	VerifiedAt    *time.Time `gorm:"column:verified_at"`
	ExpiresAt     *time.Time `gorm:"column:expires_at;index"`
	Version       int        `gorm:"column:version;not null;default:1"`
}

// TableName 指定表名
func (KYCApplicationModel) TableName() string {
	return "kyc_applications"
}

// DocumentModel 证件数据库模型
type DocumentModel struct {
	gorm.Model
	ApplicationID  string     `gorm:"column:application_id;type:varchar(64);index;not null"`
	DocumentID     string     `gorm:"column:document_id;type:varchar(64);uniqueIndex;not null"`
	DocumentType   int        `gorm:"column:document_type;not null"`
	Side           int        `gorm:"column:side;not null"`
	DocumentURL    string     `gorm:"column:document_url;type:varchar(512);not null"`
	OCRName        string     `gorm:"column:ocr_name;type:varchar(128)"`
	OCRIDNumber    string     `gorm:"column:ocr_id_number;type:varchar(64)"`
	OCRGender      string     `gorm:"column:ocr_gender;type:char(1)"`
	OCRNationality string     `gorm:"column:ocr_nationality;type:varchar(64)"`
	OCRBirthDate   *time.Time `gorm:"column:ocr_birth_date"`
	OCRAddress     string     `gorm:"column:ocr_address;type:varchar(512)"`
	OCRIssueAuth   string     `gorm:"column:ocr_issue_auth;type:varchar(128)"`
	OCRIssueDate   *time.Time `gorm:"column:ocr_issue_date"`
	OCRExpiryDate  *time.Time `gorm:"column:ocr_expiry_date"`
	OCRConfidence  float64    `gorm:"column:ocr_confidence"`
	Verified       bool       `gorm:"column:verified;default:false"`
	UploadedAt     time.Time  `gorm:"column:uploaded_at;not null"`
}

// TableName 指定表名
func (DocumentModel) TableName() string {
	return "kyc_documents"
}

// FaceVerificationModel 人脸验证数据库模型
type FaceVerificationModel struct {
	gorm.Model
	ApplicationID   string    `gorm:"column:application_id;type:varchar(64);index;not null"`
	VerificationID  string    `gorm:"column:verification_id;type:varchar(64);uniqueIndex;not null"`
	Passed          bool      `gorm:"column:passed;not null"`
	SimilarityScore float64   `gorm:"column:similarity_score;not null"`
	LivenessPassed  bool      `gorm:"column:liveness_passed;not null"`
	FaceImageURL    string    `gorm:"column:face_image_url;type:varchar(512)"`
	VerifiedAt      time.Time `gorm:"column:verified_at;not null"`
}

// TableName 指定表名
func (FaceVerificationModel) TableName() string {
	return "kyc_face_verifications"
}

// AuditRecordModel 审核记录数据库模型
type AuditRecordModel struct {
	gorm.Model
	RecordID      string `gorm:"column:record_id;type:varchar(64);uniqueIndex;not null"`
	ApplicationID string `gorm:"column:application_id;type:varchar(64);index;not null"`
	AuditorID     uint64 `gorm:"column:auditor_id;not null"`
	AuditorName   string `gorm:"column:auditor_name;type:varchar(128);not null"`
	Action        int    `gorm:"column:action;not null"`
	Reason        string `gorm:"column:reason;type:varchar(512)"`
	Comment       string `gorm:"column:comment;type:text"`
}

// TableName 指定表名
func (AuditRecordModel) TableName() string {
	return "kyc_audit_records"
}

// MerchantKYCModel 商家KYC数据库模型
type MerchantKYCModel struct {
	gorm.Model
	ApplicationID           string `gorm:"column:application_id;type:varchar(64);uniqueIndex;not null"`
	MerchantID              uint64 `gorm:"column:merchant_id;index;not null"`
	CompanyName             string `gorm:"column:company_name;type:varchar(256);not null"`
	RegistrationNumber      string `gorm:"column:registration_number;type:varchar(64);not null"`
	TaxID                   string `gorm:"column:tax_id;type:varchar(64)"`
	LegalRepresentativeName string `gorm:"column:legal_representative_name;type:varchar(128);not null"`
	LegalRepresentativeID   string `gorm:"column:legal_representative_id;type:varchar(64);not null"`
	BusinessAddress         string `gorm:"column:business_address;type:varchar(512)"`
	Country                 string `gorm:"column:country;type:varchar(32)"`
	Province                string `gorm:"column:province;type:varchar(64)"`
	City                    string `gorm:"column:city;type:varchar(64)"`
	PostalCode              string `gorm:"column:postal_code;type:varchar(16)"`
	ContactPhone            string `gorm:"column:contact_phone;type:varchar(32)"`
	ContactEmail            string `gorm:"column:contact_email;type:varchar(128)"`
	BankName                string `gorm:"column:bank_name;type:varchar(128)"`
	BankAccount             string `gorm:"column:bank_account;type:varchar(64)"`
	BankAccountName         string `gorm:"column:bank_account_name;type:varchar(128)"`
}

// TableName 指定表名
func (MerchantKYCModel) TableName() string {
	return "kyc_merchant_info"
}

// KYCRepository KYC仓储实现
type KYCRepository struct {
	db *gorm.DB
}

// NewKYCRepository 创建KYC仓储
func NewKYCRepository(db *gorm.DB) domain.KYCRepository {
	return &KYCRepository{db: db}
}

// Save 保存KYC申请
func (r *KYCRepository) Save(ctx context.Context, app *domain.KYCApplication) error {
	model := toApplicationModel(app)
	return r.db.WithContext(ctx).Create(model).Error
}

// Update 更新KYC申请
func (r *KYCRepository) Update(ctx context.Context, app *domain.KYCApplication) error {
	model := toApplicationModel(app)
	result := r.db.WithContext(ctx).Model(&KYCApplicationModel{}).
		Where("id = ? AND version = ?", model.ID, model.Version).
		Updates(map[string]any{
			"full_name":   model.FullName,
			"id_number":   model.IDNumber,
			"id_type":     model.IDType,
			"nationality": model.Nationality,
			"birth_date":  model.BirthDate,
			"gender":      model.Gender,
			"address":     model.Address,
			"country":     model.Country,
			"province":    model.Province,
			"city":        model.City,
			"postal_code": model.PostalCode,
			"status":      model.Status,
			"level":       model.Level,
			"risk_score":  model.RiskScore,
			"verified_at": model.VerifiedAt,
			"expires_at":  model.ExpiresAt,
			"version":     gorm.Expr("version + 1"),
			"updated_at":  time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("optimistic lock conflict")
	}
	return nil
}

// FindByID 根据ID查询
func (r *KYCRepository) FindByID(ctx context.Context, id uint64) (*domain.KYCApplication, error) {
	var model KYCApplicationModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.loadRelations(ctx, &model)
}

// FindByApplicationID 根据申请ID查询
func (r *KYCRepository) FindByApplicationID(ctx context.Context, applicationID string) (*domain.KYCApplication, error) {
	var model KYCApplicationModel
	if err := r.db.WithContext(ctx).Where("application_id = ?", applicationID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.loadRelations(ctx, &model)
}

// FindByUserID 根据用户ID查询最新申请
func (r *KYCRepository) FindByUserID(ctx context.Context, userID uint64) (*domain.KYCApplication, error) {
	var model KYCApplicationModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.loadRelations(ctx, &model)
}

// FindByUserIDAndStatus 根据用户ID和状态查询
func (r *KYCRepository) FindByUserIDAndStatus(ctx context.Context, userID uint64, status domain.KYCStatus) (*domain.KYCApplication, error) {
	var model KYCApplicationModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, int(status)).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.loadRelations(ctx, &model)
}

// FindPendingApplications 查询待审核申请列表
func (r *KYCRepository) FindPendingApplications(ctx context.Context, page, pageSize int, levelFilter domain.KYCLevel, countryFilter string) ([]*domain.KYCApplication, int64, error) {
	query := r.db.WithContext(ctx).Model(&KYCApplicationModel{}).
		Where("status IN ?", []int{int(domain.KYCStatusPending), int(domain.KYCStatusUnderReview)})

	if levelFilter != domain.KYCLevelUnspecified {
		query = query.Where("level = ?", int(levelFilter))
	}
	if countryFilter != "" {
		query = query.Where("country = ?", countryFilter)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []KYCApplicationModel
	offset := (page - 1) * pageSize
	if err := query.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	apps := make([]*domain.KYCApplication, len(models))
	for i, m := range models {
		app, err := r.loadRelations(ctx, &m)
		if err != nil {
			return nil, 0, err
		}
		apps[i] = app
	}

	return apps, total, nil
}

// FindExpiringApplications 查询即将过期的申请
func (r *KYCRepository) FindExpiringApplications(ctx context.Context, beforeDays int) ([]*domain.KYCApplication, error) {
	expiryDate := time.Now().AddDate(0, 0, beforeDays)
	var models []KYCApplicationModel
	if err := r.db.WithContext(ctx).
		Where("status = ? AND expires_at IS NOT NULL AND expires_at <= ?", int(domain.KYCStatusApproved), expiryDate).
		Find(&models).Error; err != nil {
		return nil, err
	}

	apps := make([]*domain.KYCApplication, len(models))
	for i, m := range models {
		app, err := r.loadRelations(ctx, &m)
		if err != nil {
			return nil, err
		}
		apps[i] = app
	}

	return apps, nil
}

// WithTx 在事务中执行
func (r *KYCRepository) WithTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, "tx", tx)
		return fn(txCtx)
	})
}

// loadRelations 加载关联数据
func (r *KYCRepository) loadRelations(ctx context.Context, model *KYCApplicationModel) (*domain.KYCApplication, error) {
	app := toApplicationDomain(model)

	// 加载证件
	var docModels []DocumentModel
	if err := r.db.WithContext(ctx).
		Where("application_id = ?", model.ApplicationID).
		Find(&docModels).Error; err != nil {
		return nil, err
	}
	for _, dm := range docModels {
		app.Documents = append(app.Documents, toDocumentDomain(&dm))
	}

	// 加载人脸验证
	var fvModel FaceVerificationModel
	if err := r.db.WithContext(ctx).
		Where("application_id = ?", model.ApplicationID).
		First(&fvModel).Error; err == nil {
		app.FaceVerification = toFaceVerificationDomain(&fvModel)
	}

	// 加载审核记录
	var auditModels []AuditRecordModel
	if err := r.db.WithContext(ctx).
		Where("application_id = ?", model.ApplicationID).
		Order("created_at ASC").
		Find(&auditModels).Error; err != nil {
		return nil, err
	}
	for _, am := range auditModels {
		app.AuditRecords = append(app.AuditRecords, toAuditRecordDomain(&am))
	}

	return app, nil
}

// toApplicationModel 转换为数据库模型
func toApplicationModel(app *domain.KYCApplication) *KYCApplicationModel {
	return &KYCApplicationModel{
		Model: gorm.Model{
			ID:        uint(app.ID),
			CreatedAt: app.CreatedAt,
			UpdatedAt: app.UpdatedAt,
		},
		ApplicationID: app.ApplicationID,
		UserID:        app.UserID,
		UserType:      int(app.UserType),
		MerchantID:    app.MerchantID,
		FullName:      app.FullName,
		IDNumber:      app.IDNumber,
		IDType:        int(app.IDType),
		Nationality:   app.Nationality,
		BirthDate:     app.BirthDate,
		Gender:        app.Gender,
		Address:       app.Address,
		Country:       app.Country,
		Province:      app.Province,
		City:          app.City,
		PostalCode:    app.PostalCode,
		Status:        int(app.Status),
		Level:         int(app.Level),
		RiskScore:     app.RiskScore,
		VerifiedAt:    app.VerifiedAt,
		ExpiresAt:     app.ExpiresAt,
		Version:       app.Version,
	}
}

// toApplicationDomain 转换为领域模型
func toApplicationDomain(model *KYCApplicationModel) *domain.KYCApplication {
	return &domain.KYCApplication{
		ID:            uint64(model.ID),
		ApplicationID: model.ApplicationID,
		UserID:        model.UserID,
		UserType:      domain.UserType(model.UserType),
		MerchantID:    model.MerchantID,
		FullName:      model.FullName,
		IDNumber:      model.IDNumber,
		IDType:        domain.IDType(model.IDType),
		Nationality:   model.Nationality,
		BirthDate:     model.BirthDate,
		Gender:        model.Gender,
		Address:       model.Address,
		Country:       model.Country,
		Province:      model.Province,
		City:          model.City,
		PostalCode:    model.PostalCode,
		Status:        domain.KYCStatus(model.Status),
		Level:         domain.KYCLevel(model.Level),
		RiskScore:     model.RiskScore,
		Documents:     []*domain.Document{},
		AuditRecords:  []*domain.AuditRecord{},
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
		VerifiedAt:    model.VerifiedAt,
		ExpiresAt:     model.ExpiresAt,
		Version:       model.Version,
	}
}

// toDocumentModel 转换为数据库模型
func toDocumentModel(doc *domain.Document) *DocumentModel {
	model := &DocumentModel{
		ApplicationID: doc.ApplicationID,
		DocumentID:    doc.DocumentID,
		DocumentType:  int(doc.DocumentType),
		Side:          int(doc.Side),
		DocumentURL:   doc.DocumentURL,
		Verified:      doc.Verified,
		UploadedAt:    doc.UploadedAt,
	}
	if doc.OCRInfo != nil {
		model.OCRName = doc.OCRInfo.Name
		model.OCRIDNumber = doc.OCRInfo.IDNumber
		model.OCRGender = doc.OCRInfo.Gender
		model.OCRNationality = doc.OCRInfo.Nationality
		model.OCRBirthDate = doc.OCRInfo.BirthDate
		model.OCRAddress = doc.OCRInfo.Address
		model.OCRIssueAuth = doc.OCRInfo.IssuingAuthority
		model.OCRIssueDate = doc.OCRInfo.IssueDate
		model.OCRExpiryDate = doc.OCRInfo.ExpiryDate
		model.OCRConfidence = doc.OCRInfo.ConfidenceScore
	}
	return model
}

// toDocumentDomain 转换为领域模型
func toDocumentDomain(model *DocumentModel) *domain.Document {
	doc := &domain.Document{
		ID:            uint64(model.ID),
		DocumentID:    model.DocumentID,
		ApplicationID: model.ApplicationID,
		DocumentType:  domain.IDType(model.DocumentType),
		Side:          domain.DocumentSide(model.Side),
		DocumentURL:   model.DocumentURL,
		Verified:      model.Verified,
		UploadedAt:    model.UploadedAt,
	}
	if model.OCRName != "" {
		doc.OCRInfo = &domain.IDDocumentInfo{
			Name:             model.OCRName,
			IDNumber:         model.OCRIDNumber,
			Gender:           model.OCRGender,
			Nationality:      model.OCRNationality,
			BirthDate:        model.OCRBirthDate,
			Address:          model.OCRAddress,
			IssuingAuthority: model.OCRIssueAuth,
			IssueDate:        model.OCRIssueDate,
			ExpiryDate:       model.OCRExpiryDate,
			ConfidenceScore:  model.OCRConfidence,
		}
	}
	return doc
}

// toFaceVerificationDomain 转换为领域模型
func toFaceVerificationDomain(model *FaceVerificationModel) *domain.FaceVerification {
	return &domain.FaceVerification{
		ID:              uint64(model.ID),
		VerificationID:  model.VerificationID,
		ApplicationID:   model.ApplicationID,
		Passed:          model.Passed,
		SimilarityScore: model.SimilarityScore,
		LivenessPassed:  model.LivenessPassed,
		FaceImageURL:    model.FaceImageURL,
		VerifiedAt:      model.VerifiedAt,
	}
}

// toAuditRecordDomain 转换为领域模型
func toAuditRecordDomain(model *AuditRecordModel) *domain.AuditRecord {
	return &domain.AuditRecord{
		ID:            uint64(model.ID),
		RecordID:      model.RecordID,
		ApplicationID: model.ApplicationID,
		AuditorID:     model.AuditorID,
		AuditorName:   model.AuditorName,
		Action:        domain.AuditAction(model.Action),
		Reason:        model.Reason,
		Comment:       model.Comment,
		CreatedAt:     model.CreatedAt,
	}
}

// DocumentRepository 证件仓储实现
type DocumentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository 创建证件仓储
func NewDocumentRepository(db *gorm.DB) domain.DocumentRepository {
	return &DocumentRepository{db: db}
}

// Save 保存证件
func (r *DocumentRepository) Save(ctx context.Context, doc *domain.Document) error {
	model := toDocumentModel(doc)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "document_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"document_url", "verified", "updated_at"}),
	}).Create(model).Error
}

// FindByID 根据ID查询
func (r *DocumentRepository) FindByID(ctx context.Context, id uint64) (*domain.Document, error) {
	var model DocumentModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDocumentDomain(&model), nil
}

// FindByDocumentID 根据证件ID查询
func (r *DocumentRepository) FindByDocumentID(ctx context.Context, documentID string) (*domain.Document, error) {
	var model DocumentModel
	if err := r.db.WithContext(ctx).Where("document_id = ?", documentID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDocumentDomain(&model), nil
}

// FindByApplicationID 查询申请的所有证件
func (r *DocumentRepository) FindByApplicationID(ctx context.Context, applicationID string) ([]*domain.Document, error) {
	var models []DocumentModel
	if err := r.db.WithContext(ctx).Where("application_id = ?", applicationID).Find(&models).Error; err != nil {
		return nil, err
	}
	docs := make([]*domain.Document, len(models))
	for i, m := range models {
		docs[i] = toDocumentDomain(&m)
	}
	return docs, nil
}

// Delete 删除证件
func (r *DocumentRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&DocumentModel{}, id).Error
}

// FaceVerificationRepository 人脸验证仓储实现
type FaceVerificationRepository struct {
	db *gorm.DB
}

// NewFaceVerificationRepository 创建人脸验证仓储
func NewFaceVerificationRepository(db *gorm.DB) domain.FaceVerificationRepository {
	return &FaceVerificationRepository{db: db}
}

// Save 保存人脸验证记录
func (r *FaceVerificationRepository) Save(ctx context.Context, fv *domain.FaceVerification) error {
	model := &FaceVerificationModel{
		ApplicationID:   fv.ApplicationID,
		VerificationID:  fv.VerificationID,
		Passed:          fv.Passed,
		SimilarityScore: fv.SimilarityScore,
		LivenessPassed:  fv.LivenessPassed,
		FaceImageURL:    fv.FaceImageURL,
		VerifiedAt:      fv.VerifiedAt,
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "verification_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"passed", "similarity_score", "liveness_passed", "updated_at"}),
	}).Create(model).Error
}

// FindByApplicationID 根据申请ID查询
func (r *FaceVerificationRepository) FindByApplicationID(ctx context.Context, applicationID string) (*domain.FaceVerification, error) {
	var model FaceVerificationModel
	if err := r.db.WithContext(ctx).Where("application_id = ?", applicationID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toFaceVerificationDomain(&model), nil
}

// FindByVerificationID 根据验证ID查询
func (r *FaceVerificationRepository) FindByVerificationID(ctx context.Context, verificationID string) (*domain.FaceVerification, error) {
	var model FaceVerificationModel
	if err := r.db.WithContext(ctx).Where("verification_id = ?", verificationID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toFaceVerificationDomain(&model), nil
}

// AuditRecordRepository 审核记录仓储实现
type AuditRecordRepository struct {
	db *gorm.DB
}

// NewAuditRecordRepository 创建审核记录仓储
func NewAuditRecordRepository(db *gorm.DB) domain.AuditRecordRepository {
	return &AuditRecordRepository{db: db}
}

// Save 保存审核记录
func (r *AuditRecordRepository) Save(ctx context.Context, record *domain.AuditRecord) error {
	model := &AuditRecordModel{
		RecordID:      record.RecordID,
		ApplicationID: record.ApplicationID,
		AuditorID:     record.AuditorID,
		AuditorName:   record.AuditorName,
		Action:        int(record.Action),
		Reason:        record.Reason,
		Comment:       record.Comment,
	}
	return r.db.WithContext(ctx).Create(model).Error
}

// FindByApplicationID 查询申请的审核记录
func (r *AuditRecordRepository) FindByApplicationID(ctx context.Context, applicationID string, page, pageSize int) ([]*domain.AuditRecord, int64, error) {
	query := r.db.WithContext(ctx).Model(&AuditRecordModel{}).Where("application_id = ?", applicationID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []AuditRecordModel
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	records := make([]*domain.AuditRecord, len(models))
	for i, m := range models {
		records[i] = toAuditRecordDomain(&m)
	}

	return records, total, nil
}

// MerchantKYCRepository 商家KYC仓储实现
type MerchantKYCRepository struct {
	db *gorm.DB
}

// NewMerchantKYCRepository 创建商家KYC仓储
func NewMerchantKYCRepository(db *gorm.DB) domain.MerchantKYCRepository {
	return &MerchantKYCRepository{db: db}
}

// Save 保存商家KYC信息
func (r *MerchantKYCRepository) Save(ctx context.Context, info *domain.MerchantKYCInfo) error {
	model := &MerchantKYCModel{
		ApplicationID:           info.ApplicationID,
		MerchantID:              info.ID,
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
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "application_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"company_name", "registration_number", "tax_id", "legal_representative_name",
			"legal_representative_id", "business_address", "updated_at",
		}),
	}).Create(model).Error
}

// FindByApplicationID 根据申请ID查询
func (r *MerchantKYCRepository) FindByApplicationID(ctx context.Context, applicationID string) (*domain.MerchantKYCInfo, error) {
	var model MerchantKYCModel
	if err := r.db.WithContext(ctx).Where("application_id = ?", applicationID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toMerchantKYCDomain(&model), nil
}

// FindByMerchantID 根据商家ID查询
func (r *MerchantKYCRepository) FindByMerchantID(ctx context.Context, merchantID uint64) (*domain.MerchantKYCInfo, error) {
	var model MerchantKYCModel
	if err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toMerchantKYCDomain(&model), nil
}

// toMerchantKYCDomain 转换为领域模型
func toMerchantKYCDomain(model *MerchantKYCModel) *domain.MerchantKYCInfo {
	return &domain.MerchantKYCInfo{
		ID:                      model.MerchantID,
		ApplicationID:           model.ApplicationID,
		CompanyName:             model.CompanyName,
		RegistrationNumber:      model.RegistrationNumber,
		TaxID:                   model.TaxID,
		LegalRepresentativeName: model.LegalRepresentativeName,
		LegalRepresentativeID:   model.LegalRepresentativeID,
		BusinessAddress:         model.BusinessAddress,
		Country:                 model.Country,
		Province:                model.Province,
		City:                    model.City,
		PostalCode:              model.PostalCode,
		ContactPhone:            model.ContactPhone,
		ContactEmail:            model.ContactEmail,
		BankName:                model.BankName,
		BankAccount:             model.BankAccount,
		BankAccountName:         model.BankAccountName,
	}
}

// GetDB 获取数据库连接（用于事务）
func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok {
		return tx
	}
	return defaultDB
}
