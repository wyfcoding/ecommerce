package model

import "time"

// ModelMetadata 存储AI模型的元数据，例如部署信息、训练历史等。
type ModelMetadata struct {
	ID             uint64
	ModelName      string
	ModelVersion   string
	ModelURI       string // 模型存储路径，例如 S3 路径
	DeploymentID   string // 部署ID
	Status         string // 部署状态 (e.g., "DEPLOYED", "FAILED", "UNDEPLOYED")
	DeployedAt     time.Time
	LastTrainedAt  time.Time
	TrainingJobIDs []string // 训练任务ID列表，存储为 JSON 字符串
	Metadata       map[string]string // 其他元数据，存储为 JSON 字符串
	ErrorMessage   string // 部署或训练失败时的错误信息
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// SentimentType 定义了情感分析的类型。
type SentimentType int32

const (
	SentimentTypeUnspecified SentimentType = 0 // 未指定
	SentimentTypePositive    SentimentType = 1 // 积极
	SentimentTypeNeutral     SentimentType = 2 // 中性
	SentimentTypeNegative    SentimentType = 3 // 消极
)