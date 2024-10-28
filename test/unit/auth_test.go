package unit

// import (
// 	"context"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestAuthService(t *testing.T) {
// 	// Setup
// 	ctx := context.Background()
// 	authService := setupAuthService()

// 	// Test cases
// 	tests := []struct {
// 		name          string
// 		apiKey        string
// 		expectedError bool
// 	}{
// 		{
// 			name:          "Valid API Key",
// 			apiKey:        "valid-key",
// 			expectedError: false,
// 		},
// 		{
// 			name:          "Invalid API Key",
// 			apiKey:        "invalid-key",
// 			expectedError: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Test token generation
// 			token, err := authService.GenerateToken(ctx, tt.apiKey)
// 			if tt.expectedError {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotEmpty(t, token)
// 			}
// 		})
// 	}
// }
