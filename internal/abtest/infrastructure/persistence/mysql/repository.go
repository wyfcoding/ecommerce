package mysql

import (
	"context"
	"encoding/json"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/abtest/v1"
	"github.com/wyfcoding/ecommerce/internal/abtest/domain"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/logging"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ExperimentModel GORM 模型.
type ExperimentModel struct {
	gorm.Model
	ExperimentID      string     `gorm:"column:experiment_id;uniqueIndex;type:varchar(64);not null"`
	Name              string     `gorm:"column:name;uniqueIndex;type:varchar(255);not null"`
	Description       string     `gorm:"column:description;type:text"`
	TrafficPercentage int32      `gorm:"column:traffic_percentage;default:100"`
	Status            int32      `gorm:"column:status;index"`
	VariationsJSON    string     `gorm:"column:variations_json;type:longtext"`
	StartedAt         *time.Time `gorm:"column:started_at"`
	EndedAt           *time.Time `gorm:"column:ended_at"`
}

func (ExperimentModel) TableName() string { return "ab_experiments" }

// AssignmentModel 记录用户被分配到的实验变量.
type AssignmentModel struct {
	gorm.Model
	UserID       string `gorm:"column:user_id;uniqueIndex:idx_user_exp;type:varchar(64);not null"`
	ExperimentID string `gorm:"column:experiment_id;uniqueIndex:idx_user_exp;type:varchar(64);not null"`
	VariationKey string `gorm:"column:variation_key;type:varchar(64);not null"`
}

func (AssignmentModel) TableName() string { return "ab_assignments" }

// EventRecordModel 记录转化事件.
type EventRecordModel struct {
	gorm.Model
	UserID       string  `gorm:"column:user_id;index;type:varchar(64)"`
	ExperimentID string  `gorm:"column:experiment_id;index;type:varchar(64)"`
	VariationKey string  `gorm:"column:variation_key;index;type:varchar(64)"`
	EventName    string  `gorm:"column:event_name;index;type:varchar(100)"`
	Value        float64 `gorm:"column:event_value;type:decimal(18,2)"`
}

func (EventRecordModel) TableName() string { return "ab_events" }

type abtestRepository struct {
	db     *database.DB
	logger *logging.Logger
}

func NewABTestRepository(db *database.DB, logger *logging.Logger) domain.ABTestRepository {
	return &abtestRepository{db: db, logger: logger}
}

// domain.ABTestRepository 实现
func (r *abtestRepository) SaveExperiment(ctx context.Context, exp *domain.Experiment) error {
	vars, _ := json.Marshal(exp.Variations)
	m := &ExperimentModel{
		ExperimentID:      exp.ID,
		Name:              exp.Name,
		Description:       exp.Description,
		TrafficPercentage: exp.TrafficPercentage,
		Status:            int32(exp.Status),
		VariationsJSON:    string(vars),
		StartedAt:         exp.StartedAt,
		EndedAt:           exp.EndedAt,
	}
	// Upsert Based on ExperimentID
	return r.db.RawDB().WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "experiment_id"}},
		UpdateAll: true,
	}).Save(m).Error
}

func (r *abtestRepository) GetExperimentByID(ctx context.Context, id string) (*domain.Experiment, error) {
	var m ExperimentModel
	if err := r.db.RawDB().WithContext(ctx).Where("experiment_id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return toExperimentDomain(&m), nil
}

func (r *abtestRepository) GetExperimentByName(ctx context.Context, name string) (*domain.Experiment, error) {
	var m ExperimentModel
	if err := r.db.RawDB().WithContext(ctx).Where("name = ?", name).First(&m).Error; err != nil {
		return nil, err
	}
	return toExperimentDomain(&m), nil
}

func (r *abtestRepository) ListExperiments(ctx context.Context, status pb.ExperimentStatus, offset, limit int) ([]*domain.Experiment, int, error) {
	var models []*ExperimentModel
	var total int64
	db := r.db.RawDB().WithContext(ctx).Model(&ExperimentModel{})
	if status != pb.ExperimentStatus_EXPERIMENT_STATUS_UNSPECIFIED {
		db = db.Where("status = ?", int32(status))
	}
	db.Count(&total)
	if err := db.Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	res := make([]*domain.Experiment, len(models))
	for i, m := range models {
		res[i] = toExperimentDomain(m)
	}
	return res, int(total), nil
}

func (r *abtestRepository) SaveAssignment(ctx context.Context, a *domain.Assignment) error {
	m := &AssignmentModel{
		UserID:       a.UserID,
		ExperimentID: a.ExperimentID,
		VariationKey: a.VariationKey,
	}
	return r.db.RawDB().WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "experiment_id"}},
		UpdateAll: true,
	}).Create(m).Error
}

func (r *abtestRepository) GetAssignment(ctx context.Context, userID, experimentID string) (*domain.Assignment, error) {
	var m AssignmentModel
	if err := r.db.RawDB().WithContext(ctx).Where("user_id = ? AND experiment_id = ?", userID, experimentID).First(&m).Error; err != nil {
		return nil, err
	}
	return &domain.Assignment{
		UserID:       m.UserID,
		ExperimentID: m.ExperimentID,
		VariationKey: m.VariationKey,
	}, nil
}

func (r *abtestRepository) TrackEvent(ctx context.Context, experimentID, variationKey, eventName string, value float64) error {
	m := &EventRecordModel{
		ExperimentID: experimentID,
		VariationKey: variationKey,
		EventName:    eventName,
		Value:        value,
	}
	return r.db.RawDB().WithContext(ctx).Create(m).Error
}

func (r *abtestRepository) GetVariationStats(ctx context.Context, experimentID string) (map[string]*domain.VariationStats, error) {
	// 统计参与人数
	var partCount []struct {
		VariationKey string
		Count        int64
	}
	r.db.RawDB().WithContext(ctx).Model(&AssignmentModel{}).
		Select("variation_key, count(*) as count").
		Where("experiment_id = ?", experimentID).
		Group("variation_key").Scan(&partCount)

	// 统计转化次数
	var convCount []struct {
		VariationKey string
		Count        int64
	}
	r.db.RawDB().WithContext(ctx).Model(&EventRecordModel{}).
		Select("variation_key, count(*) as count").
		Where("experiment_id = ?", experimentID).
		Group("variation_key").Scan(&convCount)

	// 合并结果
	stats := make(map[string]*domain.VariationStats)
	for _, p := range partCount {
		if _, ok := stats[p.VariationKey]; !ok {
			stats[p.VariationKey] = &domain.VariationStats{VariationKey: p.VariationKey}
		}
		stats[p.VariationKey].SampleSize = p.Count
	}
	for _, c := range convCount {
		if _, ok := stats[c.VariationKey]; !ok {
			stats[c.VariationKey] = &domain.VariationStats{VariationKey: c.VariationKey}
		}
		stats[c.VariationKey].Conversions = c.Count
	}

	return stats, nil
}

func toExperimentDomain(m *ExperimentModel) *domain.Experiment {
	var variations []domain.Variation
	_ = json.Unmarshal([]byte(m.VariationsJSON), &variations)
	return &domain.Experiment{
		ID:                m.ExperimentID,
		Name:              m.Name,
		Description:       m.Description,
		Variations:        variations,
		TrafficPercentage: m.TrafficPercentage,
		Status:            pb.ExperimentStatus(m.Status),
		CreatedAt:         m.CreatedAt,
		StartedAt:         m.StartedAt,
		EndedAt:           m.EndedAt,
	}
}
