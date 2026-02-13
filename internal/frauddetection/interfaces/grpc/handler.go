package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/frauddetection/v1"
	"github.com/wyfcoding/ecommerce/internal/frauddetection/application"
	"github.com/wyfcoding/ecommerce/internal/frauddetection/domain"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedFraudDetectionServiceServer
	service *application.FraudDetectionService
	logger  *slog.Logger
}

func NewServer(service *application.FraudDetectionService, logger *slog.Logger) *Server {
	return &Server{
		service: service,
		logger:  logger,
	}
}

func (s *Server) AnalyzeTransaction(ctx context.Context, req *pb.AnalyzeTransactionRequest) (*pb.RiskAssessment, error) {
	start := time.Now()
	s.logger.InfoContext(ctx, "AnalyzeTransaction called", "transaction_id", req.TransactionId, "user_id", req.UserId)

	cmd := &application.AnalyzeTransactionCommand{
		TransactionID:     req.TransactionId,
		UserID:            req.UserId,
		Amount:            req.Amount,
		Currency:          req.Currency,
		PaymentMethod:     req.PaymentMethod,
		IPAddress:         req.IpAddress,
		DeviceFingerprint: req.DeviceFingerprint,
		UserAgent:         req.UserAgent,
		MerchantID:        req.MerchantId,
		CardLastFour:      req.CardLastFour,
		Email:             req.Email,
		Phone:             req.Phone,
		Metadata:          req.Metadata,
	}

	if req.GeoLocation != nil {
		cmd.Country = req.GeoLocation.Country
		cmd.City = req.GeoLocation.City
		cmd.Latitude = req.GeoLocation.Latitude
		cmd.Longitude = req.GeoLocation.Longitude
	}

	assessment, err := s.service.AnalyzeTransaction(ctx, cmd)
	if err != nil {
		s.logger.ErrorContext(ctx, "AnalyzeTransaction failed", "error", err, "duration", time.Since(start))
		return nil, err
	}

	s.logger.InfoContext(ctx, "AnalyzeTransaction completed", "risk_level", assessment.RiskLevel.String(), "duration", time.Since(start))
	return convertAssessmentToProto(assessment), nil
}

func (s *Server) GetRiskAssessment(ctx context.Context, req *pb.GetRiskAssessmentRequest) (*pb.RiskAssessment, error) {
	start := time.Now()
	s.logger.DebugContext(ctx, "GetRiskAssessment called", "assessment_id", req.AssessmentId)

	assessment, err := s.service.GetRiskAssessment(ctx, req.AssessmentId)
	if err != nil {
		s.logger.ErrorContext(ctx, "GetRiskAssessment failed", "error", err, "duration", time.Since(start))
		return nil, err
	}

	return convertAssessmentToProto(assessment), nil
}

func (s *Server) GetBlacklist(ctx context.Context, req *pb.GetBlacklistRequest) (*pb.GetBlacklistResponse, error) {
	start := time.Now()
	s.logger.DebugContext(ctx, "GetBlacklist called", "type", req.Type)

	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}

	entries, total, err := s.service.GetBlacklist(ctx, domain.BlacklistType(req.Type.String()), page, pageSize)
	if err != nil {
		s.logger.ErrorContext(ctx, "GetBlacklist failed", "error", err, "duration", time.Since(start))
		return nil, err
	}

	return &pb.GetBlacklistResponse{
		Entries: convertBlacklistEntriesToProto(entries),
		Total:   int32(total),
	}, nil
}

func (s *Server) AddToBlacklist(ctx context.Context, req *pb.AddToBlacklistRequest) (*pb.BlacklistEntry, error) {
	start := time.Now()
	s.logger.InfoContext(ctx, "AddToBlacklist called", "type", req.Type, "value", req.Value)

	var expiresIn *time.Duration
	if req.ExpiresInSeconds > 0 {
		d := time.Duration(req.ExpiresInSeconds) * time.Second
		expiresIn = &d
	}

	cmd := &application.AddToBlacklistCommand{
		Type:      domain.BlacklistType(req.Type.String()),
		Value:     req.Value,
		Reason:    req.Reason,
		CreatedBy: req.CreatedBy,
		ExpiresIn: expiresIn,
	}

	entry, err := s.service.AddToBlacklist(ctx, cmd)
	if err != nil {
		s.logger.ErrorContext(ctx, "AddToBlacklist failed", "error", err, "duration", time.Since(start))
		return nil, err
	}

	return convertBlacklistEntryToProto(entry), nil
}

