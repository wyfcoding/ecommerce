package model

// FraudCheckResult represents the result of an anti-fraud check in the business logic layer.
type FraudCheckResult struct {
	ID             uint
	UserID         string
	IPAddress      string
	DeviceInfo     string
	OrderID        string
	Amount         uint64
	IsFraud        bool
	RiskScore      string
	Decision       string
	Message        string
	AdditionalData map[string]string
}
