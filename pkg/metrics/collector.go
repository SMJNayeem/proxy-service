package metrics

import (
	"strconv"
	"time"

	"proxy-service/internal/config"
	"proxy-service/internal/models"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type MetricsCollector struct {
	requestCounter      *prometheus.CounterVec
	requestDuration     *prometheus.HistogramVec
	errorCounter        *prometheus.CounterVec
	responseSize        *prometheus.HistogramVec
	activeConnections   *prometheus.GaugeVec
	authFailures        *prometheus.CounterVec
	authSuccesses       *prometheus.CounterVec
	authDuration        *prometheus.HistogramVec
	agentConnections    *prometheus.GaugeVec
	agentDisconnections *prometheus.CounterVec
	agentRequestCounter *prometheus.CounterVec
	agentErrors         *prometheus.CounterVec
	agentLatency        *prometheus.HistogramVec
	agentUptime         *prometheus.GaugeVec
	agentMemoryUsage    *prometheus.GaugeVec
	agentCPUUsage       *prometheus.GaugeVec
}

type ProxyHandler struct {
	config  *config.ProxyConfig
	metrics *MetricsCollector
}

func NewProxyHandler(config *config.ProxyConfig, metrics *MetricsCollector) *ProxyHandler {
	return &ProxyHandler{
		config:  config,
		metrics: metrics,
	}
}

func NewMetricsCollector() *MetricsCollector {
	mc := &MetricsCollector{
		// Existing metrics
		requestCounter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "proxy_requests_total",
				Help: "Total number of HTTP requests processed",
			},
			[]string{"customer_id", "path", "method", "status"},
		),
		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "proxy_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"customer_id", "path", "method"},
		),
		errorCounter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "proxy_errors_total",
				Help: "Total number of errors encountered",
			},
			[]string{"customer_id", "error_type"},
		),
		// Add new auth metrics
		authFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "proxy_auth_failures_total",
				Help: "Total number of authentication failures",
			},
			[]string{"reason"},
		),
		authSuccesses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "proxy_auth_successes_total",
				Help: "Total number of successful authentications",
			},
			[]string{"type"},
		),
		authDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "proxy_auth_duration_seconds",
				Help:    "Authentication request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"status"},
		),
		agentConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "proxy_agent_connections",
				Help: "Number of active agent connections",
			},
			[]string{"customer_id"},
		),

		agentDisconnections: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "proxy_agent_disconnections_total",
				Help: "Total number of agent disconnections",
			},
			[]string{"customer_id", "reason"},
		),

		agentRequestCounter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "proxy_agent_requests_total",
				Help: "Total number of requests processed by agents",
			},
			[]string{"customer_id", "agent_id", "status"},
		),

		agentErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "proxy_agent_errors_total",
				Help: "Total number of agent errors",
			},
			[]string{"customer_id", "agent_id", "error_type"},
		),

		agentLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "proxy_agent_request_duration_seconds",
				Help:    "Agent request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"customer_id", "agent_id"},
		),

		agentUptime: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "proxy_agent_uptime_seconds",
				Help: "Agent uptime in seconds",
			},
			[]string{"customer_id", "agent_id"},
		),

		agentMemoryUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "proxy_agent_memory_bytes",
				Help: "Agent memory usage in bytes",
			},
			[]string{"customer_id", "agent_id"},
		),

		agentCPUUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "proxy_agent_cpu_usage",
				Help: "Agent CPU usage percentage",
			},
			[]string{"customer_id", "agent_id"},
		),
	}
	return mc
}

// Add new auth-related methods
func (c *MetricsCollector) RecordAuthFailure(reason string) {
	c.authFailures.WithLabelValues(reason).Inc()
}

func (c *MetricsCollector) RecordAuthSuccess(authType string) {
	c.authSuccesses.WithLabelValues(authType).Inc()
}

func (c *MetricsCollector) RecordAuthDuration(duration time.Duration) {
	c.authDuration.WithLabelValues("success").Observe(duration.Seconds())
}

func (c *MetricsCollector) RecordRequestDuration(customerID, path, method string, duration time.Duration) {
	c.requestDuration.WithLabelValues(customerID, path, method).Observe(duration.Seconds())
}

func (c *MetricsCollector) RecordRequestCount(customerID, path, method string, status int) {
	c.requestCounter.WithLabelValues(customerID, path, method, strconv.Itoa(status)).Inc()
}

func (c *MetricsCollector) RecordResponseSize(customerID, path string, size float64) {
	c.responseSize.WithLabelValues(customerID, path).Observe(size)
}

