package repository

import (
	"context"
	"time"

	"proxy-service/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AgentRepository struct {
	db *mongo.Database
}

func NewAgentRepository(db *mongo.Database) *AgentRepository {
	return &AgentRepository{
		db: db,
	}
}

func (r *AgentRepository) SaveAgent(ctx context.Context, agent *models.Agent) error {
	collection := r.db.Collection("agents")

	agent.UpdatedAt = time.Now()
	if agent.CreatedAt.IsZero() {
		agent.CreatedAt = agent.UpdatedAt
	}

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"_id": agent.ID}
	update := bson.M{"$set": agent}

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *AgentRepository) GetAgent(ctx context.Context, agentID string) (*models.Agent, error) {
	collection := r.db.Collection("agents")

	var agent models.Agent
	err := collection.FindOne(ctx, bson.M{"_id": agentID}).Decode(&agent)
	if err != nil {
		return nil, err
	}

	return &agent, nil
}

func (r *AgentRepository) GetAgentsByCustomer(ctx context.Context, customerID string) ([]*models.Agent, error) {
	collection := r.db.Collection("agents")

	cursor, err := collection.Find(ctx, bson.M{"customer_id": customerID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var agents []*models.Agent
	if err = cursor.All(ctx, &agents); err != nil {
		return nil, err
	}

	return agents, nil
}

func (r *AgentRepository) UpdateAgentStatus(ctx context.Context, agentID, status string) error {
	collection := r.db.Collection("agents")

	update := bson.M{
		"$set": bson.M{
			"status":         status,
			"last_connected": time.Now(),
			"updated_at":     time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": agentID}, update)
	return err
}

func (r *AgentRepository) DeleteAgent(ctx context.Context, agentID string) error {
	collection := r.db.Collection("agents")
	_, err := collection.DeleteOne(ctx, bson.M{"_id": agentID})
	return err
}
