package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/riskanalyzer/application"
	"github.com/wyfcoding/ecommerce/internal/riskanalyzer/domain"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	cmd   *application.RiskCommandService
	query *application.RiskQueryService
}

func NewHandler(cmd *application.RiskCommandService, query *application.RiskQueryService) *Handler {
	return &Handler{
		cmd:   cmd,
		query: query,
	}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	risk := r.Group("/risk")
	{
		risk.POST("/analyze", h.AnalyzeRisk)
		risk.GET("/assessment/:target_id", h.GetAssessment)
		risk.POST("/blacklist", h.AddToBlacklist)
		risk.DELETE("/blacklist", h.RemoveFromBlacklist)
		risk.GET("/blacklist/check", h.CheckBlacklist)
	}

	credit := r.Group("/credit")
	{
		credit.POST("/evaluate/:user_id", h.EvaluateCredit)
		credit.GET("/profile/:user_id", h.GetCreditProfile)
		credit.GET("/records/:user_id", h.GetCreditRecords)
		credit.POST("/freeze", h.FreezeCreditLimit)
		credit.POST("/release", h.ReleaseCreditLimit)
		credit.POST("/record", h.AddCreditRecord)
	}
}

func (h *Handler) AnalyzeRisk(c *gin.Context) {
	var req struct {
		TargetID     string `json:"target_id" binding:"required"`
		TargetType   string `json:"target_type" binding:"required"`
		UserID       uint64 `json:"user_id" binding:"required"`
		DeviceFinger string `json:"device_finger"`
		IPAddress    string `json:"ip_address"`
		UserAgent    string `json:"user_agent"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.AnalyzeRiskCommand{
		TargetID:     req.TargetID,
		TargetType:   req.TargetType,
		UserID:       req.UserID,
		DeviceFinger: req.DeviceFinger,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
	}

	assessment, err := h.cmd.AnalyzeRisk(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toAssessmentDTO(assessment))
}

func (h *Handler) GetAssessment(c *gin.Context) {
	targetID := c.Param("target_id")

	assessment, err := h.query.GetAssessment(c.Request.Context(), targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if assessment == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "assessment not found"})
		return
	}

	c.JSON(http.StatusOK, toAssessmentDTO(assessment))
}

func (h *Handler) AddToBlacklist(c *gin.Context) {
	var req struct {
		Type   int    `json:"type" binding:"required"`
		Value  string `json:"value" binding:"required"`
		Reason string `json:"reason"`
		Source string `json:"source"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.AddToBlacklistCommand{
		Type:   domain.BlacklistType(req.Type),
		Value:  req.Value,
		Reason: req.Reason,
		Source: req.Source,
	}

	if err := h.cmd.AddToBlacklist(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "added to blacklist"})
}

func (h *Handler) RemoveFromBlacklist(c *gin.Context) {
	var req struct {
		Type  int    `json:"type" binding:"required"`
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.RemoveFromBlacklistCommand{
		Type:  domain.BlacklistType(req.Type),
		Value: req.Value,
	}

	if err := h.cmd.RemoveFromBlacklist(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "removed from blacklist"})
}

func (h *Handler) CheckBlacklist(c *gin.Context) {
	blacklistType, _ := strconv.Atoi(c.Query("type"))
	value := c.Query("value")

	if value == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value is required"})
		return
	}

	inBlacklist, err := h.query.IsInBlacklist(c.Request.Context(), domain.BlacklistType(blacklistType), value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"in_blacklist": inBlacklist})
}

func (h *Handler) EvaluateCredit(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	cmd := application.EvaluateCreditCommand{UserID: userID}

	profile, err := h.cmd.EvaluateCredit(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toCreditProfileDTO(profile))
}

func (h *Handler) GetCreditProfile(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	profile, err := h.query.GetCreditProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if profile == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "credit profile not found"})
		return
	}

	c.JSON(http.StatusOK, toCreditProfileDTO(profile))
}

func (h *Handler) GetCreditRecords(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	records, err := h.query.GetCreditRecords(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"records": toCreditRecordDTOs(records)})
}

func (h *Handler) FreezeCreditLimit(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
		Amount int64  `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.FreezeCreditLimitCommand{
		UserID: req.UserID,
		Amount: req.Amount,
	}

	if err := h.cmd.FreezeCreditLimit(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "credit limit frozen"})
}

func (h *Handler) ReleaseCreditLimit(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
		Amount int64  `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.ReleaseCreditLimitCommand{
		UserID: req.UserID,
		Amount: req.Amount,
	}

	if err := h.cmd.ReleaseCreditLimit(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "credit limit released"})
}

