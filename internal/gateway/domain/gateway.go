package domain

import (
	"errors" // 导入标准错误处理库。

	"gorm.io/gorm" // 导入GORM库。
)

// 定义Gateway模块的业务错误。
var (
	ErrRouteNotFound     = errors.New("路由不存在") // 路由规则未找到。
	ErrRateLimitExceeded = errors.New("超过限流")  // 请求超过限流阈值。
)

// RouteStatus 定义了路由规则的启用状态。
type RouteStatus int8

const (
	RouteStatusDisabled RouteStatus = 0 // 禁用：路由不生效。
	RouteStatusEnabled  RouteStatus = 1 // 启用：路由生效。
)

// Route 实体代表一个API网关的路由规则。
// 它定义了请求如何从网关转发到后端服务。
type Route struct {
	gorm.Model              // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	Path        string      `gorm:"type:varchar(255);not null;uniqueIndex;comment:路径" json:"path"` // 匹配的请求路径，唯一索引，不允许为空。
	Method      string      `gorm:"type:varchar(16);not null;comment:方法" json:"method"`            // 匹配的HTTP方法，例如“GET”，“POST”。
	Service     string      `gorm:"type:varchar(64);not null;comment:服务名" json:"service"`          // 目标后端服务名称。
	Backend     string      `gorm:"type:varchar(255);not null;comment:后端地址" json:"backend"`        // 目标后端服务的具体地址。
	Timeout     int32       `gorm:"default:5000;comment:超时时间(ms)" json:"timeout"`                  // 请求转发到后端服务的超时时间（毫秒）。
	Retries     int32       `gorm:"default:3;comment:重试次数" json:"retries"`                         // 请求转发失败后的重试次数。
	Status      RouteStatus `gorm:"default:1;comment:状态" json:"status"`                            // 路由状态，默认为启用。
	Description string      `gorm:"type:text;comment:描述" json:"description"`                       // 路由规则的描述。
}

// NewRoute 创建并返回一个新的 Route 实体实例。
// path: 路径。
// method: HTTP方法。
// service: 服务名。
// backend: 后端地址。
// timeout, retries: 超时和重试次数。
// description: 描述。
func NewRoute(path, method, service, backend string, timeout, retries int32, description string) *Route {
	return &Route{
		Path:        path,
		Method:      method,
		Service:     service,
		Backend:     backend,
		Timeout:     timeout,
		Retries:     retries,
		Status:      RouteStatusEnabled, // 默认状态为启用。
		Description: description,
	}
}

// Enable 启用路由规则。
func (r *Route) Enable() {
	r.Status = RouteStatusEnabled
}

// Disable 禁用路由规则。
func (r *Route) Disable() {
	r.Status = RouteStatusDisabled
}

// IsEnabled 检查路由规则是否已启用。
func (r *Route) IsEnabled() bool {
	return r.Status == RouteStatusEnabled
}

// UpdateBackend 更新路由规则的后端地址。
func (r *Route) UpdateBackend(backend string) {
	r.Backend = backend
}

// RateLimitRule 实体代表一个API限流规则。
// 它定义了在特定路径和方法上允许的最大请求速率。
type RateLimitRule struct {
	gorm.Model         // 嵌入gorm.Model。
	Name        string `gorm:"type:varchar(64);not null;uniqueIndex;comment:规则名称" json:"name"` // 规则名称，唯一索引，不允许为空。
	Path        string `gorm:"type:varchar(255);not null;comment:路径" json:"path"`              // 匹配的请求路径。
	Method      string `gorm:"type:varchar(16);not null;comment:方法" json:"method"`             // 匹配的HTTP方法。
	Limit       int32  `gorm:"not null;comment:限制请求数" json:"limit"`                            // 在指定时间窗口内允许的最大请求数。
	Window      int32  `gorm:"not null;comment:时间窗口(秒)" json:"window"`                         // 限流的时间窗口长度（秒）。
	Enabled     bool   `gorm:"default:true;comment:是否启用" json:"enabled"`                       // 规则是否启用，默认为启用。
	Description string `gorm:"type:text;comment:描述" json:"description"`                        // 规则描述。
}

// NewRateLimitRule 创建并返回一个新的 RateLimitRule 实体实例。
// name: 规则名称。
// path: 路径。
// method: HTTP方法。
// limit: 限制请求数。
// window: 时间窗口。
// description: 描述。
func NewRateLimitRule(name, path, method string, limit, window int32, description string) *RateLimitRule {
	return &RateLimitRule{
		Name:        name,
		Path:        path,
		Method:      method,
		Limit:       limit,
		Window:      window,
		Enabled:     true, // 默认启用。
		Description: description,
	}
}

// Enable 启用限流规则。
func (r *RateLimitRule) Enable() {
	r.Enabled = true
}

// Disable 禁用限流规则。
func (r *RateLimitRule) Disable() {
	r.Enabled = false
}

// UpdateLimit 更新限流规则的限制请求数和时间窗口。
func (r *RateLimitRule) UpdateLimit(limit, window int32) {
	r.Limit = limit
	r.Window = window
}

// APILog 实体代表一条API请求日志。
// 它记录了通过网关的API请求的详细信息。
type APILog struct {
	gorm.Model        // 嵌入gorm.Model。
	RequestID  string `gorm:"type:varchar(64);not null;index;comment:请求ID" json:"request_id"` // 请求的唯一标识符，索引字段。
	Path       string `gorm:"type:varchar(255);not null;comment:路径" json:"path"`              // 请求路径。
	Method     string `gorm:"type:varchar(16);not null;comment:方法" json:"method"`             // 请求方法。
	Service    string `gorm:"type:varchar(64);not null;comment:服务名" json:"service"`           // 目标服务名称。
	UserID     uint64 `gorm:"index;comment:用户ID" json:"user_id"`                              // 发起请求的用户ID，索引字段。
	IP         string `gorm:"type:varchar(64);comment:IP地址" json:"ip"`                        // 客户端IP地址。
	UserAgent  string `gorm:"type:varchar(255);comment:UserAgent" json:"user_agent"`          // 用户代理字符串。
	StatusCode int32  `gorm:"comment:状态码" json:"status_code"`                                 // HTTP响应状态码。
	Duration   int64  `gorm:"comment:耗时(ms)" json:"duration"`                                 // 请求处理的总耗时（毫秒）。
	Error      string `gorm:"type:text;comment:错误信息" json:"error"`                            // 错误信息（如果请求失败）。
}

// NewAPILog 创建并返回一个新的 APILog 实体实例。
// requestID: 请求ID。
// path, method, service: 请求路径、方法和目标服务。
// userID: 用户ID。
// ip, userAgent: 客户端IP和User-Agent。
func NewAPILog(requestID, path, method, service string, userID uint64, ip, userAgent string) *APILog {
	return &APILog{
		RequestID: requestID,
		Path:      path,
		Method:    method,
		Service:   service,
		UserID:    userID,
		IP:        ip,
		UserAgent: userAgent,
	}
}

// Complete 标记API请求日志为完成，记录响应状态码、耗时和错误信息。
func (l *APILog) Complete(statusCode int32, duration int64, errorMsg string) {
	l.StatusCode = statusCode
	l.Duration = duration
	l.Error = errorMsg
}
