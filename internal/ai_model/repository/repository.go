package repository

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ecommerce/internal/ai_model/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// StringArray 是一个自定义类型，用于在 GORM 中存储字符串数组为 JSON。
type StringArray []string

// Value 实现 driver.Valuer 接口。
func (sa StringArray) Value() (driver.Value, error) {
	if sa == nil {
		return nil, nil
	}
	return json.Marshal(sa)
}

// Scan 实现 sql.Scanner 接口。
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal StringArray value:", value))
	}
	return json.Unmarshal(bytes, sa)
}

// StringMap 是一个自定义类型，用于在 GORM 中存储 map[string]string 为 JSON。
type StringMap map[string]string

// Value 实现 driver.Valuer 接口。
func (sm StringMap) Value() (driver.Value, error) {
	if sm == nil {
		return nil, nil
	}
	return json.Marshal(sm)
}

// Scan 实现 sql.Scanner 接口。
func (sm *StringMap) Scan(value interface{}) error {
	if value == nil {
		*sm = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal StringMap value:", value))
	}
	return json.Unmarshal(bytes, sm)
}

// --- 接口定义 ---

// ModelMetadataRepo 定义了AI模型元数据的数据访问接口。
type ModelMetadataRepo interface {
	// CreateModelMetadata 创建一个新的模型元数据记录。
	CreateModelMetadata(ctx context.Context, metadata *model.ModelMetadata) (*model.ModelMetadata, error)
	// GetModelMetadataByID 根据ID获取模型元数据。
	GetModelMetadataByID(ctx context.Context, id uint64) (*model.ModelMetadata, error)
	// GetModelMetadataByNameAndVersion 根据模型名称和版本获取模型元数据。
	GetModelMetadataByNameAndVersion(ctx context.Context, modelName, modelVersion string) (*model.ModelMetadata, error)
	// UpdateModelMetadata 更新模型元数据。
	UpdateModelMetadata(ctx context.Context, metadata *model.ModelMetadata) (*model.ModelMetadata, error)
	// ListModelMetadata 获取模型元数据列表。
	ListModelMetadata(ctx context.Context, modelNameKeyword string, pageSize, pageToken int32) ([]*model.ModelMetadata, int32, error)
}

// --- 数据库模型 ---

// DBModelMetadata 对应数据库中的AI模型元数据表。
type DBModelMetadata struct {
	gorm.Model
	ModelName      string      `gorm:"index;not null;type:varchar(128);comment:模型名称"`
	ModelVersion   string      `gorm:"index;not null;type:varchar(64);comment:模型版本"`
	ModelURI       string      `gorm:"type:varchar(255);comment:模型存储URI"`
	DeploymentID   string      `gorm:"uniqueIndex;type:varchar(128);comment:部署ID"`
	Status         string      `gorm:"type:varchar(64);comment:部署状态"`
	DeployedAt     time.Time   `gorm:"comment:部署时间"`
	LastTrainedAt  time.Time   `gorm:"comment:上次训练时间"`
	TrainingJobIDs StringArray `gorm:"type:json;comment:训练任务ID列表"`
	Metadata       StringMap   `gorm:"type:json;comment:其他元数据"`
	ErrorMessage   string      `gorm:"type:text;comment:错误信息"`
}

// TableName 返回 DBModelMetadata 对应的表名。
func (DBModelMetadata) TableName() string {
	return "ai_model_metadata"
}

// --- 数据层核心 ---

// Data 封装了所有数据库操作的 GORM 客户端。
type Data struct {
	db *gorm.DB
}

// NewData 创建一个新的 Data 实例，并执行数据库迁移。
func NewData(db *gorm.DB) (*Data, func(), error) {
	d := &Data{
		db: db,
	}
	zap.S().Info("running database migrations for AI model service...")
	// 自动迁移所有相关的数据库表
	if err := db.AutoMigrate(
		&DBModelMetadata{},
	); err != nil {
		zap.S().Errorf("failed to migrate AI model database: %v", err)
		return nil, nil, fmt.Errorf("failed to migrate AI model database: %w", err)
	}

	cleanup := func() {
		zap.S().Info("closing AI model data layer...")
		// 可以在这里添加数据库连接关闭逻辑，如果 GORM 提供了的话
	}

	return d, cleanup, nil
}

// --- ModelMetadataRepo 实现 ---

// modelMetadataRepository 是 ModelMetadataRepo 接口的 GORM 实现。
type modelMetadataRepository struct {
	*Data
}

// NewModelMetadataRepo 创建一个新的 ModelMetadataRepo 实例。
func NewModelMetadataRepo(data *Data) ModelMetadataRepo {
	return &modelMetadataRepository{data}
}

// CreateModelMetadata 在数据库中创建一个新的模型元数据记录。
func (r *modelMetadataRepository) CreateModelMetadata(ctx context.Context, metadata *model.ModelMetadata) (*model.ModelMetadata, error) {
	dbMetadata := fromBizModelMetadata(metadata)
	if err := r.db.WithContext(ctx).Create(dbMetadata).Error; err != nil {
		zap.S().Errorf("failed to create model metadata for %s:%s in db: %v", metadata.ModelName, metadata.ModelVersion, err)
		return nil, err
	}
	zap.S().Infof("model metadata created in db: %d", dbMetadata.ID)
	return toBizModelMetadata(dbMetadata), nil
}

