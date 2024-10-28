package models

import "time"

type MetricData struct {
	ID           string    `bson:"_id" json:"id"`
	CustomerID   string    `bson:"customer_id" json:"customer_id"`
	Path         string    `bson:"path" json:"path"`
	Method       string    `bson:"method" json:"method"`
	StatusCode   int       `bson:"status_code" json:"status_code"`
	Latency      float64   `bson:"latency" json:"latency"`
	RequestSize  int64     `bson:"request_size" json:"request_size"`
	ResponseSize int64     `bson:"response_size" json:"response_size"`
	Timestamp    time.Time `bson:"timestamp" json:"timestamp"`
}

type AggregatedMetrics struct {
	TotalRequests      int64   `json:"total_requests"`
	AverageLatency     float64 `json:"average_latency"`
	ErrorRate          float64 `json:"error_rate"`
	RequestsPerMinute  float64 `json:"requests_per_minute"`
	AverageRequestSize int64   `json:"average_request_size"`
}

type CurrentMetrics struct {
	ActiveConnections int     `json:"active_connections"`
	RequestRate       float64 `json:"request_rate"`
	AverageLatency    float64 `json:"average_latency"`
	ErrorRate         float64 `json:"error_rate"`
}

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type MetricsResponse struct {
	CustomerID string            `json:"customer_id"`
	TimeRange  TimeRange         `json:"time_range"`
	Aggregated AggregatedMetrics `json:"aggregated"`
	Current    CurrentMetrics    `json:"current"`
}
