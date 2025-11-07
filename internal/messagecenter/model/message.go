package model

import "time"

// MessageType 消息类型
type MessageType string

const (
	MessageTypeSystem      MessageType = "SYSTEM"      // 系统消息
	MessageTypeOrder       MessageType = "ORDER"       // 订单消息
	MessageTypePromotion   MessageType = "PROMOTION"   // 促销消息
	MessageTypeActivity    MessageType = "ACTIVITY"    // 活动消息
	MessageTypeNotice      MessageType = "NOTICE"      // 公告消息
	MessageTypeInteraction MessageType = "INTERACTION" // 互动消息（点赞、评论等）
)

// MessageStatus 消息状态
type MessageStatus string

const (
	MessageStatusUnread MessageStatus = "UNREAD" // 未读
	MessageStatusRead   MessageStatus = "READ"   // 已读
	MessageStatusDeleted MessageStatus = "DELETED" // 已删除
)

// MessagePriority 消息优先级
type MessagePriority string

const (
	MessagePriorityLow    MessagePriority = "LOW"    // 低
	MessagePriorityNormal MessagePriority = "NORMAL" // 普通
	MessagePriorityHigh   MessagePriority = "HIGH"   // 高
	MessagePriorityUrgent MessagePriority = "URGENT" // 紧急
)

