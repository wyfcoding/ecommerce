package handler

import (
	"net/http"
	"strconv"

	"ecommerce/internal/settlement/service"
	"github.com/gin-gonic/gin"
)

// SettlementHandler handles HTTP requests for the settlement service.
type SettlementHandler struct {
	service *service.SettlementService
}

// NewSettlementHandler creates a new SettlementHandler.
func NewSettlementHandler(s *service.SettlementService) *SettlementHandler {
	return &SettlementHandler{service: s}
}

// RegisterRoutes registers all the routes for the settlement service.
func (h *SettlementHandler) RegisterRoutes(e *gin.Engine) {
	api := e.Group("/api/v1/settlements")
	{
		api.POST("/process", processOrderSettlementHandler(h.service))
		api.GET("/:record_id", getSettlementRecordHandler(h.service))
		api.GET("", listSettlementRecordsHandler(h.service))
	}
}

func processOrderSettlementHandler(s *service.SettlementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OrderID    uint64 `json:"order_id"`
			MerchantID uint64 `json:"merchant_id"`
			TotalAmount uint64 `json:"total_amount"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		record, err := s.ProcessOrderSettlement(c.Request.Context(), req.OrderID, req.MerchantID, req.TotalAmount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, record)
	}
}

func getSettlementRecordHandler(s *service.SettlementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		recordID := c.Param("record_id")
		record, err := s.GetSettlementRecord(c.Request.Context(), recordID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, record)
	}
}

func listSettlementRecordsHandler(s *service.SettlementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		merchantID, err := strconv.ParseUint(c.DefaultQuery("merchant_id", "0"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchant ID"})
			return
		}
		status := c.Query("status")
		pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 32)
		pageNum, _ := strconv.ParseInt(c.DefaultQuery("page_num", "1"), 10, 32)

		records, total, err := s.ListSettlementRecords(c.Request.Context(), merchantID, status, uint32(pageSize), uint32(pageNum))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"records": records, "total": total})
	}
}
