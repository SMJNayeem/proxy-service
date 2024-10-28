package middleware

import (
	"net/http"
	"proxy-service/pkg/logger"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error and stack trace
				logger.Error("panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
				)

				// Get customer ID if available
				customerID, exists := c.Get("customer_id")
				if exists {
					logger.Error("customer context",
						"customer_id", customerID,
					)
				}

				// Return error response
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
					"code":  "INTERNAL_SERVER_ERROR",
				})
			}
		}()

		c.Next()
	}

}
