package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusMetrics struct {
	registry *prometheus.Registry
}

func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		registry: prometheus.NewRegistry(),
	}
}

func (p *PrometheusMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}