// GetModelMetadataByID 根据ID从数据库中获取模型元数据。
func (r *modelMetadataRepository) GetModelMetadataByID(ctx context.Context, id uint64) (*model.ModelMetadata, error) {
	var dbMetadata DBModelMetadata
	if err := r.db.WithContext(ctx).First(&dbMetadata, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		zap.S().Errorf("failed to get model metadata by id %d from db: %v", id, err)
		return nil, err
	}
	return toBizModelMetadata(&dbMetadata), nil
}

// GetModelMetadataByNameAndVersion 根据模型名称和版本从数据库中获取模型元数据。
func (r *modelMetadataRepository) GetModelMetadataByNameAndVersion(ctx context.Context, modelName, modelVersion string) (*model.ModelMetadata, error) {
	var dbMetadata DBModelMetadata
	if err := r.db.WithContext(ctx).Where("model_name = ? AND model_version = ?", modelName, modelVersion).First(&dbMetadata).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		zap.S().Errorf("failed to get model metadata for %s:%s from db: %v", modelName, modelVersion, err)
		return nil, err
	}
	return toBizModelMetadata(&dbMetadata), nil
}

// UpdateModelMetadata 更新数据库中的模型元数据。
func (r *modelMetadataRepository) UpdateModelMetadata(ctx context.Context, metadata *model.ModelMetadata) (*model.ModelMetadata, error) {
	dbMetadata := fromBizModelMetadata(metadata)
	// 使用 Select 仅更新非零值字段，或者明确指定要更新的字段
	if err := r.db.WithContext(ctx).Model(&DBModelMetadata{}).Where("id = ?", metadata.ID).Updates(dbMetadata).Error; err != nil {
		zap.S().Errorf("failed to update model metadata %d in db: %v", metadata.ID, err)
		return nil, err
	}
	zap.S().Infof("model metadata updated in db: %d", metadata.ID)
	return r.GetModelMetadataByID(ctx, metadata.ID) // 返回更新后的完整元数据
}

// ListModelMetadata 从数据库中获取模型元数据列表。
func (r *modelMetadataRepository) ListModelMetadata(ctx context.Context, modelNameKeyword string, pageSize, pageToken int32) ([]*model.ModelMetadata, int32, error) {
	var dbMetadatas []*DBModelMetadata
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&DBModelMetadata{})

	if modelNameKeyword != "" {
		query = query.Where("model_name LIKE ?", "%"+modelNameKeyword+"%")
	}

	// 获取总数
	if err := query.Count(&totalCount).Error; err != nil {
		zap.S().Errorf("failed to count model metadatas: %v", err)
		return nil, 0, err
	}

	// 分页查询
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageToken <= 0 {
		pageToken = 1
	}
	offset := (pageToken - 1) * pageSize

	if err := query.Order("created_at DESC").Limit(int(pageSize)).Offset(int(offset)).Find(&dbMetadatas).Error; err != nil {
		zap.S().Errorf("failed to list model metadatas from db: %v", err)
		return nil, 0, err
	}

	bizMetadatas := make([]*model.ModelMetadata, len(dbMetadatas))
	for i, dbMetadata := range dbMetadatas {
		bizMetadatas[i] = toBizModelMetadata(dbMetadata)
	}

	return bizMetadatas, int32(totalCount), nil
}

// --- 模型转换辅助函数 ---

// toBizModelMetadata 将 DBModelMetadata 数据库模型转换为 model.ModelMetadata 业务领域模型。
func toBizModelMetadata(dbMetadata *DBModelMetadata) *model.ModelMetadata {
	if dbMetadata == nil {
		return nil
	}
	return &model.ModelMetadata{
		ID:             uint64(dbMetadata.ID),
		ModelName:      dbMetadata.ModelName,
		ModelVersion:   dbMetadata.ModelVersion,
		ModelURI:       dbMetadata.ModelURI,
		DeploymentID:   dbMetadata.DeploymentID,
		Status:         dbMetadata.Status,
		DeployedAt:     dbMetadata.DeployedAt,
		LastTrainedAt:  dbMetadata.LastTrainedAt,
		TrainingJobIDs: dbMetadata.TrainingJobIDs,
		Metadata:       dbMetadata.Metadata,
		ErrorMessage:   dbMetadata.ErrorMessage,
		CreatedAt:      dbMetadata.CreatedAt,
		UpdatedAt:      dbMetadata.UpdatedAt,
	}
}

// fromBizModelMetadata 将 model.ModelMetadata 业务领域模型转换为 DBModelMetadata 数据库模型。
func fromBizModelMetadata(bizMetadata *model.ModelMetadata) *DBModelMetadata {
	if bizMetadata == nil {
		return nil
	}
	return &DBModelMetadata{
		Model:          gorm.Model{ID: uint(bizMetadata.ID), CreatedAt: bizMetadata.CreatedAt, UpdatedAt: bizMetadata.UpdatedAt},
		ModelName:      bizMetadata.ModelName,
		ModelVersion:   bizMetadata.ModelVersion,
		ModelURI:       bizMetadata.ModelURI,
		DeploymentID:   bizMetadata.DeploymentID,
		Status:         bizMetadata.Status,
		DeployedAt:     bizMetadata.DeployedAt,
		LastTrainedAt:  bizMetadata.LastTrainedAt,
		TrainingJobIDs: bizMetadata.TrainingJobIDs,
		Metadata:       bizMetadata.Metadata,
		ErrorMessage:   bizMetadata.ErrorMessage,
	}
}