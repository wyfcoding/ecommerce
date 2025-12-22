package domain

import (
	"slices"
	"time"

	"gorm.io/gorm" // 导入GORM库。
)

// AuditEventType 定义了审计事件的类型。
// 这些类型用于对系统中发生的各种操作进行分类。
type AuditEventType string

const (
	EventTypeCreate AuditEventType = "create" // 创建资源事件。
	EventTypeUpdate AuditEventType = "update" // 更新资源事件。
	EventTypeDelete AuditEventType = "delete" // 删除资源事件。
	EventTypeLogin  AuditEventType = "login"  // 用户登录事件。
	EventTypeLogout AuditEventType = "logout" // 用户登出事件。
	EventTypeAccess AuditEventType = "access" // 资源访问事件。
	EventTypeExport AuditEventType = "export" // 数据导出事件。
	EventTypeImport AuditEventType = "import" // 数据导入事件。
)

// AuditLevel 定义了审计日志的级别。
// 这些级别用于表示事件的重要性和严重程度。
type AuditLevel string

const (
	LevelInfo     AuditLevel = "info"     // 信息级别，记录一般性操作。
	LevelWarning  AuditLevel = "warning"  // 警告级别，记录可能存在风险的操作。
	LevelError    AuditLevel = "error"    // 错误级别，记录操作失败或异常。
	LevelCritical AuditLevel = "critical" // 关键级别，记录严重的安全事件或系统故障。
)

// AuditLog 实体是审计模块的聚合根。
// 它代表一个完整的审计日志记录，包含了操作用户、事件类型、模块、操作详情、状态、时间戳等信息。
type AuditLog struct {
	gorm.Model                  // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	AuditNo      string         `gorm:"type:varchar(64);uniqueIndex;not null;comment:审计编号" json:"audit_no"`   // 审计事件的唯一编号，唯一索引，不允许为空。
	UserID       uint64         `gorm:"not null;index;comment:用户ID" json:"user_id"`                           // 执行操作的用户ID，索引字段。
	Username     string         `gorm:"type:varchar(64);not null;comment:用户名" json:"username"`                // 执行操作的用户名。
	EventType    AuditEventType `gorm:"type:varchar(32);not null;index;comment:事件类型" json:"event_type"`       // 审计事件类型，索引字段。
	Level        AuditLevel     `gorm:"type:varchar(32);not null;default:'info';comment:级别" json:"level"`     // 审计事件的级别，默认为信息级别。
	Module       string         `gorm:"type:varchar(64);not null;index;comment:模块" json:"module"`             // 事件所属的模块，索引字段。
	Action       string         `gorm:"type:varchar(64);not null;comment:操作" json:"action"`                   // 执行的具体操作。
	ResourceType string         `gorm:"type:varchar(64);comment:资源类型" json:"resource_type"`                   // 操作影响的资源类型。
	ResourceID   string         `gorm:"type:varchar(64);comment:资源ID" json:"resource_id"`                     // 操作影响的资源ID。
	OldValue     string         `gorm:"type:text;comment:旧值" json:"old_value"`                                // 资源变更前的旧值（通常为JSON字符串）。
	NewValue     string         `gorm:"type:text;comment:新值" json:"new_value"`                                // 资源变更后的新值（通常为JSON字符串）。
	IP           string         `gorm:"type:varchar(64);comment:IP地址" json:"ip"`                              // 操作发生时的IP地址。
	UserAgent    string         `gorm:"type:varchar(255);comment:用户代理" json:"user_agent"`                     // 操作发生时的User-Agent字符串。
	Status       string         `gorm:"type:varchar(32);not null;default:'success';comment:状态" json:"status"` // 操作的状态，例如“success”（成功）或“failure”（失败）。
	ErrorMsg     string         `gorm:"type:text;comment:错误信息" json:"error_msg"`                              // 错误信息（如果操作失败）。
	Duration     int64          `gorm:"comment:耗时(ms)" json:"duration"`                                       // 操作的总耗时（毫秒）。
	Timestamp    time.Time      `gorm:"not null;index;comment:时间戳" json:"timestamp"`                          // 事件发生的时间戳，索引字段。
}

// AuditPolicy 实体是审计模块的聚合根，定义了审计日志的收集和保留策略。
type AuditPolicy struct {
	gorm.Model             // 嵌入gorm.Model。
	Name          string   `gorm:"type:varchar(128);not null;comment:策略名称" json:"name"`         // 策略名称。
	Description   string   `gorm:"type:text;comment:描述" json:"description"`                     // 策略描述。
	EventTypes    []string `gorm:"type:json;serializer:json;comment:事件类型列表" json:"event_types"` // 策略关注的事件类型列表（JSON存储）。
	Modules       []string `gorm:"type:json;serializer:json;comment:模块列表" json:"modules"`       // 策略关注的模块列表（JSON存储）。
	Enabled       bool     `gorm:"default:true;comment:是否启用" json:"enabled"`                    // 策略是否启用，默认为启用。
	RetentionDays int32    `gorm:"default:90;comment:保留天数" json:"retention_days"`               // 审计日志的保留天数，过期将被删除，默认为90天。
}

