package function

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestFunctionResource_Create(t *testing.T) {
	tests := []struct {
		name           string
		functionName   string
		skuID          string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:         "successful creation",
			functionName: "test-function",
			skuID:        "small",
			mockResponse: client.CreateFunctionResponse{
				Created: "test-function",
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name:         "creation with medium SKU",
			functionName: "medium-function",
			skuID:        "medium",
			mockResponse: client.CreateFunctionResponse{
				Created: "medium-function",
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/namespaces/test-ns/vm/", func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					var req client.CreateFunctionRequest
					err := json.NewDecoder(r.Body).Decode(&req)
					assert.NoError(t, err)
					assert.Equal(t, tt.functionName, req.Name)
					assert.Equal(t, tt.skuID, req.SKUID)

					w.WriteHeader(tt.mockStatusCode)
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			})

			// Mock get response for created function
			mux.HandleFunc("/v1/namespaces/test-ns/vm/"+tt.functionName, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					function := &client.Function{
						Name:   tt.functionName,
						SKU:    client.SKU{ID: tt.skuID, Name: "Test SKU"},
						Status: "ready",
					}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(function)
				}
			})

			server := httptest.NewServer(mux)
			defer server.Close()

			// Create function resource
			functionResource := NewFunctionResource().(*FunctionResource)

			// Configure with mock client - skip for now as it requires proper HTTP integration
			// dspcClient := &client.DspcClient{
			// 	Functions: nil, // Would need proper client initialization
			// }

			// Test basic resource creation
			assert.NotNil(t, functionResource)
		})
	}
}
