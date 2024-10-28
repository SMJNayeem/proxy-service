package integration

// import (
// 	"context"
// 	"proxy-service/internal/config"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestTunnelIntegration(t *testing.T) {
// 	ctx := context.Background()
// 	cfg, err := config.Load("../../config/config.test.yaml")
// 	assert.NoError(t, err)

// 	// Test cases
// 	tests := []struct {
// 		name          string
// 		customerID    string
// 		expectedError bool
// 	}{
// 		{
// 			name:          "Valid Customer",
// 			customerID:    "test-customer",
// 			expectedError: false,
// 		},
// 		{
// 			name:          "Invalid Customer",
// 			customerID:    "invalid",
// 			expectedError: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Test tunnel request
// 			resp, err := makeProxyRequest(ctx, tt.customerID)
// 			if tt.expectedError {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, resp)
// 			}
// 		})
// 	}
// }
