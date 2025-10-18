package service

import (
	"context"
	"errors"
	"fmt"

	"ecommerce/internal/gateway/client"
)

var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrServiceNotFound = errors.New("service not found for path")
)

// GatewayService is the business logic for the API Gateway.
type GatewayService struct {
	authClient client.AuthClient
	// TODO: Add clients for other microservices (Product, Order, User, etc.)
}

// NewGatewayService creates a new GatewayService.
func NewGatewayService(authClient client.AuthClient) *GatewayService {
	return &GatewayService{authClient: authClient}
}

// HandleRequest handles a generic incoming request, including authentication and routing.
func (s *GatewayService) HandleRequest(ctx context.Context, method, path, body, accessToken string, headers map[string]string) (statusCode int32, responseBody string, responseHeaders map[string]string, err error) {
	// 1. Authentication (if access token is provided)
	if accessToken != "" {
		isValid, userID, username, authErr := s.authClient.ValidateToken(ctx, accessToken)
		if authErr != nil {
			return 401, `{"message": "Token validation failed"}`, nil, fmt.Errorf("%w: %v", ErrUnauthorized, authErr)
		}
		if !isValid {
			return 401, `{"message": "Invalid or expired token"}`, nil, ErrUnauthorized
		}
		// Add user info to context or headers for downstream services
		headers["x-user-id"] = fmt.Sprintf("%d", userID)
		headers["x-username"] = username
	}

	// 2. Routing (simplified placeholder)
	// In a real gateway, this would involve a sophisticated routing mechanism
	// based on path, method, and potentially other criteria.
	// It would then call the appropriate microservice client.
	switch {
	case path == "/v1/products/create":
		// Simulate calling Product Service
		// For now, just return a dummy response
		return 200, `{"message": "Product created successfully (simulated)"}`, nil, nil
	case path == "/v1/orders/create":
		// Simulate calling Order Service
		return 200, `{"message": "Order created successfully (simulated)"}`, nil, nil
	case path == "/v1/user/register":
		// Simulate calling User Service
		return 200, `{"message": "User registered successfully (simulated)"}`, nil, nil
	default:
		return 404, `{"message": "Not Found"}`, nil, ErrServiceNotFound
	}
}
