package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	pb "github.com/wyfcoding/ecommerce/go-api/cms/v1"
	"github.com/wyfcoding/ecommerce/internal/cms/application"
)

type Handler struct {
	svc    *application.CMSService
	logger *slog.Logger
}

func NewHandler(svc *application.CMSService, logger *slog.Logger) *Handler {
	return &Handler{
		svc:    svc,
		logger: logger.With("component", "cms_http"),
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1/cms")
	{
		v1.GET("/pages/:slug", h.GetPageBySlug)
	}
}

func (h *Handler) GetPageBySlug(c *gin.Context) {
	slug := c.Param("slug")
	page, err := h.svc.GetPage(c.Request.Context(), &pb.GetPageRequest{
		Query: &pb.GetPageRequest_Slug{Slug: slug},
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
		return
	}

	c.JSON(http.StatusOK, page)
}