func (c *MetricsCollector) RecordError(customerID, errorType string) {
	c.errorCounter.WithLabelValues(customerID, errorType).Inc()
}

func (c *MetricsCollector) UpdateActiveConnections(customerID string, count int) {
	c.activeConnections.WithLabelValues(customerID).Set(float64(count))
}

func (c *MetricsCollector) GetCurrentMetrics(customerID string) models.CurrentMetrics {
	var currentMetrics models.CurrentMetrics

	// Use the underlying gauge vector directly
	currentMetrics.ActiveConnections = int(getGaugeValue(c.activeConnections, customerID))

	// Calculate request rate from counter
	totalRequests := getCounterValue(c.requestCounter, customerID)
	currentMetrics.RequestRate = totalRequests / 60.0 // requests per second

	// Get average latency
	latencySum := getHistogramValue(c.requestDuration, customerID)
	if latencySum > 0 {
		currentMetrics.AverageLatency = latencySum
	}

	// Calculate error rate
	totalErrors := getCounterValue(c.errorCounter, customerID)
	if totalRequests > 0 {
		currentMetrics.ErrorRate = totalErrors / totalRequests
	}

	return currentMetrics
}

// Helper functions to get metric values
func getGaugeValue(vec *prometheus.GaugeVec, customerID string) float64 {
	// Get the metric value directly from the vector
	metric, err := vec.GetMetricWithLabelValues(customerID)
	if err != nil {
		return 0
	}
	return float64(metric.(prometheus.Gauge).Desc().String()[0])
}

func getCounterValue(vec *prometheus.CounterVec, customerID string) float64 {
	metric, err := vec.GetMetricWithLabelValues(customerID, "", "", "")
	if err != nil {
		return 0
	}
	return float64(metric.(prometheus.Counter).Desc().String()[0])
}

func getHistogramValue(vec *prometheus.HistogramVec, customerID string) float64 {
	metric, err := vec.GetMetricWithLabelValues(customerID, "", "")
	if err != nil {
		return 0
	}
	return float64(metric.(prometheus.Histogram).Desc().String()[0])
}

func (c *MetricsCollector) RecordRequest(customerID, path, method string, duration float64) {
	// Increment request counter
	c.requestCounter.WithLabelValues(customerID, path, method, "200").Inc()

	// Record request duration
	c.requestDuration.WithLabelValues(customerID, path, method).Observe(duration)
}

func (c *MetricsCollector) RecordAgentConnection(customerID string) {
	c.agentConnections.WithLabelValues(customerID).Inc()
}

func (c *MetricsCollector) RecordAgentDisconnection(customerID string) {
	c.agentDisconnections.WithLabelValues(customerID).Inc()
}

func (c *MetricsCollector) RecordAgentRequest(customerID, agentID string, status int) {
	c.agentRequestCounter.WithLabelValues(customerID, agentID, strconv.Itoa(status)).Inc()
}

func (c *MetricsCollector) RecordAgentError(customerID, agentID, errorType string) {
	c.agentErrors.WithLabelValues(customerID, agentID, errorType).Inc()
}

func (c *MetricsCollector) RecordAgentLatency(customerID, agentID string, duration time.Duration) {
	c.agentLatency.WithLabelValues(customerID, agentID).Observe(duration.Seconds())
}

func (c *MetricsCollector) UpdateAgentUptime(customerID, agentID string, uptime float64) {
	c.agentUptime.WithLabelValues(customerID, agentID).Set(uptime)
}

func (c *MetricsCollector) UpdateAgentMemoryUsage(customerID, agentID string, bytes float64) {
	c.agentMemoryUsage.WithLabelValues(customerID, agentID).Set(bytes)
}

func (c *MetricsCollector) UpdateAgentCPUUsage(customerID, agentID string, percentage float64) {
	c.agentCPUUsage.WithLabelValues(customerID, agentID).Set(percentage)
}

func (c *MetricsCollector) RecordAgentHeartbeat(customerID, agentID string) {
	c.agentConnections.WithLabelValues(customerID, agentID).Set(1)
}

func (c *MetricsCollector) UpdateAgentMetrics(customerID, agentID string, metrics *models.AgentMetrics) {
	c.agentMemoryUsage.WithLabelValues(customerID, agentID).Set(metrics.MemoryUsage)
	c.agentCPUUsage.WithLabelValues(customerID, agentID).Set(metrics.CPUUsage)
	c.agentUptime.WithLabelValues(customerID, agentID).Set(metrics.Uptime)
}