// Message 消息
type Message struct {
	ID          uint64          `gorm:"primarykey" json:"id"`
	MessageNo   string          `gorm:"type:varchar(64);uniqueIndex;not null;comment:消息编号" json:"messageNo"`
	Type        MessageType     `gorm:"type:varchar(20);not null;index;comment:消息类型" json:"type"`
	Priority    MessagePriority `gorm:"type:varchar(20);not null;comment:优先级" json:"priority"`
	Title       string          `gorm:"type:varchar(255);not null;comment:消息标题" json:"title"`
	Content     string          `gorm:"type:text;not null;comment:消息内容" json:"content"`
	Summary     string          `gorm:"type:varchar(500);comment:消息摘要" json:"summary"`
	ImageURL    string          `gorm:"type:varchar(255);comment:消息图片" json:"imageUrl"`
	LinkURL     string          `gorm:"type:varchar(500);comment:跳转链接" json:"linkUrl"`
	LinkType    string          `gorm:"type:varchar(50);comment:链接类型" json:"linkType"`
	SenderID    uint64          `gorm:"comment:发送者ID,0表示系统" json:"senderId"`
	SenderName  string          `gorm:"type:varchar(100);comment:发送者名称" json:"senderName"`
	TargetType  string          `gorm:"type:varchar(20);not null;comment:目标类型(ALL,USER,GROUP)" json:"targetType"`
	TargetIDs   string          `gorm:"type:text;comment:目标用户ID列表JSON" json:"targetIds"`
	Tags        string          `gorm:"type:varchar(500);comment:标签,逗号分隔" json:"tags"`
	ExpireAt    *time.Time      `gorm:"comment:过期时间" json:"expireAt"`
	PublishAt   time.Time       `gorm:"not null;comment:发布时间" json:"publishAt"`
	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   *time.Time      `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (Message) TableName() string {
	return "messages"
}

// UserMessage 用户消息关联
type UserMessage struct {
	ID        uint64        `gorm:"primarykey" json:"id"`
	UserID    uint64        `gorm:"index:idx_user_status;not null;comment:用户ID" json:"userId"`
	MessageID uint64        `gorm:"index;not null;comment:消息ID" json:"messageId"`
	Status    MessageStatus `gorm:"type:varchar(20);not null;index:idx_user_status;comment:消息状态" json:"status"`
	ReadAt    *time.Time    `gorm:"comment:阅读时间" json:"readAt"`
	CreatedAt time.Time     `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time     `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt *time.Time    `gorm:"index" json:"deletedAt,omitempty"`

	// 关联消息详情
	Message *Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
}

// TableName 指定表名
func (UserMessage) TableName() string {
	return "user_messages"
}

// MessageTemplate 消息模板
type MessageTemplate struct {
	ID          uint64      `gorm:"primarykey" json:"id"`
	Code        string      `gorm:"type:varchar(100);uniqueIndex;not null;comment:模板编码" json:"code"`
	Name        string      `gorm:"type:varchar(255);not null;comment:模板名称" json:"name"`
	Type        MessageType `gorm:"type:varchar(20);not null;comment:消息类型" json:"type"`
	Title       string      `gorm:"type:varchar(255);not null;comment:标题模板" json:"title"`
	Content     string      `gorm:"type:text;not null;comment:内容模板" json:"content"`
	Variables   string      `gorm:"type:text;comment:变量说明JSON" json:"variables"`
	IsActive    bool        `gorm:"not null;default:true;comment:是否激活" json:"isActive"`
	Description string      `gorm:"type:varchar(500);comment:模板描述" json:"description"`
	CreatedAt   time.Time   `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   *time.Time  `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (MessageTemplate) TableName() string {
	return "message_templates"
}

// MessageConfig 消息配置
type MessageConfig struct {
	ID        uint64     `gorm:"primarykey" json:"id"`
	UserID    uint64     `gorm:"uniqueIndex;not null;comment:用户ID" json:"userId"`
	// 消息类型开关
	SystemEnabled      bool      `gorm:"not null;default:true;comment:系统消息" json:"systemEnabled"`
	OrderEnabled       bool      `gorm:"not null;default:true;comment:订单消息" json:"orderEnabled"`
	PromotionEnabled   bool      `gorm:"not null;default:true;comment:促销消息" json:"promotionEnabled"`
	ActivityEnabled    bool      `gorm:"not null;default:true;comment:活动消息" json:"activityEnabled"`
	NoticeEnabled      bool      `gorm:"not null;default:true;comment:公告消息" json:"noticeEnabled"`
	InteractionEnabled bool      `gorm:"not null;default:true;comment:互动消息" json:"interactionEnabled"`
	// 推送设置
	PushEnabled  bool      `gorm:"not null;default:true;comment:推送通知" json:"pushEnabled"`
	EmailEnabled bool      `gorm:"not null;default:false;comment:邮件通知" json:"emailEnabled"`
	SMSEnabled   bool      `gorm:"not null;default:false;comment:短信通知" json:"smsEnabled"`
	// 免打扰时段
	DoNotDisturbStart string    `gorm:"type:varchar(10);comment:免打扰开始时间(HH:mm)" json:"doNotDisturbStart"`
	DoNotDisturbEnd   string    `gorm:"type:varchar(10);comment:免打扰结束时间(HH:mm)" json:"doNotDisturbEnd"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (MessageConfig) TableName() string {
	return "message_configs"
}

// MessageStatistics 消息统计
type MessageStatistics struct {
	TotalCount  int64 `json:"totalCount"`  // 总消息数
	UnreadCount int64 `json:"unreadCount"` // 未读消息数
	ReadCount   int64 `json:"readCount"`   // 已读消息数
	
	// 按类型统计
	SystemCount      int64 `json:"systemCount"`
	OrderCount       int64 `json:"orderCount"`
	PromotionCount   int64 `json:"promotionCount"`
	ActivityCount    int64 `json:"activityCount"`
	NoticeCount      int64 `json:"noticeCount"`
	InteractionCount int64 `json:"interactionCount"`
}

// IsExpired 判断消息是否过期
func (m *Message) IsExpired() bool {
	if m.ExpireAt == nil {
		return false
	}
	return time.Now().After(*m.ExpireAt)
}

// IsPublished 判断消息是否已发布
func (m *Message) IsPublished() bool {
	return time.Now().After(m.PublishAt) || time.Now().Equal(m.PublishAt)
}
