package application

import (
	"context"
	"encoding/json" // 导入JSON编码/解码库。
	"fmt"
	"time" // 导入时间库。

	"github.com/wyfcoding/ecommerce/internal/risk_security/domain/entity"     // 导入风控安全领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/risk_security/domain/repository" // 导入风控安全领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// RiskService 结构体定义了风控安全相关的应用服务。
// 它协调领域层和基础设施层，处理风险评估、黑名单管理和用户行为记录等业务逻辑。
type RiskService struct {
	repo   repository.RiskRepository // 依赖RiskRepository接口，用于数据持久化操作。
	logger *slog.Logger              // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewRiskService 创建并返回一个新的 RiskService 实例。
func NewRiskService(repo repository.RiskRepository, logger *slog.Logger) *RiskService {
	return &RiskService{
		repo:   repo,
		logger: logger,
	}
}

// EvaluateRisk 评估指定用户操作的风险。
// ctx: 上下文。
// userID: 用户ID。
// ip: 用户IP地址。
// deviceID: 用户设备ID。
// amount: 关联金额（例如，交易金额）。
// 返回风险分析结果 RiskAnalysisResult 实体和可能发生的错误。
func (s *RiskService) EvaluateRisk(ctx context.Context, userID uint64, ip, deviceID string, amount int64) (*entity.RiskAnalysisResult, error) {
	// 1. 检查黑名单：
	// 检查IP是否在黑名单中。
	isIPBlacklisted, err := s.repo.IsBlacklisted(ctx, entity.BlacklistTypeIP, ip)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check IP blacklist", "ip", ip, "error", err)
	}
	if isIPBlacklisted {
		return s.createResult(ctx, userID, entity.RiskLevelCritical, 100, "IP address found in blacklist")
	}

	// 检查用户是否在黑名单中。
	isUserBlacklisted, err := s.repo.IsBlacklisted(ctx, entity.BlacklistTypeUser, fmt.Sprintf("%d", userID)) // 用户ID转换为字符串。
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check user blacklist", "user_id", userID, "error", err)
	}
	if isUserBlacklisted {
		return s.createResult(ctx, userID, entity.RiskLevelCritical, 100, "User ID found in blacklist")
	}

	// 2. 评估风险规则（模拟逻辑）：
	score := int32(0)
	// 规则1：如果金额过大，增加风险分数。
	if amount > 1000000 { // 假设金额超过10000.00元为高风险。
		score += 20
	}
	// TODO: 实际系统中，这里会有一系列复杂的风控规则引擎来计算风险分数。

	// 3. 评估用户行为（模拟逻辑）：
	behavior, err := s.repo.GetUserBehavior(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user behavior", "user_id", userID, "error", err)
	}
	// 规则2：如果用户上次登录IP与当前IP不符，增加风险分数。
	if behavior != nil && behavior.LastLoginIP != "" && behavior.LastLoginIP != ip {
		score += 30
	}
	// TODO: 实际系统中，会结合更多用户行为数据（如登录频率、交易频率、历史异常行为）进行评估。

	// 4. 根据风险分数确定风险等级。
	level := entity.RiskLevelLow
	if score > 80 {
		level = entity.RiskLevelCritical // 严重风险。
	} else if score > 60 {
		level = entity.RiskLevelHigh // 高风险。
	} else if score > 40 {
		level = entity.RiskLevelMedium // 中风险。
	}

	// 5. 创建并保存风险分析结果。
	return s.createResult(ctx, userID, level, score, "Risk evaluation completed")
}

// createResult 是一个辅助函数，用于创建并保存 RiskAnalysisResult 实体。
func (s *RiskService) createResult(ctx context.Context, userID uint64, level entity.RiskLevel, score int32, reason string) (*entity.RiskAnalysisResult, error) {
	// 创建风险项列表。
	items := []*entity.RiskItem{
		{
			Type:      entity.RiskTypeAnomalousTransaction, // 暂时指定为异常交易类型。
			Level:     level,
			Score:     score,
			Reason:    reason,
			Timestamp: time.Now(),
		},
	}
	itemsJSON, err := json.Marshal(items) // 将风险项列表序列化为JSON字符串。
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to marshal risk items", "user_id", userID, "error", err)
		return nil, err
	}

	result := &entity.RiskAnalysisResult{
		UserID:    userID,
		RiskScore: score,
		RiskLevel: level,
		RiskItems: string(itemsJSON), // 风险项存储为JSON字符串。
	}

	// 保存风险分析结果。
	if err := s.repo.SaveAnalysisResult(ctx, result); err != nil {
		s.logger.ErrorContext(ctx, "failed to save risk analysis result", "user_id", userID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "risk analysis result saved", "user_id", userID, "risk_level", level)
	return result, nil
}

