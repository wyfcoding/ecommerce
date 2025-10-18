package marketinghandler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	v1 "ecommerce/api/marketing/v1"
	"ecommerce/internal/marketing/model"
	"ecommerce/internal/marketing/service"
	"ecommerce/pkg/logging"
)

// StartHTTPServer starts the HTTP Gateway, which proxies HTTP requests to gRPC services.
func StartHTTPServer(couponService *service.CouponService, promotionService *service.PromotionService, addr string, port int) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", addr, port)

	err := v1.RegisterMarketingServiceHandlerFromEndpoint(context.Background(), mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery())

	// Add service-specific Gin routes here
	api := r.Group("/api/v1/marketing")
	{
		// Coupon routes
		api.POST("/coupons/templates", createCouponTemplateHandler(couponService))
		api.POST("/coupons/claim", claimCouponHandler(couponService))
		api.POST("/coupons/calculate-discount", calculateDiscountHandler(couponService))

		// Promotion routes
		api.POST("/promotions", createPromotionHandler(promotionService))
		api.PUT("/promotions/:id", updatePromotionHandler(promotionService))
		api.DELETE("/promotions/:id", deletePromotionHandler(promotionService))
		api.GET("/promotions/:id", getPromotionHandler(promotionService))
		api.GET("/promotions", listPromotionsHandler(promotionService))
	}

	r.Any("/*any", gin.WrapH(mux))

	httpEndpoint := fmt.Sprintf("%s:%d", addr, port)
	server := &http.Server{
		Addr:    httpEndpoint,
		Handler: r,
	}

	zap.S().Infof("HTTP server listening at %s", httpEndpoint)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to serve HTTP: %w", err)
		}
		close(errChan)
	}()
	return server, errChan
}

func createCouponTemplateHandler(s *service.CouponService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.CouponTemplate
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		template, err := s.CreateCouponTemplate(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, template)
	}
}

func claimCouponHandler(s *service.CouponService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID     uint64 `json:"user_id"`
			TemplateID uint64 `json:"template_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		userCoupon, err := s.ClaimCoupon(c.Request.Context(), req.UserID, req.TemplateID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, userCoupon)
	}
}

func calculateDiscountHandler(s *service.CouponService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID     uint64              `json:"user_id"`
			CouponCode string              `json:"coupon_code"`
			Items      []*model.OrderItemInfo `json:"items"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		discount, err := s.CalculateDiscount(c.Request.Context(), req.UserID, req.CouponCode, req.Items)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"discount_amount": discount})
	}
}

func createPromotionHandler(s *service.PromotionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.Promotion
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		promotion, err := s.CreatePromotion(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, promotion)
	}
}

func updatePromotionHandler(s *service.PromotionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid promotion ID"})
			return
		}
		var req model.Promotion
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		req.ID = id
		promotion, err := s.UpdatePromotion(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, promotion)
	}
}

func deletePromotionHandler(s *service.PromotionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid promotion ID"})
			return
		}
		if err := s.DeletePromotion(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNoContent, nil)
	}
}

func getPromotionHandler(s *service.PromotionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid promotion ID"})
			return
		}
		promotion, err := s.GetPromotion(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, promotion)
	}
}

func listPromotionsHandler(s *service.PromotionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageSize, _ := strconv.ParseUint(c.DefaultQuery("page_size", "10"), 10, 32)
		pageNum, _ := strconv.ParseUint(c.DefaultQuery("page_num", "1"), 10, 32)
		name := c.Query("name")
		promoTypeStr := c.Query("type")
		statusStr := c.Query("status")

		var promoType *uint32
		if promoTypeStr != "" {
			p, err := strconv.ParseUint(promoTypeStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid promo type"})
				return
			}
			pt := uint32(p)
			promoType = &pt
		}

		var status *uint32
		if statusStr != "" {
			s, err := strconv.ParseUint(statusStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
				return
			}
			st := uint32(s)
			status = &st
		}

		promotions, total, err := s.ListPromotions(c.Request.Context(), uint32(pageSize), uint32(pageNum), &name, promoType, status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"promotions": promotions, "total": total})
	}
}
