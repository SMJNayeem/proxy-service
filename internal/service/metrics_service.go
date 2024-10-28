package service

import (
	"context"
	"proxy-service/internal/models"
	"proxy-service/internal/repository"
	"proxy-service/pkg/metrics"
	"time"
)

type MetricsService struct {
	repo      *repository.MetricsRepository
	collector *metrics.MetricsCollector
}

func NewMetricsService(repo *repository.MetricsRepository, collector *metrics.MetricsCollector) *MetricsService {
	return &MetricsService{
		repo:      repo,
		collector: collector,
	}
}

func (s *MetricsService) RecordMetric(ctx context.Context, metric *models.MetricData) error {
	// Save to database
	if err := s.repo.SaveMetric(ctx, metric); err != nil {
		return err
	}

	// Update real-time metrics
	s.collector.RecordRequest(metric.CustomerID, metric.Path, metric.Method, float64(metric.Latency))
	return nil
}

func (s *MetricsService) GetAggregatedMetrics(ctx context.Context, customerID string, duration time.Duration) (*models.AggregatedMetrics, error) {
	startTime := time.Now().Add(-duration)
	return s.repo.GetAggregatedMetrics(ctx, customerID, startTime)
}

func (s *MetricsService) GetMetrics(ctx context.Context, customerID string) (*models.MetricsResponse, error) {
	// Get metrics for the last hour by default
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	// Get aggregated metrics from repository
	aggregatedMetrics, err := s.repo.GetAggregatedMetrics(ctx, customerID, startTime)
	if err != nil {
		return nil, err
	}

	// Get current metrics from collector
	currentMetrics := s.collector.GetCurrentMetrics(customerID)

	// Combine the metrics into a response
	response := &models.MetricsResponse{
		CustomerID: customerID,
		TimeRange: models.TimeRange{
			Start: startTime,
			End:   endTime,
		},
		Aggregated: *aggregatedMetrics,
		Current:    currentMetrics,
	}

	return response, nil
}

func (s *MetricsService) GetCollector() *metrics.MetricsCollector {
	return s.collector
}