// AddToBlacklist 将指定类型和值的实体添加到黑名单。
// ctx: 上下文。
// bType: 黑名单类型（例如，“IP”，“User”）。
// value: 黑名单值（例如，IP地址或用户ID字符串）。
// reason: 添加到黑名单的原因。
// duration: 黑名单的有效期。
// 返回可能发生的错误。
func (s *RiskService) AddToBlacklist(ctx context.Context, bType string, value, reason string, duration time.Duration) error {
	blacklist := &entity.Blacklist{
		Type:      entity.BlacklistType(bType), // 转换黑名单类型。
		Value:     value,
		Reason:    reason,
		ExpiresAt: time.Now().Add(duration), // 设置黑名单过期时间。
	}
	if err := s.repo.SaveBlacklist(ctx, blacklist); err != nil {
		s.logger.ErrorContext(ctx, "failed to save blacklist entry", "type", bType, "value", value, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "added to blacklist", "type", bType, "value", value)
	return nil
}

// RemoveFromBlacklist 从黑名单中移除指定ID的条目。
// ctx: 上下文。
// id: 黑名单条目ID。
// 返回可能发生的错误。
func (s *RiskService) RemoveFromBlacklist(ctx context.Context, id uint64) error {
	if err := s.repo.DeleteBlacklist(ctx, id); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete blacklist entry", "id", id, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "removed from blacklist", "id", id)
	return nil
}

// RecordUserBehavior 记录或更新用户的行为数据。
// ctx: 上下文。
// userID: 用户ID。
// ip: 用户IP地址。
// deviceID: 用户设备ID。
// 返回可能发生的错误。
func (s *RiskService) RecordUserBehavior(ctx context.Context, userID uint64, ip, deviceID string) error {
	// 尝试获取现有用户行为记录。
	behavior, err := s.repo.GetUserBehavior(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user behavior for recording", "user_id", userID, "error", err)
		return err
	}

	// 如果不存在，则创建新的行为记录。
	if behavior == nil {
		behavior = &entity.UserBehavior{
			UserID: userID,
		}
	}

	// 更新行为记录的最新登录信息。
	behavior.LastLoginIP = ip
	behavior.LastLoginDevice = deviceID
	behavior.LastLoginTime = time.Now()

	// 保存用户行为记录。
	if err := s.repo.SaveUserBehavior(ctx, behavior); err != nil {
		s.logger.ErrorContext(ctx, "failed to save user behavior", "user_id", userID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "user behavior recorded", "user_id", userID, "ip", ip)
	return nil
}

// GetRiskAnalysisResult retrieves the latest risk analysis result for a user.
func (s *RiskService) GetRiskAnalysisResult(ctx context.Context, userID uint64) (*entity.RiskAnalysisResult, error) {
	// Get the latest result. Limiting to 1.
	results, err := s.repo.ListAnalysisResults(ctx, userID, 1)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no risk analysis result found for user %d", userID)
	}
	return results[0], nil
}

// GetBlacklist retrieves blacklist entries.
// Note: The repository interface for GetBlacklist takes specific type and value.
// To list all or filter, we might need a ListBlacklist method in repo.
// For now, let's implement a simple check or mock list if repo doesn't support listing all.
// Looking at repo interface: GetBlacklist(ctx, bType, value).
// It seems we can't list all blacklist entries easily with current repo interface.
// I will implement a method to check if a value is blacklisted for now, or add ListBlacklist to repo if I could (but I shouldn't modify repo interface if not necessary).
// Let's change the service method to CheckBlacklist or similar, or just return an error for now if listing is not supported.
// Actually, the HTTP handler might want to check a specific value.
// Let's implement CheckBlacklist instead of GetBlacklist (List).
func (s *RiskService) CheckBlacklist(ctx context.Context, bType string, value string) (*entity.Blacklist, error) {
	return s.repo.GetBlacklist(ctx, entity.BlacklistType(bType), value)
}

// GetUserBehavior retrieves user behavior data.
func (s *RiskService) GetUserBehavior(ctx context.Context, userID uint64) (*entity.UserBehavior, error) {
	return s.repo.GetUserBehavior(ctx, userID)
}
