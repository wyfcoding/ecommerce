package entity

import (
	"time"

	"gorm.io/gorm"
)

// AuditEventType 审计事件类型
type AuditEventType string

const (
	EventTypeCreate AuditEventType = "create"
	EventTypeUpdate AuditEventType = "update"
	EventTypeDelete AuditEventType = "delete"
	EventTypeLogin  AuditEventType = "login"
	EventTypeLogout AuditEventType = "logout"
	EventTypeAccess AuditEventType = "access"
	EventTypeExport AuditEventType = "export"
	EventTypeImport AuditEventType = "import"
)

// AuditLevel 审计级别
type AuditLevel string

const (
	LevelInfo     AuditLevel = "info"
	LevelWarning  AuditLevel = "warning"
	LevelError    AuditLevel = "error"
	LevelCritical AuditLevel = "critical"
)

// AuditLog 审计日志聚合根
type AuditLog struct {
	gorm.Model
	AuditNo      string         `gorm:"type:varchar(64);uniqueIndex;not null;comment:审计编号" json:"audit_no"`
	UserID       uint64         `gorm:"not null;index;comment:用户ID" json:"user_id"`
	Username     string         `gorm:"type:varchar(64);not null;comment:用户名" json:"username"`
	EventType    AuditEventType `gorm:"type:varchar(32);not null;index;comment:事件类型" json:"event_type"`
	Level        AuditLevel     `gorm:"type:varchar(32);not null;default:'info';comment:级别" json:"level"`
	Module       string         `gorm:"type:varchar(64);not null;index;comment:模块" json:"module"`
	Action       string         `gorm:"type:varchar(64);not null;comment:操作" json:"action"`
	ResourceType string         `gorm:"type:varchar(64);comment:资源类型" json:"resource_type"`
	ResourceID   string         `gorm:"type:varchar(64);comment:资源ID" json:"resource_id"`
	OldValue     string         `gorm:"type:text;comment:旧值" json:"old_value"`
	NewValue     string         `gorm:"type:text;comment:新值" json:"new_value"`
	IP           string         `gorm:"type:varchar(64);comment:IP地址" json:"ip"`
	UserAgent    string         `gorm:"type:varchar(255);comment:用户代理" json:"user_agent"`
	Status       string         `gorm:"type:varchar(32);not null;default:'success';comment:状态" json:"status"`
	ErrorMsg     string         `gorm:"type:text;comment:错误信息" json:"error_msg"`
	Duration     int64          `gorm:"comment:耗时(ms)" json:"duration"`
	Timestamp    time.Time      `gorm:"not null;index;comment:时间戳" json:"timestamp"`
}

// AuditPolicy 审计策略聚合根
type AuditPolicy struct {
	gorm.Model
	Name          string   `gorm:"type:varchar(128);not null;comment:策略名称" json:"name"`
	Description   string   `gorm:"type:text;comment:描述" json:"description"`
	EventTypes    []string `gorm:"type:json;serializer:json;comment:事件类型列表" json:"event_types"`
	Modules       []string `gorm:"type:json;serializer:json;comment:模块列表" json:"modules"`
	Enabled       bool     `gorm:"default:true;comment:是否启用" json:"enabled"`
	RetentionDays int32    `gorm:"default:90;comment:保留天数" json:"retention_days"`
}

// AuditReport 审计报告聚合根
type AuditReport struct {
	gorm.Model
	ReportNo    string     `gorm:"type:varchar(64);uniqueIndex;not null;comment:报告编号" json:"report_no"`
	Title       string     `gorm:"type:varchar(128);not null;comment:标题" json:"title"`
	Description string     `gorm:"type:text;comment:描述" json:"description"`
	StartDate   time.Time  `gorm:"comment:开始日期" json:"start_date"`
	EndDate     time.Time  `gorm:"comment:结束日期" json:"end_date"`
	EventTypes  []string   `gorm:"type:json;serializer:json;comment:事件类型列表" json:"event_types"`
	Modules     []string   `gorm:"type:json;serializer:json;comment:模块列表" json:"modules"`
	Status      string     `gorm:"type:varchar(32);not null;default:'draft';comment:状态" json:"status"`
	Content     string     `gorm:"type:longtext;comment:内容" json:"content"`
	GeneratedAt *time.Time `gorm:"comment:生成时间" json:"generated_at"`
}

// NewAuditLog 创建审计日志
func NewAuditLog(auditNo string, userID uint64, username string, eventType AuditEventType, module, action string) *AuditLog {
	return &AuditLog{
		AuditNo:   auditNo,
		UserID:    userID,
		Username:  username,
		EventType: eventType,
		Module:    module,
		Action:    action,
		Level:     LevelInfo,
		Status:    "success",
		Timestamp: time.Now(),
	}
}

// SetError 设置错误信息
func (a *AuditLog) SetError(errMsg string) {
	a.Status = "failure"
	a.ErrorMsg = errMsg
	a.Level = LevelError
}

// SetLevel 设置审计级别
func (a *AuditLog) SetLevel(level AuditLevel) {
	a.Level = level
}

// SetResource 设置资源信息
func (a *AuditLog) SetResource(resourceType, resourceID string) {
	a.ResourceType = resourceType
	a.ResourceID = resourceID
}

// SetChange 设置变更信息
func (a *AuditLog) SetChange(oldValue, newValue string) {
	a.OldValue = oldValue
	a.NewValue = newValue
}

// SetClientInfo 设置客户端信息
func (a *AuditLog) SetClientInfo(ip, userAgent string) {
	a.IP = ip
	a.UserAgent = userAgent
}

// SetDuration 设置操作耗时
func (a *AuditLog) SetDuration(duration int64) {
	a.Duration = duration
}

// NewAuditPolicy 创建审计策略
func NewAuditPolicy(name, description string) *AuditPolicy {
	return &AuditPolicy{
		Name:          name,
		Description:   description,
		EventTypes:    []string{},
		Modules:       []string{},
		Enabled:       true,
		RetentionDays: 90,
	}
}

// AddEventType 添加事件类型
func (p *AuditPolicy) AddEventType(eventType string) {
	for _, et := range p.EventTypes {
		if et == eventType {
			return
		}
	}
	p.EventTypes = append(p.EventTypes, eventType)
}

// AddModule 添加模块
func (p *AuditPolicy) AddModule(module string) {
	for _, m := range p.Modules {
		if m == module {
			return
		}
	}
	p.Modules = append(p.Modules, module)
}

// Enable 启用策略
func (p *AuditPolicy) Enable() {
	p.Enabled = true
}

// Disable 禁用策略
func (p *AuditPolicy) Disable() {
	p.Enabled = false
}

// NewAuditReport 创建审计报告
func NewAuditReport(reportNo, title, description string) *AuditReport {
	return &AuditReport{
		ReportNo:    reportNo,
		Title:       title,
		Description: description,
		EventTypes:  []string{},
		Modules:     []string{},
		Status:      "draft",
	}
}

// Generate 生成报告
func (r *AuditReport) Generate(content string) {
	r.Content = content
	r.Status = "generated"
	now := time.Now()
	r.GeneratedAt = &now
}

// Publish 发布报告
func (r *AuditReport) Publish() {
	r.Status = "published"
}
