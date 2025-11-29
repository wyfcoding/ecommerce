package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
)

// CircuitBreaker returns a new circuit breaker middleware
func CircuitBreaker() gin.HandlerFunc {
	st := gobreaker.Settings{
		Name:        "HTTP-Circuit-Breaker",
		MaxRequests: 0, // No limit on requests in half-open state
		Interval:    60 * time.Second,
		Timeout:     60 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.6
		},
	}
	cb := gobreaker.NewCircuitBreaker(st)

	return func(c *gin.Context) {
		_, err := cb.Execute(func() (interface{}, error) {
			c.Next()
			if c.Writer.Status() >= 500 {
				return nil, http.ErrHandlerTimeout // Treat 5xx as failure
			}
			return nil, nil
		})

		if err != nil {
			// If circuit is open or execution failed
			if err == gobreaker.ErrOpenState {
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "circuit breaker open"})
				return
			}
			// Other errors are handled by c.Next() or are just status codes
		}
	}
}
