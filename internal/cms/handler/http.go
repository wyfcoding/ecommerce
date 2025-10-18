package cmshandler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"ecommerce/internal/cms/model"
	"ecommerce/internal/cms/service"
	"ecommerce/pkg/logging"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// StartHTTPServer starts the HTTP server.
func StartHTTPServer(cmsService *service.CmsService, addr string, port int) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	r := gin.New()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery()) // Use project's GinLogger

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Register routes
	registerRoutes(r, cmsService)

	httpEndpoint := fmt.Sprintf("%s:%d", addr, port)
	server := &http.Server{
		Addr:    httpEndpoint,
		Handler: r,
	}

	zap.S().Infof("Gin HTTP server listening at %s", httpEndpoint)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to serve Gin HTTP: %w", err)
		}
		close(errChan)
	}()
	return server, errChan
}

func registerRoutes(r *gin.Engine, svc *service.CmsService) {
	api := r.Group("/api/v1/cms")
	{
		// Content Pages
		api.POST("/pages", createContentPageHandler(svc))
		api.GET("/pages/:id", getContentPageHandler(svc))
		api.GET("/pages/slug/:slug", getContentPageBySlugHandler(svc))
		api.PUT("/pages/:id", updateContentPageHandler(svc))
		api.DELETE("/pages/:id", deleteContentPageHandler(svc))
		api.GET("/pages", listContentPagesHandler(svc))

		// Content Blocks
		api.POST("/blocks", createContentBlockHandler(svc))
		api.GET("/blocks/:id", getContentBlockHandler(svc))
		api.GET("/blocks/name/:name", getContentBlockByNameHandler(svc))
		api.PUT("/blocks/:id", updateContentBlockHandler(svc))
		api.DELETE("/blocks/:id", deleteContentBlockHandler(svc))
	}
}

func createContentPageHandler(svc *service.CmsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Title       string `json:"title"`
			Slug        string `json:"slug"`
			ContentHTML string `json:"content_html"`
			Status      string `json:"status"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		page, err := svc.CreateContentPage(c.Request.Context(), req.Title, req.Slug, req.ContentHTML, req.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, page)
	}
}

func getContentPageHandler(svc *service.CmsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page ID"})
			return
		}
		page, err := svc.GetContentPage(c.Request.Context(), uint(id), "")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, page)
	}
}

func getContentPageBySlugHandler(svc *service.CmsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		page, err := svc.GetContentPage(c.Request.Context(), 0, slug)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, page)
	}
}

func updateContentPageHandler(svc *service.CmsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page ID"})
			return
		}
		var req model.ContentPage
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		req.ID = uint(id)
		page, err := svc.UpdateContentPage(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, page)
	}
}

func deleteContentPageHandler(svc *service.CmsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page ID"})
			return
		}
		if err := svc.DeleteContentPage(c.Request.Context(), uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNoContent, nil)
	}
}

func listContentPagesHandler(svc *service.CmsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		statusFilter := c.Query("status")
		pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 32)
		pageToken, _ := strconv.ParseInt(c.DefaultQuery("page_token", "0"), 10, 32)
		pages, total, err := svc.ListContentPages(c.Request.Context(), statusFilter, int32(pageSize), int32(pageToken))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"pages": pages, "total": total})
	}
}

func createContentBlockHandler(svc *service.CmsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string `json:"name"`
			ContentHTML string `json:"content_html"`
			Type        string `json:"type"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		block, err := svc.CreateContentBlock(c.Request.Context(), req.Name, req.ContentHTML, req.Type)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, block)
	}
}

func getContentBlockHandler(svc *service.CmsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid block ID"})
			return
		}
		block, err := svc.GetContentBlock(c.Request.Context(), uint(id), "")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, block)
	}
}

func getContentBlockByNameHandler(svc *service.CmsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		block, err := svc.GetContentBlock(c.Request.Context(), 0, name)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, block)
	}
}

func updateContentBlockHandler(svc *service.CmsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid block ID"})
			return
		}
		var req model.ContentBlock
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		req.ID = uint(id)
		block, err := svc.UpdateContentBlock(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, block)
	}
}

func deleteContentBlockHandler(svc *service.CmsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid block ID"})
			return
		}
		if err := svc.DeleteContentBlock(c.Request.Context(), uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNoContent, nil)
	}
}