// AuditReport 实体是审计模块的聚合根，代表一个生成的审计报告。
type AuditReport struct {
	gorm.Model             // 嵌入gorm.Model。
	ReportNo    string     `gorm:"type:varchar(64);uniqueIndex;not null;comment:报告编号" json:"report_no"` // 报告的唯一编号，唯一索引，不允许为空。
	Title       string     `gorm:"type:varchar(128);not null;comment:标题" json:"title"`                  // 报告标题。
	Description string     `gorm:"type:text;comment:描述" json:"description"`                             // 报告描述。
	StartDate   time.Time  `gorm:"comment:开始日期" json:"start_date"`                                      // 报告涵盖的开始日期。
	EndDate     time.Time  `gorm:"comment:结束日期" json:"end_date"`                                        // 报告涵盖的结束日期。
	EventTypes  []string   `gorm:"type:json;serializer:json;comment:事件类型列表" json:"event_types"`         // 报告包含的事件类型列表（JSON存储）。
	Modules     []string   `gorm:"type:json;serializer:json;comment:模块列表" json:"modules"`               // 报告包含的模块列表（JSON存储）。
	Status      string     `gorm:"type:varchar(32);not null;default:'draft';comment:状态" json:"status"`  // 报告状态，例如“draft”（草稿）、“generated”（已生成）、“published”（已发布）。
	Content     string     `gorm:"type:longtext;comment:内容" json:"content"`                             // 报告的详细内容，可以存储为文本或HTML。
	GeneratedAt *time.Time `gorm:"comment:生成时间" json:"generated_at"`                                    // 报告生成时间。
}

// NewAuditLog 创建并返回一个新的 AuditLog 实体实例。
// auditNo: 审计日志的唯一编号。
// userID, username: 执行操作的用户信息。
// eventType, module, action: 事件的分类信息。
func NewAuditLog(auditNo string, userID uint64, username string, eventType AuditEventType, module, action string) *AuditLog {
	return &AuditLog{
		AuditNo:   auditNo,
		UserID:    userID,
		Username:  username,
		EventType: eventType,
		Module:    module,
		Action:    action,
		Level:     LevelInfo,  // 默认为信息级别。
		Status:    "success",  // 默认状态为成功。
		Timestamp: time.Now(), // 记录当前时间作为事件时间戳。
	}
}

// SetError 设置审计日志的错误信息。
// errMsg: 错误消息。
func (a *AuditLog) SetError(errMsg string) {
	a.Status = "failure" // 将状态设置为失败。
	a.ErrorMsg = errMsg  // 记录错误消息。
	a.Level = LevelError // 将级别设置为错误。
}

// SetLevel 设置审计日志的级别。
func (a *AuditLog) SetLevel(level AuditLevel) {
	a.Level = level
}

// SetResource 设置审计日志所关联的资源信息。
// resourceType: 资源类型，例如“Order”，“Product”。
// resourceID: 资源ID。
func (a *AuditLog) SetResource(resourceType, resourceID string) {
	a.ResourceType = resourceType
	a.ResourceID = resourceID
}

// SetChange 设置审计日志中资源变更的前后值。
// oldValue: 变更前的资源状态（例如，JSON字符串）。
// newValue: 变更后的资源状态（例如，JSON字符串）。
func (a *AuditLog) SetChange(oldValue, newValue string) {
	a.OldValue = oldValue
	a.NewValue = newValue
}

// SetClientInfo 设置审计日志中客户端的信息。
// ip: 客户端的IP地址。
// userAgent: 客户端的User-Agent字符串。
func (a *AuditLog) SetClientInfo(ip, userAgent string) {
	a.IP = ip
	a.UserAgent = userAgent
}

// SetDuration 设置审计日志中操作的耗时。
// duration: 操作的耗时（毫秒）。
func (a *AuditLog) SetDuration(duration int64) {
	a.Duration = duration
}

// NewAuditPolicy 创建并返回一个新的 AuditPolicy 实体实例。
// name: 策略名称。
// description: 策略描述。
func NewAuditPolicy(name, description string) *AuditPolicy {
	return &AuditPolicy{
		Name:          name,
		Description:   description,
		EventTypes:    []string{}, // 初始化关注的事件类型列表。
		Modules:       []string{}, // 初始化关注的模块列表。
		Enabled:       true,       // 默认启用策略。
		RetentionDays: 90,         // 默认保留90天。
	}
}

// AddEventType 向审计策略中添加一个关注的事件类型。
func (p *AuditPolicy) AddEventType(eventType string) {
	if slices.Contains(p.EventTypes, eventType) {
		return // 避免重复添加。
	}
	p.EventTypes = append(p.EventTypes, eventType)
}

// AddModule 向审计策略中添加一个关注的模块。
func (p *AuditPolicy) AddModule(module string) {
	if slices.Contains(p.Modules, module) {
		return // 避免重复添加。
	}
	p.Modules = append(p.Modules, module)
}

// Enable 启用审计策略。
func (p *AuditPolicy) Enable() {
	p.Enabled = true
}

// Disable 禁用审计策略。
func (p *AuditPolicy) Disable() {
	p.Enabled = false
}

// NewAuditReport 创建并返回一个新的 AuditReport 实体实例。
// reportNo: 报告的唯一编号。
// title, description: 报告标题和描述。
func NewAuditReport(reportNo, title, description string) *AuditReport {
	return &AuditReport{
		ReportNo:    reportNo,
		Title:       title,
		Description: description,
		EventTypes:  []string{}, // 初始化报告包含的事件类型列表。
		Modules:     []string{}, // 初始化报告包含的模块列表。
		Status:      "draft",    // 默认状态为草稿。
	}
}

// Generate 生成报告内容，并更新报告状态和生成时间。
// content: 生成的报告的详细内容。
func (r *AuditReport) Generate(content string) {
	r.Content = content    // 存储报告内容。
	r.Status = "generated" // 更新状态为“已生成”。
	now := time.Now()
	r.GeneratedAt = &now // 记录生成时间。
}

// Publish 发布报告，更新报告状态为“published”。
func (r *AuditReport) Publish() {
	r.Status = "published"
}
