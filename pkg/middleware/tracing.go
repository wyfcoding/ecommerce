package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// TracingMiddleware creates a tracing middleware.
func TracingMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract trace context from headers
		ctx := c.Request.Context()
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(c.Request.Header))

		tracer := otel.Tracer(serviceName)
		ctx, span := tracer.Start(ctx, c.Request.Method+" "+c.FullPath())
		defer span.End()

		// Inject trace context into request context
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// Record status code if needed (omitted for brevity)
	}
}
