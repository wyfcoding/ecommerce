package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// OrderTag 订单标签
type OrderTag struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`      // 标签颜色
	Type      string    `json:"type"`       // 标签类型: SYSTEM, CUSTOM
	Priority  int       `json:"priority"`   // 优先级
	Enabled   bool      `json:"enabled"`    // 是否启用
	CreatorID uint64    `json:"creator_id"` // 创建者ID
}

// OrderTagRelation 订单标签关系
type OrderTagRelation struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	OrderID   uint64    `json:"order_id"`
	TagID     uint64    `json:"tag_id"`
	CreatorID uint64    `json:"creator_id"`
	Remark    string    `json:"remark"` // 添加标签时的备注
}

// OrderNote 订单备注
type OrderNote struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	OrderID   uint64    `json:"order_id"`
	UserID    uint64    `json:"user_id"`
	UserName  string    `json:"user_name"`  // 用户名称
	UserType  string    `json:"user_type"`  // 用户类型: ADMIN, CUSTOMER, MERCHANT
	Content   string    `json:"content"`    // 备注内容
	Type      string    `json:"type"`       // 备注类型: SYSTEM, CUSTOMER_SERVICE, OPERATION, RISK
	Priority  string    `json:"priority"`   // 优先级: HIGH, MEDIUM, LOW
	IsPrivate bool      `json:"is_private"` // 是否私有（仅管理员可见）
	Metadata  string    `json:"metadata"`   // 元数据 (JSON格式)
}

// TagManager 标签管理器
type TagManager struct {
	tagRepo OrderTagRepository
}

// NewTagManager 创建标签管理器
func NewTagManager(tagRepo OrderTagRepository) *TagManager {
	return &TagManager{
		tagRepo: tagRepo,
	}
}

// CreateTag 创建标签
func (m *TagManager) CreateTag(ctx context.Context, name, color, tagType string, priority int, creatorID uint64) (*OrderTag, error) {
	// 验证标签名称
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("tag name cannot be empty")
	}

	// 检查标签是否已存在
	existingTag, err := m.tagRepo.FindTagByName(ctx, name)
	if err == nil && existingTag != nil {
		return nil, fmt.Errorf("tag '%s' already exists", name)
	}

	// 创建标签
	tag := &OrderTag{
		Name:      name,
		Color:     color,
		Type:      tagType,
		Priority:  priority,
		Enabled:   true,
		CreatorID: creatorID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = m.tagRepo.SaveTag(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to save tag: %w", err)
	}

	return tag, nil
}

// UpdateTag 更新标签
func (m *TagManager) UpdateTag(ctx context.Context, tagID uint64, updates map[string]interface{}) (*OrderTag, error) {
	tag, err := m.tagRepo.FindTagByID(ctx, tagID)
	if err != nil {
		return nil, fmt.Errorf("failed to find tag: %w", err)
	}

	// 更新标签属性
	if name, ok := updates["name"].(string); ok && name != "" {
		// 检查名称是否重复
		existingTag, err := m.tagRepo.FindTagByName(ctx, name)
		if err == nil && existingTag != nil && uint64(existingTag.ID) != tagID {
			return nil, fmt.Errorf("tag '%s' already exists", name)
		}
		tag.Name = name
	}

	if color, ok := updates["color"].(string); ok {
		tag.Color = color
	}

	if priority, ok := updates["priority"].(int); ok {
		tag.Priority = priority
	}

	if enabled, ok := updates["enabled"].(bool); ok {
		tag.Enabled = enabled
	}

	tag.UpdatedAt = time.Now()

	err = m.tagRepo.UpdateTag(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}

	return tag, nil
}

// DeleteTag 删除标签
func (m *TagManager) DeleteTag(ctx context.Context, tagID uint64) error {
	// 检查标签是否被使用
	usageCount, err := m.tagRepo.CountTagUsage(ctx, tagID)
	if err != nil {
		return fmt.Errorf("failed to check tag usage: %w", err)
	}

	if usageCount > 0 {
		return fmt.Errorf("cannot delete tag that is in use by %d orders", usageCount)
	}

	err = m.tagRepo.DeleteTag(ctx, tagID)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	return nil
}

// AddTagsToOrder 为订单添加标签
func (m *TagManager) AddTagsToOrder(ctx context.Context, orderID uint64, tagIDs []uint64, creatorID uint64, remark string) error {
	for _, tagID := range tagIDs {
		// 检查标签是否存在
		tag, err := m.tagRepo.FindTagByID(ctx, tagID)
		if err != nil {
			return fmt.Errorf("failed to find tag %d: %w", tagID, err)
		}

		if !tag.Enabled {
			return fmt.Errorf("tag '%s' is disabled", tag.Name)
		}

		// 检查是否已添加
		exists, err := m.tagRepo.CheckTagRelation(ctx, orderID, tagID)
		if err != nil {
			return fmt.Errorf("failed to check tag relation: %w", err)
		}

		if exists {
			continue // 已存在则跳过
		}

		// 创建标签关系
		relation := &OrderTagRelation{
			OrderID:   orderID,
			TagID:     tagID,
			CreatorID: creatorID,
			Remark:    remark,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = m.tagRepo.SaveTagRelation(ctx, relation)
		if err != nil {
			return fmt.Errorf("failed to save tag relation: %w", err)
		}
	}

	return nil
}

// RemoveTagsFromOrder 从订单移除标签
func (m *TagManager) RemoveTagsFromOrder(ctx context.Context, orderID uint64, tagIDs []uint64) error {
	for _, tagID := range tagIDs {
		err := m.tagRepo.DeleteTagRelation(ctx, orderID, tagID)
		if err != nil {
			return fmt.Errorf("failed to remove tag %d: %w", tagID, err)
		}
	}

	return nil
}

// GetOrderTags 获取订单的所有标签
func (m *TagManager) GetOrderTags(ctx context.Context, orderID uint64) ([]*OrderTag, error) {
	tags, err := m.tagRepo.FindTagsByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order tags: %w", err)
	}

	return tags, nil
}

// GetOrdersByTag 根据标签获取订单
func (m *TagManager) GetOrdersByTag(ctx context.Context, tagID uint64, page, pageSize int) ([]uint64, int64, error) {
	orderIDs, total, err := m.tagRepo.FindOrdersByTagID(ctx, tagID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get orders by tag: %w", err)
	}

	return orderIDs, total, nil
}

// NoteManager 备注管理器
type NoteManager struct {
	noteRepo OrderNoteRepository
}

// NewNoteManager 创建备注管理器
func NewNoteManager(noteRepo OrderNoteRepository) *NoteManager {
	return &NoteManager{
		noteRepo: noteRepo,
	}
}

// AddNote 添加订单备注
func (m *NoteManager) AddNote(ctx context.Context, orderID, userID uint64, userName, userType, content, noteType, priority string, isPrivate bool, metadata map[string]interface{}) (*OrderNote, error) {
	// 验证备注内容
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("note content cannot be empty")
	}

	// 序列化元数据
	var metadataJSON string
	if metadata != nil {
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = string(metadataBytes)
	}

	// 创建备注
	note := &OrderNote{
		OrderID:   orderID,
		UserID:    userID,
		UserName:  userName,
		UserType:  userType,
		Content:   content,
		Type:      noteType,
		Priority:  priority,
		IsPrivate: isPrivate,
		Metadata:  metadataJSON,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := m.noteRepo.SaveNote(ctx, note)
	if err != nil {
		return nil, fmt.Errorf("failed to save note: %w", err)
	}

	return note, nil
}

// UpdateNote 更新备注
func (m *NoteManager) UpdateNote(ctx context.Context, noteID uint64, content string, metadata map[string]interface{}) (*OrderNote, error) {
	note, err := m.noteRepo.FindNoteByID(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to find note: %w", err)
	}

	// 更新内容
	if content != "" {
		note.Content = content
	}

	// 更新元数据
	if metadata != nil {
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		note.Metadata = string(metadataBytes)
	}

	note.UpdatedAt = time.Now()

	err = m.noteRepo.UpdateNote(ctx, note)
	if err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	return note, nil
}

// DeleteNote 删除备注
func (m *NoteManager) DeleteNote(ctx context.Context, noteID uint64) error {
	err := m.noteRepo.DeleteNote(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}

// GetOrderNotes 获取订单的所有备注
func (m *NoteManager) GetOrderNotes(ctx context.Context, orderID uint64, includePrivate bool, noteTypes []string, page, pageSize int) ([]*OrderNote, int64, error) {
	notes, total, err := m.noteRepo.FindNotesByOrderID(ctx, orderID, includePrivate, noteTypes, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get order notes: %w", err)
	}

	return notes, total, nil
}

// SearchNotes 搜索备注
func (m *NoteManager) SearchNotes(ctx context.Context, keyword string, userID uint64, noteTypes []string, startTime, endTime *time.Time, page, pageSize int) ([]*OrderNote, int64, error) {
	notes, total, err := m.noteRepo.SearchNotes(ctx, keyword, userID, noteTypes, startTime, endTime, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search notes: %w", err)
	}

	return notes, total, nil
}

// 预定义系统标签
var SystemTags = []struct {
	Name     string
	Color    string
	Priority int
}{
	{"高风险", "#ff4d4f", 100},
	{"待审核", "#faad14", 90},
	{"加急", "#f5222d", 80},
	{"VIP客户", "#722ed1", 70},
	{"大额订单", "#eb2f96", 60},
	{"预售订单", "#13c2c2", 50},
	{"国际订单", "#1890ff", 40},
	{"退货订单", "#52c41a", 30},
	{"投诉订单", "#fa8c16", 20},
	{"测试订单", "#8c8c8c", 10},
}

// InitializeSystemTags 初始化系统标签
func (m *TagManager) InitializeSystemTags(ctx context.Context, creatorID uint64) error {
	for _, sysTag := range SystemTags {
		// 检查是否已存在
		existingTag, err := m.tagRepo.FindTagByName(ctx, sysTag.Name)
		if err == nil && existingTag != nil {
			continue // 已存在则跳过
		}

		// 创建系统标签
		tag := &OrderTag{
			Name:      sysTag.Name,
			Color:     sysTag.Color,
			Type:      "SYSTEM",
			Priority:  sysTag.Priority,
			Enabled:   true,
			CreatorID: creatorID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = m.tagRepo.SaveTag(ctx, tag)
		if err != nil {
			return fmt.Errorf("failed to save system tag '%s': %w", sysTag.Name, err)
		}
	}

	return nil
}

// OrderTagRepository 订单标签仓储接口
type OrderTagRepository interface {
	// 标签管理
	SaveTag(ctx context.Context, tag *OrderTag) error
	FindTagByID(ctx context.Context, id uint64) (*OrderTag, error)
	FindTagByName(ctx context.Context, name string) (*OrderTag, error)
	UpdateTag(ctx context.Context, tag *OrderTag) error
	DeleteTag(ctx context.Context, id uint64) error
	ListTags(ctx context.Context, tagType string, enabledOnly bool, page, pageSize int) ([]*OrderTag, int64, error)
	CountTagUsage(ctx context.Context, tagID uint64) (int64, error)

	// 标签关系管理
	SaveTagRelation(ctx context.Context, relation *OrderTagRelation) error
	DeleteTagRelation(ctx context.Context, orderID, tagID uint64) error
	CheckTagRelation(ctx context.Context, orderID, tagID uint64) (bool, error)
	FindTagsByOrderID(ctx context.Context, orderID uint64) ([]*OrderTag, error)
	FindOrdersByTagID(ctx context.Context, tagID uint64, page, pageSize int) ([]uint64, int64, error)
	BatchUpdateTags(ctx context.Context, orderIDs []uint64, tagIDs []uint64, creatorID uint64, remark string) error
}

// OrderNoteRepository 订单备注仓储接口
type OrderNoteRepository interface {
	SaveNote(ctx context.Context, note *OrderNote) error
	FindNoteByID(ctx context.Context, id uint64) (*OrderNote, error)
	UpdateNote(ctx context.Context, note *OrderNote) error
	DeleteNote(ctx context.Context, id uint64) error
	FindNotesByOrderID(ctx context.Context, orderID uint64, includePrivate bool, noteTypes []string, page, pageSize int) ([]*OrderNote, int64, error)
	SearchNotes(ctx context.Context, keyword string, userID uint64, noteTypes []string, startTime, endTime *time.Time, page, pageSize int) ([]*OrderNote, int64, error)
}
