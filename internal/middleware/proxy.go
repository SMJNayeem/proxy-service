package middleware

import (
	"net/http"
	"proxy-service/pkg/utils"

	"github.com/gin-gonic/gin"
)

func Proxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get allowed routes from context
		allowedRoutes, exists := c.Get("allowed_routes")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "route access not configured",
				"code":  "ROUTE_ACCESS_NOT_CONFIGURED",
			})
			return
		}

		// Check if current path is allowed
		if !utils.IsRouteAllowed(c.Request.URL.Path, allowedRoutes.([]string)) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "route not allowed",
				"code":  "ROUTE_NOT_ALLOWED",
			})
			return
		}

		// Add proxy headers
		c.Request.Header.Set("X-Forwarded-For", c.ClientIP())
		c.Request.Header.Set("X-Real-IP", c.ClientIP())
		c.Request.Header.Set("X-Proxy-ID", "proxy-service")

		c.Next()
	}
}
