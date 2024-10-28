package repository

import (
	"context"
	"proxy-service/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MetricsRepository struct {
	db *mongo.Database
}

func NewMetricsRepository(db *mongo.Database) *MetricsRepository {
	return &MetricsRepository{
		db: db,
	}
}

func (r *MetricsRepository) SaveMetric(ctx context.Context, metric *models.MetricData) error {
	_, err := r.db.Collection("metrics").InsertOne(ctx, metric)
	return err
}

func (r *MetricsRepository) GetAggregatedMetrics(ctx context.Context, customerID string, startTime time.Time) (*models.AggregatedMetrics, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"customer_id": customerID,
				"timestamp":   bson.M{"$gte": startTime},
			},
		},
		{
			"$group": bson.M{
				"_id":             nil,
				"total_requests":  bson.M{"$sum": 1},
				"average_latency": bson.M{"$avg": "$latency"},
				"error_count": bson.M{"$sum": bson.M{
					"$cond": []interface{}{bson.M{"$gte": []interface{}{"$status_code", 400}}, 1, 0},
				}},
				"total_request_size": bson.M{"$sum": "$request_size"},
			},
		},
	}

	cursor, err := r.db.Collection("metrics").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return &models.AggregatedMetrics{}, nil
	}

	result := results[0]
	totalRequests := result["total_requests"].(int64)

	return &models.AggregatedMetrics{
		TotalRequests:      totalRequests,
		AverageLatency:     result["average_latency"].(float64),
		ErrorRate:          float64(result["error_count"].(int32)) / float64(totalRequests),
		RequestsPerMinute:  float64(totalRequests) / time.Since(startTime).Minutes(),
		AverageRequestSize: result["total_request_size"].(int64) / totalRequests,
	}, nil
}
