package domain

import "context"

// DeviceFingerprintReadRepository 定义设备指纹读模型仓储接口（Redis）。
type DeviceFingerprintReadRepository interface {
	Save(ctx context.Context, fp *DeviceFingerprint) error
	GetByDeviceID(ctx context.Context, deviceID string) (*DeviceFingerprint, error)
	DeleteByDeviceID(ctx context.Context, deviceID string) error
}
