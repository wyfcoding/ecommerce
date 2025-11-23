package entity

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrRouteNotFound     = errors.New("路由不存在")
	ErrRateLimitExceeded = errors.New("超过限流")
)

// RouteStatus 路由状态
type RouteStatus int8

const (
	RouteStatusDisabled RouteStatus = 0 // 禁用
	RouteStatusEnabled  RouteStatus = 1 // 启用
)

// Route 路由实体
type Route struct {
	gorm.Model
	Path        string      `gorm:"type:varchar(255);not null;uniqueIndex;comment:路径" json:"path"`
	Method      string      `gorm:"type:varchar(16);not null;comment:方法" json:"method"`
	Service     string      `gorm:"type:varchar(64);not null;comment:服务名" json:"service"`
	Backend     string      `gorm:"type:varchar(255);not null;comment:后端地址" json:"backend"`
	Timeout     int32       `gorm:"default:5000;comment:超时时间(ms)" json:"timeout"`
	Retries     int32       `gorm:"default:3;comment:重试次数" json:"retries"`
	Status      RouteStatus `gorm:"default:1;comment:状态" json:"status"`
	Description string      `gorm:"type:text;comment:描述" json:"description"`
}

// NewRoute 创建路由
func NewRoute(path, method, service, backend string, timeout, retries int32, description string) *Route {
	return &Route{
		Path:        path,
		Method:      method,
		Service:     service,
		Backend:     backend,
		Timeout:     timeout,
		Retries:     retries,
		Status:      RouteStatusEnabled,
		Description: description,
	}
}

// Enable 启用
func (r *Route) Enable() {
	r.Status = RouteStatusEnabled
}

// Disable 禁用
func (r *Route) Disable() {
	r.Status = RouteStatusDisabled
}

// IsEnabled 是否启用
func (r *Route) IsEnabled() bool {
	return r.Status == RouteStatusEnabled
}

// UpdateBackend 更新后端地址
func (r *Route) UpdateBackend(backend string) {
	r.Backend = backend
}

// RateLimitRule 限流规则实体
type RateLimitRule struct {
	gorm.Model
	Name        string `gorm:"type:varchar(64);not null;uniqueIndex;comment:规则名称" json:"name"`
	Path        string `gorm:"type:varchar(255);not null;comment:路径" json:"path"`
	Method      string `gorm:"type:varchar(16);not null;comment:方法" json:"method"`
	Limit       int32  `gorm:"not null;comment:限制请求数" json:"limit"`
	Window      int32  `gorm:"not null;comment:时间窗口(秒)" json:"window"`
	Enabled     bool   `gorm:"default:true;comment:是否启用" json:"enabled"`
	Description string `gorm:"type:text;comment:描述" json:"description"`
}

// NewRateLimitRule 创建限流规则
func NewRateLimitRule(name, path, method string, limit, window int32, description string) *RateLimitRule {
	return &RateLimitRule{
		Name:        name,
		Path:        path,
		Method:      method,
		Limit:       limit,
		Window:      window,
		Enabled:     true,
		Description: description,
	}
}

// Enable 启用
func (r *RateLimitRule) Enable() {
	r.Enabled = true
}

// Disable 禁用
func (r *RateLimitRule) Disable() {
	r.Enabled = false
}

// UpdateLimit 更新限制
func (r *RateLimitRule) UpdateLimit(limit, window int32) {
	r.Limit = limit
	r.Window = window
}

// APILog API日志实体
type APILog struct {
	gorm.Model
	RequestID  string `gorm:"type:varchar(64);not null;index;comment:请求ID" json:"request_id"`
	Path       string `gorm:"type:varchar(255);not null;comment:路径" json:"path"`
	Method     string `gorm:"type:varchar(16);not null;comment:方法" json:"method"`
	Service    string `gorm:"type:varchar(64);not null;comment:服务名" json:"service"`
	UserID     uint64 `gorm:"index;comment:用户ID" json:"user_id"`
	IP         string `gorm:"type:varchar(64);comment:IP地址" json:"ip"`
	UserAgent  string `gorm:"type:varchar(255);comment:UserAgent" json:"user_agent"`
	StatusCode int32  `gorm:"comment:状态码" json:"status_code"`
	Duration   int64  `gorm:"comment:耗时(ms)" json:"duration"`
	Error      string `gorm:"type:text;comment:错误信息" json:"error"`
}

// NewAPILog 创建API日志
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

// Complete 完成
func (l *APILog) Complete(statusCode int32, duration int64, errorMsg string) {
	l.StatusCode = statusCode
	l.Duration = duration
	l.Error = errorMsg
}
