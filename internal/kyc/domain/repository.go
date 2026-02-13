// 变更说明：KYC 仓储接口定义
package domain

import (
	"context"
)

// KYCRepository KYC申请仓储接口
type KYCRepository interface {
	// Save 保存KYC申请
	Save(ctx context.Context, app *KYCApplication) error
	
	// Update 更新KYC申请
	Update(ctx context.Context, app *KYCApplication) error
	
	// FindByID 根据ID查询
	FindByID(ctx context.Context, id uint64) (*KYCApplication, error)
	
	// FindByApplicationID 根据申请ID查询
	FindByApplicationID(ctx context.Context, applicationID string) (*KYCApplication, error)
	
	// FindByUserID 根据用户ID查询最新申请
	FindByUserID(ctx context.Context, userID uint64) (*KYCApplication, error)
	
	// FindByUserIDAndStatus 根据用户ID和状态查询
	FindByUserIDAndStatus(ctx context.Context, userID uint64, status KYCStatus) (*KYCApplication, error)
	
	// FindPendingApplications 查询待审核申请列表
	FindPendingApplications(ctx context.Context, page, pageSize int, levelFilter KYCLevel, countryFilter string) ([]*KYCApplication, int64, error)
	
	// FindExpiringApplications 查询即将过期的申请
	FindExpiringApplications(ctx context.Context, beforeDays int) ([]*KYCApplication, error)
	
	// WithTx 在事务中执行
	WithTx(ctx context.Context, fn func(txCtx context.Context) error) error
}

// DocumentRepository 证件仓储接口
type DocumentRepository interface {
	// Save 保存证件
	Save(ctx context.Context, doc *Document) error
	
	// FindByID 根据ID查询
	FindByID(ctx context.Context, id uint64) (*Document, error)
	
	// FindByDocumentID 根据证件ID查询
	FindByDocumentID(ctx context.Context, documentID string) (*Document, error)
	
	// FindByApplicationID 查询申请的所有证件
	FindByApplicationID(ctx context.Context, applicationID string) ([]*Document, error)
	
	// Delete 删除证件
	Delete(ctx context.Context, id uint64) error
}

// FaceVerificationRepository 人脸验证仓储接口
type FaceVerificationRepository interface {
	// Save 保存人脸验证记录
	Save(ctx context.Context, fv *FaceVerification) error
	
	// FindByApplicationID 根据申请ID查询
	FindByApplicationID(ctx context.Context, applicationID string) (*FaceVerification, error)
	
	// FindByVerificationID 根据验证ID查询
	FindByVerificationID(ctx context.Context, verificationID string) (*FaceVerification, error)
}

// AuditRecordRepository 审核记录仓储接口
type AuditRecordRepository interface {
	// Save 保存审核记录
	Save(ctx context.Context, record *AuditRecord) error
	
	// FindByApplicationID 查询申请的审核记录
	FindByApplicationID(ctx context.Context, applicationID string, page, pageSize int) ([]*AuditRecord, int64, error)
}

// MerchantKYCRepository 商家KYC仓储接口
type MerchantKYCRepository interface {
	// Save 保存商家KYC信息
	Save(ctx context.Context, info *MerchantKYCInfo) error
	
	// FindByApplicationID 根据申请ID查询
	FindByApplicationID(ctx context.Context, applicationID string) (*MerchantKYCInfo, error)
	
	// FindByMerchantID 根据商家ID查询
	FindByMerchantID(ctx context.Context, merchantID uint64) (*MerchantKYCInfo, error)
}

// KYCReadRepository KYC读模型仓储接口（Redis）
type KYCReadRepository interface {
	// Save 保存KYC状态到缓存
	Save(ctx context.Context, app *KYCApplication) error
	
	// FindByUserID 从缓存查询用户KYC状态
	FindByUserID(ctx context.Context, userID uint64) (*KYCApplication, error)
	
	// Delete 删除缓存
	Delete(ctx context.Context, userID uint64) error
}

// KYCLimitRepository KYC限额仓储接口
type KYCLimitRepository interface {
	// GetLimits 获取用户KYC限额
	GetLimits(ctx context.Context, userID uint64, level KYCLevel) ([]*KYCLimit, error)
	
	// UpdateUsedValue 更新已用值
	UpdateUsedValue(ctx context.Context, userID uint64, limitType string, value string) error
}
