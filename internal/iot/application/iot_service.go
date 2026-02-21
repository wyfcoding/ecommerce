package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/wyfcoding/ecommerce/internal/iot/domain"
)

type IoTService struct {
	repo domain.IoTRepository
}

func NewIoTService(repo domain.IoTRepository) *IoTService {
	return &IoTService{repo: repo}
}

func (s *IoTService) RegisterDevice(ctx context.Context, name, deviceType, ownerID string) (*domain.IoTDevice, error) {
	device := &domain.IoTDevice{
		ID:          uuid.New().String(),
		Name:        name,
		Type:        deviceType,
		Status:      domain.DeviceStatusOnline,
		OwnerID:     ownerID,
		LastSeen:    time.Now(),
		FirmwareVer: "1.0.0",
		Metadata:    make(map[string]interface{}),
	}

	if err := s.repo.SaveDevice(ctx, device); err != nil {
		return nil, err
	}

	return device, nil
}

func (s *IoTService) ReportTelemetry(ctx context.Context, deviceID string, payload map[string]interface{}) error {
	data := &domain.TelemetryData{
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Payload:   payload,
	}

	if err := s.repo.SaveTelemetry(ctx, data); err != nil {
		return err
	}

	return s.repo.UpdateStatus(ctx, deviceID, domain.DeviceStatusOnline)
}
