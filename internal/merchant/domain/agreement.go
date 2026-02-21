// 生成摘要：
// - 从 agreement 服务合并到 merchant 域。
// - 协议签署为商户入驻域子聚合，负责管理商户合同/用户隐私协议。
// - 关键实体：Agreement（协议模板）、UserAgreementRecord（用户签署记录）。
package domain

import (
	"time"

	"gorm.io/gorm"
)

// Agreement 协议版本模板。
// 管理入驻商户的电子合同以及用户的隐私/服务协议更新。
type Agreement struct {
	gorm.Model
	// Code 协议编码，全局唯一标识。
	Code string `gorm:"type:varchar(64);uniqueIndex" json:"code"`
	// Version 协议版本号。
	Version string `gorm:"type:varchar(32)" json:"version"`
	// Title 协议标题。
	Title string `gorm:"type:varchar(128)" json:"title"`
	// Content 协议正文。
	Content string `gorm:"type:text" json:"content"`
	// IsForce 是否强制签署，true 表示用户必须同意才能继续。
	IsForce bool `gorm:"default:false;comment:是否强制签署" json:"is_force"`
}

// UserAgreementRecord 用户签署存证。
// 记录用户签署某协议的时间、IP 等法律凭证信息。
type UserAgreementRecord struct {
	gorm.Model
	// UserID 签署用户 ID。
	UserID uint64 `gorm:"index" json:"user_id"`
	// AgreementID 关联的协议 ID。
	AgreementID uint `gorm:"index" json:"agreement_id"`
	// SignedAt 签署时间。
	SignedAt time.Time `json:"signed_at"`
	// ClientIP 签署时客户端 IP。
	ClientIP string `gorm:"type:varchar(45)" json:"client_ip"`
}
