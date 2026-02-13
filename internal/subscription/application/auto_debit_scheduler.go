// 变更说明：新增自动扣款和续费提醒调度服务，支持定时任务处理
package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/subscription/domain"
	"github.com/wyfcoding/pkg/messagequeue"
)

// AutoDebitScheduler 自动扣款调度服务
type AutoDebitScheduler struct {
	subscriptionRepo  domain.SubscriptionRepository
	autoDebitRepo     domain.AutoDebitRepository
	paymentGateway    domain.PaymentGateway
	notification      domain.NotificationSender
	publisher         messagequeue.EventPublisher
	logger            *slog.Logger
	maxRetryAttempts  int
	retryInterval     time.Duration
}

// NewAutoDebitScheduler 创建自动扣款调度服务
func NewAutoDebitScheduler(
	subscriptionRepo domain.SubscriptionRepository,
	autoDebitRepo domain.AutoDebitRepository,
	paymentGateway domain.PaymentGateway,
	notification domain.NotificationSender,
	publisher messagequeue.EventPublisher,
	logger *slog.Logger,
) *AutoDebitScheduler {
	return &AutoDebitScheduler{
		subscriptionRepo:  subscriptionRepo,
		autoDebitRepo:     autoDebitRepo,
		paymentGateway:    paymentGateway,
		notification:      notification,
		publisher:         publisher,
		logger:            logger,
		maxRetryAttempts:  3,
		retryInterval:     time.Hour,
	}
}

// ProcessAutoDebits 处理需要自动扣款的订阅
func (s *AutoDebitScheduler) ProcessAutoDebits(ctx context.Context) error {
	now := time.Now()
	
	query := &domain.SubscriptionQuery{
		Status:   ptrSubscriptionStatus(domain.SubscriptionStatusActive),
		AutoRenew: ptrBool(true),
		PageSize: 100,
	}
	
	subscriptions, _, err := s.subscriptionRepo.ListSubscriptions(ctx, query)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list subscriptions for auto debit", "error", err)
		return err
	}
	
	for _, sub := range subscriptions {
		if !sub.NeedsAutoDebit(now) {
			continue
		}
		
		if err := s.processAutoDebit(ctx, sub); err != nil {
			s.logger.ErrorContext(ctx, "failed to process auto debit",
				"subscription_id", sub.ID,
				"user_id", sub.UserID,
				"error", err)
			continue
		}
	}
	
	return nil
}

// processAutoDebit 处理单个订阅的自动扣款
func (s *AutoDebitScheduler) processAutoDebit(ctx context.Context, sub *domain.Subscription) error {
	plan, err := s.subscriptionRepo.GetPlan(ctx, sub.PlanID)
	if err != nil || plan == nil {
		return fmt.Errorf("failed to get plan: %w", err)
	}
	
	record := sub.CreateAutoDebitRecord(plan.Price)
	record.AttemptCount++
	record.LastAttempt = time.Now()
	
	if err := s.autoDebitRepo.Save(ctx, record); err != nil {
		return fmt.Errorf("failed to save debit record: %w", err)
	}
	
	if s.paymentGateway == nil {
		s.logger.WarnContext(ctx, "payment gateway not configured, skipping auto debit",
			"subscription_id", sub.ID)
		return nil
	}
	
	transactionID, err := s.paymentGateway.ChargeUser(ctx, sub.UserID, plan.Price, 
		fmt.Sprintf("Subscription renewal - Plan: %s", plan.Name))
	
	if err != nil {
		record.Status = domain.DebitStatusFailed
		record.ErrorMessage = err.Error()
		_ = s.autoDebitRepo.UpdateStatus(ctx, record.RecordID, record.Status, record.ErrorMessage)
		
		if s.notification != nil {
			_ = s.notification.SendAutoDebitResult(ctx, sub.UserID, false, plan.Price)
		}
		
		s.publishAutoDebitEvent(ctx, sub, record)
		return fmt.Errorf("payment failed: %w", err)
	}
	
	record.Status = domain.DebitStatusSucceeded
	_ = s.autoDebitRepo.UpdateStatus(ctx, record.RecordID, record.Status, "")
	
	sub.Renew(plan.Duration)
	if err := s.subscriptionRepo.SaveSubscription(ctx, sub); err != nil {
		s.logger.ErrorContext(ctx, "failed to renew subscription after successful payment",
			"subscription_id", sub.ID, "transaction_id", transactionID)
	}
	
	if s.notification != nil {
		_ = s.notification.SendAutoDebitResult(ctx, sub.UserID, true, plan.Price)
	}
	
	s.publishAutoDebitEvent(ctx, sub, record)
	s.logger.InfoContext(ctx, "auto debit completed successfully",
		"subscription_id", sub.ID,
		"transaction_id", transactionID)
	
	return nil
}

