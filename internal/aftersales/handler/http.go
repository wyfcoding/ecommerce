package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/aftersales/service"
	"ecommerce/pkg/response"
)

// AfterSalesHandler 售后服务HTTP处理器
type AfterSalesHandler struct {
	service service.AfterSalesService
	logger  *zap.Logger
}

// NewAfterSalesHandler 创建售后服务HTTP处理器
func NewAfterSalesHandler(service service.AfterSalesService, logger *zap.Logger) *AfterSalesHandler {
	return &AfterSalesHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册路由
func (h *AfterSalesHandler) RegisterRoutes(r *gin.RouterGroup) {
	aftersales := r.Group("/aftersales")
	{
		// 退款
		aftersales.POST("/refunds", h.ApplyRefund)
		aftersales.GET("/refunds/:id", h.GetRefund)
		aftersales.GET("/refunds", h.ListRefunds)
		aftersales.POST("/refunds/:id/approve", h.ApproveRefund)
		aftersales.POST("/refunds/:id/reject", h.RejectRefund)
		aftersales.POST("/refunds/:id/process", h.ProcessRefund)
		
		// 换货
		aftersales.POST("/exchanges", h.ApplyExchange)
		aftersales.GET("/exchanges/:id", h.GetExchange)
		aftersales.GET("/exchanges", h.ListExchanges)
		aftersales.POST("/exchanges/:id/approve", h.ApproveExchange)
		aftersales.POST("/exchanges/:id/ship", h.ShipExchange)
		
		// 维修
		aftersales.POST("/repairs", h.ApplyRepair)
		aftersales.GET("/repairs/:id", h.GetRepair)
		aftersales.GET("/repairs", h.ListRepairs)
		aftersales.POST("/repairs/:id/update-status", h.UpdateRepairStatus)
		
		// 工单
		aftersales.POST("/tickets", h.CreateTicket)
		aftersales.GET("/tickets/:id", h.GetTicket)
		aftersales.GET("/tickets", h.ListTickets)
		aftersales.POST("/tickets/:id/assign", h.AssignTicket)
		aftersales.POST("/tickets/:id/close", h.CloseTicket)
	}
}

// ApplyRefund 申请退款
func (h *AfterSalesHandler) ApplyRefund(c *gin.Context) {
	var req struct {
		OrderID    uint64 `json:"orderId" binding:"required"`
		OrderNo    string `json:"orderNo" binding:"required"`
		Reason     string `json:"reason" binding:"required"`
		Amount     int64  `json:"amount" binding:"required"`
		Images     string `json:"images"`
		ReturnType string `json:"returnType" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	userID := c.GetUint64("userID")

	refund, err := h.service.ApplyRefund(c.Request.Context(), userID, req.OrderID, req.OrderNo, req.Reason, req.Amount, req.Images, req.ReturnType)
	if err != nil {
		h.logger.Error("申请退款失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "申请退款失败", err)
		return
	}

	response.Success(c, refund)
}

// GetRefund 获取退款详情
func (h *AfterSalesHandler) GetRefund(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	refund, err := h.service.GetRefund(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "退款单不存在", err)
		return
	}

	response.Success(c, refund)
}

// ListRefunds 获取退款列表
func (h *AfterSalesHandler) ListRefunds(c *gin.Context) {
	userID := c.GetUint64("userID")
	status := c.Query("status")
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))

	refunds, total, err := h.service.ListRefunds(c.Request.Context(), userID, status, int32(pageSize), int32(pageNum))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取退款列表失败", err)
		return
	}

	response.SuccessWithPagination(c, refunds, total, int32(pageNum), int32(pageSize))
}

// ApproveRefund 审核退款
func (h *AfterSalesHandler) ApproveRefund(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	var req struct {
		Approved bool   `json:"approved"`
		Remark   string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	approverID := c.GetUint64("userID")

	if err := h.service.ApproveRefund(c.Request.Context(), id, approverID, req.Approved, req.Remark); err != nil {
		response.Error(c, http.StatusInternalServerError, "审核退款失败", err)
		return
	}

	response.Success(c, nil)
}

// RejectRefund 拒绝退款
func (h *AfterSalesHandler) RejectRefund(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := h.service.RejectRefund(c.Request.Context(), id, req.Reason); err != nil {
		response.Error(c, http.StatusInternalServerError, "拒绝退款失败", err)
		return
	}

	response.Success(c, nil)
}

// ProcessRefund 处理退款
func (h *AfterSalesHandler) ProcessRefund(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := h.service.ProcessRefund(c.Request.Context(), id); err != nil {
		response.Error(c, http.StatusInternalServerError, "处理退款失败", err)
		return
	}

	response.Success(c, nil)
}

// ApplyExchange 申请换货
func (h *AfterSalesHandler) ApplyExchange(c *gin.Context) {
	var req struct {
		OrderID uint64 `json:"orderId" binding:"required"`
		OrderNo string `json:"orderNo" binding:"required"`
		Reason  string `json:"reason" binding:"required"`
		Images  string `json:"images"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	userID := c.GetUint64("userID")

	exchange, err := h.service.ApplyExchange(c.Request.Context(), userID, req.OrderID, req.OrderNo, req.Reason, req.Images)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "申请换货失败", err)
		return
	}

	response.Success(c, exchange)
}

