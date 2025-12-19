package domain

import (
	"gorm.io/gorm"
)

// AuditLog 审计日志
// 记录所有管理端的操作行为，不可变
type AuditLog struct {
	gorm.Model
	UserID    uint   `gorm:"column:user_id;index;not null;comment:操作人ID"`
	Username  string `gorm:"column:username;type:varchar(50);not null;comment:操作人用户名(冗余)"`
	Action    string `gorm:"column:action;type:varchar(50);index;not null;comment:操作动作"`
	Resource  string `gorm:"column:resource;type:varchar(50);index;not null;comment:资源类型"`
	TargetID  string `gorm:"column:target_id;type:varchar(50);index;comment:目标资源ID"`
	ClientIP  string `gorm:"column:client_ip;type:varchar(50);comment:客户端IP"`
	Payload   string `gorm:"column:payload;type:text;comment:请求参数(JSON)"`
	Result    string `gorm:"column:result;type:text;comment:操作结果/错误信息"`
	Status    int    `gorm:"column:status;type:tinyint;default:1;comment:结果状态 1:成功 0:失败"`
	UserAgent string `gorm:"column:user_agent;type:varchar(255);comment:UserAgent"`
}
