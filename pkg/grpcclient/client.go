package grpcclient

import (
	"context"
	"time"

	"ecommerce/pkg/logging"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

var (
	grpcClientRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_client_requests_total",
			Help: "The total number of grpc client requests",
		},
		[]string{"method", "target", "status"},
	)
	grpcClientDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_client_duration_seconds",
			Help:    "The duration of grpc client requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "target"},
	)
)

func init() {
	prometheus.MustRegister(grpcClientRequests, grpcClientDuration)
}

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
	// Initialize Circuit Breaker per client target
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "grpc-client-" + target,
		MaxRequests: 0,
		Interval:    0,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.6
		},
	})

	// Initialize Rate Limiter (e.g., 1000 rps)
	limiter := rate.NewLimiter(rate.Limit(1000), 100)

	// Default options
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()), // Add tracing
		grpc.WithChainUnaryInterceptor(
			f.metricsInterceptor(),
			f.circuitBreakerInterceptor(cb),
			f.rateLimitInterceptor(limiter),
			f.retryInterceptor(),
			f.loggingInterceptor(),
		),
	}

	opts = append(defaultOpts, opts...)

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		f.logger.ErrorContext(ctx, "failed to dial grpc target", "target", target, "error", err)
		return nil, err
	}

	return conn, nil
}

func (f *ClientFactory) metricsInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(start).Seconds()

		statusStr := "success"
		if err != nil {
			statusStr = status.Code(err).String()
		}

		grpcClientRequests.WithLabelValues(method, cc.Target(), statusStr).Inc()
		grpcClientDuration.WithLabelValues(method, cc.Target()).Observe(duration)

		return err
	}
}

func (f *ClientFactory) circuitBreakerInterceptor(cb *gobreaker.CircuitBreaker) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		_, err := cb.Execute(func() (interface{}, error) {
			return nil, invoker(ctx, method, req, reply, cc, opts...)
		})
		return err
	}
}

func (f *ClientFactory) rateLimitInterceptor(limiter *rate.Limiter) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if err := limiter.Wait(ctx); err != nil {
			return status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (f *ClientFactory) retryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var err error
		for i := 0; i < 3; i++ {
			err = invoker(ctx, method, req, reply, cc, opts...)
			if err == nil {
				return nil
			}
			// Retry only on specific codes
			code := status.Code(err)
			if code != codes.Unavailable && code != codes.DeadlineExceeded {
				return err
			}
			time.Sleep(time.Duration(1<<i) * 100 * time.Millisecond)
		}
		return err
	}
}

func (f *ClientFactory) loggingInterceptor() grpc.UnaryClientInterceptor {
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
