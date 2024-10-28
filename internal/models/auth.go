package models

import (
	"time"
)

type AuthRequest struct {
	APIKey string `json:"api_key"`
}

type AuthResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
}

type TokenClaims struct {
	CustomerID    string   `json:"customer_id"`
	AllowedRoutes []string `json:"allowed_routes"`
	ExpiresAt     int64    `json:"exp"`
	IssuedAt      int64    `json:"iat"`
}

type Customer struct {
	ID            string    `bson:"_id" json:"id"`
	Name          string    `bson:"name" json:"name"`
	APIKey        string    `bson:"api_key" json:"api_key"`
	Status        string    `bson:"status" json:"status"`
	AllowedRoutes []string  `bson:"allowed_routes" json:"allowed_routes"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time `bson:"updated_at" json:"updated_at"`
}
