package grpcclient

import (
	"context"
	"time"

	"ecommerce/pkg/logging"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ClientFactory creates gRPC clients.
type ClientFactory struct {
	logger *logging.Logger
}

// NewClientFactory creates a new ClientFactory.
func NewClientFactory(logger *logging.Logger) *ClientFactory {
	return &ClientFactory{
		logger: logger,
	}
}

// NewClient creates a new gRPC client connection.
func (f *ClientFactory) NewClient(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// Default options
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()), // Add tracing
		grpc.WithUnaryInterceptor(f.unaryClientInterceptor()),
	}

	opts = append(defaultOpts, opts...)

	conn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		f.logger.ErrorContext(ctx, "failed to dial grpc target", "target", target, "error", err)
		return nil, err
	}

	return conn, nil
}

func (f *ClientFactory) unaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		cost := time.Since(start)

		if err != nil {
			f.logger.ErrorContext(ctx, "grpc client call failed",
				"method", method,
				"target", cc.Target(),
				"cost", cost,
				"error", err,
			)
		} else {
			f.logger.DebugContext(ctx, "grpc client call success",
				"method", method,
				"target", cc.Target(),
				"cost", cost,
			)
		}

		return err
	}
}
