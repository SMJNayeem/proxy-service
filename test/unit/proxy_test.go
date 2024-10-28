package unit

// import (
// 	"context"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestProxyService(t *testing.T) {
// 	// Setup
// 	ctx := context.Background()
// 	proxyService := setupProxyService()

// 	// Test cases
// 	tests := []struct {
// 		name          string
// 		customerID    string
// 		path          string
// 		expectedError bool
// 	}{
// 		{
// 			name:          "Valid Request",
// 			customerID:    "test-customer",
// 			path:          "/api/test",
// 			expectedError: false,
// 		},
// 		{
// 			name:          "Invalid Customer",
// 			customerID:    "invalid",
// 			path:          "/api/test",
// 			expectedError: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Test proxy request
// 			resp, err := proxyService.HandleRequest(ctx, tt.customerID, tt.path)
// 			if tt.expectedError {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, resp)
// 			}
// 		})
// 	}
// }