// publishAutoDebitEvent 发布自动扣款事件
func (s *AutoDebitScheduler) publishAutoDebitEvent(ctx context.Context, sub *domain.Subscription, record *domain.AutoDebitRecord) {
	if s.publisher == nil {
		return
	}
	
	event := &domain.AutoDebitExecutedEvent{
		RecordID:       record.RecordID,
		SubscriptionID: record.SubscriptionID,
		UserID:         sub.UserID,
		Amount:         record.Amount,
		Status:         record.Status,
		AttemptCount:   record.AttemptCount,
		Timestamp:      time.Now(),
	}
	
	_ = s.publisher.Publish(ctx, domain.AutoDebitExecutedEventType, record.RecordID, event)
}

// RenewalReminderService 续费提醒服务
type RenewalReminderService struct {
	subscriptionRepo domain.SubscriptionRepository
	reminderRepo     domain.RenewalReminderRepository
	notification     domain.NotificationSender
	publisher        messagequeue.EventPublisher
	logger           *slog.Logger
}

// NewRenewalReminderService 创建续费提醒服务
func NewRenewalReminderService(
	subscriptionRepo domain.SubscriptionRepository,
	reminderRepo domain.RenewalReminderRepository,
	notification domain.NotificationSender,
	publisher messagequeue.EventPublisher,
	logger *slog.Logger,
) *RenewalReminderService {
	return &RenewalReminderService{
		subscriptionRepo: subscriptionRepo,
		reminderRepo:     reminderRepo,
		notification:     notification,
		publisher:        publisher,
		logger:           logger,
	}
}

// SendRenewalReminders 发送续费提醒
func (s *RenewalReminderService) SendRenewalReminders(ctx context.Context) error {
	now := time.Now()
	
	query := &domain.SubscriptionQuery{
		Status:   ptrSubscriptionStatus(domain.SubscriptionStatusActive),
		PageSize: 100,
	}
	
	subscriptions, _, err := s.subscriptionRepo.ListSubscriptions(ctx, query)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list subscriptions for reminders", "error", err)
		return err
	}
	
	for _, sub := range subscriptions {
		shouldSend, reminderType := sub.ShouldSendReminder(now)
		if !shouldSend {
			continue
		}
		
		if sent, _ := s.reminderRepo.GetSentReminders(ctx, uint(sub.ID), reminderType); sent {
			continue
		}
		
		plan, err := s.subscriptionRepo.GetPlan(ctx, sub.PlanID)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to get plan for reminder",
				"subscription_id", sub.ID, "error", err)
			continue
		}
		
		planName := ""
		if plan != nil {
			planName = plan.Name
		}
		
		reminder := &domain.RenewalReminder{
			SubscriptionID: uint(sub.ID),
			UserID:         sub.UserID,
			PlanName:       planName,
			EndDate:        sub.EndDate,
			DaysRemaining:  int(time.Until(sub.EndDate).Hours() / 24),
			ReminderType:   reminderType,
		}
		
		if s.notification != nil {
			if err := s.notification.SendRenewalReminder(ctx, sub.UserID, reminder); err != nil {
				s.logger.ErrorContext(ctx, "failed to send renewal reminder",
					"subscription_id", sub.ID, "error", err)
				continue
			}
		}
		
		_ = s.reminderRepo.MarkAsSent(ctx, uint(sub.ID), reminderType)
		_ = s.reminderRepo.Save(ctx, reminder)
		
		s.publishExpiringEvent(ctx, sub, reminder.DaysRemaining)
		s.logger.InfoContext(ctx, "renewal reminder sent",
			"subscription_id", sub.ID,
			"reminder_type", reminderType)
	}
	
	return nil
}

// publishExpiringEvent 发布订阅即将过期事件
func (s *RenewalReminderService) publishExpiringEvent(ctx context.Context, sub *domain.Subscription, daysRemaining int) {
	if s.publisher == nil {
		return
	}
	
	event := &domain.SubscriptionExpiringEvent{
		SubscriptionID: uint64(sub.ID),
		UserID:         sub.UserID,
		PlanID:         sub.PlanID,
		EndDate:        sub.EndDate,
		DaysRemaining:  daysRemaining,
		Timestamp:      time.Now(),
	}
	
	_ = s.publisher.Publish(ctx, domain.SubscriptionExpiringEventType, fmt.Sprintf("%d", sub.ID), event)
}

// 辅助函数
func ptrSubscriptionStatus(s domain.SubscriptionStatus) *domain.SubscriptionStatus {
	return &s
}

func ptrBool(b bool) *bool {
	return &b
}
