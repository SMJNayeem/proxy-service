package middleware

import (
	"fmt"
	"net/http"
	"proxy-service/pkg/circuitbreaker"
	"proxy-service/pkg/validator"
	"time"

	"github.com/gin-gonic/gin"
)

type ResilienceMiddleware struct {
	validator      *validator.RequestValidator
	circuitBreaker *circuitbreaker.CircuitBreaker
}

func NewResilienceMiddleware(validator *validator.RequestValidator) *ResilienceMiddleware {
	cb := circuitbreaker.New(circuitbreaker.CircuitBreakerConfig{
		MaxFailures: 5,
		Timeout:     time.Second * 30,
	})

	return &ResilienceMiddleware{
		validator:      validator,
		circuitBreaker: cb,
	}
}

// Handle wraps an http.Handler with validation and circuit breaker
func (m *ResilienceMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Validate request
		if err := m.validator.ValidateRequest(c.Request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		// 2. Check circuit breaker
		if m.circuitBreaker.GetState() == circuitbreaker.StateOpen {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "service is temporarily unavailable",
			})
			c.Abort()
			return
		}

		// 3. Execute request with circuit breaker
		err := m.circuitBreaker.Execute(c.Request.Context(), func() error {
			c.Next()
			if c.Writer.Status() >= 500 {
				return fmt.Errorf("server error: %d", c.Writer.Status())
			}
			return nil
		})

		if err != nil {
			if err == circuitbreaker.ErrCircuitOpen {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "service is temporarily unavailable",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
			c.Abort()
		}
	}
}
