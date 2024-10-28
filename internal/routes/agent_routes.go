package routes

import (
	"proxy-service/internal/handler/agent"
	"proxy-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAgentRoutes(router *gin.Engine, handler *agent.AgentHandler, authMiddleware *middleware.AuthMiddleware) {
	// Agent management routes
	agentGroup := router.Group("/api/v1/agents")
	{
		// Agent connection endpoint
		agentGroup.GET("/connect", handler.HandleConnection)

		// Protected routes requiring authentication
		protected := agentGroup.Use(authMiddleware.ValidateToken())
		{
			protected.GET("/health", handler.HandleAgentHealth)
			protected.GET("/metrics", handler.HandleAgentMetrics)
			protected.POST("/register", handler.HandleAgentRegistration)
			protected.DELETE("/deregister", handler.HandleAgentDeregistration)
		}
	}
}