// GetExchange 获取换货详情
func (h *AfterSalesHandler) GetExchange(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	exchange, err := h.service.GetExchange(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "换货单不存在", err)
		return
	}

	response.Success(c, exchange)
}

// ListExchanges 获取换货列表
func (h *AfterSalesHandler) ListExchanges(c *gin.Context) {
	userID := c.GetUint64("userID")
	status := c.Query("status")
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))

	exchanges, total, err := h.service.ListExchanges(c.Request.Context(), userID, status, int32(pageSize), int32(pageNum))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取换货列表失败", err)
		return
	}

	response.SuccessWithPagination(c, exchanges, total, int32(pageNum), int32(pageSize))
}

// ApproveExchange 审核换货
func (h *AfterSalesHandler) ApproveExchange(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	var req struct {
		Approved bool   `json:"approved"`
		Remark   string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	approverID := c.GetUint64("userID")

	if err := h.service.ApproveExchange(c.Request.Context(), id, approverID, req.Approved, req.Remark); err != nil {
		response.Error(c, http.StatusInternalServerError, "审核换货失败", err)
		return
	}

	response.Success(c, nil)
}

// ShipExchange 换货发货
func (h *AfterSalesHandler) ShipExchange(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	var req struct {
		ExpressCompany string `json:"expressCompany" binding:"required"`
		ExpressNo      string `json:"expressNo" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := h.service.ShipExchange(c.Request.Context(), id, req.ExpressCompany, req.ExpressNo); err != nil {
		response.Error(c, http.StatusInternalServerError, "换货发货失败", err)
		return
	}

	response.Success(c, nil)
}

// ApplyRepair 申请维修
func (h *AfterSalesHandler) ApplyRepair(c *gin.Context) {
	var req struct {
		OrderID uint64 `json:"orderId" binding:"required"`
		OrderNo string `json:"orderNo" binding:"required"`
		Problem string `json:"problem" binding:"required"`
		Images  string `json:"images"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	userID := c.GetUint64("userID")

	repair, err := h.service.ApplyRepair(c.Request.Context(), userID, req.OrderID, req.OrderNo, req.Problem, req.Images)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "申请维修失败", err)
		return
	}

	response.Success(c, repair)
}

// GetRepair 获取维修详情
func (h *AfterSalesHandler) GetRepair(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	repair, err := h.service.GetRepair(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "维修单不存在", err)
		return
	}

	response.Success(c, repair)
}

// ListRepairs 获取维修列表
func (h *AfterSalesHandler) ListRepairs(c *gin.Context) {
	userID := c.GetUint64("userID")
	status := c.Query("status")
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))

	repairs, total, err := h.service.ListRepairs(c.Request.Context(), userID, status, int32(pageSize), int32(pageNum))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取维修列表失败", err)
		return
	}

	response.SuccessWithPagination(c, repairs, total, int32(pageNum), int32(pageSize))
}

// UpdateRepairStatus 更新维修状态
func (h *AfterSalesHandler) UpdateRepairStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		Remark string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := h.service.UpdateRepairStatus(c.Request.Context(), id, req.Status, req.Remark); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新维修状态失败", err)
		return
	}

	response.Success(c, nil)
}

// CreateTicket 创建工单
func (h *AfterSalesHandler) CreateTicket(c *gin.Context) {
	var req struct {
		Type        string `json:"type" binding:"required"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description" binding:"required"`
		OrderNo     string `json:"orderNo"`
		Images      string `json:"images"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	userID := c.GetUint64("userID")

	ticket, err := h.service.CreateTicket(c.Request.Context(), userID, req.Type, req.Title, req.Description, req.OrderNo, req.Images)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建工单失败", err)
		return
	}

	response.Success(c, ticket)
}

// GetTicket 获取工单详情
func (h *AfterSalesHandler) GetTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	ticket, err := h.service.GetTicket(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "工单不存在", err)
		return
	}

	response.Success(c, ticket)
}

// ListTickets 获取工单列表
func (h *AfterSalesHandler) ListTickets(c *gin.Context) {
	userID := c.GetUint64("userID")
	status := c.Query("status")
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))

	tickets, total, err := h.service.ListTickets(c.Request.Context(), userID, status, int32(pageSize), int32(pageNum))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取工单列表失败", err)
		return
	}

	response.SuccessWithPagination(c, tickets, total, int32(pageNum), int32(pageSize))
}

// AssignTicket 分配工单
func (h *AfterSalesHandler) AssignTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	var req struct {
		AgentID uint64 `json:"agentId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := h.service.AssignTicket(c.Request.Context(), id, req.AgentID); err != nil {
		response.Error(c, http.StatusInternalServerError, "分配工单失败", err)
		return
	}

	response.Success(c, nil)
}

// CloseTicket 关闭工单
func (h *AfterSalesHandler) CloseTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	var req struct {
		Resolution string `json:"resolution" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := h.service.CloseTicket(c.Request.Context(), id, req.Resolution); err != nil {
		response.Error(c, http.StatusInternalServerError, "关闭工单失败", err)
		return
	}

	response.Success(c, nil)
}
