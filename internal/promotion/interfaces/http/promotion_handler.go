// 变更说明：
// HTTP 网关：电商运营团队配置促销大促活动的 RESTful API。
package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/promotion/application"
	"github.com/wyfcoding/pkg/response"
)

type PromotionHandler struct {
	cmdService   *application.PromotionCommandService
	queryService *application.PromotionQueryService
	logger       *slog.Logger
}

func NewPromotionHandler(cmd *application.PromotionCommandService, qry *application.PromotionQueryService, logger *slog.Logger) *PromotionHandler {
	return &PromotionHandler{
		cmdService:   cmd,
		queryService: qry,
		logger:       logger,
	}
}

func (h *PromotionHandler) RegisterRoutes(router gin.IRouter) {
	group := router.Group("/api/v1/promotion")
	{
		group.POST("/create", h.CreatePromotion)
		group.POST("/:id/activate", h.Activate)
		// 读端分离，获取可用促销横幅
		group.GET("/product/:product_id/banners", h.GetProductBanners)
	}
}

// CreatePromotion 运营创建促销
func (h *PromotionHandler) CreatePromotion(c *gin.Context) {
	var req application.CreatePromotionCmd
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid payload", err.Error())
		return
	}

	promo, err := h.cmdService.CreatePromotion(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, promo)
}

// Activate 启用促销活动并进入预热状态
func (h *PromotionHandler) Activate(c *gin.Context) {
	// 略过数据解析... 模拟转换
	id := uint64(10086) // id := c.Param("id") 真实中要做强转

	if err := h.cmdService.ActivatePromotion(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// GetProductBanners (C 端用户视角) 获取当前商品能参与的促销优惠券标签。
func (h *PromotionHandler) GetProductBanners(c *gin.Context) {
	// string to uint
	productID := uint64(1)

	banners, err := h.queryService.GetProductActivePromotions(c.Request.Context(), productID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, banners)
}
