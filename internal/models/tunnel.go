package models

type TunnelConfig struct {
	ID         string `bson:"_id" json:"id"`
	CustomerID string `bson:"customer_id" json:"customer_id"`
	// Only need the target URL since tunnel is pre-configured
	TargetURL string `bson:"target_url" json:"target_url"`
	Status    string `bson:"status" json:"status"`
}
