package mysql

import (
	"context"
	"encoding/json"
	"time"

	"github.com/wyfcoding/ecommerce/internal/iot/domain"
	"gorm.io/gorm"
)

type DeviceModel struct {
	ID          string `gorm:"primaryKey"`
	Name        string
	Type        string
	Status      string
	FirmwareVer string
	OwnerID     string
	LastSeen    time.Time
	Metadata    string // JSON string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type IoTRepositoryImpl struct {
	db *gorm.DB
}

func NewIoTRepository(db *gorm.DB) domain.IoTRepository {
	return &IoTRepositoryImpl{db: db}
}

func (r *IoTRepositoryImpl) SaveDevice(ctx context.Context, device *domain.IoTDevice) error {
	metadata, _ := json.Marshal(device.Metadata)
	model := &DeviceModel{
		ID:          device.ID,
		Name:        device.Name,
		Type:        device.Type,
		Status:      string(device.Status),
		FirmwareVer: device.FirmwareVer,
		OwnerID:     device.OwnerID,
		LastSeen:    device.LastSeen,
		Metadata:    string(metadata),
	}
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *IoTRepositoryImpl) FindDeviceByID(ctx context.Context, id string) (*domain.IoTDevice, error) {
	var model DeviceModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}

	var metadata map[string]interface{}
	json.Unmarshal([]byte(model.Metadata), &metadata)

	return &domain.IoTDevice{
		ID:          model.ID,
		Name:        model.Name,
		Type:        model.Type,
		Status:      domain.DeviceStatus(model.Status),
		FirmwareVer: model.FirmwareVer,
		OwnerID:     model.OwnerID,
		LastSeen:    model.LastSeen,
		Metadata:    metadata,
	}, nil
}

func (r *IoTRepositoryImpl) UpdateStatus(ctx context.Context, id string, status domain.DeviceStatus) error {
	return r.db.WithContext(ctx).Model(&DeviceModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    string(status),
		"last_seen": time.Now(),
	}).Error
}

func (r *IoTRepositoryImpl) SaveTelemetry(ctx context.Context, data *domain.TelemetryData) error {
	// 实际场景通常存入 TimeSeries DB 或 MongoDB，这里简化为打印或存入日志表
	return nil
}
