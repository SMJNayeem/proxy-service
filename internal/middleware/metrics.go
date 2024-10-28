package middleware

import (
	"proxy-service/pkg/metrics"
	"time"

	"github.com/gin-gonic/gin"
)

func Metrics(collector *metrics.MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Get customer ID from context, use "unknown" if not found
		customerID, exists := c.Get("customer_id")
		custID := "unknown"
		if exists {
			custID = customerID.(string)
		}

		// Process request
		c.Next()

		// Record metrics after request is processed
		duration := time.Since(start)
		path := c.Request.URL.Path
		method := c.Request.Method
		status := c.Writer.Status()

		// Record various metrics
		collector.RecordRequestCount(custID, path, method, status)
		collector.RecordRequestDuration(custID, path, method, duration)
		collector.RecordResponseSize(custID, path, float64(c.Writer.Size()))

		// Record error metrics if applicable
		if status >= 400 {
			errorType := "client_error"
			if status >= 500 {
				errorType = "server_error"
			}
			collector.RecordError(custID, errorType)
		}
	}
}
