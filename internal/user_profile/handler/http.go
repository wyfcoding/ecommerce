package userprofilehandler

import (
	"net/http"
	"time"

	"ecommerce/internal/user_profile/model"
	"ecommerce/internal/user_profile/service"
	"github.com/gin-gonic/gin"
)

// UserProfileHandler handles HTTP requests for the user profile service.
type UserProfileHandler struct {
	service *service.UserProfileService
}

// NewUserProfileHandler creates a new UserProfileHandler.
func NewUserProfileHandler(s *service.UserProfileService) *UserProfileHandler {
	return &UserProfileHandler{service: s}
}

// RegisterRoutes registers all the routes for the user profile service.
func (h *UserProfileHandler) RegisterRoutes(e *gin.Engine) {
	api := e.Group("/api/v1/user-profiles")
	{
		api.GET("/:user_id", h.GetUserProfile)
		api.PUT("/:user_id", h.UpdateUserProfile)
		api.POST("/behavior", h.RecordUserBehavior)
	}
}

// GetUserProfile is a Gin handler for getting a user profile by ID.
func (h *UserProfileHandler) GetUserProfile(c *gin.Context) {
	userID := c.Param("user_id")

	profile, err := h.service.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateUserProfile is a Gin handler for updating a user profile.
func (h *UserProfileHandler) UpdateUserProfile(c *gin.Context) {
	userID := c.Param("user_id")

	var req model.UserProfile
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.UserID = userID

	updatedProfile, err := h.service.UpdateUserProfile(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	c.JSON(http.StatusOK, updatedProfile)
}

// RecordUserBehavior is a Gin handler for recording user behavior.
func (h *UserProfileHandler) RecordUserBehavior(c *gin.Context) {
	var req struct {
		UserID       string            `json:"user_id"`
		BehaviorType string            `json:"behavior_type"`
		ItemID       string            `json:"item_id"`
		Properties   map[string]string `json:"properties"`
		EventTime    time.Time         `json:"event_time"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.RecordUserBehavior(
		c.Request.Context(),
		req.UserID,
		req.BehaviorType,
		req.ItemID,
		req.Properties,
		req.EventTime,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user behavior recorded"})
}
