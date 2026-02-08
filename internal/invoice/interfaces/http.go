// Package interfaces 发票服务接口层
package interfaces

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/invoice/application"
	"github.com/wyfcoding/ecommerce/internal/invoice/domain"
)

// HTTPHandler HTTP 接口处理器
type HTTPHandler struct {
	commandService *application.CommandService
	queryService   *application.QueryService
}

// NewHTTPHandler 创建 HTTP 处理器
func NewHTTPHandler(
	commandService *application.CommandService,
	queryService *application.QueryService,
) *HTTPHandler {
	return &HTTPHandler{
		commandService: commandService,
		queryService:   queryService,
	}
}

// RegisterRoutes 注册路由
func (h *HTTPHandler) RegisterRoutes(r *gin.RouterGroup) {
	invoices := r.Group("/invoices")
	{
		invoices.POST("/apply", h.Apply)
		invoices.POST("/:id/issue", h.Issue)
		invoices.POST("/:id/red", h.ApplyRed)
		invoices.GET("/:id", h.Get)
		invoices.GET("", h.List)
	}
}

// ApplyRequest 申请发票请求
type ApplyRequest struct {
	OrderNo    string                       `json:"order_no" binding:"required"`
	UserID     uint64                       `json:"user_id" binding:"required"`
	MerchantID uint64                       `json:"merchant_id" binding:"required"`
	Amount     int64                        `json:"amount" binding:"required"`
	Type       int8                         `json:"type" binding:"required"`
	Medium     int8                         `json:"medium" binding:"required"`
	Title      application.InvoiceTitleDTO  `json:"title" binding:"required"`
	Items      []application.InvoiceItemDTO `json:"items" binding:"required"`
}

// Apply 申请发票
func (h *HTTPHandler) Apply(c *gin.Context) {
	var req ApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.ApplyInvoiceCommand{
		OrderNo:    req.OrderNo,
		UserID:     req.UserID,
		MerchantID: req.MerchantID,
		Amount:     req.Amount,
		Type:       domain.InvoiceType(req.Type),
		Medium:     domain.InvoiceMedium(req.Medium),
		Title:      req.Title,
		Items:      req.Items,
	}

	result, err := h.commandService.ApplyInvoice(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// IssueRequest 开具发票请求
type IssueRequest struct {
	InvoiceCode string `json:"invoice_code" binding:"required"`
	InvoiceNo   string `json:"invoice_no" binding:"required"`
	CheckCode   string `json:"check_code"`
	PDFUrl      string `json:"pdf_url"`
	XMLUrl      string `json:"xml_url"`
}

// Issue 开具发票
func (h *HTTPHandler) Issue(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req IssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.IssueInvoiceCommand{
		InvoiceID:   uint(id),
		InvoiceCode: req.InvoiceCode,
		InvoiceNo:   req.InvoiceNo,
		CheckCode:   req.CheckCode,
		PDFUrl:      req.PDFUrl,
		XMLUrl:      req.XMLUrl,
	}

	if err := h.commandService.IssueInvoice(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "issued"})
}

// ApplyRedRequest 申请红冲请求
type ApplyRedRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// ApplyRed 申请红冲
func (h *HTTPHandler) ApplyRed(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req ApplyRedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.ApplyRedInvoiceCommand{
		OriginInvoiceID: uint(id),
		Reason:          req.Reason,
	}

	result, err := h.commandService.ApplyRedInvoice(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// Get 获取发票详情
func (h *HTTPHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	invoice, err := h.queryService.GetInvoice(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// List 发票列表
func (h *HTTPHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	merchantID, _ := strconv.ParseUint(c.Query("merchant_id"), 10, 64)

	filter := &domain.InvoiceFilter{
		UserID:     userID,
		MerchantID: merchantID,
		OrderNo:    c.Query("order_no"),
		Page:       page,
		PageSize:   pageSize,
	}

	if statusStr := c.Query("status"); statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			s := domain.InvoiceStatus(status)
			filter.Status = &s
		}
	}

	result, err := h.queryService.ListInvoices(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
