package domain

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/abtest/v1"
)

// Experiment 聚合根，代表一个 AB 实验.
type Experiment struct {
	ID                string
	Name              string
	Description       string
	Variations        []Variation
	TrafficPercentage int32 // 0-100
	Status            pb.ExperimentStatus
	CreatedAt         time.Time
	StartedAt         *time.Time
	EndedAt           *time.Time
}

// Variation 实验变量.
type Variation struct {
	Key    string
	Value  string // JSON payload
	Weight int32
}

// Assignment 用户分配结果.
type Assignment struct {
	UserID         string
	ExperimentID   string
	VariationKey   string
	VariationValue string
}

// ABTestRepository 仓储接口.
type ABTestRepository interface {
	SaveExperiment(ctx context.Context, exp *Experiment) error
	GetExperimentByID(ctx context.Context, id string) (*Experiment, error)
	GetExperimentByName(ctx context.Context, name string) (*Experiment, error)
	ListExperiments(ctx context.Context, status pb.ExperimentStatus, offset, limit int) ([]*Experiment, int, error)

	SaveAssignment(ctx context.Context, assignment *Assignment) error
	GetAssignment(ctx context.Context, userID, experimentID string) (*Assignment, error)

	TrackEvent(ctx context.Context, experimentID, variationKey, eventName string, value float64) error
	GetResults(ctx context.Context, experimentID string) ([]*VariationResult, error)
}

// VariationResult 统计结果.
type VariationResult struct {
	VariationKey   string
	Participants   int64
	Conversions    int64
	ConversionRate float64
}

// Bucketer 负责根据用户 ID 分流.
type Bucketer struct{}

func (b *Bucketer) GetVariation(userID string, exp *Experiment) (string, error) {
	if exp.Status != pb.ExperimentStatus_EXPERIMENT_STATUS_RUNNING {
		return "", fmt.Errorf("experiment is not running")
	}

	// 1. 确定用户是否在总流量百分比内
	// 使用 MD5 结合 Experiment ID 实现确定性分流
	hash := b.getHash(userID, exp.ID, "traffic")
	if hash%100 >= uint32(exp.TrafficPercentage) {
		return "", nil // 不在流量范围内，返回默认值（空）
	}

	// 2. 在变化的权重之间分配
	totalWeight := int32(0)
	for _, v := range exp.Variations {
		totalWeight += v.Weight
	}
	if totalWeight == 0 {
		return "", fmt.Errorf("no variations with weight found")
	}

	hash = b.getHash(userID, exp.ID, "bucket")
	bucket := int32(hash % uint32(totalWeight))

	current := int32(0)
	for _, v := range exp.Variations {
		current += v.Weight
		if bucket < current {
			return v.Key, nil
		}
	}

	return exp.Variations[0].Key, nil
}

func (b *Bucketer) getHash(userID, experimentID, salt string) uint32 {
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%s:%s:%s", userID, experimentID, salt)))
	hexStr := hex.EncodeToString(h.Sum(nil))
	// 取前 8 位转换为 uint32
	var val uint32
	fmt.Sscanf(hexStr[:8], "%x", &val)
	return val
}