func (s *Server) RemoveFromBlacklist(ctx context.Context, req *pb.RemoveFromBlacklistRequest) (*emptypb.Empty, error) {
	start := time.Now()
	s.logger.InfoContext(ctx, "RemoveFromBlacklist called", "id", req.Id)

	if err := s.service.RemoveFromBlacklist(ctx, req.Id); err != nil {
		s.logger.ErrorContext(ctx, "RemoveFromBlacklist failed", "error", err, "duration", time.Since(start))
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetRiskRules(ctx context.Context, req *pb.GetRiskRulesRequest) (*pb.GetRiskRulesResponse, error) {
	start := time.Now()
	s.logger.DebugContext(ctx, "GetRiskRules called", "type", req.Type)

	rules, err := s.service.GetRiskRules(ctx, domain.RuleType(req.Type.String()), req.EnabledOnly)
	if err != nil {
		s.logger.ErrorContext(ctx, "GetRiskRules failed", "error", err, "duration", time.Since(start))
		return nil, err
	}

	return &pb.GetRiskRulesResponse{
		Rules: convertRiskRulesToProto(rules),
		Total: int32(len(rules)),
	}, nil
}

func (s *Server) CreateRiskRule(ctx context.Context, req *pb.CreateRiskRuleRequest) (*pb.RiskRule, error) {
	start := time.Now()
	s.logger.InfoContext(ctx, "CreateRiskRule called", "name", req.Name)

	cmd := &application.CreateRiskRuleCommand{
		Name:       req.Name,
		Type:       domain.RuleType(req.Type.String()),
		Condition:  req.Condition,
		RiskWeight: int(req.RiskWeight),
		Action:     domain.RiskAction(req.Action.String()),
		Priority:   int(req.Priority),
	}

	rule, err := s.service.CreateRiskRule(ctx, cmd)
	if err != nil {
		s.logger.ErrorContext(ctx, "CreateRiskRule failed", "error", err, "duration", time.Since(start))
		return nil, err
	}

	return convertRiskRuleToProto(rule), nil
}

func (s *Server) UpdateRiskRule(ctx context.Context, req *pb.UpdateRiskRuleRequest) (*pb.RiskRule, error) {
	start := time.Now()
	s.logger.InfoContext(ctx, "UpdateRiskRule called", "id", req.Id)

	return nil, fmt.Errorf("not implemented")
}

func (s *Server) DeleteRiskRule(ctx context.Context, req *pb.DeleteRiskRuleRequest) (*emptypb.Empty, error) {
	start := time.Now()
	s.logger.InfoContext(ctx, "DeleteRiskRule called", "id", req.Id)

	return nil, fmt.Errorf("not implemented")
}

func (s *Server) GetDeviceFingerprint(ctx context.Context, req *pb.GetDeviceFingerprintRequest) (*pb.DeviceFingerprint, error) {
	start := time.Now()
	s.logger.DebugContext(ctx, "GetDeviceFingerprint called", "fingerprint", req.Fingerprint)

	device, err := s.service.GetDeviceFingerprint(ctx, req.Fingerprint)
	if err != nil {
		s.logger.ErrorContext(ctx, "GetDeviceFingerprint failed", "error", err, "duration", time.Since(start))
		return nil, err
	}

	if device == nil {
		return nil, fmt.Errorf("device fingerprint not found")
	}

	return convertDeviceFingerprintToProto(device), nil
}

func (s *Server) RecordSuspiciousActivity(ctx context.Context, req *pb.RecordSuspiciousActivityRequest) (*emptypb.Empty, error) {
	start := time.Now()
	s.logger.WarnContext(ctx, "RecordSuspiciousActivity called", "user_id", req.UserId, "type", req.ActivityType)

	cmd := &application.RecordSuspiciousActivityCommand{
		UserID:            req.UserId,
		ActivityType:      req.ActivityType,
		Description:       req.Description,
		IPAddress:         req.IpAddress,
		DeviceFingerprint: req.DeviceFingerprint,
		Metadata:          req.Metadata,
	}

	if err := s.service.RecordSuspiciousActivity(ctx, cmd); err != nil {
		s.logger.ErrorContext(ctx, "RecordSuspiciousActivity failed", "error", err, "duration", time.Since(start))
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetUserRiskProfile(ctx context.Context, req *pb.GetUserRiskProfileRequest) (*pb.UserRiskProfile, error) {
	start := time.Now()
	s.logger.DebugContext(ctx, "GetUserRiskProfile called", "user_id", req.UserId)

	profile, err := s.service.GetUserRiskProfile(ctx, req.UserId)
	if err != nil {
		s.logger.ErrorContext(ctx, "GetUserRiskProfile failed", "error", err, "duration", time.Since(start))
		return nil, err
	}

	if profile == nil {
		return nil, fmt.Errorf("user risk profile not found")
	}

	return convertUserRiskProfileToProto(profile), nil
}

func convertAssessmentToProto(a *domain.RiskAssessment) *pb.RiskAssessment {
	factors := make([]*pb.RiskFactor, len(a.RiskFactors))
	for i, f := range a.RiskFactors {
		factors[i] = &pb.RiskFactor{
			FactorName:       f.FactorName,
			Description:      f.Description,
			Weight:           int32(f.Weight),
			ContributionScore: int32(f.ContributionScore),
		}
	}

	return &pb.RiskAssessment{
		AssessmentId:      a.ID,
		TransactionId:     a.TransactionID,
		UserId:            a.UserID,
		RiskLevel:         pb.RiskLevel(a.RiskLevel),
		RecommendedAction: pb.RiskAction(pb.RiskAction_value[string(a.RecommendedAction)]),
		RiskScore:         int32(a.RiskScore),
		RiskFactors:       factors,
		TriggeredRules:    a.TriggeredRules,
		Explanation:       a.Explanation,
		CreatedAt:         timestamppb.New(a.CreatedAt),
	}
}

func convertBlacklistEntriesToProto(entries []*domain.BlacklistEntry) []*pb.BlacklistEntry {
	result := make([]*pb.BlacklistEntry, len(entries))
	for i, e := range entries {
		result[i] = convertBlacklistEntryToProto(e)
	}
	return result
}

func convertBlacklistEntryToProto(e *domain.BlacklistEntry) *pb.BlacklistEntry {
	entry := &pb.BlacklistEntry{
		Id:        e.ID,
		Type:      pb.BlacklistType(pb.BlacklistType_value[string(e.Type)]),
		Value:     e.Value,
		Reason:    e.Reason,
		CreatedBy: e.CreatedBy,
		CreatedAt: timestamppb.New(e.CreatedAt),
		IsActive:  e.IsActive,
	}
	if e.ExpiresAt != nil {
		entry.ExpiresAt = timestamppb.New(*e.ExpiresAt)
	}
	return entry
}

func convertRiskRulesToProto(rules []*domain.RiskRule) []*pb.RiskRule {
	result := make([]*pb.RiskRule, len(rules))
	for i, r := range rules {
		result[i] = convertRiskRuleToProto(r)
	}
	return result
}

func convertRiskRuleToProto(r *domain.RiskRule) *pb.RiskRule {
	return &pb.RiskRule{
		Id:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Type:        pb.RuleType(pb.RuleType_value[string(r.Type)]),
		Condition:   r.Condition,
		RiskWeight:  int32(r.RiskWeight),
		Action:      pb.RiskAction(pb.RiskAction_value[string(r.Action)]),
		Enabled:     r.Enabled,
		Priority:    int32(r.Priority),
		CreatedAt:   timestamppb.New(r.CreatedAt),
		UpdatedAt:   timestamppb.New(r.UpdatedAt),
	}
}

func convertDeviceFingerprintToProto(d *domain.DeviceFingerprint) *pb.DeviceFingerprint {
	return &pb.DeviceFingerprint{
		Fingerprint:            d.Fingerprint,
		FirstSeenUserId:        d.FirstSeenUserID,
		AssociatedAccountsCount: int32(d.AssociatedAccountsCount),
		AssociatedIps:          d.AssociatedIPs,
		IsSuspicious:           d.IsSuspicious,
		DeviceType:             d.DeviceType,
		Os:                     d.OS,
		Browser:                d.Browser,
		FirstSeenAt:            timestamppb.New(d.FirstSeenAt),
		LastSeenAt:             timestamppb.New(d.LastSeenAt),
	}
}

func convertUserRiskProfileToProto(p *domain.UserRiskProfile) *pb.UserRiskProfile {
	profile := &pb.UserRiskProfile{
		UserId:             p.UserID,
		RiskScore:          int32(p.RiskScore),
		RiskLevel:          pb.RiskLevel(p.RiskLevel),
		TotalTransactions:  int32(p.TotalTransactions),
		FlaggedTransactions: int32(p.FlaggedTransactions),
		BlockedTransactions: int32(p.BlockedTransactions),
		RiskTags:           p.RiskTags,
	}
	if p.FirstTransactionAt != nil {
		profile.FirstTransactionAt = timestamppb.New(*p.FirstTransactionAt)
	}
	if p.LastTransactionAt != nil {
		profile.LastTransactionAt = timestamppb.New(*p.LastTransactionAt)
	}
	if p.LastAssessmentAt != nil {
		profile.LastAssessmentAt = timestamppb.New(*p.LastAssessmentAt)
	}
	return profile
}
