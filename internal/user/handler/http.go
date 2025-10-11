package handler

import (
	"net/http"

	"ecommerce/api/user/v1"
	"ecommerce/internal/user/service"
	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests for the user service.
type UserHandler struct {
	service *service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{service: s}
}

// GetUser is a Gin handler for getting a user by ID.
func (h *UserHandler) GetUser(c *gin.Context) {
	// For simplicity, we'll get the ID from the path and call the gRPC service method.
	// In a real-world scenario, you might have a separate business logic method.
	id := c.Param("id")

	// The service method expects a gRPC request, so we create one.
	// This is not ideal, but it's a way to reuse the existing gRPC service logic.
	// A better approach would be to have a shared business logic layer.
	grpcReq := &v1.GetUserRequest{Id: id} // Assuming v1 is the api package
	user, err := h.service.GetUser(c.Request.Context(), grpcReq)

	if err != nil {
		// In a real app, you'd have proper error handling and mapping to HTTP status codes.
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
