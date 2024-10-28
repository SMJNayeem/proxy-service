package handler

import (
	"net/http"
	"proxy-service/internal/service"

	"github.com/gin-gonic/gin"
)

type MetricsHandler struct {
	metricsService *service.MetricsService
}

func NewMetricsHandler(service *service.MetricsService) *MetricsHandler {
	return &MetricsHandler{
		metricsService: service,
	}
}

func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	customerID := c.GetString("customer_id")

	// Get metrics
	metrics, err := h.metricsService.GetMetrics(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get metrics",
			"code":  "METRICS_FETCH_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}
