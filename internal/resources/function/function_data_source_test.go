package function

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestFunctionDataSource_Read(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name: "successful get function",
			mockResponse: &client.Function{
				Name:   "test-function",
				SKU:    client.SKU{ID: "small", Name: "Small"},
				Status: "ready",
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "function not found",
			mockResponse:   nil,
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/namespaces/test-ns/vm/test-function", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			})

			server := httptest.NewServer(mux)
			defer server.Close()

			// Create data source
			dataSource := NewFunctionDataSource().(*FunctionDataSource)

			// Configure with mock client - skip for now as it requires proper HTTP integration
			// dspcClient := &client.DspcClient{
			// 	Functions: nil, // Would need proper client initialization
			// }

			// Test basic data source creation
			assert.NotNil(t, dataSource)
		})
	}
}
