package domain

import (
	"context"
	"time"
)

type DeviceStatus string

const (
	DeviceStatusOnline  DeviceStatus = "ONLINE"
	DeviceStatusOffline DeviceStatus = "OFFLINE"
	DeviceStatusError   DeviceStatus = "ERROR"
)

type IoTDevice struct {
	ID           string
	Name         string
	Type         string
	Status       DeviceStatus
	FirmwareVer  string
	OwnerID      string
	LastSeen     time.Time
	Metadata     map[string]interface{}
}

type TelemetryData struct {
	DeviceID  string
	Timestamp time.Time
	Payload   map[string]interface{}
}

type IoTRepository interface {
	SaveDevice(ctx context.Context, device *IoTDevice) error
	FindDeviceByID(ctx context.Context, id string) (*IoTDevice, error)
	UpdateStatus(ctx context.Context, id string, status DeviceStatus) error
	SaveTelemetry(ctx context.Context, data *TelemetryData) error
}
