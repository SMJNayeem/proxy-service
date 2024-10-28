package models

import (
	"time"

	"github.com/gorilla/websocket"
)

type Agent struct {
	ID          string                 `json:"id" bson:"_id"`
	CustomerID  string                 `json:"customer_id" bson:"customer_id"`
	Name        string                 `json:"name" bson:"name"`
	Status      string                 `json:"status" bson:"status"`
	Version     string                 `json:"version" bson:"version"`
	LastSeen    time.Time              `json:"last_seen" bson:"last_seen"`
	CreatedAt   time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" bson:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata" bson:"metadata"`
	Permissions []string               `json:"permissions" bson:"permissions"`
}

func (a *Agent) IsActive() bool {
	return a.Status == "active" &&
		time.Since(a.LastSeen) < 5*time.Minute // Consider agent inactive if not seen in 5 minutes
}

type AgentStatus struct {
	AgentID       string       `json:"agent_id"`
	Status        string       `json:"status"`
	LastHeartbeat time.Time    `json:"last_heartbeat"`
	Version       string       `json:"version"`
	Metrics       AgentMetrics `json:"metrics"`
}

type AgentMetrics struct {
	ConnectionUptime  float64 `json:"connection_uptime"`
	RequestsProcessed int64   `json:"requests_processed"`
	AverageLatency    float64 `json:"average_latency"`
	ErrorCount        int64   `json:"error_count"`
	MemoryUsage       float64 `json:"memory_usage"`
	CPUUsage          float64 `json:"cpu_usage"`
	Uptime            float64 `json:"uptime"`
}

type AgentConnection struct {
	ID         string
	CustomerID string
	Conn       *websocket.Conn
	StartTime  time.Time
	LastPing   time.Time
}
