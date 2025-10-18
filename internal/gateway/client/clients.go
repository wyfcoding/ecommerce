package client

import (
	"fmt"

	authv1 "ecommerce/api/auth/v1"
	orderv1 "ecommerce/api/order/v1"
	productv1 "ecommerce/api/product/v1"
	userv1 "ecommerce/api/user/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Clients holds all the gRPC clients for downstream services.
type Clients struct {
	AuthClient    authv1.AuthServiceClient
	ProductClient productv1.ProductClient
	OrderClient   orderv1.OrderClient
	UserClient    userv1.UserClient
}

// NewClients initializes and returns a new Clients struct.
func NewClients(authServiceAddr, productServiceAddr, orderServiceAddr, userServiceAddr *string) (*Clients, func(), error) {
	var cleanupFuncs []func()

	// Auth Service Client
	authConn, err := grpc.Dial(*authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}
	cleanupFuncs = append(cleanupFuncs, func() { authConn.Close() })
	authClient := authv1.NewAuthServiceClient(authConn)

	// Product Service Client
	productConn, err := grpc.Dial(*productServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, combineCleanup(cleanupFuncs), fmt.Errorf("failed to connect to product service: %w", err)
	}
	cleanupFuncs = append(cleanupFuncs, func() { productConn.Close() })
	productClient := productv1.NewProductClient(productConn)

	// Order Service Client
	orderConn, err := grpc.Dial(*orderServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, combineCleanup(cleanupFuncs), fmt.Errorf("failed to connect to order service: %w", err)
	}
	cleanupFuncs = append(cleanupFuncs, func() { orderConn.Close() })
	orderClient := orderv1.NewOrderClient(orderConn)

	// User Service Client
	userConn, err := grpc.Dial(*userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, combineCleanup(cleanupFuncs), fmt.Errorf("failed to connect to user service: %w", err)
	}
	cleanupFuncs = append(cleanupFuncs, func() { userConn.Close() })
	userClient := userv1.NewUserClient(userConn)

	clients := &Clients{
		AuthClient:    authClient,
		ProductClient: productClient,
		OrderClient:   orderClient,
		UserClient:    userClient,
	}

	return clients, combineCleanup(cleanupFuncs), nil
}

func combineCleanup(funcs []func()) func() {
	return func() {
		for _, f := range funcs {
			f()
		}
	}
}