func (h *Handler) AddCreditRecord(c *gin.Context) {
	var req struct {
		UserID      uint64 `json:"user_id" binding:"required"`
		RecordType  int    `json:"record_type" binding:"required"`
		ScoreChange int    `json:"score_change"`
		LimitChange int64  `json:"limit_change"`
		RelatedID   string `json:"related_id"`
		Reason      string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.AddCreditRecordCommand{
		UserID:      req.UserID,
		RecordType:  domain.CreditRecordType(req.RecordType),
		ScoreChange: req.ScoreChange,
		LimitChange: req.LimitChange,
		RelatedID:   req.RelatedID,
		Reason:      req.Reason,
	}

	if err := h.cmd.AddCreditRecord(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "credit record added"})
}

type AssessmentDTO struct {
	ID           uint64          `json:"id"`
	TargetID     string          `json:"target_id"`
	TargetType   string          `json:"target_type"`
	UserID       uint64          `json:"user_id"`
	TotalScore   int             `json:"total_score"`
	Level        string          `json:"level"`
	MatchedRules []MatchedRuleDTO `json:"matched_rules"`
	Decision     string          `json:"decision"`
	DeviceFinger string          `json:"device_finger"`
	IPAddress    string          `json:"ip_address"`
	CreatedAt    string          `json:"created_at"`
}

type MatchedRuleDTO struct {
	RuleID      uint64 `json:"rule_id"`
	RuleName    string `json:"rule_name"`
	Score       int    `json:"score"`
	Reason      string `json:"reason"`
	RawEvidence string `json:"raw_evidence"`
}

type CreditProfileDTO struct {
	ID             uint64           `json:"id"`
	UserID         uint64           `json:"user_id"`
	Score          int              `json:"score"`
	Level          string           `json:"level"`
	TotalLimit     int64            `json:"total_limit"`
	UsedLimit      int64            `json:"used_limit"`
	AvailableLimit int64            `json:"available_limit"`
	Factors        []FactorScoreDTO `json:"factors"`
	Status         string           `json:"status"`
	LastAssessedAt string           `json:"last_assessed_at"`
}

type FactorScoreDTO struct {
	Factor   string `json:"factor"`
	Score    int    `json:"score"`
	Comments string `json:"comments"`
}

type CreditRecordDTO struct {
	ID          uint64 `json:"id"`
	ProfileID   uint64 `json:"profile_id"`
	Type        int    `json:"type"`
	ScoreChange int    `json:"score_change"`
	LimitChange int64  `json:"limit_change"`
	RelatedID   string `json:"related_id"`
	Reason      string `json:"reason"`
	HappenedAt  string `json:"happened_at"`
}

func toAssessmentDTO(a *domain.RiskAssessment) *AssessmentDTO {
	dto := &AssessmentDTO{
		ID:           a.ID,
		TargetID:     a.TargetID,
		TargetType:   a.TargetType,
		UserID:       a.UserID,
		TotalScore:   a.TotalScore,
		Level:        string(a.Level),
		Decision:     a.Decision,
		DeviceFinger: a.DeviceFinger,
		IPAddress:    a.IPAddress,
		CreatedAt:    a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	dto.MatchedRules = make([]MatchedRuleDTO, len(a.MatchedRules))
	for i, r := range a.MatchedRules {
		dto.MatchedRules[i] = MatchedRuleDTO{
			RuleID:      r.RuleID,
			RuleName:    r.RuleName,
			Score:       r.Score,
			Reason:      r.Reason,
			RawEvidence: r.RawEvidence,
		}
	}

	return dto
}

func toCreditProfileDTO(p *domain.CreditProfile) *CreditProfileDTO {
	dto := &CreditProfileDTO{
		ID:             p.ID,
		UserID:         p.UserID,
		Score:          p.Score,
		Level:          string(p.Level),
		TotalLimit:     p.TotalLimit,
		UsedLimit:      p.UsedLimit,
		AvailableLimit: p.AvailableLimit,
		Status:         p.Status,
		LastAssessedAt: p.LastAssessedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	dto.Factors = make([]FactorScoreDTO, len(p.Factors))
	for i, f := range p.Factors {
		dto.Factors[i] = FactorScoreDTO{
			Factor:   string(f.Factor),
			Score:    f.Score,
			Comments: f.Comments,
		}
	}

	return dto
}

func toCreditRecordDTOs(records []*domain.CreditRecord) []CreditRecordDTO {
	dtos := make([]CreditRecordDTO, len(records))
	for i, r := range records {
		dtos[i] = CreditRecordDTO{
			ID:          r.ID,
			ProfileID:   r.ProfileID,
			Type:        int(r.Type),
			ScoreChange: r.ScoreChange,
			LimitChange: r.LimitChange,
			RelatedID:   r.RelatedID,
			Reason:      r.Reason,
			HappenedAt:  r.HappenedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return dtos
}
