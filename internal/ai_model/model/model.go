package model

import "time"

// ModelMetadata 存储AI模型的元数据，例如部署信息、训练历史等。
// 包含了模型的版本、部署状态、存储位置以及训练任务等关键信息。
type ModelMetadata struct {
	ID             uint64            // 模型元数据唯一标识符
	ModelName      string            // 模型名称 (例如："product_recommendation", "sentiment_analysis")
	ModelVersion   string            // 模型版本号
	ModelURI       string            // 模型存储路径，例如 S3 路径或本地文件路径
	DeploymentID   string            // 模型部署的唯一ID
	Status         string            // 模型部署状态 (e.g., "DEPLOYED", "FAILED", "UNDEPLOYED", "TRAINING")
	DeployedAt     time.Time         // 模型部署时间
	LastTrainedAt  time.Time         // 模型上次训练完成时间
	TrainingJobIDs []string          // 关联的训练任务ID列表，存储为 JSON 字符串
	Metadata       map[string]string // 其他自定义元数据，存储为 JSON 字符串
	ErrorMessage   string            // 部署或训练失败时的错误信息
	CreatedAt      time.Time         // 记录创建时间
	UpdatedAt      time.Time         // 记录最后更新时间
}

// SentimentType 定义了情感分析的类型。
// 用于表示文本情感的分类结果。
type SentimentType int32

const (
	SentimentTypeUnspecified SentimentType = 0 // 未指定情感类型
	SentimentTypePositive    SentimentType = 1 // 积极情感
	SentimentTypeNeutral     SentimentType = 2 // 中性情感
	SentimentTypeNegative    SentimentType = 3 // 消极情感
)
