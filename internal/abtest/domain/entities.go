// Package domain AB测试领域模型
package domain

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"
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

	// 统计指标定义
	MetricName    string // e.g. "click_purchase"
	MinSampleSize int64  // 最小样本量要求
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
	AssignedAt     time.Time
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

	// GetVariationStats 获取变体的统计数据 (样本数, 转化数, 总值)
	GetVariationStats(ctx context.Context, experimentID string) (map[string]*VariationStats, error)
}

// VariationStats 变体统计原始数据
type VariationStats struct {
	VariationKey string
	SampleSize   int64
	Conversions  int64
	TotalValue   float64
}

// VariationResult 统计分析结果 (含显著性)
type VariationResult struct {
	VariationKey   string
	Participants   int64
	Conversions    int64
	ConversionRate float64
	ZScore         float64 // Z检验值 (对比Control)
	PValue         float64 // 显著性P值
	IsSignificant  bool    // 是否显著 (Confidence > 95%)
	Lift           float64 // 提升幅度
}

// Bucketer 负责根据用户 ID 分流.
type Bucketer struct{}

func (b *Bucketer) GetVariation(userID string, exp *Experiment) (string, error) {
	if exp.Status != pb.ExperimentStatus_EXPERIMENT_STATUS_RUNNING {
		return "", fmt.Errorf("experiment is not running")
	}

	// 1. 确定用户是否在总流量百分比内
	hash := b.getHash(userID, exp.ID, "traffic")
	if hash%100 >= uint32(exp.TrafficPercentage) {
		return "", nil // 不在流量范围内
	}

	// 2. 根据权重分配
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
	var val uint32
	fmt.Sscanf(hexStr[:8], "%x", &val)
	return val
}

// Analyzer 统计分析器
type Analyzer struct{}

// AnalyzeResults 计算统计显著性 (CR转化率)
func (a *Analyzer) AnalyzeResults(stats map[string]*VariationStats, controlKey string) ([]*VariationResult, error) {
	control, ok := stats[controlKey]
	if !ok || control.SampleSize == 0 {
		return nil, fmt.Errorf("control group data missing or empty")
	}

	controlCR := float64(control.Conversions) / float64(control.SampleSize)

	var results []*VariationResult

	for key, stat := range stats {
		res := &VariationResult{
			VariationKey:   key,
			Participants:   stat.SampleSize,
			Conversions:    stat.Conversions,
			ConversionRate: 0,
		}

		if stat.SampleSize > 0 {
			res.ConversionRate = float64(stat.Conversions) / float64(stat.SampleSize)
		}

		if key != controlKey && stat.SampleSize > 0 {
			// Calculate Z-Score for two proportions
			p1 := res.ConversionRate
			p2 := controlCR
			n1 := float64(stat.SampleSize)
			n2 := float64(control.SampleSize)

			// Pooled proportion
			p := (float64(stat.Conversions) + float64(control.Conversions)) / (n1 + n2)
			se := math.Sqrt(p * (1 - p) * (1/n1 + 1/n2))

			if se > 0 {
				z := (p1 - p2) / se
				res.ZScore = z
				res.PValue = 2 * (1 - normalCDF(math.Abs(z))) // Two-tailed
				res.IsSignificant = res.PValue < 0.05
				res.Lift = (p1 - p2) / p2
			}
		}

		results = append(results, res)
	}

	return results, nil
}

// normalCDF 标准正态分布累积分布函数 (简化)
func normalCDF(x float64) float64 {
	return 0.5 * (1 + math.Erf(x/math.Sqrt2))
}
