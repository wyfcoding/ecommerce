package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/audit/domain"
	"gorm.io/gorm"
)

// AuditLogModel 审计日志写模型。
type AuditLogModel struct {
	gorm.Model
	AuditNo      string                `gorm:"column:audit_no;type:varchar(64);uniqueIndex;not null;comment:审计编号"`
	UserID       uint64                `gorm:"column:user_id;not null;index;comment:用户ID"`
	Username     string                `gorm:"column:username;type:varchar(64);not null;comment:用户名"`
	EventType    domain.AuditEventType `gorm:"column:event_type;type:varchar(32);not null;index;comment:事件类型"`
	Level        domain.AuditLevel     `gorm:"column:level;type:varchar(32);not null;default:'info';comment:级别"`
	Module       string                `gorm:"column:module;type:varchar(64);not null;index;comment:模块"`
	Action       string                `gorm:"column:action;type:varchar(64);not null;comment:操作"`
	ResourceType string                `gorm:"column:resource_type;type:varchar(64);comment:资源类型"`
	ResourceID   string                `gorm:"column:resource_id;type:varchar(64);comment:资源ID"`
	OldValue     string                `gorm:"column:old_value;type:text;comment:旧值"`
	NewValue     string                `gorm:"column:new_value;type:text;comment:新值"`
	IP           string                `gorm:"column:ip;type:varchar(64);comment:IP地址"`
	UserAgent    string                `gorm:"column:user_agent;type:varchar(255);comment:用户代理"`
	Status       string                `gorm:"column:status;type:varchar(32);not null;default:'success';comment:状态"`
	ErrorMsg     string                `gorm:"column:error_msg;type:text;comment:错误信息"`
	Duration     int64                 `gorm:"column:duration;comment:耗时(ms)"`
	Timestamp    time.Time             `gorm:"column:timestamp;not null;index;comment:时间戳"`
}

// AuditPolicyModel 审计策略写模型。
type AuditPolicyModel struct {
	gorm.Model
	Name          string   `gorm:"column:name;type:varchar(128);not null;comment:策略名称"`
	Description   string   `gorm:"column:description;type:text;comment:描述"`
	EventTypes    []string `gorm:"column:event_types;type:json;serializer:json;comment:事件类型列表"`
	Modules       []string `gorm:"column:modules;type:json;serializer:json;comment:模块列表"`
	Enabled       bool     `gorm:"column:enabled;default:true;comment:是否启用"`
	RetentionDays int32    `gorm:"column:retention_days;default:90;comment:保留天数"`
}

// AuditReportModel 审计报告写模型。
type AuditReportModel struct {
	gorm.Model
	ReportNo    string     `gorm:"column:report_no;type:varchar(64);uniqueIndex;not null;comment:报告编号"`
	Title       string     `gorm:"column:title;type:varchar(128);not null;comment:标题"`
	Description string     `gorm:"column:description;type:text;comment:描述"`
	StartDate   time.Time  `gorm:"column:start_date;comment:开始日期"`
	EndDate     time.Time  `gorm:"column:end_date;comment:结束日期"`
	EventTypes  []string   `gorm:"column:event_types;type:json;serializer:json;comment:事件类型列表"`
	Modules     []string   `gorm:"column:modules;type:json;serializer:json;comment:模块列表"`
	Status      string     `gorm:"column:status;type:varchar(32);not null;default:'draft';comment:状态"`
	Content     string     `gorm:"column:content;type:longtext;comment:内容"`
	GeneratedAt *time.Time `gorm:"column:generated_at;comment:生成时间"`
}

func (AuditLogModel) TableName() string    { return "audit_logs" }
func (AuditPolicyModel) TableName() string { return "audit_policies" }
func (AuditReportModel) TableName() string { return "audit_reports" }

func toAuditLogModel(log *domain.AuditLog) *AuditLogModel {
	if log == nil {
		return nil
	}
	return &AuditLogModel{
		Model: gorm.Model{
			ID:        log.ID,
			CreatedAt: log.CreatedAt,
			UpdatedAt: log.UpdatedAt,
		},
		AuditNo:      log.AuditNo,
		UserID:       log.UserID,
		Username:     log.Username,
		EventType:    log.EventType,
		Level:        log.Level,
		Module:       log.Module,
		Action:       log.Action,
		ResourceType: log.ResourceType,
		ResourceID:   log.ResourceID,
		OldValue:     log.OldValue,
		NewValue:     log.NewValue,
		IP:           log.IP,
		UserAgent:    log.UserAgent,
		Status:       log.Status,
		ErrorMsg:     log.ErrorMsg,
		Duration:     log.Duration,
		Timestamp:    log.Timestamp,
	}
}

func toAuditLog(model *AuditLogModel) *domain.AuditLog {
	if model == nil {
		return nil
	}
	return &domain.AuditLog{
		ID:           model.ID,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
		AuditNo:      model.AuditNo,
		UserID:       model.UserID,
		Username:     model.Username,
		EventType:    model.EventType,
		Level:        model.Level,
		Module:       model.Module,
		Action:       model.Action,
		ResourceType: model.ResourceType,
		ResourceID:   model.ResourceID,
		OldValue:     model.OldValue,
		NewValue:     model.NewValue,
		IP:           model.IP,
		UserAgent:    model.UserAgent,
		Status:       model.Status,
		ErrorMsg:     model.ErrorMsg,
		Duration:     model.Duration,
		Timestamp:    model.Timestamp,
	}
}

func toAuditPolicyModel(policy *domain.AuditPolicy) *AuditPolicyModel {
	if policy == nil {
		return nil
	}
	return &AuditPolicyModel{
		Model: gorm.Model{
			ID:        policy.ID,
			CreatedAt: policy.CreatedAt,
			UpdatedAt: policy.UpdatedAt,
		},
		Name:          policy.Name,
		Description:   policy.Description,
		EventTypes:    policy.EventTypes,
		Modules:       policy.Modules,
		Enabled:       policy.Enabled,
		RetentionDays: policy.RetentionDays,
	}
}

func toAuditPolicy(model *AuditPolicyModel) *domain.AuditPolicy {
	if model == nil {
		return nil
	}
	return &domain.AuditPolicy{
		ID:            model.ID,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
		Name:          model.Name,
		Description:   model.Description,
		EventTypes:    model.EventTypes,
		Modules:       model.Modules,
		Enabled:       model.Enabled,
		RetentionDays: model.RetentionDays,
	}
}

func toAuditReportModel(report *domain.AuditReport) *AuditReportModel {
	if report == nil {
		return nil
	}
	return &AuditReportModel{
		Model: gorm.Model{
			ID:        report.ID,
			CreatedAt: report.CreatedAt,
			UpdatedAt: report.UpdatedAt,
		},
		ReportNo:    report.ReportNo,
		Title:       report.Title,
		Description: report.Description,
		StartDate:   report.StartDate,
		EndDate:     report.EndDate,
		EventTypes:  report.EventTypes,
		Modules:     report.Modules,
		Status:      report.Status,
		Content:     report.Content,
		GeneratedAt: report.GeneratedAt,
	}
}

func toAuditReport(model *AuditReportModel) *domain.AuditReport {
	if model == nil {
		return nil
	}
	return &domain.AuditReport{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		ReportNo:    model.ReportNo,
		Title:       model.Title,
		Description: model.Description,
		StartDate:   model.StartDate,
		EndDate:     model.EndDate,
		EventTypes:  model.EventTypes,
		Modules:     model.Modules,
		Status:      model.Status,
		Content:     model.Content,
		GeneratedAt: model.GeneratedAt,
	}
}
